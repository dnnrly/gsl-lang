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
)

var tokenNames = map[TokenType]string{
	TOKEN_ILLEGAL:  "ILLEGAL",
	TOKEN_EOF:      "EOF",
	TOKEN_IDENT:    "IDENT",
	TOKEN_STRING:   "STRING",
	TOKEN_NUMBER:   "NUMBER",
	TOKEN_NODE:     "NODE",
	TOKEN_SET:      "SET",
	TOKEN_TRUE:     "TRUE",
	TOKEN_FALSE:    "FALSE",
	TOKEN_LBRACKET: "LBRACKET",
	TOKEN_RBRACKET: "RBRACKET",
	TOKEN_LBRACE:   "LBRACE",
	TOKEN_RBRACE:   "RBRACE",
	TOKEN_COMMA:    "COMMA",
	TOKEN_EQUALS:   "EQUALS",
	TOKEN_ARROW:    "ARROW",
	TOKEN_AT:       "AT",
	TOKEN_COLON:    "COLON",
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
