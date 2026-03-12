package query

import (
	"testing"
	"time"

	"github.com/dnnrly/gsl-lang"
)

func TestExecutionStatsBasic(t *testing.T) {
	// Test: basic stats collection on simple query
	input := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"api": {ID: "api", Attributes: map[string]interface{}{"type": "service"}, Sets: make(map[string]struct{})},
			"db":  {ID: "db", Attributes: map[string]interface{}{"type": "database"}, Sets: make(map[string]struct{})},
		},
		Edges: []*gsl.Edge{
			{From: "api", To: "db", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
		},
		Sets: make(map[string]*gsl.Set),
	}

	query, err := NewQueryParser(`subgraph node.type = "service"`).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	executor := NewQueryExecutor(query)
	_, stats, err := executor.ExecuteWithStats(ctx)
	if err != nil {
		t.Fatalf("ExecuteWithStats() error = %v", err)
	}

	// Input metrics
	if stats.InputNodeCount != 2 {
		t.Errorf("InputNodeCount = %d, want 2", stats.InputNodeCount)
	}
	if stats.InputEdgeCount != 1 {
		t.Errorf("InputEdgeCount = %d, want 1", stats.InputEdgeCount)
	}

	// Result metrics
	if stats.ResultNodeCount != 1 {
		t.Errorf("ResultNodeCount = %d, want 1", stats.ResultNodeCount)
	}
	if stats.ResultEdgeCount != 0 {
		t.Errorf("ResultEdgeCount = %d, want 0", stats.ResultEdgeCount)
	}

	// Changes
	if stats.NodesRemoved != 1 {
		t.Errorf("NodesRemoved = %d, want 1", stats.NodesRemoved)
	}
	if stats.EdgesRemoved != 1 {
		t.Errorf("EdgesRemoved = %d, want 1", stats.EdgesRemoved)
	}

	// Timing
	if stats.TotalDuration == 0 {
		t.Errorf("TotalDuration = 0, want > 0")
	}
	if len(stats.ExpressionDurations) != 1 {
		t.Errorf("len(ExpressionDurations) = %d, want 1", len(stats.ExpressionDurations))
	}
}

func TestExecutionStatsMultiExpression(t *testing.T) {
	// Test: stats tracking across multiple pipeline stages
	input := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"api":  {ID: "api", Attributes: map[string]interface{}{"type": "service", "critical": true}, Sets: make(map[string]struct{})},
			"db":   {ID: "db", Attributes: map[string]interface{}{"type": "database"}, Sets: make(map[string]struct{})},
			"cache": {ID: "cache", Attributes: map[string]interface{}{"type": "cache"}, Sets: make(map[string]struct{})},
		},
		Edges: []*gsl.Edge{
			{From: "api", To: "db", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
		},
		Sets: make(map[string]*gsl.Set),
	}

	// Pipeline: subgraph node.type = "service" | make node.priority = "high" where exists
	query, err := NewQueryParser(`subgraph node.type = "service" | make node.priority = "high" where exists`).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	executor := NewQueryExecutor(query)
	_, stats, err := executor.ExecuteWithStats(ctx)
	if err != nil {
		t.Fatalf("ExecuteWithStats() error = %v", err)
	}

	// Should have 2 expressions
	if len(stats.ExpressionDurations) != 2 {
		t.Errorf("len(ExpressionDurations) = %d, want 2", len(stats.ExpressionDurations))
	}

	// First expression: SubgraphExpr
	if stats.ExpressionDurations[0].ExprType != "SubgraphExpr" {
		t.Errorf("ExprType[0] = %s, want SubgraphExpr", stats.ExpressionDurations[0].ExprType)
	}
	if stats.ExpressionDurations[0].OutputNodeCount != 1 {
		t.Errorf("ExpressionDurations[0].OutputNodeCount = %d, want 1", stats.ExpressionDurations[0].OutputNodeCount)
	}

	// Second expression: MakeExpr
	if stats.ExpressionDurations[1].ExprType != "MakeExpr" {
		t.Errorf("ExprType[1] = %s, want MakeExpr", stats.ExpressionDurations[1].ExprType)
	}

	// Result should have 1 node with priority attribute
	if stats.ResultNodeCount != 1 {
		t.Errorf("ResultNodeCount = %d, want 1", stats.ResultNodeCount)
	}
}

func TestExecutionStatsNodesAdded(t *testing.T) {
	// Test: tracking nodes added by collapse operation
	input := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"a": {ID: "a", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
			"b": {ID: "b", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Collapse both nodes into one (net: -1 node)
	query, err := NewQueryParser(`collapse into MERGED where exists`).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	executor := NewQueryExecutor(query)
	_, stats, err := executor.ExecuteWithStats(ctx)
	if err != nil {
		t.Fatalf("ExecuteWithStats() error = %v", err)
	}

	// Started with 2 nodes, collapsed to 1 (net: -1 node, so 1 removed from input count)
	if stats.NodesRemoved != 1 {
		t.Errorf("NodesRemoved = %d, want 1", stats.NodesRemoved)
	}
	if stats.ResultNodeCount != 1 {
		t.Errorf("ResultNodeCount = %d, want 1", stats.ResultNodeCount)
	}
}

func TestExpressionTimingInfo(t *testing.T) {
	// Test: ExpressionTiming contains correct information
	timing := ExpressionTiming{
		Index:           0,
		ExprType:        "SubgraphExpr",
		Duration:        100 * time.Millisecond,
		InputNodeCount:  10,
		OutputNodeCount: 5,
		InputEdgeCount:  20,
		OutputEdgeCount: 8,
	}

	str := timing.String()
	if !statsContainsAll(str, "SubgraphExpr", "100ms", "10", "5", "20", "8") {
		t.Errorf("String() = %s, missing expected values", str)
	}
}

func TestStatsSummary(t *testing.T) {
	// Test: summary generation
	stats := &ExecutionStats{
		ResultNodeCount: 5,
		ResultEdgeCount: 3,
		ResultSetCount:  1,
		InputNodeCount:  10,
		InputEdgeCount:  5,
		InputSetCount:   1,
		NodesAdded:      0,
		NodesRemoved:    5,
		EdgesAdded:      0,
		EdgesRemoved:    2,
		TotalDuration:   50 * time.Millisecond,
	}

	summary := stats.Summary()
	if !statsContainsAll(summary, "5 nodes", "3 edges", "1 sets", "50ms") {
		t.Errorf("Summary() = %s, missing expected values", summary)
	}
}

func TestStatsString(t *testing.T) {
	// Test: full stats string representation
	stats := &ExecutionStats{
		ResultNodeCount: 1,
		ResultEdgeCount: 0,
		ResultSetCount:  0,
		InputNodeCount:  2,
		InputEdgeCount:  1,
		InputSetCount:   0,
		NodesRemoved:    1,
		EdgesRemoved:    1,
		TotalDuration:   10 * time.Millisecond,
		ExpressionDurations: []ExpressionTiming{
			{
				Index:           0,
				ExprType:        "SubgraphExpr",
				Duration:        10 * time.Millisecond,
				InputNodeCount:  2,
				OutputNodeCount: 1,
				InputEdgeCount:  1,
				OutputEdgeCount: 0,
			},
		},
	}

	str := stats.String()
	if !statsContainsAll(str, "Query Execution Statistics", "1 nodes", "SubgraphExpr") {
		t.Errorf("String() missing expected content: %s", str)
	}
}

func TestAverageExpressionTime(t *testing.T) {
	// Test: average timing calculation
	stats := &ExecutionStats{
		ExpressionDurations: []ExpressionTiming{
			{Duration: 10 * time.Millisecond},
			{Duration: 20 * time.Millisecond},
			{Duration: 30 * time.Millisecond},
		},
	}

	avg := stats.AverageExpressionTime()
	expected := 20 * time.Millisecond
	if avg != expected {
		t.Errorf("AverageExpressionTime() = %v, want %v", avg, expected)
	}
}

func TestAverageExpressionTimeEmpty(t *testing.T) {
	// Test: average on empty stats
	stats := &ExecutionStats{
		ExpressionDurations: []ExpressionTiming{},
	}

	avg := stats.AverageExpressionTime()
	if avg != 0 {
		t.Errorf("AverageExpressionTime() = %v, want 0", avg)
	}
}

func TestSlowestExpression(t *testing.T) {
	// Test: finding slowest expression
	stats := &ExecutionStats{
		ExpressionDurations: []ExpressionTiming{
			{Index: 0, ExprType: "FromExpr", Duration: 10 * time.Millisecond},
			{Index: 1, ExprType: "SubgraphExpr", Duration: 30 * time.Millisecond},
			{Index: 2, ExprType: "MakeExpr", Duration: 20 * time.Millisecond},
		},
	}

	slowest := stats.SlowestExpression()
	if slowest == nil || slowest.ExprType != "SubgraphExpr" || slowest.Duration != 30*time.Millisecond {
		t.Errorf("SlowestExpression() = %v, want SubgraphExpr with 30ms", slowest)
	}
}

func TestFastestExpression(t *testing.T) {
	// Test: finding fastest expression
	stats := &ExecutionStats{
		ExpressionDurations: []ExpressionTiming{
			{Index: 0, ExprType: "FromExpr", Duration: 10 * time.Millisecond},
			{Index: 1, ExprType: "SubgraphExpr", Duration: 30 * time.Millisecond},
			{Index: 2, ExprType: "MakeExpr", Duration: 20 * time.Millisecond},
		},
	}

	fastest := stats.FastestExpression()
	if fastest == nil || fastest.ExprType != "FromExpr" || fastest.Duration != 10*time.Millisecond {
		t.Errorf("FastestExpression() = %v, want FromExpr with 10ms", fastest)
	}
}

func TestPerExpressionStats(t *testing.T) {
	// Test: per-expression stats breakdown
	stats := &ExecutionStats{
		ExpressionDurations: []ExpressionTiming{
			{Index: 0, ExprType: "SubgraphExpr", Duration: 10 * time.Millisecond, InputNodeCount: 10, OutputNodeCount: 5},
			{Index: 1, ExprType: "MakeExpr", Duration: 5 * time.Millisecond, InputNodeCount: 5, OutputNodeCount: 5},
		},
	}

	perExpr := stats.PerExpressionStats()
	if len(perExpr) != 2 {
		t.Errorf("len(PerExpressionStats()) = %d, want 2", len(perExpr))
	}

	if !statsContainsAll(perExpr[0], "SubgraphExpr", "10ms", "10", "5") {
		t.Errorf("PerExpressionStats()[0] = %s, missing expected values", perExpr[0])
	}
}

func TestExecutionStatsComplexPipeline(t *testing.T) {
	// Test: stats on complex multi-step pipeline
	input := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"a": {ID: "a", Attributes: map[string]interface{}{"type": "service"}, Sets: make(map[string]struct{})},
			"b": {ID: "b", Attributes: map[string]interface{}{"type": "database"}, Sets: make(map[string]struct{})},
			"c": {ID: "c", Attributes: map[string]interface{}{"type": "service"}, Sets: make(map[string]struct{})},
		},
		Edges: []*gsl.Edge{
			{From: "a", To: "b", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
			{From: "b", To: "c", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
		},
		Sets: make(map[string]*gsl.Set),
	}

	// Complex pipeline: bind, filter, collapse, mark
	query, err := NewQueryParser(
		`(subgraph node.type = "service") as SERVICES | ` +
		`(from * | subgraph node.type = "database") as DATABASES | ` +
		`SERVICES + DATABASES | ` +
		`make node.processed = true where exists`,
	).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	executor := NewQueryExecutor(query)
	_, stats, err := executor.ExecuteWithStats(ctx)
	if err != nil {
		t.Fatalf("ExecuteWithStats() error = %v", err)
	}

	// Should track all expressions
	if len(stats.ExpressionDurations) < 4 {
		t.Errorf("len(ExpressionDurations) = %d, want >= 4", len(stats.ExpressionDurations))
	}

	// Final result should have all 3 nodes
	if stats.ResultNodeCount != 3 {
		t.Errorf("ResultNodeCount = %d, want 3", stats.ResultNodeCount)
	}
}

func TestEmptyGraphStats(t *testing.T) {
	// Test: stats on empty graph
	input := &gsl.Graph{
		Nodes: make(map[string]*gsl.Node),
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	query, err := NewQueryParser(`subgraph exists`).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	executor := NewQueryExecutor(query)
	_, stats, err := executor.ExecuteWithStats(ctx)
	if err != nil {
		t.Fatalf("ExecuteWithStats() error = %v", err)
	}

	if stats.ResultNodeCount != 0 {
		t.Errorf("ResultNodeCount = %d, want 0", stats.ResultNodeCount)
	}
	if stats.ResultEdgeCount != 0 {
		t.Errorf("ResultEdgeCount = %d, want 0", stats.ResultEdgeCount)
	}
}

// Helper function to check if string contains all substrings
func statsContainsAll(str string, substrings ...string) bool {
	for _, substr := range substrings {
		if !statsContains(str, substr) {
			return false
		}
	}
	return true
}

func statsContains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
