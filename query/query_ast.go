package query

import "strconv"

// QueryError types as defined in GQL v1.0 spec
type QueryErrorType int

const (
	ErrorInvalidQuery QueryErrorType = iota
	ErrorUnknownNodeID
	ErrorInvalidPredicate
)

func (e QueryErrorType) String() string {
	switch e {
	case ErrorInvalidQuery:
		return "InvalidQuery"
	case ErrorUnknownNodeID:
		return "UnknownNodeID"
	case ErrorInvalidPredicate:
		return "InvalidPredicate"
	default:
		return "UnknownError"
	}
}

// QueryError represents a query execution error per GQL v1.0 spec
type QueryError struct {
	Type    QueryErrorType
	Message string
	Line    int
	Column  int
}

func (e *QueryError) Error() string {
	if e.Line > 0 && e.Column > 0 {
		return e.Type.String() + " at " + strconv.Itoa(e.Line) + ":" + strconv.Itoa(e.Column) + ": " + e.Message
	}
	return e.Type.String() + ": " + e.Message
}

// Query AST for GSL Query Language

// Query is the root node of a query AST
type Query struct {
	Root Step
}

// Step is implemented by all pipeline steps
type Step interface {
	stepNode()
}

// Pipeline represents a sequence of steps
type Pipeline struct {
	Steps []Step
}

func (p *Pipeline) stepNode() {}

// StartStep initiates a pipeline with one or more node IDs
type StartStep struct {
	NodeIDs []string
	Line    int
	Column  int
}

func (s *StartStep) stepNode() {}

// FlowStep traverses edges in a specified direction
type FlowStep struct {
	Direction  string       // "in", "out", "both"
	Recursive  bool         // true if * or "recursive"
	EdgeFilter *FilterSpec  // nil if no edge filter
	Line       int
	Column     int
}

func (f *FlowStep) stepNode() {}

// FilterStep filters nodes by attribute conditions
type FilterStep struct {
	Filter *FilterSpec
	Line   int
	Column int
}

func (f *FilterStep) stepNode() {}

// MinusStep removes nodes matching a sub-pipeline
type MinusStep struct {
	Pipeline *Pipeline
	Line     int
	Column   int
}

func (m *MinusStep) stepNode() {}

// FilterSpec represents an attribute comparison
type FilterSpec struct {
	IsEdge   bool        // true for edge.attr, false for node attr
	Attr     string      // attribute name
	Op       string      // "=", "!=", "contains", "matches"
	Value    interface{} // string, float64, or bool
	Line     int
	Column   int
}

// CombinatorExpr combines two pipelines with an operator
type CombinatorExpr struct {
	Type     string      // "union", "intersect", "minus"
	Left     *Pipeline   // or *CombinatorExpr
	Right    *Pipeline   // or *CombinatorExpr
	IsParens bool        // true if parentheses were used
	Line     int
	Column   int
}

func (c *CombinatorExpr) stepNode() {}
