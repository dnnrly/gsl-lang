<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Supported GQL subset for dialect matching

**Status:** normative companion profile to [`DIALECTS_SPEC.md`](DIALECTS_SPEC.md)
(`0.2.0-draft`). It resolves the documentation requirement in §7.1 and
open decision §16.2 of that specification, and remains provisional until the
dialect specification is declared implementation-ready.

This document defines **exactly** which GQL queries a dialect `match.query`
may contain. Every other query MUST be rejected when the dialect is loaded,
and a rejected match query makes the dialect invalid (spec §7.4, §10.1).

The grammar and semantics are those of
[`../gql/QUERY_GRAMMAR.md`](../gql/QUERY_GRAMMAR.md) and
[`../gql/QUERY_SPEC.md`](../gql/QUERY_SPEC.md), restricted as described here.
This profile introduces **no new syntax**; it is a strict subset of GQL.

## 1. Accepted form

A supported match query is a single `subgraph_expr` with a predicate and
**no** traversal or scope clause:

```ebnf
match_query
    = subgraph_expr ;

subgraph_expr
    = "subgraph" , predicate ;
```

Concretely, a query is in the subset if and only if all of the following hold:

1. It parses under the full GQL grammar.
2. It contains **exactly one** expression (no `|` pipeline).
3. That expression is a `subgraph` expression.
4. The `subgraph` expression has **no** `traverse` or `scope` clause.

A valid match query evaluates to a graph, exactly as a `subgraph` expression
does in GQL (§3 of the grammar). The elements of that graph are the raw
match set; the dialect then selects only the elements whose kind matches the
definition's `representation` (spec §7.2).

## 2. Predicates

Any predicate accepted by the full grammar (§4–§13) is supported, provided it
is attached to a `subgraph` expression and does not introduce traversal:

- Equality: `node.kind == "service"`, `edge.protocol == "http"`.
- Inequality: `node.zone != "C"`.
- Existence: `node.team exists`, `edge.protocol not exists`.
- Set membership: `node in @critical`, `edge not in @deprecated`.
- Parent existence: `edge parent exists`, `edge parent not exists`.
- Dependency: `edge depends on edge.protocol == "http"`.
- Depth: `edge.depth == 0`, `edge.depth != 0`.
- Compound terms joined by `AND`: `node.team == "payments" AND node.zone == "B"`.

All terms in a compound predicate MUST target the same element kind, as in
full GQL.

### 2.1 Type-sensitive equality

As of the numeric-literal correction recorded in `DIALECTS_SPEC.md` §16 and
the GQL specification, value equality is **type-sensitive**: a number literal
matches only a numeric attribute, and a quoted literal matches only a string.
Dialect authors MUST quote values that are strings. The two queries below are
both valid GQL, but select different sets for a numeric `replicas` attribute;
only the first selects numeric values.

```gql
subgraph node.replicas == 3
```

```gql
subgraph node.replicas == "3"
```

## 3. Rejected forms

The following query forms MUST be rejected at dialect load time. Rejection is
reported as an invalid dialect, not as a matching or conformance outcome
(spec §10.1 lists match queries among the validity requirements). A query that
is syntactically invalid, or that parses but is outside the subset, is
rejected identically.

The examples below are shown in `text` fences because they are valid GQL: they
parse but leave the dialect subset. Rejection therefore cannot be decided from
the parse alone.

### 3.1 Traversal and scope

Traversal changes the selected set and is excluded from the initial subset.

```text
subgraph node.kind == "service" traverse out 1
```

`scope` is syntactic sugar for `traverse down all` and is excluded for the
same reason.

```text
subgraph edge.protocol == "http" scope
```

### 3.2 Pipelines

A pipeline applies further transformations after the match, which would change
the element set under validation.

```text
subgraph node.kind == "service" | remove orphans
```

### 3.3 Graph transformations

`make`, `remove` and `collapse` are transformations, not matches.

```text
make node.status = "reviewed" where node.team == "payments"
```

```text
remove edge where edge.protocol == "tcp"
```

```text
collapse into platform where node.team == "platform"
```

### 3.4 Graph sources, bindings and algebra

`from`, named-graph bindings and graph algebra operate on named working graphs
and are excluded.

```text
from API
```

```text
(subgraph node.kind == "service") as SVC | SVC + OTHER
```

```text
SVC + OTHER
```

### 3.5 Syntax errors

A query that does not parse is rejected regardless of form.

```invalid-gql
subgraph node.kind ==
```

## 4. Diagnostics

Implementations MUST reject an out-of-subset or malformed query when loading
the dialect, before any graph is validated, and MUST NOT approximate the query
with different semantics. Diagnostics SHOULD locate the offending field
(for example `elements[Service].match.query`) and distinguish, where
practicable:

- **Unsupported syntax** — the query parses but leaves the subset (for
  example, traversal or a pipeline).
- **Invalid syntax** — the query does not parse.

Stable machine-readable diagnostic codes remain an open decision
(`DIALECTS_SPEC.md` §16.4); this draft does not fix their form.

## 5. Examples

A minimal dialect definition using an in-subset query:

```yaml
elements:
  Service:
    representation: node
    match:
      query: 'subgraph node.kind == "service"'
    attributes:
      type: object
```

Match sets usable by a `representation: node` definition:

```gql
subgraph node.kind == "service" AND node.active exists
```

Match sets usable by a `representation: edge` definition:

```gql
subgraph edge.protocol == "http"
```

Match sets usable by a `representation: set` definition. Because a GQL
`subgraph` carries every set declared in the input graph, any in-subset query
selects all sets; dialect authors SHOULD NOT rely on the predicate to filter
sets (see the known limitation recorded in `DIALECTS_SPEC.md`).

```gql
subgraph node.kind == "service"
```

## 6. Normative references

- [`DIALECTS_SPEC.md`](DIALECTS_SPEC.md) — dialect specification (§6.2 match,
  §7 matching, §10 outcomes).
- [`../gql/QUERY_SPEC.md`](../gql/QUERY_SPEC.md) — GQL semantics.
- [`../gql/QUERY_GRAMMAR.md`](../gql/QUERY_GRAMMAR.md) — GQL grammar.
