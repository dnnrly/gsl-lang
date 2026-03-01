# GSL (Graph Specification Language) - Complete Guide for LLMs

This document contains everything needed to understand, write, and use GSL files and the Go library.

## What is GSL?

GSL is a small, declarative language for describing **directed graphs** with attributes and set-based grouping. It is:
- Human-readable and easy to parse
- Deterministic and canonicalisable
- Designed for tooling, transformation, and programmatic analysis
- NOT a visual language—purely textual

## GSL Syntax Reference

### Basic Nodes

```gsl
node A
node B [flag]
node C [weight=2, color="red"]
node D: "Display Text"
```

**Rules:**
- Node IDs must match: `[A-Za-z_][A-Za-z0-9_]*`
- Cannot use reserved keywords: `node`, `set`, `true`, `false`
- Attributes use `[key]` or `[key=value]` syntax
- Text shorthand `: "string"` sets the `text` attribute
- Text shorthand cannot be combined with attributes in same declaration

### Basic Edges

```gsl
A->B
A,B->C
D->E,F
A->B [weight=1.5, color="blue"]
A->B: "label"
A->B [weight=1.5] @setname
```

**Rules:**
- Edges expand grouped nodes automatically: `A,B->C` becomes `A->C` and `B->C`
- Cannot have grouped nodes on both sides: `A,B->C,D` is a syntax error
- Attributes supported same as nodes (no NodeRef types allowed)
- Text shorthand supported
- Duplicate edges are allowed and preserved
- Edges can have text shorthand and attributes but not both in same declaration

### Sets (Groupings)

```gsl
set critical [backup=true]
set services [env="production"]

node ServiceA @critical @services
node ServiceB @critical
A->B @services
```

**Rules:**
- Sets are named groupings that nodes and edges can belong to
- Sets are declared with: `set <name> [attributes]`
- Membership added via `@setname` suffix after node/edge declaration
- Membership accumulates across multiple declarations
- Sets created implicitly when referenced but not declared (generates warning)

### Parent-Child Relationships

```gsl
# Block syntax (syntactic sugar)
node Parent {
  node Child1
  node Child2
}

# Explicit equivalent
node Parent
node Child1 [parent=Parent]
node Child2 [parent=Parent]
```

**Rules:**
- Block syntax is shorthand for implicit `parent` attributes
- Explicit parent in block overrides implicit (generates warning)
- `parent` attribute uses NodeRef type (can reference any node ID)
- Blocks can be nested

### Attribute Types and Values

Attributes can have these value types:

```gsl
node A [
  text="string",
  count=42,
  ratio=3.14,
  enabled=true,
  disabled=false,
  parent=OtherNode,
  flag
]
```

**Value types:**
- **String**: `"anything in quotes"`, supports escape sequences: `\"`, `\\`, `\n`, `\t`
- **Number**: Integer or float, no sign/exponent: `42`, `3.14`
- **Boolean**: `true` or `false` (must be lowercase)
- **NodeRef**: Bare identifier (only allowed in node attributes, not edges/sets): `OtherNode`
- **Empty**: Bare key with no value means empty attribute: `flag`

**Rules:**
- No duplicate keys in single declaration
- Attributes are untyped when accessed via API (must type-assert)
- Nodes can have `parent` attributes pointing to other nodes
- Edges cannot have NodeRef values
- Sets cannot have NodeRef values

### Comments

```gsl
# This is a comment
node A  # Inline comments work too
# set B  # Commented out
```

## Complete GSL Example

```gsl
# Microservices architecture example
set frontend [color="blue", visible=true]
set backend [color="green"]
set critical [backup=true]

# Frontend services
node WebUI [text="Web Interface"] @frontend
node Dashboard [text="Dashboard"] @frontend

# Backend services
node API [text="API Server"] @backend @critical
node Database [text="PostgreSQL"] @backend @critical
node Cache [text="Redis"] @backend

# API structure
node AuthModule [parent=API, timeout=30] @API
node DataModule [parent=API, timeout=60] @API

# Connections
WebUI -> API [protocol="REST", timeout=5000]
Dashboard -> API
API -> Database [pool_size=20]
API -> Cache [ttl=3600]
Database -> Cache

# Complex edge declaration
AuthModule -> Database, Cache
```

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

## Important Notes

### Parsing Behavior

- **Lenient parsing**: Parse succeeds even with warnings (implicit sets, name collisions)
- **Three return values**: `Parse()` returns `(*Graph, []error, error)` where:
  - First return: The parsed graph (even if warnings exist)
  - Second return: Non-fatal warnings
  - Third return: Fatal parse error (only one of these two is non-nil)
- **Check warnings**: Always iterate the warning slice and check for issues

### Graph Properties

- **Attributes are untyped**: Values stored as `interface{}`, must type-assert to use
- **No schema validation**: GSL doesn't validate graph structure (no acyclicity checking)
- **Duplicate edges allowed**: The `Edges` slice is a multiset and preserves duplicates
- **Set membership is separate**: Nodes and edges have `.Sets` map, not list of members
- **Parent is just an attribute**: `parent` attribute is normal except it's cached in `Node.Parent`

### Serialization

- **Canonical form**: `Serialize()` produces deterministic output
- **Ordering may differ**: Serialized output may reorder elements but parses to same graph
- **Round-trip guarantee**: `parse(serialize(parse(input))) == parse(input)`
- **No data loss**: All information is preserved (attributes, sets, duplicates, etc.)

### Warnings Types

```
- "implicit set creation: %q"        // Set used but never declared
- "%d:%d: parent override inside block"  // Explicit parent differs from block parent
- "node and set name collision: %q"  // Same ID used as both node and set
```

Warnings are informational only—parsing continues.

## Best Practices

1. **Always check warnings**: Even if `err == nil`, check for non-fatal warnings
2. **Type-assert attributes**: Don't assume attribute types
3. **Build adjacency lists**: Create `map[string][]string` for algorithms before iterating `graph.Edges`
4. **Handle duplicate edges**: Remember `graph.Edges` may contain duplicates
5. **Sort for determinism**: When printing/iterating, sort node IDs for consistent output
6. **Validate before use**: Don't assume graph is acyclic or well-formed unless you validate
7. **Use set membership for queries**: Checking `node.Sets[setName]` is O(1), much faster than filtering nodes

## Common Gotchas

1. **Text shorthand and attributes clash**: Can't do `node A: "text" [attr=1]` - split into separate declarations
2. **Attributes are interface{}**: `node.Attributes["count"]` is `interface{}`, need `.(int)`
3. **NodeRef only in nodes**: Can't put `parent=SomeNode` in edge or set attributes
4. **Grouped edges on both sides fails**: `A,B->C,D` is syntax error, must be `A,B->C` or `A->C,D`
5. **Sets accumulate**: Multiple `@setname` on same node adds to existing set membership
6. **Implicit sets create warnings**: Using `@undeclared` without `set undeclared` produces warning
7. **No set-of-sets**: Sets can't contain other sets, only nodes and edges can be in sets
8. **Parent is attribute**: Setting parent doesn't create hierarchical structure in API—it's just a string attribute

## Quick Reference: Parse → Use → Serialize

```go
// 1. Parse
content, _ := os.ReadFile("input.gsl")
graph, warnings, err := gsl.Parse(bytes.NewReader(content))
if err != nil { log.Fatal(err) }

// 2. Use
for _, edge := range graph.Edges {
	if edge.From == "A" {
		fmt.Println(edge.To)
	}
}

// 3. Serialize
output := gsl.Serialize(graph)
fmt.Println(output)
```

That's everything you need to read, write, and work with GSL!
