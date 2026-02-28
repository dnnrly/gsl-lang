package gsl

// NodeRef is a distinct type for node references in attribute values.
type NodeRef string

// statement is implemented by nodeDecl, edgeDecl, setDecl
type statement interface {
	stmtNode()
}

type program struct {
	statements []statement
}

type nodeDecl struct {
	name        string
	textValue   *string     // from ": STRING"
	attrs       []attribute // from [...]
	block       []nodeDecl  // from { ... }
	memberships []string    // from @ident
	line        int
	col         int
}

type edgeDecl struct {
	left        []string
	right       []string
	textValue   *string
	attrs       []attribute
	memberships []string
	line        int
	col         int
}

type setDecl struct {
	name  string
	attrs []attribute
	line  int
	col   int
}

type attribute struct {
	key   string
	value *attrValue // nil = empty/flag
	line  int
	col   int
}

type attrValue struct {
	kind    attrValueKind
	strVal  string
	numVal  float64
	boolVal bool
	refVal  string // for NodeRef
}

type attrValueKind int

const (
	valueString  attrValueKind = iota
	valueNumber
	valueBool
	valueNodeRef
)

func (n *nodeDecl) stmtNode() {}
func (e *edgeDecl) stmtNode() {}
func (s *setDecl) stmtNode()  {}
