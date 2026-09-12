# Getting Started with GSL

**Audience:** anyone who wants to try GSL now.
**Prerequisite:** a terminal. **Time:** about two minutes to the first diagram.
**Next step:** [Concepts — the graph model](../concepts/graph-model.md).

By the end of this page you will have:

- the `gsl-query` and `gsl-diagram` tools installed,
- asked a real world graph a question and got a canonical answer,
- rendered a diagram derived from that graph,
- and a pointer to the concepts that explain *what you just did*.

---

## 1. Install the tools

You need **Go 1.26 or newer** to install from source. The reference implementation has no build-time magic — standard compiler, single step:

```bash
go install github.com/dnnrly/gsl-lang/cmd/gsl-query@latest
go install github.com/dnnrly/gsl-lang/cmd/gsl-diagram@latest
```

Prebuilt binaries for Linux, macOS and Windows are attached to the [GitHub releases](https://github.com/dnnrly/gsl-lang/releases) page (no Go required).

If you are working inside this repository, `make build` compiles both tools into `tmp/`, and `go test ./examples -run Flagship` verifies every committed example result end-to-end.

> **A note on ordering.** GSL output is *canonical* — deterministic within a given implementation and version, so diffs stay small and reviewable. The exact line ordering can vary between implementations and versions and still be valid canonical GSL. The committed flagship results in this repository were produced by the current reference implementation; building the tools from this source reproduces them byte-for-byte.

---

## 2. The two-minute test

The flagship examples are self-checked narratives — *problem → model → query → derived view*. Let's run the hero one, a 20-service retail platform:

```bash
git clone https://github.com/dnnrly/gsl-lang
cd gsl-lang/examples/flagships/01-service-many-views
```

Ask the graph a question: *"If we retire the legacy order service, which `@critical` parts break?"*

```bash
gsl-query -f q3-critical-blast.gql -i model.gsl
```

The answer is a derived subgraph — canonical, diffable GSL:

```gsl
set critical
set deprecated
set external

node gateway [replicas="4", team="gateway", zone="A"] @critical
node orders [replicas="6", team="orders", zone="B"] @critical

gateway->orders [protocol="http"]
```

Now render a view of the whole model as a Mermaid component diagram:

```bash
gsl-query 'from *' < model.gsl | gsl-diagram -f mermaid -t component
```

That is the whole loop: **one source graph → a query → a useful view**. Every committed diagram in the example is derived this way, never hand-maintained.

---

## 3. Your first query

You don't need a 20-service model to get the idea. Save this tiny graph to `model.gsl`:

```gsl
set critical

node gateway [team="gateway"] @critical
node orders [team="orders"] @critical
node orders_db [team="platform"] @critical
node payments [team="payments"]
node users [team="identity"]

gateway -> orders [protocol="http"]
orders -> payments [protocol="grpc"]
orders -> users [protocol="grpc"]
orders -> orders_db [protocol="sql"]
```

Two things to notice: `set critical` declares a named set, and `@critical` puts three nodes *in* it. That set is now queryable — "which nodes are `@critical`?":

```bash
gsl-query 'subgraph node in @critical' < model.gsl
```

```gsl
set critical

node gateway [team="gateway"] @critical
node orders [team="orders"] @critical
node orders_db [team="platform"] @critical

gateway->orders [protocol="http"]
orders->orders_db [protocol="sql"]
```

Or the blast-radius question — "if `orders` goes away, which `@critical` nodes depend on it?":

```bash
gsl-query '(subgraph node.id == "orders" traverse in all) as BLAST | from * | (subgraph node in @critical) as CRIT | BLAST & CRIT' < model.gsl
```

```gsl
set critical

node gateway [team="gateway"] @critical
node orders [team="orders"] @critical

gateway->orders [protocol="http"]
```

Reading the query left to right: *find everything that (transitively) depends on `orders`*, *take the `@critical` nodes*, *intersect them*. Three short concepts — subgraph, traversal, intersection — give you a structural answer you would otherwise eyeball.

> **Tip:** run `gsl-query "" < model.gsl` — the empty query canonises the graph. It strips comments, sorts what the specification sorts, and emits stable deterministic GSL that produces small, reviewable git diffs.

---

## 4. Your first diagram

Any graph — the full model or a derived subgraph — can become a Mermaid or PlantUML view:

```bash
gsl-query 'from *' < model.gsl | gsl-diagram -f mermaid -t component > model.mmd
gsl-query 'from *' < model.gsl | gsl-diagram -f plantuml > model.puml
```

Render the `.mmd` in any Mermaid viewer (GitHub renders it natively in Markdown). The diagram is a *view*: change the model, re-run the pipeline, and the diagram updates with it.

---

## 5. "What was that?" — where each piece lives

You just touched the whole mental model without realising it:

| You used… | That is… | Explained in |
|---|---|---|
| `node … [team="…"]` | an untyped attribute on a node | [Attributes](../concepts/attributes.md) |
| `set critical` / `@critical` | a named set and membership syntax | [Named sets](../concepts/sets.md) |
| `gateway -> orders` | a directed edge | [Edges as a multiset](../concepts/edges.md) |
| `subgraph` / `traverse` | selecting and expanding a subgraph | [Graphs and nodes](../concepts/graph-model.md) |
| `(…) as BLAST` / `&` | named graphs and graph algebra | [Query tutorial](../../QUERY_TUTORIAL.md) |
| `gsl-query ""` | canonicalisation | [Canonical form](../concepts/canonical-form.md) |
| the canonical result | "one graph, many views" in miniature | [Graphs and nodes](../concepts/graph-model.md) |

Guessing at `@critical` membership or drawing a blast-radius arrow by hand is exactly the rot GSL exists to remove: the model holds the facts, and the views — diagrams, subsets, reports — are re-derived from it.

---

**Next up:** [Concepts — the graph model](../concepts/graph-model.md) explains what a graph *is* in GSL, in plain language and without any grammar. When you want to go deeper into queries, the [query tutorial](../../QUERY_TUTORIAL.md) is the step-by-step GQL learning path.