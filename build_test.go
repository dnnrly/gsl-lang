package gsl

import (
	"strings"
	"testing"
)

func TestBuildSimpleNode(t *testing.T) {
	prog := &program{
		statements: []statement{
			&nodeDecl{name: "A", line: 1, col: 1},
		},
	}
	g, warns, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warns) != 0 {
		t.Errorf("unexpected warnings: %v", warns)
	}
	nodes := g.GetNodes()
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	n, ok := nodes["A"]
	if !ok {
		t.Fatal("node A not found")
	}
	if n.ID != "A" {
		t.Errorf("expected ID %q, got %q", "A", n.ID)
	}
}

func TestBuildNodeWithAttrs(t *testing.T) {
	prog := &program{
		statements: []statement{
			&nodeDecl{
				name: "A",
				attrs: []attribute{
					{key: "label", value: &attrValue{kind: valueString, strVal: "hello"}},
					{key: "weight", value: &attrValue{kind: valueNumber, numVal: 3.14}},
					{key: "active", value: &attrValue{kind: valueBool, boolVal: true}},
					{key: "ref", value: &attrValue{kind: valueNodeRef, refVal: "B"}},
					{key: "flag", value: nil},
				},
				line: 1, col: 1,
			},
		},
	}
	g, _, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nodes := g.GetNodes()
	n := nodes["A"]
	if n.Attributes["label"] != "hello" {
		t.Errorf("expected label %q, got %v", "hello", n.Attributes["label"])
	}
	if n.Attributes["weight"] != 3.14 {
		t.Errorf("expected weight 3.14, got %v", n.Attributes["weight"])
	}
	if n.Attributes["active"] != true {
		t.Errorf("expected active true, got %v", n.Attributes["active"])
	}
	if n.Attributes["ref"] != NodeRef("B") {
		t.Errorf("expected ref NodeRef(B), got %v", n.Attributes["ref"])
	}
	if n.Attributes["flag"] != nil {
		t.Errorf("expected flag nil, got %v", n.Attributes["flag"])
	}
}

func TestBuildNodeTextShorthand(t *testing.T) {
	text := "Start"
	prog := &program{
		statements: []statement{
			&nodeDecl{name: "A", textValue: &text, line: 1, col: 1},
		},
	}
	g, _, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nodes := g.GetNodes()
	if nodes["A"].Attributes["text"] != "Start" {
		t.Errorf("expected text %q, got %v", "Start", nodes["A"].Attributes["text"])
	}
}

func TestBuildNodeMerging(t *testing.T) {
	prog := &program{
		statements: []statement{
			&nodeDecl{
				name: "A",
				attrs: []attribute{
					{key: "color", value: &attrValue{kind: valueString, strVal: "red"}},
					{key: "size", value: &attrValue{kind: valueNumber, numVal: 10}},
				},
				memberships: []string{"s1"},
				line:        1, col: 1,
			},
			&setDecl{name: "s1", line: 2, col: 1},
			&setDecl{name: "s2", line: 3, col: 1},
			&nodeDecl{
				name: "A",
				attrs: []attribute{
					{key: "color", value: &attrValue{kind: valueString, strVal: "blue"}},
				},
				memberships: []string{"s2"},
				line:        4, col: 1,
			},
		},
	}
	g, _, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nodes := g.GetNodes()
	n := nodes["A"]
	if n.Attributes["color"] != "blue" {
		t.Errorf("expected color %q (last-write-wins), got %v", "blue", n.Attributes["color"])
	}
	if n.Attributes["size"] != float64(10) {
		t.Errorf("expected size 10, got %v", n.Attributes["size"])
	}
	if _, ok := n.Sets["s1"]; !ok {
		t.Error("expected membership in s1")
	}
	if _, ok := n.Sets["s2"]; !ok {
		t.Error("expected membership in s2")
	}
}

func TestBuildBlockDesugaring(t *testing.T) {
	prog := &program{
		statements: []statement{
			&nodeDecl{
				name: "Parent",
				block: []nodeDecl{
					{name: "Child", line: 2, col: 5},
				},
				line: 1, col: 1,
			},
		},
	}
	g, warns, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warns) != 0 {
		t.Errorf("unexpected warnings: %v", warns)
	}
	nodes := g.GetNodes()
	child := nodes["Child"]
	if child == nil {
		t.Fatal("child node not found")
	}
	if child.Attributes["parent"] != NodeRef("Parent") {
		t.Errorf("expected parent NodeRef(Parent), got %v", child.Attributes["parent"])
	}
	if child.Parent == nil || *child.Parent != "Parent" {
		t.Errorf("expected Parent field = %q, got %v", "Parent", child.Parent)
	}
}

func TestBuildBlockParentOverride(t *testing.T) {
	prog := &program{
		statements: []statement{
			&nodeDecl{
				name: "Parent",
				block: []nodeDecl{
					{
						name: "Child",
						attrs: []attribute{
							{key: "parent", value: &attrValue{kind: valueNodeRef, refVal: "Other"}},
						},
						line: 2, col: 5,
					},
				},
				line: 1, col: 1,
			},
		},
	}
	g, warns, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nodes := g.GetNodes()
	child := nodes["Child"]
	if child.Attributes["parent"] != NodeRef("Other") {
		t.Errorf("expected parent NodeRef(Other), got %v", child.Attributes["parent"])
	}
	if child.Parent == nil || *child.Parent != "Other" {
		t.Errorf("expected Parent field = %q, got %v", "Other", child.Parent)
	}
	// Should have a warning about parent override
	found := false
	for _, w := range warns {
		if strings.Contains(w.Error(), "parent override") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'parent override' warning")
	}
}

func TestBuildNestedBlocks(t *testing.T) {
	prog := &program{
		statements: []statement{
			&nodeDecl{
				name: "A",
				block: []nodeDecl{
					{
						name: "B",
						block: []nodeDecl{
							{name: "C", line: 3, col: 9},
						},
						line: 2, col: 5,
					},
				},
				line: 1, col: 1,
			},
		},
	}
	g, _, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nodes := g.GetNodes()
	b := nodes["B"]
	if b.Parent == nil || *b.Parent != "A" {
		t.Errorf("expected B.Parent = %q, got %v", "A", b.Parent)
	}
	c := nodes["C"]
	if c.Parent == nil || *c.Parent != "B" {
		t.Errorf("expected C.Parent = %q, got %v", "B", c.Parent)
	}
}

func TestBuildSimpleEdge(t *testing.T) {
	prog := &program{
		statements: []statement{
			&nodeDecl{name: "A", line: 1, col: 1},
			&nodeDecl{name: "B", line: 2, col: 1},
			&edgeDecl{left: []string{"A"}, right: []string{"B"}, line: 3, col: 1},
		},
	}
	g, _, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	edges := g.GetEdges()
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(edges))
	}
	if edges[0].From != "A" || edges[0].To != "B" {
		t.Errorf("expected A->B, got %s->%s", edges[0].From, edges[0].To)
	}
}

func TestBuildGroupedEdgeLeft(t *testing.T) {
	prog := &program{
		statements: []statement{
			&edgeDecl{left: []string{"A", "B"}, right: []string{"C"}, line: 1, col: 1},
		},
	}
	g, _, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	edges := g.GetEdges()
	if len(edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(edges))
	}
	if edges[0].From != "A" || edges[0].To != "C" {
		t.Errorf("expected A->C, got %s->%s", edges[0].From, edges[0].To)
	}
	if edges[1].From != "B" || edges[1].To != "C" {
		t.Errorf("expected B->C, got %s->%s", edges[1].From, edges[1].To)
	}
}

func TestBuildGroupedEdgeRight(t *testing.T) {
	prog := &program{
		statements: []statement{
			&edgeDecl{left: []string{"C"}, right: []string{"D", "E"}, line: 1, col: 1},
		},
	}
	g, _, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	edges := g.GetEdges()
	if len(edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(edges))
	}
	if edges[0].From != "C" || edges[0].To != "D" {
		t.Errorf("expected C->D, got %s->%s", edges[0].From, edges[0].To)
	}
	if edges[1].From != "C" || edges[1].To != "E" {
		t.Errorf("expected C->E, got %s->%s", edges[1].From, edges[1].To)
	}
}

func TestBuildEdgeTextShorthand(t *testing.T) {
	text := "Next"
	prog := &program{
		statements: []statement{
			&edgeDecl{
				left: []string{"A"}, right: []string{"B"},
				textValue: &text,
				line:      1, col: 1,
			},
		},
	}
	g, _, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	edges := g.GetEdges()
	if edges[0].Attributes["text"] != "Next" {
		t.Errorf("expected text %q, got %v", "Next", edges[0].Attributes["text"])
	}
}

func TestBuildEdgeNodeRefError(t *testing.T) {
	prog := &program{
		statements: []statement{
			&edgeDecl{
				left:  []string{"A"},
				right: []string{"B"},
				attrs: []attribute{
					{key: "target", value: &attrValue{kind: valueNodeRef, refVal: "X"}, line: 1, col: 10},
				},
				line: 1, col: 1,
			},
		},
	}
	_, _, err := buildGraph(prog)
	if err == nil {
		t.Fatal("expected error for NodeRef in edge attr")
	}
	if !strings.Contains(err.Error(), "NodeRef") {
		t.Errorf("expected NodeRef error, got: %v", err)
	}
}

func TestBuildSetDeclaration(t *testing.T) {
	prog := &program{
		statements: []statement{
			&setDecl{
				name: "cluster",
				attrs: []attribute{
					{key: "visible", value: nil},
					{key: "color", value: &attrValue{kind: valueString, strVal: "red"}},
				},
				line: 1, col: 1,
			},
		},
	}
	g, _, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sets := g.GetSets()
	s, ok := sets["cluster"]
	if !ok {
		t.Fatal("set cluster not found")
	}
	if s.Attributes["color"] != "red" {
		t.Errorf("expected color %q, got %v", "red", s.Attributes["color"])
	}
	if s.Attributes["visible"] != nil {
		t.Errorf("expected visible nil, got %v", s.Attributes["visible"])
	}
}

func TestBuildSetMerging(t *testing.T) {
	prog := &program{
		statements: []statement{
			&setDecl{
				name: "s",
				attrs: []attribute{
					{key: "a", value: &attrValue{kind: valueString, strVal: "first"}},
					{key: "b", value: &attrValue{kind: valueNumber, numVal: 1}},
				},
				line: 1, col: 1,
			},
			&setDecl{
				name: "s",
				attrs: []attribute{
					{key: "a", value: &attrValue{kind: valueString, strVal: "second"}},
				},
				line: 2, col: 1,
			},
		},
	}
	g, _, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sets := g.GetSets()
	s := sets["s"]
	if s.Attributes["a"] != "second" {
		t.Errorf("expected a %q (last-write-wins), got %v", "second", s.Attributes["a"])
	}
	if s.Attributes["b"] != float64(1) {
		t.Errorf("expected b 1, got %v", s.Attributes["b"])
	}
}

func TestBuildSetNodeRefError(t *testing.T) {
	prog := &program{
		statements: []statement{
			&setDecl{
				name: "s",
				attrs: []attribute{
					{key: "ref", value: &attrValue{kind: valueNodeRef, refVal: "X"}, line: 1, col: 10},
				},
				line: 1, col: 1,
			},
		},
	}
	_, _, err := buildGraph(prog)
	if err == nil {
		t.Fatal("expected error for NodeRef in set attr")
	}
	if !strings.Contains(err.Error(), "NodeRef") {
		t.Errorf("expected NodeRef error, got: %v", err)
	}
}

func TestBuildImplicitSetCreation(t *testing.T) {
	prog := &program{
		statements: []statement{
			&nodeDecl{
				name:        "A",
				memberships: []string{"undeclared"},
				line:        1, col: 1,
			},
		},
	}
	g, warns, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sets := g.GetSets()
	if _, ok := sets["undeclared"]; !ok {
		t.Fatal("expected implicit set to be created")
	}
	found := false
	for _, w := range warns {
		if strings.Contains(w.Error(), "implicit set") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'implicit set' warning")
	}
}

func TestBuildNodeSetNameCollision(t *testing.T) {
	prog := &program{
		statements: []statement{
			&setDecl{name: "X", line: 1, col: 1},
			&nodeDecl{name: "X", line: 2, col: 1},
		},
	}
	_, warns, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, w := range warns {
		if strings.Contains(w.Error(), "collision") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected name collision warning")
	}
}

func TestBuildEdgeMembership(t *testing.T) {
	prog := &program{
		statements: []statement{
			&setDecl{name: "flow", line: 1, col: 1},
			&edgeDecl{
				left: []string{"A"}, right: []string{"B"},
				memberships: []string{"flow"},
				line:        2, col: 1,
			},
		},
	}
	g, _, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	edges := g.GetEdges()
	if _, ok := edges[0].Sets["flow"]; !ok {
		t.Error("expected edge membership in flow")
	}
}

func TestBuildForwardDeclaredNodes(t *testing.T) {
	prog := &program{
		statements: []statement{
			&edgeDecl{left: []string{"X"}, right: []string{"Y"}, line: 1, col: 1},
		},
	}
	g, _, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nodes := g.GetNodes()
	if _, ok := nodes["X"]; !ok {
		t.Error("expected forward-declared node X")
	}
	if _, ok := nodes["Y"]; !ok {
		t.Error("expected forward-declared node Y")
	}
	if len(nodes["X"].Attributes) != 0 {
		t.Error("expected empty attributes for forward-declared node")
	}
}
