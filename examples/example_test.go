package gsl_test

import (
	"bytes"
	"fmt"
	"os"
	"sort"

	gsl "github.com/dnnrly/gsl-lang"
)

// Example_parseSimpleWorkflow demonstrates parsing a basic workflow graph.
func Example_parseSimpleWorkflow() {
	content, _ := os.ReadFile("simple_workflow.gsl")
	graph, _, _ := gsl.Parse(bytes.NewReader(content))

	fmt.Printf("Workflow has %d nodes and %d edges\n", len(graph.Nodes), len(graph.Edges))

	// Print nodes in sorted order for consistent output
	nodeIDs := make([]string, 0, len(graph.Nodes))
	for id := range graph.Nodes {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Strings(nodeIDs)

	fmt.Println("Nodes:")
	for _, id := range nodeIDs {
		node := graph.Nodes[id]
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

	fmt.Printf("System has %d nodes and %d sets\n", len(graph.Nodes), len(graph.Sets))

	// Print sets with their attributes
	setNames := make([]string, 0, len(graph.Sets))
	for name := range graph.Sets {
		setNames = append(setNames, name)
	}
	sort.Strings(setNames)

	fmt.Println("Sets:")
	for _, name := range setNames {
		set := graph.Sets[name]
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

	fmt.Printf("Microservices: %d services\n", len(graph.Nodes))

	// Find critical services
	critical := make([]string, 0)
	for nodeID, node := range graph.Nodes {
		if _, ok := node.Sets["critical"]; ok {
			critical = append(critical, nodeID)
		}
	}
	sort.Strings(critical)

	fmt.Println("Critical services:")
	for _, id := range critical {
		node := graph.Nodes[id]
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

	// Find all services that OrderService depends on
	deps := make([]string, 0)
	for _, edge := range graph.Edges {
		if edge.From == "OrderService" {
			deps = append(deps, edge.To)
		}
	}
	sort.Strings(deps)

	fmt.Printf("OrderService depends on %d services:\n", len(deps))
	for _, target := range deps {
		if node, ok := graph.Nodes[target]; ok {
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

	fmt.Printf("Pipeline has %d stages with %d transformations\n", len(graph.Sets), len(graph.Edges))

	// Group nodes by stage
	stages := make(map[string][]string)
	for nodeID, node := range graph.Nodes {
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

	fmt.Printf("Round-trip: %d nodes → %d nodes\n", len(graph.Nodes), len(graph2.Nodes))
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

	// Count attributes
	nodeAttrs := make(map[string]int)
	for _, node := range graph.Nodes {
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

	fmt.Printf("Graph: %d sets created\n", len(graph.Sets))
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

	fmt.Printf("Graph: %d nodes, %d sets\n", len(graph.Nodes), len(graph.Sets))
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

	fmt.Printf("Graph: %d nodes\n", len(graph.Nodes))
	fmt.Printf("Warnings: %d\n", len(warnings))

	// Print warnings
	for _, w := range warnings {
		fmt.Printf("  - %v\n", w)
	}

	// Show parent relationships
	fmt.Println("Parent relationships:")
	for nodeID, node := range graph.Nodes {
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
