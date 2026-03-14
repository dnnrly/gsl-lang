package query

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				require.FileExists(t, path, "required file missing")
			}

			// Parse the input graph
			graphData, err := os.ReadFile(graphPath)
			require.NoError(t, err, "failed to read graph.gsl")

			graph, errs, err := gsl.Parse(bytes.NewReader(graphData))
			require.NoError(t, err, "failed to parse graph.gsl")
			require.Empty(t, errs, "parsing errors in graph.gsl")

			// Read the query
			queryData, err := os.ReadFile(queryPath)
			require.NoError(t, err, "failed to read query.gql")

			// Read expected result
			resultData, err := os.ReadFile(resultPath)
			require.NoError(t, err, "failed to read result.gsl")

			expectedResult, errs, err := gsl.Parse(bytes.NewReader(resultData))
			require.NoError(t, err, "failed to parse result.gsl")
			require.Empty(t, errs, "parsing errors in result.gsl")

			// Parse and execute the query
			queryParser := NewQueryParser(string(queryData))
			parsedQuery, err := queryParser.Parse()
			require.NoError(t, err, "failed to parse query")

			// Execute the query
			ctx := &QueryContext{
				InputGraph:  graph,
				NamedGraphs: make(map[string]*gsl.Graph),
			}
			result, err := parsedQuery.Execute(ctx)
			require.NoError(t, err, "failed to execute query")

			// Extract the graph from the result
			var actualGraph *gsl.Graph
			switch v := result.(type) {
			case GraphValue:
				actualGraph = v.Graph
			default:
				require.Fail(t, "unexpected result type", "got %T", result)
			}

			// Compare serialized versions
			actual := gsl.Serialize(actualGraph)
			expected := gsl.Serialize(expectedResult)

			assert.Equal(t, expected, actual, "query result mismatch")
		})
	}
}
