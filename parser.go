package gsl

import (
	"fmt"
	"strconv"
	"strings"
)

type parser struct {
	tokens []Token
	pos    int
	errors []error
}

func newParser(tokens []Token) *parser {
	return &parser{tokens: tokens}
}

func (p *parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

func (p *parser) advance() Token {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func (p *parser) expect(t TokenType) Token {
	tok := p.peek()
	if tok.Type != t {
		p.addError("expected %s, got %s (%q) at %d:%d", t, tok.Type, tok.Literal, tok.Line, tok.Column)
		return tok
	}
	return p.advance()
}

func (p *parser) addError(format string, args ...interface{}) {
	p.errors = append(p.errors, fmt.Errorf(format, args...))
}

func (p *parser) isReserved(tok Token) bool {
	switch tok.Type {
	case TOKEN_NODE, TOKEN_SET, TOKEN_TRUE, TOKEN_FALSE:
		return true
	}
	return false
}

func (p *parser) expectIdent() Token {
	tok := p.peek()
	if p.isReserved(tok) {
		p.addError("reserved keyword %q cannot be used as identifier at %d:%d", tok.Literal, tok.Line, tok.Column)
		p.advance()
		return Token{Type: TOKEN_IDENT, Literal: tok.Literal, Line: tok.Line, Column: tok.Column}
	}
	return p.expect(TOKEN_IDENT)
}

func parse(input string) (*program, []error) {
	l, err := NewLexer(strings.NewReader(input))
	if err != nil {
		return nil, []error{err}
	}
	tokens := l.Tokenize()
	p := newParser(tokens)
	prog := p.parseProgram()
	return prog, p.errors
}

func (p *parser) parseProgram() *program {
	prog := &program{}
	for p.peek().Type != TOKEN_EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			prog.statements = append(prog.statements, stmt)
		}
	}
	return prog
}

func (p *parser) parseStatement() statement {
	tok := p.peek()
	switch tok.Type {
	case TOKEN_NODE:
		return p.parseNodeDecl()
	case TOKEN_SET:
		return p.parseSetDecl()
	case TOKEN_IDENT:
		return p.parseEdgeDecl()
	default:
		p.addError("unexpected token %s (%q) at %d:%d", tok.Type, tok.Literal, tok.Line, tok.Column)
		p.advance()
		return nil
	}
}

func (p *parser) parseNodeDecl() *nodeDecl {
	tok := p.advance() // consume TOKEN_NODE
	nd := &nodeDecl{line: tok.Line, col: tok.Column}

	nameTok := p.peek()
	if nameTok.Type != TOKEN_IDENT {
		if p.isReserved(nameTok) {
			p.addError("reserved keyword %q cannot be used as identifier at %d:%d", nameTok.Literal, nameTok.Line, nameTok.Column)
			p.advance()
			return nil
		}
		p.addError("expected identifier after 'node', got %s (%q) at %d:%d", nameTok.Type, nameTok.Literal, nameTok.Line, nameTok.Column)
		return nil
	}
	nd.name = p.advance().Literal

	// Parse optional suffix
	switch p.peek().Type {
	case TOKEN_LBRACKET:
		nd.attrs = p.parseAttributeList(true)
	case TOKEN_COLON:
		p.advance() // consume ':'
		strTok := p.expect(TOKEN_STRING)
		if strTok.Type == TOKEN_STRING {
			nd.textValue = &strTok.Literal
		}
	case TOKEN_LBRACE:
		nd.block = p.parseBlock(nd.name)
	}

	nd.memberships = p.parseMemberships()
	return nd
}

func (p *parser) parseEdgeDecl() *edgeDecl {
	firstTok := p.peek()
	ed := &edgeDecl{line: firstTok.Line, col: firstTok.Column}

	left := p.parseNodeList()
	ed.left = left

	p.expect(TOKEN_ARROW)

	right := p.parseNodeList()
	ed.right = right

	if len(left) > 1 && len(right) > 1 {
		p.addError("both sides of edge cannot be grouped at %d:%d", firstTok.Line, firstTok.Column)
	}

	// Parse optional edge suffix
	switch p.peek().Type {
	case TOKEN_LBRACKET:
		ed.attrs = p.parseAttributeList(false)
	case TOKEN_COLON:
		p.advance() // consume ':'
		strTok := p.expect(TOKEN_STRING)
		if strTok.Type == TOKEN_STRING {
			ed.textValue = &strTok.Literal
		}
	}

	ed.memberships = p.parseMemberships()
	return ed
}

func (p *parser) parseSetDecl() *setDecl {
	tok := p.advance() // consume TOKEN_SET
	sd := &setDecl{line: tok.Line, col: tok.Column}

	nameTok := p.expectIdent()
	if nameTok.Type != TOKEN_IDENT {
		return nil
	}
	sd.name = nameTok.Literal

	if p.peek().Type == TOKEN_LBRACKET {
		sd.attrs = p.parseAttributeList(false)
	}

	return sd
}

func (p *parser) parseNodeList() []string {
	var names []string
	ident := p.expectIdent()
	if ident.Type == TOKEN_IDENT {
		names = append(names, ident.Literal)
	}
	for p.peek().Type == TOKEN_COMMA {
		p.advance() // consume ','
		ident = p.expectIdent()
		if ident.Type == TOKEN_IDENT {
			names = append(names, ident.Literal)
		}
	}
	return names
}

func (p *parser) parseAttributeList(inNodeContext bool) []attribute {
	p.advance() // consume '['
	var attrs []attribute
	seen := map[string]bool{}

	if p.peek().Type != TOKEN_RBRACKET {
		attr := p.parseAttribute(inNodeContext)
		if seen[attr.key] {
			p.addError("duplicate attribute key %q at %d:%d", attr.key, attr.line, attr.col)
		}
		seen[attr.key] = true
		attrs = append(attrs, attr)

		for p.peek().Type == TOKEN_COMMA {
			p.advance() // consume ','
			attr = p.parseAttribute(inNodeContext)
			if seen[attr.key] {
				p.addError("duplicate attribute key %q at %d:%d", attr.key, attr.line, attr.col)
			}
			seen[attr.key] = true
			attrs = append(attrs, attr)
		}
	}

	p.expect(TOKEN_RBRACKET)
	return attrs
}

func (p *parser) parseAttribute(inNodeContext bool) attribute {
	tok := p.expectIdent()
	attr := attribute{
		key:  tok.Literal,
		line: tok.Line,
		col:  tok.Column,
	}

	if p.peek().Type == TOKEN_EQUALS {
		p.advance() // consume '='
		attr.value = p.parseValue(inNodeContext)
	}

	return attr
}

func (p *parser) parseValue(inNodeContext bool) *attrValue {
	tok := p.peek()
	switch tok.Type {
	case TOKEN_STRING:
		p.advance()
		return &attrValue{kind: valueString, strVal: tok.Literal}
	case TOKEN_NUMBER:
		p.advance()
		n, err := strconv.ParseFloat(tok.Literal, 64)
		if err != nil {
			p.addError("invalid number %q at %d:%d", tok.Literal, tok.Line, tok.Column)
		}
		return &attrValue{kind: valueNumber, numVal: n}
	case TOKEN_TRUE:
		p.advance()
		return &attrValue{kind: valueBool, boolVal: true}
	case TOKEN_FALSE:
		p.advance()
		return &attrValue{kind: valueBool, boolVal: false}
	case TOKEN_IDENT:
		if inNodeContext {
			p.advance()
			return &attrValue{kind: valueNodeRef, refVal: tok.Literal}
		}
		p.addError("node references are not allowed in this context at %d:%d", tok.Line, tok.Column)
		p.advance()
		return &attrValue{kind: valueNodeRef, refVal: tok.Literal}
	default:
		p.addError("expected value, got %s (%q) at %d:%d", tok.Type, tok.Literal, tok.Line, tok.Column)
		p.advance()
		return &attrValue{}
	}
}

func (p *parser) parseMemberships() []string {
	var memberships []string
	for p.peek().Type == TOKEN_AT {
		p.advance() // consume '@'
		ident := p.expectIdent()
		if ident.Type == TOKEN_IDENT {
			memberships = append(memberships, ident.Literal)
		}
	}
	return memberships
}

func (p *parser) parseBlock(parentName string) []nodeDecl {
	p.advance() // consume '{'
	var children []nodeDecl
	for p.peek().Type != TOKEN_RBRACE && p.peek().Type != TOKEN_EOF {
		if p.peek().Type == TOKEN_NODE {
			nd := p.parseNodeDecl()
			if nd != nil {
				children = append(children, *nd)
			}
		} else {
			tok := p.peek()
			p.addError("expected node declaration inside block, got %s (%q) at %d:%d", tok.Type, tok.Literal, tok.Line, tok.Column)
			p.advance()
		}
	}
	p.expect(TOKEN_RBRACE)
	return children
}
