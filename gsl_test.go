package gsl

import (
	"strings"
	"testing"
)

func TestParseSimpleInput(t *testing.T) {
	g, parseErr := Parse(strings.NewReader("node A"))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("unexpected error: %v", parseErr)
	}
	if parseErr != nil && parseErr.HasWarnings() {
		t.Errorf("unexpected warnings: %v", parseErr.Warnings)
	}
	nodes := g.GetNodes()
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if _, ok := nodes["A"]; !ok {
		t.Error("expected node A")
	}
}

func TestParseReadmeExample(t *testing.T) {
	input := `set flow [color="blue"]
node A: "Start" @flow
node B [flag]
A->B [weight=1.2] @flow`

	g, parseErr := Parse(strings.NewReader(input))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("unexpected error: %v", parseErr)
	}
	nodes := g.GetNodes()
	edges := g.GetEdges()
	sets := g.GetSets()
	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodes))
	}
	if len(edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(edges))
	}
	if len(sets) != 1 {
		t.Errorf("expected 1 set, got %d", len(sets))
	}
	a := nodes["A"]
	if a == nil {
		t.Fatal("expected node A")
	}
	if a.Attributes["text"] != "Start" {
		t.Errorf("expected A.text='Start', got %v", a.Attributes["text"])
	}
	if _, ok := a.Sets["flow"]; !ok {
		t.Error("expected A in set flow")
	}
	b := nodes["B"]
	if b == nil {
		t.Fatal("expected node B")
	}
	if b.Attributes["flag"] != nil {
		t.Errorf("expected B.flag=nil, got %v", b.Attributes["flag"])
	}
	e := edges[0]
	if e.From != "A" || e.To != "B" {
		t.Errorf("expected edge A->B, got %s->%s", e.From, e.To)
	}
	if e.Attributes["weight"] != 1.2 {
		t.Errorf("expected weight=1.2, got %v", e.Attributes["weight"])
	}
	if _, ok := e.Sets["flow"]; !ok {
		t.Error("expected edge in set flow")
	}
	s := sets["flow"]
	if s == nil {
		t.Fatal("expected set flow")
	}
	if s.Attributes["color"] != "blue" {
		t.Errorf("expected flow.color='blue', got %v", s.Attributes["color"])
	}
}

func TestRoundTripSimple(t *testing.T) {
	assertRoundTrip(t, "node A")
}

func TestRoundTripWithAttrs(t *testing.T) {
	assertRoundTrip(t, "node A [x=1, flag]")
}

func TestRoundTripWithMembership(t *testing.T) {
	assertRoundTrip(t, "set s1\nnode A @s1")
}

func TestRoundTripEdge(t *testing.T) {
	assertRoundTrip(t, "A->B [w=1.2] @flow")
}

func TestRoundTripGroupedEdge(t *testing.T) {
	input := "A,B->C"
	g1, parseErr := Parse(strings.NewReader(input))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("first parse failed: %v", parseErr)
	}
	canonical := Serialize(g1)
	g2, parseErr := Parse(strings.NewReader(canonical))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("second parse failed: %v", parseErr)
	}
	// Grouped edge expands to 2 edges
	edges1 := g1.GetEdges()
	if len(edges1) != 2 {
		t.Fatalf("expected 2 edges after expansion, got %d", len(edges1))
	}
	assertGraphsEqual(t, g1, g2)
}

func TestRoundTripBlock(t *testing.T) {
	input := "node C { node D }"
	g1, parseErr := Parse(strings.NewReader(input))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("first parse failed: %v", parseErr)
	}
	canonical := Serialize(g1)
	g2, parseErr := Parse(strings.NewReader(canonical))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("second parse failed: %v", parseErr)
	}
	assertGraphsEqual(t, g1, g2)
}

func TestRoundTripTextShorthand(t *testing.T) {
	input := `node A: "Hello"`
	g1, parseErr := Parse(strings.NewReader(input))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("first parse failed: %v", parseErr)
	}
	canonical := Serialize(g1)
	// Should use [text="Hello"] not shorthand
	if strings.Contains(canonical, `:`) {
		t.Errorf("canonical should not use text shorthand, got %q", canonical)
	}
	g2, parseErr := Parse(strings.NewReader(canonical))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("second parse failed: %v", parseErr)
	}
	assertGraphsEqual(t, g1, g2)
}

func TestRoundTripFullProgram(t *testing.T) {
	input := `set flow [color="blue"]
node A: "Start" @flow
node B [flag]
A->B [weight=1.2] @flow`
	assertRoundTrip(t, input)
}

func TestParseWithWarnings(t *testing.T) {
	input := "node A @undeclared"
	g, parseErr := Parse(strings.NewReader(input))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("unexpected error: %v", parseErr)
	}
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	sets := g.GetSets()
	if _, ok := sets["undeclared"]; !ok {
		t.Error("expected implicit set to exist")
	}
	found := false
	if parseErr != nil {
		for _, w := range parseErr.Warnings {
			if strings.Contains(w.Error(), "implicit set") {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected 'implicit set' warning")
	}
}

func TestParseWithErrors(t *testing.T) {
	input := "node [invalid]"
	_, parseErr := Parse(strings.NewReader(input))
	if parseErr == nil || !parseErr.HasError() {
		t.Fatal("expected error for invalid input")
	}
}

// assertRoundTrip parses input, serializes, re-parses, and checks equality.
func assertRoundTrip(t *testing.T, input string) {
	t.Helper()
	g1, parseErr := Parse(strings.NewReader(input))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("first parse failed: %v", parseErr)
	}
	canonical := Serialize(g1)
	g2, parseErr := Parse(strings.NewReader(canonical))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("second parse failed (canonical=%q): %v", canonical, parseErr)
	}
	assertGraphsEqual(t, g1, g2)
}

// assertGraphsEqual checks that two graphs have the same structure.
func assertGraphsEqual(t *testing.T, g1, g2 *Graph) {
	t.Helper()

	nodes1 := g1.GetNodes()
	nodes2 := g2.GetNodes()
	sets1 := g1.GetSets()
	sets2 := g2.GetSets()
	edges1 := g1.GetEdges()
	edges2 := g2.GetEdges()

	// Compare nodes
	if len(nodes1) != len(nodes2) {
		t.Errorf("node count: g1=%d, g2=%d", len(nodes1), len(nodes2))
		return
	}
	for id, n1 := range nodes1 {
		n2, ok := nodes2[id]
		if !ok {
			t.Errorf("node %q in g1 but not g2", id)
			continue
		}
		if len(n1.Attributes) != len(n2.Attributes) {
			t.Errorf("node %q attr count: g1=%d, g2=%d", id, len(n1.Attributes), len(n2.Attributes))
			continue
		}
		for k, v1 := range n1.Attributes {
			v2, ok := n2.Attributes[k]
			if !ok {
				t.Errorf("node %q attr %q in g1 but not g2", id, k)
				continue
			}
			if v1 != v2 {
				t.Errorf("node %q attr %q: g1=%v, g2=%v", id, k, v1, v2)
			}
		}
		if len(n1.Sets) != len(n2.Sets) {
			t.Errorf("node %q set count: g1=%d, g2=%d", id, len(n1.Sets), len(n2.Sets))
		}
		for s := range n1.Sets {
			if _, ok := n2.Sets[s]; !ok {
				t.Errorf("node %q set %q in g1 but not g2", id, s)
			}
		}
	}

	// Compare sets
	if len(sets1) != len(sets2) {
		t.Errorf("set count: g1=%d, g2=%d", len(sets1), len(sets2))
		return
	}
	for id, s1 := range sets1 {
		s2, ok := sets2[id]
		if !ok {
			t.Errorf("set %q in g1 but not g2", id)
			continue
		}
		if len(s1.Attributes) != len(s2.Attributes) {
			t.Errorf("set %q attr count: g1=%d, g2=%d", id, len(s1.Attributes), len(s2.Attributes))
		}
		for k, v1 := range s1.Attributes {
			v2, ok := s2.Attributes[k]
			if !ok {
				t.Errorf("set %q attr %q in g1 but not g2", id, k)
				continue
			}
			if v1 != v2 {
				t.Errorf("set %q attr %q: g1=%v, g2=%v", id, k, v1, v2)
			}
		}
	}

	// Compare edges
	if len(edges1) != len(edges2) {
		t.Errorf("edge count: g1=%d, g2=%d", len(edges1), len(edges2))
		return
	}
	for i := range edges1 {
		e1 := edges1[i]
		e2 := edges2[i]
		if e1.From != e2.From || e1.To != e2.To {
			t.Errorf("edge %d: g1=%s->%s, g2=%s->%s", i, e1.From, e1.To, e2.From, e2.To)
		}
		if len(e1.Attributes) != len(e2.Attributes) {
			t.Errorf("edge %d attr count: g1=%d, g2=%d", i, len(e1.Attributes), len(e2.Attributes))
		}
		for k, v1 := range e1.Attributes {
			v2, ok := e2.Attributes[k]
			if !ok {
				t.Errorf("edge %d attr %q in g1 but not g2", i, k)
				continue
			}
			if v1 != v2 {
				t.Errorf("edge %d attr %q: g1=%v, g2=%v", i, k, v1, v2)
			}
		}
		if len(e1.Sets) != len(e2.Sets) {
			t.Errorf("edge %d set count: g1=%d, g2=%d", i, len(e1.Sets), len(e2.Sets))
		}
		for s := range e1.Sets {
			if _, ok := e2.Sets[s]; !ok {
				t.Errorf("edge %d set %q in g1 but not g2", i, s)
			}
		}
	}
}
