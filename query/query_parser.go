package query

import (
	"fmt"
	"strconv"
	"strings"

	gsl "github.com/dnnrly/gsl-lang"
)

// QueryParser parses GSL Query Language
type QueryParser struct {
	tokens []gsl.Token
	pos    int
	errors []error
}

// ParseQuery parses a query string and returns a Query AST
func ParseQuery(input string) (*Query, []error) {
	l, err := NewQueryLexer(strings.NewReader(input))
	if err != nil {
		return nil, []error{err}
	}
	tokens := l.Tokenize()
	p := NewQueryParser(tokens)
	q := p.parseQuery()
	return q, p.errors
}

// NewQueryParser creates a new query parser
func NewQueryParser(tokens []gsl.Token) *QueryParser {
	return &QueryParser{tokens: tokens}
}

func (p *QueryParser) peek() gsl.Token {
	if p.pos >= len(p.tokens) {
		return gsl.Token{Type: gsl.TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

func (p *QueryParser) peekAhead(n int) gsl.Token {
	if p.pos+n >= len(p.tokens) {
		return gsl.Token{Type: gsl.TOKEN_EOF}
	}
	return p.tokens[p.pos+n]
}

func (p *QueryParser) advance() gsl.Token {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func (p *QueryParser) expect(t gsl.TokenType) gsl.Token {
	tok := p.peek()
	if tok.Type != t {
		p.addError("expected %s, got %s (%q) at %d:%d", t, tok.Type, tok.Literal, tok.Line, tok.Column)
		return tok
	}
	return p.advance()
}

func (p *QueryParser) addError(format string, args ...interface{}) {
	p.errors = append(p.errors, fmt.Errorf(format, args...))
}

func (p *QueryParser) parseQuery() *Query {
	// Skip leading whitespace/comments
	for p.peek().Type == gsl.TOKEN_EOF {
		if len(p.errors) == 0 {
			p.addError("empty query")
		}
		return &Query{Root: &Pipeline{}}
	}

	step := p.parseTopLevel()
	return &Query{Root: step}
}

func (p *QueryParser) parseTopLevel() Step {
	// Top level can be:
	// - A parenthesized group: (...) or (...) union (...)
	// - A pipeline: start ... | flow ... | where ...
	// - A combinator: pipeline union pipeline

	// Check for parenthesized group
	if p.peek().Type == gsl.TOKEN_LPAREN {
		step := p.parseCombinator()
		if step == nil {
			return &Pipeline{}
		}
		
		// Check for trailing tokens after parenthesized expression
		tok := p.peek()
		if tok.Type != gsl.TOKEN_EOF {
			p.addError("unexpected token %q at %d:%d, expected end of query", tok.Literal, tok.Line, tok.Column)
		}
		
		return step
	}

	// Otherwise parse as pipeline, but check for combinators after
	left := p.parsePipeline()
	if left == nil {
		return &Pipeline{}
	}

	// Check for combinator operators at top level
	tok := p.peek()
	if tok.Type == gsl.TOKEN_UNION || tok.Type == gsl.TOKEN_INTERSECT {
		result := p.parseCombinatorWithLeft(left)
		// Check for trailing tokens after combinator expression
		tok := p.peek()
		if tok.Type != gsl.TOKEN_EOF {
			p.addError("unexpected token %q at %d:%d, expected end of query", tok.Literal, tok.Line, tok.Column)
		}
		return result
	}

	// Check if minus is a combinator or was missed as a step
	if tok.Type == gsl.TOKEN_MINUS {
		// It's only a combinator if it's at the top level after a complete pipeline
		// If the last step in left could accept minus, we shouldn't treat it as combinator
		// For now, always treat top-level minus as combinator
		result := p.parseCombinatorWithLeft(left)
		// Check for trailing tokens after combinator expression
		tok := p.peek()
		if tok.Type != gsl.TOKEN_EOF {
			p.addError("unexpected token %q at %d:%d, expected end of query", tok.Literal, tok.Line, tok.Column)
		}
		return result
	}

	// Check for trailing tokens - any token other than EOF is an error
	if tok.Type != gsl.TOKEN_EOF {
		p.addError("unexpected token %q at %d:%d, expected end of query or combinator operator", tok.Literal, tok.Line, tok.Column)
	}

	return left
}

func (p *QueryParser) parseCombinator() Step {
	// Either we're at '(' or we already have a left side
	var left *Pipeline

	if p.peek().Type == gsl.TOKEN_LPAREN {
		p.advance() // consume '('
		left = p.parsePipeline()
		p.expect(gsl.TOKEN_RPAREN)
	} else {
		left = p.parsePipeline()
	}

	if left == nil {
		return nil
	}

	// Check if there's a combinator operator following
	result := p.parseCombinatorWithLeft(left)
	if result != nil {
		return result
	}
	
	// No combinator operator found, just return the pipeline as-is
	// This handles cases like "(start A)" which is just a parenthesized pipeline
	return left
}

func (p *QueryParser) parseCombinatorWithLeft(left *Pipeline) *CombinatorExpr {
	tok := p.peek()
	switch tok.Type {
	case gsl.TOKEN_UNION, gsl.TOKEN_INTERSECT:
		op := tok.Literal
		line := tok.Line
		col := tok.Column
		p.advance() // consume operator

		// Parse right side (could be another combinator)
		var right *Pipeline
		if p.peek().Type == gsl.TOKEN_LPAREN {
			p.advance() // consume '('
			right = p.parsePipeline()
			p.expect(gsl.TOKEN_RPAREN)
		} else {
			right = p.parsePipeline()
		}

		if right == nil {
			return nil
		}

		expr := &CombinatorExpr{
			Type:     op,
			Left:     left,
			Right:    right,
			IsParens: false,
			Line:     line,
			Column:   col,
		}

		// Check for chained combinators (left-associative)
		if p.peek().Type == gsl.TOKEN_UNION || p.peek().Type == gsl.TOKEN_INTERSECT || p.peek().Type == gsl.TOKEN_MINUS {
			// Need to check if this is a minus operator or something else
			if p.peek().Type == gsl.TOKEN_MINUS {
				nextTok := p.peekAhead(1)
				// If minus is followed by something that starts a pipeline, it's a combinator
				if nextTok.Type != gsl.TOKEN_LPAREN && nextTok.Type != gsl.TOKEN_START && nextTok.Type != gsl.TOKEN_IDENT {
					return expr
				}
			}
			// Convert to right pipeline for next iteration
			rightPipeline := &Pipeline{Steps: []Step{expr}}
			return p.parseCombinatorWithLeft(rightPipeline)
		}

		return expr
	case gsl.TOKEN_MINUS:
		// Minus as a combinator (not a step)
		op := tok.Literal
		line := tok.Line
		col := tok.Column
		p.advance() // consume operator

		// Parse right side
		var right *Pipeline
		if p.peek().Type == gsl.TOKEN_LPAREN {
			p.advance() // consume '('
			right = p.parsePipeline()
			p.expect(gsl.TOKEN_RPAREN)
		} else {
			right = p.parsePipeline()
		}

		if right == nil {
			return nil
		}

		expr := &CombinatorExpr{
			Type:     op,
			Left:     left,
			Right:    right,
			IsParens: false,
			Line:     line,
			Column:   col,
		}

		// Check for chained combinators
		if p.peek().Type == gsl.TOKEN_UNION || p.peek().Type == gsl.TOKEN_INTERSECT || p.peek().Type == gsl.TOKEN_MINUS {
			rightPipeline := &Pipeline{Steps: []Step{expr}}
			return p.parseCombinatorWithLeft(rightPipeline)
		}

		return expr
	default:
		return nil
	}
}

func (p *QueryParser) parsePipeline() *Pipeline {
	var steps []Step

	// Parse first step
	step := p.parseStep()
	if step == nil {
		return nil
	}
	steps = append(steps, step)

	// Parse remaining steps separated by pipes
	for p.peek().Type == gsl.TOKEN_PIPE {
		p.advance() // consume '|'

		// Check if next token is a valid step starter or if it's at end
		nextTok := p.peek()
		if nextTok.Type == gsl.TOKEN_EOF || nextTok.Type == gsl.TOKEN_RPAREN ||
			nextTok.Type == gsl.TOKEN_UNION || nextTok.Type == gsl.TOKEN_INTERSECT {
			p.addError("expected step after '|' at %d:%d", nextTok.Line, nextTok.Column)
			break
		}

		// Check if it's a minus that won't be parsed as a step
		if nextTok.Type == gsl.TOKEN_MINUS {
			minusNextTok := p.peekAhead(1)
			if minusNextTok.Type != gsl.TOKEN_LPAREN && minusNextTok.Type != gsl.TOKEN_START && minusNextTok.Type != gsl.TOKEN_IDENT {
				p.addError("expected step after '|' at %d:%d", nextTok.Line, nextTok.Column)
				break
			}
			// Otherwise, minus will be parsed as a step, so continue
		}

		step = p.parseStep()
		if step == nil {
			break
		}
		steps = append(steps, step)
	}

	return &Pipeline{Steps: steps}
}

func (p *QueryParser) parseStep() Step {
	tok := p.peek()

	switch tok.Type {
	case gsl.TOKEN_START:
		return p.parseStartStep()
	case gsl.TOKEN_FLOW:
		return p.parseFlowStep()
	case gsl.TOKEN_WHERE:
		return p.parseWhereStep()
	case gsl.TOKEN_MINUS:
		// Check if this is a minus step (minus followed by paren or start keyword)
		// vs a minus operator (at end of pipeline)
		nextTok := p.peekAhead(1)
		if nextTok.Type == gsl.TOKEN_LPAREN || nextTok.Type == gsl.TOKEN_START || nextTok.Type == gsl.TOKEN_IDENT {
			return p.parseMinusStep()
		}
		// Otherwise it's a combinator, stop here
		return nil
	default:
		if tok.Type == gsl.TOKEN_EOF || tok.Type == gsl.TOKEN_RPAREN || tok.Type == gsl.TOKEN_PIPE ||
			tok.Type == gsl.TOKEN_UNION || tok.Type == gsl.TOKEN_INTERSECT || tok.Type == gsl.TOKEN_MINUS {
			return nil
		}
		p.addError("unexpected token %s (%q) at %d:%d", tok.Type, tok.Literal, tok.Line, tok.Column)
		p.advance()
		return nil
	}
}

func (p *QueryParser) parseStartStep() *StartStep {
	tok := p.expect(gsl.TOKEN_START)
	if tok.Type != gsl.TOKEN_START {
		return nil
	}

	startLine := tok.Line
	startCol := tok.Column

	var nodeIDs []string

	// Parse comma-separated node IDs (strings or idents)
	idTok := p.peek()
	switch idTok.Type {
	case gsl.TOKEN_STRING:
		p.advance()
		nodeIDs = append(nodeIDs, idTok.Literal)
	case gsl.TOKEN_IDENT:
		p.advance()
		nodeIDs = append(nodeIDs, idTok.Literal)
	default:
		p.addError("expected node ID in start step at %d:%d", idTok.Line, idTok.Column)
		return nil
	}

	// Handle comma-separated list
	for p.peek().Type == gsl.TOKEN_COMMA {
		p.advance() // consume ','
		idTok = p.peek()
		switch idTok.Type {
		case gsl.TOKEN_STRING:
			p.advance()
			nodeIDs = append(nodeIDs, idTok.Literal)
		case gsl.TOKEN_IDENT:
			p.advance()
			nodeIDs = append(nodeIDs, idTok.Literal)
		default:
			p.addError("expected node ID after comma at %d:%d", idTok.Line, idTok.Column)
			break
		}
	}

	return &StartStep{
		NodeIDs: nodeIDs,
		Line:    startLine,
		Column:  startCol,
	}
}

func (p *QueryParser) parseFlowStep() *FlowStep {
	tok := p.expect(gsl.TOKEN_FLOW)
	if tok.Type != gsl.TOKEN_FLOW {
		return nil
	}

	flowLine := tok.Line
	flowCol := tok.Column

	// Parse direction: in, out, both
	dirTok := p.peek()
	var direction string
	switch dirTok.Type {
	case gsl.TOKEN_IN, gsl.TOKEN_OUT, gsl.TOKEN_BOTH:
		direction = dirTok.Literal
		p.advance()
	default:
		p.addError("expected direction (in/out/both) in flow step at %d:%d", dirTok.Line, dirTok.Column)
		return nil
	}

	// Check for recursion modifiers or edge filter
	var recursive bool
	var edgeFilter *FilterSpec

	for {
		tok := p.peek()
		switch tok.Type {
		case gsl.TOKEN_RECURSIVE, gsl.TOKEN_STAR:
			if recursive {
				p.addError("recursive already specified at %d:%d", tok.Line, tok.Column)
			}
			recursive = true
			p.advance()

		case gsl.TOKEN_WHERE:
			// Edge filter must be preceded by "where"
			p.advance() // consume 'where'
			edgeFilter = p.parseFilterSpec(true) // true = edge filter
			if edgeFilter == nil {
				return nil
			}
			// After edge filter, can still have recursive
			if p.peek().Type == gsl.TOKEN_RECURSIVE || p.peek().Type == gsl.TOKEN_STAR {
				recursive = true
				p.advance()
			}

		default:
			// End of flow step
			goto done
		}
	}

done:
	return &FlowStep{
		Direction:  direction,
		Recursive:  recursive,
		EdgeFilter: edgeFilter,
		Line:       flowLine,
		Column:     flowCol,
	}
}

func (p *QueryParser) parseWhereStep() *FilterStep {
	tok := p.expect(gsl.TOKEN_WHERE)
	if tok.Type != gsl.TOKEN_WHERE {
		return nil
	}

	whereLine := tok.Line
	whereCol := tok.Column

	filter := p.parseFilterSpec(false) // false = node filter
	if filter == nil {
		return nil
	}

	return &FilterStep{
		Filter: filter,
		Line:   whereLine,
		Column: whereCol,
	}
}

func (p *QueryParser) parseMinusStep() *MinusStep {
	tok := p.expect(gsl.TOKEN_MINUS)
	if tok.Type != gsl.TOKEN_MINUS {
		return nil
	}

	minusLine := tok.Line
	minusCol := tok.Column

	// Parse sub-pipeline (can be parenthesized)
	var subPipeline *Pipeline
	if p.peek().Type == gsl.TOKEN_LPAREN {
		p.advance() // consume '('
		subPipeline = p.parsePipeline()
		p.expect(gsl.TOKEN_RPAREN)
	} else {
		subPipeline = p.parsePipeline()
	}

	if subPipeline == nil {
		return nil
	}

	return &MinusStep{
		Pipeline: subPipeline,
		Line:     minusLine,
		Column:   minusCol,
	}
}

func (p *QueryParser) parseFilterSpec(isEdgeFilter bool) *FilterSpec {
	// If edge filter, first token might be "edge"
	if isEdgeFilter {
		if p.peek().Type == gsl.TOKEN_EDGE {
			p.advance() // consume 'edge'
		}
		// Expect dot before attribute
		p.expect(gsl.TOKEN_DOT)
	}

	// Parse attribute name
	attrTok := p.peek()
	if attrTok.Type != gsl.TOKEN_IDENT {
		p.addError("expected attribute name at %d:%d", attrTok.Line, attrTok.Column)
		return nil
	}
	p.advance()
	attr := attrTok.Literal

	filterLine := attrTok.Line
	filterCol := attrTok.Column

	// Parse operator
	opTok := p.peek()
	var op string
	switch opTok.Type {
	case gsl.TOKEN_EQUALS:
		op = "="
		p.advance()
	case gsl.TOKEN_NE:
		op = "!="
		p.advance()
	case gsl.TOKEN_LT:
		op = "<"
		p.advance()
	case gsl.TOKEN_LE:
		op = "<="
		p.advance()
	case gsl.TOKEN_GT:
		op = ">"
		p.advance()
	case gsl.TOKEN_GE:
		op = ">="
		p.advance()
	default:
		p.addError("expected operator (=, !=, <, <=, >, >=) at %d:%d", opTok.Line, opTok.Column)
		return nil
	}

	// Parse value
	valTok := p.peek()
	var value interface{}
	switch valTok.Type {
	case gsl.TOKEN_STRING:
		p.advance()
		value = valTok.Literal
	case gsl.TOKEN_NUMBER:
		p.advance()
		n, err := strconv.ParseFloat(valTok.Literal, 64)
		if err != nil {
			p.addError("invalid number %q at %d:%d", valTok.Literal, valTok.Line, valTok.Column)
		}
		value = n
	case gsl.TOKEN_TRUE:
		p.advance()
		value = true
	case gsl.TOKEN_FALSE:
		p.advance()
		value = false
	default:
		p.addError("expected value (string, number, boolean) at %d:%d", valTok.Line, valTok.Column)
		return nil
	}

	return &FilterSpec{
		IsEdge: isEdgeFilter,
		Attr:   attr,
		Op:     op,
		Value:  value,
		Line:   filterLine,
		Column: filterCol,
	}
}
