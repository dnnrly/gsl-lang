package main

import (
	"fmt"
	"os"

	gsl "github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/cmd/gsl-diagram/formats"
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

			inputName := inputFile
			if inputName == "" {
				inputName = "<stdin>"
			}

			cfg := &Config{
				InputFile:   inputFile,
				OutputFile:  outputFile,
				DiagramType: diagramType,
				Converter:   factory,
				InputName:   inputName,
			}

			return Execute(cfg)
		},
	}

	rootCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input GSL file (read from stdin if not provided)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output diagram file (write to stdout if not provided)")
	rootCmd.Flags().StringVarP(&format, "format", "f", "mermaid", "Output format: mermaid, plantuml (default: mermaid)")
	rootCmd.Flags().StringVarP(&diagramType, "type", "t", "component", "Diagram type: component (default), graph, sequence")

	helpCmd := &cobra.Command{
		Use:   "help",
		Short: "Show help for gsl-diagram",
		Run: func(cmd *cobra.Command, args []string) {
			_ = rootCmd.Help()
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			cli.PrintVersion("gsl-diagram", Version, Commit, BuildDate)
		},
	}

	// Load guides and build AI command tree
	gslDesc, _, gslErr := cli.LoadGuide(gsl.Guides, "GSL_GUIDE.md")
	_, goGuideContent, _ := cli.LoadGuide(gsl.Guides, "GO_REFERENCE.md")
	seqDesc, seqContent, seqErr := cli.LoadGuide(gsl.Guides, "SEQUENCE_GUIDE.md")

	cli.BuildAICommand(rootCmd, []cli.GuideSpec{
		{Use: "gsl", Desc: gslDesc, Content: goGuideContent, Err: gslErr},
		{Use: "sequence", Desc: seqDesc, Content: seqContent, Err: seqErr},
	}, "AI topics:", "")

	rootCmd.AddCommand(helpCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
