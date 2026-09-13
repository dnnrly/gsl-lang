<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Concept: Named Sets

**Audience:** everyone. **Prerequisite:** [Untyped attributes](attributes.md).
**Next step:** [Structure and nesting](structure-and-nesting.md).

A **set** is a named collection of nodes (and edges). It is GSL's grouping mechanism — and because membership is stored as first-class data, a set is something you can **query**, not just something you read.

> **A set is not a type.** A set classifies and groups graph elements; it does not impose a type hierarchy. `@critical` does not say what a node *is*, and a node can be a member of any number of sets. Membership is data — it can change, and the change is a reviewable diff.

Sets are also **independent of each other**: nothing forces overlapping sets to be related, so a node can carry several simultaneous classifications — its environment, geography, trust boundary, data sensitivity and owner — without any one of them being imposed as "the" grouping. Since dimensions stay independent, any *combination* is queryable and no derived hierarchy needs maintaining. Flagship [03 — architecture across boundaries](../../examples/flagships/03-architecture-across-boundaries/README.md) is a full worked example of five orthogonal dimensions held on one flat graph; the query answers double as the proof that no dimension is derived from another.

## Declaring a set and membership

```gsl
set critical

node gateway [team="gateway"] @critical
node payments [team="payments"]
```

Two different statements are at work:

- `set critical` — *declare a named set.*
- `@critical` on a node — *claim membership in it.*

`payments` is in the model but not in the set. This separation is the point: a diagram of the whole system and a diagram of "just the critical bits" are two **views of the same file**, not two documents to keep in sync.

## Sets have attributes too

A set works like a node for its own metadata — `set critical [description="Systemically important"]` — so the category itself can carry explanation without polluting every member.

## Querying membership

Membership becomes the vocabulary of the questions you can't easily eyeball:

```bash
gsl-query 'subgraph node in @critical' < model.gsl
```

Returns the canonical subgraph of exactly the critical nodes and the edges *between* them. Compare the cheap, mechanical, always-current answer above with the alternative — a human re-reading a diagram to decide which parts matter.

## Sets as the vocabulary of review

What you name says what you review. Common patterns in the flagship examples:

- `@critical` — systemically important.
- `@deprecated` — scheduled for retirement (drives migration backlogs).
- `@external` — outside your boundary.
- `@async` — applied to *edges* to tag asynchronous calls.

A code review of a GSL change is literally a diff of set membership — "this deployment moved two services out of `@critical`" — which is the asset-behaviour the whole format is built for.

---

**Next:** [Structure and nesting](structure-and-nesting.md) — containment, and the parent relationship.