package gsl

import (
	"io"
)

// Parse reads a GSL document and produces a Graph.
// Returns the graph and a ParseError (nil on success).
// The ParseError distinguishes between fatal errors and non-fatal warnings.
func Parse(r io.Reader) (*Graph, *ParseError) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, &ParseError{
			Message: "failed to read input",
			Err:     err,
		}
	}

	prog, parseErrors := parse(string(data))
	if len(parseErrors) > 0 {
		return nil, &ParseError{
			Message: "parse failed",
			Err:     parseErrors[0],
		}
	}

	graph, warnings, buildErr := buildGraph(prog)
	if buildErr != nil {
		return nil, &ParseError{
			Message:  "build failed",
			Err:      buildErr,
			Warnings: warnings,
		}
	}

	// Success: return graph with any warnings
	if len(warnings) > 0 {
		return graph, &ParseError{
			Message:  "",
			Warnings: warnings,
		}
	}

	return graph, nil
}
