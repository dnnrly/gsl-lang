package query

import (
	"testing"

	"github.com/dnnrly/gsl-lang"
)

func TestIntegrationComplexPipeline(t *testing.T) {
	// Test: Extract critical services, collapse by team, mark as high-priority
	input := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"api": {ID: "api", Attributes: map[string]interface{}{"type": "service", "team": "platform"}, Sets: make(map[string]struct{})},
			"db": {ID: "db", Attributes: map[string]interface{}{"type": "database", "team": "platform"}, Sets: make(map[string]struct{})},
			"cache": {ID: "cache", Attributes: map[string]interface{}{"type": "cache", "team": "platform"}, Sets: make(map[string]struct{})},
			"web": {ID: "web", Attributes: map[string]interface{}{"type": "service", "team": "frontend"}, Sets: make(map[string]struct{})},
		},
		Edges: []*gsl.Edge{
			{From: "web", To: "api", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
			{From: "api", To: "db", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
			{From: "api", To: "cache", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
		},
		Sets: make(map[string]*gsl.Set),
	}

	// Pipeline: subgraph node.team = "platform" | collapse into PLATFORM where exists | make node.critical = true where exists
	// Note: subgraph filters to team=platform nodes only (api, db, cache)
	// Then collapse merges them into PLATFORM
	// Then make marks PLATFORM as critical
	query, err := NewQueryParser(`subgraph node.team = "platform" | collapse into PLATFORM where exists | make node.critical = true where exists`).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	graphValue := result.(GraphValue)
	graph := graphValue.Graph

	// After subgraph filters to platform nodes, we only have api, db, cache
	// After collapse, they become PLATFORM
	// So should have only 1 node
	if len(graph.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(graph.Nodes))
	}

	// Should have collapsed node
	if _, exists := graph.Nodes["PLATFORM"]; !exists {
		t.Errorf("collapsed node PLATFORM not found")
	}

	// Should have critical attribute
	node := graph.Nodes["PLATFORM"]
	if critical, ok := node.Attributes["critical"]; !ok || critical != true {
		t.Errorf("critical attribute not set on PLATFORM")
	}
}

func TestIntegrationMultipleBindings(t *testing.T) {
	// Test: bind services, bind databases, union them
	input := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"api": {ID: "api", Attributes: map[string]interface{}{"type": "service"}, Sets: make(map[string]struct{})},
			"web": {ID: "web", Attributes: map[string]interface{}{"type": "service"}, Sets: make(map[string]struct{})},
			"db": {ID: "db", Attributes: map[string]interface{}{"type": "database"}, Sets: make(map[string]struct{})},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Pipeline: (subgraph node.type = "service") as SERVICES | (from * | subgraph node.type = "database") as DATABASES | SERVICES + DATABASES
	query, err := NewQueryParser(`(subgraph node.type = "service") as SERVICES | (from * | subgraph node.type = "database") as DATABASES | SERVICES + DATABASES`).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	graphValue := result.(GraphValue)
	graph := graphValue.Graph

	// Union should include all 3 nodes
	if len(graph.Nodes) != 3 {
		t.Errorf("expected 3 nodes in union, got %d", len(graph.Nodes))
	}

	// Check named graphs were bound
	if _, exists := ctx.NamedGraphs["SERVICES"]; !exists {
		t.Errorf("SERVICES binding not found")
	}
	if _, exists := ctx.NamedGraphs["DATABASES"]; !exists {
		t.Errorf("DATABASES binding not found")
	}
}

func TestIntegrationTraversalAndTransform(t *testing.T) {
	// Test: find critical nodes and their neighbors, mark as related
	input := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"critical": {ID: "critical", Attributes: map[string]interface{}{"critical": true}, Sets: make(map[string]struct{})},
			"neighbor1": {ID: "neighbor1", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
			"neighbor2": {ID: "neighbor2", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
			"isolated": {ID: "isolated", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
		},
		Edges: []*gsl.Edge{
			{From: "critical", To: "neighbor1", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
			{From: "neighbor1", To: "neighbor2", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
		},
		Sets: make(map[string]*gsl.Set),
	}

	// Pipeline: subgraph node.critical = true traverse out 2 | make node.related = true where exists
	query, err := NewQueryParser(`subgraph node.critical = true traverse out 2 | make node.related = true where exists`).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	graphValue := result.(GraphValue)
	graph := graphValue.Graph

	// Should have critical, neighbor1, neighbor2 (traversed 2 hops out)
	if len(graph.Nodes) != 3 {
		t.Errorf("expected 3 nodes after traversal, got %d", len(graph.Nodes))
	}

	// All should have related = true
	for id, node := range graph.Nodes {
		if related, ok := node.Attributes["related"]; !ok || related != true {
			t.Errorf("node %s missing related attribute", id)
		}
	}
}

func TestIntegrationRemoveAndFilter(t *testing.T) {
	// Test: filter active nodes and clean orphans
	input := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"active1": {ID: "active1", Attributes: map[string]interface{}{"status": "active"}, Sets: make(map[string]struct{})},
			"active2": {ID: "active2", Attributes: map[string]interface{}{"status": "active"}, Sets: make(map[string]struct{})},
			"deprecated": {ID: "deprecated", Attributes: map[string]interface{}{"status": "deprecated"}, Sets: make(map[string]struct{})},
			"orphan": {ID: "orphan", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
		},
		Edges: []*gsl.Edge{
			{From: "active1", To: "active2", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
		},
		Sets: make(map[string]*gsl.Set),
	}

	// Pipeline: subgraph node.status = "active" | remove orphans
	// This filters to active nodes only (active1, active2 with edge between them)
	// Then remove orphans (should remove nothing since both have edges)
	query, err := NewQueryParser(`subgraph node.status = "active" | remove orphans`).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	graphValue := result.(GraphValue)
	graph := graphValue.Graph

	// Should have both active nodes
	if len(graph.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(graph.Nodes))
	}

	if _, exists := graph.Nodes["active1"]; !exists {
		t.Errorf("active1 node not found")
	}
	if _, exists := graph.Nodes["active2"]; !exists {
		t.Errorf("active2 node not found")
	}
}

func TestIntegrationEmptyGraph(t *testing.T) {
	// Test: all operations on empty graph
	input := &gsl.Graph{
		Nodes: make(map[string]*gsl.Node),
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	tests := []struct {
		name  string
		query string
	}{
		{name: "subgraph on empty", query: "subgraph exists"},
		{name: "make on empty", query: "make node.x = 1 where exists"},
		{name: "remove on empty", query: "remove orphans"},
		{name: "collapse on empty", query: "collapse into X where exists"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := NewQueryParser(tt.query).Parse()
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			ctx := &QueryContext{
				InputGraph:  input,
				NamedGraphs: make(map[string]*gsl.Graph),
			}

			result, err := q.Execute(ctx)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			graphValue := result.(GraphValue)
			if graphValue.Graph == nil {
				t.Errorf("returned nil graph")
			}
		})
	}
}

func TestIntegrationSingleNode(t *testing.T) {
	// Test: operations on single-node graph
	input := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Pipeline: make node.label = "single" where exists | collapse into MERGED where exists
	query, err := NewQueryParser(`make node.label = "single" where exists | collapse into MERGED where exists`).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	graphValue := result.(GraphValue)
	graph := graphValue.Graph

	// Should have merged node with label
	if _, exists := graph.Nodes["MERGED"]; !exists {
		t.Errorf("MERGED node not found")
	}

	if label, ok := graph.Nodes["MERGED"].Attributes["label"]; !ok || label != "single" {
		t.Errorf("label attribute not preserved")
	}
}

func TestIntegrationCyclePreservation(t *testing.T) {
	// Test: cyclic graphs are preserved through operations
	input := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
			"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
			{From: "B", To: "C", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
			{From: "C", To: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
		},
		Sets: make(map[string]*gsl.Set),
	}

	// Pipeline: make node.visited = true where exists | subgraph exists
	query, err := NewQueryParser(`make node.visited = true where exists | subgraph exists`).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	graphValue := result.(GraphValue)
	graph := graphValue.Graph

	// All 3 nodes and 3 edges preserved
	if len(graph.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(graph.Nodes))
	}

	if len(graph.Edges) != 3 {
		t.Errorf("expected 3 edges, got %d", len(graph.Edges))
	}

	// All nodes marked as visited
	for _, node := range graph.Nodes {
		if visited, ok := node.Attributes["visited"]; !ok || visited != true {
			t.Errorf("node %s not marked visited", node.ID)
		}
	}
}

func TestIntegrationSetPreservation(t *testing.T) {
	// Test: sets are preserved through complex operations
	input := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"critical": {}, "monitored": {}}},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: map[string]struct{}{"monitored": {}}},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
		},
		Sets: map[string]*gsl.Set{
			"critical": {ID: "critical", Attributes: map[string]interface{}{}},
			"monitored": {ID: "monitored", Attributes: map[string]interface{}{}},
		},
	}

	// Pipeline: subgraph in critical | collapse into CRITICAL where exists
	query, err := NewQueryParser(`subgraph in critical | collapse into CRITICAL where exists`).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	graphValue := result.(GraphValue)
	graph := graphValue.Graph

	// Sets should be preserved
	if len(graph.Sets) != 2 {
		t.Errorf("expected 2 sets, got %d", len(graph.Sets))
	}

	if _, exists := graph.Sets["critical"]; !exists {
		t.Errorf("critical set not found")
	}

	if _, exists := graph.Sets["monitored"]; !exists {
		t.Errorf("monitored set not found")
	}
}

func TestIntegrationErrorPropagation(t *testing.T) {
	// Test: errors in pipeline stop execution
	input := &gsl.Graph{
		Nodes: map[string]*gsl.Node{},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{name: "missing graph reference", query: "G_MISSING + G1", wantErr: true},
		{name: "rebind attempt", query: "(from *) as X | (from *) as X", wantErr: true},
		{name: "invalid graph reference", query: "invalid-name + G1", wantErr: true},
		{name: "type mismatch in predicate", query: "subgraph node.x = true AND edge.y = 1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := NewQueryParser(tt.query).Parse()
			if err != nil && tt.wantErr {
				// Parse error is acceptable
				return
			}
			if err != nil {
				t.Fatalf("Parse() unexpected error = %v", err)
			}

			ctx := &QueryContext{
				InputGraph:  input,
				NamedGraphs: make(map[string]*gsl.Graph),
			}

			_, err = q.Execute(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIntegrationDuplicateEdges(t *testing.T) {
	// Test: duplicate edges are preserved through operations
	input := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
			"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
		},
		Edges: []*gsl.Edge{
			{From: "A", To: "B", Attributes: map[string]interface{}{"type": "sync"}, Sets: make(map[string]struct{})},
			{From: "A", To: "B", Attributes: map[string]interface{}{"type": "async"}, Sets: make(map[string]struct{})},
		},
		Sets: make(map[string]*gsl.Set),
	}

	// Pipeline: make node.marked = true where exists
	query, err := NewQueryParser(`make node.marked = true where exists`).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	graphValue := result.(GraphValue)
	graph := graphValue.Graph

	// Duplicate edges should be preserved
	if len(graph.Edges) != 2 {
		t.Errorf("expected 2 edges (duplicates), got %d", len(graph.Edges))
	}
}

func TestIntegrationAttributePropagation(t *testing.T) {
	// Test: attributes propagate correctly through transformations
	input := &gsl.Graph{
		Nodes: map[string]*gsl.Node{
			"A": {ID: "A", Attributes: map[string]interface{}{"owner": "alice", "env": "prod"}, Sets: make(map[string]struct{})},
			"B": {ID: "B", Attributes: map[string]interface{}{"owner": "bob", "env": "staging"}, Sets: make(map[string]struct{})},
		},
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Pipeline: subgraph node.owner = "alice" | make node.env = "prod" where exists
	query, err := NewQueryParser(`subgraph node.owner = "alice" | make node.env = "prod" where exists`).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	graphValue := result.(GraphValue)
	graph := graphValue.Graph

	// Should have A with owner=alice and env=prod
	nodeA := graph.Nodes["A"]
	if nodeA == nil {
		t.Fatalf("node A not found")
	}

	owner, _ := nodeA.Attributes["owner"]
	env, _ := nodeA.Attributes["env"]

	if owner != "alice" || env != "prod" {
		t.Errorf("attributes not propagated correctly: owner=%v, env=%v", owner, env)
	}
}

func TestIntegrationLargeGraph(t *testing.T) {
	// Test: operations on larger graph
	input := &gsl.Graph{
		Nodes: make(map[string]*gsl.Node),
		Edges: []*gsl.Edge{},
		Sets:  make(map[string]*gsl.Set),
	}

	// Create 100 nodes
	for i := 0; i < 100; i++ {
		id := "node_" + string(rune('0' + i % 10)) + string(rune('A' + i / 10))
		isService := i % 2 == 0
		input.Nodes[id] = &gsl.Node{
			ID: id,
			Attributes: map[string]interface{}{
				"type": map[bool]string{true: "service", false: "database"}[isService],
				"index": i,
			},
			Sets: make(map[string]struct{}),
		}
	}

	// Create edges
	for i := 0; i < 90; i++ {
		fromId := "node_" + string(rune('0' + i % 10)) + string(rune('A' + i / 10))
		toId := "node_" + string(rune('0' + (i+1) % 10)) + string(rune('A' + (i + 1) / 10))
		input.Edges = append(input.Edges, &gsl.Edge{
			From:       fromId,
			To:         toId,
			Attributes: map[string]interface{}{},
			Sets:       make(map[string]struct{}),
		})
	}

	// Pipeline: subgraph node.type = "service" | make node.priority = "high" where exists
	query, err := NewQueryParser(`subgraph node.type = "service" | make node.priority = "high" where exists`).Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := &QueryContext{
		InputGraph:  input,
		NamedGraphs: make(map[string]*gsl.Graph),
	}

	result, err := query.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	graphValue := result.(GraphValue)
	graph := graphValue.Graph

	// Should have 50 service nodes
	if len(graph.Nodes) != 50 {
		t.Errorf("expected 50 service nodes, got %d", len(graph.Nodes))
	}

	// All should have priority=high
	for _, node := range graph.Nodes {
		if priority, ok := node.Attributes["priority"]; !ok || priority != "high" {
			t.Errorf("node %s missing priority attribute", node.ID)
		}
	}
}
