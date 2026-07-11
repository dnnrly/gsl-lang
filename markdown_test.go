package gsl_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/query"
)

// TestMarkdownCodeBlocks validates all code blocks in markdown files.
// - "gsl" blocks must parse successfully
// - "invalid-gsl" blocks must fail to parse
// - "gql" blocks must parse successfully
// - "invalid-gql" blocks must fail to parse
func TestMarkdownCodeBlocks(t *testing.T) {
	markdownFiles, err := filepath.Glob("*.md")
	if err != nil {
		t.Fatalf("failed to find markdown files: %v", err)
	}

	if len(markdownFiles) == 0 {
		t.Fatalf("no markdown files found")
	}

	validGSLBlocks := 0
	invalidGSLBlocks := 0
	validGQLBlocks := 0
	invalidGQLBlocks := 0

	for _, mdFile := range markdownFiles {
		content, err := os.ReadFile(mdFile)
		if err != nil {
			t.Errorf("failed to read %s: %v", mdFile, err)
			continue
		}

		// Extract all code blocks
		blocks := extractCodeBlocks(string(content))

		for i, block := range blocks {
			switch block.language {
			case "gsl":
				validGSLBlocks++
				if err := testValidGSL(block.code); err != nil {
					t.Errorf("%s block %d (line %d): %v\nCode:\n%s",
						mdFile, i+1, block.lineNumber, err, block.code)
				}
			case "invalid-gsl":
				invalidGSLBlocks++
				if err := testInvalidGSL(block.code); err != nil {
					t.Errorf("%s block %d (line %d): expected parse to fail, but got: %v\nCode:\n%s",
						mdFile, i+1, block.lineNumber, err, block.code)
				}
			case "gql":
				validGQLBlocks++
				if err := testValidGQL(block.code); err != nil {
					t.Errorf("%s block %d (line %d): %v\nCode:\n%s",
						mdFile, i+1, block.lineNumber, err, block.code)
				}
			case "invalid-gql":
				invalidGQLBlocks++
				if err := testInvalidGQL(block.code); err != nil {
					t.Errorf("%s block %d (line %d): expected parse to fail, but got: %v\nCode:\n%s",
						mdFile, i+1, block.lineNumber, err, block.code)
				}
			}
		}
	}

	totalGSL := validGSLBlocks + invalidGSLBlocks
	totalGQL := validGQLBlocks + invalidGQLBlocks
	t.Logf("Validated %d valid GSL blocks and %d invalid GSL blocks", validGSLBlocks, invalidGSLBlocks)
	t.Logf("Validated %d valid GQL blocks and %d invalid GQL blocks", validGQLBlocks, invalidGQLBlocks)

	if totalGSL == 0 && totalGQL == 0 {
		t.Fatalf("no gsl, invalid-gsl, gql, or invalid-gql code blocks found in markdown files")
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

		if language == "gsl" || language == "invalid-gsl" || language == "gql" || language == "invalid-gql" {
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
	_, parseErr := gsl.Parse(io.NopCloser(strings.NewReader(code)))
	if parseErr != nil && parseErr.HasError() {
		return fmt.Errorf("parse failed: %w", parseErr)
	}
	return nil
}

// testInvalidGSL ensures the code fails to parse
func testInvalidGSL(code string) error {
	_, parseErr := gsl.Parse(io.NopCloser(strings.NewReader(code)))
	if parseErr == nil || !parseErr.HasError() {
		return fmt.Errorf("expected parse to fail but it succeeded")
	}
	return nil
}

// testValidGQL ensures the GQL code parses successfully
func testValidGQL(code string) error {
	_, err := query.NewQueryParser(code).Parse()
	if err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}
	return nil
}

// testInvalidGQL ensures the GQL code fails to parse
func testInvalidGQL(code string) error {
	_, err := query.NewQueryParser(code).Parse()
	if err == nil {
		return fmt.Errorf("expected parse to fail but it succeeded")
	}
	return nil
}
