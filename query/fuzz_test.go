//go:build fuzz

package query

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gsl "github.com/dnnrly/gsl-lang"
)

// loadGQLFiles walks the testdata directory and returns unique query.gql contents.
func loadGQLFiles(dirs ...string) []string {
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
			data, err := os.ReadFile(path)
			if err != nil {
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

// loadFirstGraph parses the first graph.gsl it finds in the given directories
// for use as a test graph in execution fuzzing.
func loadFirstGraph(dirs ...string) *gsl.Graph {
	for _, dir := range dirs {
		var graph *gsl.Graph
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if d.Name() == "graph.gsl" && graph == nil {
				data, readErr := os.ReadFile(path)
				if readErr != nil {
					return nil
				}
				parsed, parseErr := gsl.Parse(bytes.NewReader(data))
				if parseErr == nil || !parseErr.HasError() {
					graph = parsed
				}
			}
			return nil
		})
		if graph != nil {
			return graph
		}
	}
	return gsl.NewGraph()
}

func FuzzQueryParse(f *testing.F) {
	f.Add("subgraph node.team == \"payments\"")
	f.Add("from *")
	f.Add("subgraph node.id == \"a\" traverse out 2")
	f.Add("make node.reviewed = true where node.team == \"payments\"")
	f.Add("remove orphans")
	f.Add("remove edge where edge.protocol == \"tcp\"")
	f.Add("remove node.debug where node.debug exists")
	f.Add("collapse into platform where node.team == \"platform\"")
	f.Add("subgraph node in @critical")
	f.Add("subgraph edge parent exists")
	f.Add("subgraph edge.depth == 0")
	f.Add("subgraph node.env exists")
	f.Add("subgraph node.env not exists")
	f.Add("subgraph node.team == \"api\" | make node.reviewed = true where node.team == \"api\" | remove orphans")
	f.Add("(subgraph node.team == \"payments\") as PAY | from * | (subgraph node.team == \"fraud\") as FRAUD | PAY + FRAUD")
	f.Add("subgraph node.id == \"a\" traverse up 1")
	f.Add("subgraph node.id == \"b\" traverse down 1")
	f.Add("subgraph edge.protocol == \"http\" scope")
	f.Add("")
	f.Add("subgraph node.exists = true")
	f.Add("subgraph node.team == \"payments\" | from *")
	f.Add("subgraph node.id == \"root\" traverse out 2")

	// Add all query.gql test fixtures as seed corpus
	for _, input := range loadGQLFiles("testdata") {
		f.Add(input)
	}

	f.Fuzz(func(t *testing.T, input string) {
		parser := NewQueryParser(input)
		query, err := parser.Parse()
		_ = query
		_ = err
	})
}

func FuzzQueryExecute(f *testing.F) {
	testGraph := loadFirstGraph("testdata")

	f.Add("from *")
	f.Add("subgraph node.exists = true")
	f.Add("subgraph node.id == \"a\"")
	f.Add("remove orphans")
	f.Add("subgraph node.id == \"a\" traverse out 2")

	for _, input := range loadGQLFiles("testdata") {
		f.Add(input)
	}

	f.Fuzz(func(t *testing.T, input string) {
		parser := NewQueryParser(input)
		query, err := parser.Parse()
		if err != nil {
			return
		}

		ctx := &QueryContext{
			InputGraph:  testGraph,
			NamedGraphs: make(map[string]*gsl.Graph),
		}
		result, err := query.Execute(ctx)
		_ = result
		_ = err
	})
}
