package query

import (
	"testing"

	"github.com/dnnrly/gsl-lang"
)

// TestMakeNodeAttributeBasic tests basic node attribute assignment
func TestMakeNodeAttributeBasic(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	// Set team on all nodes
	expr := &MakeExpr{
		Target: "node",
		Attr:   "team",
		Value:  "payments",
		Pred:   &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	for _, node := range gv.Graph.GetNodes() {
		if val, ok := node.Attributes["team"]; !ok || val != "payments" {
			t.Errorf("Expected team=payments, got %v", val)
		}
	}
}

// TestMakeNodeAttributeWithPredicate tests attribute assignment with predicate
func TestMakeNodeAttributeWithPredicate(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"type": "service"}},
			"B": {ID: "B", Attributes: map[string]interface{}{"type": "database"}},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	// Set owner only on services
	expr := &MakeExpr{
		Target: "node",
		Attr:   "owner",
		Value:  "alice",
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
	nodeA := gv.Graph.GetNodes()["A"]
	nodeB := gv.Graph.GetNodes()["B"]

	if owner, ok := nodeA.Attributes["owner"]; !ok || owner != "alice" {
		t.Errorf("Expected A owner=alice, got %v", owner)
	}
	if _, ok := nodeB.Attributes["owner"]; ok {
		t.Error("Expected B to not have owner")
	}
}

// TestMakeEdgeAttributeBasic tests basic edge attribute assignment
func TestMakeEdgeAttributeBasic(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}},
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

	// Set protocol on all edges
	expr := &MakeExpr{
		Target: "edge",
		Attr:   "protocol",
		Value:  "http",
		Pred:   &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.GetEdges()) != 1 {
		t.Fatalf("Expected 1 edge, got %d", len(gv.Graph.GetEdges()))
	}
	if val, ok := gv.Graph.GetEdges()[0].Attributes["protocol"]; !ok || val != "http" {
		t.Errorf("Expected protocol=http, got %v", val)
	}
}

// TestMakeEdgeAttributeWithPredicate tests edge attribute assignment with predicate
func TestMakeEdgeAttributeWithPredicate(t *testing.T) {
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

	// Set timeout on HTTP edges only
	expr := &MakeExpr{
		Target: "edge",
		Attr:   "timeout",
		Value:  30,
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
	if timeout, ok := gv.Graph.GetEdges()[0].Attributes["timeout"]; !ok || timeout != 30 {
		t.Errorf("Expected first edge timeout=30, got %v", timeout)
	}
	if _, ok := gv.Graph.GetEdges()[1].Attributes["timeout"]; ok {
		t.Error("Expected second edge to not have timeout")
	}
}

// TestMakeOverwriteAttribute tests overwriting existing attributes
func TestMakeOverwriteAttribute(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"team": "fraud"}},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	// Overwrite team attribute
	expr := &MakeExpr{
		Target: "node",
		Attr:   "team",
		Value:  "payments",
		Pred:   &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if team, ok := gv.Graph.GetNodes()["A"].Attributes["team"]; !ok || team != "payments" {
		t.Errorf("Expected team=payments, got %v", team)
	}
}

// TestMakePreservesOtherAttributes tests that make preserves other attributes
func TestMakePreservesOtherAttributes(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"team": "payments", "owner": "alice"}},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	// Add new attribute while preserving existing ones
	expr := &MakeExpr{
		Target: "node",
		Attr:   "env",
		Value:  "prod",
		Pred:   &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	attrs := gv.Graph.GetNodes()["A"].Attributes
	if attrs["team"] != "payments" {
		t.Error("Expected team to be preserved")
	}
	if attrs["owner"] != "alice" {
		t.Error("Expected owner to be preserved")
	}
	if attrs["env"] != "prod" {
		t.Error("Expected env to be set")
	}
}

// TestMakeStringValue tests string value assignment
func TestMakeStringValue(t *testing.T) {
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

	expr := &MakeExpr{
		Target: "node",
		Attr:   "name",
		Value:  "example",
		Pred:   &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if val := gv.Graph.GetNodes()["A"].Attributes["name"]; val != "example" {
		t.Errorf("Expected 'example', got %v", val)
	}
}

// TestMakeBooleanValue tests boolean value assignment
func TestMakeBooleanValue(t *testing.T) {
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

	expr := &MakeExpr{
		Target: "node",
		Attr:   "critical",
		Value:  true,
		Pred:   &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if val := gv.Graph.GetNodes()["A"].Attributes["critical"]; val != true {
		t.Errorf("Expected true, got %v", val)
	}
}

// TestMakeNumericValue tests numeric value assignment
func TestMakeNumericValue(t *testing.T) {
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

	expr := &MakeExpr{
		Target: "node",
		Attr:   "priority",
		Value:  "42", // parseValue treats numbers as strings
		Pred:   &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if val := gv.Graph.GetNodes()["A"].Attributes["priority"]; val != "42" {
		t.Errorf("Expected '42', got %v", val)
	}
}

// TestMakeExprParser tests parsing of make expressions
func TestMakeExprParser(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
		exprType string
	}{
		{"make node.team = \"payments\" where exists", false, "Make"},
		{"make edge.protocol = \"http\" where exists", false, "Make"},
		{"make node.priority = 5 where node.type = \"service\"", false, "Make"},
		{"make edge.timeout = 30 where exists", false, "Make"},
		{"make", true, ""},
		{"make node.attr", true, ""},
		{"make node.attr = value", true, ""},
	}

	for _, tt := range tests {
		p := newExpressionParser(tt.input)
		expr, err := p.parse()
		
		if (err != nil) != tt.wantErr {
			t.Errorf("parse(%q): wantErr=%v, got=%v", tt.input, tt.wantErr, err)
			continue
		}
		
		if err == nil {
			if _, ok := expr.(*MakeExpr); !ok {
				t.Errorf("parse(%q): expected MakeExpr, got %T", tt.input, expr)
			}
		}
	}
}

// TestMakeInPipeline tests make in query pipeline
func TestMakeInPipeline(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"type": "service"}},
			"B": {ID: "B", Attributes: map[string]interface{}{"type": "database"}},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	query, err := NewQueryParser("make node.owner = \"alice\" where node.type = \"service\"").Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	gv := result.(GraphValue)
	if owner, ok := gv.Graph.GetNodes()["A"].Attributes["owner"]; !ok || owner != "alice" {
		t.Errorf("Expected A owner=alice, got %v", owner)
	}
	if _, ok := gv.Graph.GetNodes()["B"].Attributes["owner"]; ok {
		t.Error("Expected B to not have owner")
	}
}

// TestMakeChained tests make chained with other operations
func TestMakeChained(t *testing.T) {
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

	// Make operation followed by subgraph
	query, err := NewQueryParser("make node.marked = true where exists | subgraph node.marked = true").Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.GetNodes()) != 3 {
		t.Errorf("Expected 3 nodes in result, got %d", len(gv.Graph.GetNodes()))
	}
}

// TestMakePreservesSets tests that make preserves sets
func TestMakePreservesSets(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}},
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

	expr := &MakeExpr{
		Target: "node",
		Attr:   "team",
		Value:  "payments",
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

// TestMakePreservesSets tests that make preserves nodes and edges
func TestMakePreservesStructure(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}},
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

	expr := &MakeExpr{
		Target: "node",
		Attr:   "team",
		Value:  "payments",
		Pred:   &ExistsPredicate{},
	}

	result, err := expr.Apply(ctx, GraphValue{graph})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.GetNodes()) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(gv.Graph.GetNodes()))
	}
	if len(gv.Graph.GetEdges()) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(gv.Graph.GetEdges()))
	}
}

// TestMakeEmptyGraph tests make on empty graph
func TestMakeEmptyGraph(t *testing.T) {
	graph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	})

	ctx := &QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	expr := &MakeExpr{
		Target: "node",
		Attr:   "team",
		Value:  "payments",
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

// TestMakeMixedTargets tests that mixed targets produce error
func TestMakeMixedTargets(t *testing.T) {
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

	// Mixed target predicate (node and edge)
	expr := &MakeExpr{
		Target: "node",
		Attr:   "team",
		Value:  "payments",
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

// TestMakeParserQuotedValues tests parsing quoted values
func TestMakeParserQuotedValues(t *testing.T) {
	query, err := NewQueryParser(`make node.team = "payments" where exists`).Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(query.Expressions) != 1 {
		t.Fatalf("Expected 1 expression, got %d", len(query.Expressions))
	}

	expr := query.Expressions[0].(*MakeExpr)
	if expr.Value != "payments" {
		t.Errorf("Expected value 'payments', got %v", expr.Value)
	}
}

// TestMakeParserNumericValues tests parsing numeric values
func TestMakeParserNumericValues(t *testing.T) {
	query, err := NewQueryParser("make node.priority = 42 where exists").Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expr := query.Expressions[0].(*MakeExpr)
	if expr.Value != "42" {
		t.Errorf("Expected '42', got %v", expr.Value)
	}
}

// TestMakeParserBooleanValues tests parsing boolean values
func TestMakeParserBooleanValues(t *testing.T) {
	query, err := NewQueryParser("make node.critical = true where exists").Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expr := query.Expressions[0].(*MakeExpr)
	if expr.Value != true {
		t.Errorf("Expected true, got %v", expr.Value)
	}
}
