//go:build fuzz

package gsl_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gsl "github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/query"
)

// loadGraphQueryPairs walks testdata directories and returns pairs of
// (graph.gsl content, query.gql content) from each fixture directory.
func loadGraphQueryPairs(dirs ...string) [][2]string {
	type pair struct{ graph, query string }
	seen := map[pair]bool{}
	var pairs [][2]string

	for _, dir := range dirs {
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if d.Name() != "query.gql" {
				return nil
			}
			testDir := filepath.Dir(path)
			graphPath := filepath.Join(testDir, "graph.gsl")

			graphData, gErr := os.ReadFile(graphPath)
			queryData, qErr := os.ReadFile(path)
			if gErr != nil || qErr != nil {
				return nil
			}

			graph := strings.TrimSpace(string(graphData))
			query := strings.TrimSpace(string(queryData))
			if graph == "" || query == "" {
				return nil
			}

			p := pair{graph, query}
			if !seen[p] {
				seen[p] = true
				pairs = append(pairs, [2]string{graph, query})
			}
			return nil
		})
	}
	return pairs
}

// loadAllGraphs returns all unique graph.gsl contents.
func loadAllGraphs(dirs ...string) []string {
	var inputs []string
	seen := map[string]bool{}
	for _, dir := range dirs {
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if d.Name() != "graph.gsl" {
				return nil
			}
			data, rErr := os.ReadFile(path)
			if rErr != nil {
				return nil
			}
			content := strings.TrimSpace(string(data))
			if content != "" && !seen[content] {
				seen[content] = true
				inputs = append(inputs, content)
			}
			return nil
		})
	}
	return inputs
}

// loadAllQueries returns all unique query.gql contents.
func loadAllQueries(dirs ...string) []string {
	var inputs []string
	seen := map[string]bool{}
	for _, dir := range dirs {
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if d.Name() != "query.gql" {
				return nil
			}
			data, rErr := os.ReadFile(path)
			if rErr != nil {
				return nil
			}
			content := strings.TrimSpace(string(data))
			if content != "" && !seen[content] {
				seen[content] = true
				inputs = append(inputs, content)
			}
			return nil
		})
	}
	return inputs
}

func FuzzGraphQuery(f *testing.F) {
	// Seed from known-valid graph+query pairs in testdata
	for _, p := range loadGraphQueryPairs("testdata", "query/testdata") {
		f.Add(p[0], p[1])
	}

	// Also seed cross-product of popular graphs x popular queries so the
	// fuzzer can mutate from combinations that are likely interesting.
	graphs := loadAllGraphs("testdata", "query/testdata")
	queries := loadAllQueries("testdata", "query/testdata")

	// Pick a representative subset for the cross-product seed (limit
	// to avoid combinatorial explosion in seed corpus).
	graphSample := graphs
	if len(graphSample) > 10 {
		graphSample = graphSample[:10]
	}
	querySample := queries
	if len(querySample) > 10 {
		querySample = querySample[:10]
	}
	for _, g := range graphSample {
		for _, q := range querySample {
			f.Add(g, q)
		}
	}

	// Minimal inline seeds for fast baseline
	f.Add("node a\na->b", "from *")
	f.Add("node a", "subgraph node.id == \"a\"")
	f.Add("node a\nnode b\na->b", "remove orphans")
	f.Add("node a [x=1]\nnode b [x=2]", "subgraph node.x == \"1\"")

	f.Fuzz(func(t *testing.T, graphGSL string, queryGQL string) {
		// Parse graph
		g, parseErr := gsl.Parse(strings.NewReader(graphGSL))
		if parseErr != nil && parseErr.HasError() {
			return
		}
		if g == nil {
			return
		}

		// Parse query
		parser := query.NewQueryParser(queryGQL)
		q, qErr := parser.Parse()
		if qErr != nil {
			return
		}

		// Execute query against graph
		ctx := &query.QueryContext{
			InputGraph:  g,
			NamedGraphs: make(map[string]*gsl.Graph),
		}
		result, execErr := q.Execute(ctx)
		_ = result
		_ = execErr
	})
}
