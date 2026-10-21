<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Supported GQL subset for dialect matching

**Status:** normative companion profile to [`DIALECTS_SPEC.md`](DIALECTS_SPEC.md)
(`0.3.0-draft`). It resolves the documentation requirement in §7.1 and
open decision §16.2 of that specification, and remains provisional until the
dialect specification is declared implementation-ready.

This document defines **exactly** which match expressions a dialect
`match.query` may contain. Every other expression MUST be rejected when the
dialect is loaded, and a rejected match expression makes the dialect invalid
(spec §6.2, §7.4, §10.1).

The grammar and semantics are those of
[`../gql/QUERY_GRAMMAR.md`](../gql/QUERY_GRAMMAR.md) and
[`../gql/QUERY_SPEC.md`](../gql/QUERY_SPEC.md), restricted as described here.
This profile introduces **no new syntax**: it is a strict subset of GQL, and
each accepted expression expands to a full GQL `subgraph` predicate.

## 1. Accepted form

A match expression is a single **attribute predicate** with no element prefix.
The `subgraph` wrapper and the element namespace are implicit; both are
supplied by the definition's `representation` (spec §6.1, §6.2).

```ebnf
match_query
    = attribute_predicate ;

attribute_predicate
    = attribute_term , { "AND" , attribute_term } ;
```

where each `attribute_term` is one of the predicate forms in Section 2, written
with a **bare attribute name** rather than a `node.`/`edge.` prefixed one.

**Expansion.** Before evaluation, an implementation MUST expand the expression
by:

1. wrapping it in `subgraph`; and
2. prefixing every attribute reference with the definition's `representation`.

For a `representation: node` definition:

```text
kind == "service"
```

expands to:

```text
subgraph node.kind == "service"
```

For a `representation: edge` definition, the same expression expands to:

```text
subgraph edge.kind == "service"
```

The expansion produces an expression that parses under the full GQL grammar and
evaluates to a graph, exactly as a `subgraph` expression does in GQL (§3 of the
grammar). The elements of that graph are the raw match set; the dialect then
selects only the elements whose kind matches the definition's `representation`
(spec §7.2). Because the expansion fixes the element prefix to the
representation, the raw match set and the selected kind agree by construction.

**Set representation.** The `subgraph` grammar admits only `node.` and `edge.`
element prefixes; there is no `set.` prefix. A `representation: set` definition
therefore cannot be expanded with its representation as the prefix. Its `query`
MUST still be a well-formed in-subset attribute predicate, and implementations
MUST validate it for syntax and subset membership, but they MUST NOT evaluate it
to filter sets: a `set` definition selects every set in the input graph (spec
§7.5).

## 2. Predicates

Any predicate accepted by the full grammar (§4–§13) is supported, provided it is
written as a bare attribute term joined by `AND`:

- Equality: `kind == "service"`, `protocol == "http"`.
- Inequality: `zone != "C"`.
- Existence: `team exists`, `protocol not exists`.
- Set membership: `in @critical`, `not in @deprecated`.
- Parent existence: `parent exists`, `parent not exists`.
- Dependency: `depends on protocol == "http"`.
- Depth: `depth == 0`, `depth != 0`.
- Compound terms joined by `AND`: `team == "payments" AND zone == "B"`.

All terms in a compound predicate target the definition's representation, so
they necessarily refer to the same element kind.

### 2.1 Type-sensitive equality

As of the numeric-literal correction recorded in `DIALECTS_SPEC.md` §16 and
the GQL specification, value equality is **type-sensitive**: a number literal
matches only a numeric attribute, and a quoted literal matches only a string.
Dialect authors MUST quote values that are strings. The two expressions below
are both valid, but select different sets for a numeric `replicas` attribute;
only the first selects numeric values.

```text
replicas == 3
```

```text
replicas == "3"
```

### 2.2 Rejected explicit prefixes

A match expression MUST NOT contain an explicit element prefix. Writing
`node.kind == "service"` (or `edge.`/`set.`) is an invalid dialect, even when
the prefix agrees with the definition's representation, because the prefix is
defined by the representation and stating it twice invites disagreement.

## 3. Rejected forms

The following expression forms MUST be rejected at dialect load time. Rejection
is reported as an invalid dialect, not as a matching or conformance outcome
(spec §10.1 lists match expressions among the validity requirements). A query
that is syntactically invalid, or that parses but is outside the subset, is
rejected identically.

The examples below are shown in `text` fences because they are valid GQL: they
parse but leave the dialect subset. Rejection therefore cannot be decided from
the parse alone.

### 3.1 Explicit element prefixes

The element prefix is implied by the representation and MUST NOT be written.

```text
node.kind == "service"
```

```text
edge.protocol == "http"
```

### 3.2 An explicit `subgraph` wrapper

The `subgraph` keyword is implied and MUST NOT be written. An expression that
starts with `subgraph` is outside the subset (it is the expanded form, not the
accepted input form).

```text
subgraph node.kind == "service"
```

### 3.3 Traversal and scope

Traversal changes the selected set and is excluded from the initial subset.

```text
kind == "service" traverse out 1
```

`scope` is syntactic sugar for `traverse down all` and is excluded for the same
reason.

```text
protocol == "http" scope
```

### 3.4 Pipelines

A pipeline applies further transformations after the match, which would change
the element set under validation.

```text
kind == "service" | remove orphans
```

### 3.5 Graph transformations

`make`, `remove` and `collapse` are transformations, not matches.

```text
make status = "reviewed" where team == "payments"
```

```text
remove edge where protocol == "tcp"
```

```text
collapse into platform where team == "platform"
```

### 3.6 Graph sources, bindings and algebra

`from`, named-graph bindings and graph algebra operate on named working graphs
and are excluded.

```text
from API
```

```text
(kind == "service") as SVC | SVC + OTHER
```

```text
SVC + OTHER
```

### 3.7 Syntax errors

An expression that does not expand to a valid GQL query is rejected regardless
of form.

```invalid-gql
kind ==
```

## 4. Diagnostics

Implementations MUST reject an out-of-subset or malformed expression when
loading the dialect, before any graph is validated, and MUST NOT approximate the
expression with different semantics. Diagnostics SHOULD locate the offending
field (for example `elements[Service].match.query`) and distinguish, where
practicable:

- **Unsupported syntax** — the expression validates as GQL after expansion but
  leaves the subset (for example, an explicit element prefix, an explicit
  `subgraph` wrapper, traversal or a pipeline).
- **Invalid syntax** — the expression does not expand to a valid GQL query.

Stable machine-readable diagnostic codes remain an open decision
(`DIALECTS_SPEC.md` §16.4); this draft does not fix their form.

## 5. Examples

A minimal dialect definition using an in-subset expression:

```yaml
elements:
  Service:
    representation: node
    match:
      query: 'kind == "service"'
    attributes:
      type: object
```

This expands to `subgraph node.kind == "service"`.

Match sets usable by a `representation: node` definition:

```text
kind == "service" AND active exists
```

Match sets usable by a `representation: edge` definition:

```text
protocol == "http"
```

Match sets usable by a `representation: set` definition. Because a GQL
`subgraph` carries every set declared in the input graph, any in-subset
expression selects all sets; the predicate is validated but never evaluated, so
dialect authors SHOULD NOT rely on it to filter sets (see the known limitation
recorded in `DIALECTS_SPEC.md`).

```text
kind == "service"
```

## 6. Normative references

- [`DIALECTS_SPEC.md`](DIALECTS_SPEC.md) — dialect specification (§6.1
  representation, §6.2 match, §7 matching, §10 outcomes).
- [`../gql/QUERY_SPEC.md`](../gql/QUERY_SPEC.md) — GQL semantics.
- [`../gql/QUERY_GRAMMAR.md`](../gql/QUERY_GRAMMAR.md) — GQL grammar.
