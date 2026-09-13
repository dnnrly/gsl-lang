<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Architecture Archaeology — an advanced study

> **The problem:** nobody owns an accurate picture of how parts of the
> system really fit together. Either you have inherited one undocumented
> monolith, or the facts are spread across hundreds of repositories and no
> document ever assembled them.
>
> **The GSL answer:** archaeology is recovery with record-keeping. Every
> fact you dig up is stored **with its provenance** — where it was found and
> how much you trust it — in the same canonical, diffable format as the rest
> of your graph. The model is the deliverable; it improves in git as evidence
> mounts; diagrams are products of queries, drawn fresh.

This study is the re-home for two examples that used to sit in the flagship
line: the single-system archaeology (12 nodes, one undocumented monolith) and
the many-systems archaeology (18 nodes, seven independently-investigated
repositories). They share one deliberate discipline — **never present a
recovered fact without its evidence** — and they are best read together, so
the flagship line now carries the core GSL story in five examples while this
study carries the specialty.

## Why archaeology is an advanced study, not a flagship

The five flagships tell the [core GSL argument](../../flagships/README.md):
graph → canonical text → query → derived view. Archaeology is the same
argument applied to a specialised recovery workflow — it leans on
[flagship 01](../../flagships/01-service-many-views/README.md) for views and
blast radius, on [flagship 05](../../flagships/05-llm-assisted-modelling/README.md)
for the agent loop that produces drafts, and contributes the `source` /
`confidence` provenance convention that 05 consumes. It is a rewarding read,
but go through the flagship line once first.

## The two parts

| Part | What it recovers | Model | Queries |
|------|------------------|-------|---------|
| [Part 1 — one system](one-system/README.md) | a decade-old undocumented monolith | `one-system/model.gsl`, 12 nodes | `q1..q5` + committed results |
| [Part 2 — many systems](many-systems/README.md) | an enterprise across seven repositories | `many-systems/model.gsl`, 18 nodes | `q1..q7` + committed results |

### Part 1 — one system

Targets a single legacy monolith. The archaeology loop lives in git: a
week-one observation is recorded *from memory* (`source="interview"`,
`@suspected`); finding its cron config promotes it to
`source="deploy"`, `@confirmed` — a one-line diff that *is* the audit trail.
Questions asked against `one-system/model.gsl`:

- the **confirmed skeleton** — the diagram you'll bet on (confirmed nodes,
  high-confidence edges only, orphans removed);
- **suspected structure feeding critical parts** — the prioritised
  investigation list, derived;
- the monolith's **one-hop world**, and what a worker drives downstream;
- the **trust-this-last list** — everything we only know from interviews.

### Part 2 — many systems

Scales the discipline across independent repositories. Each investigation
produces one fragment in `many-systems/discovered/`; concatenating the
fragments (`cat discovered/*.gsl > model.gsl`) is the **composition**, and
GSL's declarative merge turns seven fragments into one queryable graph.
Questions asked against `many-systems/model.gsl`:

- who calls `payment_service`, and what its blast radius is;
- whether the `order_completed` publishers and consumers actually line up
  (one deliberately does *not* — the naming mismatch task);
- the **architect's review queue** — every low- or medium-confidence edge,
  as the deliverable of the archaeology.

## Provenance discipline

Both parts use the same conventions. They are **modelling conventions, not
language features** — GSL stores untyped attributes, so nothing enforces
them, and that is a choice:

- `source` — `code`, `config`, `docs`, `repo` (for a node), or `logs` /
  `deploy` / `interview` (single-system flavour);
- `confidence` — `high` / `medium` / `low`;
- set tags — `@confirmed`, `@suspected`, `@critical` keep the working
  hypothesis explicit.

The value is that all of it is **queryable**: risk lists, review queues and
trust-this-last lists are outputs of `gsl-query`, not paragraphs you have to
re-derive.

## Run everything

```bash
go test ./examples -run Flagship -v     # the five flagship examples
go test ./examples -run Advanced -v     # this study: models, queries, results, READMEs
cd one-system   && gsl-query 'subgraph node in @confirmed | remove edge where edge.confidence != "high" | remove orphans' < model.gsl  # the diagram you'll bet on
cd many-systems && ./compose.sh && gsl-query 'subgraph edge.confidence == "low"' < model.gsl                                    # what to verify
```

## Files

| Path | What it is |
|------|-----------|
| `one-system/model.gsl` | the single-monolith archaeology ledger, 12 nodes |
| `one-system/q1..q5` | single-system questions as executable query/result pairs |
| `one-system/views/` | derived charts for Part 1 |
| `many-systems/discovered/*.gsl` | seven independent investigation fragments |
| `many-systems/model.gsl` | the composed enterprise graph, 18 nodes |
| `many-systems/compose.sh` | reproduces `model.gsl` from the fragments |
| `many-systems/q1..q7` | enterprise questions as executable query/result pairs |
| `many-systems/views/` | derived charts for Part 2 |