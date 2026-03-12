package query

import (
	"fmt"
	"time"
)

// ExecutionStats tracks metrics about query execution
type ExecutionStats struct {
	// Result graph metrics
	ResultNodeCount int
	ResultEdgeCount int
	ResultSetCount  int

	// Input graph metrics
	InputNodeCount int
	InputEdgeCount int
	InputSetCount  int

	// Transformation metrics
	NodesAdded    int
	NodesRemoved  int
	EdgesAdded    int
	EdgesRemoved  int
	AttributesSet int

	// Timing metrics
	TotalDuration      time.Duration
	ExpressionDurations []ExpressionTiming

	// Traversal metrics (from subgraph with traverse)
	NodesTraversed int
	MaxTraversalDepth int
}

// ExpressionTiming tracks execution time for a single expression
type ExpressionTiming struct {
	Index    int           // 0-based position in pipeline
	ExprType string        // e.g., "SubgraphExpr", "MakeExpr"
	Duration time.Duration
	InputNodeCount  int
	OutputNodeCount int
	InputEdgeCount  int
	OutputEdgeCount int
}

// QueryExecutor wraps Query execution with statistics collection
type QueryExecutor struct {
	Query *Query
	CollectStats bool
}

// NewQueryExecutor creates a new executor for a query
func NewQueryExecutor(q *Query) *QueryExecutor {
	return &QueryExecutor{
		Query: q,
		CollectStats: true,
	}
}

// ExecuteWithStats runs the query and collects execution statistics
func (qe *QueryExecutor) ExecuteWithStats(ctx *QueryContext) (Value, *ExecutionStats, error) {
	stats := &ExecutionStats{
		InputNodeCount: len(ctx.InputGraph.Nodes),
		InputEdgeCount: len(ctx.InputGraph.Edges),
		InputSetCount:  len(ctx.InputGraph.Sets),
		ExpressionDurations: make([]ExpressionTiming, 0, len(qe.Query.Expressions)),
	}

	startTime := time.Now()

	// Initialize with input graph
	var value Value = GraphValue{ctx.InputGraph}

	// Apply each expression in sequence, tracking stats
	for i, expr := range qe.Query.Expressions {
		exprStart := time.Now()

		// Get before state
		beforeNodes := 0
		beforeEdges := 0
		if beforeGraphValue, ok := value.(GraphValue); ok && beforeGraphValue.Graph != nil {
			beforeNodes = len(beforeGraphValue.Graph.Nodes)
			beforeEdges = len(beforeGraphValue.Graph.Edges)
		}

		// Execute expression
		var err error
		value, err = expr.Apply(ctx, value)
		if err != nil {
			return nil, nil, err
		}

		// Get after state
		afterNodes := 0
		afterEdges := 0
		if afterGraphValue, ok := value.(GraphValue); ok && afterGraphValue.Graph != nil {
			afterNodes = len(afterGraphValue.Graph.Nodes)
			afterEdges = len(afterGraphValue.Graph.Edges)
		}

		// Record timing
		exprDuration := time.Since(exprStart)
		exprType := getExpressionType(expr)
		stats.ExpressionDurations = append(stats.ExpressionDurations, ExpressionTiming{
			Index:    i,
			ExprType: exprType,
			Duration: exprDuration,
			InputNodeCount:  beforeNodes,
			OutputNodeCount: afterNodes,
			InputEdgeCount:  beforeEdges,
			OutputEdgeCount: afterEdges,
		})

		// Track changes
		if afterNodes > beforeNodes {
			stats.NodesAdded += afterNodes - beforeNodes
		} else if afterNodes < beforeNodes {
			stats.NodesRemoved += beforeNodes - afterNodes
		}

		if afterEdges > beforeEdges {
			stats.EdgesAdded += afterEdges - beforeEdges
		} else if afterEdges < beforeEdges {
			stats.EdgesRemoved += beforeEdges - afterEdges
		}
	}

	stats.TotalDuration = time.Since(startTime)

	// Capture final result metrics
	if finalGraphValue, ok := value.(GraphValue); ok && finalGraphValue.Graph != nil {
		stats.ResultNodeCount = len(finalGraphValue.Graph.Nodes)
		stats.ResultEdgeCount = len(finalGraphValue.Graph.Edges)
		stats.ResultSetCount = len(finalGraphValue.Graph.Sets)
	}

	return value, stats, nil
}

// getExpressionType returns human-readable type name for an expression
func getExpressionType(expr Expression) string {
	switch expr.(type) {
	case *FromExpr:
		return "FromExpr"
	case *BindExpr:
		return "BindExpr"
	case *SubgraphExpr:
		return "SubgraphExpr"
	case *RemoveEdgeExpr:
		return "RemoveEdgeExpr"
	case *RemoveAttributeExpr:
		return "RemoveAttributeExpr"
	case *RemoveOrphansExpr:
		return "RemoveOrphansExpr"
	case *MakeExpr:
		return "MakeExpr"
	case *CollapseExpr:
		return "CollapseExpr"
	case *GraphAlgebraExpr:
		return "GraphAlgebraExpr"
	case *IdentityExpr:
		return "IdentityExpr"
	default:
		return "UnknownExpr"
	}
}

// String returns a human-readable summary of execution statistics
func (s *ExecutionStats) String() string {
	output := fmt.Sprintf("Query Execution Statistics\n")
	output += fmt.Sprintf("==========================\n\n")

	output += fmt.Sprintf("Result Graph:\n")
	output += fmt.Sprintf("  Nodes: %d (input: %d, +%d, -%d)\n",
		s.ResultNodeCount, s.InputNodeCount, s.NodesAdded, s.NodesRemoved)
	output += fmt.Sprintf("  Edges: %d (input: %d, +%d, -%d)\n",
		s.ResultEdgeCount, s.InputEdgeCount, s.EdgesAdded, s.EdgesRemoved)
	output += fmt.Sprintf("  Sets:  %d (input: %d)\n", s.ResultSetCount, s.InputSetCount)

	output += fmt.Sprintf("\nExecution Pipeline:\n")
	for i, timing := range s.ExpressionDurations {
		output += fmt.Sprintf("  [%d] %s\n", i+1, timing.String())
	}

	output += fmt.Sprintf("\nTiming:\n")
	output += fmt.Sprintf("  Total: %v\n", s.TotalDuration)

	return output
}

// String returns a human-readable summary of a single expression timing
func (et *ExpressionTiming) String() string {
	return fmt.Sprintf("%s %v (%d → %d nodes, %d → %d edges)",
		et.ExprType, et.Duration,
		et.InputNodeCount, et.OutputNodeCount,
		et.InputEdgeCount, et.OutputEdgeCount)
}

// Summary returns a compact one-line summary of execution statistics
func (s *ExecutionStats) Summary() string {
	return fmt.Sprintf(
		"Result: %d nodes, %d edges, %d sets | Time: %v | Changes: +%d/%d nodes, +%d/%d edges",
		s.ResultNodeCount, s.ResultEdgeCount, s.ResultSetCount,
		s.TotalDuration,
		s.NodesAdded, s.NodesRemoved,
		s.EdgesAdded, s.EdgesRemoved,
	)
}

// PerExpressionStats returns a detailed breakdown of stats per expression
func (s *ExecutionStats) PerExpressionStats() []string {
	result := make([]string, 0, len(s.ExpressionDurations))
	for _, timing := range s.ExpressionDurations {
		result = append(result, timing.String())
	}
	return result
}

// AverageExpressionTime returns the average execution time per expression
func (s *ExecutionStats) AverageExpressionTime() time.Duration {
	if len(s.ExpressionDurations) == 0 {
		return 0
	}
	total := time.Duration(0)
	for _, timing := range s.ExpressionDurations {
		total += timing.Duration
	}
	return total / time.Duration(len(s.ExpressionDurations))
}

// SlowestExpression returns the expression that took the longest
func (s *ExecutionStats) SlowestExpression() *ExpressionTiming {
	if len(s.ExpressionDurations) == 0 {
		return nil
	}
	slowest := &s.ExpressionDurations[0]
	for i := 1; i < len(s.ExpressionDurations); i++ {
		if s.ExpressionDurations[i].Duration > slowest.Duration {
			slowest = &s.ExpressionDurations[i]
		}
	}
	return slowest
}

// FastestExpression returns the expression that took the least time
func (s *ExecutionStats) FastestExpression() *ExpressionTiming {
	if len(s.ExpressionDurations) == 0 {
		return nil
	}
	fastest := &s.ExpressionDurations[0]
	for i := 1; i < len(s.ExpressionDurations); i++ {
		if s.ExpressionDurations[i].Duration < fastest.Duration {
			fastest = &s.ExpressionDurations[i]
		}
	}
	return fastest
}
