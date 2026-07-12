package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gsl "github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/cmd/gsl-diagram/formats"
)

func TestFixtures(t *testing.T) {
	testdataDir := filepath.Join("testdata", "sequence-diagrams")

	err := filepath.WalkDir(testdataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "input.gsl" {
			return nil
		}

		testDir := filepath.Dir(path)
		relPath, _ := filepath.Rel(testdataDir, testDir)

		t.Run(relPath, func(t *testing.T) {
			inputPath := filepath.Join(testDir, "input.gsl")
			expectedPath := filepath.Join(testDir, "output.puml")

			requireFileExists(t, inputPath)
			requireFileExists(t, expectedPath)

			inputData, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("failed to read input.gsl: %v", err)
			}

			graph, parseErr := gsl.Parse(bytes.NewReader(inputData))
			if parseErr != nil && parseErr.HasError() {
				t.Fatalf("failed to parse input.gsl: %v", parseErr)
			}

			expectedData, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("failed to read output.puml: %v", err)
			}

			actual := convertSequenceDiagram(graph)
			expected := string(expectedData)

			actualNorm := normalizeDiagramOutput(actual)
			expectedNorm := normalizeDiagramOutput(expected)

			if actualNorm != expectedNorm {
				t.Errorf("converter output mismatch\n--- expected ---\n%s\n--- actual ---\n%s", expected, actual)
			}
		})

		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk testdata directory: %v", err)
	}
}

// convertSequenceDiagram converts a GSL graph to PlantUML sequence diagram syntax.
func convertSequenceDiagram(graph *gsl.Graph) string {
	factory, err := formats.GetFactory("plantuml")
	if err != nil {
		return ""
	}
	conv := factory("sequence")
	return conv.Convert(graph)
}

func requireFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("required file missing: %s", path)
	}
}

func normalizeDiagramOutput(s string) string {
	s = strings.TrimSpace(s)

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	result := strings.Join(lines, "\n")

	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}

	if len(result) > 0 && result[len(result)-1] != '\n' {
		result += "\n"
	}

	return result
}
