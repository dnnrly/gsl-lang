package gsl

import "testing"

// ============================================================================
// Node Accessor Tests
// ============================================================================

func TestNodeGetString(t *testing.T) {
	n := &Node{
		Attributes: map[string]interface{}{
			"text":   "hello",
			"number": 42,
			"flag":   true,
		},
	}

	// Successful retrieval
	val, ok := n.GetString("text")
	if !ok || val != "hello" {
		t.Errorf("expected (hello, true), got (%s, %v)", val, ok)
	}

	// Type mismatch
	val, ok = n.GetString("number")
	if ok {
		t.Errorf("expected (empty, false) for number, got (%s, %v)", val, ok)
	}

	// Missing key
	val, ok = n.GetString("missing")
	if ok {
		t.Errorf("expected (empty, false) for missing key, got (%s, %v)", val, ok)
	}

	// Nil node
	var nilNode *Node
	val, ok = nilNode.GetString("text")
	if ok {
		t.Errorf("expected (empty, false) for nil node, got (%s, %v)", val, ok)
	}
}

func TestNodeGetInt(t *testing.T) {
	n := &Node{
		Attributes: map[string]interface{}{
			"count":  42.0,
			"text":   "hello",
			"intVal": int64(100),
		},
	}

	// Successful retrieval from float64
	val, ok := n.GetInt("count")
	if !ok || val != 42 {
		t.Errorf("expected (42, true), got (%d, %v)", val, ok)
	}

	// Type mismatch
	val, ok = n.GetInt("text")
	if ok {
		t.Errorf("expected (0, false) for string, got (%d, %v)", val, ok)
	}

	// Missing key
	val, ok = n.GetInt("missing")
	if ok {
		t.Errorf("expected (0, false) for missing key, got (%d, %v)", val, ok)
	}
}

func TestNodeGetBool(t *testing.T) {
	n := &Node{
		Attributes: map[string]interface{}{
			"enabled":  true,
			"disabled": false,
			"number":   42,
		},
	}

	// Successful retrieval
	val, ok := n.GetBool("enabled")
	if !ok || val != true {
		t.Errorf("expected (true, true), got (%v, %v)", val, ok)
	}

	// False value still succeeds
	val, ok = n.GetBool("disabled")
	if !ok || val != false {
		t.Errorf("expected (false, true), got (%v, %v)", val, ok)
	}

	// Type mismatch
	val, ok = n.GetBool("number")
	if ok {
		t.Errorf("expected (false, false) for number, got (%v, %v)", val, ok)
	}
}

func TestNodeGetRef(t *testing.T) {
	parent := NodeRef("Parent")
	n := &Node{
		Attributes: map[string]interface{}{
			"parent":  parent,
			"sibling": "text",
		},
	}

	// Successful retrieval
	val, ok := n.GetRef("parent")
	if !ok || *val != parent {
		t.Errorf("expected (Parent, true), got (%v, %v)", val, ok)
	}

	// Type mismatch
	val, ok = n.GetRef("sibling")
	if ok {
		t.Errorf("expected (nil, false) for string, got (%v, %v)", val, ok)
	}

	// Missing key
	val, ok = n.GetRef("missing")
	if ok {
		t.Errorf("expected (nil, false) for missing key, got (%v, %v)", val, ok)
	}
}

func TestNodeSetAttribute(t *testing.T) {
	n := &Node{
		ID:         "A",
		Attributes: make(map[string]interface{}),
	}

	// Set string
	err := n.SetAttribute("text", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Attributes["text"] != "hello" {
		t.Error("failed to set string attribute")
	}

	// Set number
	err = n.SetAttribute("count", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Set bool
	err = n.SetAttribute("enabled", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty key should fail
	err = n.SetAttribute("", "value")
	if err == nil {
		t.Error("expected error for empty key")
	}

	// Nil node should fail
	var nilNode *Node
	err = nilNode.SetAttribute("key", "value")
	if err == nil {
		t.Error("expected error for nil node")
	}
}

// ============================================================================
// Edge Accessor Tests
// ============================================================================

func TestEdgeGetString(t *testing.T) {
	e := &Edge{
		From: "A",
		To:   "B",
		Attributes: map[string]interface{}{
			"label": "connection",
			"weight": 3.14,
		},
	}

	val, ok := e.GetString("label")
	if !ok || val != "connection" {
		t.Errorf("expected (connection, true), got (%s, %v)", val, ok)
	}

	val, ok = e.GetString("weight")
	if ok {
		t.Errorf("expected (empty, false) for number, got (%s, %v)", val, ok)
	}
}

func TestEdgeGetInt(t *testing.T) {
	e := &Edge{
		From: "A",
		To:   "B",
		Attributes: map[string]interface{}{
			"weight": 5.0,
		},
	}

	val, ok := e.GetInt("weight")
	if !ok || val != 5 {
		t.Errorf("expected (5, true), got (%d, %v)", val, ok)
	}
}

func TestEdgeGetBool(t *testing.T) {
	e := &Edge{
		From: "A",
		To:   "B",
		Attributes: map[string]interface{}{
			"active": true,
		},
	}

	val, ok := e.GetBool("active")
	if !ok || val != true {
		t.Errorf("expected (true, true), got (%v, %v)", val, ok)
	}
}

func TestEdgeSetAttribute(t *testing.T) {
	e := &Edge{
		From:       "A",
		To:         "B",
		Attributes: make(map[string]interface{}),
	}

	err := e.SetAttribute("label", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = e.SetAttribute("", "value")
	if err == nil {
		t.Error("expected error for empty key")
	}

	var nilEdge *Edge
	err = nilEdge.SetAttribute("key", "value")
	if err == nil {
		t.Error("expected error for nil edge")
	}
}

// ============================================================================
// Set Accessor Tests
// ============================================================================

func TestSetGetString(t *testing.T) {
	s := &Set{
		ID: "critical",
		Attributes: map[string]interface{}{
			"description": "critical services",
			"priority":    1,
		},
	}

	val, ok := s.GetString("description")
	if !ok || val != "critical services" {
		t.Errorf("expected (critical services, true), got (%s, %v)", val, ok)
	}

	val, ok = s.GetString("priority")
	if ok {
		t.Errorf("expected (empty, false) for number, got (%s, %v)", val, ok)
	}
}

func TestSetGetInt(t *testing.T) {
	s := &Set{
		ID: "services",
		Attributes: map[string]interface{}{
			"priority": 1.0,
		},
	}

	val, ok := s.GetInt("priority")
	if !ok || val != 1 {
		t.Errorf("expected (1, true), got (%d, %v)", val, ok)
	}
}

func TestSetGetBool(t *testing.T) {
	s := &Set{
		ID: "prod",
		Attributes: map[string]interface{}{
			"critical": true,
		},
	}

	val, ok := s.GetBool("critical")
	if !ok || val != true {
		t.Errorf("expected (true, true), got (%v, %v)", val, ok)
	}
}

func TestSetSetAttribute(t *testing.T) {
	s := &Set{
		ID:         "test",
		Attributes: make(map[string]interface{}),
	}

	err := s.SetAttribute("label", "test set")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.SetAttribute("", "value")
	if err == nil {
		t.Error("expected error for empty key")
	}

	var nilSet *Set
	err = nilSet.SetAttribute("key", "value")
	if err == nil {
		t.Error("expected error for nil set")
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestTypedAccessorIntegration(t *testing.T) {
	// Build a graph using typed accessors
	g := NewGraph()
	_, err := g.AddNode("API", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Set attributes using typed setters
	apiNode := g.GetNode("API")
	if apiNode == nil {
		t.Fatal("node not found")
	}

	err = apiNode.SetAttribute("timeout", int64(5000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = apiNode.SetAttribute("critical", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Retrieve using typed getters
	timeout, ok := apiNode.GetInt("timeout")
	if !ok || timeout != 5000 {
		t.Errorf("expected timeout 5000, got %d", timeout)
	}

	critical, ok := apiNode.GetBool("critical")
	if !ok || !critical {
		t.Errorf("expected critical true, got %v", critical)
	}

	// Missing attribute
	missingVal, ok := apiNode.GetInt("missing")
	if ok {
		t.Errorf("expected missing attribute to return false, got %v", ok)
	}
	if missingVal != 0 {
		t.Errorf("expected zero value for missing attribute, got %d", missingVal)
	}
}
