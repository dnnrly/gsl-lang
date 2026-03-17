package query

import (
	"fmt"
	"sort"

	"github.com/dnnrly/gsl-lang"
)

// FromExpr selects a graph from context
type FromExpr struct {
	IsWildcard bool   // true if "*", false if named graph
	Name       string // graph name (empty if wildcard)
}

// Apply returns the selected graph
func (e *FromExpr) Apply(ctx *QueryContext, input Value) (Value, error) {
	if e.IsWildcard {
		// from * → return input graph
		return GraphValue{ctx.InputGraph}, nil
	}

	// from NAME → return named graph
	graph, exists := ctx.NamedGraphs[e.Name]
	if !exists {
		return nil, fmt.Errorf("named graph not found: %s", e.Name)
	}

	return GraphValue{graph}, nil
}

// BindExpr binds a subpipeline result to a named graph
// Execution: evaluate subpipeline, store result, return input unchanged
type BindExpr struct {
	Pipeline *Query // Subpipeline to evaluate
	Name     string // Name to bind result to
}

// Apply executes the subpipeline, stores result, returns input unchanged
func (e *BindExpr) Apply(ctx *QueryContext, input Value) (Value, error) {
	// Check if name already bound (immutability rule)
	if _, exists := ctx.NamedGraphs[e.Name]; exists {
		return nil, fmt.Errorf("named graph already bound: %s", e.Name)
	}

	// Execute subpipeline
	result, err := e.Pipeline.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("subpipeline failed: %w", err)
	}

	// Extract graph from result
	graphValue, ok := result.(GraphValue)
	if !ok {
		return nil, fmt.Errorf("subpipeline must return a graph")
	}

	// Store the result
	ctx.NamedGraphs[e.Name] = graphValue.Graph

	// Return input unchanged
	return input, nil
}

// TraversalConfig specifies traversal direction and depth
type TraversalConfig struct {
	Direction string // "in", "out", or "both"
	Depth     int    // number of hops; 0 means no traversal
}

// SubgraphExpr extracts a subgraph matching a predicate, with optional traversal
// Traversal expands the subgraph structurally (not predicate-based)
type SubgraphExpr struct {
	Pred      Predicate        // Predicate to match nodes or edges
	Traversal *TraversalConfig // nil if no traversal
}

// Apply filters graph to subgraph matching predicate, then optionally traverses
func (e *SubgraphExpr) Apply(ctx *QueryContext, input Value) (Value, error) {
	// Extract graph
	graphValue, ok := input.(GraphValue)
	if !ok {
		return nil, fmt.Errorf("subgraph requires a graph input")
	}

	graph := graphValue.Graph
	targetType := e.Pred.TargetType()

	// Detect mixed targets
	if targetType == "error" {
		return nil, fmt.Errorf("predicate mixes node and edge targets")
	}

	// Build base subgraph (returns nodes and edges)
	baseNodes, baseEdges := e.buildSubgraph(graph, targetType)

	// If traversal requested, expand from base nodes
	if e.Traversal != nil && e.Traversal.Depth > 0 {
		baseNodes = e.traverse(graph, baseNodes, e.Traversal)
		// After traversal, rebuild base edges (include all edges between nodes in traversal result)
		baseEdges = make(map[int]bool)
		for i, edge := range graph.Edges {
			if baseNodes[edge.From] && baseNodes[edge.To] {
				baseEdges[i] = true
			}
		}
	}

	// Construct result graph
	result := &gsl.Graph{
		Nodes: make(map[string]*gsl.Node),
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Copy all sets unchanged
	for id, set := range graph.Sets {
		result.Sets[id] = set
	}

	// Add matched nodes
	for id := range baseNodes {
		if node, exists := graph.Nodes[id]; exists {
			result.Nodes[id] = node
		}
	}

	// Add matched edges (or edges where both endpoints are in baseNodes if edge predicate)
	if targetType == "edge" {
		// For edge predicates, only add edges that matched the predicate
		for idx := range baseEdges {
			result.Edges = append(result.Edges, graph.Edges[idx])
		}
	} else {
		// For node predicates, add all edges where both endpoints are in baseNodes
		for _, edge := range graph.Edges {
			if baseNodes[edge.From] && baseNodes[edge.To] {
				result.Edges = append(result.Edges, edge)
			}
		}
	}

	return GraphValue{result}, nil
}

// buildSubgraph constructs the initial subgraph matching the predicate
// Returns a set of node IDs and a set of edge indices included in the subgraph
func (e *SubgraphExpr) buildSubgraph(graph *gsl.Graph, targetType string) (map[string]bool, map[int]bool) {
	nodes := make(map[string]bool)
	edges := make(map[int]bool)

	switch targetType {
	case "node":
		// Node predicate: include matching nodes
		for id, node := range graph.Nodes {
			if e.Pred.EvaluateNode(node) {
				nodes[id] = true
			}
		}

	case "edge":
		// Edge predicate: include endpoints of matching edges
		for i, edge := range graph.Edges {
			if e.Pred.EvaluateEdge(edge) {
				nodes[edge.From] = true
				nodes[edge.To] = true
				edges[i] = true
			}
		}

	default:
		// Empty target: try nodes first, then edges, then all
		for id, node := range graph.Nodes {
			if e.Pred.EvaluateNode(node) {
				nodes[id] = true
			}
		}

		if len(nodes) == 0 {
			// No matching nodes, try edges
			for i, edge := range graph.Edges {
				if e.Pred.EvaluateEdge(edge) {
					nodes[edge.From] = true
					nodes[edge.To] = true
					edges[i] = true
				}
			}
		}

		if len(nodes) == 0 {
			// No edges either, include all (for exists predicate)
			for id := range graph.Nodes {
				nodes[id] = true
			}
		}
	}

	return nodes, edges
}

// traverse expands the node set via breadth-first traversal
func (e *SubgraphExpr) traverse(graph *gsl.Graph, startNodes map[string]bool, cfg *TraversalConfig) map[string]bool {
	result := make(map[string]bool)
	for id := range startNodes {
		result[id] = true
	}

	visited := make(map[string]bool)
	for id := range startNodes {
		visited[id] = true
	}

	// Breadth-first traversal
	frontier := make([]string, 0)
	for id := range startNodes {
		frontier = append(frontier, id)
	}

	for depth := 0; depth < cfg.Depth && len(frontier) > 0; depth++ {
		nextFrontier := make([]string, 0)

		for _, nodeID := range frontier {
			neighbors := e.getNeighbors(graph, nodeID, cfg.Direction)
			for _, neighbor := range neighbors {
				if !visited[neighbor] {
					visited[neighbor] = true
					result[neighbor] = true
					nextFrontier = append(nextFrontier, neighbor)
				}
			}
		}

		frontier = nextFrontier
	}

	return result
}

// getNeighbors returns node IDs reachable from nodeID in the given direction
func (e *SubgraphExpr) getNeighbors(graph *gsl.Graph, nodeID string, direction string) []string {
	neighbors := make(map[string]bool)

	for _, edge := range graph.Edges {
		switch direction {
		case "out":
			// Outgoing edges: from→to
			if edge.From == nodeID {
				neighbors[edge.To] = true
			}
		case "in":
			// Incoming edges: to→from
			if edge.To == nodeID {
				neighbors[edge.From] = true
			}
		case "both":
			// Both directions
			if edge.From == nodeID {
				neighbors[edge.To] = true
			}
			if edge.To == nodeID {
				neighbors[edge.From] = true
			}
		}
	}

	result := make([]string, 0, len(neighbors))
	for id := range neighbors {
		result = append(result, id)
	}
	return result
}

// RemoveEdgeExpr removes edges matching a predicate
// Nodes remain in the graph
type RemoveEdgeExpr struct {
	Pred Predicate
}

// Apply removes edges matching the predicate
func (e *RemoveEdgeExpr) Apply(ctx *QueryContext, input Value) (Value, error) {
	// Extract graph
	graphValue, ok := input.(GraphValue)
	if !ok {
		return nil, fmt.Errorf("remove edge requires a graph input")
	}

	graph := graphValue.Graph

	// Detect target type
	targetType := e.Pred.TargetType()
	if targetType == "error" {
		return nil, fmt.Errorf("predicate mixes node and edge targets")
	}

	// Filter edges: keep those that don't match the predicate
	result := &gsl.Graph{
		Nodes: make(map[string]*gsl.Node),
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Copy all nodes and sets unchanged
	for id, node := range graph.Nodes {
		result.Nodes[id] = node
	}
	for id, set := range graph.Sets {
		result.Sets[id] = set
	}

	// Keep edges that don't match the predicate
	for _, edge := range graph.Edges {
		// Check if this edge matches the predicate (only for edge predicates)
		if targetType == "edge" {
			if !e.Pred.EvaluateEdge(edge) {
				result.Edges = append(result.Edges, edge)
			}
		} else if targetType == "" {
			// Empty target: try edges first (per removal semantics)
			if !e.Pred.EvaluateEdge(edge) {
				result.Edges = append(result.Edges, edge)
			}
		} else {
			// Node predicate on remove edge: error or keep all?
			// Per spec, "remove edge where" uses edge predicates only
			// For now, keep all edges if not an edge predicate
			result.Edges = append(result.Edges, edge)
		}
	}

	return GraphValue{result}, nil
}

// RemoveAttributeExpr removes an attribute from nodes or edges matching a predicate
type RemoveAttributeExpr struct {
	Target string    // "node" or "edge"
	Attr   string    // attribute name to remove
	Pred   Predicate // predicate to select targets
}

// Apply removes attributes from matching nodes or edges
func (e *RemoveAttributeExpr) Apply(ctx *QueryContext, input Value) (Value, error) {
	// Extract graph
	graphValue, ok := input.(GraphValue)
	if !ok {
		return nil, fmt.Errorf("remove attribute requires a graph input")
	}

	graph := graphValue.Graph

	// Detect mixed targets
	targetType := e.Pred.TargetType()
	if targetType == "error" {
		return nil, fmt.Errorf("predicate mixes node and edge targets")
	}

	// Create result graph (clone structure)
	result := &gsl.Graph{
		Nodes: make(map[string]*gsl.Node),
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Copy all sets
	for id, set := range graph.Sets {
		result.Sets[id] = set
	}

	// Process nodes
	for id, node := range graph.Nodes {
		nodeCopy := *node
		nodeCopy.Attributes = make(map[string]interface{})

		// Copy attributes from original
		for k, v := range node.Attributes {
			nodeCopy.Attributes[k] = v
		}

		// Remove attribute if this node matches the predicate
		if e.Target == "node" && (targetType == "" || targetType == "node") {
			if e.Pred.EvaluateNode(node) {
				delete(nodeCopy.Attributes, e.Attr)
			}
		}

		result.Nodes[id] = &nodeCopy
	}

	// Process edges
	for _, edge := range graph.Edges {
		edgeCopy := *edge
		edgeCopy.Attributes = make(map[string]interface{})

		// Copy attributes from original
		for k, v := range edge.Attributes {
			edgeCopy.Attributes[k] = v
		}

		// Remove attribute if this edge matches the predicate
		if e.Target == "edge" && (targetType == "" || targetType == "edge") {
			if e.Pred.EvaluateEdge(edge) {
				delete(edgeCopy.Attributes, e.Attr)
			}
		}

		result.Edges = append(result.Edges, &edgeCopy)
	}

	return GraphValue{result}, nil
}

// RemoveOrphansExpr removes nodes with no incident edges
// A self-loop counts as an incident edge
type RemoveOrphansExpr struct{}

// Apply removes isolated nodes
func (e *RemoveOrphansExpr) Apply(ctx *QueryContext, input Value) (Value, error) {
	// Extract graph
	graphValue, ok := input.(GraphValue)
	if !ok {
		return nil, fmt.Errorf("remove orphans requires a graph input")
	}

	graph := graphValue.Graph

	// Identify nodes with at least one incident edge
	hasEdge := make(map[string]bool)
	for _, edge := range graph.Edges {
		hasEdge[edge.From] = true
		hasEdge[edge.To] = true
	}

	// Build result graph
	result := &gsl.Graph{
		Nodes: make(map[string]*gsl.Node),
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Copy all sets unchanged
	for id, set := range graph.Sets {
		result.Sets[id] = set
	}

	// Keep only nodes with incident edges
	for id, node := range graph.Nodes {
		if hasEdge[id] {
			result.Nodes[id] = node
		}
	}

	// Copy all edges (endpoints still exist)
	result.Edges = append(result.Edges, graph.Edges...)

	return GraphValue{result}, nil
}

// CollapseExpr merges multiple nodes matching a predicate into a single node
// Edge rewriting and attribute merging follows the spec
type CollapseExpr struct {
	NodeID  string    // target node ID for collapsed nodes
	Pred    Predicate // predicate to select nodes to collapse
}

// Apply merges nodes matching the predicate into a single node
func (e *CollapseExpr) Apply(ctx *QueryContext, input Value) (Value, error) {
	// Extract graph
	graphValue, ok := input.(GraphValue)
	if !ok {
		return nil, fmt.Errorf("collapse requires a graph input")
	}

	graph := graphValue.Graph

	// Detect target type - collapse only works on nodes
	targetType := e.Pred.TargetType()
	if targetType == "error" {
		return nil, fmt.Errorf("predicate mixes node and edge targets")
	}
	if targetType == "edge" {
		return nil, fmt.Errorf("collapse only works with node predicates")
	}

	// Find nodes to collapse
	collapsedSet := make(map[string]bool)
	for id, node := range graph.Nodes {
		if e.Pred.EvaluateNode(node) {
			collapsedSet[id] = true
		}
	}

	// If no nodes match, return unchanged
	if len(collapsedSet) == 0 {
		return GraphValue{graph}, nil
	}

	// Create result graph
	result := &gsl.Graph{
		Nodes: make(map[string]*gsl.Node),
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Copy all sets unchanged
	for id, set := range graph.Sets {
		result.Sets[id] = set
	}

	// Copy nodes that are not being collapsed
	for id, node := range graph.Nodes {
		if !collapsedSet[id] {
			result.Nodes[id] = node
		}
	}

	// Create the collapsed node with merged attributes
	mergedAttrs := make(map[string]interface{})
	for _, id := range e.sortedNodeIDs(graph, collapsedSet) {
		// Apply attributes in order (last-write-wins)
		node := graph.Nodes[id]
		if node != nil && node.Attributes != nil {
			for k, v := range node.Attributes {
				mergedAttrs[k] = v
			}
		}
	}

	result.Nodes[e.NodeID] = &gsl.Node{
		ID:         e.NodeID,
		Attributes: mergedAttrs,
		Sets:       make(map[string]struct{}),
	}

	// Process edges: rewrite external edges, remove internal edges
	seenEdges := make(map[string]bool) // for deduplication

	for _, edge := range graph.Edges {
		fromCollapsed := collapsedSet[edge.From]
		toCollapsed := collapsedSet[edge.To]

		// Skip internal edges (both endpoints are collapsed)
		if fromCollapsed && toCollapsed {
			continue
		}

		// Determine new source and target
		newFrom := edge.From
		if fromCollapsed {
			newFrom = e.NodeID
		}

		newTo := edge.To
		if toCollapsed {
			newTo = e.NodeID
		}

		// Create new edge
		newEdge := &gsl.Edge{
			From:       newFrom,
			To:         newTo,
			Attributes: edge.Attributes,
			Sets:       edge.Sets,
		}

		// Deduplication key (per spec): from, to, attributes
		key := e.edgeKey(newEdge)
		if !seenEdges[key] {
			seenEdges[key] = true
			result.Edges = append(result.Edges, newEdge)
		}
	}

	return GraphValue{result}, nil
}

// sortedNodeIDs returns collapsed node IDs in a deterministic order
func (e *CollapseExpr) sortedNodeIDs(graph *gsl.Graph, collapsedSet map[string]bool) []string {
	var ids []string
	for id := range collapsedSet {
		ids = append(ids, id)
	}
	// Simple string sort for determinism
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			if ids[i] > ids[j] {
				ids[i], ids[j] = ids[j], ids[i]
			}
		}
	}
	return ids
}

// edgeKey creates a deduplication key for an edge
func (e *CollapseExpr) edgeKey(edge *gsl.Edge) string {
	// Key: from|to|attribute_hash
	// For simplicity, we use from|to and check if exact edge exists
	// In a real implementation, we'd hash attributes
	// Sort attribute keys for deterministic output
	keys := make([]string, 0, len(edge.Attributes))
	for k := range edge.Attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	var attrs string
	for _, k := range keys {
		if attrs != "" {
			attrs += ","
		}
		attrs += fmt.Sprintf("%s:%v", k, edge.Attributes[k])
	}
	return edge.From + "|" + edge.To + "|" + attrs
}

// MakeExpr assigns or overwrites an attribute on nodes or edges matching a predicate
type MakeExpr struct {
	Target string      // "node" or "edge"
	Attr   string      // attribute name
	Value  interface{} // value to assign
	Pred   Predicate   // predicate to select targets
}

// Apply assigns attributes to matching nodes or edges
func (e *MakeExpr) Apply(ctx *QueryContext, input Value) (Value, error) {
	// Extract graph
	graphValue, ok := input.(GraphValue)
	if !ok {
		return nil, fmt.Errorf("make requires a graph input")
	}

	graph := graphValue.Graph

	// Detect mixed targets
	targetType := e.Pred.TargetType()
	if targetType == "error" {
		return nil, fmt.Errorf("predicate mixes node and edge targets")
	}

	// Create result graph (clone structure)
	result := &gsl.Graph{
		Nodes: make(map[string]*gsl.Node),
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Copy all sets
	for id, set := range graph.Sets {
		result.Sets[id] = set
	}

	// Process nodes
	for id, node := range graph.Nodes {
		nodeCopy := *node
		nodeCopy.Attributes = make(map[string]interface{})

		// Copy attributes from original
		for k, v := range node.Attributes {
			nodeCopy.Attributes[k] = v
		}

		// Set attribute if this node matches the predicate
		if e.Target == "node" && (targetType == "" || targetType == "node") {
			if e.Pred.EvaluateNode(node) {
				nodeCopy.Attributes[e.Attr] = e.Value
			}
		}

		result.Nodes[id] = &nodeCopy
	}

	// Process edges
	for _, edge := range graph.Edges {
		edgeCopy := *edge
		edgeCopy.Attributes = make(map[string]interface{})

		// Copy attributes from original
		for k, v := range edge.Attributes {
			edgeCopy.Attributes[k] = v
		}

		// Set attribute if this edge matches the predicate
		if e.Target == "edge" && (targetType == "" || targetType == "edge") {
			if e.Pred.EvaluateEdge(edge) {
				edgeCopy.Attributes[e.Attr] = e.Value
			}
		}

		result.Edges = append(result.Edges, &edgeCopy)
	}

	return GraphValue{result}, nil
}

// GraphAlgebraExpr combines two named graphs using an operator
// Operators: +, &, -, ^
type GraphAlgebraExpr struct {
	LeftRef  string // graph name (or "*" for input)
	RightRef string // graph name (or "*" for input)
	Operator string // "+", "&", "-", "^"
}

// Apply combines two graphs according to the operator
func (e *GraphAlgebraExpr) Apply(ctx *QueryContext, input Value) (Value, error) {
	// Resolve left graph
	left, err := e.resolveGraph(ctx, e.LeftRef)
	if err != nil {
		return nil, err
	}

	// Resolve right graph
	right, err := e.resolveGraph(ctx, e.RightRef)
	if err != nil {
		return nil, err
	}

	// Apply operator
	switch e.Operator {
	case "+":
		return GraphValue{e.union(left, right)}, nil
	case "&":
		return GraphValue{e.intersection(left, right)}, nil
	case "-":
		return GraphValue{e.difference(left, right)}, nil
	case "^":
		return GraphValue{e.symmetricDifference(left, right)}, nil
	default:
		return nil, fmt.Errorf("unknown algebra operator: %s", e.Operator)
	}
}

// resolveGraph returns the graph referenced by a name
func (e *GraphAlgebraExpr) resolveGraph(ctx *QueryContext, ref string) (*gsl.Graph, error) {
	if ref == "*" {
		return ctx.InputGraph, nil
	}

	graph, exists := ctx.NamedGraphs[ref]
	if !exists {
		return nil, fmt.Errorf("named graph not found: %s", ref)
	}

	return graph, nil
}

// union combines all nodes and edges from both graphs
// For shared nodes: left attributes first, right overwrites conflicts (last-write-wins)
func (e *GraphAlgebraExpr) union(left *gsl.Graph, right *gsl.Graph) *gsl.Graph {
	result := &gsl.Graph{
		Nodes: make(map[string]*gsl.Node),
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Copy left nodes and sets
	for id, node := range left.Nodes {
		nodeCopy := *node
		nodeCopy.Attributes = make(map[string]interface{})
		for k, v := range node.Attributes {
			nodeCopy.Attributes[k] = v
		}
		nodeCopy.Sets = make(map[string]struct{})
		for s := range node.Sets {
			nodeCopy.Sets[s] = struct{}{}
		}
		result.Nodes[id] = &nodeCopy
	}

	// Copy left edges
	result.Edges = append(result.Edges, left.Edges...)

	// Copy left sets
	for id, set := range left.Sets {
		setCopy := *set
		setCopy.Attributes = make(map[string]interface{})
		for k, v := range set.Attributes {
			setCopy.Attributes[k] = v
		}
		result.Sets[id] = &setCopy
	}

	// Merge right nodes (right overwrites left for attributes)
	for id, node := range right.Nodes {
		if existing, exists := result.Nodes[id]; exists {
			// Node exists in both: right overwrites left (last-write-wins)
			for k, v := range node.Attributes {
				existing.Attributes[k] = v
			}
			// Add right's set memberships
			for s := range node.Sets {
				existing.Sets[s] = struct{}{}
			}
		} else {
			// New node from right
			nodeCopy := *node
			nodeCopy.Attributes = make(map[string]interface{})
			for k, v := range node.Attributes {
				nodeCopy.Attributes[k] = v
			}
			nodeCopy.Sets = make(map[string]struct{})
			for s := range node.Sets {
				nodeCopy.Sets[s] = struct{}{}
			}
			result.Nodes[id] = &nodeCopy
		}
	}

	// Add right edges (duplicates preserved)
	result.Edges = append(result.Edges, right.Edges...)

	// Merge right sets
	for id, set := range right.Sets {
		if existing, exists := result.Sets[id]; exists {
			// Set exists in both: right overwrites left
			for k, v := range set.Attributes {
				existing.Attributes[k] = v
			}
		} else {
			// New set from right
			setCopy := *set
			setCopy.Attributes = make(map[string]interface{})
			for k, v := range set.Attributes {
				setCopy.Attributes[k] = v
			}
			result.Sets[id] = &setCopy
		}
	}

	return result
}

// intersection returns only nodes/edges present in both graphs
// A node is in intersection if it exists in both graphs
// An edge is in intersection if endpoints both exist in result and edge exists in both graphs
func (e *GraphAlgebraExpr) intersection(left *gsl.Graph, right *gsl.Graph) *gsl.Graph {
	result := &gsl.Graph{
		Nodes: make(map[string]*gsl.Node),
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Include nodes that exist in both graphs
	for id, node := range left.Nodes {
		if _, exists := right.Nodes[id]; exists {
			// Node exists in both
			nodeCopy := *node
			nodeCopy.Attributes = make(map[string]interface{})
			for k, v := range node.Attributes {
				nodeCopy.Attributes[k] = v
			}
			nodeCopy.Sets = make(map[string]struct{})
			for s := range node.Sets {
				nodeCopy.Sets[s] = struct{}{}
			}
			result.Nodes[id] = &nodeCopy
		}
	}

	// Build edge set from right for fast lookup
	rightEdges := make(map[string]bool) // key: from|to
	for _, edge := range right.Edges {
		rightEdges[edge.From+"|"+edge.To] = true
	}

	// Include edges that exist in both graphs (both endpoints in result, edge in both)
	for _, edge := range left.Edges {
		key := edge.From + "|" + edge.To
		if rightEdges[key] && result.Nodes[edge.From] != nil && result.Nodes[edge.To] != nil {
			edgeCopy := *edge
			edgeCopy.Attributes = make(map[string]interface{})
			for k, v := range edge.Attributes {
				edgeCopy.Attributes[k] = v
			}
			edgeCopy.Sets = make(map[string]struct{})
			for s := range edge.Sets {
				edgeCopy.Sets[s] = struct{}{}
			}
			result.Edges = append(result.Edges, &edgeCopy)
		}
	}

	// Include sets that exist in both graphs
	for id, set := range left.Sets {
		if _, exists := right.Sets[id]; exists {
			setCopy := *set
			setCopy.Attributes = make(map[string]interface{})
			for k, v := range set.Attributes {
				setCopy.Attributes[k] = v
			}
			result.Sets[id] = &setCopy
		}
	}

	return result
}

// difference returns nodes/edges in left but not in right
// A node is in difference if it exists in left but not right
// An edge is in difference if endpoints both exist in result and edge exists in left but not right
func (e *GraphAlgebraExpr) difference(left *gsl.Graph, right *gsl.Graph) *gsl.Graph {
	result := &gsl.Graph{
		Nodes: make(map[string]*gsl.Node),
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Include nodes that exist in left but not right
	for id, node := range left.Nodes {
		if _, exists := right.Nodes[id]; !exists {
			nodeCopy := *node
			nodeCopy.Attributes = make(map[string]interface{})
			for k, v := range node.Attributes {
				nodeCopy.Attributes[k] = v
			}
			nodeCopy.Sets = make(map[string]struct{})
			for s := range node.Sets {
				nodeCopy.Sets[s] = struct{}{}
			}
			result.Nodes[id] = &nodeCopy
		}
	}

	// Build edge set from right for fast lookup
	rightEdges := make(map[string]bool)
	for _, edge := range right.Edges {
		rightEdges[edge.From+"|"+edge.To] = true
	}

	// Include edges that exist in left but not right
	for _, edge := range left.Edges {
		key := edge.From + "|" + edge.To
		if !rightEdges[key] && result.Nodes[edge.From] != nil && result.Nodes[edge.To] != nil {
			edgeCopy := *edge
			edgeCopy.Attributes = make(map[string]interface{})
			for k, v := range edge.Attributes {
				edgeCopy.Attributes[k] = v
			}
			edgeCopy.Sets = make(map[string]struct{})
			for s := range edge.Sets {
				edgeCopy.Sets[s] = struct{}{}
			}
			result.Edges = append(result.Edges, &edgeCopy)
		}
	}

	// Include sets that exist in left but not right
	for id, set := range left.Sets {
		if _, exists := right.Sets[id]; !exists {
			setCopy := *set
			setCopy.Attributes = make(map[string]interface{})
			for k, v := range set.Attributes {
				setCopy.Attributes[k] = v
			}
			result.Sets[id] = &setCopy
		}
	}

	return result
}

// symmetricDifference returns nodes/edges in exactly one graph
// A node is in symDiff if it exists in left or right but not both
// An edge is in symDiff if endpoints both exist in result and edge is in exactly one graph
func (e *GraphAlgebraExpr) symmetricDifference(left *gsl.Graph, right *gsl.Graph) *gsl.Graph {
	result := &gsl.Graph{
		Nodes: make(map[string]*gsl.Node),
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Include nodes from left that don't exist in right
	for id, node := range left.Nodes {
		if _, exists := right.Nodes[id]; !exists {
			nodeCopy := *node
			nodeCopy.Attributes = make(map[string]interface{})
			for k, v := range node.Attributes {
				nodeCopy.Attributes[k] = v
			}
			nodeCopy.Sets = make(map[string]struct{})
			for s := range node.Sets {
				nodeCopy.Sets[s] = struct{}{}
			}
			result.Nodes[id] = &nodeCopy
		}
	}

	// Include nodes from right that don't exist in left
	for id, node := range right.Nodes {
		if _, exists := left.Nodes[id]; !exists {
			nodeCopy := *node
			nodeCopy.Attributes = make(map[string]interface{})
			for k, v := range node.Attributes {
				nodeCopy.Attributes[k] = v
			}
			nodeCopy.Sets = make(map[string]struct{})
			for s := range node.Sets {
				nodeCopy.Sets[s] = struct{}{}
			}
			result.Nodes[id] = &nodeCopy
		}
	}

	// Build edge sets for lookup
	leftEdges := make(map[string]bool)
	rightEdges := make(map[string]bool)
	for _, edge := range left.Edges {
		leftEdges[edge.From+"|"+edge.To] = true
	}
	for _, edge := range right.Edges {
		rightEdges[edge.From+"|"+edge.To] = true
	}

	// Include edges from left that don't exist in right
	for _, edge := range left.Edges {
		key := edge.From + "|" + edge.To
		if !rightEdges[key] && result.Nodes[edge.From] != nil && result.Nodes[edge.To] != nil {
			edgeCopy := *edge
			edgeCopy.Attributes = make(map[string]interface{})
			for k, v := range edge.Attributes {
				edgeCopy.Attributes[k] = v
			}
			edgeCopy.Sets = make(map[string]struct{})
			for s := range edge.Sets {
				edgeCopy.Sets[s] = struct{}{}
			}
			result.Edges = append(result.Edges, &edgeCopy)
		}
	}

	// Include edges from right that don't exist in left
	for _, edge := range right.Edges {
		key := edge.From + "|" + edge.To
		if !leftEdges[key] && result.Nodes[edge.From] != nil && result.Nodes[edge.To] != nil {
			edgeCopy := *edge
			edgeCopy.Attributes = make(map[string]interface{})
			for k, v := range edge.Attributes {
				edgeCopy.Attributes[k] = v
			}
			edgeCopy.Sets = make(map[string]struct{})
			for s := range edge.Sets {
				edgeCopy.Sets[s] = struct{}{}
			}
			result.Edges = append(result.Edges, &edgeCopy)
		}
	}

	// Include sets from left that don't exist in right
	for id, set := range left.Sets {
		if _, exists := right.Sets[id]; !exists {
			setCopy := *set
			setCopy.Attributes = make(map[string]interface{})
			for k, v := range set.Attributes {
				setCopy.Attributes[k] = v
			}
			result.Sets[id] = &setCopy
		}
	}

	// Include sets from right that don't exist in left
	for id, set := range right.Sets {
		if _, exists := left.Sets[id]; !exists {
			setCopy := *set
			setCopy.Attributes = make(map[string]interface{})
			for k, v := range set.Attributes {
				setCopy.Attributes[k] = v
			}
			result.Sets[id] = &setCopy
		}
	}

	return result
}
