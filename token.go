package gsl

type TokenType int

const (
	// Special
	TOKEN_ILLEGAL TokenType = iota
	TOKEN_EOF

	// Literals
	TOKEN_IDENT
	TOKEN_STRING
	TOKEN_NUMBER

	// Keywords
	TOKEN_NODE
	TOKEN_SET
	TOKEN_TRUE
	TOKEN_FALSE

	// Symbols
	TOKEN_LBRACKET // [
	TOKEN_RBRACKET // ]
	TOKEN_LBRACE   // {
	TOKEN_RBRACE   // }
	TOKEN_COMMA    // ,
	TOKEN_EQUALS   // =
	TOKEN_ARROW    // ->
	TOKEN_AT       // @
	TOKEN_COLON    // :

	// Query language tokens
	TOKEN_LPAREN   // (
	TOKEN_RPAREN   // )
	TOKEN_PIPE     // |
	TOKEN_STAR     // *
	TOKEN_NE       // !=
	TOKEN_LT       // <
	TOKEN_LE       // <=
	TOKEN_GT       // >
	TOKEN_GE       // >=
	TOKEN_DOT      // .

	// Query keywords
	TOKEN_START
	TOKEN_FLOW
	TOKEN_WHERE
	TOKEN_MINUS
	TOKEN_UNION
	TOKEN_INTERSECT
	TOKEN_IN
	TOKEN_OUT
	TOKEN_BOTH
	TOKEN_RECURSIVE
	TOKEN_EDGE
)

var tokenNames = map[TokenType]string{
	TOKEN_ILLEGAL:   "ILLEGAL",
	TOKEN_EOF:       "EOF",
	TOKEN_IDENT:     "IDENT",
	TOKEN_STRING:    "STRING",
	TOKEN_NUMBER:    "NUMBER",
	TOKEN_NODE:      "NODE",
	TOKEN_SET:       "SET",
	TOKEN_TRUE:      "TRUE",
	TOKEN_FALSE:     "FALSE",
	TOKEN_LBRACKET:  "LBRACKET",
	TOKEN_RBRACKET:  "RBRACKET",
	TOKEN_LBRACE:    "LBRACE",
	TOKEN_RBRACE:    "RBRACE",
	TOKEN_COMMA:     "COMMA",
	TOKEN_EQUALS:    "EQUALS",
	TOKEN_ARROW:     "ARROW",
	TOKEN_AT:        "AT",
	TOKEN_COLON:     "COLON",
	TOKEN_LPAREN:    "LPAREN",
	TOKEN_RPAREN:    "RPAREN",
	TOKEN_PIPE:      "PIPE",
	TOKEN_STAR:      "STAR",
	TOKEN_NE:       "NE",
	TOKEN_LT:       "LT",
	TOKEN_LE:       "LE",
	TOKEN_GT:       "GT",
	TOKEN_GE:       "GE",
	TOKEN_DOT:      "DOT",
	TOKEN_START:     "START",
	TOKEN_FLOW:      "FLOW",
	TOKEN_WHERE:     "WHERE",
	TOKEN_MINUS:     "MINUS",
	TOKEN_UNION:     "UNION",
	TOKEN_INTERSECT: "INTERSECT",
	TOKEN_IN:        "IN",
	TOKEN_OUT:       "OUT",
	TOKEN_BOTH:      "BOTH",
	TOKEN_RECURSIVE: "RECURSIVE",
	TOKEN_EDGE:      "EDGE",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}
