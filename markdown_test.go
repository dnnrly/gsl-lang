package gsl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestMarkdownCodeBlocks validates all code blocks in markdown files.
// - "gsl" blocks must parse successfully
// - "invalid-gsl" blocks must fail to parse
func TestMarkdownCodeBlocks(t *testing.T) {
	markdownFiles, err := filepath.Glob("*.md")
	if err != nil {
		t.Fatalf("failed to find markdown files: %v", err)
	}

	if len(markdownFiles) == 0 {
		t.Fatalf("no markdown files found")
	}

	validBlocks := 0
	invalidBlocks := 0

	for _, mdFile := range markdownFiles {
		content, err := os.ReadFile(mdFile)
		if err != nil {
			t.Errorf("failed to read %s: %v", mdFile, err)
			continue
		}

		// Extract all code blocks
		blocks := extractCodeBlocks(string(content))

		for i, block := range blocks {
			if block.language == "gsl" {
				validBlocks++
				if err := testValidGSL(block.code); err != nil {
					t.Errorf("%s block %d (line %d): %v\nCode:\n%s",
						mdFile, i+1, block.lineNumber, err, block.code)
				}
			} else if block.language == "invalid-gsl" {
				invalidBlocks++
				if err := testInvalidGSL(block.code); err != nil {
					t.Errorf("%s block %d (line %d): expected parse to fail, but got: %v\nCode:\n%s",
						mdFile, i+1, block.lineNumber, err, block.code)
				}
			}
		}
	}

	t.Logf("Validated %d valid GSL blocks and %d invalid GSL blocks", validBlocks, invalidBlocks)

	if validBlocks == 0 && invalidBlocks == 0 {
		t.Fatalf("no gsl or invalid-gsl code blocks found in markdown files")
	}
}

type codeBlock struct {
	language   string
	code       string
	lineNumber int
}

// extractCodeBlocks extracts all code blocks with their language labels from markdown content
func extractCodeBlocks(content string) []codeBlock {
	// Match markdown code fences: ```language\n...code...\n```
	// Using (?sm) for both DOTALL and MULTILINE modes
	re := regexp.MustCompile("(?sm)^```([a-z-]*)\n(.*?)\n```")

	var blocks []codeBlock
	matches := re.FindAllStringSubmatchIndex(content, -1)

	lineNumber := 1
	lastEnd := 0

	for _, match := range matches {
		// Count lines up to this match
		lineNumber += strings.Count(content[lastEnd:match[0]], "\n")

		// match[2:4] is the language capture group
		// match[4:6] is the code capture group
		language := content[match[2]:match[3]]
		code := content[match[4]:match[5]]
		lastEnd = match[1]

		if language == "gsl" || language == "invalid-gsl" {
			blocks = append(blocks, codeBlock{
				language:   language,
				code:       code,
				lineNumber: lineNumber,
			})
		}

		lineNumber += strings.Count(content[match[0]:match[1]], "\n")
	}

	return blocks
}

// testValidGSL ensures the code parses successfully
func testValidGSL(code string) error {
	_, _, err := Parse(io.NopCloser(strings.NewReader(code)))
	if err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}
	return nil
}

// testInvalidGSL ensures the code fails to parse
func testInvalidGSL(code string) error {
	_, _, err := Parse(io.NopCloser(strings.NewReader(code)))
	if err == nil {
		return fmt.Errorf("expected parse to fail but it succeeded")
	}
	return nil
}
