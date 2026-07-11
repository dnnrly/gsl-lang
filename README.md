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

## For LLMs and AI Agents

GSL **is a language** — not just a Go library. This repository is the **canonical home** of the language: normative specification, reference implementation, CLI tools, LSP server, query language, and VS Code extension.

| Start with this file | If you need... |
|---|---|
| [`SPEC.md`](SPEC.md) | The authoritative language specification |
| [`GRAMMAR.md`](GRAMMAR.md) | The formal grammar (for implementing a parser) |
| [`GSL_GUIDE.md`](GSL_GUIDE.md) | A self-contained GSL syntax & semantics reference |
| [`QUERY_SPEC.md`](QUERY_SPEC.md) | The query language specification |
| [`GQL_GUIDE.md`](GQL_GUIDE.md) | A self-contained GQL syntax & semantics reference |
| [`GO_REFERENCE.md`](GO_REFERENCE.md) | The Go reference implementation guide |
| [`AGENTS.md`](AGENTS.md) | Development instructions for contributing |

---

## Table of Contents

- [What GSL Describes](#what-gsl-describes)
- [Quick Example](#quick-example)
- [Core Features](#core-features)
  - [Nodes](#nodes)
  - [Edges](#edges)
  - [Sets and Membership](#sets-and-membership)
  - [Parent Relationships](#parent-relationships)
  - [Edge Labels and Scoping](#edge-labels-and-scoping)
- [Design Goals](#design-goals)
- [Canonical Behaviour](#canonical-behaviour)
- [Tools](#tools)
    - [gsl-diagram](#gsl-diagram)
- [Using the Library](#using-the-library)
  - [Programmatic Usage (Go)](#programmatic-usage-go)
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

### Edge Labels and Scoping

GSL supports **edge labels** and **scoped edges** for expressing explicit dependencies between edges.

**Edge labels** assign a name to an edge:

```gsl
E1: A -> B
E2: B -> C [weight=2]
```

Labels are globally unique and enable explicit dependency references.

**Edge scoping** allows nesting edges to express implicit dependencies:

```gsl
# B -> C implicitly depends on A -> B
A -> B {
    B -> C
}
```

This is syntactic sugar for:

```gsl
E1: A -> B
B -> C [parent=E1]
```

Scoped edges flatten to explicit `parent` attributes during canonicalization.

**Explicit `parent`** references any labeled edge:

```gsl
E1: A -> B
C -> D [parent=E1]
```

Scoped edges automatically get implicit `parent` on their parent edge. You cannot use explicit `parent` inside a scoped edge - use labels on the parent edge instead:

Use cases:
- Dependency graphs where edges represent tasks
- Data pipelines with explicit stage ordering
- Workflow orchestration with explicit prerequisites

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

GSL includes a **query language** for selecting, filtering, and transforming graphs using a pipeline-based syntax.

Queries enable you to:

* Extract subgraphs by filtering nodes and edges
* Traverse graph neighbourhoods (incoming, outgoing, bidirectional)
* Assign and remove attributes
* Merge (collapse) nodes
* Combine multiple graphs using set operations (union, intersection, difference)

### Query Concepts

The core idea is a **pipeline of expressions** where each expression receives a graph and produces a new graph:

```
input graph → expr₁ → expr₂ → … → result graph
```

Expressions are separated by `|` and evaluated left-to-right, similar to a Unix shell pipeline but for graphs.

### Pipeline Expressions

| Expression | Syntax | Purpose |
|---|---|---|
| **Source** | `from *` or `from NAME` | Switch the working graph |
| **Subgraph** | `subgraph <predicate> [traverse <dir> <depth>]` | Filter nodes or edges, optionally traverse |
| **Make** | `make <path> = <value> where <predicate>` | Assign attributes to matching elements |
| **Remove** | `remove edge where <predicate>` | Delete matching edges |
| **Remove** | `remove node.<attr> where <predicate>` | Delete attributes from matching nodes |
| **Remove** | `remove orphans` | Delete nodes with no incident edges |
| **Collapse** | `collapse into <id> where <predicate>` | Merge matching nodes into one |
| **Binding** | `(<pipeline>) as NAME` | Save pipeline result as a named graph |
| **Algebra** | `NAME + NAME2`, `NAME & NAME2`, etc. | Combine named graphs (union, intersection, etc.) |

### Subgraph Filtering (Most Common)

Extract subgraphs by matching nodes or edges:

#### Node Matching

```
subgraph node.team == "payments"
```

Selects all nodes where `team` equals `"payments"` and includes edges between matched nodes only.

#### Edge Matching

```
subgraph edge.protocol == "grpc"
```

Selects all edges where `protocol` is `"grpc"` and includes their source and target nodes.

#### Traversal

After matching, optionally explore the graph neighbourhood:

```
subgraph node.team == "payments" traverse out 1
subgraph node.team == "payments" traverse in all
```

**Directions:** `in`, `out`, `both`  
**Depths:** `1`, `2`, `N` (hops), or `all` (unlimited)

### Predicates

Predicates filter by attributes, set membership, or existence:

| Form | Example | Meaning |
|---|---|---|
| Equality | `node.team == "payments"` | Attribute equals value |
| Inequality | `node.zone != "C"` | Attribute does not equal value |
| Exists | `node.team exists` | Attribute is present |
| Not exists | `edge.debug not exists` | Attribute is absent |
| Set membership | `node in @critical` | Node belongs to set |
| Set non-membership | `edge not in @deprecated` | Node does not belong to set |
| Compound | `node.team == "payments" AND node.zone == "B"` | Both conditions true |

**Important:** Cannot mix `node.` and `edge.` in one predicate. Only `AND` is supported (no `OR`).

### Transformation Examples

#### Example 1: Basic Filtering

```
subgraph node.team == "payments"
```

Result: All nodes from the payments team and edges between them.

#### Example 2: Filtering + Cleanup

```
subgraph node.team == "payments" | remove orphans
```

Result: Payments team nodes with any orphaned nodes removed.

#### Example 3: Traversal + Removal

```
subgraph node.team == "payments" traverse out 1 | remove edge where edge.protocol == "tcp"
```

Result: Payments team and their direct outbound neighbours, excluding TCP edges.

#### Example 4: Node Collapse

```
subgraph node.zone == "A" | collapse into zone_a_cluster where node.team == "platform"
```

Result: All nodes in zone A, with platform team nodes merged into a single `zone_a_cluster` node.

#### Example 5: Named Graphs + Set Operations

```
(subgraph node.team == "payments") as PAY
| from *
| (subgraph node.team == "identity") as ID
| PAY + ID
```

Result: Union of payments team and identity team nodes.

### Query Combinators (Graph Algebra)

After binding named graphs, combine them:

```
GRAPH1 + GRAPH2    # Union: all nodes and edges from both
GRAPH1 & GRAPH2    # Intersection: only shared elements
GRAPH1 - GRAPH2    # Difference: in GRAPH1 but not GRAPH2
GRAPH1 ^ GRAPH2    # Symmetric difference: in exactly one
```

When the same node appears in both graphs, attributes from the right-hand side overwrite conflicts.

### More Examples

For additional examples and detailed explanations, see:

* [QUERY_TUTORIAL.md](QUERY_TUTORIAL.md) — Step-by-step learning guide  
* [QUERY_SPEC.md](QUERY_SPEC.md) — Complete formal specification  
* [query/](query/) — Go package documentation

---

### Go API Reference

Parse and serialize queries programmatically:

```go
import "github.com/dnnrly/gsl-lang/query"

// Parse a query string
q, errs := query.ParseQuery(`subgraph node.team == "payments" | remove orphans`)

// Serialize back to a query string
queryStr := query.SerializeQuery(q)
```

**Functions:**

| Function | Description |
|---|---|
| `query.ParseQuery(input string) (*Query, []error)` | Parse a GQL query string into an AST |
| `query.SerializeQuery(q *Query) string` | Serialize a query AST back to a string |

**AST Types:** `Query`, `Pipeline`, `StartStep`, `FlowStep`, `FilterStep`, `MinusStep`, `CombinatorExpr`, `FilterSpec` — see `query/` package for full definitions.

---

## Using the Library

The GSL library provides a Go API for parsing and manipulating GSL documents programmatically.

### For LLMs and AI Agents

If you are an LLM or AI agent that needs to work with GSL, see **[GSL_GUIDE.md](GSL_GUIDE.md)**.

The GSL Guide is a self-contained reference that covers:
- Complete GSL syntax with examples
- Language semantics and design notes
- Best practices and common gotchas

You can copy the entire guide and use it as context for your tasks.

### Programmatic Usage (Go)

```bash
go get github.com/dnnrly/gsl-lang
```

```go
graph, warnings, err := gsl.Parse(bytes.NewReader(content))
canonical := gsl.Serialize(graph)
```

For full API reference, code patterns, and algorithm implementations, see **[GO_REFERENCE.md](GO_REFERENCE.md)**.

For query language usage, see the [`query/`](query/) package.

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
