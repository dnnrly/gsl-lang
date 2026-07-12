//go:build fuzz

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

// loadSequenceFixtureInputs loads all input.gsl files from sequence diagram test fixtures.
func loadSequenceFixtureInputs() []string {
	var inputs []string
	seen := map[string]bool{}

	dir := filepath.Join("testdata", "sequence-diagrams")
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if d.Name() != "input.gsl" {
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
	return inputs
}

func FuzzSequencePlantUML(f *testing.F) {
	// Hand-crafted seeds covering edge cases
	f.Add("node A")
	f.Add("A->B: \"Hello\"")
	f.Add("A->B: \"Hello\" { B->C }")
	f.Add("A->B [arrow=\"async\"]")
	f.Add("A->B [activate=true]")
	f.Add("scope: A->B")
	f.Add("A->B [parent=scope]")
	f.Add("node A [shape=\"actor\"]")
	f.Add("node A [text=\"Display Name\"]")
	f.Add("A->A: \"Self\"")
	f.Add("A->B: \"L1\" { B->C: \"L2\" { C->D: \"L3\" } }")
	f.Add("A->B [arrow=\"return\"]")
	f.Add("A->B [arrow=\"dependency\"]")
	f.Add("A->B [arrow=\"strong\"]")
	f.Add("A->B: \"X\" { B->C C->A }")
	f.Add("")
	f.Add("# comment\nA->B")
	f.Add("node A\nnode B\nnode C\nA->B\nB->C\nC->A")

	// Seed from fixture inputs
	for _, input := range loadSequenceFixtureInputs() {
		f.Add(input)
	}

	f.Fuzz(func(t *testing.T, input string) {
		graph, parseErr := gsl.Parse(bytes.NewReader([]byte(input)))
		if parseErr != nil && parseErr.HasError() {
			return
		}
		if graph == nil {
			return
		}

		factory, err := formats.GetFactory("plantuml")
		if err != nil {
			return
		}
		conv := factory("sequence")
		output := conv.Convert(graph)
		_ = output
	})
}

func FuzzSequenceMermaid(f *testing.F) {
	// Hand-crafted seeds covering edge cases
	f.Add("node A")
	f.Add("A->B: \"Hello\"")
	f.Add("A->B: \"Hello\" { B->C }")
	f.Add("A->B [arrow=\"async\"]")
	f.Add("A->B [activate=true]")
	f.Add("scope: A->B")
	f.Add("A->B [parent=scope]")
	f.Add("node A [shape=\"actor\"]")
	f.Add("node A [text=\"Display Name\"]")
	f.Add("A->A: \"Self\"")
	f.Add("A->B: \"L1\" { B->C: \"L2\" { C->D: \"L3\" } }")
	f.Add("A->B [arrow=\"return\"]")
	f.Add("A->B [arrow=\"dependency\"]")
	f.Add("A->B [arrow=\"strong\"]")
	f.Add("A->B: \"X\" { B->C C->A }")
	f.Add("")
	f.Add("# comment\nA->B")
	f.Add("node A\nnode B\nnode C\nA->B\nB->C\nC->A")
	f.Add("node A [shape=\"database\"]")
	f.Add("node A [shape=\"boundary\"]")
	f.Add("node A [shape=\"queue\"]")

	// Seed from fixture inputs
	for _, input := range loadSequenceFixtureInputs() {
		f.Add(input)
	}

	f.Fuzz(func(t *testing.T, input string) {
		graph, parseErr := gsl.Parse(bytes.NewReader([]byte(input)))
		if parseErr != nil && parseErr.HasError() {
			return
		}
		if graph == nil {
			return
		}

		factory, err := formats.GetFactory("mermaid")
		if err != nil {
			return
		}
		conv := factory("sequence")
		output := conv.Convert(graph)
		_ = output
	})
}

func FuzzSequenceBoth(f *testing.F) {
	// Seeds that exercise both converters equally
	f.Add("node A\nA->B: \"Hello\"")
	f.Add("A->B { B->C }")
	f.Add("A->B [arrow=\"async\"]: \"Fire\"")
	f.Add("scope: A->B\nB->C\nA->D [parent=scope]")
	f.Add("A->B: \"Deep\" { B->C: \"Deeper\" { C->D } }")
	f.Add("A->B [activate=true]\nA->B [arrow=\"return\"]")
	f.Add("node X [shape=\"actor\"]\nnode Y [shape=\"database\"]\nX->Y: \"Query\"")

	for _, input := range loadSequenceFixtureInputs() {
		f.Add(input)
	}

	f.Fuzz(func(t *testing.T, input string) {
		graph, parseErr := gsl.Parse(bytes.NewReader([]byte(input)))
		if parseErr != nil && parseErr.HasError() {
			return
		}
		if graph == nil {
			return
		}

		// Convert with both formats - neither should panic
		for _, format := range []string{"plantuml", "mermaid"} {
			factory, err := formats.GetFactory(format)
			if err != nil {
				continue
			}
			conv := factory("sequence")
			output := conv.Convert(graph)

			// Basic sanity: output should not be empty for non-empty graphs
			if len(output) == 0 && len(graph.GetEdges()) > 0 {
				t.Errorf("empty output from %s converter for input: %q", format, input)
			}
		}
	})
}
