package query

import (
	"testing"

	"github.com/dnnrly/gsl-lang"
)

// TestBindSimple tests basic binding: (from *) as NAME
func TestBindSimple(t *testing.T) {
	inputGraph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  inputGraph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	parser := NewQueryParser("(from *) as SNAPSHOT")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	// Result should be input graph (binding returns input unchanged)
	gv := result.(GraphValue)
	if gv.Graph != inputGraph {
		t.Fatal("Binding should return input graph")
	}

	// Named graph should be stored
	if _, exists := ctx.NamedGraphs["SNAPSHOT"]; !exists {
		t.Fatal("Named graph should be stored")
	}

	if ctx.NamedGraphs["SNAPSHOT"] != inputGraph {
		t.Fatal("Named graph should be the bound graph")
	}
}

// TestBindMultiple tests chaining bindings
func TestBindMultiple(t *testing.T) {
	input1 := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	})

	input2 := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph: input1,
		NamedGraphs: map[string]*gsl.Graph{
			"OTHER": input2,
		},
	}

	// Query: bind input as FIRST, then bind OTHER as SECOND
	// After first binding, the value is still input1
	// We can't retrieve OTHER in the second binding because the second binding
	// would receive input1 as its input value (not used, since (from OTHER) starts fresh)
	parser := NewQueryParser("(from *) as FIRST | (from OTHER) as SECOND")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	// Both should be bound
	if ctx.NamedGraphs["FIRST"] != input1 {
		t.Fatal("FIRST should be bound to input")
	}

	if ctx.NamedGraphs["SECOND"] != input2 {
		t.Fatal("SECOND should be bound to OTHER")
	}

	// Result is from the binding, which returns the input to the binding
	// The second binding receives input1 (from previous binding), returns it unchanged
	gv := result.(GraphValue)
	if gv.Graph != input1 {
		t.Fatal("Result should be input1 (returned unchanged by second binding)")
	}
}

// TestBindImmutability tests that names cannot be rebound
func TestBindImmutability(t *testing.T) {
	inputGraph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  inputGraph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// First binding should succeed
	parser := NewQueryParser("(from *) as MYNAME")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	_, err = query.Execute(ctx)
	if err != nil {
		t.Fatalf("First binding failed: %v", err)
	}

	// Second binding with same name should fail
	parser = NewQueryParser("(from *) as MYNAME")
	query, err = parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	_, err = query.Execute(ctx)
	if err == nil {
		t.Fatal("Rebinding should fail")
	}
	if err.Error() != "named graph already bound: MYNAME" {
		t.Fatalf("Wrong error message: %v", err)
	}
}

// TestBindInvalidSyntax tests parsing errors
func TestBindInvalidSyntax(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"unclosed paren", "(from * as NAME", true},
		{"missing as", "(from *) NAME", true},
		{"missing name", "(from *) as", true},
		{"invalid name", "(from *) as lowercase", true},
		{"invalid name format", "(from *) as 123", true},
		{"empty pipeline", "() as NAME", false}, // Empty pipeline is valid (identity)
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

// TestBindNestedPipeline tests complex subpipelines in binding
func TestBindNestedPipeline(t *testing.T) {
	graph1 := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"X": {ID: "X", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	})

	graph2 := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"Y": {ID: "Y", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph: graph1,
		NamedGraphs: map[string]*gsl.Graph{
			"G2": graph2,
		},
	}

	// Bind a pipeline: (from G2) as RESULT
	// This executes the subpipeline "from G2" which returns graph2, then binds it
	parser := NewQueryParser("(from G2) as RESULT")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	_, err = query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	// RESULT should be graph2 (result of from G2)
	if ctx.NamedGraphs["RESULT"] != graph2 {
		t.Fatal("RESULT should be G2")
	}
}

// TestBindThenUse tests binding then using named graph
func TestBindThenUse(t *testing.T) {
	originalInput := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"ORIG": {ID: "ORIG", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  originalInput,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Query: (from *) as SAVED | from SAVED
	// This binds input, then retrieves it
	parser := NewQueryParser("(from *) as SAVED | from SAVED")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	if gv.Graph != originalInput {
		t.Fatal("Result should be the original input retrieved from SAVED")
	}
}

// TestBindExprDirectly tests BindExpr.Apply() directly
func TestBindExprDirectly(t *testing.T) {
	inputGraph := newTestGraph(testGraphInput{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	})

	ctx := &QueryContext{
		InputGraph:  inputGraph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Create a subpipeline that returns input
	subQuery := &Query{Expressions: []Expression{&IdentityExpr{}}}

	// Create binding expression
	bind := &BindExpr{Pipeline: subQuery, Name: "TEST"}

	// Apply it
	input := GraphValue{inputGraph}
	result, err := bind.Apply(ctx, input)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Result should be input unchanged
	if result.(GraphValue).Graph != inputGraph {
		t.Fatal("Binding should return input unchanged")
	}

	// Named graph should be stored
	if ctx.NamedGraphs["TEST"] != inputGraph {
		t.Fatal("Named graph should be stored")
	}

	// Attempting to rebind should fail
	bind2 := &BindExpr{Pipeline: subQuery, Name: "TEST"}
	_, err = bind2.Apply(ctx, input)
	if err == nil {
		t.Fatal("Rebinding should fail")
	}
}

// TestBindValidNames tests valid binding names
func TestBindValidNames(t *testing.T) {
	validNames := []string{"A", "ABC", "A_B_C", "A0B1C2"}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			ctx := &QueryContext{
				InputGraph:  newTestGraph(testGraphInput{Nodes: map[string]*gsl.Node{}, Edges: []*gsl.Edge{}, Sets: map[string]*gsl.Set{}}),
				NamedGraphs: map[string]*gsl.Graph{},
			}

			input := "(from *) as " + name
			parser := NewQueryParser(input)
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
