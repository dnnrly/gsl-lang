---
name: gsl-go-library
description: Complete reference for using GSL (Graph Specification Language) in Go applications. Covers API usage, data structures, algorithms, and best practices. Use when working with GSL in Go code.
---

# GSL Go Library - Complete Guide for LLMs

This document contains everything needed to use GSL in Go applications.

**First, read [LLM_GUIDE.md](LLM_GUIDE.md) to understand the GSL language itself.**

## Go API Reference

### Import and Basic Setup

```go
import (
	"bytes"
	"os"
	gsl "github.com/dnnrly/gsl-lang"
)

// Parse from file
content, _ := os.ReadFile("graph.gsl")
graph, warnings, err := gsl.Parse(bytes.NewReader(content))
if err != nil {
	// Fatal parse error
}

// Process warnings (non-fatal)
for _, w := range warnings {
	fmt.Printf("Warning: %v\n", w)
}
```

### Data Structures

```go
type Graph struct {
	Nodes map[string]*Node  // ID -> Node
	Edges []*Edge           // All edges (multiset, allows duplicates)
	Sets  map[string]*Set   // ID -> Set
}

type Node struct {
	ID         string                 // Node identifier
	Attributes map[string]interface{} // Attributes (untyped)
	Sets       map[string]struct{}    // Set membership (keys only)
	Parent     *string                // Cached parent reference
}

type Edge struct {
	From       string                 // Source node ID
	To         string                 // Target node ID
	Attributes map[string]interface{} // Attributes (untyped)
	Sets       map[string]struct{}    // Set membership (keys only)
}

type Set struct {
	ID         string                 // Set identifier
	Attributes map[string]interface{} // Attributes (untyped)
}
```

### Common Operations

#### Access Nodes

```go
// Iterate all nodes
for nodeID, node := range graph.Nodes {
	fmt.Println(nodeID)
	
	// Get text attribute
	if text, ok := node.Attributes["text"]; ok {
		fmt.Println(text.(string))
	}
	
	// Check parent
	if parent, ok := node.Attributes["parent"]; ok {
		fmt.Println(parent.(string))
	}
}
```

#### Traverse Edges

```go
// Find outbound edges from a node
for _, edge := range graph.Edges {
	if edge.From == "API" {
		fmt.Printf("%s -> %s\n", edge.From, edge.To)
		
		if method, ok := edge.Attributes["method"]; ok {
			fmt.Println(method)
		}
	}
}

// Find inbound edges to a node
for _, edge := range graph.Edges {
	if edge.To == "Database" {
		fmt.Printf("%s -> %s\n", edge.From, edge.To)
	}
}
```

#### Query Sets

```go
// Find nodes in a set
for nodeID, node := range graph.Nodes {
	if _, isCritical := node.Sets["critical"]; isCritical {
		fmt.Println(nodeID)
	}
}

// Find edges in a set
for _, edge := range graph.Edges {
	if _, isProd := edge.Sets["production"]; isProd {
		fmt.Printf("%s -> %s\n", edge.From, edge.To)
	}
}

// Check all sets a node belongs to
for setName := range node.Sets {
	fmt.Println(setName)
}
```

#### Serialize

```go
canonical := gsl.Serialize(graph)
fmt.Println(canonical)

// Re-parse to verify round-trip
graph2, _, _ := gsl.Parse(bytes.NewReader([]byte(canonical)))
// graph and graph2 are semantically equivalent
```

#### Clone (Deep Copy)

```go
// Create an independent copy of the graph
original := gsl.NewGraph()
original.AddNode("A", nil)
original.AddNode("B", nil)
original.AddEdge("A", "B", nil)

cloned := original.Clone()

// Mutations to the clone do NOT affect the original
cloned.AddNode("C", nil)

len(original.GetNodes()) // 2
len(cloned.GetNodes())   // 3

// All attributes and set memberships are preserved
clonedNodeA := cloned.GetNode("A")
// clonedNodeA.Attributes is a deep copy of original's attributes
```

**Use cases:**
- Safe mutations for query transformations (e.g., remove nodes, add edges)
- Experimental graph modifications without side effects
- Creating multiple variations of a base graph
- Testing graph operations in isolation

## Algorithm Patterns

### Topological Sort (Kahn's Algorithm)

```go
inDegree := make(map[string]int)
outEdges := make(map[string][]string)

for nodeID := range graph.Nodes {
	inDegree[nodeID] = 0
	outEdges[nodeID] = []string{}
}

for _, edge := range graph.Edges {
	inDegree[edge.To]++
	outEdges[edge.From] = append(outEdges[edge.From], edge.To)
}

queue := []string{}
for nodeID, degree := range inDegree {
	if degree == 0 {
		queue = append(queue, nodeID)
	}
}

sorted := []string{}
for len(queue) > 0 {
	current := queue[0]
	queue = queue[1:]
	sorted = append(sorted, current)
	
	for _, neighbor := range outEdges[current] {
		inDegree[neighbor]--
		if inDegree[neighbor] == 0 {
			queue = append(queue, neighbor)
		}
	}
}
```

### Cycle Detection (DFS)

```go
WHITE, GRAY, BLACK := 0, 1, 2
color := make(map[string]int)
for nodeID := range graph.Nodes {
	color[nodeID] = WHITE
}

adjList := make(map[string][]string)
for nodeID := range graph.Nodes {
	adjList[nodeID] = []string{}
}
for _, edge := range graph.Edges {
	adjList[edge.From] = append(adjList[edge.From], edge.To)
}

var dfs func(string) bool
dfs = func(node string) bool {
	color[node] = GRAY
	for _, neighbor := range adjList[node] {
		if color[neighbor] == GRAY {
			return true  // Cycle found
		}
		if color[neighbor] == WHITE && dfs(neighbor) {
			return true
		}
	}
	color[node] = BLACK
	return false
}

hasCycle := false
for nodeID := range graph.Nodes {
	if color[nodeID] == WHITE && dfs(nodeID) {
		hasCycle = true
		break
	}
}
```

### Path Finding (DFS)

```go
adjList := make(map[string][]string)
for nodeID := range graph.Nodes {
	adjList[nodeID] = []string{}
}
for _, edge := range graph.Edges {
	adjList[edge.From] = append(adjList[edge.From], edge.To)
}

var allPaths [][]string

var dfs func(string, string, []string, map[string]bool)
dfs = func(current, target string, path []string, visited map[string]bool) {
	if current == target {
		pathCopy := make([]string, len(path))
		copy(pathCopy, path)
		allPaths = append(allPaths, pathCopy)
		return
	}
	
	for _, neighbor := range adjList[current] {
		if !visited[neighbor] {
			visited[neighbor] = true
			dfs(neighbor, target, append(path, neighbor), visited)
			visited[neighbor] = false
		}
	}
}

visited := make(map[string]bool)
visited["A"] = true
dfs("A", "Z", []string{"A"}, visited)

// allPaths now contains all paths from A to Z
```

## Go-Specific Usage Patterns

### Parse Correctly

- **Three return values**: `Parse()` returns `(*Graph, []error, error)` where:
  - First return: The parsed graph (even if warnings exist)
  - Second return: Non-fatal warnings (check these!)
  - Third return: Fatal parse error
- **Always check both returns**: Warnings may indicate issues (implicit sets, name collisions)

### Work with Untyped Attributes

Attributes in nodes, edges, and sets are stored as `interface{}`. Always type-assert:

```go
// DON'T: count := node.Attributes["count"] + 1  // Error!
// DO:
if val, ok := node.Attributes["count"]; ok {
	count := val.(int) + 1
}
```

Common types:
- String: `"text"` → `string`
- Integer: `42` → `int` (stores as `float64`, convert if needed)
- Float: `3.14` → `float64`
- Boolean: `true`/`false` → `bool`
- NodeRef: `OtherNode` → `string`
- Empty: `flag` → `bool` (true if present)

### Build Adjacency Lists for Algorithms

Before running graph algorithms, build an adjacency list once:

```go
adjList := make(map[string][]string)
for nodeID := range graph.Nodes {
	adjList[nodeID] = []string{}
}
for _, edge := range graph.Edges {
	adjList[edge.From] = append(adjList[edge.From], edge.To)
}
// Now use adjList for DFS, BFS, etc.
```

### Handle Duplicate Edges

`graph.Edges` is a multiset—the same edge can appear multiple times:

```go
// This is valid and both edges are stored:
// A -> B [weight=1]
// A -> B [weight=2]

// If you need unique edges:
seen := make(map[string]bool)
for _, edge := range graph.Edges {
	key := edge.From + "->" + edge.To
	if !seen[key] {
		// Process unique edge
		seen[key] = true
	}
}
```

### Sort for Determinism

Maps are unordered in Go. For deterministic output, always sort:

```go
nodeIDs := make([]string, 0, len(graph.Nodes))
for id := range graph.Nodes {
	nodeIDs = append(nodeIDs, id)
}
sort.Strings(nodeIDs)

for _, id := range nodeIDs {
	node := graph.Nodes[id]
	// Process in order
}
```

### Query Set Membership Efficiently

Set membership is O(1) using map lookup:

```go
// Fast: O(1)
if _, ok := node.Sets["critical"]; ok {
	// Node is in "critical" set
}

// Slower: O(n) - avoid when possible
count := 0
for nodeID, node := range graph.Nodes {
	if _, ok := node.Sets["critical"]; ok {
		count++
	}
}
```

## Best Practices

1. **Always check warnings**: Even if `err == nil`, check the warnings slice
2. **Type-assert defensively**: Use `val, ok := ...` pattern for all attribute access
3. **Build adjacency lists once**: Don't rebuild during algorithms
4. **Remember duplicate edges**: When filtering/deduplicating, check for edge duplication
5. **Sort for determinism**: Any output should sort node IDs for reproducibility
6. **Don't validate structure**: GSL doesn't enforce acyclicity or tree properties
7. **Use set membership for queries**: `node.Sets[setName]` is O(1), much faster than filtering

## Common Gotchas

1. **Attributes are `interface{}`**: `node.Attributes["count"]` is `interface{}`, need `.(int)`
2. **Numbers are `float64`**: Integer literals in GSL parse as `float64`; convert if needed
3. **Unordered maps**: Iterating `graph.Nodes` or `graph.Sets` is random; sort IDs for consistency
4. **Duplicate edges matter**: `A->B` can appear multiple times; don't assume uniqueness
5. **Parent is attribute**: Node.Parent is cached from `parent` attribute; it's just a string
6. **No hierarchy in API**: Parent-child relationships are not enforced; check `parent` attribute
7. **Warnings are separate**: Non-fatal warnings are in second return; they don't stop parsing

## Quick Reference: Parse → Use → Serialize

```go
// 1. Parse
content, _ := os.ReadFile("input.gsl")
graph, warnings, err := gsl.Parse(bytes.NewReader(content))
if err != nil { log.Fatal(err) }
for _, w := range warnings { fmt.Println(w) }

// 2. Use (example: find all edges from node A)
for _, edge := range graph.Edges {
	if edge.From == "A" {
		fmt.Println(edge.To)
	}
}

// 3. Serialize
output := gsl.Serialize(graph)
fmt.Println(output)
```

That's everything you need to use GSL in Go!
