package query

import (
	"fmt"
	"strings"

	"github.com/dnnrly/gsl-lang"
)

// Predicate evaluates conditions on nodes and edges
// An expression can target either nodes or edges, not both (mixed error in SPEC)
type Predicate interface {
	// EvaluateNode returns true if node matches predicate
	EvaluateNode(node *gsl.Node) bool

	// EvaluateEdge returns true if edge matches predicate
	EvaluateEdge(edge *gsl.Edge) bool

	// TargetType returns "node", "edge", or "" (error if both)
	TargetType() string
}

// ExistsPredicate matches nodes/edges that exist (always true)
type ExistsPredicate struct{}

func (p *ExistsPredicate) EvaluateNode(node *gsl.Node) bool { return true }
func (p *ExistsPredicate) EvaluateEdge(edge *gsl.Edge) bool { return true }
func (p *ExistsPredicate) TargetType() string               { return "" }

// NodeExistsPredicate matches all nodes (syntactic sugar for exists)
type NodeExistsPredicate struct{}

func (p *NodeExistsPredicate) EvaluateNode(node *gsl.Node) bool { return true }
func (p *NodeExistsPredicate) EvaluateEdge(edge *gsl.Edge) bool { return false }
func (p *NodeExistsPredicate) TargetType() string               { return "node" }

// EdgeExistsPredicate matches all edges
type EdgeExistsPredicate struct{}

func (p *EdgeExistsPredicate) EvaluateNode(node *gsl.Node) bool { return false }
func (p *EdgeExistsPredicate) EvaluateEdge(edge *gsl.Edge) bool { return true }
func (p *EdgeExistsPredicate) TargetType() string               { return "edge" }

// AttributeEqualsPredicate matches nodes/edges with attribute = value
type AttributeEqualsPredicate struct {
	Target string      // "node" or "edge"
	Name   string      // attribute name
	Value  interface{} // expected value
}

func (p *AttributeEqualsPredicate) EvaluateNode(node *gsl.Node) bool {
	if node == nil || node.Attributes == nil {
		return false
	}
	val, exists := node.Attributes[p.Name]
	if !exists {
		return false
	}
	return val == p.Value
}

func (p *AttributeEqualsPredicate) EvaluateEdge(edge *gsl.Edge) bool {
	if edge == nil || edge.Attributes == nil {
		return false
	}
	val, exists := edge.Attributes[p.Name]
	if !exists {
		return false
	}
	return val == p.Value
}

func (p *AttributeEqualsPredicate) TargetType() string { return p.Target }

// SetMembershipPredicate matches nodes/edges in a set
type SetMembershipPredicate struct {
	Target string // "node" or "edge"
	SetID  string // set name
}

func (p *SetMembershipPredicate) EvaluateNode(node *gsl.Node) bool {
	if node == nil || node.Sets == nil {
		return false
	}
	_, exists := node.Sets[p.SetID]
	return exists
}

func (p *SetMembershipPredicate) EvaluateEdge(edge *gsl.Edge) bool {
	if edge == nil || edge.Sets == nil {
		return false
	}
	_, exists := edge.Sets[p.SetID]
	return exists
}

func (p *SetMembershipPredicate) TargetType() string { return p.Target }

// SetNotMembershipPredicate matches nodes/edges NOT in a set
type SetNotMembershipPredicate struct {
	Target string // "node" or "edge"
	SetID  string // set name
}

func (p *SetNotMembershipPredicate) EvaluateNode(node *gsl.Node) bool {
	if node == nil || node.Sets == nil {
		return true // No sets = not in this set
	}
	_, exists := node.Sets[p.SetID]
	return !exists
}

func (p *SetNotMembershipPredicate) EvaluateEdge(edge *gsl.Edge) bool {
	if edge == nil || edge.Sets == nil {
		return true // No sets = not in this set
	}
	_, exists := edge.Sets[p.SetID]
	return !exists
}

func (p *SetNotMembershipPredicate) TargetType() string { return p.Target }

// AndPredicate combines predicates with AND (both must be true)
type AndPredicate struct {
	Left  Predicate
	Right Predicate
}

func (p *AndPredicate) EvaluateNode(node *gsl.Node) bool {
	return p.Left.EvaluateNode(node) && p.Right.EvaluateNode(node)
}

func (p *AndPredicate) EvaluateEdge(edge *gsl.Edge) bool {
	return p.Left.EvaluateEdge(edge) && p.Right.EvaluateEdge(edge)
}

func (p *AndPredicate) TargetType() string {
	left := p.Left.TargetType()
	right := p.Right.TargetType()

	// If either is empty, use the other
	if left == "" {
		return right
	}
	if right == "" {
		return left
	}

	// If both are specified and different, error
	if left != right {
		return "error"
	}

	return left
}

// NotPredicate inverts a predicate
type NotPredicate struct {
	Inner Predicate
}

func (p *NotPredicate) EvaluateNode(node *gsl.Node) bool {
	return !p.Inner.EvaluateNode(node)
}

func (p *NotPredicate) EvaluateEdge(edge *gsl.Edge) bool {
	return !p.Inner.EvaluateEdge(edge)
}

func (p *NotPredicate) TargetType() string {
	return p.Inner.TargetType()
}

// ParsePredicate parses a predicate string
// Formats:
//   exists
//   node.attr = value
//   edge.attr = value
//   in SETNAME
//   not in SETNAME
//   pred1 AND pred2
func ParsePredicate(input string) (Predicate, error) {
	input = strings.TrimSpace(input)

	if input == "" {
		return nil, fmt.Errorf("empty predicate")
	}

	// Try parsing as simple predicate first
	if !strings.Contains(input, " AND ") {
		return parseSimplePredicate(input)
	}

	// Split on AND
	parts := strings.Split(input, " AND ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid AND predicate")
	}

	left, err := ParsePredicate(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, err
	}

	right, err := ParsePredicate(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, err
	}

	return &AndPredicate{Left: left, Right: right}, nil
}

// parseSimplePredicate parses a single predicate (no AND)
func parseSimplePredicate(input string) (Predicate, error) {
	input = strings.TrimSpace(input)

	// Handle "not" prefix
	if strings.HasPrefix(input, "not ") {
		inner, err := parseSimplePredicate(strings.TrimSpace(input[4:]))
		if err != nil {
			return nil, err
		}
		return &NotPredicate{Inner: inner}, nil
	}

	// Handle "exists"
	if input == "exists" {
		return &ExistsPredicate{}, nil
	}

	// Handle "in SETNAME"
	if strings.HasPrefix(input, "in ") {
		setName := strings.TrimSpace(input[3:])
		if setName == "" {
			return nil, fmt.Errorf("set name required for 'in'")
		}
		return &SetMembershipPredicate{Target: "", SetID: setName}, nil
	}

	// Handle "not in SETNAME"
	if strings.HasPrefix(input, "not in ") {
		setName := strings.TrimSpace(input[7:])
		if setName == "" {
			return nil, fmt.Errorf("set name required for 'not in'")
		}
		return &SetNotMembershipPredicate{Target: "", SetID: setName}, nil
	}

	// Handle "node.attr = value"
	if strings.HasPrefix(input, "node.") {
		return parseNodePredicate(input)
	}

	// Handle "edge.attr = value"
	if strings.HasPrefix(input, "edge.") {
		return parseEdgePredicate(input)
	}

	return nil, fmt.Errorf("unknown predicate: %s", input)
}

// parseNodePredicate parses "node.attr = value"
func parseNodePredicate(input string) (Predicate, error) {
	if !strings.HasPrefix(input, "node.") {
		return nil, fmt.Errorf("expected node. prefix")
	}

	rest := input[5:] // skip "node."

	// Find the = sign
	idx := strings.Index(rest, " = ")
	if idx == -1 {
		return nil, fmt.Errorf("expected ' = ' in node predicate")
	}

	attrName := strings.TrimSpace(rest[:idx])
	valueStr := strings.TrimSpace(rest[idx+3:])

	if attrName == "" || valueStr == "" {
		return nil, fmt.Errorf("invalid node predicate: %s", input)
	}

	// Parse value (simple: string, number, boolean)
	value := parseValue(valueStr)

	return &AttributeEqualsPredicate{
		Target: "node",
		Name:   attrName,
		Value:  value,
	}, nil
}

// parseEdgePredicate parses "edge.attr = value"
func parseEdgePredicate(input string) (Predicate, error) {
	if !strings.HasPrefix(input, "edge.") {
		return nil, fmt.Errorf("expected edge. prefix")
	}

	rest := input[5:] // skip "edge."

	// Find the = sign
	idx := strings.Index(rest, " = ")
	if idx == -1 {
		return nil, fmt.Errorf("expected ' = ' in edge predicate")
	}

	attrName := strings.TrimSpace(rest[:idx])
	valueStr := strings.TrimSpace(rest[idx+3:])

	if attrName == "" || valueStr == "" {
		return nil, fmt.Errorf("invalid edge predicate: %s", input)
	}

	// Parse value
	value := parseValue(valueStr)

	return &AttributeEqualsPredicate{
		Target: "edge",
		Name:   attrName,
		Value:  value,
	}, nil
}

// parseValue parses a string value into string, number, or boolean
func parseValue(s string) interface{} {
	s = strings.TrimSpace(s)

	// Handle string literals: "..." or '...'
	if (strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) ||
		(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		return s[1 : len(s)-1]
	}

	// Handle boolean
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	// Try number (simple check)
	if isNumericString(s) {
		// Keep as string for now (type-sensitive equality per spec)
		// Could parse to int/float but spec says string != number
		return s
	}

	// Default: return as string
	return s
}

// isNumericString checks if string looks like a number
func isNumericString(s string) bool {
	if s == "" {
		return false
	}

	// Check for minus sign
	start := 0
	if s[0] == '-' || s[0] == '+' {
		start = 1
	}

	hasDigit := false
	hasDot := false

	for i := start; i < len(s); i++ {
		ch := s[i]
		if ch >= '0' && ch <= '9' {
			hasDigit = true
		} else if ch == '.' && !hasDot {
			hasDot = true
		} else {
			return false
		}
	}

	return hasDigit
}
