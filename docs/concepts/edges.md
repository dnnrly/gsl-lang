# Concept: Edges as a Multiset

**Audience:** everyone. **Prerequisite:** [Graphs and nodes](graph-model.md).
**Next step:** [Named sets](sets.md).

An edge is a **directed arrow between two nodes**. Every edge may carry its own attributes, a membership, and — uniquely — a *relationship to other edges*. But the detail that shapes everything downstream is this: **edges form a multiset.**

## Edges are directed

```gsl
node gateway
node orders

gateway->orders [protocol="http"]
```

`gateway->orders` is not the same as `orders->gateway`. Direction is the point: requests flow one way, and queries can walk edges in *either* direction (`traverse in` vs `traverse out`) — which is exactly what turns "what does X call?" into "what depends on X?".

## Edges are a multiset, not a set

The same directed edge may appear **more than once**, and GSL preserves every occurrence:

```gsl
node gateway
node orders

gateway->orders [protocol="http"]
gateway->orders [protocol="http"]
```

Why does that matter? Because the duplication is sometimes *information* — two transports, two flows, two invocations. A multiset keeps your facts intact. When you want the collisions removed, queries collapse them for you (`collapse`), and set membership on an edge is the normal way to tag *a particular* occurrence rather than all of them.

## Edges carry the same luggage as nodes

An edge can have attributes, set membership, and nesting just like a node:

```gsl
set async

node checkout
node orders

checkout->orders [retries="3"] @async
```

`@async` marks *this edge* as a member of the `async` set — queryable the same way node membership is: `subgraph edge in @async`.

## Dependencies between edges

Here is where edges go beyond any node-centric model. An edge can be **labeled**, and a scoped (or `parent=`-declared) edge can state that it runs *only after* another edge — a relationship *between edges*:

```gsl
GATE: controlplane -> approve
PROMOTE: registry -> canary [parent=GATE]
```

Read that as *"`PROMOTE` cannot run until `GATE` has happened"* — while both remain first-class edges in the graph. This is how flagship [03 — release prerequisites](../../examples/flagships/03-release-prerequisites/README.md) models "when can we promote?" as pure graph. Promotions, approvals, gates — a whole class of workflow is just edges with edges depending on them. See [Parents, scopes and edge dependencies](parents-and-scopes.md).

## Why you can traverse both ways

Because edges are stored, not lost, every query can walk them in **both** directions without extra modelling:

- `subgraph node in @critical traverse out 1` — what critical things *call*.
- `subgraph node in @critical traverse in 1` — what critical things are *called by* (the dependency side).

The whole "what depends on X?" operation is a *direction* on the same stored edges.

---

**Next:** [Named sets](sets.md) — tagging things with a name and asking who is in it.