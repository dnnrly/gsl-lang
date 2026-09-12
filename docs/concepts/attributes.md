# Concept: Untyped Attributes

**Audience:** everyone. **Prerequisite:** [Edges — relationships between nodes](edges.md).
**Next step:** [Named sets](sets.md).

Attributes attach additional information to graph elements — nodes, edges and sets all carry **`key="value"` pairs** that describe them. Attributes are deliberately **untyped**: GSL stores a small vocabulary of values (strings, numbers, booleans, identifiers) and leaves interpretation to you.

## The surface syntax

```gsl
node api [version="2.1", internal=true, replicas=3]
```

That is a string, a boolean, and a number on one node. Edges and sets take the same brackets:

```gsl
node checkout
node orders

set async [description="Non-blocking calls"]

checkout->orders [retries="3", timeout_ms=500] @async
```

## What "untyped" buys you

- **No schema to migrate.** Adding a new attribute is a text edit, not a type change. There is no attribute registry to update.
- **Truth lives where it makes sense.** `team="gateway"` on a node, `retries="3"` on an edge, `description=...` on a set — attributes are attached to the exact thing they describe.
- **Queries read them naturally.** `subgraph node.team == "payments"` and `subgraph edge.retries == "3"` are ordinary predicates, with one honest caveat below.

## The honest caveat

Because values are untyped, comparison is **string-ish by default**: a predicate must agree with how the value was written. `replicas="4"` is a different value from `replicas=4`, and queries treat them differently. This is the price of no-type-migrations, and it is paid once, at the boundary: model in text, interpret where it matters.

> **For the Go implementation:** reading an attribute returns an untyped value that you type-assert, with convenient accessors for the common cases — see the [Go reference](../../GO_REFERENCE.md#go-api-reference). That is an implementation detail, not part of the mental model.

## One reserved key

`parent` is special — it creates structure. See [Structure and nesting](structure-and-nesting.md) next.

---

**Next:** [Named sets](sets.md) — grouping nodes and edges into queryable categories.