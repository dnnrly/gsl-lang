package gsl

import (
	"testing"
)

func TestSerializeEmptyGraph(t *testing.T) {
	g := &Graph{
		Nodes: make(map[string]*Node),
		Sets:  make(map[string]*Set),
	}
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
	g := &Graph{
		Nodes: map[string]*Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: make(map[string]*Set),
	}
	got := Serialize(g)
	expected := "node A"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeNodeWithAttrs(t *testing.T) {
	g := &Graph{
		Nodes: map[string]*Node{
			"A": {
				ID: "A",
				Attributes: map[string]interface{}{
					"flag":   nil,
					"weight": float64(2),
				},
				Sets: map[string]struct{}{},
			},
		},
		Sets: make(map[string]*Set),
	}
	got := Serialize(g)
	expected := "node A [flag, weight=2]"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeNodeWithTextAttr(t *testing.T) {
	g := &Graph{
		Nodes: map[string]*Node{
			"A": {
				ID:         "A",
				Attributes: map[string]interface{}{"text": "Hello"},
				Sets:       map[string]struct{}{},
			},
		},
		Sets: make(map[string]*Set),
	}
	got := Serialize(g)
	expected := `node A [text="Hello"]`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeNodeWithMembership(t *testing.T) {
	g := &Graph{
		Nodes: map[string]*Node{
			"A": {
				ID:         "A",
				Attributes: map[string]interface{}{},
				Sets:       map[string]struct{}{"cluster": {}},
			},
		},
		Sets: map[string]*Set{
			"cluster": {ID: "cluster", Attributes: map[string]interface{}{}},
		},
	}
	got := Serialize(g)
	expected := "set cluster\n\nnode A @cluster"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeNodeWithParent(t *testing.T) {
	parent := "B"
	g := &Graph{
		Nodes: map[string]*Node{
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
		Sets: make(map[string]*Set),
	}
	got := Serialize(g)
	expected := "node A [parent=B]\nnode B"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeEdge(t *testing.T) {
	g := &Graph{
		Nodes: map[string]*Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: make(map[string]*Set),
		Edges: []*Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
	}
	got := Serialize(g)
	expected := "node A\nnode B\n\nA->B"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeEdgeWithAttrs(t *testing.T) {
	g := &Graph{
		Nodes: map[string]*Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: make(map[string]*Set),
		Edges: []*Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{"weight": 1.2}, Sets: map[string]struct{}{}},
		},
	}
	got := Serialize(g)
	expected := "node A\nnode B\n\nA->B [weight=1.2]"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeEdgeWithMembership(t *testing.T) {
	g := &Graph{
		Nodes: map[string]*Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*Set{
			"flow": {ID: "flow", Attributes: map[string]interface{}{}},
		},
		Edges: []*Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"flow": {}}},
		},
	}
	got := Serialize(g)
	expected := "set flow\n\nnode A\nnode B\n\nA->B @flow"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeSet(t *testing.T) {
	g := &Graph{
		Nodes: make(map[string]*Node),
		Sets: map[string]*Set{
			"flow": {ID: "flow", Attributes: map[string]interface{}{}},
		},
	}
	got := Serialize(g)
	expected := "set flow"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeSetWithAttrs(t *testing.T) {
	g := &Graph{
		Nodes: make(map[string]*Node),
		Sets: map[string]*Set{
			"flow": {ID: "flow", Attributes: map[string]interface{}{"color": "blue"}},
		},
	}
	got := Serialize(g)
	expected := `set flow [color="blue"]`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeOrdering(t *testing.T) {
	g := &Graph{
		Nodes: map[string]*Node{
			"Z": {ID: "Z", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
		Sets: map[string]*Set{
			"beta":  {ID: "beta", Attributes: map[string]interface{}{}},
			"alpha": {ID: "alpha", Attributes: map[string]interface{}{}},
		},
		Edges: []*Edge{
			{From: "Z", To: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
			{From: "A", To: "Z", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}},
		},
	}
	got := Serialize(g)
	expected := "set alpha\nset beta\n\nnode A\nnode Z\n\nZ->A\nA->Z"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeStringEscaping(t *testing.T) {
	g := &Graph{
		Nodes: map[string]*Node{
			"A": {
				ID:         "A",
				Attributes: map[string]interface{}{"msg": "say \"hi\"\nand\t\\done"},
				Sets:       map[string]struct{}{},
			},
		},
		Sets: make(map[string]*Set),
	}
	got := Serialize(g)
	expected := `node A [msg="say \"hi\"\nand\t\\done"]`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeBoolAttrs(t *testing.T) {
	g := &Graph{
		Nodes: map[string]*Node{
			"A": {
				ID: "A",
				Attributes: map[string]interface{}{
					"active":  true,
					"deleted": false,
				},
				Sets: map[string]struct{}{},
			},
		},
		Sets: make(map[string]*Set),
	}
	got := Serialize(g)
	expected := "node A [active=true, deleted=false]"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeNumberFormatting(t *testing.T) {
	g := &Graph{
		Nodes: map[string]*Node{
			"A": {
				ID: "A",
				Attributes: map[string]interface{}{
					"decimal": 1.2,
					"whole":   float64(2),
				},
				Sets: map[string]struct{}{},
			},
		},
		Sets: make(map[string]*Set),
	}
	got := Serialize(g)
	expected := "node A [decimal=1.2, whole=2]"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
