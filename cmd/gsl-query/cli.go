package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	gsl "github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/query"
)

// Config holds the query execution configuration
type Config struct {
	InputFile  string
	OutputFile string
	Query      string
	InputName  string // Display name for input (file path or "<stdin>")
}

// Execute runs the query pipeline
func Execute(cfg *Config) error {
	// Read input GSL
	var input []byte
	var err error

	if cfg.InputFile != "" {
		input, err = os.ReadFile(cfg.InputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
	} else {
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
	}

	// Parse GSL
	graph, parseErr := gsl.Parse(bytes.NewReader(input))
	if parseErr != nil && parseErr.HasError() {
		fmt.Fprintf(os.Stderr, "Error: failed to parse %s: %v\n", cfg.InputName, parseErr)
		return parseErr
	}

	// Log warnings if any
	if parseErr != nil && parseErr.HasWarnings() {
		for _, w := range parseErr.Warnings {
			fmt.Fprintf(os.Stderr, "Warning [%s]: %v\n", cfg.InputName, w)
		}
	}

	// Parse query
	parser := query.NewQueryParser(cfg.Query)
	q, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse query: %v\n", err)
		return err
	}

	// Execute query
	ctx := &query.QueryContext{
		InputGraph:  graph,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	result, err := q.Execute(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: query execution failed: %v\n", err)
		return err
	}

	// Extract graph from result
	var resultGraph *gsl.Graph
	switch v := result.(type) {
	case query.GraphValue:
		resultGraph = v.Graph
	default:
		return fmt.Errorf("unexpected query result type: %T", result)
	}

	// Serialize result
	output := gsl.Serialize(resultGraph)

	// Write output
	if cfg.OutputFile != "" {
		if err := os.WriteFile(cfg.OutputFile, []byte(output), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to write output file: %v\n", err)
			return err
		}
	} else {
		fmt.Print(output)
	}

	return nil
}
