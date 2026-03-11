package query

import (
	"fmt"
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
	// Two-stage parser:
	// 1. Split on |
	// 2. Parse each expression individually

	expressions := []Expression{}

	// Split on |
	pipelineParts := p.splitPipeline()

	// Parse each part as an expression
	for _, part := range pipelineParts {
		exprParser := newExpressionParser(part)
		expr, err := exprParser.parse()
		if err != nil {
			return nil, fmt.Errorf("failed to parse expression: %w", err)
		}
		expressions = append(expressions, expr)
	}

	return &Query{Expressions: expressions}, nil
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
