package query

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	gsl "github.com/dnnrly/gsl-lang"
)

func TestFixtures(t *testing.T) {
	testdataDir := filepath.Join("testdata")
	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		testName := entry.Name()
		t.Run(testName, func(t *testing.T) {
			testDir := filepath.Join(testdataDir, testName)

			// Read input files
			graphPath := filepath.Join(testDir, "graph.gsl")
			queryPath := filepath.Join(testDir, "query.gql")
			resultPath := filepath.Join(testDir, "result.gsl")

			// Check that all required files exist
			for _, path := range []string{graphPath, queryPath, resultPath} {
				if _, err := os.Stat(path); err != nil {
					t.Fatalf("required file missing: %s", path)
				}
			}

			// Parse the input graph
			graphData, err := os.ReadFile(graphPath)
			if err != nil {
				t.Fatalf("failed to read graph.gsl: %v", err)
			}

			graph, errs, err := gsl.Parse(bytes.NewReader(graphData))
			if err != nil {
				t.Fatalf("failed to parse graph.gsl: %v", err)
			}
			if len(errs) > 0 {
				t.Fatalf("failed to parse graph.gsl: %v", errs)
			}

			// Read the query
			queryData, err := os.ReadFile(queryPath)
			if err != nil {
				t.Fatalf("failed to read query.gql: %v", err)
			}

			// Read expected result
			resultData, err := os.ReadFile(resultPath)
			if err != nil {
				t.Fatalf("failed to read result.gsl: %v", err)
			}

			expectedResult, errs, err := gsl.Parse(bytes.NewReader(resultData))
			if err != nil {
				t.Fatalf("failed to parse result.gsl: %v", err)
			}
			if len(errs) > 0 {
				t.Fatalf("failed to parse result.gsl: %v", errs)
			}

			// TODO: Run the query and compare with expected result
			_ = graph
			_ = queryData
			_ = expectedResult
			t.Logf("Test infrastructure ready. Query test data loaded successfully.")
		})
	}
}
