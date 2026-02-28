# GSL (GSL-Lang)

**GSL** is a small, declarative language for describing directed graphs with attributes and set-based grouping.

It is designed to be:

* Human-readable
* Deterministic
* Canonicalisable
* Easy to parse
* Easy to diff

GSL is not a visual graph language.
It is a textual graph representation designed for tooling, transformation, and programmatic analysis.

---

## What GSL Describes

A GSL document defines:

* Nodes
* Directed edges
* Sets (groupings)
* Arbitrary attributes on all of the above

---

## Quick Example

```gsl
# Declare sets
set flow [color="blue"]

# Declare nodes
node A: "Start" @flow
node B [flag]

# Declare edges
A->B [weight=1.2] @flow
```

This defines:

* 2 nodes
* 1 directed edge
* 1 set
* Attributes on nodes, edges, and sets

---

## Core Features

### Nodes

```gsl
node A
node B [flag, weight=2]
node C: "Hello"
```

### Edges

```gsl
A->B
A,B->C
C->D,E
```

Grouped edges expand automatically.

Duplicate edges are allowed.

---

### Sets and Membership

```gsl
set cluster [visible]

node A @cluster
A->B @cluster
```

Sets are named groupings.
Membership accumulates across declarations.

---

### Parent Relationships

```gsl
node C {
    node D
}
```

This is syntactic sugar for:

```gsl
node D [parent=C]
```

`parent` is treated as a normal attribute.

---

## Design Goals

GSL is designed to:

* Round-trip cleanly
* Produce a canonical internal representation
* Merge repeated declarations
* Preserve duplicate edges
* Keep semantics simple and explicit

It intentionally avoids:

* Edge identity
* Schema enforcement
* Graph correctness constraints (acyclicity, tree validity, etc.)

---

## Canonical Behaviour

A compliant parser must ensure:

```
parse(serialize(parse(input))) == parse(input)
```

Grouped edges expand.
Blocks become explicit `parent` attributes.
Implicit sets are materialised.

---

## Using the Library

The GSL library provides a Go API for parsing and manipulating GSL documents programmatically.

### Installation

```bash
go get github.com/dnnrly/gsl-lang
```

### Basic Usage

```go
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	gsl "github.com/dnnrly/gsl-lang"
)

func main() {
	// Read a GSL file
	content, err := os.ReadFile("graph.gsl")
	if err != nil {
		log.Fatal(err)
	}

	// Parse the GSL
	graph, warnings, err := gsl.Parse(bytes.NewReader(content))
	if err != nil {
		log.Fatal(err)
	}

	// Check for non-fatal warnings
	for _, w := range warnings {
		fmt.Printf("Warning: %v\n", w)
	}

	// Access the graph
	fmt.Printf("Nodes: %d, Edges: %d, Sets: %d\n", 
		len(graph.Nodes), len(graph.Edges), len(graph.Sets))
}
```

### Common Patterns

#### 1. Accessing Nodes

```go
// Get all nodes
for nodeID, node := range graph.Nodes {
	fmt.Printf("Node: %s\n", nodeID)
	
	// Access attributes
	if text, ok := node.Attributes["text"]; ok {
		fmt.Printf("  Text: %v\n", text)
	}
	
	// Check parent relationship
	if parent, ok := node.Attributes["parent"]; ok {
		fmt.Printf("  Parent: %v\n", parent)
	}
}
```

#### 2. Traversing Edges

```go
// Find all outbound edges from a node
for _, edge := range graph.Edges {
	if edge.From == "NodeA" {
		fmt.Printf("NodeA -> %s\n", edge.To)
		
		// Access edge attributes
		if method, ok := edge.Attributes["method"]; ok {
			fmt.Printf("  Method: %v\n", method)
		}
	}
}
```

#### 3. Working with Sets

```go
// Find all nodes in a specific set
for nodeID, node := range graph.Nodes {
	if _, isMember := node.Sets["critical"]; isMember {
		fmt.Printf("Critical node: %s\n", nodeID)
	}
}

// Find all edges in a specific set
for _, edge := range graph.Edges {
	if _, isMember := edge.Sets["production"]; isMember {
		fmt.Printf("Production edge: %s -> %s\n", edge.From, edge.To)
	}
}
```

#### 4. Computing Graph Statistics

```go
// Count nodes by set membership
setCounts := make(map[string]int)
for _, node := range graph.Nodes {
	for setName := range node.Sets {
		setCounts[setName]++
	}
}

for setName, count := range setCounts {
	fmt.Printf("%s: %d nodes\n", setName, count)
}
```

#### 5. Round-Tripping (Parse → Modify → Serialize)

```go
// Parse
graph, _, err := gsl.Parse(bytes.NewReader(content))
if err != nil {
	log.Fatal(err)
}

// Serialize back to canonical form
canonical := gsl.Serialize(graph)
fmt.Println(canonical)

// The serialized form can be re-parsed to produce an identical graph
graph2, _, _ := gsl.Parse(bytes.NewReader([]byte(canonical)))
// graph and graph2 are semantically equivalent
```

### API Reference

#### `Parse(io.Reader) (*Graph, []error, error)`

Parses GSL input and returns:
- `*Graph`: The parsed graph structure
- `[]error`: Non-fatal warnings (implicit set creation, name collisions, etc.)
- `error`: Fatal parse error, if any

#### `Serialize(*Graph) string`

Serializes a graph to canonical GSL form. The output can be re-parsed to produce an identical graph.

#### Graph Structure

```go
type Graph struct {
	Nodes map[string]*Node  // Nodes indexed by ID
	Edges []*Edge           // All edges (allows duplicates)
	Sets  map[string]*Set   // Named sets indexed by name
}

type Node struct {
	ID         string                 // Node identifier
	Attributes map[string]interface{} // Key-value attributes
	Sets       map[string]struct{}    // Set membership
	Parent     *string                // Cached parent reference
}

type Edge struct {
	From       string                 // Source node ID
	To         string                 // Target node ID
	Attributes map[string]interface{} // Key-value attributes
	Sets       map[string]struct{}    // Set membership
}

type Set struct {
	ID         string                 // Set identifier
	Attributes map[string]interface{} // Key-value attributes
}
```

### Examples and Tests

The [`examples/`](examples/) directory contains:
- 7 example GSL files demonstrating different graph patterns
- 10 runnable Example tests showing common usage patterns
- Documentation of warning types

Run examples:
```bash
go test ./examples -v
```

View example code:
- [Simple workflow parsing](examples/example_test.go#L13)
- [Set membership queries](examples/example_test.go#L74)
- [Edge traversal](examples/example_test.go#L106)
- [Serialization round-trip](examples/example_test.go#L172)

### Important Notes

- **Parsing is lenient**: Warnings are non-fatal. Parse will succeed even if implicit sets are created or name collisions occur.
- **Canonical form**: Serialized output may have different ordering than input but represents the same graph.
- **Graph structure**: No validation of graph properties (acyclicity, tree validity, etc.) is performed.
- **Attributes are untyped**: All attributes are stored as `interface{}`. Type assertion is needed for safety.
- **Duplicate edges preserved**: The graph preserves multiple edges between the same nodes (multiset).

---

## Reference

The formal language specification is defined in:

* [the specification](SPEC.md)
* [the grammar](GRAMMAR.md)
* [examples](examples/)
