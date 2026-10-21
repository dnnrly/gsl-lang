// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

package dialect

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	gsl "github.com/dnnrly/gsl-lang/implementations/go"
	"gopkg.in/yaml.v3"
)

// conformanceDialectsDir returns the repository-relative path to the shared,
// language-neutral dialect conformance fixtures. It is expressed relative to
// this package's directory (go test runs each package's tests from its own
// source directory), so it resolves to
// <repo-root>/tests/conformance/dialects.
func conformanceDialectsDir() string {
	return filepath.Join("..", "..", "..", "tests", "conformance", "dialects")
}

// expectation is the shared, language-neutral expected outcome for a dialect
// fixture. It is deliberately minimal so that a conformant implementation in
// any language can satisfy it:
//
//   - outcome:    conforming | non-conforming | invalid-dialect
//   - projection: optional; definition -> element -> projected attribute map
//   - violations: optional; the set of (definition, element, location)
//
// All fields except outcome are optional and asserted only when present.
type expectation struct {
	Outcome    string                    `yaml:"outcome"`
	Projection map[string]map[string]any `yaml:"projection"`
	Violations []expectedViolation       `yaml:"violations"`
}

type expectedViolation struct {
	Definition string `yaml:"definition"`
	Element    string `yaml:"element"`
	Location   string `yaml:"location"`
}

func TestConformanceFixtures(t *testing.T) {
	testdataDir := conformanceDialectsDir()

	var fixtures []string
	walkErr := filepath.WalkDir(testdataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "dialect.yaml" {
			fixtures = append(fixtures, filepath.Dir(path))
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("failed to walk conformance fixture directory: %v", walkErr)
	}
	if len(fixtures) == 0 {
		t.Fatalf("no dialect fixtures found under %s", testdataDir)
	}
	sort.Strings(fixtures)

	for _, dir := range fixtures {
		rel, _ := filepath.Rel(testdataDir, dir)
		t.Run(rel, func(t *testing.T) {
			runConformanceFixture(t, dir)
		})
	}
}

func runConformanceFixture(t *testing.T, dir string) {
	t.Helper()

	var exp expectation
	expBytes, err := os.ReadFile(filepath.Join(dir, "expected.yaml"))
	if err != nil {
		t.Fatalf("failed to read expected.yaml: %v", err)
	}
	if err := yaml.Unmarshal(expBytes, &exp); err != nil {
		t.Fatalf("failed to parse expected.yaml: %v", err)
	}
	if exp.Outcome == "" {
		t.Fatal("expected.yaml is missing an outcome")
	}

	d, loadErr := LoadFile(filepath.Join(dir, "dialect.yaml"))

	if exp.Outcome == "invalid-dialect" {
		if loadErr == nil {
			t.Fatal("expected the dialect to be rejected, but it loaded")
		}
		var le *LoadError
		if !errors.As(loadErr, &le) {
			t.Fatalf("expected a *LoadError, got %T: %v", loadErr, loadErr)
		}
		return
	}
	if loadErr != nil {
		t.Fatalf("failed to load dialect: %v", loadErr)
	}

	graphPath := filepath.Join(dir, "graph.gsl")
	graphData, err := os.ReadFile(graphPath)
	if err != nil {
		t.Fatalf("failed to read graph.gsl: %v", err)
	}
	graph, perr := gsl.Parse(strings.NewReader(string(graphData)))
	if perr != nil && perr.HasError() {
		t.Fatalf("failed to parse graph.gsl: %v", perr)
	}
	if perr != nil && perr.HasWarnings() {
		t.Fatalf("graph.gsl parsed with warnings: %v", perr)
	}

	if exp.Projection != nil {
		assertProjection(t, d, graph, exp.Projection)
	}

	report, err := d.Validate(graph)
	if err != nil {
		t.Fatalf("Validate returned an incomplete-validation error: %v", err)
	}
	if got := report.Outcome.String(); got != exp.Outcome {
		t.Fatalf("outcome = %q, want %q (violations: %v)", got, exp.Outcome, report.Violations)
	}

	assertViolations(t, report, exp.Violations)
}

func assertProjection(t *testing.T, d *Dialect, graph *gsl.Graph, want map[string]map[string]any) {
	t.Helper()

	selected, err := d.Match(graph)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	got := map[string]map[string]any{}
	for _, s := range selected {
		if got[s.Definition] == nil {
			got[s.Definition] = map[string]any{}
		}
		got[s.Definition][s.Element] = s.Attributes
	}

	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal actual projection: %v", err)
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal expected projection: %v", err)
	}

	var gotNorm, wantNorm any
	if err := json.Unmarshal(gotJSON, &gotNorm); err != nil {
		t.Fatalf("normalise actual projection: %v", err)
	}
	if err := json.Unmarshal(wantJSON, &wantNorm); err != nil {
		t.Fatalf("normalise expected projection: %v", err)
	}

	if !reflect.DeepEqual(gotNorm, wantNorm) {
		t.Errorf("projection mismatch\n got: %s\nwant: %s", gotJSON, wantJSON)
	}
}

func assertViolations(t *testing.T, report *Report, want []expectedViolation) {
	t.Helper()

	got := make([]string, 0, len(report.Violations))
	for _, v := range report.Violations {
		got = append(got, v.Definition+"\x00"+v.Element+"\x00"+v.InstanceLocation)
	}
	expect := make([]string, 0, len(want))
	for _, v := range want {
		expect = append(expect, v.Definition+"\x00"+v.Element+"\x00"+v.Location)
	}
	sort.Strings(got)
	sort.Strings(expect)

	if !reflect.DeepEqual(got, expect) {
		t.Errorf("violation set mismatch\n got: %v\nwant: %v\nraw: %v", got, expect, report.Violations)
	}
}
