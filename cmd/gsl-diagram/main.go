package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	gsl "github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/cmd/gsl-diagram/formats"
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
	rootCmd.Flags().StringVarP(&diagramType, "type", "t", "component", "Diagram type (default: component)")

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
			printVersion()
		},
	}

	// Load guides
	_, gslDesc, _, gslErr := loadGuide("LLM_GUIDE.md")
	_, _, goGuideContent, goGuideErr := loadGuide("GO_GUIDE.md")

	// ai command - lists available guides
	aiCmd := &cobra.Command{
		Use:   "ai",
		Short: "Show AI/LLM guides",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("AI topics:")
			fmt.Println()
			if gslErr == nil {
				fmt.Printf("  gsl - %s\n", gslDesc)
			}
		},
	}

	// ai gsl command - prints GO_GUIDE.md
	if goGuideErr == nil {
		aiGSLCmd := &cobra.Command{
			Use:   "gsl",
			Short: gslDesc,
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(goGuideContent)
			},
		}
		aiCmd.AddCommand(aiGSLCmd)
	}

	rootCmd.AddCommand(helpCmd, versionCmd, aiCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Printf("gsl-diagram version %s\n", Version)
	fmt.Printf("Commit: %s\n", Commit)
	
	// Parse BuildDate if it's in a standard format
	if BuildDate != "unknown" {
		if t, err := time.Parse(time.RFC3339, BuildDate); err == nil {
			fmt.Printf("Build Date: %s\n", t.Format("2006-01-02 15:04:05 MST"))
		} else {
			fmt.Printf("Build Date: %s\n", BuildDate)
		}
	} else {
		fmt.Printf("Build Date: %s\n", BuildDate)
	}
	
	// Include Go version info
	if info, ok := debug.ReadBuildInfo(); ok {
		fmt.Printf("Go: %s\n", info.GoVersion)
	}
}

// extractFrontmatter extracts YAML frontmatter from markdown
// Returns name, description, and remaining content
func extractFrontmatter(md string) (name, description, content string) {
	if !strings.HasPrefix(md, "---") {
		return "", "", md
	}

	// Find closing ---
	rest := md[3:]
	endIdx := strings.Index(rest, "---")
	if endIdx == -1 {
		return "", "", md
	}

	frontmatter := rest[:endIdx]
	content = strings.TrimPrefix(rest[endIdx+3:], "\n")

	// Extract name and description from YAML
	for _, line := range strings.Split(frontmatter, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			name = strings.Trim(strings.TrimPrefix(line, "name:"), " \"'")
		} else if strings.HasPrefix(line, "description:") {
			description = strings.Trim(strings.TrimPrefix(line, "description:"), " \"'")
		}
	}

	return name, description, content
}

// loadGuide loads a guide file and extracts its metadata
func loadGuide(filename string) (name, description, content string, err error) {
	data, err := gsl.Guides.ReadFile(filename)
	if err == nil {
		name, description, content = extractFrontmatter(string(data))
		return name, description, content, nil
	}

	return "", "", "", fmt.Errorf("could not find %s: %w", filename, err)
}
