package query

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	gsl "github.com/dnnrly/gsl-lang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixtures(t *testing.T) {
	testdataDir := filepath.Join("testdata")
	err := filepath.WalkDir(testdataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "query.gql" {
			return nil
		}

		testDir := filepath.Dir(path)
		relPath, _ := filepath.Rel(testdataDir, testDir)
		t.Run(relPath, func(t *testing.T) {
			// Read input files
			graphPath := filepath.Join(testDir, "graph.gsl")
			queryPath := filepath.Join(testDir, "query.gql")
			resultPath := filepath.Join(testDir, "result.gsl")

			for _, p := range []string{graphPath, queryPath, resultPath} {
				require.FileExists(t, p, "required file missing")
			}

			graphData, err := os.ReadFile(graphPath)
			require.NoError(t, err, "failed to read graph.gsl")

			graph, parseErr := gsl.Parse(bytes.NewReader(graphData))
			require.False(t, parseErr != nil && parseErr.HasError(), "failed to parse graph.gsl: %v", parseErr)
			require.False(t, parseErr != nil && parseErr.HasWarnings(), "parsing warnings in graph.gsl")

			queryData, err := os.ReadFile(queryPath)
			require.NoError(t, err, "failed to read query.gql")

			resultData, err := os.ReadFile(resultPath)
			require.NoError(t, err, "failed to read result.gsl")

			expectedResult, parseErr := gsl.Parse(bytes.NewReader(resultData))
			require.False(t, parseErr != nil && parseErr.HasError(), "failed to parse result.gsl: %v", parseErr)
			require.False(t, parseErr != nil && parseErr.HasWarnings(), "parsing warnings in result.gsl")

			queryParser := NewQueryParser(string(queryData))
			parsedQuery, err := queryParser.Parse()
			require.NoError(t, err, "failed to parse query")

			ctx := &QueryContext{
				InputGraph:  graph,
				NamedGraphs: make(map[string]*gsl.Graph),
			}
			result, err := parsedQuery.Execute(ctx)
			require.NoError(t, err, "failed to execute query")

			var actualGraph *gsl.Graph
			switch v := result.(type) {
			case GraphValue:
				actualGraph = v.Graph
			default:
				require.Fail(t, "unexpected result type", "got %T", result)
			}

			actual := gsl.Serialize(actualGraph)
			expected := gsl.Serialize(expectedResult)

			assert.Equal(t, expected, actual, "query result mismatch")
		})

		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk testdata directory: %v", err)
	}
}
