package gsl

import "fmt"

// ============================================================================
// Node Attribute Accessors
// ============================================================================

// GetString returns the string value of a node attribute.
// Returns (value, true) if the attribute exists and is a string.
// Returns ("", false) if the attribute is missing or not a string.
func (n *Node) GetString(key string) (string, bool) {
	if n == nil || n.Attributes == nil {
		return "", false
	}
	val, ok := n.Attributes[key]
	if !ok {
		return "", false
	}
	s, isString := val.(string)
	return s, isString
}

// GetInt returns the int64 value of a node attribute.
// Returns (value, true) if the attribute exists and is a number.
// Returns (0, false) if the attribute is missing or not a number.
func (n *Node) GetInt(key string) (int64, bool) {
	if n == nil || n.Attributes == nil {
		return 0, false
	}
	val, ok := n.Attributes[key]
	if !ok {
		return 0, false
	}
	// Handle both float64 (from JSON) and int64
	switch v := val.(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	default:
		return 0, false
	}
}

// GetBool returns the bool value of a node attribute.
// Returns (value, true) if the attribute exists and is a bool.
// Returns (false, false) if the attribute is missing or not a bool.
func (n *Node) GetBool(key string) (bool, bool) {
	if n == nil || n.Attributes == nil {
		return false, false
	}
	val, ok := n.Attributes[key]
	if !ok {
		return false, false
	}
	b, isBool := val.(bool)
	return b, isBool
}

// GetRef returns the NodeRef value of a node attribute.
// Returns (value, true) if the attribute exists and is a NodeRef.
// Returns (nil, false) if the attribute is missing or not a NodeRef.
func (n *Node) GetRef(key string) (*NodeRef, bool) {
	if n == nil || n.Attributes == nil {
		return nil, false
	}
	val, ok := n.Attributes[key]
	if !ok {
		return nil, false
	}
	ref, isRef := val.(NodeRef)
	if !isRef {
		return nil, false
	}
	return &ref, true
}

// SetAttribute sets an attribute on a node.
// Returns an error if validation fails.
func (n *Node) SetAttribute(key string, val interface{}) error {
	if n == nil {
		return fmt.Errorf("cannot set attribute on nil node")
	}
	if key == "" {
		return fmt.Errorf("attribute key cannot be empty")
	}
	if n.Attributes == nil {
		n.Attributes = make(map[string]interface{})
	}
	n.Attributes[key] = val
	return nil
}

// ============================================================================
// Edge Attribute Accessors
// ============================================================================

// GetString returns the string value of an edge attribute.
// Returns (value, true) if the attribute exists and is a string.
// Returns ("", false) if the attribute is missing or not a string.
func (e *Edge) GetString(key string) (string, bool) {
	if e == nil || e.Attributes == nil {
		return "", false
	}
	val, ok := e.Attributes[key]
	if !ok {
		return "", false
	}
	s, isString := val.(string)
	return s, isString
}

// GetInt returns the int64 value of an edge attribute.
// Returns (value, true) if the attribute exists and is a number.
// Returns (0, false) if the attribute is missing or not a number.
func (e *Edge) GetInt(key string) (int64, bool) {
	if e == nil || e.Attributes == nil {
		return 0, false
	}
	val, ok := e.Attributes[key]
	if !ok {
		return 0, false
	}
	// Handle both float64 (from JSON) and int64
	switch v := val.(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	default:
		return 0, false
	}
}

// GetBool returns the bool value of an edge attribute.
// Returns (value, true) if the attribute exists and is a bool.
// Returns (false, false) if the attribute is missing or not a bool.
func (e *Edge) GetBool(key string) (bool, bool) {
	if e == nil || e.Attributes == nil {
		return false, false
	}
	val, ok := e.Attributes[key]
	if !ok {
		return false, false
	}
	b, isBool := val.(bool)
	return b, isBool
}

// SetAttribute sets an attribute on an edge.
// Returns an error if validation fails.
func (e *Edge) SetAttribute(key string, val interface{}) error {
	if e == nil {
		return fmt.Errorf("cannot set attribute on nil edge")
	}
	if key == "" {
		return fmt.Errorf("attribute key cannot be empty")
	}
	if e.Attributes == nil {
		e.Attributes = make(map[string]interface{})
	}
	e.Attributes[key] = val
	return nil
}

// ============================================================================
// Set Attribute Accessors
// ============================================================================

// GetString returns the string value of a set attribute.
// Returns (value, true) if the attribute exists and is a string.
// Returns ("", false) if the attribute is missing or not a string.
func (s *Set) GetString(key string) (string, bool) {
	if s == nil || s.Attributes == nil {
		return "", false
	}
	val, ok := s.Attributes[key]
	if !ok {
		return "", false
	}
	str, isString := val.(string)
	return str, isString
}

// GetInt returns the int64 value of a set attribute.
// Returns (value, true) if the attribute exists and is a number.
// Returns (0, false) if the attribute is missing or not a number.
func (s *Set) GetInt(key string) (int64, bool) {
	if s == nil || s.Attributes == nil {
		return 0, false
	}
	val, ok := s.Attributes[key]
	if !ok {
		return 0, false
	}
	// Handle both float64 (from JSON) and int64
	switch v := val.(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	default:
		return 0, false
	}
}

// GetBool returns the bool value of a set attribute.
// Returns (value, true) if the attribute exists and is a bool.
// Returns (false, false) if the attribute is missing or not a bool.
func (s *Set) GetBool(key string) (bool, bool) {
	if s == nil || s.Attributes == nil {
		return false, false
	}
	val, ok := s.Attributes[key]
	if !ok {
		return false, false
	}
	b, isBool := val.(bool)
	return b, isBool
}

// SetAttribute sets an attribute on a set.
// Returns an error if validation fails.
func (s *Set) SetAttribute(key string, val interface{}) error {
	if s == nil {
		return fmt.Errorf("cannot set attribute on nil set")
	}
	if key == "" {
		return fmt.Errorf("attribute key cannot be empty")
	}
	if s.Attributes == nil {
		s.Attributes = make(map[string]interface{})
	}
	s.Attributes[key] = val
	return nil
}
