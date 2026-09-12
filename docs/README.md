# GSL Documentation — The Learning Journey

> **The source of truth for the graphs behind your diagrams.**

GSL is a canonical, diffable text format for modelling graph-shaped knowledge — architectures, dependencies, workflows, organisational and relationship structures — from which you derive the views you need. **One graph, many views**: the graph file is the durable asset; every diagram, subset, report or analysis is a derived view.

This index is the route from *first contact* to *competent use*. Each tier states its audience, its prerequisite, and what to read next — you should never need to open the formal specification until you choose to.

---

## The journey at a glance

| Tier | What you get | Prerequisite | Next step |
|---|---|---|---|
| **[Getting Started](getting-started/README.md)** | Tools installed and your first query + diagram in two minutes | Nothing | Concepts |
| **[Concepts](concepts/graph-model.md)** | The graph model in plain words — no grammar required | Getting Started | Tutorials |
| **[Tutorials](tutorials/README.md)** | Goal-oriented paths: the flagship examples and the GQL learning path | Concepts | Cookbook |
| **[Cookbook](cookbook/README.md)** | Short recipes: problem → GSL → answer → diagram | Concepts + basic queries | Language Guide / Query Guide |
| **[Language Guide](../GSL_GUIDE.md)** | The language, systematically, in one document | Concepts + Cookbook | Reference |
| **[Query Guide](../GQL_GUIDE.md)** · **[Query tutorial](../QUERY_TUTORIAL.md)** | GQL, the query/transformation language, as reference + a step-by-step learning path | Concepts + Tutorials | Reference / Specification |
| **[Reference](../GRAMMAR.md)** · **[Query grammar](../QUERY_GRAMMAR.md)** · **[Go reference](../GO_REFERENCE.md)** · **[Sequence guide](../SEQUENCE_GUIDE.md)** | Precise syntax and behavioural detail, including the Go API and the sequence-diagram dialect | When you need precision | Specification |
| **[Specification](../SPEC.md)** · **[Query spec](../QUERY_SPEC.md)** | The normative, RFC 2119-style definition of the language — implementation-facing, never diluted | When you must be exact | — |
| **[Ecosystem](../README.md#implementations)** | `gsl-query`, `gsl-diagram`, `gsl-lsp` and the VS Code extension | Whenever | — |

## How to read these tiers

- **The journey is circular, not linear.** Concepts cross-reference each other; the Cookbook points back at the flagship examples; the Reference and Specification are there the moment you need *precision* rather than an explanation.
- **The flagship examples are the spine.** Every tier links to the five runnable narratives under `examples/flagships/` — they are the same story told end to end: *problem → model → query → derived view*.
- **GQL (the query language) is deliberately a post-Concepts topic.** You can be productive with just a graph file and `gsl-query ""`; queries are how you derive the views.
- **Honest about maturity.** The core language is stable and heavily tested. GQL is an extensive, well-tested *Revised Draft* — the flagship examples only use behaviour covered by the tested query fixtures.

## Tiers in more detail

### Getting Started

Install the tools, run the flagship two-minute test, ask your first query, render your first diagram, and learn where to turn for the *why*. **[Start here](getting-started/README.md)** — no prerequisites.

### Concepts

The mental model before the syntax: what a graph is in GSL, how edges behave as a multiset, what named sets and untyped attributes give you, how parent scopes and edge dependencies work, how parsing stays lenient, and what canonical form means for version control. Seven pages, one concept each, linked rather than repeated:

1. [Graphs and nodes](concepts/graph-model.md)
2. [Edges as a multiset](concepts/edges.md)
3. [Named sets](concepts/sets.md)
4. [Untyped attributes](concepts/attributes.md)
5. [Parents, scopes and edge dependencies](concepts/parents-and-scopes.md)
6. [Lenient parsing and last-write-wins](concepts/parsing-and-merging.md)
7. [Canonical form and diffability](concepts/canonical-form.md)

### Tutorials

Two goal-oriented paths, both executable:

- **The flagship examples** — five complete narratives for real problems: [the flagship index](../examples/flagships/README.md).
- **The query tutorial** — a step-by-step GQL learning path: [QUERY_TUTORIAL](../QUERY_TUTORIAL.md).

See the [tutorials index](tutorials/README.md) for how the paths fit together.

### Cookbook

Recipes for specific jobs — *"retire a service — who breaks?"*, *"what's in `@critical`?"*, *"regenerate the diagrams in CI"* — each showing the problem, the model, the query, the exact answer, and the diagram command. See the [cookbook index](cookbook/README.md).

### Language Guide, Query Guide, Reference, Specification

These tiers live at the repository root and stay there:

- **Language Guide** → [GSL_GUIDE.md](../GSL_GUIDE.md)
- **Query Guide** → [GQL_GUIDE.md](../GQL_GUIDE.md) and the [query tutorial](../QUERY_TUTORIAL.md)
- **Reference** → [GRAMMAR.md](../GRAMMAR.md), [QUERY_GRAMMAR.md](../QUERY_GRAMMAR.md), [GO_REFERENCE.md](../GO_REFERENCE.md), [SEQUENCE_GUIDE.md](../SEQUENCE_GUIDE.md)
- **Specification** → [SPEC.md](../SPEC.md), [QUERY_SPEC.md](../QUERY_SPEC.md)

### Ecosystem

The tools that consume and emit GSL — [implementations](../README.md#implementations) in the README, the [gsl-query](../cmd/gsl-query/README.md) and [gsl-diagram](../cmd/gsl-diagram/README.md) command references, the [LSP server](../lsp/), and the [VS Code extension](../editors/vscode/).

### For AI agents and LLMs

The agent-oriented index and embedded guides sit at the repository root: start at [llms.txt](../llms.txt), then [GSL_GUIDE](../GSL_GUIDE.md), [GQL_GUIDE](../GQL_GUIDE.md) and [GO_REFERENCE](../GO_REFERENCE.md). The [LLM-assisted modelling flagship](../examples/flagships/05-llm-assisted-modelling/README.md) is the worked experiment.