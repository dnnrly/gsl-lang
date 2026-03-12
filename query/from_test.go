package query

import (
	"testing"

	"github.com/dnnrly/gsl-lang"
)

// TestFromWildcard tests "from *" - selects input graph
func TestFromWildcard(t *testing.T) {
	// Create input graph
	inputGraph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph:  inputGraph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	parser := NewQueryParser("from *")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	if gv.Graph != inputGraph {
		t.Fatal("from * should return input graph")
	}
}

// TestFromNamedGraph tests "from NAME" - selects named graph
func TestFromNamedGraph(t *testing.T) {
	namedGraph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	}

	inputGraph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph: inputGraph,
		NamedGraphs: map[string]*gsl.Graph{
			"SERVICES": namedGraph,
		},
	}

	parser := NewQueryParser("from SERVICES")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	if gv.Graph != namedGraph {
		t.Fatal("from SERVICES should return named graph")
	}
}

// TestFromMissingNamedGraph tests error when named graph doesn't exist
func TestFromMissingNamedGraph(t *testing.T) {
	ctx := &QueryContext{
		InputGraph:  &gsl.Graph{Nodes: map[string]*gsl.Node{}, Edges: []*gsl.Edge{}, Sets: map[string]*gsl.Set{}},
		NamedGraphs: map[string]*gsl.Graph{},
	}

	parser := NewQueryParser("from NONEXISTENT")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	_, err = query.Execute(ctx)
	if err == nil {
		t.Fatal("Expected error for missing named graph")
	}
	if err.Error() != "named graph not found: NONEXISTENT" {
		t.Fatalf("Wrong error message: %v", err)
	}
}

// TestFromInvalidSyntax tests parsing errors
func TestFromInvalidSyntax(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		errMsg    string
	}{
		{"from alone", "from", true, "from requires an argument"},
		{"from lowercase", "from services", true, "invalid graph name"},
		{"from digit start", "from 123GRAPH", true, "invalid graph name"},
		{"from hyphen", "from GRAPH-NAME", true, "invalid graph name"},
		{"from space in name", "from GRAPH NAME", true, "invalid graph name"},
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

// TestFromValidGraphNames tests valid graph name patterns
func TestFromValidGraphNames(t *testing.T) {
	validNames := []string{
		"A",
		"AB",
		"ABC",
		"A0",
		"A_B",
		"A_B_C",
		"A0B1C2_D_E_F",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			ctx := &QueryContext{
				InputGraph: &gsl.Graph{Nodes: map[string]*gsl.Node{}, Edges: []*gsl.Edge{}, Sets: map[string]*gsl.Set{}},
				NamedGraphs: map[string]*gsl.Graph{
					name: &gsl.Graph{Nodes: map[string]*gsl.Node{}, Edges: []*gsl.Edge{}, Sets: map[string]*gsl.Set{}},
				},
			}

			parser := NewQueryParser("from " + name)
			query, err := parser.Parse()
			if err != nil {
				t.Fatalf("Failed to parse valid name: %v", err)
			}

			_, err = query.Execute(ctx)
			if err != nil {
				t.Fatalf("Failed to execute: %v", err)
			}
		})
	}
}

// TestFromInvalidGraphNames tests invalid graph name patterns
func TestFromInvalidGraphNames(t *testing.T) {
	invalidNames := []string{
		"",          // empty
		"a",         // lowercase
		"abc",       // lowercase
		"_ABC",      // underscore first
		"0ABC",      // digit first
		"A-B",       // hyphen
		"A.B",       // dot
		"A B",       // space
		"A\tB",      // tab
	}

	for _, name := range invalidNames {
		t.Run(name, func(t *testing.T) {
			parser := NewQueryParser("from " + name)
			_, err := parser.Parse()
			if err == nil {
				t.Fatal("Expected error for invalid name")
			}
		})
	}
}

// TestFromInPipeline tests "from" as part of a pipeline
func TestFromInPipeline(t *testing.T) {
	named := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph: &gsl.Graph{
			Nodes: map[string]*gsl.Node{
				"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			},
			Edges: []*gsl.Edge{},
			Sets:  map[string]*gsl.Set{},
		},
		NamedGraphs: map[string]*gsl.Graph{
			"NAMED": named,
		},
	}

	parser := NewQueryParser("from NAMED")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	if len(gv.Graph.Nodes) != 1 {
		t.Fatal("Pipeline should execute from and return correct graph")
	}
	if _, ok := gv.Graph.Nodes["C"]; !ok {
		t.Fatal("Pipeline should contain node C")
	}
}

// TestFromExprDirectly tests FromExpr.Apply() directly
func TestFromExprDirectly(t *testing.T) {
	namedGraph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"X": {ID: "X", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	}

	inputGraph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"Y": {ID: "Y", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph: inputGraph,
		NamedGraphs: map[string]*gsl.Graph{
			"NAMED": namedGraph,
		},
	}

	// Test wildcard
	expr := &FromExpr{IsWildcard: true}
	result, err := expr.Apply(ctx, GraphValue{})
	if err != nil {
		t.Fatalf("Wildcard apply failed: %v", err)
	}
	if result.(GraphValue).Graph != inputGraph {
		t.Fatal("Wildcard should return input graph")
	}

	// Test named graph
	expr = &FromExpr{IsWildcard: false, Name: "NAMED"}
	result, err = expr.Apply(ctx, GraphValue{})
	if err != nil {
		t.Fatalf("Named apply failed: %v", err)
	}
	if result.(GraphValue).Graph != namedGraph {
		t.Fatal("Named should return named graph")
	}

	// Test missing named graph
	expr = &FromExpr{IsWildcard: false, Name: "MISSING"}
	_, err = expr.Apply(ctx, GraphValue{})
	if err == nil {
		t.Fatal("Expected error for missing graph")
	}
}
