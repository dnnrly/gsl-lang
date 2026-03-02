# GSL Query Language

A pipeline-style query language for selecting and filtering nodes from GSL graphs.

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/dnnrly/gsl-lang/query"
)

func main() {
	// Parse a query
	q, errs := query.ParseQuery(`start "AuthService" | flow out recursive | where critical = true`)
	if len(errs) > 0 {
		log.Fatal(errs)
	}

	// Serialize back to string
	queryStr := query.SerializeQuery(q)
	fmt.Println(queryStr)
	// Output: start "AuthService" | flow out recursive | where critical = true
}
```

## Syntax

The query language uses a **pipeline** model where operations are chained with `|`.

### Pipeline Steps

- **`start <node_ids>`** — Select starting nodes
  - Node IDs can be unquoted identifiers or quoted strings
  - Multiple IDs separated by commas: `start A, B, "Service"`

- **`flow <direction> [recursive|*] [where edge.attr op value]`** — Traverse edges
  - Direction: `in` (incoming), `out` (outgoing), `both`
  - Recursion: optional `recursive` or `*` for infinite traversal
  - Edge filter: optional `where edge.attr op value` to filter edges
  - Examples:
    - `flow out` — one level of outbound edges
    - `flow in recursive` — all incoming edges (recursive)
    - `flow out * where edge.color = "Blue"` — filtered recursive traversal

- **`where <attr> <op> <value>`** — Filter nodes by attribute
  - Filters the current selection by node properties
  - Multiple filters can be chained: `| where type = "Service" | where critical = true`

- **`minus <pipeline>`** — Remove nodes from selection
  - Subtracts results of a sub-pipeline from current selection
  - Example: `start A | flow out | minus (start B)`

### Combinators

Combine two pipelines:

- **`union`** — Merge results (set union)
- **`intersect`** — Intersection of results
- **`minus`** — Set difference

Example:
```
(start A | flow out) union (start B | flow in)
```

### Operators

Per GQL v1.0 specification, supported operators are:

- `=` — Equality
- `!=` — Inequality
- `<` — Less than
- `<=` — Less than or equal
- `>` — Greater than
- `>=` — Greater than or equal

### Values

- **Strings**: `"quoted value"`
- **Numbers**: `42`, `3.14`
- **Booleans**: `true`, `false`

## Examples

```go
// Simple flow
q, _ := query.ParseQuery("start A | flow out")

// Recursive traversal
q, _ := query.ParseQuery("start A | flow out recursive")

// With node filter
q, _ := query.ParseQuery("start A | flow out | where status = \"active\"")

// With edge filter
q, _ := query.ParseQuery(`start "Service" | flow out where edge.color = "Blue"`)

// Combinators
q, _ := query.ParseQuery("(start A | flow out) union (start B | flow in)")

// Complex pipeline
q, _ := query.ParseQuery(`
	start "AuthService" |
	flow out recursive |
	where critical = true |
	minus (start "Deprecated")
`)
```

## API

### Functions

- **`ParseQuery(input string) (*Query, []error)`** — Parse a query string
  - Returns the AST and any parse errors
  - Returns `(*Query, []error{...})` on success with empty error slice

- **`SerializeQuery(q *Query) string`** — Convert AST back to query string
  - Produces canonical form suitable for re-parsing
  - Useful for round-trip testing and debugging

### Types

#### AST Types

All AST types are exported for programmatic manipulation:

- `Query` — Root node with a `Root Step` field
- `Pipeline` — Sequence of steps
- `StartStep` — Selects starting nodes
- `FlowStep` — Traverses edges
- `FilterStep` — Filters nodes
- `MinusStep` — Subtracts nodes
- `CombinatorExpr` — Combines pipelines (union/intersect/minus)
- `FilterSpec` — Attribute comparison (for both node and edge filters)

#### Error Types

- `QueryError` — Structured error with type, message, and position
- `QueryErrorType` — Error classification:
  - `ErrorInvalidQuery` — Malformed AST (parse error)
  - `ErrorUnknownNodeID` — Unknown node in execution
  - `ErrorInvalidPredicate` — Invalid operator or type mismatch

Usage:
```go
if err != nil {
    if qErr, ok := err.(*QueryError); ok {
        switch qErr.Type {
        case ErrorUnknownNodeID:
            // Handle missing node
        case ErrorInvalidPredicate:
            // Handle type/operator mismatch
        case ErrorInvalidQuery:
            // Handle malformed query
        }
    }
}
```

## GQL v1.0 Compliance

This implementation conforms to the **GSL Query Language (GQL) v1.0 Production Specification**, including:

- **Deterministic evaluation** — Identical queries on identical graphs produce identical results
- **Set semantics** — All intermediate results are node sets; duplicates automatically removed
- **Strict subgraph construction** — Result includes all edges with both endpoints in the result set
- **Defined error types** — Three specific error types per spec:
  - `InvalidQuery` — Malformed AST
  - `UnknownNodeID` — Node ID does not exist in graph
  - `InvalidPredicate` — Invalid operator or type mismatch
- **No implementation-defined behavior** — All semantics fully specified

## Design

- **Separate from graph definition**: Query is a consumer of GSL graphs, not part of graph definition
- **Pipeline model**: Operations compose naturally left-to-right
- **Unambiguous grammar**: No conflicts between operators (e.g., `minus` as step vs operator)
- **Round-trip safe**: `parse(serialize(parse(x))) == parse(x)`
- **Comprehensive errors**: Includes line and column information with typed error codes

## Notes

- Queries are stateless and can be safely parsed and serialized concurrently
- The query language does **not** execute against graphs - it only parses queries into an AST
- For execution against a graph, use the AST to drive graph traversal logic in your application
- Token types are imported from the root `gsl` package
