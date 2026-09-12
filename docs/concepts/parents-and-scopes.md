# Concept: Parents, Scopes and Edge Dependencies

**Audience:** everyone using nesting or workflows. **Prerequisite:** [Edges as a multiset](edges.md).
**Next step:** [Lenient parsing and last-write-wins](parsing-and-merging.md).

A graph is a flat list of declarations on disk — but GSL gives you **structure on top of it** with two mechanisms that matter in very different ways: *parent scopes* for nodes, and *edge dependencies* for sequencing.

## Parent scopes: two shapes, one meaning

A child declares its parent either in **flat** style, with the reserved `parent` attribute:

```gsl
node C
node B [parent=C]
node A [parent=B]
```

or in **nested block** style, which means exactly the same thing:

```gsl
node C {
    node B {
        node A
    }
}
```

Both describe the same scaffold: `C` contains `B`, which contains `A`. Serialization may choose block or flat form depending on context, but the *semantics* are identical — that equivalence is covered by the round-trip guarantee (see [Canonical form](canonical-form.md)).

## Why nesting is information, not decoration

- **Components of components**: a database *inside* a service, a zone *inside* an availability region.
- **Fully-qualified identity**: a scoped node is addressed by its full path — `node "checkout/cart"` if `cart` lives under `checkout`.
- **Diagram structure**: converters render nesting as component boundaries.

## Edge dependencies: relationships *between edges*

This is GSL's distinctive trick, and it works for anything gated or sequential. Label an edge, then let another edge declare a `parent`:

```gsl
CONFIG: config -> controlplane [stage="apply"]
GATE: controlplane -> approve [stage="gate", parent=CONFIG]
```

Read the second line as: *`GATE` runs only after `CONFIG` has completed.* Both are still ordinary edges — queryable, renderable, in the same graph — but the model now records an ordering **between dependencies**, not just between nodes.

Nested scoped blocks express chains of this naturally:

```gsl
GATE: controlplane -> approve
PROMOTE: registry -> canary [parent=GATE] {
    CANARY: canary -> prod
    SMOKE: prod -> smoke
}
```

`PROMOTE` depends on `GATE`, and the steps inside its scope (`CANARY`, `SMOKE`) follow the promotion chain. Note the scoped blocks themselves must not carry a `parent` attribute — the label-level `[parent=...]` plus nesting covers the relationship.

This is how flagship [03 — release prerequisites](../../examples/flagships/03-release-prerequisites/README.md) answers "when can we promote?" without a single workflow keyword — promotions, approvals, release gates and task pipelines are all just edges with edges depending on them. GQL exposes the structure through the edge-dependency predicates (`edge parent exists`, `edge.depth`, `edge depends on ... scope` — see the [query tutorial](../../QUERY_TUTORIAL.md)).

> **Version note:** scoped/edge-dependency syntax is part of the language and the current reference implementation, but the released binary may lag — build the tools from source if a `LABEL: a -> b { ... }` model fails to parse on your install ([Getting Started](../getting-started/README.md#1-install-the-tools)).

## What changes on the way out

When serialized, children that were *only* in a nested block can be emitted level-by-level with an explicit `parent=` attribute — the point of canonical form is that the model you get back is the model you wrote, whichever honest-shaped typing you used.

---

**Next:** [Lenient parsing and last-write-wins](parsing-and-merging.md) — why merging two edits almost never loses your work.