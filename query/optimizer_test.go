package query

import (
	"testing"

	"github.com/dnnrly/gsl-lang"
)

func TestOptimizerBasic(t *testing.T) {
	// Test: basic optimizer creation
	opt := NewQueryOptimizer()
	if !opt.EnableReordering || !opt.EnableFilterPushdown || !opt.EnableEarlyFilter {
		t.Errorf("optimizer not initialized with correct flags")
	}
	if opt.Stats == nil {
		t.Errorf("optimizer stats not initialized")
	}
}

func TestOptimizerNilQuery(t *testing.T) {
	// Test: optimizer handles nil query gracefully
	opt := NewQueryOptimizer()
	result, err := opt.Optimize(nil)
	if err != nil {
		t.Errorf("Optimize(nil) error = %v", err)
	}
	if result != nil {
		t.Errorf("Optimize(nil) = %v, want nil", result)
	}
}

func TestOptimizerEmptyQuery(t *testing.T) {
	// Test: optimizer handles empty query
	opt := NewQueryOptimizer()
	q := &Query{Expressions: []Expression{}}
	result, err := opt.Optimize(q)
	if err != nil {
		t.Errorf("Optimize(empty) error = %v", err)
	}
	if len(result.Expressions) != 0 {
		t.Errorf("Optimize(empty) expressions = %d, want 0", len(result.Expressions))
	}
}

func TestOptimizerNoOptimizationsNeeded(t *testing.T) {
	// Test: query with single expression needs no optimization
	opt := NewQueryOptimizer()
	q := &Query{
		Expressions: []Expression{&IdentityExpr{}},
	}
	result, err := opt.Optimize(q)
	if err != nil {
		t.Errorf("Optimize() error = %v", err)
	}
	if len(result.Expressions) != 1 {
		t.Errorf("expression count = %d, want 1", len(result.Expressions))
	}
	if opt.Stats.OriginalExprCount != 1 {
		t.Errorf("OriginalExprCount = %d, want 1", opt.Stats.OriginalExprCount)
	}
}

func TestOptimizerEarlyFilter(t *testing.T) {
	// Test: optimizer moves filters to beginning
	opt := NewQueryOptimizer()
	opt.EnableReordering = false

	q := &Query{
		Expressions: []Expression{
			&MakeExpr{
				Target: "node",
				Attr:   "x",
				Value:  1,
				Pred:   &ExistsPredicate{},
			},
			&SubgraphExpr{
				Pred:      &ExistsPredicate{},
				Traversal: nil,
			},
		},
	}

	result, err := opt.Optimize(q)
	if err != nil {
		t.Errorf("Optimize() error = %v", err)
	}

	// Should have SubgraphExpr first
	if _, ok := result.Expressions[0].(*SubgraphExpr); !ok {
		t.Errorf("first expression type = %T, want *SubgraphExpr", result.Expressions[0])
	}

	if opt.Stats.EarlyFilterCount != 1 {
		t.Errorf("EarlyFilterCount = %d, want 1", opt.Stats.EarlyFilterCount)
	}
}

func TestOptimizerFilterDoesNotReorderWithTraversal(t *testing.T) {
	// Test: subgraph with traversal is not moved (has side effect)
	opt := NewQueryOptimizer()
	opt.EnableReordering = false

	q := &Query{
		Expressions: []Expression{
			&MakeExpr{
				Target: "node",
				Attr:   "x",
				Value:  1,
				Pred:   &ExistsPredicate{},
			},
			&SubgraphExpr{
				Pred: &ExistsPredicate{},
				Traversal: &TraversalConfig{
					Direction: "out",
					Depth:     1,
				},
			},
		},
	}

	result, err := opt.Optimize(q)
	if err != nil {
		t.Errorf("Optimize() error = %v", err)
	}

	// Order should not change (traversal makes filter stateful)
	if _, ok := result.Expressions[0].(*MakeExpr); !ok {
		t.Errorf("first expression type = %T, want *MakeExpr", result.Expressions[0])
	}
}

func TestOptimizerFromNotReordered(t *testing.T) {
	// Test: From expressions never reordered (context change)
	opt := NewQueryOptimizer()

	q := &Query{
		Expressions: []Expression{
			&RemoveOrphansExpr{},
			&FromExpr{IsWildcard: true},
		},
	}

	result, err := opt.Optimize(q)
	if err != nil {
		t.Errorf("Optimize() error = %v", err)
	}

	// Order must be preserved
	if _, ok := result.Expressions[0].(*RemoveOrphansExpr); !ok {
		t.Errorf("first expression = %T, want *RemoveOrphansExpr", result.Expressions[0])
	}
	if _, ok := result.Expressions[1].(*FromExpr); !ok {
		t.Errorf("second expression = %T, want *FromExpr", result.Expressions[1])
	}
}

func TestOptimizerBindNotReordered(t *testing.T) {
	// Test: Bind expressions never reordered (creates named graph)
	opt := NewQueryOptimizer()

	q := &Query{
		Expressions: []Expression{
			&RemoveOrphansExpr{},
			&BindExpr{
				Pipeline: &Query{Expressions: []Expression{&IdentityExpr{}}},
				Name:     "G",
			},
		},
	}

	result, err := opt.Optimize(q)
	if err != nil {
		t.Errorf("Optimize() error = %v", err)
	}

	// Order must be preserved
	if _, ok := result.Expressions[0].(*RemoveOrphansExpr); !ok {
		t.Errorf("first expression = %T, want *RemoveOrphansExpr", result.Expressions[0])
	}
	if _, ok := result.Expressions[1].(*BindExpr); !ok {
		t.Errorf("second expression = %T, want *BindExpr", result.Expressions[1])
	}
}

func TestOptimizerStats(t *testing.T) {
	// Test: stats collection
	opt := NewQueryOptimizer()

	q := &Query{
		Expressions: []Expression{
			&MakeExpr{Target: "node", Attr: "x", Value: 1, Pred: &ExistsPredicate{}},
			&SubgraphExpr{Pred: &ExistsPredicate{}, Traversal: nil},
			&RemoveOrphansExpr{},
		},
	}

	result, err := opt.Optimize(q)
	if err != nil {
		t.Errorf("Optimize() error = %v", err)
	}

	stats := opt.Stats
	if stats.OriginalExprCount != 3 {
		t.Errorf("OriginalExprCount = %d, want 3", stats.OriginalExprCount)
	}
	if stats.OptimizedExprCount != 3 {
		t.Errorf("OptimizedExprCount = %d, want 3", stats.OptimizedExprCount)
	}
	if len(result.Expressions) != 3 {
		t.Errorf("result expression count = %d, want 3", len(result.Expressions))
	}
}

func TestOptimizerStatsString(t *testing.T) {
	// Test: stats string representation
	stats := &OptimizationStats{
		OriginalExprCount:   5,
		OptimizedExprCount:  4,
		ReorderingApplied:   true,
		FilterPushdownCount: 1,
		EarlyFilterCount:    2,
		EstimatedSavings:    20.0,
		ReorderedExpressions: []string{"Moved 2 filter(s) to beginning"},
	}

	str := stats.String()
	if !statsContainsStr(str, "Query Optimization Report") {
		t.Errorf("String() missing header")
	}
	if !statsContainsStr(str, "5") || !statsContainsStr(str, "4") {
		t.Errorf("String() missing expression counts")
	}
	if !statsContainsStr(str, "20.0") {
		t.Errorf("String() missing savings estimate")
	}
}

func TestOptimizerStatsSummary(t *testing.T) {
	// Test: stats summary
	stats := &OptimizationStats{
		OriginalExprCount:  5,
		OptimizedExprCount: 3,
		EstimatedSavings:   40.0,
	}

	summary := stats.Summary()
	if !statsContainsStr(summary, "5") || !statsContainsStr(summary, "3") {
		t.Errorf("Summary missing expression counts")
	}
	if !statsContainsStr(summary, "40") {
		t.Errorf("Summary missing savings estimate")
	}
}

func TestExpressionCosts(t *testing.T) {
	// Test: expression cost values are as expected
	tests := []struct {
		name     string
		expr     Expression
		expectedCost ExpressionCost
	}{
		{"RemoveOrphans", &RemoveOrphansExpr{}, CostRemoveOrphans},
		{"RemoveEdge", &RemoveEdgeExpr{Pred: &ExistsPredicate{}}, CostRemoveEdge},
		{"RemoveAttr", &RemoveAttributeExpr{Pred: &ExistsPredicate{}}, CostRemoveAttr},
		{"Subgraph", &SubgraphExpr{Pred: &ExistsPredicate{}}, CostSubgraph},
		{"Make", &MakeExpr{Pred: &ExistsPredicate{}}, CostMake},
		{"Collapse", &CollapseExpr{Pred: &ExistsPredicate{}}, CostCollapse},
		{"Algebra", &GraphAlgebraExpr{Operator: "+"}, CostAlgebra},
		{"From", &FromExpr{}, CostFrom},
		{"Bind", &BindExpr{}, CostBind},
		{"Identity", &IdentityExpr{}, CostIdentity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := getExpressionCost(tt.expr)
			if cost != tt.expectedCost {
				t.Errorf("getExpressionCost(%s) = %d, want %d", tt.name, cost, tt.expectedCost)
			}
		})
	}
}

func TestCanSwap(t *testing.T) {
	// Test: swap eligibility rules
	tests := []struct {
		name   string
		left   Expression
		right  Expression
		canSwap bool
	}{
		{"identity + identity", &IdentityExpr{}, &IdentityExpr{}, true},
		{"make + subgraph", &MakeExpr{}, &SubgraphExpr{}, true},
		{"from + remove", &FromExpr{}, &RemoveOrphansExpr{}, false},
		{"remove + from", &RemoveOrphansExpr{}, &FromExpr{}, false},
		{"bind + make", &BindExpr{}, &MakeExpr{}, false},
		{"make + bind", &MakeExpr{}, &BindExpr{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canSwap(tt.left, tt.right)
			if result != tt.canSwap {
				t.Errorf("canSwap(%T, %T) = %v, want %v", tt.left, tt.right, result, tt.canSwap)
			}
		})
	}
}

func TestOptimizerDisabledOptimizations(t *testing.T) {
	// Test: optimizer respects disabled optimizations
	opt := NewQueryOptimizer()
	opt.EnableEarlyFilter = false
	opt.EnableReordering = false
	opt.EnableFilterPushdown = false

	q := &Query{
		Expressions: []Expression{
			&MakeExpr{Target: "node", Attr: "x", Value: 1, Pred: &ExistsPredicate{}},
			&SubgraphExpr{Pred: &ExistsPredicate{}, Traversal: nil},
		},
	}

	result, err := opt.Optimize(q)
	if err != nil {
		t.Errorf("Optimize() error = %v", err)
	}

	// Order should not change
	if _, ok := result.Expressions[0].(*MakeExpr); !ok {
		t.Errorf("first expression = %T, want *MakeExpr", result.Expressions[0])
	}
	if _, ok := result.Expressions[1].(*SubgraphExpr); !ok {
		t.Errorf("second expression = %T, want *SubgraphExpr", result.Expressions[1])
	}
}

func TestOptimizerComplexPipeline(t *testing.T) {
	// Test: optimizer on realistic complex pipeline
	opt := NewQueryOptimizer()

	q := &Query{
		Expressions: []Expression{
			&MakeExpr{Target: "node", Attr: "priority", Value: "high", Pred: &ExistsPredicate{}},
			&CollapseExpr{NodeID: "MERGED", Pred: &ExistsPredicate{}},
			&SubgraphExpr{Pred: &ExistsPredicate{}, Traversal: nil},
			&RemoveOrphansExpr{},
		},
	}

	result, err := opt.Optimize(q)
	if err != nil {
		t.Errorf("Optimize() error = %v", err)
	}

	// Should have reordered: RemoveOrphans should be first
	if _, ok := result.Expressions[0].(*RemoveOrphansExpr); !ok {
		t.Errorf("first expression after optimization = %T, want *RemoveOrphansExpr", result.Expressions[0])
	}
}

func TestOptimizerPreservesSemantics(t *testing.T) {
	// Test: optimization doesn't change semantics
	input := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"a": {ID: "a", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
			"b": {ID: "b", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	queryStr := `remove orphans | make node.marked = true where exists`
	q, err := NewQueryParser(queryStr).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Execute original
	ctx1 := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}
	result1, err := q.Execute(ctx1)
	if err != nil {
		t.Fatalf("Execute(original) error = %v", err)
	}

	// Optimize and execute
	opt := NewQueryOptimizer()
	optimized, err := opt.Optimize(q)
	if err != nil {
		t.Fatalf("Optimize() error = %v", err)
	}

	ctx2 := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}
	result2, err := optimized.Execute(ctx2)
	if err != nil {
		t.Fatalf("Execute(optimized) error = %v", err)
	}

	// Results should be identical
	gv1 := result1.(GraphValue)
	gv2 := result2.(GraphValue)

	if len(gv1.Graph.Nodes) != len(gv2.Graph.Nodes) {
		t.Errorf("node count mismatch: %d vs %d", len(gv1.Graph.Nodes), len(gv2.Graph.Nodes))
	}
	if len(gv1.Graph.Edges) != len(gv2.Graph.Edges) {
		t.Errorf("edge count mismatch: %d vs %d", len(gv1.Graph.Edges), len(gv2.Graph.Edges))
	}
}

// Helper function for string containment checks
func statsContainsStr(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
