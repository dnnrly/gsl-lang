package query

import (
	"github.com/dnnrly/gsl-lang"
)

// testGraphInput is a struct to hold graph components for test fixtures.
type testGraphInput struct {
	Nodes map[string]*gsl.Node
	Edges []*gsl.Edge
	Sets  map[string]*gsl.Set
}

// newTestGraph creates a graph from nodes, edges, and sets.
// Used for test fixture construction to avoid direct field access.
func newTestGraph(input testGraphInput) *gsl.Graph {
	g := gsl.NewGraph()
	// Use internal function to set state (only for tests)
	g.SetInternalState(input.Nodes, input.Edges, input.Sets)
	return g
}
