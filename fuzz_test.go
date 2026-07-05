//go:build fuzz

package gsl

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// loadGSLFiles walks the given directories and returns the unique contents
// of all .gsl files (excluding result.gsl which are expected outputs).
func loadGSLFiles(dirs ...string) []string {
	var inputs []string
	seen := map[string]bool{}

	for _, dir := range dirs {
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".gsl") {
				return nil
			}
			// Skip result.gsl files (they are expected outputs, not inputs)
			if strings.HasSuffix(path, "result.gsl") {
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

func FuzzLexer(f *testing.F) {
	f.Add([]byte("node A"))
	f.Add([]byte("A->B"))
	f.Add([]byte("set s [color=\"blue\"]"))
	f.Add([]byte(""))
	f.Add([]byte("node A { node B }"))
	f.Add([]byte("A,B->C,D"))
	f.Add([]byte("node A @s1 @s2"))
	f.Add([]byte(`node A: "Start" @flow`))
	f.Add([]byte("# just a comment"))
	f.Add([]byte(`"unterminated`))
	f.Add([]byte("-"))
	f.Add([]byte(`"hello\"world"`))
	f.Add([]byte("true false"))
	f.Add([]byte("node A # inline comment\nset B"))
	f.Add([]byte("E1: A -> B { B -> C }"))

	for _, input := range loadGSLFiles("examples", "query/testdata") {
		f.Add([]byte(input))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		l, err := NewLexer(bytes.NewReader(data))
		if err != nil {
			return
		}
		tokens := l.Tokenize()
		_ = tokens
	})
}

func FuzzParse(f *testing.F) {
	f.Add("node A")
	f.Add("A->B")
	f.Add("set s [color=\"blue\"]")
	f.Add("")
	f.Add("node A { node B }")
	f.Add("A->B [weight=1.2] @flow")
	f.Add("node A: \"Start\" @flow")
	f.Add("node B [flag]")
	f.Add("A,B->C")
	f.Add("C->D,E")
	f.Add("A->B @flow")
	f.Add("A->B [weight=1.2] @flow")
	f.Add("node A { node B node C }")
	f.Add("E1: A -> B")
	f.Add("A -> B { B -> C }")
	f.Add("E1: A -> B { B -> C }")
	f.Add("A: a -> b { B: b -> c { c -> d } }")
	f.Add("A -> B { node C C -> D }")
	f.Add("A -> B { set flow B -> C }")
	f.Add("A -> B [parent = E1]")
	f.Add("node A [parent=NodeA]")
	f.Add(`node A [key="value"]`)
	f.Add("node A [key=42]")
	f.Add("node A [key=3.14]")
	f.Add("node A [key=true]")
	f.Add("node A [key=false]")
	f.Add("node A [key=1, flag]")
	f.Add("node A [x=1] @s1 @s2")
	f.Add("# comment")
	f.Add("set flow [color=\"blue\"]\nnode A: \"Start\" @flow\nnode B [flag]\nA->B [weight=1.2] @flow")
	f.Add("node A")
	f.Add("node A [color=\"red\", color=\"blue\"]")
	f.Add("A,B -> C,D")
	f.Add("A->B [target=C]")
	f.Add("set colors [primary=Red]")
	f.Add("node node")
	f.Add("node A: \"Start\"\nnode B: \"End\"\nA->B")

	for _, input := range loadGSLFiles("examples", "query/testdata") {
		f.Add(input)
	}

	f.Fuzz(func(t *testing.T, input string) {
		prog, errs := parse(input)
		_ = prog
		_ = errs
	})
}

func FuzzRoundTrip(f *testing.F) {
	f.Add("node A")
	f.Add("A->B")
	f.Add("set s [color=\"blue\"]")
	f.Add("node A { node B }")
	f.Add("A->B [weight=1.2] @flow")
	f.Add("set flow [color=\"blue\"]\nnode A: \"Start\" @flow\nnode B [flag]\nA->B [weight=1.2] @flow")
	f.Add("E1: A -> B")
	f.Add("A -> B { B -> C }")
	f.Add("E1: A -> B { B -> C }")
	f.Add("A: a -> b { B: b -> c { c -> d } }")
	f.Add("E1: A -> B\nB -> C [parent = E1]")
	f.Add("A -> B { node C C -> D }")
	f.Add("A -> B { set flow B -> C }")
	f.Add("node A [x=1] @s1 @s2")
	f.Add("A:A->A,A")

	for _, input := range loadGSLFiles("examples", "query/testdata") {
		f.Add(input)
	}

	f.Fuzz(func(t *testing.T, input string) {
		g, parseErr := Parse(strings.NewReader(input))
		if parseErr != nil && parseErr.HasError() {
			return
		}
		if g == nil {
			return
		}

		canonical := Serialize(g)

		g2, parseErr2 := Parse(strings.NewReader(canonical))
		if parseErr2 != nil && parseErr2.HasError() {
			t.Errorf("round-trip re-parse failed: input=%q canonical=%q err=%v", input, canonical, parseErr2)
		}
		if g2 == nil {
			t.Errorf("nil graph after round-trip: input=%q canonical=%q", input, canonical)
		}
	})
}
