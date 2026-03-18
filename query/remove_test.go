package query

import (
	"testing"

	"github.com/dnnrly/gsl-lang"
)

// TestRemoveEdgeBasic tests basic edge removal
func TestRemoveEdgeBasic(t *testing.T) {
	// Create simple graph: A -> B, A -> C
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}},
			{From: "A", To: "C", Attributes: map[string]interface{}{}},
		},
		Sets: make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	// Remove all edges with exists predicate
	expr := &RemoveEdgeExpr{
		Pred: &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.GetEdges()) != 0 {
		t.Errorf("Expected 0 edges, got %d", len(gv.Graph.GetEdges()))
	}
	if len(gv.Graph.GetNodes()) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(gv.Graph.GetNodes()))
	}
}

// TestRemoveEdgeWithPredicate tests edge removal with attribute predicate
func TestRemoveEdgeWithPredicate(t *testing.T) {
	// Create graph with labeled edges
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{"type": "http"}},
			{From: "A", To: "C", Attributes: map[string]interface{}{"type": "grpc"}},
		},
		Sets: make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	// Remove edges with type=http
	expr := &RemoveEdgeExpr{
		Pred: &AttributeEqualsPredicate{
			Target: "edge",
			Name:   "type",
			Value:  "http",
		},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.GetEdges()) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(gv.Graph.GetEdges()))
	}
	if gv.Graph.GetEdges()[0].To != "C" {
		t.Errorf("Expected remaining edge to C, got %s", gv.Graph.GetEdges()[0].To)
	}
}

// TestRemoveEdgePreservesNodesAndSets tests that nodes and sets are preserved
func TestRemoveEdgePreservesNodesAndSets(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"foo": "bar"}},
			"B": {ID: "B", Attributes: map[string]interface{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}},
		},
		Sets: map[string]*gsl.Set{
			"CRITICAL": {ID: "CRITICAL", Attributes: map[string]interface{}{}},
		},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	expr := &RemoveEdgeExpr{Pred: &ExistsPredicate{}}
	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.GetNodes()) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(gv.Graph.GetNodes()))
	}
	if _, hasSet := gv.Graph.GetSets()["CRITICAL"]; !hasSet {
		t.Error("Expected set CRITICAL to be preserved")
	}
	if gv.Graph.GetNodes()["A"].Attributes["foo"] != "bar" {
		t.Error("Expected node attributes to be preserved")
	}
}

// TestRemoveOrphansBasic tests orphan removal
func TestRemoveOrphansBasic(t *testing.T) {
	// A -> B, C is orphan
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}},
		},
		Sets: make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	expr := &RemoveOrphansExpr{}
	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.GetNodes()) != 2 {
		t.Errorf("Expected 2 nodes (A, B), got %d", len(gv.Graph.GetNodes()))
	}
	if _, hasC := gv.Graph.GetNodes()["C"]; hasC {
		t.Error("Expected orphan C to be removed")
	}
	if len(gv.Graph.GetEdges()) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(gv.Graph.GetEdges()))
	}
}

// TestRemoveOrphansWithSelfLoop tests that self-loops prevent orphan removal
func TestRemoveOrphansWithSelfLoop(t *testing.T) {
	// A has self-loop, B is true orphan
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "A", Attributes: map[string]interface{}{}},
		},
		Sets: make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	expr := &RemoveOrphansExpr{}
	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.GetNodes()) != 1 {
		t.Errorf("Expected 1 node (A with self-loop), got %d", len(gv.Graph.GetNodes()))
	}
	if _, hasA := gv.Graph.GetNodes()["A"]; !hasA {
		t.Error("Expected node A with self-loop to remain")
	}
	if _, hasB := gv.Graph.GetNodes()["B"]; hasB {
		t.Error("Expected orphan B to be removed")
	}
}

// TestRemoveOrphansEmpty tests remove orphans on empty graph
func TestRemoveOrphansEmpty(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	expr := &RemoveOrphansExpr{}
	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.GetNodes()) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(gv.Graph.GetNodes()))
	}
}

// TestRemoveAttributeBasic tests basic attribute removal
func TestRemoveAttributeBasic(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"team": "payments", "owner": "alice"}},
			"B": {ID: "B", Attributes: map[string]interface{}{"team": "fraud", "owner": "bob"}},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	// Remove owner attribute from all nodes
	expr := &RemoveAttributeExpr{
		Target: "node",
		Attr:   "owner",
		Pred:   &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	for _, node := range gv.Graph.GetNodes() {
		if _, hasOwner := node.Attributes["owner"]; hasOwner {
			t.Error("Expected owner attribute to be removed")
		}
		if _, hasTeam := node.Attributes["team"]; !hasTeam {
			t.Error("Expected team attribute to remain")
		}
	}
}

// TestRemoveAttributeWithPredicate tests attribute removal with predicate
func TestRemoveAttributeWithPredicate(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"team": "payments", "owner": "alice"}},
			"B": {ID: "B", Attributes: map[string]interface{}{"team": "fraud", "owner": "bob"}},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	// Remove owner from team=payments nodes only
	expr := &RemoveAttributeExpr{
		Target: "node",
		Attr:   "owner",
		Pred: &AttributeEqualsPredicate{
			Target: "node",
			Name:   "team",
			Value:  "payments",
		},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	nodeA := gv.Graph.GetNodes()["A"]
	nodeB := gv.Graph.GetNodes()["B"]

	if _, hasOwner := nodeA.Attributes["owner"]; hasOwner {
		t.Error("Expected owner removed from A")
	}
	if owner, hasOwner := nodeB.Attributes["owner"]; !hasOwner || owner != "bob" {
		t.Error("Expected owner preserved on B")
	}
}

// TestRemoveAttributeEdges tests edge attribute removal
func TestRemoveAttributeEdges(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{"type": "http", "timeout": 30}},
		},
		Sets: make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	// Remove timeout from all edges
	expr := &RemoveAttributeExpr{
		Target: "edge",
		Attr:   "timeout",
		Pred:   &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	edge := gv.Graph.GetEdges()[0]

	if _, hasTimeout := edge.Attributes["timeout"]; hasTimeout {
		t.Error("Expected timeout to be removed")
	}
	if edgeType, hasType := edge.Attributes["type"]; !hasType || edgeType != "http" {
		t.Error("Expected type to be preserved")
	}
}

// TestRemoveExprParser tests parsing of remove expressions
func TestRemoveExprParser(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
		exprType string
	}{
		{"remove orphans", false, "RemoveOrphans"},
		{"remove edge where exists", false, "RemoveEdge"},
		{"remove node.owner where exists", false, "RemoveAttribute"},
		{"remove edge.timeout where exists", false, "RemoveAttribute"},
		{"remove", true, ""},
		{"remove edge", true, ""},
		{"remove node.attr", true, ""},
	}

	for _, tt := range tests {
		p := newExpressionParser(tt.input)
		expr, err := p.parse()
		
		if (err != nil) != tt.wantErr {
			t.Errorf("parse(%q): wantErr=%v, got=%v", tt.input, tt.wantErr, err)
			continue
		}
		
		if err == nil {
			switch tt.exprType {
			case "RemoveOrphans":
				if _, ok := expr.(*RemoveOrphansExpr); !ok {
					t.Errorf("parse(%q): expected RemoveOrphansExpr, got %T", tt.input, expr)
				}
			case "RemoveEdge":
				if _, ok := expr.(*RemoveEdgeExpr); !ok {
					t.Errorf("parse(%q): expected RemoveEdgeExpr, got %T", tt.input, expr)
				}
			case "RemoveAttribute":
				if _, ok := expr.(*RemoveAttributeExpr); !ok {
					t.Errorf("parse(%q): expected RemoveAttributeExpr, got %T", tt.input, expr)
				}
			}
		}
	}
}

// TestRemoveEdgeInPipeline tests remove edge in query pipeline
func TestRemoveEdgeInPipeline(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{"type": "http"}},
			{From: "B", To: "C", Attributes: map[string]interface{}{"type": "grpc"}},
		},
		Sets: make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	query, err := NewQueryParser("remove edge where edge.type = \"http\"").Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.GetEdges()) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(gv.Graph.GetEdges()))
	}
	if gv.Graph.GetEdges()[0].To != "C" {
		t.Errorf("Expected B->C, got %s->%s", gv.Graph.GetEdges()[0].From, gv.Graph.GetEdges()[0].To)
	}
}

// TestRemoveOrphansInPipeline tests remove orphans in query pipeline
func TestRemoveOrphansInPipeline(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}},
		},
		Sets: make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	query, err := NewQueryParser("remove orphans").Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.GetNodes()) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(gv.Graph.GetNodes()))
	}
}

// TestRemoveEdgeChained tests remove edge chained with other operations
func TestRemoveEdgeChained(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"team": "payments"}},
			"B": {ID: "B", Attributes: map[string]interface{}{"team": "fraud"}},
			"C": {ID: "C", Attributes: map[string]interface{}{"team": "payments"}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}},
			{From: "B", To: "C", Attributes: map[string]interface{}{}},
		},
		Sets: make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	// First remove orphans, then extract subgraph
	query, err := NewQueryParser("remove orphans | subgraph node.team = \"payments\"").Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.GetNodes()) != 2 {
		t.Errorf("Expected 2 nodes (A, C), got %d", len(gv.Graph.GetNodes()))
	}
}

// TestRemoveAttributeEdgeCases tests edge cases for attribute removal
func TestRemoveAttributeEdgeCases(t *testing.T) {
	// Graph where attribute doesn't exist
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"team": "payments"}},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	// Try to remove non-existent attribute
	expr := &RemoveAttributeExpr{
		Target: "node",
		Attr:   "nonexistent",
		Pred:   &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if gv.Graph.GetNodes()["A"].Attributes["team"] != "payments" {
		t.Error("Expected team attribute to remain")
	}
}

// TestRemoveEdgeEmptyGraph tests remove edge on empty graph
func TestRemoveEdgeEmptyGraph(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	expr := &RemoveEdgeExpr{Pred: &ExistsPredicate{}}
	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.GetEdges()) != 0 || len(gv.Graph.GetNodes()) != 0 {
		t.Error("Expected empty graph to remain empty")
	}
}

// TestRemoveOrphansPreservesSets tests that remove orphans preserves sets
func TestRemoveOrphansPreservesSets(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}},
		},
		Sets: map[string]*gsl.Set{
			"CRITICAL": {ID: "CRITICAL", Attributes: map[string]interface{}{}},
		},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	expr := &RemoveOrphansExpr{}
	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if _, hasSet := gv.Graph.GetSets()["CRITICAL"]; !hasSet {
		t.Error("Expected set to be preserved")
	}
}
