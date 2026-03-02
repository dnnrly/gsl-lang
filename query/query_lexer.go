package query

import (
	"io"
	gsl "github.com/dnnrly/gsl-lang"
)

// QueryLexer is a specialized lexer for the GSL Query Language
type QueryLexer struct {
	input []rune
	pos   int
	line  int
	col   int
}

// NewQueryLexer creates a new query lexer from an io.Reader
func NewQueryLexer(r io.Reader) (*QueryLexer, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &QueryLexer{
		input: []rune(string(data)),
		pos:   0,
		line:  1,
		col:   1,
	}, nil
}

func (l *QueryLexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *QueryLexer) advance() rune {
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

func (l *QueryLexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *QueryLexer) skipComment() {
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.advance()
	}
}

func (l *QueryLexer) readString() gsl.Token {
	line := l.line
	col := l.col
	l.advance() // consume opening "

	var lit []rune
	for {
		if l.pos >= len(l.input) {
			return gsl.Token{Type: gsl.TOKEN_ILLEGAL, Literal: string(lit), Line: line, Column: col}
		}
		ch := l.advance()
		if ch == '"' {
			return gsl.Token{Type: gsl.TOKEN_STRING, Literal: string(lit), Line: line, Column: col}
		}
		if ch == '\\' {
			if l.pos >= len(l.input) {
				return gsl.Token{Type: gsl.TOKEN_ILLEGAL, Literal: string(lit), Line: line, Column: col}
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

func (l *QueryLexer) readNumber() gsl.Token {
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
	return gsl.Token{Type: gsl.TOKEN_NUMBER, Literal: string(l.input[start:l.pos]), Line: line, Column: col}
}

func (l *QueryLexer) readIdentifier() gsl.Token {
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

	// Query-specific keywords
	keywords := map[string]gsl.TokenType{
		"start":     gsl.TOKEN_START,
		"flow":      gsl.TOKEN_FLOW,
		"where":     gsl.TOKEN_WHERE,
		"minus":     gsl.TOKEN_MINUS,
		"union":     gsl.TOKEN_UNION,
		"intersect": gsl.TOKEN_INTERSECT,
		"in":        gsl.TOKEN_IN,
		"out":       gsl.TOKEN_OUT,
		"both":      gsl.TOKEN_BOTH,
		"recursive": gsl.TOKEN_RECURSIVE,
		"edge":      gsl.TOKEN_EDGE,
		"true":      gsl.TOKEN_TRUE,
		"false":     gsl.TOKEN_FALSE,
	}

	if tt, ok := keywords[literal]; ok {
		return gsl.Token{Type: tt, Literal: literal, Line: line, Column: col}
	}
	return gsl.Token{Type: gsl.TOKEN_IDENT, Literal: literal, Line: line, Column: col}
}

func (l *QueryLexer) NextToken() gsl.Token {
	for {
		l.skipWhitespace()
		if l.pos >= len(l.input) {
			return gsl.Token{Type: gsl.TOKEN_EOF, Literal: "", Line: l.line, Column: l.col}
		}

		ch := l.peek()

		if ch == '#' {
			l.skipComment()
			continue
		}

		line := l.line
		col := l.col

		switch ch {
		case '(':
			l.advance()
			return gsl.Token{Type: gsl.TOKEN_LPAREN, Literal: "(", Line: line, Column: col}
		case ')':
			l.advance()
			return gsl.Token{Type: gsl.TOKEN_RPAREN, Literal: ")", Line: line, Column: col}
		case '|':
			l.advance()
			return gsl.Token{Type: gsl.TOKEN_PIPE, Literal: "|", Line: line, Column: col}
		case '*':
			l.advance()
			return gsl.Token{Type: gsl.TOKEN_STAR, Literal: "*", Line: line, Column: col}
		case '.':
			l.advance()
			return gsl.Token{Type: gsl.TOKEN_DOT, Literal: ".", Line: line, Column: col}
		case '=':
			l.advance()
			return gsl.Token{Type: gsl.TOKEN_EQUALS, Literal: "=", Line: line, Column: col}
		case '!':
			l.advance()
			if l.pos < len(l.input) && l.input[l.pos] == '=' {
				l.advance()
				return gsl.Token{Type: gsl.TOKEN_NE, Literal: "!=", Line: line, Column: col}
			}
			return gsl.Token{Type: gsl.TOKEN_ILLEGAL, Literal: "!", Line: line, Column: col}
		case '<':
			l.advance()
			if l.pos < len(l.input) && l.input[l.pos] == '=' {
				l.advance()
				return gsl.Token{Type: gsl.TOKEN_LE, Literal: "<=", Line: line, Column: col}
			}
			return gsl.Token{Type: gsl.TOKEN_LT, Literal: "<", Line: line, Column: col}
		case '>':
			l.advance()
			if l.pos < len(l.input) && l.input[l.pos] == '=' {
				l.advance()
				return gsl.Token{Type: gsl.TOKEN_GE, Literal: ">=", Line: line, Column: col}
			}
			return gsl.Token{Type: gsl.TOKEN_GT, Literal: ">", Line: line, Column: col}
		case ',':
			l.advance()
			return gsl.Token{Type: gsl.TOKEN_COMMA, Literal: ",", Line: line, Column: col}
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
			return gsl.Token{Type: gsl.TOKEN_ILLEGAL, Literal: string(ch), Line: line, Column: col}
		}
	}
}

func (l *QueryLexer) Tokenize() []gsl.Token {
	var tokens []gsl.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == gsl.TOKEN_EOF {
			break
		}
	}
	return tokens
}
