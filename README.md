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

## Table of Contents

- [What GSL Describes](#what-gsl-describes)
- [Quick Example](#quick-example)
- [Core Features](#core-features)
  - [Nodes](#nodes)
  - [Edges](#edges)
  - [Sets and Membership](#sets-and-membership)
  - [Parent Relationships](#parent-relationships)
- [Design Goals](#design-goals)
- [Canonical Behaviour](#canonical-behaviour)
- [Tools](#tools)
    - [gsl-diagram](#gsl-diagram)
- [Using the Library](#using-the-library)
  - [Installation](#installation)
  - [Basic Usage](#basic-usage)
  - [Common Patterns](#common-patterns)
  - [API Reference](#api-reference)
  - [Examples and Tests](#examples-and-tests)
  - [Important Notes](#important-notes)
- [Reference](#reference)

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

## Tools

### gsl-diagram

Convert GSL graphs to visual diagram formats (Mermaid, PlantUML).

#### Installation

```bash
go build -o gsl-diagram ./cmd/gsl-diagram
```

#### Usage

```bash
gsl-diagram -i graph.gsl -f mermaid -t component
gsl-diagram -i graph.gsl -f plantuml
cat graph.gsl | gsl-diagram -f mermaid > diagram.mmd
```

**Supported formats:**
- **Mermaid**: Component diagrams and flowcharts
- **PlantUML**: Component diagrams

See [cmd/gsl-diagram/README.md](cmd/gsl-diagram/README.md) for full documentation and examples.

---

## GSL Query Language

GSL also includes a **query language** for selecting and filtering nodes from a graph using a pipeline-based syntax.

### Query Quick Start

```go
import (
	gsl "github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/query"
)

// Parse a query
q, errs := query.ParseQuery(`start "AuthService" | flow out recursive | where critical = true`)
if len(errs) > 0 {
    log.Fatal(errs)
}

// Serialize back to string
queryStr := query.SerializeQuery(q)
fmt.Println(queryStr)
```

### Query Syntax

The query language uses a **pipeline** model where operations are chained with `|`:

**Pipeline Steps:**
- `start <node_ids>` — select starting nodes
- `flow <direction> [recursive|*] [where edge.attr op value]` — traverse edges
  - Direction: `in`, `out`, `both`
  - Optional `recursive` or `*` for recursive traversal
  - Optional edge filter
- `where <attr> <op> <value>` — filter nodes by attribute
- `minus <pipeline>` — remove nodes from selection

**Combinators:**
- `union` — merge two pipelines' results
- `intersect` — intersection of two pipelines' results
- `minus` — difference of two pipelines' results

**Operators (per GQL v1.0 spec):**
- `=` — equality
- `!=` — inequality
- `<` — less than
- `<=` — less than or equal
- `>` — greater than
- `>=` — greater than or equal

**Values:**
- Strings (quoted): `"value"`
- Numbers: `42`, `3.14`
- Booleans: `true`, `false`

### Query Examples

```gsl-query
# Select nodes and follow outbound edges
start A, B | flow out

# Recursive traversal with edge filter
start "Service" | flow out where edge.color = "Blue" recursive

# Filter by node attributes
start A | where status = "active"

# Multiple filters in a pipeline
start A | flow out | where critical = true | where type != "archived"

# Combine pipelines with union
(start A | flow out) union (start B | flow in)

# Complex combinators
(start X | flow out recursive) union (start Y) minus (start Z | flow both)
```

### Query API

Import the query package:
```go
import "github.com/dnnrly/gsl-lang/query"
```

#### `query.ParseQuery(input string) (*Query, []error)`

Parses a query string and returns:
- `*Query`: The parsed query AST
- `[]error`: Parse errors, if any

#### `query.SerializeQuery(q *Query) string`

Serializes a query AST back to a query string.

#### Query AST Structure

```go
type Query struct {
    Root Step  // Entry point of the query
}

type Pipeline struct {
    Steps []Step  // Sequence of pipeline steps
}

type StartStep struct {
    NodeIDs []string  // Starting node IDs
}

type FlowStep struct {
    Direction   string       // "in", "out", "both"
    Recursive   bool         // true if * or "recursive"
    EdgeFilter  *FilterSpec  // Optional edge filter
}

type FilterStep struct {
    Filter *FilterSpec  // Node attribute filter
}

type MinusStep struct {
    Pipeline *Pipeline  // Sub-pipeline to subtract
}

type CombinatorExpr struct {
    Type  string      // "union", "intersect", "minus"
    Left  *Pipeline   // Left operand
    Right *Pipeline   // Right operand
}

type FilterSpec struct {
    IsEdge bool        // true for edge filters, false for node filters
    Attr   string      // Attribute name
    Op     string      // Operator: "=", "!=", "contains", "matches"
    Value  interface{} // Comparison value
}
```

---

## Using the Library

The GSL library provides a Go API for parsing and manipulating GSL documents programmatically.

### For LLMs and AI Agents

If you are an LLM or AI agent that needs to work with GSL, see **[LLM_GUIDE.md](LLM_GUIDE.md)**.

The LLM Guide is a self-contained reference that covers:
- Complete GSL syntax with examples
- Go API reference with code patterns
- Algorithm implementations (topological sort, cycle detection, path finding)
- Best practices and common gotchas

You can copy the entire guide and use it as context for your tasks.

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
