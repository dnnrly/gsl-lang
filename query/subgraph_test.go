package query

import (
	"testing"

	"github.com/dnnrly/gsl-lang"
)

// TestSubgraphNodePredicate tests node-targeted subgraph extraction
func TestSubgraphNodePredicate(t *testing.T) {
	// Create a graph with multiple nodes
	graph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"type": "service"}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{"type": "db"}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{"type": "service"}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "B", To: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "A", To: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph: select services (A, C)
	parser := NewQueryParser("subgraph node.type = service")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should have A and C
	if len(gv.Graph.Nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(gv.Graph.Nodes))
	}
	if _, ok := gv.Graph.Nodes["A"]; !ok {
		t.Fatal("Node A should be included")
	}
	if _, ok := gv.Graph.Nodes["C"]; !ok {
		t.Fatal("Node C should be included")
	}
	if _, ok := gv.Graph.Nodes["B"]; ok {
		t.Fatal("Node B should not be included")
	}

	// Should only include edge A→C (both ends in result)
	if len(gv.Graph.Edges) != 1 {
		t.Fatalf("Expected 1 edge, got %d", len(gv.Graph.Edges))
	}
	if gv.Graph.Edges[0].From != "A" || gv.Graph.Edges[0].To != "C" {
		t.Fatalf("Expected edge A→C, got %s→%s", gv.Graph.Edges[0].From, gv.Graph.Edges[0].To)
	}
}

// TestSubgraphEdgePredicate tests edge-targeted subgraph extraction
func TestSubgraphEdgePredicate(t *testing.T) {
	graph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{"protocol": "http"}, Sets: map[string]struct{}{}},
			{From: "B", To: "C", Attributes: map[string]interface{}{"protocol": "grpc"}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph: select http edges
	parser := NewQueryParser("subgraph edge.protocol = http")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should include A and B (endpoints of http edge), not C
	if len(gv.Graph.Nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(gv.Graph.Nodes))
	}
	if _, ok := gv.Graph.Nodes["A"]; !ok {
		t.Fatal("Node A should be included")
	}
	if _, ok := gv.Graph.Nodes["B"]; !ok {
		t.Fatal("Node B should be included")
	}
	if _, ok := gv.Graph.Nodes["C"]; ok {
		t.Fatal("Node C should not be included")
	}

	// Should only include http edge
	if len(gv.Graph.Edges) != 1 {
		t.Fatalf("Expected 1 edge, got %d", len(gv.Graph.Edges))
	}
	if gv.Graph.Edges[0].From != "A" || gv.Graph.Edges[0].To != "B" {
		t.Fatalf("Expected edge A→B, got %s→%s", gv.Graph.Edges[0].From, gv.Graph.Edges[0].To)
	}
}

// TestSubgraphSetMembership tests set-based node filtering
func TestSubgraphSetMembership(t *testing.T) {
	graph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"CRITICAL": {}}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"CRITICAL": {}}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "B", To: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "A", To: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph: select nodes in CRITICAL set
	parser := NewQueryParser("subgraph in CRITICAL")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should have A and C
	if len(gv.Graph.Nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(gv.Graph.Nodes))
	}

	// Should include A→C edge
	if len(gv.Graph.Edges) != 1 {
		t.Fatalf("Expected 1 edge, got %d", len(gv.Graph.Edges))
	}
}

// TestSubgraphExistsPredicateAllElements tests exists predicate (match all)
func TestSubgraphExistsPredicateAllElements(t *testing.T) {
	graph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph with exists (match all)
	parser := NewQueryParser("subgraph exists")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should include everything
	if len(gv.Graph.Nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(gv.Graph.Nodes))
	}
	if len(gv.Graph.Edges) != 1 {
		t.Fatalf("Expected 1 edge, got %d", len(gv.Graph.Edges))
	}
}

// TestSubgraphPreserveDuplicateEdges tests that duplicate edges are preserved
func TestSubgraphPreserveDuplicateEdges(t *testing.T) {
	graph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	parser := NewQueryParser("subgraph exists")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should preserve both edges
	if len(gv.Graph.Edges) != 2 {
		t.Fatalf("Expected 2 edges, got %d", len(gv.Graph.Edges))
	}
}

// TestSubgraphEmptyResult tests subgraph with no matches
func TestSubgraphEmptyResult(t *testing.T) {
	graph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"type": "service"}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph: select databases (none exist)
	parser := NewQueryParser("subgraph node.type = db")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should have empty result
	if len(gv.Graph.Nodes) != 0 {
		t.Fatalf("Expected 0 nodes, got %d", len(gv.Graph.Nodes))
	}
	if len(gv.Graph.Edges) != 0 {
		t.Fatalf("Expected 0 edges, got %d", len(gv.Graph.Edges))
	}
}

// TestSubgraphAndPredicate tests AND combination in subgraph
func TestSubgraphAndPredicate(t *testing.T) {
	graph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {
				ID:         "A",
				Attributes: map[string]interface{}{"type": "service", "zone": "us"},
				Sets:       map[string]struct{}{"CRITICAL": {}},
			},
			"B": {
				ID:         "B",
				Attributes: map[string]interface{}{"type": "service", "zone": "eu"},
				Sets:       map[string]struct{}{},
			},
			"C": {
				ID:         "C",
				Attributes: map[string]interface{}{"type": "db", "zone": "us"},
				Sets:       map[string]struct{}{"CRITICAL": {}},
			},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph: select CRITICAL services in us zone
	parser := NewQueryParser("subgraph in CRITICAL AND node.zone = us")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should include A and C (both critical, both us zone)
	if len(gv.Graph.Nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(gv.Graph.Nodes))
	}
	if _, ok := gv.Graph.Nodes["A"]; !ok {
		t.Fatal("Node A should be included")
	}
	if _, ok := gv.Graph.Nodes["C"]; !ok {
		t.Fatal("Node C should be included")
	}

	// Should include A→C edge
	if len(gv.Graph.Edges) != 1 {
		t.Fatalf("Expected 1 edge, got %d", len(gv.Graph.Edges))
	}
}

// TestSubgraphNotPredicate tests NOT prefix in subgraph
func TestSubgraphNotPredicate(t *testing.T) {
	graph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"DEPRECATED": {}}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"DEPRECATED": {}}},
		},
		Edges: []*gsl.Edge{
			{From: "B", To: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph: select non-deprecated nodes
	parser := NewQueryParser("subgraph not in DEPRECATED")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should include only B
	if len(gv.Graph.Nodes) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(gv.Graph.Nodes))
	}
	if _, ok := gv.Graph.Nodes["B"]; !ok {
		t.Fatal("Node B should be included")
	}

	// No edges (B has no edges to other non-deprecated nodes)
	if len(gv.Graph.Edges) != 0 {
		t.Fatalf("Expected 0 edges, got %d", len(gv.Graph.Edges))
	}
}

// TestSubgraphInPipeline tests subgraph as part of a pipeline
func TestSubgraphInPipeline(t *testing.T) {
	graph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"type": "service"}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{"type": "db"}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Pipeline: extract services, then bind to SERVICES
	// Note: (pipeline) as NAME executes pipeline relative to current input
	parser := NewQueryParser("subgraph node.type = service | (from *) as SERVICES")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	// Result should be the subgraph (services only)
	gv := result.(GraphValue)
	if len(gv.Graph.Nodes) != 1 {
		t.Fatalf("Expected 1 node in result, got %d", len(gv.Graph.Nodes))
	}
	if _, ok := gv.Graph.Nodes["A"]; !ok {
		t.Fatal("Node A should be in result")
	}

	// Named graph should be bound (contains input graph, because (from *) resets to input)
	if _, ok := ctx.NamedGraphs["SERVICES"]; !ok {
		t.Fatal("SERVICES should be bound")
	}
	// SERVICES contains the input graph (from * resets to InputGraph)
	if len(ctx.NamedGraphs["SERVICES"].Nodes) != 2 {
		t.Fatalf("SERVICES should contain input graph (2 nodes), got %d", len(ctx.NamedGraphs["SERVICES"].Nodes))
	}
}

// TestSubgraphInvalidSyntax tests parsing errors
func TestSubgraphInvalidSyntax(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"subgraph alone", "subgraph", true},
		{"invalid predicate", "subgraph invalid_pred", true},
		{"no predicate", "subgraph ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewQueryParser(tt.input)
			_, err := parser.Parse()
			if !tt.wantError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if tt.wantError && err == nil {
				t.Fatal("Expected error but got none")
			}
		})
	}
}

// TestSubgraphPreservesSets tests that sets are preserved in subgraph
func TestSubgraphPreservesSets(t *testing.T) {
	graph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets: map[string]*gsl.Set{
			"CUSTOM": {ID: "CUSTOM", Attributes: map[string]interface{}{"color": "blue"}},
		},
	}

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	parser := NewQueryParser("subgraph exists")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Sets should be preserved
	if len(gv.Graph.Sets) != 1 {
		t.Fatalf("Expected 1 set, got %d", len(gv.Graph.Sets))
	}
	if _, ok := gv.Graph.Sets["CUSTOM"]; !ok {
		t.Fatal("CUSTOM set should be preserved")
	}
}

// TestSubgraphMixedTargetsError tests error when predicate mixes targets
func TestSubgraphMixedTargetsError(t *testing.T) {
	graph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Create a predicate that mixes node and edge targets
	parser := NewQueryParser("subgraph node.attr = val AND edge.attr = val2")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	_, err = query.Execute(ctx)
	if err == nil {
		t.Fatal("Expected error for mixed targets")
	}
	if err.Error() != "predicate mixes node and edge targets" {
		t.Fatalf("Wrong error message: %v", err)
	}
}

// TestSubgraphDirectlyOnExpr tests SubgraphExpr.Apply() directly
func TestSubgraphDirectlyOnExpr(t *testing.T) {
	graph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"X": {ID: "X", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"Y": {ID: "Y", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "X", To: "Y", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	pred := &ExistsPredicate{}
	expr := &SubgraphExpr{Pred: pred}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.Nodes) != 2 || len(gv.Graph.Edges) != 1 {
		t.Fatal("Should preserve all nodes and edges with exists predicate")
	}
}

// TestSubgraphNodeIsolation tests nodes with no edges to matched nodes
func TestSubgraphNodeIsolation(t *testing.T) {
	graph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"tier": "web"}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{"tier": "api"}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{"tier": "web"}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Select web tier (A, C) - isolated nodes should be included but without edges between them
	parser := NewQueryParser("subgraph node.tier = web")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should include A and C (no edge between them)
	if len(gv.Graph.Nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(gv.Graph.Nodes))
	}
	if len(gv.Graph.Edges) != 0 {
		t.Fatalf("Expected 0 edges, got %d", len(gv.Graph.Edges))
	}
}
