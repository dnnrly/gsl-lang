package main

import (
	"fmt"
	"os"

	"github.com/dnnrly/gsl-lang/cmd/gsl-diagram/formats"
	"github.com/spf13/cobra"
)

func main() {
	var inputFile, outputFile, format, diagramType string

	rootCmd := &cobra.Command{
		Use:   "gsl-diagram",
		Short: "Convert GSL graphs to diagrams",
		Long:  "Convert GSL (Graph Specification Language) documents to various diagram formats",
		RunE: func(cmd *cobra.Command, args []string) error {
			factory, err := formats.GetFactory(format)
			if err != nil {
				return err
			}

			cfg := &Config{
				InputFile:   inputFile,
				OutputFile:  outputFile,
				DiagramType: diagramType,
				Converter:   factory,
			}

			return Execute(cfg)
		},
	}

	rootCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input GSL file (read from stdin if not provided)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output diagram file (write to stdout if not provided)")
	rootCmd.Flags().StringVarP(&format, "format", "f", "mermaid", "Output format: mermaid, plantuml (default: mermaid)")
	rootCmd.Flags().StringVarP(&diagramType, "type", "t", "component", "Diagram type (default: component)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
