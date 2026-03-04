package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	gsl "github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/cmd/gsl-diagram/converter"
)

// Config holds the diagram conversion configuration
type Config struct {
	InputFile   string
	OutputFile  string
	DiagramType string
	Converter   converter.Factory
	InputName   string // Display name for input (file path or "<stdin>")
}

// Execute runs the conversion pipeline
func Execute(cfg *Config) error {
	// Read input
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
	graph, warnings, err := gsl.Parse(bytes.NewReader(input))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse %s: %v\n", cfg.InputName, err)
		return err
	}

	// Log warnings if any
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "Warning [%s]: %v\n", cfg.InputName, w)
	}

	// Convert to diagram
	conv := cfg.Converter(cfg.DiagramType)
	output := conv.Convert(graph)

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
