package query

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var queryLexer = lexer.MustSimple([]lexer.SimpleRule{
	{"Comment", `#[^\n]*`},
	{"String", `"[^"]*"`},
	{"Number", `[-+]?(?:\d+\.?\d*|\.\d+)`},
	{"Operator", `==|!=`},
	{"Assign", `=`},
	{"Pipe", `\|`},
	{"LParen", `\(`},
	{"RParen", `\)`},
	{"Dot", `\.`},
	{"At", `@`},
	{"AlgebraOp", `[+&\-^]`},
	{"Star", `\*`},
	{"Ident", `[a-zA-Z_][a-zA-Z0-9_]*`},
	{"Whitespace", `\s+`},
})

// QueryAST is the top-level grammar rule
type QueryAST struct {
	Expressions []*ExpressionAST `@@ ( Pipe @@ )*`
}

// ExpressionAST is a union of all expression types — first match wins
type ExpressionAST struct {
	Binding      *BindingAST      `  @@`
	Subgraph     *SubgraphAST     `| @@`
	Make         *MakeAST         `| @@`
	Remove       *RemoveAST       `| @@`
	Collapse     *CollapseAST     `| @@`
	From         *FromAST         `| @@`
	GraphAlgebra *GraphAlgebraAST `| @@`
}

// BindingAST: "(" pipeline ")" "as" NAME
type BindingAST struct {
	Pipeline *QueryAST `LParen @@ RParen`
	Name     string    `"as" @Ident`
}

// SubgraphAST: "subgraph" predicate ["traverse" direction depth]
type SubgraphAST struct {
	Predicate *PredicateAST `"subgraph" @@`
	Traverse  *TraverseAST  `( @@ )?`
}

// TraverseAST: "traverse" direction depth
type TraverseAST struct {
	Direction string `"traverse" @( "in" | "out" | "both" )`
	Depth     string `@( "all" | Number )`
}

// MakeAST: "make" attr_path "=" value "where" predicate
type MakeAST struct {
	Path      *AttributePathAST `"make" @@`
	Value     *ValueAST         `Assign @@`
	Predicate *PredicateAST     `"where" @@`
}

// RemoveAST: union of remove forms
type RemoveAST struct {
	Orphans       bool              `  "remove" ( @"orphans"`
	EdgePredicate *PredicateAST     `| "edge" "where" @@`
	AttrPath      *AttributePathAST `| @@`
	AttrPredicate *PredicateAST     `  "where" @@ )`
}

// CollapseAST: "collapse" "into" id "where" predicate
type CollapseAST struct {
	NodeID    string        `"collapse" "into" @Ident`
	Predicate *PredicateAST `"where" @@`
}

// FromAST: "from" ("*" | NAME)
type FromAST struct {
	Wildcard bool   `"from" ( @Star`
	Name     string `| @Ident )`
}

// GraphAlgebraAST: REF OP REF
type GraphAlgebraAST struct {
	Left     string `@( Star | Ident )`
	Operator string `@AlgebraOp`
	Right    string `@( Star | Ident )`
}

// PredicateAST: term { "AND" term }
type PredicateAST struct {
	Terms []*PredicateTermAST `@@ ( "AND" @@ )*`
}

// PredicateTermAST: union of predicate forms
type PredicateTermAST struct {
	SetPredicate  *SetPredicateAST  `  @@`
	AttrPredicate *AttrPredicateAST `| @@`
}

// SetPredicateAST: ("node"|"edge") ["not"] "in" "@" ident
type SetPredicateAST struct {
	Element string `@( "node" | "edge" )`
	Not     bool   `@"not"?`
	SetName string `"in" At @Ident`
}

// AttrPredicateAST: path (op value | "exists" | "not" "exists")
type AttrPredicateAST struct {
	Path      *AttributePathAST `@@`
	Operator  string            `( @( Operator )`
	Value     *ValueAST         `  @@`
	Exists    bool              `| @"exists"`
	NotExists bool              `| "not" @"exists" )`
}

// AttributePathAST: ("node"|"edge") "." ident
type AttributePathAST struct {
	Element string `@( "node" | "edge" )`
	Attr    string `Dot @Ident`
}

// ValueAST: string | bool | number | bare ident
type ValueAST struct {
	String *string  `  @String`
	Bool   *Boolean `| @( "true" | "false" )`
	Number *string  `| @Number`
	Ident  *string  `| @Ident`
}

// Boolean captures true/false literals
type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true"
	return nil
}

var queryParser = participle.MustBuild[QueryAST](
	participle.Lexer(queryLexer),
	participle.Elide("Comment", "Whitespace"),
	participle.UseLookahead(3),
)
