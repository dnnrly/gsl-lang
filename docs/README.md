<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# GSL Documentation — The Learning Journey

> **The source of truth for the graphs behind your diagrams.**

GSL is a canonical, diffable text format for modelling graph-shaped knowledge — architectures, dependencies, workflows, organisational and relationship structures — from which you derive the views you need. **One graph, many views**: the graph file is the durable asset; every diagram, subset, report or analysis is a derived view.

The two languages have separate jobs: **GSL is the graph language** — it describes the graph and its facts; **GQL is the query language** — it reads and transforms graphs to produce derived graphs. This journey teaches GSL first, then introduces GQL when you have a graph and something to ask of it.

This index is the route from *first contact* to *competent use*. Each tier states its audience, its prerequisite, and what to read next — you should never need to open the formal specification until you choose to.

---

## The journey at a glance

| Tier | What you get | Prerequisite | Next step |
|---|---|---|---|
| **[Getting Started](getting-started/README.md)** | Tools installed and your first query + diagram in two minutes | Nothing | Concepts |
| **[Concepts](concepts/graph-model.md)** | The mental model in plain words — graph, nodes, edges, attributes, sets, structure, queries — no grammar required | Getting Started | Modelling |
| **[Modelling with GSL](tutorials/modelling-with-gsl.md)** | How to decide what goes in a model: node, edge, attribute, set, nested, or edge dependency? | Concepts (through queries) | Tutorials |
| **[Tutorials](tutorials/README.md)** | Goal-oriented paths: the flagship examples, the GQL learning path, and the modelling guide | Concepts + Modelling | Cookbook |
| **[Cookbook](cookbook/README.md)** | Short recipes: problem → GSL → answer → diagram | Concepts + basic queries | Language Guide / Query Guide |
| **[Language Guide](../GSL_GUIDE.md)** | The language, systematically, in one document | Concepts + Cookbook | Reference |
| **[Query Guide](../GQL_GUIDE.md)** · **[Query tutorial](../QUERY_TUTORIAL.md)** | GQL, the query/transformation language, as reference + a step-by-step learning path | Concepts + Tutorials | Reference / Specification |
| **[Reference](../GRAMMAR.md)** · **[Query grammar](../QUERY_GRAMMAR.md)** · **[Go reference](../GO_REFERENCE.md)** · **[Sequence guide](../SEQUENCE_GUIDE.md)** | Precise syntax and behavioural detail, including the Go API and the sequence-diagram dialect | When you need precision | Specification |
| **[Specification](../SPEC.md)** · **[Query spec](../QUERY_SPEC.md)** | The normative, RFC 2119-style definition of the language — implementation-facing, never diluted | When you must be exact | — |
| **[Ecosystem](../README.md#implementations)** | `gsl-query`, `gsl-diagram`, `gsl-lsp` and the VS Code extension | Whenever | — |

## How to read these tiers

- **The journey is circular, not linear.** Concepts cross-reference each other; the Cookbook points back at the flagship examples; the Reference and Specification are there the moment you need *precision* rather than an explanation.
- **The flagship examples are the spine.** Every tier links to the six runnable narratives under `examples/flagships/` — they are the same story told end to end: *problem → model → query → derived view*.
- **GQL is deliberately a post-Concepts topic.** You can be productive with just a graph file and `gsl-query ""`; the concept page [Queries and derived views](concepts/queries-and-views.md) explains where queries fit before you meet their grammar.
- **The advanced concept pages can wait.** Parsing rules and canonical form explain *why text graphs behave well in git*; they are not prerequisites for modelling or querying.
- **Honest about maturity.** The core language is stable and heavily tested. GQL is an extensive, well-tested *Revised Draft* — the flagship examples only use behaviour covered by the tested query fixtures.

## Tiers in more detail

### Getting Started

Install the tools, run the flagship two-minute test, ask your first query, render your first diagram, and learn where to turn for the *why*. **[Start here](getting-started/README.md)** — no prerequisites.

### Concepts

The mental model before the syntax: what a graph is in GSL, how edges represent relationships, what untyped attributes and named sets give you, how structure and edge dependencies work, where queries and derived views fit, and what leniency and canonical form mean for version control. Nine pages, one concept each, linked rather than repeated; the last two are advanced and can wait until you need them:

1. [Graphs and nodes](concepts/graph-model.md)
2. [Edges — relationships between nodes](concepts/edges.md)
3. [Untyped attributes](concepts/attributes.md)
4. [Named sets](concepts/sets.md)
5. [Structure and nesting](concepts/structure-and-nesting.md)
6. [Edge dependencies](concepts/edge-dependencies.md)
7. [Queries and derived views](concepts/queries-and-views.md)
8. [Lenient parsing and last-write-wins](concepts/parsing-and-merging.md) *(advanced — defer until you need it)*
9. [Canonical form and diffability](concepts/canonical-form.md) *(advanced — defer until you need it)*

### Tutorials

Three goal-oriented paths, all executable:

- **Modelling with GSL** — deciding what belongs in a model (node, edge, attribute, set, nesting, edge dependency): [the modelling guide](tutorials/modelling-with-gsl.md).
- **The flagship examples** — six complete narratives for real problems: [the flagship index](../examples/flagships/README.md).
- **The query tutorial** — a step-by-step GQL learning path: [QUERY_TUTORIAL](../QUERY_TUTORIAL.md).

See the [tutorials index](tutorials/README.md) for how the paths fit together.

### Cookbook

Recipes for specific jobs — *"retire a service — who breaks?"*, *"what's in `@critical`?"*, *"regenerate the diagrams in CI"* — each showing the problem, the model, the query, the exact answer, and the diagram command. Recipes 1–4 use only the core language, so they are reachable straight after Getting Started. See the [cookbook index](cookbook/README.md).

### Language Guide, Query Guide, Reference, Specification

These tiers live at the repository root and stay there:

- **Language Guide** → [GSL_GUIDE.md](../GSL_GUIDE.md)
- **Query Guide** → [GQL_GUIDE.md](../GQL_GUIDE.md) and the [query tutorial](../QUERY_TUTORIAL.md)
- **Reference** → [GRAMMAR.md](../GRAMMAR.md), [QUERY_GRAMMAR.md](../QUERY_GRAMMAR.md), [GO_REFERENCE.md](../GO_REFERENCE.md), [SEQUENCE_GUIDE.md](../SEQUENCE_GUIDE.md)
- **Specification** → [SPEC.md](../SPEC.md), [QUERY_SPEC.md](../QUERY_SPEC.md)

### Ecosystem

The tools that consume and emit GSL — [implementations](../README.md#implementations) in the README, the [gsl-query](../cmd/gsl-query/README.md) and [gsl-diagram](../cmd/gsl-diagram/README.md) command references, the [LSP server](../lsp/), and the [VS Code extension](../editors/vscode/).

### For AI agents and LLMs

The agent-oriented index and embedded guides sit at the repository root: start at [llms.txt](../llms.txt), then [GSL_GUIDE](../GSL_GUIDE.md), [GQL_GUIDE](../GQL_GUIDE.md) and [GO_REFERENCE](../GO_REFERENCE.md). The [LLM-assisted modelling flagship](../examples/flagships/05-llm-assisted-modelling/README.md) is the worked experiment. The [enterprise architecture archaeology flagship](../examples/flagships/06-enterprise-architecture-archaeology/README.md) shows how that approach scales across many repositories. The human learning path above does not require any LLM mental model.