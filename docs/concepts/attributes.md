# Concept: Untyped Attributes

**Audience:** everyone. **Prerequisite:** [Named sets](sets.md).
**Next step:** [Parents, scopes and edge dependencies](parents-and-scopes.md).

Nodes, sets and edges all carry **attributes** — `key="value"` pairs that describe them. Attributes are deliberately **untyped**: GSL stores a small vocabulary of values (strings, numbers, booleans, identifiers) and leaves interpretation to you.

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

Because values are untyped, comparison is **string-ish by default**: a predicate must agree with how the value was written. In Go, reading `node.Attributes["replicas"]` returns an `interface{}` that you **type-assert** (`node.GetInt("replicas", 0)` exists for convenience) — see the [Go reference](../../GO_REFERENCE.md#go-api-reference). This is the price of no-type-migrations, and it is paid once, at the boundary: model in text, interpret where it matters.

## One reserved key

`parent` is special — it creates structure. See [Parents, scopes and edge dependencies](parents-and-scopes.md) next.

---

**Next:** [Parents, scopes and edge dependencies](parents-and-scopes.md) — where attributes become structure, and edges depend on edges.