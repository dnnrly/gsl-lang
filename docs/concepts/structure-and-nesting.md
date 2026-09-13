<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Concept: Structure and Nesting

**Audience:** everyone modelling systems with containment. **Prerequisite:** [Attributes](attributes.md).
**Next step:** [Edge dependencies](edge-dependencies.md).

A GSL file is a sequence of declarations, but GSL gives you **structure on top of it**: a node can *contain* other nodes. Nesting is how you record that one thing genuinely lives inside another.

## Two shapes, one meaning

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

Containment is a real relationship in your model, not an organisational convenience:

- **Components of components**: a database *inside* a service, a zone *inside* an availability region, a table *inside* a database.
- **Fully-qualified identity**: a scoped node is addressed by its full path — `node "checkout/cart"` if `cart` lives under `checkout`. The same name in two different parents is two different things.
- **Diagram structure**: converters render nesting as component boundaries, so containment survives into derived views.

## What changes on the way out

When serialized, children that were *only* in a nested block can be emitted level-by-level with an explicit `parent=` attribute — the point of canonical form is that the model you get back is the model you wrote, whichever spelling you used.

Use containment when "lives inside" is the relationship you are recording. The next page covers the other thing the `parent` attribute can express — a *relationship between relationships*.

---

**Next:** [Edge dependencies](edge-dependencies.md) — allowing one relationship itself to depend on another.