package query

import (
	"testing"

	"github.com/dnnrly/gsl-lang"
)

// TestTraverseOutDirection tests outgoing edge traversal
func TestTraverseOutDirection(t *testing.T) {
	// Graph: A → B → C → D
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"START": {}}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"D": {ID: "D", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "B", To: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "C", To: "D", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph: start from A (in START), traverse out 2 hops
	parser := NewQueryParser("subgraph in START traverse out 2")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should include A, B, C (2 hops out from A)
	if len(gv.Graph.GetNodes()) != 3 {
		t.Fatalf("Expected 3 nodes, got %d", len(gv.Graph.GetNodes()))
	}
	if !contains(gv.Graph.GetNodes(), "A", "B", "C") {
		t.Fatal("Should contain A, B, C")
	}

	// Should have edges A→B, B→C
	if len(gv.Graph.GetEdges()) != 2 {
		t.Fatalf("Expected 2 edges, got %d", len(gv.Graph.GetEdges()))
	}
}

// TestTraverseInDirection tests incoming edge traversal
func TestTraverseInDirection(t *testing.T) {
	// Graph: A → B → C
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"START": {}}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "B", To: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph: start from C (in START), traverse in 2 hops
	parser := NewQueryParser("subgraph in START traverse in 2")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should include C, B, A (2 hops in from C)
	if len(gv.Graph.GetNodes()) != 3 {
		t.Fatalf("Expected 3 nodes, got %d", len(gv.Graph.GetNodes()))
	}
	if !contains(gv.Graph.GetNodes(), "A", "B", "C") {
		t.Fatal("Should contain A, B, C")
	}

	// Should have edges A→B, B→C
	if len(gv.Graph.GetEdges()) != 2 {
		t.Fatalf("Expected 2 edges, got %d", len(gv.Graph.GetEdges()))
	}
}

// TestTraverseBothDirection tests bidirectional traversal
func TestTraverseBothDirection(t *testing.T) {
	// Graph: A → B ← C (B is center)
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"START": {}}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "C", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph: start from B (in START), traverse both 1 hop
	parser := NewQueryParser("subgraph in START traverse both 1")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should include A, B, C (both directions)
	if len(gv.Graph.GetNodes()) != 3 {
		t.Fatalf("Expected 3 nodes, got %d", len(gv.Graph.GetNodes()))
	}
	if !contains(gv.Graph.GetNodes(), "A", "B", "C") {
		t.Fatal("Should contain A, B, C")
	}

	// Should have both edges
	if len(gv.Graph.GetEdges()) != 2 {
		t.Fatalf("Expected 2 edges, got %d", len(gv.Graph.GetEdges()))
	}
}

// TestTraverseAll traverses unlimited depth
func TestTraverseAll(t *testing.T) {
	// Graph: A → B → C → D → E
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"START": {}}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"D": {ID: "D", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"E": {ID: "E", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "B", To: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "C", To: "D", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "D", To: "E", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph: start from A, traverse out all
	parser := NewQueryParser("subgraph in START traverse out all")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should include all nodes A-E
	if len(gv.Graph.GetNodes()) != 5 {
		t.Fatalf("Expected 5 nodes, got %d", len(gv.Graph.GetNodes()))
	}
	if !contains(gv.Graph.GetNodes(), "A", "B", "C", "D", "E") {
		t.Fatal("Should contain all nodes")
	}

	// Should have all edges
	if len(gv.Graph.GetEdges()) != 4 {
		t.Fatalf("Expected 4 edges, got %d", len(gv.Graph.GetEdges()))
	}
}

// TestTraverseCycles handles cyclic graphs correctly
func TestTraverseCycles(t *testing.T) {
	// Graph: A → B → C → A (cycle)
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"START": {}}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "B", To: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "C", To: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph: start from A, traverse out 3 hops
	parser := NewQueryParser("subgraph in START traverse out 3")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should include all nodes (visited set prevents revisiting)
	if len(gv.Graph.GetNodes()) != 3 {
		t.Fatalf("Expected 3 nodes, got %d", len(gv.Graph.GetNodes()))
	}

	// Should have all edges
	if len(gv.Graph.GetEdges()) != 3 {
		t.Fatalf("Expected 3 edges, got %d", len(gv.Graph.GetEdges()))
	}
}

// TestTraverseInvalidDirection tests error handling
func TestTraverseInvalidDirection(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"invalid direction", "subgraph exists traverse invalid 1", true},
		{"missing depth", "subgraph exists traverse out", true},
		{"invalid depth", "subgraph exists traverse out abc", true},
		{"negative depth", "subgraph exists traverse out -1", true},
		{"zero depth", "subgraph exists traverse out 0", true},
		{"valid all", "subgraph exists traverse out all", false},
		{"valid numeric", "subgraph exists traverse in 5", false},
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

// TestTraverseWithPredicate tests traversal combined with predicate filtering
func TestTraverseWithPredicate(t *testing.T) {
	// Graph: All marked A (critical), B, C (normal)
	// A → B → C
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {
				ID:         "A",
				Attributes: map[string]interface{}{},
				Sets:       map[string]struct{}{"CRITICAL": {}},
			},
			"B": {
				ID:         "B",
				Attributes: map[string]interface{}{},
				Sets:       map[string]struct{}{},
			},
			"C": {
				ID:         "C",
				Attributes: map[string]interface{}{},
				Sets:       map[string]struct{}{},
			},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "B", To: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph: start from CRITICAL nodes, traverse out 2
	parser := NewQueryParser("subgraph in CRITICAL traverse out 2")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should include A, B, C (starting from CRITICAL A, traverse 2 hops)
	if len(gv.Graph.GetNodes()) != 3 {
		t.Fatalf("Expected 3 nodes, got %d", len(gv.Graph.GetNodes()))
	}

	// Should have both edges
	if len(gv.Graph.GetEdges()) != 2 {
		t.Fatalf("Expected 2 edges, got %d", len(gv.Graph.GetEdges()))
	}
}

// TestTraversePreservesEdgeDuplicates tests that duplicate edges are included
func TestTraversePreservesEdgeDuplicates(t *testing.T) {
	// Graph with duplicate edge: A → B → C, and a duplicate A→B
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"START": {}}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}}, // duplicate
			{From: "B", To: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph: start from A, traverse out all
	parser := NewQueryParser("subgraph in START traverse out all")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should preserve duplicate edges
	if len(gv.Graph.GetEdges()) != 3 {
		t.Fatalf("Expected 3 edges (including duplicate), got %d", len(gv.Graph.GetEdges()))
	}
}

// TestTraverseNoEdgesOutOfSubgraph tests that edges don't extend outside subgraph
func TestTraverseNoEdgesOutOfSubgraph(t *testing.T) {
	// Graph: A → B → C, D (isolated)
	// Traverse from A but C has edge to D (outside traversal result)
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"START": {}}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"D": {ID: "D", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "B", To: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "C", To: "D", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph: start from A, traverse out 2 hops (should reach A, B, C, not D)
	parser := NewQueryParser("subgraph in START traverse out 2")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should have A, B, C (not D)
	if len(gv.Graph.GetNodes()) != 3 {
		t.Fatalf("Expected 3 nodes, got %d", len(gv.Graph.GetNodes()))
	}
	if gv.Graph.GetNodes()["D"] != nil {
		t.Fatal("Should not include D")
	}

	// Should only include edges between A, B, C (not C→D)
	if len(gv.Graph.GetEdges()) != 2 {
		t.Fatalf("Expected 2 edges, got %d", len(gv.Graph.GetEdges()))
	}
}

// TestTraverseSelfLoop tests traversal with self-loops
func TestTraverseSelfLoop(t *testing.T) {
	// Graph: A self-loop, A → B
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"START": {}}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}}, // self-loop
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph: start from A, traverse out 1
	parser := NewQueryParser("subgraph in START traverse out 1")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should include A, B
	if len(gv.Graph.GetNodes()) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(gv.Graph.GetNodes()))
	}

	// Should include self-loop and A→B
	if len(gv.Graph.GetEdges()) != 2 {
		t.Fatalf("Expected 2 edges, got %d", len(gv.Graph.GetEdges()))
	}
}

// TestTraverseWithoutTraversal tests subgraph without traversal still works
func TestTraverseWithoutTraversal(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Subgraph without traversal
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
	if len(gv.Graph.GetNodes()) != 2 || len(gv.Graph.GetEdges()) != 1 {
		t.Fatal("Subgraph without traversal should return base subgraph")
	}
}

// Helper function to check if nodes are present
func contains(nodes map[string]*gsl.Node, ids ...string) bool {
	for _, id := range ids {
		if nodes[id] == nil {
			return false
		}
	}
	return true
}

// TestTraverseUpDirection tests up traversal (follow DependsOn chain)
func TestTraverseUpDirection(t *testing.T) {
	// Graph: A → B (depends on E1: C → D)
	// Up from A or B should find C and D (parent edge's endpoints)
	parent := &gsl.Edge{From: "C", To: "D", Label: "E1"}
	child := &gsl.Edge{From: "A", To: "B", DependsOn: "E1"}
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"START": {}}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"D": {ID: "D", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{parent, child},
		Sets:  map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	parser := NewQueryParser("subgraph in START traverse up 1")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should include A, B, C, D (A from START, parent C-D via up)
	if !contains(gv.Graph.GetNodes(), "A", "B", "C", "D") {
		t.Fatal("Should contain A, B, C, D")
	}
}

// TestTraverseDownDirection tests down traversal (follow Children chain)
func TestTraverseDownDirection(t *testing.T) {
	// Graph: A → B (label E1), B → C (depends on E1)
	// Down from A or B should find C
	parent := &gsl.Edge{From: "A", To: "B", Label: "E1"}
	child := &gsl.Edge{From: "B", To: "C", DependsOn: "E1"}
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"START": {}}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{parent, child},
		Sets:  map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	parser := NewQueryParser("subgraph in START traverse down 1")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	// Should include A, B, C (A+B from START, C via down)
	if !contains(gv.Graph.GetNodes(), "A", "B", "C") {
		t.Fatal("Should contain A, B, C")
	}
}

// TestTraverseDownAll tests down traversal to unlimited depth
func TestTraverseDownAll(t *testing.T) {
	// Chain: A → B (E1) → C (depends E1) → D (E2) → E (depends E2)
	e1 := &gsl.Edge{From: "A", To: "B", Label: "E1"}
	c1 := &gsl.Edge{From: "B", To: "C", DependsOn: "E1", Label: "E2"}
	c2 := &gsl.Edge{From: "C", To: "D", DependsOn: "E2", Label: "E3"}
	c3 := &gsl.Edge{From: "D", To: "E", DependsOn: "E3"}
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"START": {}}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"D": {ID: "D", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"E": {ID: "E", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{e1, c1, c2, c3},
		Sets:  map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	parser := NewQueryParser("subgraph in START traverse down all")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	if !contains(gv.Graph.GetNodes(), "A", "B", "C", "D", "E") {
		t.Fatal("Should contain all nodes in the dependency chain")
	}
}

// TestTraverseCombinedOutUp tests combined out and up directions
func TestTraverseCombinedOutUp(t *testing.T) {
	// Graph: A → B (E1), C → D (depends E1)
	// Out from A reaches B; Up from B reaches C, D
	parent := &gsl.Edge{From: "C", To: "D", Label: "E1"}
	child := &gsl.Edge{From: "A", To: "B", DependsOn: "E1", Label: "E2"}
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"START": {}}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"D": {ID: "D", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{parent, child},
		Sets:  map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Parse with combined directions
	parser := NewQueryParser("subgraph in START traverse out up 1")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	if !contains(gv.Graph.GetNodes(), "A", "B", "C", "D") {
		t.Fatal("Should contain A, B (out from A) and C, D (up from B)")
	}
}

// TestSubgraphScope tests the scope keyword (= traverse down all)
func TestSubgraphScope(t *testing.T) {
	// Chain: A → B (E1) → C (depends E1) → D (depends... labeled, no further)
	e1 := &gsl.Edge{From: "A", To: "B", Label: "E1"}
	c1 := &gsl.Edge{From: "B", To: "C", DependsOn: "E1", Label: "E2"}
	c2 := &gsl.Edge{From: "C", To: "D", DependsOn: "E2"}
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"START": {}}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"D": {ID: "D", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{e1, c1, c2},
		Sets:  map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	parser := NewQueryParser("subgraph in START scope")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	if !contains(gv.Graph.GetNodes(), "A", "B", "C", "D") {
		t.Fatal("Scope should expand to all nodes in dependency chain")
	}
}

// TestTraverseInvalidDirections tests error handling for invalid direction combinations
func TestTraverseInvalidDirections(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"invalid direction", "subgraph exists traverse invalid 1", true},
		{"missing depth", "subgraph exists traverse out up", true},
		{"valid combined", "subgraph exists traverse out up 2", false},
		{"valid up only", "subgraph exists traverse up 1", false},
		{"valid down only", "subgraph exists traverse down 3", false},
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
