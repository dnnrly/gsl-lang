package gsl_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/cmd/gsl-diagram/formats"
	"github.com/dnnrly/gsl-lang/query"
)

// TestFlagshipModelValidation ensures every flagship model parses cleanly
// with no errors and no warnings (models must declare their sets explicitly).
func TestFlagshipModelValidation(t *testing.T) {
	for _, dir := range flagshipDirs(t) {
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

// TestFlagshipQueries runs every qN-*.gql against its flagship model and
// byte-compares the canonical serialisation against the committed
// qN-*.result.gsl. The canonical GSL output is the deterministic contract.
func TestFlagshipQueries(t *testing.T) {
	queryCount, resultCount := 0, 0
	for _, dir := range flagshipDirs(t) {
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
	t.Logf("executed %d flagship queries against %d committed result files", queryCount, resultCount)
}

// TestFlagshipDiagramConverters feeds every committed query result through
// the Mermaid and PlantUML converters. Diagrams are VIEWS, not a byte
// contract (the converters iterate Go maps), so assertions are lenient:
// conversion succeeds and the output names at least one node of the graph.
func TestFlagshipDiagramConverters(t *testing.T) {
	formatsTypes := []struct {
		format string
		dtype  string
	}{
		{"mermaid", "component"},
		{"mermaid", "graph"},
		{"plantuml", "component"},
	}
	for _, dir := range flagshipDirs(t) {
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

// TestFlagshipDiagramConvertersSequence exercises the sequence dialect on
// edge-dependency pipelines (flagship 03), where scoped blocks map to
// activations.
func TestFlagshipDiagramConvertersSequence(t *testing.T) {
	dir := "flagships/03-release-prerequisites"
	resultBytes := readFlagshipFile(t, dir, "q2-gated-deploy-spine.result.gsl")
	graph, parseErr := gsl.Parse(bytes.NewReader([]byte(resultBytes)))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("failed to parse: %v", parseErr)
	}
	for _, format := range []string{"mermaid", "plantuml"} {
		factory, err := formats.GetFactory(format)
		if err != nil {
			t.Fatalf("GetFactory(%s): %v", format, err)
		}
		out := factory("sequence").Convert(graph)
		if out == "" {
			t.Errorf("%s sequence conversion produced empty output", format)
		}
	}
}

func TestFlagshipREADMEs(t *testing.T) {
	files, err := filepath.Glob("flagships/*/README.md")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if index := "flagships/README.md"; fileExists(index) {
		files = append(files, index)
	}
	if len(files) == 0 {
		t.Fatalf("no flagship README files found")
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
	t.Logf("Validated '%d valid / %d invalid' gsl+gql blocks across %d flagship READMEs", validBlocks, invalidBlocks, len(files))
}

// TestFlagshipIndex asserts the required flagship set is present.
func TestFlagshipIndex(t *testing.T) {
	dirs := flagshipDirs(t)
	if len(dirs) < 5 {
		t.Fatalf("expected at least 5 flagship examples, found %d: %v", len(dirs), dirs)
	}
}

func flagshipDirs(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir("flagships")
	if err != nil {
		t.Fatalf("cannot read flagships dir (run from module root or ./examples): %v", err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			if _, err := os.Stat(filepath.Join("flagships", e.Name(), "model.gsl")); err == nil {
				dirs = append(dirs, filepath.Join("flagships", e.Name()))
			}
		}
	}
	if len(dirs) == 0 {
		t.Fatalf("no flagships found")
	}
	return dirs
}

func readFlagshipFile(t *testing.T, dir, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("failed to read %s: %v", filepath.Join(dir, name), err)
	}
	return string(b)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// diagramNamesNode reports whether the diagram output references any node in
// the graph (a plain, lenient "the view includes real content" check).
func diagramNamesNode(out string, graph *gsl.Graph) bool {
	for id := range graph.GetNodes() {
		if flagshipNodeIDRe(id).MatchString(out) {
			return true
		}
	}
	return false
}

func flagshipNodeIDRe(id string) *regexp.Regexp {
	return regexp.MustCompile(`\b` + regexp.QuoteMeta(id) + `\b`)
}

// ---- README code-block validation (parallel to the root markdown_test) ----

type flagshipBlock struct {
	language   string
	code       string
	lineNumber int
}

var flagshipFenceRe = regexp.MustCompile("(?sm)^```([a-z-]*)\n(.*?)\n```")

func flagshipCodeBlocks(content string) []flagshipBlock {
	var blocks []flagshipBlock
	matches := flagshipFenceRe.FindAllStringSubmatchIndex(content, -1)
	lineNumber := 1
	lastEnd := 0
	for _, m := range matches {
		lineNumber += strings.Count(content[lastEnd:m[0]], "\n")
		language := content[m[2]:m[3]]
		code := content[m[4]:m[5]]
		lastEnd = m[1]
		if language == "gsl" || language == "invalid-gsl" || language == "gql" || language == "invalid-gql" {
			blocks = append(blocks, flagshipBlock{language, code, lineNumber})
		}
		lineNumber += strings.Count(content[m[0]:m[1]], "\n")
	}
	return blocks
}

func validateGSL(code string) error {
	_, parseErr := gsl.Parse(io.NopCloser(strings.NewReader(code)))
	if parseErr != nil && parseErr.HasError() {
		return fmt.Errorf("parse failed: %w", parseErr)
	}
	return nil
}

func validateGSLInvalid(code string) error {
	_, parseErr := gsl.Parse(io.NopCloser(strings.NewReader(code)))
	if parseErr == nil || !parseErr.HasError() {
		return fmt.Errorf("expected parse to fail but it succeeded")
	}
	return nil
}

func validateGQL(code string) error {
	_, err := query.NewQueryParser(code).Parse()
	if err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}
	return nil
}

func validateGQLInvalid(code string) error {
	_, err := query.NewQueryParser(code).Parse()
	if err == nil {
		return fmt.Errorf("expected parse to fail but it succeeded")
	}
	return nil
}