package main

import (
	"fmt"
	"os"

	gsl "github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/cmd/internal/cli"
	"github.com/spf13/cobra"
)

// Build information injected by goreleaser
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func main() {
	var inputFile, outputFile, queryFile string
	var query string

	rootCmd := &cobra.Command{
		Use:   "gsl-query [query]",
		Short: "Query GSL graphs",
		Long:  "Execute queries against GSL (Graph Specification Language) graphs and output filtered/transformed results.\n\nFor AI/LLM guidance, use: gsl-query help ai",
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

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			cli.PrintVersion("gsl-query", Version, Commit, BuildDate)
		},
	}

	// Load guides
	gslDesc, gslContent, gslErr := cli.LoadGuide(gsl.Guides, "LLM_GUIDE.md")
	queryDesc, queryContent, queryErr := cli.LoadGuide(gsl.Guides, "QUERY_AI_GUIDE.md")

	guides := []cli.GuideSpec{
		{Use: "gsl", Desc: gslDesc, Content: gslContent, Err: gslErr},
		{Use: "query", Desc: queryDesc, Content: queryContent, Err: queryErr},
	}

	// Build AI command trees
	cli.BuildAICommand(rootCmd, guides, "Available AI/LLM guides:", "Run 'gsl-query ai <guide>' to view a guide")
	cli.BuildAICommand(helpCmd, guides, "Available AI/LLM guides:", "Run 'gsl-query help ai <guide>' to view a guide")

	rootCmd.AddCommand(helpCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
