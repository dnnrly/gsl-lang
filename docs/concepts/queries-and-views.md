# Concept: Queries and Derived Views

**Audience:** everyone ready to ask questions of a graph. **Prerequisite:** [Edge dependencies](edge-dependencies.md) and the pages before it.
**Next step:** [Modelling with GSL](../tutorials/modelling-with-gsl.md), then the [query tutorial](../../QUERY_TUTORIAL.md).

So far you have met two kinds of thing in a GSL file: **facts** (nodes, edges, structure) and **metadata about those facts** (attributes, sets). This page introduces the other two layers of the mental model, and the relationship between GSL's two languages.

## GSL is the graph language. GQL is the query language.

| | GSL | GQL |
|---|---|---|
| What it is | The graph language — the file format that **describes** a graph and its facts | The query language — commands that **read and transform** a graph to produce new graphs |
| You write it as | `node`, `edge`, `set`, attributes, nesting | `subgraph`, `traverse`, `make`, `remove`, `collapse`, graph algebra |
| It answers | "What is the structure?" | "What is the answer to my question?" |

The distinction is the load-bearing one: GSL is the durable asset; GQL is how you derive different views from it. They are two languages with two jobs, usually joined in one pipeline (`gsl-query` takes a GSL graph and a GQL query on stdin).

## The mental model in four layers

1. **The graph / facts** — nodes, edges, structure. This is what GSL describes, and it is the source of truth.
2. **Metadata about those facts** — attributes and sets: what you know *about* nodes, edges and groups.
3. **Operations on the graph** — GQL: selection, traversal, combining and transforming graphs to produce *derived graphs*.
4. **Views derived from the graph** — diagrams, reports, focused subgraphs: any useful representation obtained by operating on the graph, never hand-maintained alongside it.

Everything in GSL is one of these four things. When you read "one graph, many views", it means: **keep layer 1 (and its metadata) as the durable, version-controlled asset; obtain every layer-4 view through layer-3 operations.**

## A view, end to end

The small model from [Getting Started](../getting-started/README.md):

```gsl
set critical

node gateway [team="gateway"] @critical
node orders [team="orders"] @critical
node payments [team="payments"]

gateway->orders [protocol="http"]
orders->payments [protocol="grpc"]
```

A question ("which nodes are `@critical`?") is a GQL operation:

```bash
gsl-query 'subgraph node in @critical' < model.gsl
```

The answer is a new graph — a **derived view**, returned as canonical, diffable GSL rather than a picture. Render it and it becomes a diagram:

```bash
gsl-query 'subgraph node in @critical' < model.gsl | gsl-diagram -f mermaid -t component
```

Same graph, two views, both derived. Change the model and re-run the pipeline: the views stay honest, because they are never maintained by hand.

## Where the mechanics live

This page is the framing, not the reference. The step-by-step learning path is the [query tutorial](../../QUERY_TUTORIAL.md), the complete reference is the [GQL guide](../../GQL_GUIDE.md), and the [cookbook](../cookbook/README.md) shows small questions answered end to end. The flagship examples — especially [01 — one graph, many views](../../examples/flagships/01-service-many-views/README.md) — show the same pattern at 20-node scale.

---

That completes the core mental model. From here, start [Modelling with GSL](../tutorials/modelling-with-gsl.md) or jump straight into the [query tutorial](../../QUERY_TUTORIAL.md). If you want the full picture before doing anything, the two remaining concept pages explain the behaviours that make a text graph a version-controlled asset: [lenient parsing and last-write-wins](parsing-and-merging.md), then [canonical form](canonical-form.md).

**Next:** [Modelling with GSL](../tutorials/modelling-with-gsl.md) — what to put in a model in the first place.