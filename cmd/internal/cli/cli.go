package cli

import (
	"embed"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// PrintVersion prints version information for a CLI tool.
func PrintVersion(tool, version, commit, buildDate string) {
	fmt.Printf("%s version %s\n", tool, version)
	fmt.Printf("Commit: %s\n", commit)

	if buildDate != "unknown" {
		if t, err := time.Parse(time.RFC3339, buildDate); err == nil {
			fmt.Printf("Build Date: %s\n", t.Format("2006-01-02 15:04:05 MST"))
		} else {
			fmt.Printf("Build Date: %s\n", buildDate)
		}
	} else {
		fmt.Printf("Build Date: %s\n", buildDate)
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		fmt.Printf("Go: %s\n", info.GoVersion)
	}
}

// extractFrontmatter extracts YAML frontmatter from markdown.
// Returns name, description, and remaining content.
func extractFrontmatter(md string) (name, description, content string) {
	if !strings.HasPrefix(md, "---") {
		return "", "", md
	}

	rest := md[3:]
	endIdx := strings.Index(rest, "---")
	if endIdx == -1 {
		return "", "", md
	}

	frontmatter := rest[:endIdx]
	content = strings.TrimPrefix(rest[endIdx+3:], "\n")

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

// LoadGuide loads a guide from an embedded filesystem and extracts its
// frontmatter description and content.
func LoadGuide(fs embed.FS, filename string) (desc, content string, err error) {
	data, err := fs.ReadFile(filename)
	if err != nil {
		return "", "", fmt.Errorf("could not find %s: %w", filename, err)
	}

	_, description, body := extractFrontmatter(string(data))
	return description, body, nil
}

// GuideSpec defines a guide for the AI command tree.
type GuideSpec struct {
	Use     string // cobra subcommand name (e.g. "gsl", "query")
	Desc    string // short description from frontmatter
	Content string // full guide content
	Err     error  // load error — guide is skipped if non-nil
}

// BuildAICommand creates an "ai" command with subcommands for each
// successfully loaded guide and adds it to parent.
func BuildAICommand(parent *cobra.Command, guides []GuideSpec, header, hint string) {
	aiCmd := &cobra.Command{
		Use:   "ai",
		Short: "Show AI/LLM guides",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(header)
			fmt.Println()
			for _, g := range guides {
				if g.Err == nil {
					fmt.Printf("  %-6s - %s\n", g.Use, g.Desc)
				}
			}
			if hint != "" {
				fmt.Println()
				fmt.Println(hint)
			}
		},
	}

	for _, g := range guides {
		if g.Err == nil {
			guide := g
			aiCmd.AddCommand(&cobra.Command{
				Use:   guide.Use,
				Short: guide.Desc,
				Run: func(cmd *cobra.Command, args []string) {
					fmt.Println(guide.Content)
				},
			})
		}
	}

	parent.AddCommand(aiCmd)
}
