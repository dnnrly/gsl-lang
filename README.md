<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# GSL

> **The source of truth for the graphs behind your diagrams.**

Architecture diagrams go stale. Not because drawing is hard — because the picture becomes the thing everyone maintains, while the information it represents has no authoritative home. Change a service boundary and you edit *N* diagrams by hand; "what breaks?" gets answered by gut feel.

GSL makes the graph itself the source of truth — a canonical, diffable text format for graph-shaped knowledge (architectures, dependencies, workflows, organisational and relationship structures), kept in version control like any other source. Every diagram, subset, report or analysis is a *view* derived from that one graph:

> **One graph, many views.** The diagram is a view; the graph is the asset.

```text
graph in Git
   ├── architecture diagram
   ├── team view
   ├── blast radius
   ├── dependency report
   └── or consumed directly by other tooling
```

Renderers like Mermaid, D2 and Graphviz draw one view extremely well. GSL is for the different case: the graph is the durable asset, the views are many, and they are derived from one source — in sync by construction rather than by discipline. A diagram is one possible consumer of that graph, not the reason it has to exist: the same structure can feed a report, a CI check or a language model directly, with no picture involved.

> **Status: a serious experiment.** The implementation and specification are mature and heavily tested; the idea itself is young and unproven in the real world — whether this model genuinely helps people remains an open question. Evaluate it on its merits; [Project status](#project-status) is the honest detail.

---

## Try it in 30 seconds

Suppose you model a small retail platform. One file, `model.gsl`, is the whole story:

```gsl
set critical

node gateway [team="gateway"] @critical
node orders [team="orders"] @critical
node orders_db [text="PostgreSQL", team="platform"] @critical
node payments [team="payments"]
node users [team="identity"]

gateway -> orders [protocol="http"]
orders -> payments [protocol="grpc"]
orders -> users [protocol="grpc"]
orders -> orders_db [protocol="sql"]
```

Now ask the graph a question: *"If we retire `orders`, which `@critical` parts are affected?"*

```gql
(subgraph node.id == "orders" traverse in all) as BLAST | from * | (subgraph node in @critical) as CRIT | BLAST & CRIT
```

Read the pipeline left to right: `BLAST` grabs everything that transitively depends on `orders`; `CRIT` isolates the `@critical` nodes; the `&` keeps only what appears in both. The answer comes back as canonical, diffable GSL:

```gsl
set critical

node gateway [team="gateway"] @critical
node orders [team="orders"] @critical

gateway->orders [protocol="http"]
```

Retiring `orders` breaks the `@critical` gateway — a fact you derived from the model, not from gut feel. Re-run it any time the model changes; the answer stays in sync. If the diagram were the source of truth, that is a trail you'd trace by eye across every outdated picture in your wiki.

This is the whole idea. Everything below shows it at scale.

---

## Why GSL?

Every picture is an output. GSL makes the graph the first-class thing you keep: questions about your system become queries, and all the pictures you need are re-derived, not re-drawn:

- A component diagram for each audience
- The blast radius of retiring a service
- A per-team dependency view
- A migration backlog for deprecated components
- A text report consumed by other tooling

The same graph answers all of them. Nothing to hand-maintain.

Concretely, GSL gives you:

- **Truth you can diff.** Deterministic serialisation means small, reviewable diffs in git — a PR that renames a dependency is one line, not a picture you redraw.
- **Answers, not eyeballing.** Ask "what depends on this?" and get a canonical subgraph back, not a path traced by a human.
- **One source, many views.** Derive whatever subset, summary or diagram you need from a single graph — no duplicated structure, no drift.
- **Relationships between relationships.** Edges can declare dependencies on other edges (a promotion that waits on an approval), so workflow prerequisites are data, not a hand-drawn arrow.
- **Structure without a schema.** Carry arbitrary attributes (`team`, `protocol`, `owner`, `confidence`) and query on them. No schema step, no database to stand up.
- **Text, not a picture.** The structure is plain text: a PR can review it, CI can check it, other tooling can consume it — and when a language model needs to understand a system, the graph itself is what it reads, not an image of it.

Honest boundaries: GSL is a file format and a derivation tool. It does not do graph layout, schema validation, persistence or database-scale querying — and it does not try to. See [Project status](#project-status) and [Compared with alternatives](#gsl-compared-with-alternatives) for where that is a feature and where someone else is the better fit.

---

## What can I use it for?

| Job | What it looks like | Full example |
|---|---|---|
| **Keep architecture docs truthful** | One canonical model; diagrams, team views and impact answers derived from it | [01 — one graph, many views](examples/flagships/01-service-many-views/README.md) |
| **Reconstruct an undocumented system** | Record recovered facts with provenance (`source`, `confidence`) while you dig; risk lists become queries | [02 — architecture archaeology](examples/flagships/02-architecture-archaeology/README.md) |
| **Model workflow prerequisites** | A deploy edge that depends on an approval edge; query "what does the gate unlock?" | [03 — release prerequisites](examples/flagships/03-release-prerequisites/README.md) |
| **Model relationship networks of any kind** | Financial exposure, org structures, relationship graphs — the same operations, a different domain | [04 — financial relationships](examples/flagships/04-financial-network/README.md) |

GSL is a *graph* language. It is not software-only: the same text format, query language and derived views describe financial networks, organisational structures and any other relationship-shaped knowledge. Whatever the domain, the point is the same: the graph is the asset; the views are derived.

---

## Flagship examples

Six complete, runnable narratives — **problem → model → query → derived view** — each with a two-minute test you can run. They are the best place to understand why GSL exists, end to end. See the [flagships index](examples/flagships/README.md).

| Example | The question it answers | The distinctive idea |
|---|---|---|
| [01 — one graph, many views](examples/flagships/01-service-many-views/README.md) | "Retire this service — who breaks?" | One 20-service model, seven derived views (blast radius, team-level, critical reach, migration). |
| [02 — architecture archaeology](examples/flagships/02-architecture-archaeology/README.md) | "We inherited an undocumented monolith" | The model carries provenance; risk lists are queries, not guesses. |
| [03 — release prerequisites](examples/flagships/03-release-prerequisites/README.md) | "When can we promote?" | A relationship *between edges* — a promotion gated on an approval. Workflow as a graph. |
| [04 — financial relationships](examples/flagships/04-financial-network/README.md) | "If the CCP fails, who is exposed?" | The same operations as 01, in a non-software domain. |
| [05 — LLM-assisted modelling](examples/flagships/05-llm-assisted-modelling/README.md) | "Can an agent turn prose into a model?" | An honest experiment: parser + canonical diff + queries make agent output reviewable. |
| [06 — enterprise architecture archaeology](examples/flagships/06-enterprise-architecture-archaeology/README.md) | "What does our enterprise architecture actually look like?" | Independent investigation fragments from many repos compose into one queryable graph. |

[**01 — one graph, many views**](examples/flagships/01-service-many-views/README.md) is the hero example: a single `model.gsl` for a retail platform.

![The full architecture, derived from one graph](examples/flagships/01-service-many-views/views/full.component.svg)

![The critical blast radius of retiring the legacy order service](examples/flagships/01-service-many-views/views/critical-blast.graph.svg)

---

## Try it

You need **Go 1.26 or newer** to install from source. The reference implementation has no build-time magic — standard compiler, single step.

```bash
go install github.com/dnnrly/gsl-lang/cmd/gsl-query@latest
go install github.com/dnnrly/gsl-lang/cmd/gsl-diagram@latest
```

Prebuilt binaries for Linux, macOS and Windows are attached to the [GitHub releases](https://github.com/dnnrly/gsl-lang/releases) page (no Go required).

### Thirty seconds with a real example

Clone the repository (or copy the examples from it) and point the tools at a flagship model:

```bash
git clone https://github.com/dnnrly/gsl-lang
cd gsl-lang/examples/flagships/01-service-many-views

# 1. The graph, as durable text — canonicalised for stable git diffs
gsl-query "" < model.gsl

# 2. A derived view: "what breaks if we retire legacy-orders?"
gsl-query '(subgraph node.id == "legacy_orders" traverse in all) as BLAST | from * | (subgraph node in @critical) as CRIT | BLAST & CRIT' < model.gsl

# 3. A diagram of the full model (Mermaid component view)
gsl-query 'from *' < model.gsl | gsl-diagram -f mermaid -t component

# 4. The same graph as PlantUML source — a second diagram dialect
gsl-query 'from *' < model.gsl | gsl-diagram -f plantuml
```

Every command outputs **canonical GSL or diagram source** (`.mmd`/`.puml`) — stable, reviewable, diffable. `gsl-diagram` supports Mermaid (component, graph, sequence) and PlantUML (component, sequence); see [cmd/gsl-diagram/README.md](cmd/gsl-diagram/README.md) for the full converter reference. The output is diagram *source*; pipe it to your usual renderer (`mermaid-cli`, `plantuml`) when you want a picture.

### The two-minute test

The flagship examples are self-checked: every committed result is byte-compared against the real CLI on every test run.

```bash
go test ./examples -run Flagship -v
```

---

## GSL compared with alternatives

GSL is not "better" than these — it is *different*: model-first, render-later. Choose GSL when the graph is an asset you will version, query and reshape. Choose the alternative when it is the fastest path to the thing you need.

| Alternative | Choose *it* instead of GSL when… | Choose GSL instead when… |
|---|---|---|
| **Mermaid** | You want a picture in Markdown/GitHub with zero tooling, or a one-off throwaway diagram | The structure will be queried, reused or re-derived; you want the diagram *and* queryable data from one source |
| **D2** | Diagram *appearance/layout* is the success criterion and you want premium auto-layout | You need the structure as queryable data and layout is a derived view, not the product |
| **Graphviz / DOT** | You need serious auto-layout algorithms (dot/neato) or heavy rendering control | You want canonical ordering, sets, edge-dependency semantics, deterministic diffs, and a query layer DOT lacks |
| **JSON / YAML** | You need maximum interop with generic tools; your structure is small or relational, not graph-shaped | Your data is graph-shaped and you want graph semantics, sets, edge dependencies and querying without writing a custom model |
| **GraphML / GraphSON** | You must exchange graphs between specific tools/ecosystems | You need a human-authored, versionable, queryable graph you control end to end |
| **Graph databases (Neo4j, …)** | You need persistence, indexing, live transactions or thousands-scale querying | The graph lives in a repository as code; queries are file-level deterministic transformations |
| **Structurizr / C4 DSL** | You have committed to C4/Structurizr as your architecture standard | You model graphs beyond C4 (workflows, task dependencies, financial networks) or want a C4-agnostic format |

The converters mean these compose rather than compete: a GSL graph can *become* a Mermaid or PlantUML view. The composition is the point.

---

## Learn GSL

Ordered for learning, not for reference completeness:

1. **[Documentation](docs/README.md)** — the learning journey: Getting Started → Concepts → Tutorials → Cookbook, with the guides and specifications as reference tiers.
2. **[Flagship examples](examples/flagships/README.md)** — why GSL exists, problem-first; the narratives the documentation links back to.
3. **[GSL Guide](GSL_GUIDE.md)** — the language in one self-contained document (syntax, semantics, design notes).
4. **[Query tutorial](QUERY_TUTORIAL.md)** — a step-by-step learning path for GQL, the query language.
5. **[GQL Guide](GQL_GUIDE.md)** — GQL as a self-contained reference.
6. **[Go reference](GO_REFERENCE.md)** — the Go API and algorithm patterns, for programmatic use.
7. **[Examples](examples/README.md)** — a catalog of graphs demonstrating individual language features.

Targeted at AI and LLM tooling: start at **[llms.txt](llms.txt)** for the agent-oriented index.

---

## Specification

GSL is defined by a normative, RFC 2119-style specification, not by the implementation:

- **[SPEC.md](SPEC.md)** — the authoritative language specification (v1.0.0 Draft): grammar, semantics, canonicalisation guarantee.
- **[GRAMMAR.md](GRAMMAR.md)** — the formal grammar, for implementing a parser.
- **[QUERY_SPEC.md](QUERY_SPEC.md)** — the query language specification (v0.4.0 Revised Draft).
- **[QUERY_GRAMMAR.md](QUERY_GRAMMAR.md)** — the formal GQL grammar.

Every `gsl` and `gql` code block in this repository's markdown — at the root and under [`docs/`](docs/README.md) — is automatically parsed on test. The documentation cannot drift from the language it describes.

---

## Implementations

- **Go library** — the reference implementation; a hand-written parser with a canonical-form guarantee (`parse(serialize(parse(x))) == parse(x)`), standard-library core only. `go get github.com/dnnrly/gsl-lang`.
- **`gsl-query`** — run GQL pipelines (subgraph, traverse, make, remove, collapse, graph algebra) against a GSL graph; emits canonical GSL.
- **`gsl-diagram`** — render any GSL document (or derived view) to Mermaid or PlantUML.
- **`gsl-lsp`** — a language server for GSL and GQL (completion, hover, diagnostics, formatting). Source lives in [`lsp/`](lsp/); a preliminary VS Code extension is in [`editors/vscode/`](editors/vscode/).

The CLI tools are Unix-composable: GSL in, canonical GSL or a diagram out, via stdin/stdout.

---

## Project status

**The implementation is mature; the adoption question is open.** The specification and reference implementation are small, disciplined and heavily tested — every claim here is enforced by tests, not by description — but this is an honest experiment, not an established standard. What remains genuinely open is whether the model proves useful to people beyond this repository.

| Element | Maturity | Basis |
|---|---|---|
| **Language & specification** | High | v1.0.0 Draft, RFC 2119 style, canonicalisation guarantee enforced by round-trip and fuzz tests |
| **Go implementation** | High | Standard-library core, 68.8% statement coverage (measured by `make test`), 9 fuzz targets, acceptance tests |
| **Query language (GQL)** | Medium | Large, well-tested surface — but a *Revised Draft* spec, unproven with real users |
| **Tooling** | Medium | `gsl-query` and `gsl-diagram` work and are documented; the LSP and VS Code extension are early |
| **Ecosystem** | Low | One reference implementation, no third-party integrations yet |

What GSL deliberately is not — and why that is by design:

- **Not a layout engine.** Diagram output exists; appearance is a view, not the product. Render through Mermaid, PlantUML and others.
- **Not a schema or validation framework.** Graphs are accepted as written, with non-fatal warnings. Staying small is the point.
- **Not a database.** No persistence, indexing or live multi-user querying. Queries are file-in, file-out transformations.
- **Not a general data format.** GSL is for graph-shaped knowledge only.

The closest thing to a gotcha: GQL is the differentiator, and it is honestly labelled a Revised Draft. The flagship examples only use behaviour covered by its tested fixtures.

---

## Contributing

The parser is hand-written, the core is standard-library-only, and the tests run against the language itself. Contribution guidance and project conventions live in [AGENTS.md](AGENTS.md). The [code of conduct](CODE_OF_CONDUCT.md) applies.

---

## Licensing

The original GSL work is authored and maintained by **Pascal Dennerly**, who holds its copyright. The repository is intentionally dual-licensed so that code and documentation can be reused under the terms that best fit each:

| Content | Licence |
|---|---|
| **Implementation/source code** — Go sources, CLI tools, LSP server, VS Code extension, build and configuration files | [Apache-2.0](LICENSE) |
| **GSL specification** — `SPEC.md`, `GRAMMAR.md`, `QUERY_SPEC.md`, `QUERY_GRAMMAR.md` | [CC-BY-4.0](LICENSE-CC-BY-4.0) |
| **Technical documentation/tutorials** — `docs/`, the guides, READMEs, contribution docs | [CC-BY-4.0](LICENSE-CC-BY-4.0) |
| **Code examples** — `.gsl`/`.gql` files under `examples/` and the test fixtures, and fenced code blocks shown in the documentation | [Apache-2.0](LICENSE) |

Individual files carry an SPDX header identifying their licence; where a document is CC-BY-4.0, **fenced code blocks embedded in it remain Apache-2.0**. Reuse the prose under CC-BY-4.0 and the code under Apache-2.0. The CLI tools report their licence in `version` output, and the embedded guides ship under CC-BY-4.0 with the binaries.

These licences cover the material published in this repository. They grant permission to use that material; they do not claim ownership of, or restrict independent implementations of, the GSL language itself.