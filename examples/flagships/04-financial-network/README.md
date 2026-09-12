# 04 — A Financial Relationships Network (Non-Software Domain)

> **The problem:** in financial networks the *relationships* are the thing
> that matters. If a clearing house fails, which institutions are exposed —
> and how does that exposure spread through the interbank web? Regulators
> and risk teams answer this with spreadsheets and bespoke software.
>
> **The GSL answer:** the network is a graph; GSL encodes it as a canonical,
> diffable text file. The same blast-radius queries that answered "retire
> legacy-orders?" in flagship 01 answer "the CCP fails?" here — *unchanged*.
> GSL serves graph-shaped knowledge, not just software.

---

## The model

Ten institutions and critical rails, with `type`, `territory`, and edge
semantics for `member`, `clears`, `lending`, `direct`, `settles`:

- 4 banks, a broker, an exchange
- 2 payment rails (`swift`, `chaps`), a central counterparty (`ccp`), and a
  settlement treasury
- `@critical` marks systemically important participants and rails
- interbank lending forms a **cycle** (alpha → beta → gamma → delta → alpha)
- derivatives exposure carries a `notional` value

![Full network](views/full.graph.mmd)

Everything you see in this example is derived from `model.gsl`.

---

## View 1 — CCP failure: the blast radius

*"If the central counterparty implodes, who is exposed?"*

```bash
gsl-query '(subgraph node.id == "ccp" traverse in all) as BLAST | from BLAST' < model.gsl \
| gsl-diagram -f mermaid -t graph
```

![CCP blast radius](views/ccp-blast-radius.graph.mmd)

The transitive dependency cone — every institution that clears through the
CCP *plus* everyone reachable through the interbank web. Note what the query
doesn't do: it doesn't ask you to pre-list the counterparties. The graph
answers.

---

## View 2 — Critical exposure

*"Of the exposed institutions, which are themselves systemically critical?"*

```bash
gsl-query '(subgraph node.id == "ccp" traverse in all) as BLAST | from * | (subgraph node in @critical) as CRIT | BLAST & CRIT' < model.gsl \
| gsl-diagram -f mermaid -t graph
```

![Critical exposure](views/critical-exposed.graph.mmd)

`alpha_bank` is exposed to the CCP and is itself `@critical`. A one-line
pipeline derives the cross-product of two structural facts. No spreadsheet,
no waiting.

---

## View 3 — The derivatives book

*"Who clears derivatives, through whom?"*

```bash
gsl-query 'subgraph edge.instrument == "derivatives"' < model.gsl \
| gsl-diagram -f mermaid -t graph
```

![Derivatives cleared](views/derivatives-cleared.graph.mmd)

Edge attributes select by *instrument* — the predicate is on the
relationship, not the institution.

---

## View 4 — Institution-type dependency view

*"Zoom out: how do bank / rail / counterparty types connect?"*

```bash
gsl-query 'collapse into banks where node.type == "bank" | collapse into rail_systems where node.type == "payment_rail" | collapse into counterparties where node.type == "central_counterparty"' < model.gsl \
| gsl-diagram -f mermaid -t graph
```

![Type-level view](views/type-collapse.graph.mmd)

Ten institutions collapse to a handful of typed participants. Parallel edges
are preserved (banks use `rail_systems` as members *and* one has direct
access) — no information silently disappears when you zoom out.

---

## View 5 — The interbank lending web

*"Show me the bilateral obligations between banks."*

```bash
gsl-query 'subgraph edge.instrument == "lending"' < model.gsl \
| gsl-diagram -f mermaid -t graph
```

![Interbank lending web](views/interbank-lending.graph.mmd)

A genuine *cycle* — the graph-shaped scenario spreadsheet software handles
badly. GSL models, queries, and diffs it as naturally as anything else.

---

## Run the "two-minute test"

```bash
gsl-query '(subgraph node.id == "ccp" traverse in all) as BLAST | from BLAST' < model.gsl                     # CCP blast radius
gsl-query 'subgraph edge.instrument == "derivatives"' < model.gsl                                               # derivatives book
gsl-query 'collapse into banks where node.type == "bank" | collapse into rail_systems where node.type == "payment_rail"' < model.gsl
```

All three are literally Query 2 / Query 4 / Query 5 from flagship 01 with
the *domain* changed. The language didn't care.

---

## The point of example 04

Flagships 01–03 are software-shaped. This one is not — and that's the
evidence that **GSL is a general graph language**. The operations that
mattered (blast radius, critical-sets intersection, attribute-filtered
views, collapse by type, cycles) are genuine across domains. If you need
"relationships that matter" in a text file that survives code review and
git, the domain need not be microservices.

---

## Limitations

- Attributes are **untyped strings** — `notional="400"` cannot be summed or
  compared with a `>` in GQL today. Structural questions are first-class;
  numeric aggregation needs a script over the serialized output (or a
  future GQL feature).
- Direction conventions must be documented. Here "A -> B" means "A is
  exposed to / participates in B"; the *meaning* of an edge lives in its
  attribute and the human convention, not in the grammar.
- `@critical` (and any exposure modelling) reflects a judgement call. The
  model is as good as the assumptions tagged into it; GSL makes those
  assumptions explicit and reviewable, not hidden.
- This is a *structural* model. It says who is exposed and via what
  instrument; it says nothing about loss-given-default or contagion
  timing.