package gsl

import (
	"testing"
)

// testGraph is a helper to construct a Graph for testing.
// It takes maps of nodes, sets, and edges and returns a properly initialized Graph.
// This avoids relying on struct literals with private fields.
func testGraph(nodes map[string]*Node, sets map[string]*Set, edges []*Edge) *Graph {
	g := &Graph{
		nodes: make(map[string]*Node),
		sets:  make(map[string]*Set),
		edges: make([]*Edge, 0),
	}
	if nodes != nil {
		g.nodes = nodes
	}
	if sets != nil {
		g.sets = sets
	}
	if edges != nil {
		g.edges = edges
	}
	return g
}

func TestSerializeEmptyGraph(t *testing.T) {
	g := testGraph(
		map[string]*Node{},
		map[string]*Set{},
		nil,
	)
	got := Serialize(g)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestSerializeNilGraph(t *testing.T) {
	got := Serialize(nil)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestSerializeSingleNode(t *testing.T) {
	g := testGraph(
		map[string]*Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		map[string]*Set{},
		nil,
	)
	got := Serialize(g)
	expected := "node A"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeNodeWithAttrs(t *testing.T) {
	g := testGraph(
		map[string]*Node{
			"A": {
				ID: "A",
				Attributes: map[string]interface{}{
					"flag":   nil,
					"weight": float64(2),
				},
				Sets: map[string]struct{}{},
			},
		},
		map[string]*Set{},
		nil,
	)
	got := Serialize(g)
	expected := "node A [flag, weight=2]"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeNodeWithTextAttr(t *testing.T) {
	g := testGraph(
		map[string]*Node{
			"A": {
				ID:         "A",
				Attributes: map[string]interface{}{"text": "Hello"},
				Sets:       map[string]struct{}{},
			},
		},
		map[string]*Set{},
		nil,
	)
	got := Serialize(g)
	expected := `node A [text="Hello"]`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeNodeWithMembership(t *testing.T) {
	g := testGraph(
		map[string]*Node{
			"A": {
				ID:         "A",
				Attributes: map[string]interface{}{},
				Sets:       map[string]struct{}{"cluster": {}},
			},
		},
		map[string]*Set{
			"cluster": {ID: "cluster", Attributes: map[string]interface{}{}},
		},
		nil,
	)
	got := Serialize(g)
	expected := "set cluster\n\nnode A @cluster"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeNodeWithParent(t *testing.T) {
	parent := "B"
	g := testGraph(
		map[string]*Node{
			"A": {
				ID:         "A",
				Attributes: map[string]interface{}{"parent": NodeRef("B")},
				Sets:       map[string]struct{}{},
				Parent:     &parent,
			},
			"B": {
				ID:         "B",
				Attributes: map[string]interface{}{},
				Sets:       map[string]struct{}{},
			},
		},
		map[string]*Set{},
		nil,
	)
	got := Serialize(g)
	expected := "node B {\n    node A\n}"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeNestedNodeHierarchy(t *testing.T) {
	parentB := "A"
	parentC := "B"
	g := testGraph(
		map[string]*Node{
			"A": {
				ID:         "A",
				Attributes: map[string]interface{}{},
				Sets:       map[string]struct{}{},
			},
			"B": {
				ID:         "B",
				Attributes: map[string]interface{}{"color": "red"},
				Sets:       map[string]struct{}{},
				Parent:     &parentB,
			},
			"C": {
				ID:         "C",
				Attributes: map[string]interface{}{"flag": nil},
				Sets:       map[string]struct{}{"group": {}},
				Parent:     &parentC,
			},
		},
		map[string]*Set{
			"group": {ID: "group", Attributes: map[string]interface{}{}},
		},
		nil,
	)
	got := Serialize(g)
	expected := "set group\n\nnode A {\n    node B [color=\"red\"] {\n        node C [flag] @group\n    }\n}"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeNodeWithParentMissing(t *testing.T) {
	parent := "Missing"
	g := testGraph(
		map[string]*Node{
			"A": {
				ID:         "A",
				Attributes: map[string]interface{}{"parent": NodeRef("Missing")},
				Sets:       map[string]struct{}{},
				Parent:     &parent,
			},
		},
		map[string]*Set{},
		nil,
	)
	got := Serialize(g)
	// Parent doesn't exist in graph, so A outputs at root level with explicit parent
	expected := "node A [parent=Missing]"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeEdge(t *testing.T) {
	g := testGraph(
		map[string]*Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		map[string]*Set{},
		[]*Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
	)
	got := Serialize(g)
	expected := "node A\nnode B\n\nA->B"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeEdgeWithAttrs(t *testing.T) {
	g := testGraph(
		map[string]*Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		map[string]*Set{},
		[]*Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{"weight": 1.2}, Sets: map[string]struct{}{}},
		},
	)
	got := Serialize(g)
	expected := "node A\nnode B\n\nA->B [weight=1.2]"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeEdgeWithMembership(t *testing.T) {
	g := testGraph(
		map[string]*Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		map[string]*Set{
			"flow": {ID: "flow", Attributes: map[string]interface{}{}},
		},
		[]*Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"flow": {}}},
		},
	)
	got := Serialize(g)
	expected := "set flow\n\nnode A\nnode B\n\nA->B @flow"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeEdgeWithNestedScope(t *testing.T) {
	childEdge := &Edge{
		From:       "B",
		To:         "C",
		Parent:     "E1",
		Attributes: map[string]interface{}{},
		Sets:       map[string]struct{}{},
	}
	g := testGraph(
		map[string]*Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		map[string]*Set{},
		[]*Edge{
			{
				From:       "A",
				To:         "B",
				Label:      "E1",
				Attributes: map[string]interface{}{},
				Sets:       map[string]struct{}{},
				Children:   []*Edge{childEdge},
			},
			childEdge,
		},
	)
	got := Serialize(g)
	expected := "node A\nnode B\nnode C\n\nE1: A->B {\n    B->C\n}"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeEdgeWithDeeplyNestedScope(t *testing.T) {
	edgeCD := &Edge{
		From:       "C",
		To:         "D",
		Parent:     "E2",
		Attributes: map[string]interface{}{},
		Sets:       map[string]struct{}{},
	}
	edgeE2 := &Edge{
		From:       "B",
		To:         "C",
		Label:      "E2",
		Parent:     "E1",
		Attributes: map[string]interface{}{},
		Sets:       map[string]struct{}{},
		Children:   []*Edge{edgeCD},
	}
	g := testGraph(
		map[string]*Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"D": {ID: "D", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		map[string]*Set{},
		[]*Edge{
			{
				From:       "A",
				To:         "B",
				Label:      "E1",
				Attributes: map[string]interface{}{},
				Sets:       map[string]struct{}{},
				Children:   []*Edge{edgeE2},
			},
			edgeE2,
			edgeCD,
		},
	)
	got := Serialize(g)
	expected := "node A\nnode B\nnode C\nnode D\n\nE1: A->B {\n    E2: B->C {\n        C->D\n    }\n}"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeEdgeCyclicDependency(t *testing.T) {
	g := testGraph(
		map[string]*Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		map[string]*Set{},
		[]*Edge{
			{
				From:       "A",
				To:         "B",
				Label:      "E1",
				Parent:     "E3",
				Attributes: map[string]interface{}{"parent": "E3"},
				Sets:       map[string]struct{}{},
			},
			{
				From:       "B",
				To:         "C",
				Label:      "E2",
				Parent:     "E1",
				Attributes: map[string]interface{}{"parent": "E1"},
				Sets:       map[string]struct{}{},
			},
			{
				From:       "C",
				To:         "A",
				Label:      "E3",
				Parent:     "E2",
				Attributes: map[string]interface{}{"parent": "E2"},
				Sets:       map[string]struct{}{},
			},
		},
	)
	got := Serialize(g)
	// All three edges are in a cycle, so none is nested.
	// Each appears at root level with explicit parent attribute.
	expected := "node A\nnode B\nnode C\n\nE1: A->B [parent=E3]\nE2: B->C [parent=E1]\nE3: C->A [parent=E2]"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeSet(t *testing.T) {
	g := testGraph(
		map[string]*Node{},
		map[string]*Set{
			"flow": {ID: "flow", Attributes: map[string]interface{}{}},
		},
		nil,
	)
	got := Serialize(g)
	expected := "set flow"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeSetWithAttrs(t *testing.T) {
	g := testGraph(
		map[string]*Node{},
		map[string]*Set{
			"flow": {ID: "flow", Attributes: map[string]interface{}{"color": "blue"}},
		},
		nil,
	)
	got := Serialize(g)
	expected := `set flow [color="blue"]`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeOrdering(t *testing.T) {
	g := testGraph(
		map[string]*Node{
			"Z": {ID: "Z", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		map[string]*Set{
			"beta":  {ID: "beta", Attributes: map[string]interface{}{}},
			"alpha": {ID: "alpha", Attributes: map[string]interface{}{}},
		},
		[]*Edge{
			{From: "Z", To: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "A", To: "Z", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
	)
	got := Serialize(g)
	// Nodes ordered by first edge appearance: Z (edge 0), A (edge 0).
	// Both at edge 0, tiebreak by ID: A then Z.
	expected := "set alpha\nset beta\n\nnode A\nnode Z\n\nZ->A\nA->Z"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeStringEscaping(t *testing.T) {
	g := testGraph(
		map[string]*Node{
			"A": {
				ID:         "A",
				Attributes: map[string]interface{}{"msg": "say \"hi\"\nand\t\\done"},
				Sets:       map[string]struct{}{},
			},
		},
		map[string]*Set{},
		nil,
	)
	got := Serialize(g)
	expected := `node A [msg="say \"hi\"\nand\t\\done"]`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeBoolAttrs(t *testing.T) {
	g := testGraph(
		map[string]*Node{
			"A": {
				ID: "A",
				Attributes: map[string]interface{}{
					"active":  true,
					"deleted": false,
				},
				Sets: map[string]struct{}{},
			},
		},
		map[string]*Set{},
		nil,
	)
	got := Serialize(g)
	expected := "node A [active=true, deleted=false]"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeNumberFormatting(t *testing.T) {
	g := testGraph(
		map[string]*Node{
			"A": {
				ID: "A",
				Attributes: map[string]interface{}{
					"decimal": 1.2,
					"whole":   float64(2),
				},
				Sets: map[string]struct{}{},
			},
		},
		map[string]*Set{},
		nil,
	)
	got := Serialize(g)
	expected := "node A [decimal=1.2, whole=2]"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeGroupedEdgeLeft(t *testing.T) {
	// Multiple edges with same label and same To → left-grouped form
	g := testGraph(
		map[string]*Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		map[string]*Set{},
		[]*Edge{
			{From: "A", To: "Z", Label: "X", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "B", To: "Z", Label: "X", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
	)
	got := Serialize(g)
	// Edges share To=Z, so should be left-grouped: X: A,B->Z
	expected := "node A\nnode B\n\nX: A,B->Z"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeGroupedEdgeRight(t *testing.T) {
	// Multiple edges with same label and same From → right-grouped form
	g := testGraph(
		map[string]*Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		map[string]*Set{},
		[]*Edge{
			{From: "A", To: "B", Label: "X", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "A", To: "C", Label: "X", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
	)
	got := Serialize(g)
	// Edges share From=A, so should be right-grouped: X: A->B,C
	expected := "node A\nnode B\nnode C\n\nX: A->B,C"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeGroupedEdgeDuplicatePairs(t *testing.T) {
	// Duplicate identical edges with same label group to one declaration
	g := testGraph(
		map[string]*Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		map[string]*Set{},
		[]*Edge{
			{From: "A", To: "A", Label: "X", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "A", To: "A", Label: "X", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
	)
	got := Serialize(g)
	// Both edges identical (A→A, label X); both conditions true, left-grouped chosen
	expected := "node A\n\nX: A,A->A"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeGroupedEdgeSingle(t *testing.T) {
	// Single labeled edge should serialize normally (no grouping needed)
	g := testGraph(
		map[string]*Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		map[string]*Set{},
		[]*Edge{
			{From: "A", To: "B", Label: "E1", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
	)
	got := Serialize(g)
	expected := "node A\nnode B\n\nE1: A->B"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
