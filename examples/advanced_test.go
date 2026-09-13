package gsl_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/cmd/gsl-diagram/formats"
	"github.com/dnnrly/gsl-lang/query"
)

// TestAdvancedModelValidation ensures every advanced-study model parses
// cleanly with no errors and no warnings (models must declare their sets
// explicitly), matching the flagship contract for node counts.
func TestAdvancedModelValidation(t *testing.T) {
	for _, dir := range advancedDirs(t) {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			modelBytes := readFlagshipFile(t, dir, "model.gsl")
			graph, parseErr := gsl.Parse(bytes.NewReader([]byte(modelBytes)))
			if parseErr != nil && parseErr.HasError() {
				t.Fatalf("model.gsl failed to parse: %v", parseErr)
			}
			if parseErr != nil && parseErr.HasWarnings() {
				t.Errorf("model.gsl produced warnings (should declare all sets explicitly): %v", parseErr.Warnings)
			}
			n := len(graph.GetNodes())
			if n < 5 || n > 40 {
				t.Errorf("unexpected node count %d (aim for a two-minute example: 12-20)", n)
			}
		})
	}
}

// TestAdvancedQueries runs every qN-*.gql against its advanced-study model
// and byte-compares the canonical serialisation against the committed
// qN-*.result.gsl, exactly as the flagship suite does.
func TestAdvancedQueries(t *testing.T) {
	queryCount, resultCount := 0, 0
	for _, dir := range advancedDirs(t) {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			modelBytes := readFlagshipFile(t, dir, "model.gsl")
			graph, parseErr := gsl.Parse(bytes.NewReader([]byte(modelBytes)))
			if parseErr != nil && parseErr.HasError() {
				t.Fatalf("model.gsl failed to parse: %v", parseErr)
			}

			gqlFiles, err := filepath.Glob(filepath.Join(dir, "q*.gql"))
			if err != nil {
				t.Fatalf("glob: %v", err)
			}
			if len(gqlFiles) == 0 {
				t.Fatalf("no q*.gql query files found")
			}

			for _, qf := range gqlFiles {
				resultFile := strings.TrimSuffix(qf, ".gql") + ".result.gsl"
				if _, err := os.Stat(resultFile); err != nil {
					t.Errorf("%s has no matching .result.gsl; regenerate: gsl-query -f %s -i model.gsl -o %s",
						filepath.Base(qf), filepath.Base(qf), filepath.Base(resultFile))
					continue
				}

				queryBytes := readFlagshipFile(t, dir, filepath.Base(qf))
				expected := strings.TrimRight(readFlagshipFile(t, dir, filepath.Base(resultFile)), "\n")

				parser := query.NewQueryParser(string(queryBytes))
				q, perr := parser.Parse()
				if perr != nil {
					t.Errorf("%s failed to parse: %v", filepath.Base(qf), perr)
					continue
				}

				ctx := &query.QueryContext{
					InputGraph:  graph,
					NamedGraphs: make(map[string]*gsl.Graph),
				}
				result, xerr := q.Execute(ctx)
				if xerr != nil {
					t.Errorf("%s failed to execute: %v", filepath.Base(qf), xerr)
					continue
				}

				gv, ok := result.(query.GraphValue)
				if !ok {
					t.Errorf("%s returned unexpected result type %T", filepath.Base(qf), result)
					continue
				}

				got := strings.TrimRight(gsl.Serialize(gv.Graph), "\n")
				if got != expected {
					t.Errorf("%s output mismatch.\n--- got ---\n%s\n--- want ---\n%s",
						filepath.Base(qf), got, expected)
				}
				queryCount++
			}

			// no stale results without a matching query
			resultFiles, err := filepath.Glob(filepath.Join(dir, "q*.result.gsl"))
			if err != nil {
				t.Fatalf("glob: %v", err)
			}
			for _, rf := range resultFiles {
				gqlFile := strings.TrimSuffix(rf, ".result.gsl") + ".gql"
				if _, err := os.Stat(gqlFile); err != nil {
					t.Errorf("%s has no matching .gql file", filepath.Base(rf))
				}
			}
			resultCount += len(resultFiles)
		})
	}
	t.Logf("executed %d advanced queries against %d committed result files", queryCount, resultCount)
}

// TestAdvancedDiagramConverters feeds every committed advanced result
// through the Mermaid and PlantUML converters (the same lenient assertions
// as the flagship suite).
func TestAdvancedDiagramConverters(t *testing.T) {
	formatsTypes := []struct {
		format string
		dtype  string
	}{
		{"mermaid", "component"},
		{"mermaid", "graph"},
		{"plantuml", "component"},
	}
	for _, dir := range advancedDirs(t) {
		resultFiles, err := filepath.Glob(filepath.Join(dir, "q*.result.gsl"))
		if err != nil {
			t.Fatalf("glob: %v", err)
		}
		for _, rf := range resultFiles {
			resultBytes := readFlagshipFile(t, dir, filepath.Base(rf))
			graph, parseErr := gsl.Parse(bytes.NewReader([]byte(resultBytes)))
			if parseErr != nil && parseErr.HasError() {
				t.Fatalf("%s failed to parse: %v", filepath.Base(rf), parseErr)
			}
			for _, ft := range formatsTypes {
				factory, ferr := formats.GetFactory(ft.format)
				if ferr != nil {
					t.Fatalf("GetFactory(%s): %v", ft.format, ferr)
				}
				out := factory(ft.dtype).Convert(graph)
				if !diagramNamesNode(out, graph) {
					t.Errorf("%s as %s/%s produced no recognised node reference", filepath.Base(rf), ft.format, ft.dtype)
				}
			}
		}
	}
}

// TestAdvancedREADMEs validates every gsl/gql/invalid code block in the
// advanced study's READMEs (study index and both parts).
func TestAdvancedREADMEs(t *testing.T) {
	files, err := filepath.Glob("advanced/*/README.md")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	partFiles, err := filepath.Glob("advanced/*/*/README.md")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	files = append(files, partFiles...)
	if len(files) == 0 {
		t.Fatalf("no advanced study README files found")
	}
	validBlocks, invalidBlocks := 0, 0
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			t.Errorf("failed to read %s: %v", f, err)
			continue
		}
		for i, block := range flagshipCodeBlocks(string(content)) {
			switch block.language {
			case "gsl":
				validBlocks++
				if err := validateGSL(block.code); err != nil {
					t.Errorf("%s block %d (line %d): %v", f, i+1, block.lineNumber, err)
				}
			case "invalid-gsl":
				invalidBlocks++
				if err := validateGSLInvalid(block.code); err != nil {
					t.Errorf("%s block %d (line %d): %v", f, i+1, block.lineNumber, err)
				}
			case "gql":
				validBlocks++
				if err := validateGQL(block.code); err != nil {
					t.Errorf("%s block %d (line %d): %v", f, i+1, block.lineNumber, err)
				}
			case "invalid-gql":
				invalidBlocks++
				if err := validateGQLInvalid(block.code); err != nil {
					t.Errorf("%s block %d (line %d): %v", f, i+1, block.lineNumber, err)
				}
			}
		}
	}
	t.Logf("Validated '%d valid / %d invalid' gsl+gql blocks across %d advanced study READMEs", validBlocks, invalidBlocks, len(files))
}

// advancedDirs returns the scenario directories under advanced/ (each study
// contains one directory per scenario, each holding a model.gsl).
func advancedDirs(t *testing.T) []string {
	t.Helper()
	studies, err := os.ReadDir("advanced")
	if err != nil {
		t.Fatalf("cannot read advanced dir (run from module root or ./examples): %v", err)
	}
	var dirs []string
	for _, s := range studies {
		if !s.IsDir() {
			continue
		}
		scenarios, err := os.ReadDir(filepath.Join("advanced", s.Name()))
		if err != nil {
			t.Fatalf("cannot read advanced/%s: %v", s.Name(), err)
		}
		for _, sc := range scenarios {
			if sc.IsDir() {
				if _, err := os.Stat(filepath.Join("advanced", s.Name(), sc.Name(), "model.gsl")); err == nil {
					dirs = append(dirs, filepath.Join("advanced", s.Name(), sc.Name()))
				}
			}
		}
	}
	if len(dirs) == 0 {
		t.Fatalf("no advanced scenarios found")
	}
	return dirs
}