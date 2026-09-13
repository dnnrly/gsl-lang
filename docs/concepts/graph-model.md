<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Concept: Graphs and Nodes

**Audience:** everyone. **Prerequisite:** [Getting Started](../getting-started/README.md).
**Next step:** [Edges — relationships between nodes](edges.md).

A GSL file describes a **graph**: a collection of **nodes** connected by **directed edges**. That's the entire data model, and it's deliberately small.

## What a graph is in GSL

A GSL document is a flat, ordered sequence of declarations — sets, nodes, edges, nested scopes. There is no `graph` header keyword: the file *is* the graph, and its identity is the path in version control.

```gsl
set critical

node gateway [team="gateway"] @critical
node orders [team="orders"] @critical
node payments [team="payments"]

gateway->orders [protocol="http"]
orders->payments [protocol="grpc"]
```

There are two basic statements to notice:

- `node gateway [...]` — *place a node.*
- `gateway->orders [...]` — *connect two nodes with a directed arrow.*

That is the entire data model. The rest of GSL's power comes from what you can **do** with this tiny structure: query it, filter it, derive diagrams from it, and merge it cleanly because it is canonical text.

> **Background:** this is what graph theory calls a *graph* — a collection of **vertices** (we call them *nodes*) connected by **edges**. If you have drawn a network topology or an architecture diagram, you already know the shape; there is no deeper theory you need to use GSL. ([Graph — discrete mathematics](https://en.wikipedia.org/wiki/Graph_(discrete_mathematics)))

## Nodes are the subjects

Nodes are *things*: services, people, databases, components, decisions, permissions. Anything your graph's story is about. A node can carry:

- **attributes** — untyped `key="value"` pairs describing it (see [Attributes](attributes.md)),
- **set membership** — `@critical` marks it as part of a named set (see [Named sets](sets.md)),
- **a parent scope** — nesting it inside another node (see [Structure and nesting](structure-and-nesting.md)).

Here the actual syntax is enough — a node is:

```gsl
node me [colour="blue"] @critical
```

Identifiers are their own thing in GSL: `node gateway` names a node `gateway`, while quoted `node "checkout service" []` records a *label*. If the identifier reads cleanly unquoted, GSL treats it as an ID; if it needs spaces or symbols, quote it and it shows up in output as a label.

## One graph, many views

The graph file is the **asset**; every diagram, blast-radius answer, team view and report is a **derived view**. Because the source is text, stakeholders read it directly; because the tooling serialises it canonically, it diffs cleanly in version control — the review is the model itself, not a screenshot of a diagram.

You already saw the loop in Getting Started:

- **Ask**: `gsl-query '(subgraph node.id == "orders" traverse in all) as BLAST | from * | (subgraph node in @critical) as CRIT | BLAST & CRIT' < model.gsl` — which `@critical` nodes depend on `orders`?
- **Render**: `gsl-query 'from *' < model.gsl | gsl-diagram -f mermaid -t component` — the whole model as a component diagram.

Same graph, two very different views — plus every committed flagship diagram, which are re-derived by the same pipeline rather than hand-maintained.

## What GSL deliberately does not do

Honest limits, all of which are *features*:

- **No schema validation.** GSL stores structure. It will not refuse a graph with a cycle or a node that is simultaneously a parent — use the tools, and check the patterns that matter to you.
- **No layout or rendering.** gsl-diagram asks the target format to place things. The graph knows *what* is connected, not *where* it goes.
- **No home for your data model.** Graphs are graphs: if there is a *type system* in your organisation's sense, that is your application's job.

These limits are what keep the format small: canonical, diffable, mergeable — the properties that make text formats useful as source-of-truth assets.

## The shapes this supports

Because the model is just nodes and directed edges, one language covers many problem shapes — service architecture, dependency graphs, workflows, organisational reporting, security boundaries, relationship networks. The flagship examples explore real ones: [service retirement](../../examples/flagships/01-service-many-views/README.md), [financial-network criticality](../../examples/flagships/04-financial-network/README.md), [LLM-assisted modelling](../../examples/flagships/05-llm-assisted-modelling/README.md), and [independent classification dimensions](../../examples/flagships/03-architecture-across-boundaries/README.md).

---

**Next:** [Edges — relationships between nodes](edges.md) — the relationship that makes a graph a *graph*, and how duplicate relationships can be distinct facts.