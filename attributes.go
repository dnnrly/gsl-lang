package gsl

import "fmt"

// ============================================================================
// AttributeMap Type and Methods
// ============================================================================

// AttributeMap is a map of attribute names to values.
type AttributeMap map[string]interface{}

// GetString returns the string value of an attribute.
// Returns (value, true) if the attribute exists and is a string.
// Returns ("", false) if the attribute is missing or not a string.
func (a AttributeMap) GetString(key string) (string, bool) {
	if a == nil {
		return "", false
	}
	val, ok := a[key]
	if !ok {
		return "", false
	}
	s, isString := val.(string)
	return s, isString
}

// GetBool returns the bool value of an attribute.
// Returns (value, true) if the attribute exists and is a bool.
// Returns (false, false) if the attribute is missing or not a bool.
func (a AttributeMap) GetBool(key string) (bool, bool) {
	if a == nil {
		return false, false
	}
	val, ok := a[key]
	if !ok {
		return false, false
	}
	b, isBool := val.(bool)
	return b, isBool
}

// GetInt returns the int64 value of an attribute.
// Returns (value, true) if the attribute exists and is a number.
// Returns (0, false) if the attribute is missing or not a number.
func (a AttributeMap) GetInt(key string) (int64, bool) {
	if a == nil {
		return 0, false
	}
	val, ok := a[key]
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

// GetRef returns the NodeRef value of an attribute.
// Returns (value, true) if the attribute exists and is a NodeRef.
// Returns (nil, false) if the attribute is missing or not a NodeRef.
func (a AttributeMap) GetRef(key string) (*NodeRef, bool) {
	if a == nil {
		return nil, false
	}
	val, ok := a[key]
	if !ok {
		return nil, false
	}
	ref, isRef := val.(NodeRef)
	if !isRef {
		return nil, false
	}
	return &ref, true
}

// SetAttribute sets an attribute value.
// Returns an error if the key is empty.
func (a AttributeMap) SetAttribute(key string, val interface{}) error {
	if key == "" {
		return fmt.Errorf("attribute key cannot be empty")
	}
	a[key] = val
	return nil
}

// ============================================================================
// Node Wrapper Methods
// ============================================================================

// GetString returns the string value of a node attribute.
func (n *Node) GetString(key string) (string, bool) {
	if n == nil {
		return "", false
	}
	return n.Attributes.GetString(key)
}

// GetInt returns the int64 value of a node attribute.
func (n *Node) GetInt(key string) (int64, bool) {
	if n == nil {
		return 0, false
	}
	return n.Attributes.GetInt(key)
}

// GetBool returns the bool value of a node attribute.
func (n *Node) GetBool(key string) (bool, bool) {
	if n == nil {
		return false, false
	}
	return n.Attributes.GetBool(key)
}

// GetRef returns the NodeRef value of a node attribute.
func (n *Node) GetRef(key string) (*NodeRef, bool) {
	if n == nil {
		return nil, false
	}
	return n.Attributes.GetRef(key)
}

// SetAttribute sets an attribute on a node.
func (n *Node) SetAttribute(key string, val interface{}) error {
	if n == nil {
		return fmt.Errorf("cannot set attribute on nil node")
	}
	if n.Attributes == nil {
		n.Attributes = make(AttributeMap)
	}
	return n.Attributes.SetAttribute(key, val)
}

// ============================================================================
// Edge Wrapper Methods
// ============================================================================

// GetString returns the string value of an edge attribute.
func (e *Edge) GetString(key string) (string, bool) {
	if e == nil {
		return "", false
	}
	return e.Attributes.GetString(key)
}

// GetInt returns the int64 value of an edge attribute.
func (e *Edge) GetInt(key string) (int64, bool) {
	if e == nil {
		return 0, false
	}
	return e.Attributes.GetInt(key)
}

// GetBool returns the bool value of an edge attribute.
func (e *Edge) GetBool(key string) (bool, bool) {
	if e == nil {
		return false, false
	}
	return e.Attributes.GetBool(key)
}

// SetAttribute sets an attribute on an edge.
func (e *Edge) SetAttribute(key string, val interface{}) error {
	if e == nil {
		return fmt.Errorf("cannot set attribute on nil edge")
	}
	if e.Attributes == nil {
		e.Attributes = make(AttributeMap)
	}
	return e.Attributes.SetAttribute(key, val)
}

// ============================================================================
// Set Wrapper Methods
// ============================================================================

// GetString returns the string value of a set attribute.
func (s *Set) GetString(key string) (string, bool) {
	if s == nil {
		return "", false
	}
	return s.Attributes.GetString(key)
}

// GetInt returns the int64 value of a set attribute.
func (s *Set) GetInt(key string) (int64, bool) {
	if s == nil {
		return 0, false
	}
	return s.Attributes.GetInt(key)
}

// GetBool returns the bool value of a set attribute.
func (s *Set) GetBool(key string) (bool, bool) {
	if s == nil {
		return false, false
	}
	return s.Attributes.GetBool(key)
}

// SetAttribute sets an attribute on a set.
func (s *Set) SetAttribute(key string, val interface{}) error {
	if s == nil {
		return fmt.Errorf("cannot set attribute on nil set")
	}
	if s.Attributes == nil {
		s.Attributes = make(AttributeMap)
	}
	return s.Attributes.SetAttribute(key, val)
}
