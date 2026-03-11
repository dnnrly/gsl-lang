package query

import "fmt"

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
