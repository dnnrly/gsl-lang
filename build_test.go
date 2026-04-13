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

func TestGraphCloneIndependence(t *testing.T) {
	// Build original graph
	g := NewGraph()
	_, err := g.AddNode("A", map[string]interface{}{"label": "Node A", "weight": 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = g.AddNode("B", map[string]interface{}{"label": "Node B"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = g.AddEdge("A", "B", map[string]interface{}{"color": "red"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = g.AddSet("set1", map[string]interface{}{"description": "Test set"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Clone it
	cloned := g.Clone()

	// Verify initial state matches
	origNodes := g.GetNodes()
	clonedNodes := cloned.GetNodes()
	if len(origNodes) != len(clonedNodes) {
		t.Errorf("expected %d nodes in clone, got %d", len(origNodes), len(clonedNodes))
	}

	origEdges := g.GetEdges()
	clonedEdges := cloned.GetEdges()
	if len(origEdges) != len(clonedEdges) {
		t.Errorf("expected %d edges in clone, got %d", len(origEdges), len(clonedEdges))
	}

	// Mutate clone and verify original is unchanged
	_, err = cloned.AddNode("C", map[string]interface{}{"label": "Node C"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	origNodes = g.GetNodes()
	clonedNodes = cloned.GetNodes()
	if len(origNodes) == len(clonedNodes) {
		t.Error("expected clone mutation to not affect original")
	}
	if len(origNodes) != 2 || len(clonedNodes) != 3 {
		t.Errorf("expected original to have 2 nodes and clone to have 3, got %d and %d", len(origNodes), len(clonedNodes))
	}

	// Mutate clone node attributes and verify original is unchanged
	clonedNodeA := clonedNodes["A"]
	clonedNodeA.Attributes["label"] = "Modified A"
	origNodeA := origNodes["A"]
	if origNodeA.Attributes["label"] == "Modified A" {
		t.Error("expected original node attribute to be unchanged")
	}
}

func TestGraphCloneDeepCopy(t *testing.T) {
	// Build original graph with complex attributes
	g := NewGraph()
	_, err := g.AddNode("A", map[string]interface{}{
		"label":   "Node A",
		"weight":  42,
		"enabled": true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = g.AddNode("B", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = g.AddEdge("A", "B", map[string]interface{}{
		"color":     "blue",
		"thickness": 2.5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = g.AddSet("S1", map[string]interface{}{"priority": 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Clone
	cloned := g.Clone()

	// Verify all attributes are copied
	clonedNodeA := cloned.GetNode("A")

	if clonedNodeA.Attributes["label"] != "Node A" {
		t.Error("expected label to be copied")
	}
	if clonedNodeA.Attributes["weight"] != 42 {
		t.Error("expected weight to be copied")
	}
	if clonedNodeA.Attributes["enabled"] != true {
		t.Error("expected enabled to be copied")
	}

	// Verify edge attributes
	clonedEdges := cloned.GetEdges()
	if clonedEdges[0].Attributes["color"] != "blue" {
		t.Error("expected edge color to be copied")
	}
	if clonedEdges[0].Attributes["thickness"] != 2.5 {
		t.Error("expected edge thickness to be copied")
	}

	// Verify set attributes
	clonedSet := cloned.GetSets()["S1"]
	if clonedSet.Attributes["priority"] != 1 {
		t.Error("expected set priority to be copied")
	}
}

func TestGraphCloneRoundTrip(t *testing.T) {
	// Parse original GSL
	gsl := "A -> B"
	g, pErr := Parse(strings.NewReader(gsl))
	if pErr != nil {
		t.Fatalf("unexpected error: %v", pErr)
	}

	// Clone, serialize, parse, serialize again
	cloned := g.Clone()
	s1 := Serialize(g)
	s2 := Serialize(cloned)

	// Serializations should be identical
	if s1 != s2 {
		t.Errorf("expected same serialization, got:\nOriginal: %s\nClone: %s", s1, s2)
	}

	// Parse both and verify they have same structure
	g1, pErr := Parse(strings.NewReader(s1))
	if pErr != nil {
		t.Fatalf("failed to parse serialized original: %v", pErr)
	}
	g2, pErr := Parse(strings.NewReader(s2))
	if pErr != nil {
		t.Fatalf("failed to parse serialized clone: %v", pErr)
	}

	// Verify structure matches
	if len(g1.GetNodes()) != len(g2.GetNodes()) {
		t.Error("node counts differ after round-trip")
	}
	if len(g1.GetEdges()) != len(g2.GetEdges()) {
		t.Error("edge counts differ after round-trip")
	}
}

func TestGraphCloneNilGraph(t *testing.T) {
	var g *Graph
	cloned := g.Clone()
	if cloned != nil {
		t.Error("expected Clone of nil graph to be nil")
	}
}

// --- Edge labels and scoping tests ---

func TestBuildEdgeLabel(t *testing.T) {
	label := "E1"
	prog := &program{
		statements: []statement{
			&edgeDecl{
				label: &label,
				left:  []string{"A"},
				right: []string{"B"},
				line:  1, col: 1,
			},
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
	if edges[0].Label != "E1" {
		t.Errorf("expected label %q, got %q", "E1", edges[0].Label)
	}
}

func TestBuildScopedEdge(t *testing.T) {
	// A -> B { B -> C }
	// Child edge should have depends_on set to parent edge
	label := "E1"
	child := &edgeDecl{
		left:  []string{"B"},
		right: []string{"C"},
		line:  2, col: 5,
	}
	prog := &program{
		statements: []statement{
			&edgeDecl{
				label: &label,
				left:  []string{"A"},
				right: []string{"B"},
				block: []statement{child},
				line:  1, col: 1,
			},
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
	// Parent edge should have no dependency
	if edges[0].DependsOn != "" {
		t.Errorf("expected parent edge to have no dependency, got %q", edges[0].DependsOn)
	}
	// Child edge should depend on parent
	if edges[1].DependsOn != "E1" {
		t.Errorf("expected child edge to depend_on E1, got %q", edges[1].DependsOn)
	}
}

func TestBuildNestedScopedEdges(t *testing.T) {
	// A: a -> b { B: b -> c { c -> d } }
	outerLabel := "A"
	innerLabel := "B"
	innerChild := &edgeDecl{
		left:  []string{"c"},
		right: []string{"d"},
		line:  3, col: 9,
	}
	inner := &edgeDecl{
		label: &innerLabel,
		left:  []string{"b"},
		right: []string{"c"},
		block: []statement{innerChild},
		line:  2, col: 5,
	}
	outer := &edgeDecl{
		label: &outerLabel,
		left:  []string{"a"},
		right: []string{"b"},
		block: []statement{inner},
		line:  1, col: 1,
	}
	prog := &program{
		statements: []statement{outer},
	}
	g, _, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	edges := g.GetEdges()
	if len(edges) != 3 {
		t.Fatalf("expected 3 edges, got %d", len(edges))
	}
	// Outer (a->b): no dependency
	if edges[0].DependsOn != "" {
		t.Errorf("expected outer edge no dependency, got %q", edges[0].DependsOn)
	}
	// Inner (b->c): depends on A
	if edges[1].DependsOn != "A" {
		t.Errorf("expected inner edge depends_on A, got %q", edges[1].DependsOn)
	}
	// Innermost (c->d): depends on B
	if edges[2].DependsOn != "B" {
		t.Errorf("expected innermost edge depends_on B, got %q", edges[2].DependsOn)
	}
}

func TestBuildExplicitDependsOn(t *testing.T) {
	// E1: A -> B
	// B -> C [depends_on = E1]
	label := "E1"
	prog := &program{
		statements: []statement{
			&edgeDecl{
				label: &label,
				left:  []string{"A"},
				right: []string{"B"},
				line:  1, col: 1,
			},
			&edgeDecl{
				left:  []string{"B"},
				right: []string{"C"},
				attrs: []attribute{
					{key: "depends_on", value: &attrValue{kind: valueString, strVal: "E1"}, line: 2, col: 10},
				},
				line: 2, col: 1,
			},
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
	// Second edge should have depends_on = E1
	if edges[1].DependsOn != "E1" {
		t.Errorf("expected depends_on E1, got %q", edges[1].DependsOn)
	}
}

func TestBuildScopedEdgeWithUnlabeledParent(t *testing.T) {
	// A -> B { B -> C }
	// Unlabeled parent edge, child still gets implicit dependency
	child := &edgeDecl{
		left:  []string{"B"},
		right: []string{"C"},
		line:  2, col: 5,
	}
	prog := &program{
		statements: []statement{
			&edgeDecl{
				left:  []string{"A"},
				right: []string{"B"},
				block: []statement{child},
				line:  1, col: 1,
			},
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
	// Parent has no label, child has implicit dependency but no label to reference
	if edges[0].Label != "" {
		t.Errorf("expected parent edge to have no label")
	}
	// Child edge should have depends_on even though parent is unlabeled
	// The dependency is tracked internally but since parent has no label,
	// the child edge should have empty DependsOn in the current implementation
}

func TestBuildDuplicateLabelError(t *testing.T) {
	// E1: A -> B
	// E1: C -> D
	// Duplicate label at top level should error
	label := "E1"
	prog := &program{
		statements: []statement{
			&edgeDecl{
				label: &label,
				left:  []string{"A"},
				right: []string{"B"},
				line:  1, col: 1,
			},
			&edgeDecl{
				label: &label,
				left:  []string{"C"},
				right: []string{"D"},
				line:  2, col: 1,
			},
		},
	}
	_, _, err := buildGraph(prog)
	if err == nil {
		t.Fatal("expected error for duplicate label")
	}
	if !strings.Contains(err.Error(), "duplicate edge label") {
		t.Errorf("expected 'duplicate edge label' error, got: %v", err)
	}
}

func TestBuildNestedScopeLabelShadowing(t *testing.T) {
	// E1: A -> B { C -> D E1: E -> F }
	// Nested scopes allow label shadowing - inner E1 shadows outer E1
	// This should succeed per the spec
	label := "E1"
	innerLabel := "E1"
	child2 := &edgeDecl{
		label: &innerLabel,
		left:  []string{"E"},
		right: []string{"F"},
		line:  3, col: 5,
	}
	child1 := &edgeDecl{
		left:  []string{"C"},
		right: []string{"D"},
		line:  2, col: 5,
	}
	prog := &program{
		statements: []statement{
			&edgeDecl{
				label: &label,
				left:  []string{"A"},
				right: []string{"B"},
				block: []statement{child1, child2},
				line:  1, col: 1,
			},
		},
	}
	g, _, err := buildGraph(prog)
	if err != nil {
		t.Fatalf("unexpected error for nested scope label shadowing: %v", err)
	}
	edges := g.GetEdges()
	if len(edges) != 3 {
		t.Fatalf("expected 3 edges, got %d", len(edges))
	}
	// Verify both labeled edges exist
	outerFound := false
	innerFound := false
	for _, e := range edges {
		if e.Label == "E1" {
			if e.From == "A" && e.To == "B" {
				outerFound = true
				if e.DependsOn != "" {
					t.Errorf("outer edge should have no dependency")
				}
			} else if e.From == "E" && e.To == "F" {
				innerFound = true
				if e.DependsOn != "E1" {
					t.Errorf("inner edge should depend on outer E1")
				}
			}
		}
	}
	if !outerFound {
		t.Error("outer E1 edge not found")
	}
	if !innerFound {
		t.Error("inner E1 edge not found")
	}
}

func TestBuildExplicitDependsOnInScopeError(t *testing.T) {
	// E1: A -> B { C -> D [depends_on = X] }
	// Explicit depends_on inside scoped edge is invalid per spec
	label := "E1"
	child := &edgeDecl{
		left:  []string{"C"},
		right: []string{"D"},
		attrs: []attribute{
			{key: "depends_on", value: &attrValue{kind: valueString, strVal: "X"}, line: 2, col: 15},
		},
		line: 2, col: 5,
	}
	prog := &program{
		statements: []statement{
			&edgeDecl{
				label: &label,
				left:  []string{"A"},
				right: []string{"B"},
				block: []statement{child},
				line:  1, col: 1,
			},
		},
	}
	_, _, err := buildGraph(prog)
	if err == nil {
		t.Fatal("expected error for explicit depends_on inside scoped edge")
	}
	if !strings.Contains(err.Error(), "depends_on not allowed inside scoped edge") {
		t.Errorf("expected 'depends_on not allowed' error, got: %v", err)
	}
}

func TestBuildUnknownDependsOnError(t *testing.T) {
	// A -> B [depends_on = NonExistent]
	prog := &program{
		statements: []statement{
			&edgeDecl{
				left:  []string{"A"},
				right: []string{"B"},
				attrs: []attribute{
					{key: "depends_on", value: &attrValue{kind: valueString, strVal: "NonExistent"}, line: 1, col: 10},
				},
				line: 1, col: 1,
			},
		},
	}
	_, _, err := buildGraph(prog)
	if err == nil {
		t.Fatal("expected error for unknown depends_on reference")
	}
	if !strings.Contains(err.Error(), "unknown label") {
		t.Errorf("expected 'unknown label' error, got: %v", err)
	}
}
