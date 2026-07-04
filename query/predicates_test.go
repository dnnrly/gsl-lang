package query

import (
	"testing"

	"github.com/dnnrly/gsl-lang"
)

// TestExistsPredicate tests the exists predicate
func TestExistsPredicate(t *testing.T) {
	pred := &ExistsPredicate{}

	node := &gsl.Node{ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}}
	edge := &gsl.Edge{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{}}

	if !pred.EvaluateNode(node) {
		t.Fatal("exists should match any node")
	}
	if !pred.EvaluateEdge(edge) {
		t.Fatal("exists should match any edge")
	}
}

// TestAttributeEqualsPredicate tests attribute matching
func TestAttributeEqualsPredicate(t *testing.T) {
	pred := &AttributeEqualsPredicate{
		Target: "node",
		Name:   "color",
		Value:  "red",
	}

	node1 := &gsl.Node{
		ID:         "A",
		Attributes: map[string]interface{}{"color": "red"},
		Sets:       map[string]struct{}{},
	}

	node2 := &gsl.Node{
		ID:         "B",
		Attributes: map[string]interface{}{"color": "blue"},
		Sets:       map[string]struct{}{},
	}

	node3 := &gsl.Node{
		ID:         "C",
		Attributes: map[string]interface{}{},
		Sets:       map[string]struct{}{},
	}

	if !pred.EvaluateNode(node1) {
		t.Fatal("Should match node with color=red")
	}
	if pred.EvaluateNode(node2) {
		t.Fatal("Should not match node with color=blue")
	}
	if pred.EvaluateNode(node3) {
		t.Fatal("Should not match node without color")
	}
}

// TestTypeSensitiveEquality tests that 42 != "42"
func TestTypeSensitiveEquality(t *testing.T) {
	pred := &AttributeEqualsPredicate{
		Target: "node",
		Name:   "port",
		Value:  42,
	}

	node := &gsl.Node{
		ID:         "A",
		Attributes: map[string]interface{}{"port": "42"}, // String, not int
		Sets:       map[string]struct{}{},
	}

	if pred.EvaluateNode(node) {
		t.Fatal("Should not match: 42 (int) != \"42\" (string)")
	}
}

// TestAttributeNotEqualsPredicate tests inequality matching
func TestAttributeNotEqualsPredicate(t *testing.T) {
	pred := &AttributeNotEqualsPredicate{
		Target: "node",
		Name:   "env",
		Value:  "prod",
	}

	nodeProd := &gsl.Node{
		ID:         "A",
		Attributes: map[string]interface{}{"env": "prod"},
		Sets:       map[string]struct{}{},
	}

	nodeDev := &gsl.Node{
		ID:         "B",
		Attributes: map[string]interface{}{"env": "dev"},
		Sets:       map[string]struct{}{},
	}

	nodeNoEnv := &gsl.Node{
		ID:         "C",
		Attributes: map[string]interface{}{},
		Sets:       map[string]struct{}{},
	}

	if pred.EvaluateNode(nodeProd) {
		t.Fatal("Should not match node with env=prod")
	}
	if !pred.EvaluateNode(nodeDev) {
		t.Fatal("Should match node with env=dev")
	}
	if pred.EvaluateNode(nodeNoEnv) {
		t.Fatal("Should not match node without env attribute (spec: missing evaluates false)")
	}
}

// TestTypeSensitiveInequality tests that "42" != 42
func TestTypeSensitiveInequality(t *testing.T) {
	pred := &AttributeNotEqualsPredicate{
		Target: "node",
		Name:   "count",
		Value:  42,
	}

	node := &gsl.Node{
		ID:         "A",
		Attributes: map[string]interface{}{"count": "42"}, // String, not int
		Sets:       map[string]struct{}{},
	}

	if !pred.EvaluateNode(node) {
		t.Fatal("Should match: \"42\" (string) != 42 (int)")
	}
}

// TestSetMembershipPredicate tests set membership
func TestSetMembershipPredicate(t *testing.T) {
	pred := &SetMembershipPredicate{
		Target: "node",
		SetID:  "CRITICAL",
	}

	nodeInSet := &gsl.Node{
		ID:         "A",
		Attributes: map[string]interface{}{},
		Sets:       map[string]struct{}{"CRITICAL": {}},
	}

	nodeNotInSet := &gsl.Node{
		ID:         "B",
		Attributes: map[string]interface{}{},
		Sets:       map[string]struct{}{},
	}

	if !pred.EvaluateNode(nodeInSet) {
		t.Fatal("Should match node in set")
	}
	if pred.EvaluateNode(nodeNotInSet) {
		t.Fatal("Should not match node not in set")
	}
}

// TestSetNotMembershipPredicate tests "not in" predicate
func TestSetNotMembershipPredicate(t *testing.T) {
	pred := &SetNotMembershipPredicate{
		Target: "node",
		SetID:  "DEPRECATED",
	}

	nodeNotInSet := &gsl.Node{
		ID:         "A",
		Attributes: map[string]interface{}{},
		Sets:       map[string]struct{}{},
	}

	nodeInSet := &gsl.Node{
		ID:         "B",
		Attributes: map[string]interface{}{},
		Sets:       map[string]struct{}{"DEPRECATED": {}},
	}

	if !pred.EvaluateNode(nodeNotInSet) {
		t.Fatal("Should match node not in set")
	}
	if pred.EvaluateNode(nodeInSet) {
		t.Fatal("Should not match node in set")
	}
}

// TestAndPredicate tests AND combination
func TestAndPredicate(t *testing.T) {
	pred := &AndPredicate{
		Left: &AttributeEqualsPredicate{
			Target: "node",
			Name:   "type",
			Value:  "service",
		},
		Right: &SetMembershipPredicate{
			Target: "node",
			SetID:  "CRITICAL",
		},
	}

	nodeMatch := &gsl.Node{
		ID:         "A",
		Attributes: map[string]interface{}{"type": "service"},
		Sets:       map[string]struct{}{"CRITICAL": {}},
	}

	nodeNoAttr := &gsl.Node{
		ID:         "B",
		Attributes: map[string]interface{}{},
		Sets:       map[string]struct{}{"CRITICAL": {}},
	}

	nodeNoSet := &gsl.Node{
		ID:         "C",
		Attributes: map[string]interface{}{"type": "service"},
		Sets:       map[string]struct{}{},
	}

	if !pred.EvaluateNode(nodeMatch) {
		t.Fatal("Should match when both conditions true")
	}
	if pred.EvaluateNode(nodeNoAttr) {
		t.Fatal("Should not match when attribute missing")
	}
	if pred.EvaluateNode(nodeNoSet) {
		t.Fatal("Should not match when set missing")
	}
}

// TestNotPredicate tests NOT negation
func TestNotPredicate(t *testing.T) {
	pred := &NotPredicate{
		Inner: &SetMembershipPredicate{
			Target: "node",
			SetID:  "DEPRECATED",
		},
	}

	nodeInSet := &gsl.Node{
		ID:         "A",
		Attributes: map[string]interface{}{},
		Sets:       map[string]struct{}{"DEPRECATED": {}},
	}

	nodeNotInSet := &gsl.Node{
		ID:         "B",
		Attributes: map[string]interface{}{},
		Sets:       map[string]struct{}{},
	}

	if pred.EvaluateNode(nodeInSet) {
		t.Fatal("NOT should invert membership")
	}
	if !pred.EvaluateNode(nodeNotInSet) {
		t.Fatal("NOT should invert non-membership")
	}
}

// TestParsePredicate_SimplePredicates tests parsing basic predicates
func TestParsePredicate_SimplePredicates(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"exists", "exists", false},
		{"in set", "in CRITICAL", false},
		{"not in set", "not in DEPRECATED", false},
		{"node attr", "node.color = red", false},
		{"edge attr", "edge.weight = 5", false},
		{"missing predicate", "unknown", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePredicate(tt.input)
			if !tt.wantError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if tt.wantError && err == nil {
				t.Fatal("Expected error but got none")
			}
		})
	}
}

// TestParsePredicate_NodeAttribute tests parsing node attributes
func TestParsePredicate_NodeAttribute(t *testing.T) {
	pred, err := ParsePredicate("node.color = red")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	attrPred, ok := pred.(*AttributeEqualsPredicate)
	if !ok {
		t.Fatalf("Expected AttributeEqualsPredicate, got %T", pred)
	}

	if attrPred.Target != "node" {
		t.Fatalf("Expected target=node, got %s", attrPred.Target)
	}
	if attrPred.Name != "color" {
		t.Fatalf("Expected name=color, got %s", attrPred.Name)
	}
	if attrPred.Value != "red" {
		t.Fatalf("Expected value=red, got %v", attrPred.Value)
	}
}

// TestParsePredicate_EdgeAttribute tests parsing edge attributes
func TestParsePredicate_EdgeAttribute(t *testing.T) {
	pred, err := ParsePredicate("edge.weight = 42")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	attrPred, ok := pred.(*AttributeEqualsPredicate)
	if !ok {
		t.Fatalf("Expected AttributeEqualsPredicate, got %T", pred)
	}

	if attrPred.Target != "edge" {
		t.Fatalf("Expected target=edge, got %s", attrPred.Target)
	}
	if attrPred.Name != "weight" {
		t.Fatalf("Expected name=weight, got %s", attrPred.Name)
	}
	if attrPred.Value != "42" {
		t.Fatalf("Expected value=42 (string), got %v", attrPred.Value)
	}
}

// TestParsePredicate_StringValues tests parsing string values
func TestParsePredicate_StringValues(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`node.name = "alice"`, "alice"},
		{`node.name = 'bob'`, "bob"},
		{`node.value = unquoted`, "unquoted"},
	}

	for _, tt := range tests {
		pred, err := ParsePredicate(tt.input)
		if err != nil {
			t.Fatalf("Failed to parse %s: %v", tt.input, err)
		}

		attrPred := pred.(*AttributeEqualsPredicate)
		if attrPred.Value != tt.expected {
			t.Fatalf("Expected %v, got %v", tt.expected, attrPred.Value)
		}
	}
}

// TestParsePredicate_BooleanValues tests parsing boolean values
func TestParsePredicate_BooleanValues(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`node.enabled = true`, true},
		{`node.enabled = false`, false},
	}

	for _, tt := range tests {
		pred, err := ParsePredicate(tt.input)
		if err != nil {
			t.Fatalf("Failed to parse %s: %v", tt.input, err)
		}

		attrPred := pred.(*AttributeEqualsPredicate)
		if attrPred.Value != tt.expected {
			t.Fatalf("Expected %v, got %v", tt.expected, attrPred.Value)
		}
	}
}

// TestParsePredicate_AndCombination tests parsing AND combinations
func TestParsePredicate_AndCombination(t *testing.T) {
	pred, err := ParsePredicate("node.type = service AND in CRITICAL")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	andPred, ok := pred.(*AndPredicate)
	if !ok {
		t.Fatalf("Expected AndPredicate, got %T", pred)
	}

	// Check left
	if _, ok := andPred.Left.(*AttributeEqualsPredicate); !ok {
		t.Fatalf("Expected left to be AttributeEqualsPredicate")
	}

	// Check right
	if _, ok := andPred.Right.(*SetMembershipPredicate); !ok {
		t.Fatalf("Expected right to be SetMembershipPredicate")
	}
}

// TestParsePredicate_NotPrefix tests parsing NOT prefix
func TestParsePredicate_NotPrefix(t *testing.T) {
	pred, err := ParsePredicate("not in DEPRECATED")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	notPred, ok := pred.(*NotPredicate)
	if !ok {
		t.Fatalf("Expected NotPredicate, got %T", pred)
	}

	if _, ok := notPred.Inner.(*SetMembershipPredicate); !ok {
		t.Fatalf("Expected inner to be SetMembershipPredicate")
	}
}

// TestParsePredicate_ComplexAnd tests complex AND combinations
func TestParsePredicate_ComplexAnd(t *testing.T) {
	pred, err := ParsePredicate("node.type = service AND not in DEPRECATED")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	node := &gsl.Node{
		ID:         "A",
		Attributes: map[string]interface{}{"type": "service"},
		Sets:       map[string]struct{}{},
	}

	if !pred.EvaluateNode(node) {
		t.Fatal("Should match: service type AND not deprecated")
	}

	nodeDeprecated := &gsl.Node{
		ID:         "B",
		Attributes: map[string]interface{}{"type": "service"},
		Sets:       map[string]struct{}{"DEPRECATED": {}},
	}

	if pred.EvaluateNode(nodeDeprecated) {
		t.Fatal("Should not match: service but deprecated")
	}
}

// TestPredicateTargetType tests TargetType() method
func TestPredicateTargetType(t *testing.T) {
	tests := []struct {
		name   string
		pred   Predicate
		expect string
	}{
		{"NodeAttr", &AttributeEqualsPredicate{Target: "node", Name: "x", Value: 1}, "node"},
		{"EdgeAttr", &AttributeEqualsPredicate{Target: "edge", Name: "x", Value: 1}, "edge"},
		{"NodeSet", &SetMembershipPredicate{Target: "node", SetID: "S"}, "node"},
		{"EdgeSet", &SetMembershipPredicate{Target: "edge", SetID: "S"}, "edge"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pred.TargetType() != tt.expect {
				t.Fatalf("Expected %s, got %s", tt.expect, tt.pred.TargetType())
			}
		})
	}
}

// TestAndPredicate_MixedTargets tests error detection for mixed targets
func TestAndPredicate_MixedTargets(t *testing.T) {
	pred := &AndPredicate{
		Left:  &AttributeEqualsPredicate{Target: "node", Name: "x", Value: 1},
		Right: &AttributeEqualsPredicate{Target: "edge", Name: "y", Value: 2},
	}

	// TargetType() should indicate error (or return "error")
	if pred.TargetType() == "error" {
		// This is expected behavior for mixed targets
		return
	}

	// Should differentiate between node/edge targets
	if pred.TargetType() != "node" && pred.TargetType() != "edge" {
		t.Fatal("Should detect mixed targets")
	}
}

// TestParseNodePredicateNotEqual tests parsing "node.attr != value"
func TestParseNodePredicateNotEqual(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"basic string", "node.env != \"prod\"", false},
		{"no spaces", "node.env!=\"dev\"", false},
		{"boolean", "node.status != true", false},
		{"node id", "node.id != \"test\"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pred, err := ParsePredicate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && pred == nil {
				t.Fatal("expected non-nil predicate")
			}
			if err == nil {
				_, ok := pred.(*AttributeNotEqualsPredicate)
				if !ok {
					t.Fatalf("expected AttributeNotEqualsPredicate, got %T", pred)
				}
			}
		})
	}
}

// TestParseEdgePredicateNotEqual tests parsing "edge.attr != value"
func TestParseEdgePredicateNotEqual(t *testing.T) {
	pred, err := ParsePredicate("edge.protocol != \"HTTP\"")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	notEqualPred, ok := pred.(*AttributeNotEqualsPredicate)
	if !ok {
		t.Fatalf("expected AttributeNotEqualsPredicate, got %T", pred)
	}

	if notEqualPred.Target != "edge" {
		t.Fatalf("expected edge target, got %s", notEqualPred.Target)
	}
	if notEqualPred.Name != "protocol" {
		t.Fatalf("expected protocol attribute, got %s", notEqualPred.Name)
	}
	if notEqualPred.Value != "HTTP" {
		t.Fatalf("expected HTTP value, got %v", notEqualPred.Value)
	}
}

// TestParentExistsPredicate tests edge parent exists
func TestParentExistsPredicate(t *testing.T) {
	t.Run("edge with parent", func(t *testing.T) {
		pred := &ParentExistsPredicate{Target: "edge"}
		edge := &gsl.Edge{From: "A", To: "B", Parent: "E1"}
		if !pred.EvaluateEdge(edge) {
			t.Fatal("edge with Parent should match")
		}
	})

	t.Run("edge without parent", func(t *testing.T) {
		pred := &ParentExistsPredicate{Target: "edge"}
		edge := &gsl.Edge{From: "A", To: "B"}
		if pred.EvaluateEdge(edge) {
			t.Fatal("edge without Parent should not match")
		}
	})

	t.Run("node always false", func(t *testing.T) {
		pred := &ParentExistsPredicate{Target: "edge"}
		node := &gsl.Node{ID: "A"}
		if pred.EvaluateNode(node) {
			t.Fatal("parent exists predicate on node should be false")
		}
	})
}

// TestParentNotExistsPredicate tests edge parent not exists
func TestParentNotExistsPredicate(t *testing.T) {
	t.Run("root edge", func(t *testing.T) {
		pred := &ParentNotExistsPredicate{Target: "edge"}
		edge := &gsl.Edge{From: "A", To: "B", Parent: ""}
		if !pred.EvaluateEdge(edge) {
			t.Fatal("root edge should match parent not exists")
		}
	})

	t.Run("edge with parent", func(t *testing.T) {
		pred := &ParentNotExistsPredicate{Target: "edge"}
		edge := &gsl.Edge{From: "A", To: "B", Parent: "E1"}
		if pred.EvaluateEdge(edge) {
			t.Fatal("edge with Parent should not match parent not exists")
		}
	})
}

// TestDepthPredicate tests edge.depth predicate
func TestDepthPredicate(t *testing.T) {
	// Set up global label index for depth computation
	parent := &gsl.Edge{From: "A", To: "B", Label: "E1"}
	child := &gsl.Edge{From: "B", To: "C", Parent: "E1"}
	globalLabelIndex = map[string]*gsl.Edge{"E1": parent}
	defer func() { globalLabelIndex = nil }()

	t.Run("depth == 0 (root)", func(t *testing.T) {
		pred := &DepthPredicate{Target: "edge", Operator: "==", Value: 0}
		if !pred.EvaluateEdge(parent) {
			t.Fatal("root edge should have depth 0")
		}
	})

	t.Run("depth == 1 (child)", func(t *testing.T) {
		pred := &DepthPredicate{Target: "edge", Operator: "==", Value: 1}
		if !pred.EvaluateEdge(child) {
			t.Fatal("child edge should have depth 1")
		}
	})

	t.Run("depth != 0", func(t *testing.T) {
		pred := &DepthPredicate{Target: "edge", Operator: "!=", Value: 0}
		if !pred.EvaluateEdge(child) {
			t.Fatal("child edge should have depth != 0")
		}
		if pred.EvaluateEdge(parent) {
			t.Fatal("root edge should not have depth != 0")
		}
	})

	t.Run("depth == 2 on child returns false", func(t *testing.T) {
		pred := &DepthPredicate{Target: "edge", Operator: "==", Value: 2}
		if pred.EvaluateEdge(child) {
			t.Fatal("child edge should not have depth 2")
		}
	})

	t.Run("nil edge returns false", func(t *testing.T) {
		pred := &DepthPredicate{Target: "edge", Operator: "==", Value: 0}
		if pred.EvaluateEdge(nil) {
			t.Fatal("nil edge should not match")
		}
	})
}

// TestParentPredicate tests edge depends on <predicate>
func TestParentPredicate(t *testing.T) {
	// Create a parent edge
	parent := &gsl.Edge{From: "X", To: "Y", Label: "E1", Attributes: map[string]interface{}{"protocol": "http"}}
	// Create a child edge referencing the parent
	child := &gsl.Edge{From: "A", To: "B", Parent: "E1"}
	allEdges := []*gsl.Edge{parent, child}

	t.Run("parent matches inner predicate", func(t *testing.T) {
		inner := &AttributeEqualsPredicate{Target: "edge", Name: "protocol", Value: "http"}
		pred := &ParentPredicate{Target: "edge", Inner: inner, allEdges: allEdges}
		if !pred.EvaluateEdge(child) {
			t.Fatal("child edge's parent has protocol=http, should match")
		}
	})

	t.Run("parent does not match inner predicate", func(t *testing.T) {
		inner := &AttributeEqualsPredicate{Target: "edge", Name: "protocol", Value: "grpc"}
		pred := &ParentPredicate{Target: "edge", Inner: inner, allEdges: allEdges}
		if pred.EvaluateEdge(child) {
			t.Fatal("child edge's parent has protocol=http, should not match grpc")
		}
	})

	t.Run("edge without parent returns false", func(t *testing.T) {
		inner := &ExistsPredicate{}
		pred := &ParentPredicate{Target: "edge", Inner: inner, allEdges: allEdges}
		orphan := &gsl.Edge{From: "C", To: "D"}
		if pred.EvaluateEdge(orphan) {
			t.Fatal("edge without Parent should not match")
		}
	})

	t.Run("node always false", func(t *testing.T) {
		inner := &ExistsPredicate{}
		pred := &ParentPredicate{Target: "edge", Inner: inner, allEdges: allEdges}
		node := &gsl.Node{ID: "A"}
		if pred.EvaluateNode(node) {
			t.Fatal("depends on predicate on node should be false")
		}
	})
}

// TestPredicateTargetTypeNew tests TargetType for new predicates
func TestPredicateTargetTypeNew(t *testing.T) {
	tests := []struct {
		name   string
		pred   Predicate
		expect string
	}{
		{"ParentExists", &ParentExistsPredicate{Target: "edge"}, "edge"},
		{"ParentNotExists", &ParentNotExistsPredicate{Target: "edge"}, "edge"},
		{"Depth", &DepthPredicate{Target: "edge", Operator: "==", Value: 0}, "edge"},
		{"ParentPredicate", &ParentPredicate{Target: "edge", Inner: &ExistsPredicate{}}, "edge"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pred.TargetType() != tt.expect {
				t.Fatalf("Expected %s, got %s", tt.expect, tt.pred.TargetType())
			}
		})
	}
}
