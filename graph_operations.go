package gsl

import "fmt"

// NewGraph creates a new empty graph.
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
		sets:  make(map[string]*Set),
		edges: make([]*Edge, 0),
	}
}

// GetNodes returns a read-only copy of the graph's node map.
// Changes to the returned map do not affect the graph.
func (g *Graph) GetNodes() map[string]*Node {
	if g == nil {
		return nil
	}
	result := make(map[string]*Node, len(g.nodes))
	for k, v := range g.nodes {
		result[k] = v
	}
	return result
}

// GetEdges returns a read-only copy of the graph's edge slice.
// Changes to the returned slice do not affect the graph.
func (g *Graph) GetEdges() []*Edge {
	if g == nil {
		return nil
	}
	result := make([]*Edge, len(g.edges))
	copy(result, g.edges)
	return result
}

// GetSets returns a read-only copy of the graph's set map.
// Changes to the returned map do not affect the graph.
func (g *Graph) GetSets() map[string]*Set {
	if g == nil {
		return nil
	}
	result := make(map[string]*Set, len(g.sets))
	for k, v := range g.sets {
		result[k] = v
	}
	return result
}

// GetNode returns the node with the given ID, or nil if not found.
func (g *Graph) GetNode(id string) *Node {
	if g == nil {
		return nil
	}
	return g.nodes[id]
}

// AddNode adds or updates a node in the graph.
// Returns the created/updated node and an error if validation fails.
// If a node with the same ID already exists, it is returned and no error occurs.
func (g *Graph) AddNode(id string, attrs map[string]interface{}) (*Node, error) {
	if g == nil {
		return nil, fmt.Errorf("cannot add node to nil graph")
	}
	if id == "" {
		return nil, fmt.Errorf("node ID cannot be empty")
	}

	// Check if node already exists
	if n, ok := g.nodes[id]; ok {
		return n, nil
	}

	// Create new node
	n := &Node{
		ID:         id,
		Attributes: make(AttributeMap),
		Sets:       make(map[string]struct{}),
	}

	// Copy attributes if provided
	for k, v := range attrs {
		n.Attributes[k] = v
	}

	// Cache parent field if present
	if p, ok := n.Attributes["parent"]; ok {
		if ref, isRef := p.(NodeRef); isRef {
			s := string(ref)
			n.Parent = &s
		}
	}

	g.nodes[id] = n
	return n, nil
}

// AddEdge adds an edge to the graph.
// Returns the created edge and an error if validation fails.
// Validates that both from and to nodes exist.
func (g *Graph) AddEdge(from, to string, attrs map[string]interface{}) (*Edge, error) {
	if g == nil {
		return nil, fmt.Errorf("cannot add edge to nil graph")
	}
	if from == "" {
		return nil, fmt.Errorf("edge from node ID cannot be empty")
	}
	if to == "" {
		return nil, fmt.Errorf("edge to node ID cannot be empty")
	}

	// Validate that both nodes exist
	if _, ok := g.nodes[from]; !ok {
		return nil, fmt.Errorf("edge from node %q does not exist", from)
	}
	if _, ok := g.nodes[to]; !ok {
		return nil, fmt.Errorf("edge to node %q does not exist", to)
	}

	// Create new edge
	e := &Edge{
		From:       from,
		To:         to,
		Attributes: make(AttributeMap),
		Sets:       make(map[string]struct{}),
	}

	// Copy attributes if provided
	for k, v := range attrs {
		e.Attributes[k] = v
	}

	g.edges = append(g.edges, e)
	return e, nil
}

// RemoveNode removes a node from the graph.
// Returns an error if the node does not exist or has dangling edges.
func (g *Graph) RemoveNode(id string) error {
	if g == nil {
		return fmt.Errorf("cannot remove node from nil graph")
	}
	if id == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	// Check if node exists
	if _, ok := g.nodes[id]; !ok {
		return fmt.Errorf("node %q does not exist", id)
	}

	// Check for dangling edges
	for _, e := range g.edges {
		if e.From == id || e.To == id {
			return fmt.Errorf("cannot remove node %q: has dangling edges", id)
		}
	}

	delete(g.nodes, id)
	return nil
}

// RemoveEdge removes an edge from the graph.
// Returns an error if the edge does not exist.
func (g *Graph) RemoveEdge(from, to string) error {
	if g == nil {
		return fmt.Errorf("cannot remove edge from nil graph")
	}
	if from == "" {
		return fmt.Errorf("edge from node ID cannot be empty")
	}
	if to == "" {
		return fmt.Errorf("edge to node ID cannot be empty")
	}

	// Find and remove the edge
	for i, e := range g.edges {
		if e.From == from && e.To == to {
			g.edges = append(g.edges[:i], g.edges[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("edge from %q to %q does not exist", from, to)
}

// SetInternalState sets the internal state of a graph.
// This is a testing-only method and should not be used in production code.
// It directly sets nodes, edges, and sets without validation.
func (g *Graph) SetInternalState(nodes map[string]*Node, edges []*Edge, sets map[string]*Set) {
	if g == nil {
		return
	}
	if nodes != nil {
		g.nodes = nodes
	}
	if edges != nil {
		g.edges = edges
	}
	if sets != nil {
		g.sets = sets
	}
}

// AddSet adds or updates a set in the graph.
// Returns the created/updated set and an error if validation fails.
// If a set with the same ID already exists, it is returned and no error occurs.
func (g *Graph) AddSet(id string, attrs map[string]interface{}) (*Set, error) {
	if g == nil {
		return nil, fmt.Errorf("cannot add set to nil graph")
	}
	if id == "" {
		return nil, fmt.Errorf("set ID cannot be empty")
	}

	// Check if set already exists
	if s, ok := g.sets[id]; ok {
		return s, nil
	}

	// Create new set
	s := &Set{
		ID:         id,
		Attributes: make(AttributeMap),
	}

	// Copy attributes if provided
	for k, v := range attrs {
		s.Attributes[k] = v
	}

	g.sets[id] = s
	return s, nil
}

// AddExistingNode adds a pre-created node to the graph, preserving all node state.
// Used by graph operations that need to preserve node attributes and set memberships.
func (g *Graph) AddExistingNode(node *Node) error {
	if g == nil {
		return fmt.Errorf("cannot add node to nil graph")
	}
	if node == nil || node.ID == "" {
		return fmt.Errorf("node is nil or has empty ID")
	}
	g.nodes[node.ID] = node
	return nil
}

// AddExistingEdge adds a pre-created edge to the graph, preserving all edge state.
// Used by graph operations that need to preserve edge attributes and set memberships.
func (g *Graph) AddExistingEdge(edge *Edge) error {
	if g == nil {
		return fmt.Errorf("cannot add edge to nil graph")
	}
	if edge == nil {
		return fmt.Errorf("edge is nil")
	}
	g.edges = append(g.edges, edge)
	return nil
}

// AddExistingSet adds a pre-created set to the graph, preserving all set state.
// Used by graph operations that need to preserve set attributes.
func (g *Graph) AddExistingSet(set *Set) error {
	if g == nil {
		return fmt.Errorf("cannot add set to nil graph")
	}
	if set == nil || set.ID == "" {
		return fmt.Errorf("set is nil or has empty ID")
	}
	g.sets[set.ID] = set
	return nil
}

// Clone creates a deep copy of the graph.
// All nodes, edges, and sets are copied with their attributes and set memberships.
// The cloned graph is independent: mutations to the clone do not affect the original.
// copyAttrs creates a deep copy of an attribute map.
func copyAttrs(src AttributeMap) AttributeMap {
	if src == nil {
		return nil
	}
	dst := make(AttributeMap, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// copySets creates a deep copy of a set membership map.
func copySets(src map[string]struct{}) map[string]struct{} {
	if src == nil {
		return nil
	}
	dst := make(map[string]struct{}, len(src))
	for k := range src {
		dst[k] = struct{}{}
	}
	return dst
}

func (g *Graph) Clone() *Graph {
	if g == nil {
		return nil
	}

	cloned := NewGraph()

	// Deep copy nodes
	for id, node := range g.nodes {
		newNode := &Node{
			ID:         node.ID,
			Attributes: copyAttrs(node.Attributes),
			Sets:       copySets(node.Sets),
		}
		// Copy parent reference if present
		if node.Parent != nil {
			parent := *node.Parent
			newNode.Parent = &parent
		}
		cloned.nodes[id] = newNode
	}

	// Deep copy edges
	for _, edge := range g.edges {
		newEdge := &Edge{
			From:       edge.From,
			To:         edge.To,
			Attributes: copyAttrs(edge.Attributes),
			Sets:       copySets(edge.Sets),
		}
		cloned.edges = append(cloned.edges, newEdge)
	}

	// Deep copy sets
	for id, set := range g.sets {
		newSet := &Set{
			ID:         set.ID,
			Attributes: copyAttrs(set.Attributes),
		}
		cloned.sets[id] = newSet
	}

	return cloned
}
