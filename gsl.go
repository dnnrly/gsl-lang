package gsl

import (
	"io"
)

// Parse reads a GSL document and produces a Graph.
// Returns the graph, a slice of non-fatal warnings, and an error if parsing failed.
func Parse(r io.Reader) (*Graph, []error, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}

	prog, parseErrors := parse(string(data))
	if len(parseErrors) > 0 {
		return nil, nil, parseErrors[0]
	}

	graph, warnings, buildErr := buildGraph(prog)
	return graph, warnings, buildErr
}
