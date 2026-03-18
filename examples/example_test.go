package gsl_test

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	gsl "github.com/dnnrly/gsl-lang"
)

// Example_parseSimpleWorkflow demonstrates parsing a basic workflow graph.
func Example_parseSimpleWorkflow() {
	content, _ := os.ReadFile("simple_workflow.gsl")
	graph, _, _ := gsl.Parse(bytes.NewReader(content))

	nodes := graph.GetNodes()
	edges := graph.GetEdges()
	fmt.Printf("Workflow has %d nodes and %d edges\n", len(nodes), len(edges))

	// Print nodes in sorted order for consistent output
	nodeIDs := make([]string, 0, len(nodes))
	for id := range nodes {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Strings(nodeIDs)

	fmt.Println("Nodes:")
	for _, id := range nodeIDs {
		node := nodes[id]
		if text, ok := node.Attributes["text"]; ok {
			fmt.Printf("  %s: %v\n", id, text)
		}
	}

	// Output:
	// Workflow has 6 nodes and 5 edges
	// Nodes:
	//   Decision: Check Result
	//   Failure: Error Handling
	//   ProcessA: Validate Input
	//   ProcessB: Transform Data
	//   Start: Begin Process
	//   Success: Output Generated
}

// Example_parseHierarchicalSystem demonstrates parsing a graph with parent-child relationships.
func Example_parseHierarchicalSystem() {
	content, _ := os.ReadFile("hierarchical_system.gsl")
	graph, _, _ := gsl.Parse(bytes.NewReader(content))

	nodes := graph.GetNodes()
	sets := graph.GetSets()
	fmt.Printf("System has %d nodes and %d sets\n", len(nodes), len(sets))

	// Print sets with their attributes
	setNames := make([]string, 0, len(sets))
	for name := range sets {
		setNames = append(setNames, name)
	}
	sort.Strings(setNames)

	fmt.Println("Sets:")
	for _, name := range setNames {
		set := sets[name]
		fmt.Printf("  %s: %v\n", name, set.Attributes)
	}

	// Output:
	// System has 9 nodes and 3 sets
	// Sets:
	//   backend: map[color:green]
	//   database: map[color:red]
	//   frontend: map[color:blue]
}

// Example_microservicesArchitecture demonstrates analyzing a microservices graph.
func Example_microservicesArchitecture() {
	content, _ := os.ReadFile("microservices.gsl")
	graph, _, _ := gsl.Parse(bytes.NewReader(content))

	nodes := graph.GetNodes()
	fmt.Printf("Microservices: %d services\n", len(nodes))

	// Find critical services
	critical := make([]string, 0)
	for nodeID, node := range nodes {
		if _, ok := node.Sets["critical"]; ok {
			critical = append(critical, nodeID)
		}
	}
	sort.Strings(critical)

	fmt.Println("Critical services:")
	for _, id := range critical {
		node := nodes[id]
		if text, ok := node.Attributes["text"]; ok {
			fmt.Printf("  %s (%v)\n", id, text)
		}
	}

	// Output:
	// Microservices: 7 services
	// Critical services:
	//   Database (PostgreSQL)
	//   OrderService (Order Service)
	//   UserService (User Service)
}

// Example_queryNodeDependencies demonstrates querying outbound edges from a node.
func Example_queryNodeDependencies() {
	content, _ := os.ReadFile("microservices.gsl")
	graph, _, _ := gsl.Parse(bytes.NewReader(content))

	nodes := graph.GetNodes()
	edges := graph.GetEdges()
	// Find all services that OrderService depends on
	deps := make([]string, 0)
	for _, edge := range edges {
		if edge.From == "OrderService" {
			deps = append(deps, edge.To)
		}
	}
	sort.Strings(deps)

	fmt.Printf("OrderService depends on %d services:\n", len(deps))
	for _, target := range deps {
		if node, ok := nodes[target]; ok {
			if text, ok := node.Attributes["text"]; ok {
				fmt.Printf("  %s (%v)\n", target, text)
			}
		}
	}

	// Output:
	// OrderService depends on 3 services:
	//   Cache (Redis Cache)
	//   Database (PostgreSQL)
	//   PaymentService (Payment Service)
}

// Example_parseDataPipeline demonstrates parsing an ETL pipeline graph.
func Example_parseDataPipeline() {
	content, _ := os.ReadFile("data_pipeline.gsl")
	graph, _, _ := gsl.Parse(bytes.NewReader(content))

	nodes := graph.GetNodes()
	edges := graph.GetEdges()
	sets := graph.GetSets()
	fmt.Printf("Pipeline has %d stages with %d transformations\n", len(sets), len(edges))

	// Group nodes by stage
	stages := make(map[string][]string)
	for nodeID, node := range nodes {
		for stageName := range node.Sets {
			stages[stageName] = append(stages[stageName], nodeID)
		}
	}

	stageNames := make([]string, 0, len(stages))
	for name := range stages {
		stageNames = append(stageNames, name)
	}
	sort.Strings(stageNames)

	fmt.Println("Pipeline stages:")
	for _, stageName := range stageNames {
		nodes := stages[stageName]
		sort.Strings(nodes)
		fmt.Printf("  %s: %d nodes\n", stageName, len(nodes))
	}

	// Output:
	// Pipeline has 3 stages with 7 transformations
	// Pipeline stages:
	//   intake: 2 nodes
	//   output: 3 nodes
	//   processing: 3 nodes
}

// Example_serializeGraph demonstrates round-tripping: parsing and re-serializing.
func Example_serializeGraph() {
	// Parse a simple graph
	input := `node A: "Start"
node B: "End"
A->B`

	graph, _, _ := gsl.Parse(bytes.NewReader([]byte(input)))
	serialized := gsl.Serialize(graph)

	// Re-parse the serialized form
	graph2, _, _ := gsl.Parse(bytes.NewReader([]byte(serialized)))

	fmt.Printf("Round-trip: %d nodes → %d nodes\n", len(graph.GetNodes()), len(graph2.GetNodes()))
	fmt.Println("Serialized:")
	fmt.Println(serialized)

	// Output:
	// Round-trip: 2 nodes → 2 nodes
	// Serialized:
	// node A [text="Start"]
	// node B [text="End"]
	//
	// A->B
}

// Example_graphStatistics demonstrates basic graph statistics.
func Example_graphStatistics() {
	content, _ := os.ReadFile("microservices.gsl")
	graph, _, _ := gsl.Parse(bytes.NewReader(content))

	nodes := graph.GetNodes()
	// Count attributes
	nodeAttrs := make(map[string]int)
	for _, node := range nodes {
		for key := range node.Attributes {
			nodeAttrs[key]++
		}
	}

	attrs := make([]string, 0, len(nodeAttrs))
	for attr := range nodeAttrs {
		attrs = append(attrs, attr)
	}
	sort.Strings(attrs)

	fmt.Println("Node attribute usage:")
	for _, attr := range attrs {
		fmt.Printf("  %s: %d nodes\n", attr, nodeAttrs[attr])
	}

	// Output:
	// Node attribute usage:
	//   port: 5 nodes
	//   replicas: 3 nodes
	//   text: 7 nodes
	//   ttl: 1 nodes
	//   version: 1 nodes
}

// Example_implicitSets demonstrates parsing with implicit set creation warnings.
func Example_implicitSets() {
	content, _ := os.ReadFile("implicit_sets.gsl")
	graph, warnings, _ := gsl.Parse(bytes.NewReader(content))

	fmt.Printf("Graph: %d sets created\n", len(graph.GetSets()))
	fmt.Printf("Warnings: %d\n", len(warnings))

	// Print warnings
	for _, w := range warnings {
		fmt.Printf("  - %v\n", w)
	}

	// Output:
	// Graph: 2 sets created
	// Warnings: 2
	//   - implicit set creation: "active"
	//   - implicit set creation: "deprecated"
}

// Example_nameCollision demonstrates parsing with name collision warnings.
func Example_nameCollision() {
	content, _ := os.ReadFile("name_collision.gsl")
	graph, warnings, _ := gsl.Parse(bytes.NewReader(content))

	fmt.Printf("Graph: %d nodes, %d sets\n", len(graph.GetNodes()), len(graph.GetSets()))
	fmt.Printf("Warnings: %d\n", len(warnings))

	// Print warnings
	for _, w := range warnings {
		fmt.Printf("  - %v\n", w)
	}

	// Output:
	// Graph: 3 nodes, 1 sets
	// Warnings: 1
	//   - node and set name collision: "auth"
}

// Example_parentOverride demonstrates parsing with parent override warnings.
func Example_parentOverride() {
	content, _ := os.ReadFile("parent_override.gsl")
	graph, warnings, _ := gsl.Parse(bytes.NewReader(content))

	nodes := graph.GetNodes()
	fmt.Printf("Graph: %d nodes\n", len(nodes))
	fmt.Printf("Warnings: %d\n", len(warnings))

	// Print warnings
	for _, w := range warnings {
		fmt.Printf("  - %v\n", w)
	}

	// Show parent relationships (sorted for deterministic output)
	fmt.Println("Parent relationships:")
	nodeIDs := make([]string, 0)
	for nodeID := range nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Strings(nodeIDs)
	for _, nodeID := range nodeIDs {
		node := nodes[nodeID]
		if parent, ok := node.Attributes["parent"]; ok {
			fmt.Printf("  %s -> %v\n", nodeID, parent)
		}
	}

	// Output:
	// Graph: 4 nodes
	// Warnings: 1
	//   - 7:3: parent override inside block
	// Parent relationships:
	//   Child1 -> Parent1
	//   Child2 -> Parent2
}

// Example_topologicalSort demonstrates topological sorting on a task dependency graph.
func Example_topologicalSort() {
	content, _ := os.ReadFile("task_scheduling.gsl")
	graph, _, _ := gsl.Parse(bytes.NewReader(content))

	nodes := graph.GetNodes()
	edges := graph.GetEdges()
	// Build adjacency list for in-degree calculation
	inDegree := make(map[string]int)
	outEdges := make(map[string][]string)

	// Initialize all nodes
	for nodeID := range nodes {
		inDegree[nodeID] = 0
		outEdges[nodeID] = []string{}
	}

	// Count in-degrees and build adjacency list
	for _, edge := range edges {
		inDegree[edge.To]++
		outEdges[edge.From] = append(outEdges[edge.From], edge.To)
	}

	// Kahn's algorithm for topological sort
	queue := []string{}
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	sort.Strings(queue) // Deterministic ordering

	sortedOrder := []string{}
	for len(queue) > 0 {
		// Pop from front
		current := queue[0]
		queue = queue[1:]
		sortedOrder = append(sortedOrder, current)

		// Process neighbors
		neighbors := outEdges[current]
		sort.Strings(neighbors)
		for _, neighbor := range neighbors {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Print execution order
	fmt.Println("Task execution order:")
	for i, nodeID := range sortedOrder {
		node := nodes[nodeID]
		text := ""
		if t, ok := node.Attributes["text"]; ok {
			text = fmt.Sprintf(" (%v)", t)
		}
		fmt.Printf("  %d. %s%s\n", i+1, nodeID, text)
	}

	// Output:
	// Task execution order:
	//   1. Setup (Setup Environment)
	//   2. Build (Build Binary)
	//   3. UnitTests (Run Unit Tests)
	//   4. Package (Create Package)
	//   5. Integration (Integration Tests)
	//   6. Deploy (Deploy to Prod)
	//   7. Smoke (Smoke Tests)
}

// Example_cycleDetection demonstrates detecting cycles in a dependency graph using DFS.
func Example_cycleDetection() {
	content, _ := os.ReadFile("circular_dependencies.gsl")
	graph, _, _ := gsl.Parse(bytes.NewReader(content))

	nodes := graph.GetNodes()
	edges := graph.GetEdges()
	// Build adjacency list
	adjList := make(map[string][]string)
	for nodeID := range nodes {
		adjList[nodeID] = []string{}
	}
	for _, edge := range edges {
		adjList[edge.From] = append(adjList[edge.From], edge.To)
	}

	// DFS-based cycle detection
	WHITE, GRAY, BLACK := 0, 1, 2
	color := make(map[string]int)
	for nodeID := range nodes {
		color[nodeID] = WHITE
	}

	hasCycle := false
	var cycleNodes []string

	var dfs func(string, []string) bool
	dfs = func(node string, path []string) bool {
		color[node] = GRAY
		path = append(path, node)

		neighbors := adjList[node]
		sort.Strings(neighbors)

		for _, neighbor := range neighbors {
			if color[neighbor] == GRAY {
				// Found a back edge - cycle detected
				cycleNodes = path
				cycleNodes = append(cycleNodes, neighbor)
				return true
			}
			if color[neighbor] == WHITE {
				if dfs(neighbor, path) {
					return true
				}
			}
		}

		color[node] = BLACK
		return false
	}

	// Check each unvisited node
	nodeIDs := make([]string, 0, len(nodes))
	for nodeID := range nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Strings(nodeIDs)

	for _, nodeID := range nodeIDs {
		if color[nodeID] == WHITE {
			if dfs(nodeID, []string{}) {
				hasCycle = true
				break
			}
		}
	}

	// Report results
	if hasCycle {
		fmt.Println("Cycle detected in graph!")
		fmt.Printf("Cycle: %s -> %s\n", strings.Join(cycleNodes, " -> "), cycleNodes[0])
	} else {
		fmt.Println("No cycles found")
	}

	// Output:
	// Cycle detected in graph!
	// Cycle: APIGateway -> AuthService -> ConfigService -> LogService -> AuthService -> APIGateway
}

// Example_pathFinding demonstrates finding all paths between two nodes using DFS.
func Example_pathFinding() {
	content, _ := os.ReadFile("social_network.gsl")
	graph, _, _ := gsl.Parse(bytes.NewReader(content))

	nodes := graph.GetNodes()
	edges := graph.GetEdges()
	// Build adjacency list
	adjList := make(map[string][]string)
	for nodeID := range nodes {
		adjList[nodeID] = []string{}
	}
	for _, edge := range edges {
		adjList[edge.From] = append(adjList[edge.From], edge.To)
	}

	// DFS to find all paths
	var allPaths [][]string

	var dfs func(string, string, []string, map[string]bool)
	dfs = func(current, target string, path []string, visited map[string]bool) {
		if current == target {
			// Make a copy of path and add to results
			pathCopy := make([]string, len(path))
			copy(pathCopy, path)
			allPaths = append(allPaths, pathCopy)
			return
		}

		neighbors := adjList[current]
		sort.Strings(neighbors)

		for _, neighbor := range neighbors {
			if !visited[neighbor] {
				visited[neighbor] = true
				dfs(neighbor, target, append(path, neighbor), visited)
				visited[neighbor] = false
			}
		}
	}

	// Find paths from Alice to Frank
	start, end := "Alice", "Frank"
	visited := make(map[string]bool)
	visited[start] = true
	dfs(start, end, []string{start}, visited)

	// Print results
	fmt.Printf("All paths from %s to %s:\n", start, end)
	if len(allPaths) == 0 {
		fmt.Println("  (no paths found)")
	} else {
		for i, path := range allPaths {
			fmt.Printf("  %d. %s\n", i+1, strings.Join(path, " -> "))
		}
	}

	// Output:
	// All paths from Alice to Frank:
	//   1. Alice -> Bob -> Diana -> Eve -> Frank
	//   2. Alice -> Bob -> Frank
	//   3. Alice -> Charlie -> Diana -> Eve -> Frank
}
