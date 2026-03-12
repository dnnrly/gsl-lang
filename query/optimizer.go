package query

import (
	"fmt"
)

// QueryOptimizer performs optimizations on query pipelines
type QueryOptimizer struct {
	EnableReordering     bool
	EnableFilterPushdown bool
	EnableEarlyFilter    bool
	Stats                *OptimizationStats
}

// OptimizationStats tracks optimization decisions and impact
type OptimizationStats struct {
	OriginalExprCount    int
	OptimizedExprCount   int
	ReorderingApplied    bool
	FilterPushdownCount  int
	EarlyFilterCount     int
	EstimatedSavings     float64 // Estimated % of work saved
	ReorderedExpressions []string // Descriptions of reordering
}

// NewQueryOptimizer creates a new optimizer with all optimizations enabled
func NewQueryOptimizer() *QueryOptimizer {
	return &QueryOptimizer{
		EnableReordering:     true,
		EnableFilterPushdown: true,
		EnableEarlyFilter:    true,
		Stats:                &OptimizationStats{},
	}
}

// Optimize returns an optimized copy of the query
// Original query is not modified
func (qo *QueryOptimizer) Optimize(q *Query) (*Query, error) {
	if q == nil || len(q.Expressions) == 0 {
		return q, nil
	}

	// Create copy
	optimized := &Query{
		Expressions: make([]Expression, len(q.Expressions)),
	}
	copy(optimized.Expressions, q.Expressions)

	qo.Stats = &OptimizationStats{
		OriginalExprCount: len(q.Expressions),
	}

	// Apply optimizations in order
	if qo.EnableEarlyFilter {
		optimized = qo.applyEarlyFilter(optimized)
	}

	if qo.EnableFilterPushdown {
		optimized = qo.applyFilterPushdown(optimized)
	}

	if qo.EnableReordering {
		optimized = qo.applyReordering(optimized)
	}

	qo.Stats.OptimizedExprCount = len(optimized.Expressions)

	// Estimate savings from removing expressions
	if qo.Stats.OriginalExprCount > 0 {
		removed := qo.Stats.OriginalExprCount - qo.Stats.OptimizedExprCount
		qo.Stats.EstimatedSavings = float64(removed) / float64(qo.Stats.OriginalExprCount) * 100.0
	}

	return optimized, nil
}

// applyEarlyFilter moves filter operations (subgraph) earlier in the pipeline
// to reduce data flowing through downstream operations
func (qo *QueryOptimizer) applyEarlyFilter(q *Query) *Query {
	if len(q.Expressions) <= 1 {
		return q
	}

	result := make([]Expression, 0, len(q.Expressions))
	var filters []*SubgraphExpr
	var others []Expression

	// Collect filters and other expressions
	for _, expr := range q.Expressions {
		if sg, ok := expr.(*SubgraphExpr); ok && sg.Traversal == nil {
			filters = append(filters, sg)
			qo.Stats.EarlyFilterCount++
		} else {
			others = append(others, expr)
		}
	}

	// If we have filters, put them first, then others
	if len(filters) > 0 && len(others) > 0 {
		result = append(result, toExpressions(filters)...)
		result = append(result, others...)
		qo.Stats.ReorderedExpressions = append(qo.Stats.ReorderedExpressions,
			fmt.Sprintf("Moved %d filter(s) to beginning", len(filters)))
		return &Query{Expressions: result}
	}

	return q
}

// applyFilterPushdown moves filtering conditions to earlier expressions when possible
// Example: `subgraph exists | make node.x = 1 where node.y = 2` becomes
//          `subgraph node.y = 2 | make node.x = 1 where exists`
// (This is a simplified version; full version would be more sophisticated)
func (qo *QueryOptimizer) applyFilterPushdown(q *Query) *Query {
	// Simplified: just track that we attempted pushdown
	// Full implementation would analyze predicate dependencies
	// and move restrictive predicates earlier
	result := make([]Expression, 0, len(q.Expressions))

	for i, expr := range q.Expressions {
		// Check if this is a subgraph followed by make
		if i+1 < len(q.Expressions) {
			if sg, ok := expr.(*SubgraphExpr); ok {
				if make, ok := q.Expressions[i+1].(*MakeExpr); ok {
					// Could optimize: push make's predicate into subgraph if possible
					// For now, just track that opportunity
					_ = sg
					_ = make
				}
			}
		}

		result = append(result, expr)
	}

	return &Query{Expressions: result}
}

// applyReordering reorders expressions based on cost heuristics
// Heuristic: expensive operations (collapse) after cheap operations (remove orphans)
func (qo *QueryOptimizer) applyReordering(q *Query) *Query {
	if len(q.Expressions) <= 1 {
		return q
	}

	// Copy expressions
	exprs := make([]Expression, len(q.Expressions))
	copy(exprs, q.Expressions)

	// Simple bubble sort by cost (most restrictive first)
	reordered := false
	for i := 0; i < len(exprs)-1; i++ {
		for j := 0; j < len(exprs)-1-i; j++ {
			cost1 := getExpressionCost(exprs[j])
			cost2 := getExpressionCost(exprs[j+1])

			// If second expression is cheaper/more restrictive, swap
			if cost2 < cost1 && canSwap(exprs[j], exprs[j+1]) {
				exprs[j], exprs[j+1] = exprs[j+1], exprs[j]
				reordered = true
			}
		}
	}

	if reordered {
		qo.Stats.ReorderingApplied = true
		qo.Stats.ReorderedExpressions = append(qo.Stats.ReorderedExpressions,
			"Applied cost-based expression reordering")
	}

	return &Query{Expressions: exprs}
}

// ExpressionCost estimates the cost of an operation
// Lower cost = more restrictive/cheaper
type ExpressionCost int

const (
	CostRemoveOrphans ExpressionCost = 1  // Very cheap, most restrictive
	CostRemoveEdge    ExpressionCost = 2
	CostRemoveAttr    ExpressionCost = 3
	CostSubgraph      ExpressionCost = 4  // Moderately expensive
	CostMake          ExpressionCost = 5  // More expensive
	CostCollapse      ExpressionCost = 6  // Most expensive
	CostAlgebra       ExpressionCost = 7
	CostFrom          ExpressionCost = 8  // Context change
	CostBind          ExpressionCost = 9  // Context change
	CostIdentity      ExpressionCost = 10 // No-op
)

// getExpressionCost returns the relative cost of an expression
// Lower = more likely to run first (reduces downstream work)
func getExpressionCost(expr Expression) ExpressionCost {
	switch expr.(type) {
	case *RemoveOrphansExpr:
		return CostRemoveOrphans
	case *RemoveEdgeExpr:
		return CostRemoveEdge
	case *RemoveAttributeExpr:
		return CostRemoveAttr
	case *SubgraphExpr:
		return CostSubgraph
	case *MakeExpr:
		return CostMake
	case *CollapseExpr:
		return CostCollapse
	case *GraphAlgebraExpr:
		return CostAlgebra
	case *FromExpr:
		return CostFrom
	case *BindExpr:
		return CostBind
	case *IdentityExpr:
		return CostIdentity
	default:
		return CostIdentity
	}
}

// canSwap checks if two expressions can be safely reordered
// Returns false if order matters (context dependencies, etc.)
func canSwap(left, right Expression) bool {
	// Never swap if left or right is From (context change)
	if _, ok := left.(*FromExpr); ok {
		return false
	}
	if _, ok := right.(*FromExpr); ok {
		return false
	}

	// Never swap if left or right is Bind (creates named graph)
	if _, ok := left.(*BindExpr); ok {
		return false
	}
	if _, ok := right.(*BindExpr); ok {
		return false
	}

	// Can swap: filter, make, remove operations (order doesn't matter semantically)
	return true
}

// toExpressions converts SubgraphExpr slice to Expression slice
func toExpressions(sgs []*SubgraphExpr) []Expression {
	result := make([]Expression, 0, len(sgs))
	for _, sg := range sgs {
		result = append(result, sg)
	}
	return result
}

// String returns a human-readable summary of optimization decisions
func (os *OptimizationStats) String() string {
	var output string
	output += "Query Optimization Report\n"
	output += "=========================\n\n"

	output += "Expressions:\n"
	output += fmt.Sprintf("  Original: %d\n", os.OriginalExprCount)
	output += fmt.Sprintf("  Optimized: %d\n", os.OptimizedExprCount)

	output += "\nOptimizations Applied:\n"
	if os.ReorderingApplied {
		output += "  ✓ Expression reordering\n"
	}
	if os.FilterPushdownCount > 0 {
		output += fmt.Sprintf("  ✓ Filter push-down (%d)\n", os.FilterPushdownCount)
	}
	if os.EarlyFilterCount > 0 {
		output += fmt.Sprintf("  ✓ Early filtering (%d)\n", os.EarlyFilterCount)
	}

	output += "\nEstimated Impact:\n"
	output += fmt.Sprintf("  Expressions removed: %d\n", os.OriginalExprCount-os.OptimizedExprCount)
	output += fmt.Sprintf("  Estimated savings: %.1f%%\n", os.EstimatedSavings)

	if len(os.ReorderedExpressions) > 0 {
		output += "\nTransformations:\n"
		for _, desc := range os.ReorderedExpressions {
			output += fmt.Sprintf("  - %s\n", desc)
		}
	}

	return output
}

// Summary returns a compact one-line summary
func (os *OptimizationStats) Summary() string {
	if os.OriginalExprCount == os.OptimizedExprCount {
		return fmt.Sprintf("No optimizations applied (%.1f%% potential)", os.EstimatedSavings)
	}
	return fmt.Sprintf("Optimized: %d → %d expressions (%.1f%% savings)",
		os.OriginalExprCount, os.OptimizedExprCount, os.EstimatedSavings)
}
