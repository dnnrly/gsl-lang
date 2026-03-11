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

// SubgraphExpr extracts a subgraph matching a predicate
// Without traversal, includes matching nodes and edges between them
type SubgraphExpr struct {
	Pred Predicate // Predicate to match nodes or edges
}

// Apply filters graph to subgraph matching predicate
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

	// Clone the graph (start with everything)
	result := &gsl.Graph{
		Nodes: make(map[string]*gsl.Node),
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Copy all sets unchanged
	for id, set := range graph.Sets {
		result.Sets[id] = set
	}

	switch targetType {
	case "node":
		// Node predicate: include matching nodes + edges between them
		matchedNodes := make(map[string]bool)

		// Filter nodes
		for id, node := range graph.Nodes {
			if e.Pred.EvaluateNode(node) {
				matchedNodes[id] = true
				result.Nodes[id] = node
			}
		}

		// Include edges where both source and target are matched
		for _, edge := range graph.Edges {
			if matchedNodes[edge.From] && matchedNodes[edge.To] {
				result.Edges = append(result.Edges, edge)
			}
		}

	case "edge":
		// Edge predicate: include matching edges + their source/target nodes
		matchedNodeIDs := make(map[string]bool)

		// Filter edges and collect node IDs
		for _, edge := range graph.Edges {
			if e.Pred.EvaluateEdge(edge) {
				result.Edges = append(result.Edges, edge)
				matchedNodeIDs[edge.From] = true
				matchedNodeIDs[edge.To] = true
			}
		}

		// Include source and target nodes
		for id := range matchedNodeIDs {
			if node, exists := graph.Nodes[id]; exists {
				result.Nodes[id] = node
			}
		}

	default:
		// Empty target: check if predicate works on nodes or edges
		// Try evaluating on nodes first (for set membership predicates without target)
		matchedNodes := make(map[string]bool)
		hasMatchingNodes := false

		// Try as node predicate
		for id, node := range graph.Nodes {
			if e.Pred.EvaluateNode(node) {
				matchedNodes[id] = true
				result.Nodes[id] = node
				hasMatchingNodes = true
			}
		}

		if hasMatchingNodes {
			// Include edges where both source and target are matched
			for _, edge := range graph.Edges {
				if matchedNodes[edge.From] && matchedNodes[edge.To] {
					result.Edges = append(result.Edges, edge)
				}
			}
		} else {
			// Try as edge predicate (for exists or other universal predicates)
			matchedNodeIDs := make(map[string]bool)
			for _, edge := range graph.Edges {
				if e.Pred.EvaluateEdge(edge) {
					result.Edges = append(result.Edges, edge)
					matchedNodeIDs[edge.From] = true
					matchedNodeIDs[edge.To] = true
				}
			}

			// Include source and target nodes
			for id := range matchedNodeIDs {
				if node, exists := graph.Nodes[id]; exists {
					result.Nodes[id] = node
				}
			}

			// If no edges matched either, include all nodes (for exists predicate)
			if len(matchedNodeIDs) == 0 {
				for id, node := range graph.Nodes {
					result.Nodes[id] = node
				}
				for _, edge := range graph.Edges {
					result.Edges = append(result.Edges, edge)
				}
			}
		}
	}

	return GraphValue{result}, nil
}
