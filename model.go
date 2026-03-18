package gsl

// Graph is the top-level semantic model produced by parsing a GSL document.
type Graph struct {
	nodes map[string]*Node
	sets  map[string]*Set
	edges []*Edge
}

// Node represents a node in the graph.
type Node struct {
	ID         string
	Attributes map[string]interface{}
	Sets       map[string]struct{}
	Parent     *string // cached from Attributes["parent"] if it's a NodeRef
}

// Edge represents a directed edge in the graph.
type Edge struct {
	From       string
	To         string
	Attributes map[string]interface{}
	Sets       map[string]struct{}
}

// Set represents a named set/grouping.
type Set struct {
	ID         string
	Attributes map[string]interface{}
	declared   bool // true if explicitly declared via `set`
}
