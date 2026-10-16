package gsl

import (
	"strings"
	"testing"
)

func tokenize(t *testing.T, input string) []Token {
	t.Helper()
	l, err := NewLexer(strings.NewReader(input))
	if err != nil {
		t.Fatalf("NewLexer error: %v", err)
	}
	return l.Tokenize()
}

func assertToken(t *testing.T, tok Token, expectedType TokenType, expectedLiteral string) {
	t.Helper()
	if tok.Type != expectedType {
		t.Errorf("expected type %s, got %s (literal=%q)", expectedType, tok.Type, tok.Literal)
	}
	if tok.Literal != expectedLiteral {
		t.Errorf("expected literal %q, got %q", expectedLiteral, tok.Literal)
	}
}

func TestSingleCharSymbols(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"[", TOKEN_LBRACKET},
		{"]", TOKEN_RBRACKET},
		{"{", TOKEN_LBRACE},
		{"}", TOKEN_RBRACE},
		{",", TOKEN_COMMA},
		{"=", TOKEN_EQUALS},
		{"@", TOKEN_AT},
		{":", TOKEN_COLON},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := tokenize(t, tt.input)
			if len(tokens) != 2 {
				t.Fatalf("expected 2 tokens, got %d", len(tokens))
			}
			assertToken(t, tokens[0], tt.expected, tt.input)
			assertToken(t, tokens[1], TOKEN_EOF, "")
		})
	}
}

func TestArrow(t *testing.T) {
	tokens := tokenize(t, "->")
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	assertToken(t, tokens[0], TOKEN_ARROW, "->")
	assertToken(t, tokens[1], TOKEN_EOF, "")
}

func TestKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"node", TOKEN_NODE},
		{"set", TOKEN_SET},
		{"true", TOKEN_TRUE},
		{"false", TOKEN_FALSE},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := tokenize(t, tt.input)
			if len(tokens) != 2 {
				t.Fatalf("expected 2 tokens, got %d", len(tokens))
			}
			assertToken(t, tokens[0], tt.expected, tt.input)
		})
	}
}

func TestIdentifier(t *testing.T) {
	tokens := tokenize(t, "myVar_1")
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	assertToken(t, tokens[0], TOKEN_IDENT, "myVar_1")
}

func TestString(t *testing.T) {
	tokens := tokenize(t, `"hello world"`)
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	assertToken(t, tokens[0], TOKEN_STRING, "hello world")
}

func TestIntegerNumber(t *testing.T) {
	tokens := tokenize(t, "42")
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	assertToken(t, tokens[0], TOKEN_NUMBER, "42")
}

func TestDecimalNumber(t *testing.T) {
	tokens := tokenize(t, "3.14")
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	assertToken(t, tokens[0], TOKEN_NUMBER, "3.14")
}

func TestNodeStatement(t *testing.T) {
	tokens := tokenize(t, `node A: "Start" @flow`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_NODE, "node"},
		{TOKEN_IDENT, "A"},
		{TOKEN_COLON, ":"},
		{TOKEN_STRING, "Start"},
		{TOKEN_AT, "@"},
		{TOKEN_IDENT, "flow"},
		{TOKEN_EOF, ""},
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, e := range expected {
		assertToken(t, tokens[i], e.typ, e.lit)
	}
}

func TestEdgeStatement(t *testing.T) {
	tokens := tokenize(t, `A->B [weight=1.2] @flow`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_IDENT, "A"},
		{TOKEN_ARROW, "->"},
		{TOKEN_IDENT, "B"},
		{TOKEN_LBRACKET, "["},
		{TOKEN_IDENT, "weight"},
		{TOKEN_EQUALS, "="},
		{TOKEN_NUMBER, "1.2"},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_AT, "@"},
		{TOKEN_IDENT, "flow"},
		{TOKEN_EOF, ""},
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, e := range expected {
		assertToken(t, tokens[i], e.typ, e.lit)
	}
}

func TestSetStatement(t *testing.T) {
	tokens := tokenize(t, `set cluster [visible]`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_SET, "set"},
		{TOKEN_IDENT, "cluster"},
		{TOKEN_LBRACKET, "["},
		{TOKEN_IDENT, "visible"},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_EOF, ""},
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, e := range expected {
		assertToken(t, tokens[i], e.typ, e.lit)
	}
}

func TestCommentsSkipped(t *testing.T) {
	tokens := tokenize(t, "# comment\nnode A")
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	assertToken(t, tokens[0], TOKEN_NODE, "node")
	assertToken(t, tokens[1], TOKEN_IDENT, "A")
	assertToken(t, tokens[2], TOKEN_EOF, "")
}

func TestUnterminatedString(t *testing.T) {
	tokens := tokenize(t, `"unterminated`)
	if len(tokens) < 1 {
		t.Fatal("expected at least 1 token")
	}
	if tokens[0].Type != TOKEN_ILLEGAL {
		t.Errorf("expected TOKEN_ILLEGAL, got %s", tokens[0].Type)
	}
}

func TestBareDash(t *testing.T) {
	tokens := tokenize(t, "-")
	if len(tokens) < 1 {
		t.Fatal("expected at least 1 token")
	}
	assertToken(t, tokens[0], TOKEN_ILLEGAL, "-")
}

func TestStringEscapes(t *testing.T) {
	tokens := tokenize(t, `"hello\"world"`)
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	assertToken(t, tokens[0], TOKEN_STRING, `hello"world`)
}

func TestStringEscapeBackslash(t *testing.T) {
	tokens := tokenize(t, `"a\\b"`)
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	assertToken(t, tokens[0], TOKEN_STRING, `a\b`)
}

func TestStringEscapeNewlineTab(t *testing.T) {
	tokens := tokenize(t, `"a\nb\tc"`)
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	assertToken(t, tokens[0], TOKEN_STRING, "a\nb\tc")
}

func TestBooleans(t *testing.T) {
	tokens := tokenize(t, "true false")
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	assertToken(t, tokens[0], TOKEN_TRUE, "true")
	assertToken(t, tokens[1], TOKEN_FALSE, "false")
}

func TestPositionTracking(t *testing.T) {
	tokens := tokenize(t, "node A\nset B")
	// node: line 1, col 1
	if tokens[0].Line != 1 || tokens[0].Column != 1 {
		t.Errorf("node: expected 1:1, got %d:%d", tokens[0].Line, tokens[0].Column)
	}
	// A: line 1, col 6
	if tokens[1].Line != 1 || tokens[1].Column != 6 {
		t.Errorf("A: expected 1:6, got %d:%d", tokens[1].Line, tokens[1].Column)
	}
	// set: line 2, col 1
	if tokens[2].Line != 2 || tokens[2].Column != 1 {
		t.Errorf("set: expected 2:1, got %d:%d", tokens[2].Line, tokens[2].Column)
	}
	// B: line 2, col 5
	if tokens[3].Line != 2 || tokens[3].Column != 5 {
		t.Errorf("B: expected 2:5, got %d:%d", tokens[3].Line, tokens[3].Column)
	}
}

func TestEmptyInput(t *testing.T) {
	tokens := tokenize(t, "")
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	assertToken(t, tokens[0], TOKEN_EOF, "")
}

func TestTokenTypeString(t *testing.T) {
	if TOKEN_NODE.String() != "NODE" {
		t.Errorf("expected NODE, got %s", TOKEN_NODE.String())
	}
	if TOKEN_ARROW.String() != "ARROW" {
		t.Errorf("expected ARROW, got %s", TOKEN_ARROW.String())
	}
	if TokenType(999).String() != "UNKNOWN" {
		t.Errorf("expected UNKNOWN, got %s", TokenType(999).String())
	}
}

func TestInlineComment(t *testing.T) {
	tokens := tokenize(t, "node A # this is a comment\nset B")
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_NODE, "node"},
		{TOKEN_IDENT, "A"},
		{TOKEN_SET, "set"},
		{TOKEN_IDENT, "B"},
		{TOKEN_EOF, ""},
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, e := range expected {
		assertToken(t, tokens[i], e.typ, e.lit)
	}
}

func TestUnderscoreIdentifier(t *testing.T) {
	tokens := tokenize(t, "_foo")
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	assertToken(t, tokens[0], TOKEN_IDENT, "_foo")
}
