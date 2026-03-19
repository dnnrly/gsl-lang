package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Build information injected by goreleaser
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

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
	// Try to find the file
	execPath, _ := os.Executable()
	cwd, _ := os.Getwd()

	searchPaths := []string{
		// Try relative to executable (for installed binaries)
		filepath.Join(filepath.Dir(execPath), "..", "..", filename),
		filepath.Join(filepath.Dir(execPath), filename),
		// Try current working directory
		filepath.Join(cwd, filename),
		// Try repo structure from cwd
		filename,
	}

	var data []byte
	var lastErr error
	for _, path := range searchPaths {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
		lastErr = err
	}

	if err != nil {
		return "", "", "", fmt.Errorf("could not find %s: %w", filename, lastErr)
	}

	name, description, content = extractFrontmatter(string(data))
	return name, description, content, nil
}

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
			printVersion()
		},
	}

	// Load guides
	_, gslDesc, gslContent, gslErr := loadGuide("LLM_GUIDE.md")
	_, queryDesc, queryContent, queryErr := loadGuide("query/QUERY_AI_GUIDE.md")

	// help ai command - lists available guides
	helpAICmd := &cobra.Command{
		Use:   "ai",
		Short: "Show AI/LLM guides",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Available AI/LLM guides:")
			fmt.Println()
			if gslErr == nil {
				fmt.Printf("  gsl    - %s\n", gslDesc)
			}
			if queryErr == nil {
				fmt.Printf("  query  - %s\n", queryDesc)
			}
			fmt.Println()
			fmt.Println("Run 'gsl-query help ai <guide>' to view a guide")
		},
	}

	// help ai gsl command - prints LLM_GUIDE.md
	if gslErr == nil {
		helpAIGSLCmd := &cobra.Command{
			Use:   "gsl",
			Short: gslDesc,
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(gslContent)
			},
		}
		helpAICmd.AddCommand(helpAIGSLCmd)
	}

	// help ai query command - prints QUERY_AI_GUIDE.md
	if queryErr == nil {
		helpAIQueryCmd := &cobra.Command{
			Use:   "query",
			Short: queryDesc,
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(queryContent)
			},
		}
		helpAICmd.AddCommand(helpAIQueryCmd)
	}

	helpCmd.AddCommand(helpAICmd)
	rootCmd.AddCommand(helpCmd, versionCmd)

	// ai command - synonym for help ai
	aiCmd := &cobra.Command{
		Use:   "ai",
		Short: "Show AI/LLM guides",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Available AI/LLM guides:")
			fmt.Println()
			if gslErr == nil {
				fmt.Printf("  gsl    - %s\n", gslDesc)
			}
			if queryErr == nil {
				fmt.Printf("  query  - %s\n", queryDesc)
			}
			fmt.Println()
			fmt.Println("Run 'gsl-query ai <guide>' to view a guide")
		},
	}

	// ai gsl command
	if gslErr == nil {
		aiGSLCmd := &cobra.Command{
			Use:   "gsl",
			Short: gslDesc,
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(gslContent)
			},
		}
		aiCmd.AddCommand(aiGSLCmd)
	}

	// ai query command
	if queryErr == nil {
		aiQueryCmd := &cobra.Command{
			Use:   "query",
			Short: queryDesc,
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(queryContent)
			},
		}
		aiCmd.AddCommand(aiQueryCmd)
	}

	rootCmd.AddCommand(aiCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Printf("gsl-query version %s\n", Version)
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
