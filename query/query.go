package query

import (
	"fmt"

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
}

// NewQueryParser creates a new parser for the input string
func NewQueryParser(input string) *QueryParser {
	return &QueryParser{input: input}
}

// Parse parses the input string and returns a Query
func (p *QueryParser) Parse() (*Query, error) {
	return parseQuery(p.input)
}

// IdentityExpr returns the input unchanged
type IdentityExpr struct{}

func (e *IdentityExpr) Apply(ctx *QueryContext, input Value) (Value, error) {
	return input, nil
}

// expressionParser is a thin wrapper for backward compatibility with tests
type expressionParser struct {
	input string
}

func newExpressionParser(input string) *expressionParser {
	return &expressionParser{input: input}
}

// parse uses the new participle-based parser to parse a single expression
func (p *expressionParser) parse() (Expression, error) {
	// Delegate to the new parser for compatibility
	parser := NewQueryParser(p.input)
	query, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	
	// Extract the first (and usually only) expression
	if len(query.Expressions) == 0 {
		return &IdentityExpr{}, nil
	}
	return query.Expressions[0], nil
}

// parseGraphAlgebra parses only graph algebra expressions for test compatibility
func (p *expressionParser) parseGraphAlgebra() (Expression, error) {
	// Try to parse as graph algebra
	expr, err := p.parse()
	if err != nil {
		return nil, err
	}
	
	// Check if it's actually a graph algebra expression
	alg, ok := expr.(*GraphAlgebraExpr)
	if !ok {
		// Not a graph algebra expression
		return nil, fmt.Errorf("not a graph algebra expression")
	}
	
	// Validate the graph references
	if !isValidGraphRef(alg.LeftRef) {
		return nil, fmt.Errorf("invalid left graph reference: %s", alg.LeftRef)
	}
	if !isValidGraphRef(alg.RightRef) {
		return nil, fmt.Errorf("invalid right graph reference: %s", alg.RightRef)
	}
	
	return alg, nil
}
