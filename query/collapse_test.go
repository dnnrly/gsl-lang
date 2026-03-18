package query

import (
	"testing"

	"github.com/dnnrly/gsl-lang"
)

// TestCollapseBasic tests basic collapse operation
func TestCollapseBasic(t *testing.T) {
	// Graph: A -> B, C -> B, B -> D
	// Collapse A and C into SERVICES
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"type": "service"}},
			"B": {ID: "B", Attributes: map[string]interface{}{"type": "db"}},
			"C": {ID: "C", Attributes: map[string]interface{}{"type": "service"}},
			"D": {ID: "D", Attributes: map[string]interface{}{"type": "cache"}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}},
			{From: "C", To: "B", Attributes: map[string]interface{}{}},
			{From: "B", To: "D", Attributes: map[string]interface{}{}},
		},
		Sets: make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	expr := &CollapseExpr{
		NodeID: "SERVICES",
		Pred: &AttributeEqualsPredicate{
			Target: "node",
			Name:   "type",
			Value:  "service",
		},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)

	// Check nodes: SERVICES, B, D (A, C removed)
	if len(gv.Graph.GetNodes()) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(gv.Graph.GetNodes()))
	}
	if _, hasServices := gv.Graph.GetNodes()["SERVICES"]; !hasServices {
		t.Error("Expected SERVICES node")
	}
	if _, hasA := gv.Graph.GetNodes()["A"]; hasA {
		t.Error("Expected A to be removed")
	}
	if _, hasC := gv.Graph.GetNodes()["C"]; hasC {
		t.Error("Expected C to be removed")
	}

	// Check edges: SERVICES -> B, B -> D
	if len(gv.Graph.GetEdges()) != 2 {
		t.Errorf("Expected 2 edges, got %d", len(gv.Graph.GetEdges()))
	}
}

// TestCollapseAttributeMerge tests attribute merging during collapse
func TestCollapseAttributeMerge(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"team": "payments", "owner": "alice"}},
			"B": {ID: "B", Attributes: map[string]interface{}{"team": "fraud", "env": "prod"}},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	expr := &CollapseExpr{
		NodeID: "MERGED",
		Pred:   &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	merged := gv.Graph.GetNodes()["MERGED"]

	// Attributes should be merged with last-write-wins
	// Since we sort deterministically, we get: A, B
	// A: team=payments, owner=alice
	// B: team=fraud, env=prod
	// Result: team=fraud (B overwrites), owner=alice, env=prod

	if merged.Attributes["team"] != "fraud" {
		t.Errorf("Expected team=fraud, got %v", merged.Attributes["team"])
	}
	if merged.Attributes["owner"] != "alice" {
		t.Errorf("Expected owner=alice, got %v", merged.Attributes["owner"])
	}
	if merged.Attributes["env"] != "prod" {
		t.Errorf("Expected env=prod, got %v", merged.Attributes["env"])
	}
}

// TestCollapseRemovesInternalEdges tests that edges between collapsed nodes are removed
func TestCollapseRemovesInternalEdges(t *testing.T) {
	// Graph: A -> B, B -> C, A -> C (A and B will be collapsed)
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"type": "service"}},
			"B": {ID: "B", Attributes: map[string]interface{}{"type": "service"}},
			"C": {ID: "C", Attributes: map[string]interface{}{"type": "db"}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}},
			{From: "B", To: "C", Attributes: map[string]interface{}{}},
			{From: "A", To: "C", Attributes: map[string]interface{}{}},
		},
		Sets: make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	expr := &CollapseExpr{
		NodeID: "SERVICES",
		Pred: &AttributeEqualsPredicate{
			Target: "node",
			Name:   "type",
			Value:  "service",
		},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)

	// Should have only 2 edges: SERVICES -> C (A->B is internal, A->C and B->C both become SERVICES->C)
	// After dedup, just SERVICES -> C
	if len(gv.Graph.GetEdges()) != 1 {
		t.Errorf("Expected 1 edge after dedup, got %d", len(gv.Graph.GetEdges()))
	}
	if gv.Graph.GetEdges()[0].From != "SERVICES" || gv.Graph.GetEdges()[0].To != "C" {
		t.Errorf("Expected SERVICES->C, got %s->%s", gv.Graph.GetEdges()[0].From, gv.Graph.GetEdges()[0].To)
	}
}

// TestCollapseEdgeDeduplication tests edge deduplication during collapse
func TestCollapseEdgeDeduplication(t *testing.T) {
	// Graph: A -> C, B -> C (where A and B have same outgoing edge attributes)
	// D -> A, D -> B (both will become D -> MERGED)
	// When A and B collapse, edges are deduplicated
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"type": "service"}},
			"B": {ID: "B", Attributes: map[string]interface{}{"type": "service"}},
			"C": {ID: "C", Attributes: map[string]interface{}{"type": "db"}},
			"D": {ID: "D", Attributes: map[string]interface{}{"type": "client"}},
		},
		Edges: []*gsl.Edge{
			{From: "D", To: "A", Attributes: map[string]interface{}{"type": "http"}},
			{From: "D", To: "B", Attributes: map[string]interface{}{"type": "http"}},
			{From: "A", To: "C", Attributes: map[string]interface{}{}},
			{From: "B", To: "C", Attributes: map[string]interface{}{}},
		},
		Sets: make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	expr := &CollapseExpr{
		NodeID: "SERVICES",
		Pred: &AttributeEqualsPredicate{
			Target: "node",
			Name:   "type",
			Value:  "service",
		},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)

	// Should have: D -> SERVICES (deduplicated, both had same attributes)
	// SERVICES -> C (deduplicated from A->C and B->C with same attributes)
	// Total: 2 edges
	if len(gv.Graph.GetEdges()) != 2 {
		t.Errorf("Expected 2 edges after dedup, got %d", len(gv.Graph.GetEdges()))
	}
}

// TestCollapsePreservesSets tests that sets are preserved during collapse
func TestCollapsePreservesSets(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets: map[string]*gsl.Set{
			"CRITICAL": {ID: "CRITICAL", Attributes: map[string]interface{}{}},
		},
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	expr := &CollapseExpr{
		NodeID: "MERGED",
		Pred:   &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if _, hasSet := gv.Graph.GetSets()["CRITICAL"]; !hasSet {
		t.Error("Expected set to be preserved")
	}
}

// TestCollapseNoMatch tests collapse when no nodes match
func TestCollapseNoMatch(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"type": "service"}},
			"B": {ID: "B", Attributes: map[string]interface{}{"type": "db"}},
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

	expr := &CollapseExpr{
		NodeID: "MERGED",
		Pred: &AttributeEqualsPredicate{
			Target: "node",
			Name:   "type",
			Value:  "cache", // no nodes match
		},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)

	// Graph should be unchanged
	if len(gv.Graph.GetNodes()) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(gv.Graph.GetNodes()))
	}
	if _, hasMerged := gv.Graph.GetNodes()["MERGED"]; hasMerged {
		t.Error("Expected no MERGED node")
	}
}

// TestCollapseParserBasic tests parsing of collapse expressions
func TestCollapseParserBasic(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"collapse into MERGED where exists", false},
		{"collapse into MERGED where node.type = \"service\"", false},
		{"collapse into X where in CRITICAL", false},
		{"collapse", true},
		{"collapse into MERGED", true},
		{"collapse MERGED where exists", true},
	}

	for _, tt := range tests {
		p := newExpressionParser(tt.input)
		expr, err := p.parse()
		
		if (err != nil) != tt.wantErr {
			t.Errorf("parse(%q): wantErr=%v, got=%v", tt.input, tt.wantErr, err)
			continue
		}
		
		if err == nil {
			if _, ok := expr.(*CollapseExpr); !ok {
				t.Errorf("parse(%q): expected CollapseExpr, got %T", tt.input, expr)
			}
		}
	}
}

// TestCollapseInPipeline tests collapse in query pipeline
func TestCollapseInPipeline(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"type": "service"}},
			"B": {ID: "B", Attributes: map[string]interface{}{"type": "service"}},
			"C": {ID: "C", Attributes: map[string]interface{}{"type": "db"}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "C", Attributes: map[string]interface{}{}},
			{From: "B", To: "C", Attributes: map[string]interface{}{}},
		},
		Sets: make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	query, err := NewQueryParser("collapse into SERVICES where node.type = \"service\"").Parse()
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
	if _, hasServices := gv.Graph.GetNodes()["SERVICES"]; !hasServices {
		t.Error("Expected SERVICES node")
	}
}

// TestCollapseChained tests collapse chained with other operations
func TestCollapseChained(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"team": "payments"}},
			"B": {ID: "B", Attributes: map[string]interface{}{"team": "payments"}},
			"C": {ID: "C", Attributes: map[string]interface{}{"team": "fraud"}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "C", Attributes: map[string]interface{}{}},
			{From: "B", To: "C", Attributes: map[string]interface{}{}},
		},
		Sets: make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	query, err := NewQueryParser("collapse into PAYMENTS where node.team = \"payments\" | make node.critical = true where exists").Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	gv := result.(GraphValue)

	// Check that all nodes are marked critical
	for _, node := range gv.Graph.GetNodes() {
		if val := node.Attributes["critical"]; val != true {
			t.Errorf("Expected critical=true, got %v", val)
		}
	}
}

// TestCollapseSelfLoop tests collapse with self-loops and external edges
func TestCollapseSelfLoop(t *testing.T) {
	// A -> A (self-loop), A -> B, C -> A
	// Collapse A into MERGED, keeping B and C
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"type": "service"}},
			"B": {ID: "B", Attributes: map[string]interface{}{"type": "db"}},
			"C": {ID: "C", Attributes: map[string]interface{}{"type": "client"}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "A", Attributes: map[string]interface{}{}},
			{From: "A", To: "B", Attributes: map[string]interface{}{}},
			{From: "C", To: "A", Attributes: map[string]interface{}{}},
		},
		Sets: make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	expr := &CollapseExpr{
		NodeID: "MERGED",
		Pred: &AttributeEqualsPredicate{
			Target: "node",
			Name:   "type",
			Value:  "service",
		},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)

	// Should have: C -> MERGED (from C->A), MERGED -> B (from A->B)
	// A -> A becomes internal edge (removed)
	// Total: 2 edges
	if len(gv.Graph.GetEdges()) != 2 {
		t.Errorf("Expected 2 edges, got %d", len(gv.Graph.GetEdges()))
	}
}

// TestCollapseEmptyGraph tests collapse on empty graph
func TestCollapseEmptyGraph(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	expr := &CollapseExpr{
		NodeID: "MERGED",
		Pred:   &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.GetNodes()) != 0 {
		t.Error("Expected empty graph to remain empty")
	}
}

// TestCollapseSingleNode tests collapse of single node
func TestCollapseSingleNode(t *testing.T) {
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

	expr := &CollapseExpr{
		NodeID: "SINGLE",
		Pred:   &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.GetNodes()) != 1 {
		t.Errorf("Expected 1 node, got %d", len(gv.Graph.GetNodes()))
	}
	if node, ok := gv.Graph.GetNodes()["SINGLE"]; !ok {
		t.Error("Expected SINGLE node")
	} else {
		if node.Attributes["team"] != "payments" {
			t.Error("Expected team attribute preserved")
		}
	}
}

// TestCollapseMixedTargets tests that mixed targets produce error
func TestCollapseMixedTargets(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	// Mixed target predicate
	expr := &CollapseExpr{
		NodeID: "MERGED",
		Pred: &AndPredicate{
			Left:  &AttributeEqualsPredicate{Target: "node", Name: "type", Value: "service"},
			Right: &AttributeEqualsPredicate{Target: "edge", Name: "type", Value: "http"},
		},
	}

	_, err := expr.Apply(ctx, GraphValue{graph})
	if err == nil {
		t.Error("Expected mixed target error")
	}
}

// TestCollapseEdgePredicate tests that edge predicates are rejected
func TestCollapseEdgePredicate(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	// Edge predicate (not allowed for collapse)
	expr := &CollapseExpr{
		NodeID: "MERGED",
		Pred: &AttributeEqualsPredicate{
			Target: "edge",
			Name:   "type",
			Value:  "http",
		},
	}

	_, err := expr.Apply(ctx, GraphValue{graph})
	if err == nil {
		t.Error("Expected edge predicate error")
	}
}
