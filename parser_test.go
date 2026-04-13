package gsl

import (
	"testing"
)

func mustParse(t *testing.T, input string) *program {
	t.Helper()
	prog, errs := parse(input)
	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}
	return prog
}

func expectErrors(t *testing.T, input string) []error {
	t.Helper()
	_, errs := parse(input)
	if len(errs) == 0 {
		t.Fatal("expected parse errors, got none")
	}
	return errs
}

// --- Node declarations ---

func TestParseBasicNode(t *testing.T) {
	prog := mustParse(t, "node A")
	if len(prog.statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.statements))
	}
	nd, ok := prog.statements[0].(*nodeDecl)
	if !ok {
		t.Fatalf("expected *nodeDecl, got %T", prog.statements[0])
	}
	if nd.name != "A" {
		t.Errorf("expected name 'A', got %q", nd.name)
	}
	if nd.textValue != nil {
		t.Errorf("expected no text value")
	}
	if len(nd.attrs) != 0 {
		t.Errorf("expected no attrs, got %d", len(nd.attrs))
	}
	if len(nd.memberships) != 0 {
		t.Errorf("expected no memberships, got %d", len(nd.memberships))
	}
}

func TestParseNodeWithAttrs(t *testing.T) {
	prog := mustParse(t, "node B [flag, weight=2]")
	nd := prog.statements[0].(*nodeDecl)
	if nd.name != "B" {
		t.Errorf("expected name 'B', got %q", nd.name)
	}
	if len(nd.attrs) != 2 {
		t.Fatalf("expected 2 attrs, got %d", len(nd.attrs))
	}
	if nd.attrs[0].key != "flag" || nd.attrs[0].value != nil {
		t.Errorf("expected flag attr with no value, got key=%q value=%v", nd.attrs[0].key, nd.attrs[0].value)
	}
	if nd.attrs[1].key != "weight" || nd.attrs[1].value == nil {
		t.Fatalf("expected weight attr with value")
	}
	if nd.attrs[1].value.kind != valueNumber || nd.attrs[1].value.numVal != 2 {
		t.Errorf("expected weight=2, got kind=%d numVal=%f", nd.attrs[1].value.kind, nd.attrs[1].value.numVal)
	}
}

func TestParseNodeTextShorthand(t *testing.T) {
	prog := mustParse(t, `node C: "Hello"`)
	nd := prog.statements[0].(*nodeDecl)
	if nd.name != "C" {
		t.Errorf("expected name 'C', got %q", nd.name)
	}
	if nd.textValue == nil || *nd.textValue != "Hello" {
		t.Errorf("expected text value 'Hello', got %v", nd.textValue)
	}
}

func TestParseNodeMembership(t *testing.T) {
	prog := mustParse(t, "node A @cluster")
	nd := prog.statements[0].(*nodeDecl)
	if len(nd.memberships) != 1 || nd.memberships[0] != "cluster" {
		t.Errorf("expected membership [cluster], got %v", nd.memberships)
	}
}

func TestParseNodeTextShorthandWithMembership(t *testing.T) {
	prog := mustParse(t, `node A: "Start" @flow`)
	nd := prog.statements[0].(*nodeDecl)
	if nd.textValue == nil || *nd.textValue != "Start" {
		t.Errorf("expected text 'Start', got %v", nd.textValue)
	}
	if len(nd.memberships) != 1 || nd.memberships[0] != "flow" {
		t.Errorf("expected membership [flow], got %v", nd.memberships)
	}
}

func TestParseNodeAttrsWithMultipleMemberships(t *testing.T) {
	prog := mustParse(t, "node A [x=1] @s1 @s2")
	nd := prog.statements[0].(*nodeDecl)
	if len(nd.attrs) != 1 {
		t.Fatalf("expected 1 attr, got %d", len(nd.attrs))
	}
	if nd.attrs[0].key != "x" || nd.attrs[0].value.numVal != 1 {
		t.Errorf("expected x=1")
	}
	if len(nd.memberships) != 2 {
		t.Fatalf("expected 2 memberships, got %d", len(nd.memberships))
	}
	if nd.memberships[0] != "s1" || nd.memberships[1] != "s2" {
		t.Errorf("expected [s1, s2], got %v", nd.memberships)
	}
}

// --- Block syntax ---

func TestParseBlockWithChild(t *testing.T) {
	prog := mustParse(t, "node C { node D }")
	nd := prog.statements[0].(*nodeDecl)
	if nd.name != "C" {
		t.Errorf("expected 'C', got %q", nd.name)
	}
	if len(nd.block) != 1 {
		t.Fatalf("expected 1 block child, got %d", len(nd.block))
	}
	if nd.block[0].name != "D" {
		t.Errorf("expected child 'D', got %q", nd.block[0].name)
	}
}

func TestParseNestedBlocks(t *testing.T) {
	prog := mustParse(t, "node A { node B { node C } }")
	nd := prog.statements[0].(*nodeDecl)
	if nd.name != "A" {
		t.Errorf("expected 'A'")
	}
	if len(nd.block) != 1 || nd.block[0].name != "B" {
		t.Fatalf("expected child 'B'")
	}
	if len(nd.block[0].block) != 1 || nd.block[0].block[0].name != "C" {
		t.Fatalf("expected grandchild 'C'")
	}
}

// --- Edge declarations ---

func TestParseSimpleEdge(t *testing.T) {
	prog := mustParse(t, "A->B")
	ed := prog.statements[0].(*edgeDecl)
	if len(ed.left) != 1 || ed.left[0] != "A" {
		t.Errorf("expected left [A], got %v", ed.left)
	}
	if len(ed.right) != 1 || ed.right[0] != "B" {
		t.Errorf("expected right [B], got %v", ed.right)
	}
}

func TestParseEdgeWithAttrs(t *testing.T) {
	prog := mustParse(t, "A->B [weight=1.2]")
	ed := prog.statements[0].(*edgeDecl)
	if len(ed.attrs) != 1 {
		t.Fatalf("expected 1 attr, got %d", len(ed.attrs))
	}
	if ed.attrs[0].key != "weight" || ed.attrs[0].value.numVal != 1.2 {
		t.Errorf("expected weight=1.2")
	}
}

func TestParseEdgeTextShorthand(t *testing.T) {
	prog := mustParse(t, `A->B: "Next"`)
	ed := prog.statements[0].(*edgeDecl)
	if ed.textValue == nil || *ed.textValue != "Next" {
		t.Errorf("expected text 'Next', got %v", ed.textValue)
	}
}

func TestParseEdgeGroupedLeft(t *testing.T) {
	prog := mustParse(t, "A,B->C")
	ed := prog.statements[0].(*edgeDecl)
	if len(ed.left) != 2 || ed.left[0] != "A" || ed.left[1] != "B" {
		t.Errorf("expected left [A, B], got %v", ed.left)
	}
	if len(ed.right) != 1 || ed.right[0] != "C" {
		t.Errorf("expected right [C], got %v", ed.right)
	}
}

func TestParseEdgeGroupedRight(t *testing.T) {
	prog := mustParse(t, "C->D,E")
	ed := prog.statements[0].(*edgeDecl)
	if len(ed.left) != 1 || ed.left[0] != "C" {
		t.Errorf("expected left [C], got %v", ed.left)
	}
	if len(ed.right) != 2 || ed.right[0] != "D" || ed.right[1] != "E" {
		t.Errorf("expected right [D, E], got %v", ed.right)
	}
}

func TestParseEdgeWithMembership(t *testing.T) {
	prog := mustParse(t, "A->B @flow")
	ed := prog.statements[0].(*edgeDecl)
	if len(ed.memberships) != 1 || ed.memberships[0] != "flow" {
		t.Errorf("expected membership [flow], got %v", ed.memberships)
	}
}

func TestParseEdgeAttrsAndMembership(t *testing.T) {
	prog := mustParse(t, "A->B [weight=1.2] @flow")
	ed := prog.statements[0].(*edgeDecl)
	if len(ed.attrs) != 1 {
		t.Fatalf("expected 1 attr")
	}
	if ed.attrs[0].key != "weight" {
		t.Errorf("expected key 'weight'")
	}
	if len(ed.memberships) != 1 || ed.memberships[0] != "flow" {
		t.Errorf("expected membership [flow], got %v", ed.memberships)
	}
}

// --- Set declarations ---

func TestParseBareSet(t *testing.T) {
	prog := mustParse(t, "set flow")
	sd := prog.statements[0].(*setDecl)
	if sd.name != "flow" {
		t.Errorf("expected name 'flow', got %q", sd.name)
	}
	if len(sd.attrs) != 0 {
		t.Errorf("expected no attrs, got %d", len(sd.attrs))
	}
}

func TestParseSetWithAttrs(t *testing.T) {
	prog := mustParse(t, `set flow [color="blue"]`)
	sd := prog.statements[0].(*setDecl)
	if sd.name != "flow" {
		t.Errorf("expected name 'flow', got %q", sd.name)
	}
	if len(sd.attrs) != 1 {
		t.Fatalf("expected 1 attr, got %d", len(sd.attrs))
	}
	if sd.attrs[0].key != "color" || sd.attrs[0].value.strVal != "blue" {
		t.Errorf("expected color=blue")
	}
}

// --- Full program ---

func TestParseFullProgram(t *testing.T) {
	input := `
# Declare sets
set flow [color="blue"]

# Declare nodes
node A: "Start" @flow
node B [flag]

# Declare edges
A->B [weight=1.2] @flow
`
	prog := mustParse(t, input)
	if len(prog.statements) != 4 {
		t.Fatalf("expected 4 statements, got %d", len(prog.statements))
	}

	sd, ok := prog.statements[0].(*setDecl)
	if !ok {
		t.Fatalf("expected setDecl, got %T", prog.statements[0])
	}
	if sd.name != "flow" {
		t.Errorf("expected set name 'flow'")
	}

	nd1, ok := prog.statements[1].(*nodeDecl)
	if !ok {
		t.Fatalf("expected nodeDecl, got %T", prog.statements[1])
	}
	if nd1.name != "A" || nd1.textValue == nil || *nd1.textValue != "Start" {
		t.Errorf("expected node A with text Start")
	}
	if len(nd1.memberships) != 1 || nd1.memberships[0] != "flow" {
		t.Errorf("expected node A membership [flow]")
	}

	nd2, ok := prog.statements[2].(*nodeDecl)
	if !ok {
		t.Fatalf("expected nodeDecl, got %T", prog.statements[2])
	}
	if nd2.name != "B" || len(nd2.attrs) != 1 || nd2.attrs[0].key != "flag" {
		t.Errorf("expected node B with flag attr")
	}

	ed, ok := prog.statements[3].(*edgeDecl)
	if !ok {
		t.Fatalf("expected edgeDecl, got %T", prog.statements[3])
	}
	if ed.left[0] != "A" || ed.right[0] != "B" {
		t.Errorf("expected edge A->B")
	}
	if len(ed.attrs) != 1 || ed.attrs[0].key != "weight" {
		t.Errorf("expected weight attr on edge")
	}
	if len(ed.memberships) != 1 || ed.memberships[0] != "flow" {
		t.Errorf("expected edge membership [flow]")
	}
}

// --- Error cases ---

func TestParseErrorBothSidesGrouped(t *testing.T) {
	errs := expectErrors(t, "A,B->C,D")
	found := false
	for _, e := range errs {
		if contains(e.Error(), "both sides") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'both sides' error, got: %v", errs)
	}
}

func TestParseErrorMissingIdentifierAfterNode(t *testing.T) {
	errs := expectErrors(t, "node [attrs]")
	found := false
	for _, e := range errs {
		if contains(e.Error(), "expected") && contains(e.Error(), "identifier") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing identifier error, got: %v", errs)
	}
}

func TestParseErrorKeywordAsIdentifier(t *testing.T) {
	errs := expectErrors(t, "node node")
	found := false
	for _, e := range errs {
		if contains(e.Error(), "reserved keyword") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected reserved keyword error, got: %v", errs)
	}
}

func TestParseErrorDuplicateAttrs(t *testing.T) {
	errs := expectErrors(t, "node A [key=1, key=2]")
	found := false
	for _, e := range errs {
		if contains(e.Error(), "duplicate attribute") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected duplicate attribute error, got: %v", errs)
	}
}

// --- Attribute values ---

func TestParseAttrValueString(t *testing.T) {
	prog := mustParse(t, `node A [key="value"]`)
	nd := prog.statements[0].(*nodeDecl)
	v := nd.attrs[0].value
	if v.kind != valueString || v.strVal != "value" {
		t.Errorf("expected string value 'value', got kind=%d strVal=%q", v.kind, v.strVal)
	}
}

func TestParseAttrValueInteger(t *testing.T) {
	prog := mustParse(t, "node A [key=42]")
	nd := prog.statements[0].(*nodeDecl)
	v := nd.attrs[0].value
	if v.kind != valueNumber || v.numVal != 42 {
		t.Errorf("expected number 42, got kind=%d numVal=%f", v.kind, v.numVal)
	}
}

func TestParseAttrValueFloat(t *testing.T) {
	prog := mustParse(t, "node A [key=3.14]")
	nd := prog.statements[0].(*nodeDecl)
	v := nd.attrs[0].value
	if v.kind != valueNumber || v.numVal != 3.14 {
		t.Errorf("expected number 3.14, got kind=%d numVal=%f", v.kind, v.numVal)
	}
}

func TestParseAttrValueBoolTrue(t *testing.T) {
	prog := mustParse(t, "node A [key=true]")
	nd := prog.statements[0].(*nodeDecl)
	v := nd.attrs[0].value
	if v.kind != valueBool || v.boolVal != true {
		t.Errorf("expected bool true, got kind=%d boolVal=%v", v.kind, v.boolVal)
	}
}

func TestParseAttrValueBoolFalse(t *testing.T) {
	prog := mustParse(t, "node A [key=false]")
	nd := prog.statements[0].(*nodeDecl)
	v := nd.attrs[0].value
	if v.kind != valueBool || v.boolVal != false {
		t.Errorf("expected bool false, got kind=%d boolVal=%v", v.kind, v.boolVal)
	}
}

func TestParseAttrValueNodeRef(t *testing.T) {
	prog := mustParse(t, "node A [parent=NodeA]")
	nd := prog.statements[0].(*nodeDecl)
	v := nd.attrs[0].value
	if v.kind != valueNodeRef || v.refVal != "NodeA" {
		t.Errorf("expected noderef NodeA, got kind=%d refVal=%q", v.kind, v.refVal)
	}
}

func TestParseNodeRefNotAllowedInEdgeContext(t *testing.T) {
	errs := expectErrors(t, "A->B [ref=SomeNode]")
	found := false
	for _, e := range errs {
		if contains(e.Error(), "node references are not allowed") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected node ref error in edge context, got: %v", errs)
	}
}

func TestParseNodeRefNotAllowedInSetContext(t *testing.T) {
	errs := expectErrors(t, "set flow [ref=SomeNode]")
	found := false
	for _, e := range errs {
		if contains(e.Error(), "node references are not allowed") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected node ref error in set context, got: %v", errs)
	}
}

// --- Empty program ---

func TestParseEmptyProgram(t *testing.T) {
	prog := mustParse(t, "")
	if len(prog.statements) != 0 {
		t.Errorf("expected 0 statements, got %d", len(prog.statements))
	}
}

// --- Multiple block children ---

func TestParseBlockMultipleChildren(t *testing.T) {
	prog := mustParse(t, "node A { node B  node C }")
	nd := prog.statements[0].(*nodeDecl)
	if len(nd.block) != 2 {
		t.Fatalf("expected 2 block children, got %d", len(nd.block))
	}
	if nd.block[0].name != "B" || nd.block[1].name != "C" {
		t.Errorf("expected children B, C; got %q, %q", nd.block[0].name, nd.block[1].name)
	}
}

// --- Edge with text shorthand and membership ---

func TestParseEdgeTextShorthandWithMembership(t *testing.T) {
	prog := mustParse(t, `A->B: "label" @flow`)
	ed := prog.statements[0].(*edgeDecl)
	if ed.textValue == nil || *ed.textValue != "label" {
		t.Errorf("expected text 'label'")
	}
	if len(ed.memberships) != 1 || ed.memberships[0] != "flow" {
		t.Errorf("expected membership [flow], got %v", ed.memberships)
	}
}

// --- Flag attribute (no value) ---

func TestParseAttrFlagNoValue(t *testing.T) {
	prog := mustParse(t, "node A [visible]")
	nd := prog.statements[0].(*nodeDecl)
	if len(nd.attrs) != 1 {
		t.Fatalf("expected 1 attr")
	}
	if nd.attrs[0].key != "visible" || nd.attrs[0].value != nil {
		t.Errorf("expected flag attr 'visible' with nil value")
	}
}

// --- Comments only ---

func TestParseCommentsOnly(t *testing.T) {
	prog := mustParse(t, "# just a comment\n# another one")
	if len(prog.statements) != 0 {
		t.Errorf("expected 0 statements, got %d", len(prog.statements))
	}
}

// helper
// --- Edge labels and scoping ---

func TestParseEdgeLabel(t *testing.T) {
	prog := mustParse(t, "E1: A -> B")
	ed := prog.statements[0].(*edgeDecl)
	if ed.label == nil || *ed.label != "E1" {
		t.Errorf("expected label 'E1', got %v", ed.label)
	}
}

func TestParseScopedEdge(t *testing.T) {
	prog := mustParse(t, "A -> B { B -> C }")
	ed := prog.statements[0].(*edgeDecl)
	if len(ed.block) != 1 {
		t.Fatalf("expected 1 statement in block, got %d", len(ed.block))
	}
	child, ok := ed.block[0].(*edgeDecl)
	if !ok {
		t.Fatalf("expected edgeDecl in block, got %T", ed.block[0])
	}
	if child.left[0] != "B" || child.right[0] != "C" {
		t.Errorf("expected B->C, got %v->%v", child.left, child.right)
	}
}

func TestParseLabeledScopedEdge(t *testing.T) {
	prog := mustParse(t, "E1: A -> B { B -> C }")
	ed := prog.statements[0].(*edgeDecl)
	if ed.label == nil || *ed.label != "E1" {
		t.Errorf("expected label 'E1', got %v", ed.label)
	}
	if len(ed.block) != 1 {
		t.Fatalf("expected 1 statement in block, got %d", len(ed.block))
	}
}

func TestParseNestedScopedEdges(t *testing.T) {
	prog := mustParse(t, "A: a -> b { B: b -> c { c -> d } }")
	outer := prog.statements[0].(*edgeDecl)
	if outer.label == nil || *outer.label != "A" {
		t.Errorf("expected outer label 'A'")
	}
	if len(outer.block) != 1 {
		t.Fatalf("expected 1 statement in outer block")
	}
	inner, ok := outer.block[0].(*edgeDecl)
	if !ok {
		t.Fatalf("expected edgeDecl in outer block")
	}
	if inner.label == nil || *inner.label != "B" {
		t.Errorf("expected inner label 'B'")
	}
	if len(inner.block) != 1 {
		t.Fatalf("expected 1 statement in inner block")
	}
}

func TestParseScopedEdgeWithNodes(t *testing.T) {
	prog := mustParse(t, "A -> B { node C C -> D }")
	ed := prog.statements[0].(*edgeDecl)
	if len(ed.block) != 2 {
		t.Fatalf("expected 2 statements in block, got %d", len(ed.block))
	}
	_, ok1 := ed.block[0].(*nodeDecl)
	if !ok1 {
		t.Errorf("expected nodeDecl at block[0], got %T", ed.block[0])
	}
	_, ok2 := ed.block[1].(*edgeDecl)
	if !ok2 {
		t.Errorf("expected edgeDecl at block[1], got %T", ed.block[1])
	}
}

func TestParseScopedEdgeWithSet(t *testing.T) {
	prog := mustParse(t, "A -> B { set flow B -> C }")
	ed := prog.statements[0].(*edgeDecl)
	if len(ed.block) != 2 {
		t.Fatalf("expected 2 statements in block, got %d", len(ed.block))
	}
	_, ok1 := ed.block[0].(*setDecl)
	if !ok1 {
		t.Errorf("expected setDecl at block[0], got %T", ed.block[0])
	}
	_, ok2 := ed.block[1].(*edgeDecl)
	if !ok2 {
		t.Errorf("expected edgeDecl at block[1], got %T", ed.block[1])
	}
}

func TestParseDependsOn(t *testing.T) {
	prog := mustParse(t, "A -> B [depends_on = E1]")
	ed := prog.statements[0].(*edgeDecl)
	found := false
	for _, attr := range ed.attrs {
		if attr.key == "depends_on" && attr.value != nil && attr.value.strVal == "E1" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected depends_on=E1 attribute")
	}
}

// --- Error cases for edge scoping ---

func TestParseErrorDuplicateLabelInScope(t *testing.T) {
	// This will be caught at build time, not parse time
	// Parsing should succeed
	prog, errs := parse("E1: A -> B { E1: C -> D }")
	if len(errs) > 0 {
		t.Logf("parse errors (may be expected): %v", errs)
	}
	if len(prog.statements) != 1 {
		t.Fatalf("expected 1 statement")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
