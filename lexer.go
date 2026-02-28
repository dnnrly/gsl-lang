package gsl

import (
	"io"
)

var keywords = map[string]TokenType{
	"node":  TOKEN_NODE,
	"set":   TOKEN_SET,
	"true":  TOKEN_TRUE,
	"false": TOKEN_FALSE,
}

type Lexer struct {
	input []rune
	pos   int
	line  int
	col   int
}

func NewLexer(r io.Reader) (*Lexer, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &Lexer{
		input: []rune(string(data)),
		pos:   0,
		line:  1,
		col:   1,
	}, nil
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) advance() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := l.input[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) skipComment() {
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.advance()
	}
}

func (l *Lexer) readString() Token {
	line := l.line
	col := l.col
	l.advance() // consume opening "

	var lit []rune
	for {
		if l.pos >= len(l.input) {
			return Token{Type: TOKEN_ILLEGAL, Literal: string(lit), Line: line, Column: col}
		}
		ch := l.advance()
		if ch == '"' {
			return Token{Type: TOKEN_STRING, Literal: string(lit), Line: line, Column: col}
		}
		if ch == '\\' {
			if l.pos >= len(l.input) {
				return Token{Type: TOKEN_ILLEGAL, Literal: string(lit), Line: line, Column: col}
			}
			esc := l.advance()
			switch esc {
			case '"':
				lit = append(lit, '"')
			case '\\':
				lit = append(lit, '\\')
			case 'n':
				lit = append(lit, '\n')
			case 't':
				lit = append(lit, '\t')
			default:
				lit = append(lit, '\\', esc)
			}
		} else {
			lit = append(lit, ch)
		}
	}
}

func (l *Lexer) readNumber() Token {
	line := l.line
	col := l.col
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] >= '0' && l.input[l.pos] <= '9' {
		l.advance()
	}
	if l.pos < len(l.input) && l.input[l.pos] == '.' {
		l.advance()
		for l.pos < len(l.input) && l.input[l.pos] >= '0' && l.input[l.pos] <= '9' {
			l.advance()
		}
	}
	return Token{Type: TOKEN_NUMBER, Literal: string(l.input[start:l.pos]), Line: line, Column: col}
}

func (l *Lexer) readIdentifier() Token {
	line := l.line
	col := l.col
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' {
			l.advance()
		} else {
			break
		}
	}
	literal := string(l.input[start:l.pos])
	if tt, ok := keywords[literal]; ok {
		return Token{Type: tt, Literal: literal, Line: line, Column: col}
	}
	return Token{Type: TOKEN_IDENT, Literal: literal, Line: line, Column: col}
}

func (l *Lexer) NextToken() Token {
	for {
		l.skipWhitespace()
		if l.pos >= len(l.input) {
			return Token{Type: TOKEN_EOF, Literal: "", Line: l.line, Column: l.col}
		}

		ch := l.peek()

		if ch == '#' {
			l.skipComment()
			continue
		}

		line := l.line
		col := l.col

		switch ch {
		case '[':
			l.advance()
			return Token{Type: TOKEN_LBRACKET, Literal: "[", Line: line, Column: col}
		case ']':
			l.advance()
			return Token{Type: TOKEN_RBRACKET, Literal: "]", Line: line, Column: col}
		case '{':
			l.advance()
			return Token{Type: TOKEN_LBRACE, Literal: "{", Line: line, Column: col}
		case '}':
			l.advance()
			return Token{Type: TOKEN_RBRACE, Literal: "}", Line: line, Column: col}
		case ',':
			l.advance()
			return Token{Type: TOKEN_COMMA, Literal: ",", Line: line, Column: col}
		case '=':
			l.advance()
			return Token{Type: TOKEN_EQUALS, Literal: "=", Line: line, Column: col}
		case '@':
			l.advance()
			return Token{Type: TOKEN_AT, Literal: "@", Line: line, Column: col}
		case ':':
			l.advance()
			return Token{Type: TOKEN_COLON, Literal: ":", Line: line, Column: col}
		case '-':
			l.advance()
			if l.pos < len(l.input) && l.input[l.pos] == '>' {
				l.advance()
				return Token{Type: TOKEN_ARROW, Literal: "->", Line: line, Column: col}
			}
			return Token{Type: TOKEN_ILLEGAL, Literal: "-", Line: line, Column: col}
		case '"':
			return l.readString()
		default:
			if ch >= '0' && ch <= '9' {
				return l.readNumber()
			}
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || ch == '_' {
				return l.readIdentifier()
			}
			l.advance()
			return Token{Type: TOKEN_ILLEGAL, Literal: string(ch), Line: line, Column: col}
		}
	}
}

func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TOKEN_EOF {
			break
		}
	}
	return tokens
}
