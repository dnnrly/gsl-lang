<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Modelling with GSL

**Audience:** people about to write their first real model.
**Prerequisite:** [Concepts](../concepts/graph-model.md) through [Queries and derived views](../concepts/queries-and-views.md).
**Next step:** the [query tutorial](../../QUERY_TUTORIAL.md), or a [cookbook recipe](../cookbook/README.md) to see a full question answered.

The syntax pages tell you what GSL accepts. This page is about the other question:

> What should I actually put into a GSL model?

There is no single right answer — modelling is judgement. But the judgement follows a small set of principles, and they all point in the same direction.

## 1. Start from the questions, not the format

The purpose of a model is to answer questions. Before you decide what goes in it, decide what you will ask of it:

> Model the things and relationships that matter to the questions you want to ask.

If the questions are "which services are `@critical`?" and "if we retire `orders`, who breaks?", then the model needs nodes for the services, a set marking what is critical, and edges for the dependencies. It does not need the CI pipeline's cache settings. Everything below is a way of asking, *for each fact, which layer it belongs in*.

## 2. What is a node?

A node is a *thing* your questions are about — a service, a database, a person, a decision, a component, a zone. The test: **if a fact has relationships of its own, it is a node**, because edges attach to nodes.

```gsl
node gateway [team="gateway"]
node orders [team="orders"]
node orders_db [team="platform"]
```

Why is `orders_db` a node and not an attribute of `orders`? Because it has relationships of its own — `orders` calls it, but it could also serve other services. Things that a question might want to stand alone, connect, or filter on are nodes.

## 3. What is an edge?

An edge represents a **relationship** between two nodes: a call, a dependency, a reporting line, a sequence step. Direction is real — `gateway -> orders` is not `orders -> gateway`.

```gsl
gateway -> orders [protocol="http"]
orders -> orders_db [protocol="sql"]
```

Two edges between the same pair of nodes are not a mistake; they can be **two distinct facts** (two transports, two flows). GSL keeps every occurrence — more on that in [§9](#9-preserving-information).

The decision rule: if your question is about the *connection* between two things — "what calls what?", "what depends on what?" — that connection is an edge, even when it looks as simple as a label you might have written in a diagram.

## 4. What is an attribute?

An attribute is a **fact about a single element**: `team`, `version`, `protocol`, `owner`. Attributes are untyped and schema-free — adding one is a text edit, not a type migration.

```gsl
node gateway [replicas="4", team="gateway"]
```

Attributes are the default home for facts that belong to exactly one node, edge or set and that you will mostly *read* rather than *group by*. When a fact starts deciding a lot of membership questions, it may graduate to a set (next).

## 5. What is a set?

A set is a **classification that cuts across many elements** — something you will want to pick out *as a group*.

> **A set is not a type.** Sets classify and group; they do not impose a hierarchy. `@critical` does not say what a node *is*, and a node can be in any number of sets.

```gsl
set critical

node gateway @critical
node orders @critical
node payments
```

Why is `@critical` a set and `team="payments"` an attribute? `team` describes one owner of one node; `critical` is a property many nodes share and that your questions group by ("which nodes are `@critical`?"). And membership is *data* — a service leaving `@critical` is a one-line change you can review in a diff, not a type change a compiler enforces. If a single node can belong to several groupings, a set is the honest home.

## 6. When to nest

Use nesting when the relationship really is **containment** — one thing lives inside another: a database inside a service, a zone inside an availability region. Containment gives you fully-qualified identity (`node "checkout/orders_db"`) and component boundaries in derived diagrams.

```gsl
node checkout {
    node orders_db
}
```

Do not use nesting as an organisational convenience ("all my team's services under a `my_team` node"). If the relationship is not containment, it should be an edge or a set, not a parent scope.

## 7. When to use edge dependencies

Use edge dependencies when the subject of your question is a **relationship between relationships**: an approval before a promotion, a gate before a deploy, one pipeline stage after another.

```gsl
CONFIG: config -> controlplane [stage="apply"]
GATE: controlplane -> approve [stage="gate", parent=CONFIG]
```

Here the fact being recorded is not "which node depends on which node" but "the `CONFIG` *edge* must complete before the `GATE` edge". If your workflow's questions are about ordering, this is the construct, not a hand-drawn arrow beside the graph. See [Edge dependencies](../concepts/edge-dependencies.md).

## 8. Match the detail to the questions

A model is most valuable when it is small enough to keep current. The question to ask about every candidate fact:

> Will this fact ever answer a useful question?

If not, leave it out. Adding facts is cheap later; a model that nobody can keep current rots faster than a model that is too small. The flagship's own limits are a good example of what *not* to model: no runbooks, no SLAs, no runtime traffic — just the structure that answers the architecture's questions ([flagship 01 — "What this example is not"](../../examples/flagships/01-service-many-views/README.md)).

Start minimal, and add a fact when a question actually needs it.

## 9. Preserving information

One principle ties the others together:

> GSL tries to preserve information rather than force it prematurely into a particular view.

Each construct protects a different kind of fact until you are ready to interpret it:

- **Multiset edges** preserve distinct relationship facts — two identical-looking edges stay two facts, rather than collapsing into one.
- **Untyped attributes** avoid forcing a rigid schema before you know what matters.
- **Sets** preserve cross-cutting classifications as data, separate from any type hierarchy.
- **Nesting** preserves structure.
- **Edge dependencies** preserve relationships between relationships.
- **Derived views** mean you never duplicate the source-of-truth information into a diagram that drifts.

When you are unsure, prefer to *keep* a fact in the model and let a view drop it, rather than force an interpretation early.

---

## A worked example

Let's build the model for the questions from [§1](#1-start-from-the-questions-not-the-format): *"which services are `@critical`?"*, *"if we retire `orders`, who breaks?"*, *"what does the payments team own?"*.

**Step 1 — the things.** Services and databases are things with relationships: nodes.

```gsl
node gateway
node orders
node payments
node orders_db
```

**Step 2 — the relationships.** Calls and dependencies are connections: edges.

```gsl
gateway -> orders [protocol="http"]
orders -> payments [protocol="grpc"]
orders -> orders_db [protocol="sql"]
```

**Step 3 — the facts about single things.** Who owns it, how it is configured: attributes.

```gsl
node gateway [team="gateway", replicas="4"]
node orders [team="orders"]
node payments [team="payments"]
node orders_db [team="platform"]
```

**Step 4 — the classifications.** Which services are load-bearing: a set, queried as data.

```gsl
set critical

node gateway @critical
node orders @critical
```

Now the questions answer themselves:

```bash
gsl-query 'subgraph node in @critical' < model.gsl
gsl-query 'subgraph node.id == "orders" traverse in 1' < model.gsl
gsl-query 'subgraph node.team == "payments"' < model.gsl
```

Each answer is a derived view. Nothing was duplicated, nothing was hand-drawn.

**Next:** learn the query language step by step in the [query tutorial](../../QUERY_TUTORIAL.md), or jump to a [cookbook recipe](../cookbook/README.md) for a small question answered end to end.