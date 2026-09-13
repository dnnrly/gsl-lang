<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Concept: Edges — Relationships Between Nodes

**Audience:** everyone. **Prerequisite:** [Graphs and nodes](graph-model.md).
**Next step:** [Attributes](attributes.md).

An edge represents a **relationship between two nodes**: `gateway -> orders` says the gateway depends on orders. Every edge may carry its own attributes and its own set membership, and later you will encounter edge *dependencies*, where edges relate to other edges. But the relationship is the starting point.

## Edges are directed

```gsl
node gateway
node orders

gateway->orders [protocol="http"]
```

`gateway->orders` is not the same as `orders->gateway`. Direction is the point: requests flow one way, and queries can walk edges in *either* direction (`traverse in` vs `traverse out`) — which is exactly what turns "what does X call?" into "what depends on X?".

> **Background:** a graph whose edges each carry a direction is a [directed graph](https://en.wikipedia.org/wiki/Directed_graph). "Calls", "sends to" and "depends on" are relationships with a direction, which is why GSL stores one.

## Duplicate edges are distinct facts

Why would two **apparently identical** edges ever need to stay distinct? Because the duplication can itself be the information:

```gsl
node gateway
node orders

gateway->orders [protocol="http"]
gateway->orders [protocol="http"]
```

Two transports between the same pair, two invocations, two flows. GSL treats edges as a **multiset** — a collection that keeps every occurrence — precisely so those distinct facts are not merged away. When you want the collisions removed, queries can collapse them for you (`collapse`), and set membership on an edge is the normal way to tag *a particular* occurrence rather than all of them.

> **Background:** a collection that keeps every occurrence — duplicates and all — is a [multiset](https://en.wikipedia.org/wiki/Multiset). "The graph contains a multiset of edges" is just a compact way of saying duplicate edges stay distinct; there is no further counting or grouping behaviour to learn for GSL.

## Edges can carry attributes and membership

An edge takes the same brackets as a node, and can be a member of a set:

```gsl
set async

node checkout
node orders

checkout->orders [retries="3"] @async
```

`@async` marks *this edge* as a member of the `async` set — queryable the same way node membership is: `subgraph edge in @async`.

## Relationships between relationships

Edges can go further: an edge can be **labeled**, and a scoped (or `parent=`-declared) edge can state that it runs *only after* another edge — a relationship *between edges*:

```gsl
GATE: controlplane -> approve
PROMOTE: registry -> canary [parent=GATE]
```

Read that as *"`PROMOTE` cannot run until `GATE` has happened"* — while both remain first-class edges in the graph. This is how flagship [03 — release prerequisites](../../examples/flagships/03-release-prerequisites/README.md) models "when can we promote?" as pure graph. Promotions, approvals, gates — a whole class of workflow is just edges with edges depending on them. See [Edge dependencies](edge-dependencies.md).

## Why you can traverse both ways

Because edges are stored, not lost, every query can walk them in **both** directions without extra modelling:

- `subgraph node in @critical traverse out 1` — what critical things *call*.
- `subgraph node in @critical traverse in 1` — what critical things are *called by* (the dependency side).

The whole "what depends on X?" operation is a *direction* on the same stored edges.

---

**Next:** [Attributes](attributes.md) — the `key="value"` facts attached to nodes, edges and sets.