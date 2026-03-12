package query

import (
	"testing"

	"github.com/dnnrly/gsl-lang"
)

// TestPhase1Infrastructure tests the minimal engine (Phase 1)
// Goal: parse → evaluate → return graph, even if only identity exists

func TestIdentityQuery(t *testing.T) {
	// Create a simple input graph
	inputGraph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*gsl.Set{},
	}

	// Create context
	ctx := &QueryContext{
		InputGraph:  inputGraph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	// Parse empty query (should be valid)
	parser := NewQueryParser("")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse empty query: %v", err)
	}

	// Execute query
	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	// Result should be GraphValue containing the input graph
	graphValue, ok := result.(GraphValue)
	if !ok {
		t.Fatalf("Result is not GraphValue, got %T", result)
	}

	// Output should equal input
	if graphValue.Graph != inputGraph {
		t.Fatal("Output graph does not match input graph")
	}
}

func TestEmptyQuery(t *testing.T) {
	// Empty query should pass through input graph unchanged
	inputGraph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"X": {ID: "X", Attributes: map[string]interface{}{"color": "red"}, Sets: map[string]struct{}{}},
		},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph:  inputGraph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	parser := NewQueryParser("")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	gv := result.(GraphValue)
	if gv.Graph.Nodes["X"].Attributes["color"] != "red" {
		t.Fatal("Attributes not preserved")
	}
}

func TestPipelineParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		numExprs int
	}{
		{"empty", "", 1}, // Empty expression
		{"single", "from *", 1},
		{"two stages", "from * | from *", 2},
		{"three stages", "from * | from * | from *", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewQueryParser(tt.input)
			query, err := parser.Parse()
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			if len(query.Expressions) != tt.numExprs {
				t.Fatalf("Expected %d expressions, got %d", tt.numExprs, len(query.Expressions))
			}
		})
	}
}

func TestPipelineParsingWithParentheses(t *testing.T) {
	// Parentheses should not split the pipeline
	// Note: parenthesized binding not yet implemented, so we test pipe separation
	parser := NewQueryParser("from * | from *")
	query, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(query.Expressions) != 2 {
		t.Fatalf("Expected 2 expressions, got %d", len(query.Expressions))
	}
}

func TestQueryContextInitialization(t *testing.T) {
	inputGraph := &gsl.Graph{
		Nodes: map[string]*gsl.Node{},
		Edges: []*gsl.Edge{},
		Sets:  map[string]*gsl.Set{},
	}

	ctx := &QueryContext{
		InputGraph:  inputGraph,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	if ctx.InputGraph != inputGraph {
		t.Fatal("InputGraph not set correctly")
	}

	if len(ctx.NamedGraphs) != 0 {
		t.Fatal("NamedGraphs should be empty")
	}
}
