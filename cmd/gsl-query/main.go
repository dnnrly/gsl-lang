package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var inputFile, outputFile, queryFile string
	var query string

	rootCmd := &cobra.Command{
		Use:   "gsl-query [query]",
		Short: "Query GSL graphs",
		Long:  "Execute queries against GSL (Graph Specification Language) graphs and output filtered/transformed results",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine query source
			if queryFile != "" {
				// Read from file
				content, err := os.ReadFile(queryFile)
				if err != nil {
					return fmt.Errorf("failed to read query file: %w", err)
				}
				query = string(content)
			} else if len(args) > 0 {
				// From positional argument
				query = args[0]
			} else {
				return fmt.Errorf("query required: provide as argument or use --query-file")
			}

			inputName := inputFile
			if inputName == "" {
				inputName = "<stdin>"
			}

			cfg := &Config{
				InputFile:  inputFile,
				OutputFile: outputFile,
				Query:      query,
				InputName:  inputName,
			}

			return Execute(cfg)
		},
	}

	rootCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input GSL file (read from stdin if not provided)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output GSL file (write to stdout if not provided)")
	rootCmd.Flags().StringVarP(&queryFile, "query-file", "f", "", "Read query from file")

	helpCmd := &cobra.Command{
		Use:   "help",
		Short: "Show help for gsl-query",
		Run: func(cmd *cobra.Command, args []string) {
			_ = rootCmd.Help()
		},
	}

	rootCmd.AddCommand(helpCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
