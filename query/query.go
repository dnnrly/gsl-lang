package query

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dnnrly/gsl-lang"
)

// Value represents a value flowing through the query pipeline.
// In v1, this is always a graph. Future versions may support NodeSet, EdgeSet, Scalar.
type Value interface{}

// GraphValue wraps a Graph as a Value
type GraphValue struct {
	Graph *gsl.Graph
}

// QueryContext holds the state for query execution
type QueryContext struct {
	InputGraph  *gsl.Graph
	NamedGraphs map[string]*gsl.Graph
}

// Expression is implemented by all query expressions.
// Each expression transforms an input value into an output value.
type Expression interface {
	Apply(ctx *QueryContext, input Value) (Value, error)
}

// Query represents a pipeline of expressions
type Query struct {
	Expressions []Expression
}

// Execute runs the query pipeline, starting with InputGraph from ctx
func (q *Query) Execute(ctx *QueryContext) (Value, error) {
	// Initialize with input graph
	var value Value = GraphValue{ctx.InputGraph}

	// Apply each expression in sequence
	for _, expr := range q.Expressions {
		var err error
		value, err = expr.Apply(ctx, value)
		if err != nil {
			return nil, err
		}
	}

	return value, nil
}

// QueryParser parses a query string into a Query AST
type QueryParser struct {
	input string
	pos   int
}

// NewQueryParser creates a new parser for the input string
func NewQueryParser(input string) *QueryParser {
	return &QueryParser{input: input, pos: 0}
}

// Parse parses the input string and returns a Query
func (p *QueryParser) Parse() (*Query, error) {
	return parseQuery(p.input)
}

// splitPipeline splits the input on | while respecting parentheses
func (p *QueryParser) splitPipeline() []string {
	var parts []string
	var current string
	parenDepth := 0

	for i := 0; i < len(p.input); i++ {
		ch := p.input[i]
		switch ch {
		case '(':
			parenDepth++
			current += string(ch)
		case ')':
			parenDepth--
			current += string(ch)
		case '|':
			if parenDepth == 0 {
				part := trimSpace(current)
				parts = append(parts, part) // Always append, even if empty
				current = ""
			} else {
				current += string(ch)
			}
		default:
			current += string(ch)
		}
	}

	// Add final part
	part := trimSpace(current)
	parts = append(parts, part) // Always append final part

	return parts
}

// expressionParser parses a single expression
type expressionParser struct {
	input string
}

func newExpressionParser(input string) *expressionParser {
	return &expressionParser{input: input}
}

// parse parses the expression. Dispatches to specific parsers.
func (p *expressionParser) parse() (Expression, error) {
	input := trimSpace(p.input)

	// Empty expression: identity (pass-through)
	if input == "" {
		return &IdentityExpr{}, nil
	}

	// Check for binding: (pipeline) as NAME
	if strings.HasPrefix(input, "(") {
		return p.parseBind()
	}

	// Dispatch to specific expression parsers
	if strings.HasPrefix(input, "from ") || input == "from" {
		return p.parseFrom()
	}

	// Subgraph expression
	if strings.HasPrefix(input, "subgraph ") || input == "subgraph" {
		return p.parseSubgraph()
	}

	// Remove expression
	if strings.HasPrefix(input, "remove ") || input == "remove" {
		return p.parseRemove()
	}

	// Make expression
	if strings.HasPrefix(input, "make ") || input == "make" {
		return p.parseMake()
	}

	// Collapse expression
	if strings.HasPrefix(input, "collapse ") || input == "collapse" {
		return p.parseCollapse()
	}

	// Graph algebra expression (try to parse as algebra)
	expr, err := p.parseGraphAlgebra()
	if err == nil {
		return expr, nil
	}
	// If not algebra, try other patterns or fail below

	// Unknown expression type
	return nil, fmt.Errorf("unknown expression: %s", input)
}

// parseBind parses "(pipeline) as NAME"
func (p *expressionParser) parseBind() (Expression, error) {
	input := trimSpace(p.input)

	// Find matching closing paren
	closeParenIdx := p.findClosingParen(input, 0)
	if closeParenIdx == -1 {
		return nil, fmt.Errorf("unclosed parenthesis in binding")
	}

	// Extract pipeline
	pipelineStr := trimSpace(input[1:closeParenIdx])

	// Rest should be "as NAME"
	rest := trimSpace(input[closeParenIdx+1:])
	if !strings.HasPrefix(rest, "as ") {
		return nil, fmt.Errorf("binding must have 'as NAME' after pipeline")
	}

	// Extract name
	name := trimSpace(rest[3:]) // skip "as "
	if name == "" {
		return nil, fmt.Errorf("binding requires a name")
	}

	// Validate name
	if !isValidGraphName(name) {
		return nil, fmt.Errorf("invalid graph name: %s (must match [A-Z][A-Z0-9_]*)", name)
	}

	// Parse the subpipeline
	subParser := NewQueryParser(pipelineStr)
	subQuery, err := subParser.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse subpipeline: %w", err)
	}

	return &BindExpr{Pipeline: subQuery, Name: name}, nil
}

// findClosingParen finds the index of the closing paren that matches the opening paren at startIdx
func (p *expressionParser) findClosingParen(input string, startIdx int) int {
	if startIdx >= len(input) || input[startIdx] != '(' {
		return -1
	}

	depth := 1
	for i := startIdx + 1; i < len(input); i++ {
		switch input[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}

	return -1 // No matching closing paren
}

// parseFrom parses "from *" or "from NAME"
func (p *expressionParser) parseFrom() (Expression, error) {
	input := trimSpace(p.input)

	if input == "from" {
		return nil, fmt.Errorf("from requires an argument: 'from *' or 'from NAME'")
	}

	if !strings.HasPrefix(input, "from ") {
		return nil, fmt.Errorf("invalid from expression: %s", input)
	}

	// Extract the argument after "from "
	arg := trimSpace(input[5:]) // skip "from "

	if arg == "" {
		return nil, fmt.Errorf("from requires an argument: 'from *' or 'from NAME'")
	}

	// Check for wildcard
	if arg == "*" {
		return &FromExpr{IsWildcard: true, Name: ""}, nil
	}

	// Named graph - must be uppercase [A-Z][A-Z0-9_]*
	if !isValidGraphName(arg) {
		return nil, fmt.Errorf("invalid graph name: %s (must match [A-Z][A-Z0-9_]*)", arg)
	}

	return &FromExpr{IsWildcard: false, Name: arg}, nil
}

// parseSubgraph parses "subgraph <predicate> [traverse <direction> <depth>]"
func (p *expressionParser) parseSubgraph() (Expression, error) {
	input := trimSpace(p.input)

	if input == "subgraph" {
		return nil, fmt.Errorf("subgraph requires a predicate")
	}

	if !strings.HasPrefix(input, "subgraph ") {
		return nil, fmt.Errorf("invalid subgraph expression: %s", input)
	}

	// Extract everything after "subgraph "
	rest := trimSpace(input[9:]) // skip "subgraph "

	if rest == "" {
		return nil, fmt.Errorf("subgraph requires a predicate")
	}

	// Check for traverse clause
	traverseIdx := strings.Index(rest, " traverse ")
	var predicateStr, traversalStr string

	if traverseIdx == -1 {
		// No traversal
		predicateStr = rest
		traversalStr = ""
	} else {
		// Split predicate and traversal
		predicateStr = trimSpace(rest[:traverseIdx])
		traversalStr = trimSpace(rest[traverseIdx+10:]) // skip " traverse "
	}

	if predicateStr == "" {
		return nil, fmt.Errorf("subgraph requires a predicate")
	}

	// Parse the predicate
	pred, err := ParsePredicate(predicateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse predicate: %w", err)
	}

	// Parse traversal if present
	var traversal *TraversalConfig
	if traversalStr != "" {
		trav, err := parseTraversal(traversalStr)
		if err != nil {
			return nil, err
		}
		traversal = trav
	}

	return &SubgraphExpr{Pred: pred, Traversal: traversal}, nil
}

// parseRemove parses remove expressions:
//   remove edge where <predicate>
//   remove node.<attr> where <predicate>
//   remove edge.<attr> where <predicate>
//   remove orphans
func (p *expressionParser) parseRemove() (Expression, error) {
	input := trimSpace(p.input)

	if input == "remove" {
		return nil, fmt.Errorf("remove requires an argument")
	}

	if !strings.HasPrefix(input, "remove ") {
		return nil, fmt.Errorf("invalid remove expression: %s", input)
	}

	// Extract everything after "remove "
	rest := trimSpace(input[7:])

	if rest == "" {
		return nil, fmt.Errorf("remove requires an argument")
	}

	// Handle "remove orphans"
	if rest == "orphans" {
		return &RemoveOrphansExpr{}, nil
	}

	// Handle "remove edge where <predicate>"
	if strings.HasPrefix(rest, "edge where ") {
		predicateStr := trimSpace(rest[11:]) // skip "edge where "
		if predicateStr == "" {
			return nil, fmt.Errorf("remove edge where requires a predicate")
		}
		pred, err := ParsePredicate(predicateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse predicate: %w", err)
		}
		return &RemoveEdgeExpr{Pred: pred}, nil
	}

	// Handle "remove node.<attr> where <predicate>"
	if strings.HasPrefix(rest, "node.") {
		return p.parseRemoveAttribute("node", rest)
	}

	// Handle "remove edge.<attr> where <predicate>"
	if strings.HasPrefix(rest, "edge.") {
		return p.parseRemoveAttribute("edge", rest)
	}

	return nil, fmt.Errorf("unknown remove expression: %s", input)
}

// parseRemoveAttribute parses "node.<attr> where <pred>" or "edge.<attr> where <pred>"
func (p *expressionParser) parseRemoveAttribute(target string, input string) (Expression, error) {
	// Extract attribute name and predicate
	// Format: "node.<attr> where <predicate>"
	prefix := target + "."
	if !strings.HasPrefix(input, prefix) {
		return nil, fmt.Errorf("expected %s. prefix", target)
	}

	rest := input[len(prefix):]

	// Find "where"
	whereIdx := strings.Index(rest, " where ")
	if whereIdx == -1 {
		return nil, fmt.Errorf("remove %s requires ' where <predicate>'", target)
	}

	attrName := trimSpace(rest[:whereIdx])
	predicateStr := trimSpace(rest[whereIdx+7:]) // skip " where "

	if attrName == "" {
		return nil, fmt.Errorf("attribute name required for remove %s", target)
	}

	if predicateStr == "" {
		return nil, fmt.Errorf("predicate required for remove %s", target)
	}

	// Parse the predicate
	pred, err := ParsePredicate(predicateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse predicate: %w", err)
	}

	return &RemoveAttributeExpr{
		Target: target,
		Attr:   attrName,
		Pred:   pred,
	}, nil
}

// parseMake parses make expressions:
//   make node.<attr> = <value> where <predicate>
//   make edge.<attr> = <value> where <predicate>
func (p *expressionParser) parseMake() (Expression, error) {
	input := trimSpace(p.input)

	if input == "make" {
		return nil, fmt.Errorf("make requires an argument")
	}

	if !strings.HasPrefix(input, "make ") {
		return nil, fmt.Errorf("invalid make expression: %s", input)
	}

	// Extract everything after "make "
	rest := trimSpace(input[5:])

	if rest == "" {
		return nil, fmt.Errorf("make requires an argument")
	}

	// Handle "make node.<attr> = <value> where <predicate>"
	if strings.HasPrefix(rest, "node.") {
		return p.parseMakeAttribute("node", rest)
	}

	// Handle "make edge.<attr> = <value> where <predicate>"
	if strings.HasPrefix(rest, "edge.") {
		return p.parseMakeAttribute("edge", rest)
	}

	return nil, fmt.Errorf("unknown make expression: %s", input)
}

// parseMakeAttribute parses "node.<attr> = <value> where <pred>" or "edge.<attr> = <value> where <pred>"
func (p *expressionParser) parseMakeAttribute(target string, input string) (Expression, error) {
	prefix := target + "."
	if !strings.HasPrefix(input, prefix) {
		return nil, fmt.Errorf("expected %s. prefix", target)
	}

	rest := input[len(prefix):]

	// Find first " = "
	eqIdx := strings.Index(rest, " = ")
	if eqIdx == -1 {
		return nil, fmt.Errorf("make %s requires ' = <value>'", target)
	}

	attrName := trimSpace(rest[:eqIdx])
	afterEq := trimSpace(rest[eqIdx+3:]) // skip " = "

	if attrName == "" {
		return nil, fmt.Errorf("attribute name required for make %s", target)
	}

	// Find "where" clause
	whereIdx := strings.Index(afterEq, " where ")
	if whereIdx == -1 {
		return nil, fmt.Errorf("make %s requires ' where <predicate>'", target)
	}

	valueStr := trimSpace(afterEq[:whereIdx])
	predicateStr := trimSpace(afterEq[whereIdx+7:]) // skip " where "

	if valueStr == "" {
		return nil, fmt.Errorf("value required for make %s", target)
	}

	if predicateStr == "" {
		return nil, fmt.Errorf("predicate required for make %s", target)
	}

	// Parse the value
	value := parseValue(valueStr)

	// Parse the predicate
	pred, err := ParsePredicate(predicateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse predicate: %w", err)
	}

	return &MakeExpr{
		Target: target,
		Attr:   attrName,
		Value:  value,
		Pred:   pred,
	}, nil
}

// parseCollapse parses collapse expressions:
//   collapse into <id> where <predicate>
func (p *expressionParser) parseCollapse() (Expression, error) {
	input := trimSpace(p.input)

	if input == "collapse" {
		return nil, fmt.Errorf("collapse requires an argument")
	}

	if !strings.HasPrefix(input, "collapse ") {
		return nil, fmt.Errorf("invalid collapse expression: %s", input)
	}

	// Extract everything after "collapse "
	rest := trimSpace(input[9:])

	if rest == "" {
		return nil, fmt.Errorf("collapse requires an argument")
	}

	// Handle "collapse into <id> where <predicate>"
	if !strings.HasPrefix(rest, "into ") {
		return nil, fmt.Errorf("collapse requires 'into <id> where <predicate>'")
	}

	rest = trimSpace(rest[5:]) // skip "into "

	// Find "where"
	whereIdx := strings.Index(rest, " where ")
	if whereIdx == -1 {
		return nil, fmt.Errorf("collapse requires ' where <predicate>'")
	}

	nodeID := trimSpace(rest[:whereIdx])
	predicateStr := trimSpace(rest[whereIdx+7:]) // skip " where "

	if nodeID == "" {
		return nil, fmt.Errorf("collapse requires a node ID")
	}

	if predicateStr == "" {
		return nil, fmt.Errorf("collapse requires a predicate")
	}

	// Validate node ID (any non-empty identifier is acceptable)
	// Per spec: "<id> MUST be a valid node identifier"
	// We accept any string as a valid node identifier

	// Parse the predicate
	pred, err := ParsePredicate(predicateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse predicate: %w", err)
	}

	return &CollapseExpr{
		NodeID: nodeID,
		Pred:   pred,
	}, nil
}

// parseTraversal parses "<direction> <depth>"
// direction: in, out, both
// depth: number (1, 2, 3, ...) or "all"
func parseTraversal(input string) (*TraversalConfig, error) {
	parts := strings.Fields(input)

	if len(parts) != 2 {
		return nil, fmt.Errorf("traverse requires direction and depth: traverse <direction> <depth>")
	}

	direction := parts[0]
	depthStr := parts[1]

	// Validate direction
	if direction != "in" && direction != "out" && direction != "both" {
		return nil, fmt.Errorf("invalid traverse direction: %s (must be in, out, or both)", direction)
	}

	// Parse depth
	var depth int
	if depthStr == "all" {
		depth = 999999 // Large number for unlimited traversal
	} else {
		d, err := strconv.Atoi(depthStr)
		if err != nil {
			return nil, fmt.Errorf("invalid traverse depth: %s (must be number or 'all')", depthStr)
		}
		if d <= 0 {
			return nil, fmt.Errorf("traverse depth must be positive")
		}
		depth = d
	}

	return &TraversalConfig{Direction: direction, Depth: depth}, nil
}

// isValidGraphName checks if a name matches [A-Z][A-Z0-9_]*
func isValidGraphName(name string) bool {
	if len(name) == 0 {
		return false
	}

	// First character must be uppercase
	if name[0] < 'A' || name[0] > 'Z' {
		return false
	}

	// Remaining characters must be uppercase, digits, or underscore
	for i := 1; i < len(name); i++ {
		ch := name[i]
		if !((ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
			return false
		}
	}

	return true
}

// IdentityExpr returns the input unchanged
type IdentityExpr struct{}

func (e *IdentityExpr) Apply(ctx *QueryContext, input Value) (Value, error) {
	return input, nil
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

// parseGraphAlgebra parses graph algebra expressions: GRAPHREF OP GRAPHREF
// where OP is one of: +, &, -, ^
// GRAPHREF is either "*" (input graph) or a named graph name
func (p *expressionParser) parseGraphAlgebra() (Expression, error) {
	input := trimSpace(p.input)

	// Try to find operators: +, &, -, ^
	// We need to find the operator at the top level (respecting no parens in basic algebra)
	var opIdx int = -1
	var op string

	for _, candidate := range []string{"+", "&", "-", "^"} {
		idx := strings.Index(input, " "+candidate+" ")
		if idx != -1 {
			opIdx = idx
			op = candidate
			break
		}
	}

	// No operator found
	if opIdx == -1 {
		return nil, fmt.Errorf("not a graph algebra expression")
	}

	// Split on operator
	leftRef := trimSpace(input[:opIdx])
	rightRef := trimSpace(input[opIdx+3:]) // skip " OP "

	// Validate graph references
	if !isValidGraphRef(leftRef) {
		return nil, fmt.Errorf("invalid left graph reference: %s", leftRef)
	}
	if !isValidGraphRef(rightRef) {
		return nil, fmt.Errorf("invalid right graph reference: %s", rightRef)
	}

	return &GraphAlgebraExpr{
		LeftRef:  leftRef,
		RightRef: rightRef,
		Operator: op,
	}, nil
}

// isValidGraphRef checks if a reference is valid (either "*" or a named graph name)
func isValidGraphRef(ref string) bool {
	if ref == "*" {
		return true
	}
	// Must match named graph naming: [A-Z][A-Z0-9_]*
	return isValidGraphName(ref)
}
