<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Concept: Canonical Form and Diffability

**Audience:** everyone treating a graph as an asset. **Prerequisite:** [Lenient parsing and last-write-wins](parsing-and-merging.md).
**Next step:** [Tutorials](../tutorials/README.md).

**Canonical form** is the property that makes text graphs into *version-controlled assets*: within a given implementation and version, the same logical graph serializes to a **deterministic text**, so diffs are small, reviewable and meaningful. No reordering noise, no regenerated blobs, no "did anything actually change?" commits.

## Two guarantees, stated plainly

1. **Deterministic serialization.** Within a given implementation and version, given the same graph, the serializer emits the same bytes — with defined ordering rules, not "whatever a map iterate gave us".
2. **Round-trip stability.** `parse(serialize(parse(x)))` is semantically identical to `parse(x)`. You can always describe what canonization did, because it provably changed nothing but the shape of the text.

The formal ordering rules are in [SPEC §11 Canonicalisation](../../SPEC.md#11-canonicalisation); the short version for everyday use:

- **Sets** — sorted by ID.
- **Nodes** — ordered by *first appearance in an edge declaration*; nodes that appear in no edge are sorted by ID.
- **Edges** — preserved in declaration order (as a multiset, see [Edges — relationships between nodes](edges.md)).
- **Children** — the same rules within each parent scope.

## Seeing it happen

"A node order that follows the edges, not the writer" matters more than it sounds. Both blocks below are the same graph; only the author wrote the node lines in a different order. Running `gsl-query ""` — the empty query — *canonises* the input:

```bash
gsl-query "" < model.gsl
```

Before (authored, hand-ordered):

```gsl
node payments [team="payments"]
node gateway [team="gateway"] @critical
node orders [team="orders"] @critical

gateway->orders [protocol="http"]
orders->payments [protocol="grpc"]
```

After (canonical):

```gsl
set critical

node gateway [team="gateway"] @critical
node orders [team="orders"] @critical
node payments [team="payments"]

gateway->orders [protocol="http"]
orders->payments [protocol="grpc"]
```

The node order now reads the way the graph *is* (following the arrows), and the implicit `critical` set is spelled out. Same graph, smaller future diffs: an author shuffling `payments` around can no longer pollute a genuine change with cosmetic churn.

## Why this matters

- **Reviews read the truth.** A diff of `@critical` membership changing is a *semantic* diff, readable in review, exactly like a code review.
- **Query output is shareable.** Answers from `gsl-query` are canonical GSL — paste one into an issue and anyone can describe exactly what they ran.
- **CI can check freshness.** Because serialization is deterministic, a pipeline can regenerate and diff (see the [render-views-in-CI recipe](../cookbook/render-a-view-for-ci.md)).

## Two honest caveats

- **Authored files need not be canonical.** The flagship models are hand-written and therefore *not* byte-identical to their canonical form (comments, order, even `set` spelling). Canonical form is what the **toolchain emits**, not a requirement on what you type. If you enforce freshness, compare *tool-vs-tool* (regenerate and diff), never *tool-vs-authored-file*.
- **Canonical ordering is per-implementation.** Deterministic within a build and version, but exact ordering can vary between versions — still valid canonical GSL either way. Build from the repository source to reproduce the committed example outputs byte-for-byte (see [Getting Started](../getting-started/README.md#1-install-the-tools)).

---

**Next:** choose a route — [Modelling with GSL](../tutorials/modelling-with-gsl.md), [the flagship tutorials and query tutorial](../tutorials/README.md), or [a specific recipe](../cookbook/README.md).