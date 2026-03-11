package query

import (
	"fmt"

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

	// Build base subgraph
	baseNodes := e.buildSubgraph(graph, targetType)

	// If traversal requested, expand from base nodes
	if e.Traversal != nil && e.Traversal.Depth > 0 {
		baseNodes = e.traverse(graph, baseNodes, e.Traversal)
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

	// Add edges where both endpoints are in baseNodes
	for _, edge := range graph.Edges {
		if baseNodes[edge.From] && baseNodes[edge.To] {
			result.Edges = append(result.Edges, edge)
		}
	}

	return GraphValue{result}, nil
}

// buildSubgraph constructs the initial subgraph matching the predicate
// Returns a set of node IDs included in the subgraph
func (e *SubgraphExpr) buildSubgraph(graph *gsl.Graph, targetType string) map[string]bool {
	nodes := make(map[string]bool)

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
		for _, edge := range graph.Edges {
			if e.Pred.EvaluateEdge(edge) {
				nodes[edge.From] = true
				nodes[edge.To] = true
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
			for _, edge := range graph.Edges {
				if e.Pred.EvaluateEdge(edge) {
					nodes[edge.From] = true
					nodes[edge.To] = true
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

	return nodes
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
