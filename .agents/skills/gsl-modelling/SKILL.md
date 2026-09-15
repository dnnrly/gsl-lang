---
name: gsl-modelling
description: >-
  Turn arbitrary source material (prose, reports, tickets, event logs, interviews) into a
  source-faithful GSL graph model, with optional analysis and GQL queries. Use when the user
  asks to "model this in GSL", "create a GSL graph of X", or wants a durable, queryable graph
  of relationships (architectures, workflows, organisations, networks). Use ONLY when a
  graph-shaped model is genuinely the right deliverable — not for casual questions that prose
  or a table answers better. Follow the source; never "fix" it; no invented facts, edges,
  attributes, or conventions.
license: CC-BY-4.0
compatibility: host-agnostic; optional step-up tools gsl-query, gsl-diagram, gsl-lsp
metadata:
  version: 1.0.0
---

<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# GSL Modelling Skill

Turn source material into a **faithful GSL graph model** — a durable,
queryable description of what the source says, nothing more. GSL is the
graph specification language of the gsl-lang project; GQL is its query
language.

## Role

Source → faithful GSL model → interpretation. The model is the artifact.
Diagrams, subsets, and reports are views derived from it, only when the
user asks.

## Fidelity guardrails (non-negotiable)

1. **Model the source, not the truth.** Claims that contradict well-known
   facts are modelled as the source states them — never corrected.
   ("The Earth is flat" stays exactly that, if that is what the source says
   and the user wants it modelled.)
2. **No invented content.** No nodes, edges, attributes, or sets that the
   source does not support. When a fact is weak, keep it identifiable as
   weak — never silently upgrade it to fact.
3. **Preserve contradictions.** A source that disagrees with itself keeps
   both sides findable. Don't resolve by deletion unless the user decides.
4. **No example-specific conventions.** Never copy headers, licensing,
   attribute vocabularies, or style from any example file into new output
   unless the user asks.
5. **No copyright or SPDX header by default** on generated files. Add one
   only when the user explicitly requests it.
6. **Existing models are treaty territory.** Never modify one without
   explicit instruction.

## Primary workflow

Steps marked *internal* happen inside your reasoning; the rest are
user-visible.

1. **Scope (user-visible).** Confirm the source, the output shape (model
   only / + analysis / + queries), the mode, granularity, and whether an
   existing model is in play.
2. **Analyse (internal).** Inventory loci, relationships, attributes,
   boundaries, contradictions, unsupported claims — before writing
   anything. Method: `references/MODELLING.md`.
3. **Fidelity triage (user-visible only when consequential).** Local
   ambiguity → proceed. Ambiguity that would change the model's shape or
   set a convention → pause and ask.
4. **Review (internal).** Run the checklists in
   `references/MODELLING.md` (delegated reviewers if the host supports
   them; reviewers raise findings, you own the model).
5. **Construct (internal).** Draft GSL against
   `references/GSL_GUIDE.md` + `references/GRAMMAR.md`.
6. **Self-check (internal).** Correctness + consistency checklists; fix
   structural errors before delivery.
7. **Deliver (user-visible).** The model (+ analysis / queries as
   scoped). Include your analysis notes with any weak facts, so
   uncertainty stays visible.
8. **Offer step-up (user-visible).** Offer `gsl-query ""` validation,
   a representative GQL question, and a diagram — as offers with exact
   commands. Never claim you validated/rendered/executed something you did
   not: see `references/TOOLING.md`.

Run the fidelity review twice — once on the draft, once just before
delivery (see "Correction-temptation review" in MODELLING.md).

## Modes

| Mode | When | What it adds |
|---|---|---|
| Quick | small, single-fact source | minimal ceremony; only consequential pauses |
| Standard (default) | most sources | full self-review with checklists |
| Thorough | large or messy sources | delegated specialist reviewers where supported; extra fidelity passes |
| Interactive | ambiguous or convention-heavy sources | pauses at consequential ambiguity; user joins convention-setting |

Choose a sensible default from source complexity; switch on request.

## Existing-model handling

- **generate** — new model from source (default).
- **by-example** — adopt the conventions of a model the **user supplies**.
- **extend** — add to an existing model; follow its conventions; never
  restructure unless asked.
- **modify / compare** — only on explicit instruction.

## GQL position (GSL-first, GQL-aware)

GQL reads and transforms the graph to answer questions on the model.
Without tooling, write queries to express intent (and explain them). With tooling and
step-up present, you may run them. Respect GQL's maturity: it is a Revised
Draft — never present syntax that the specification or its tests do not
support as established; use the spec and the fixture tests as authority.
Restrict generated queries to constructs shown in
`references/GQL_GUIDE.md` + `references/QUERY_GRAMMAR.md`.

## Authority

1. GSL semantics (`references/GSL_GUIDE.md`, `references/GRAMMAR.md`).
2. Explicit user instruction.
3. An explicitly supplied existing model's conventions.
4. This skill's defaults.

## Navigation

| Task | Open |
|---|---|
| Method, fidelity priorities, checklists, delegation | `references/MODELLING.md` |
| What belongs in a model (node/edge/attr/set) | `references/modelling-with-gsl.md` |
| Write correct GSL | `references/GSL_GUIDE.md`, `references/GRAMMAR.md` |
| Formulate / check GQL | `references/GQL_GUIDE.md`, `references/QUERY_GRAMMAR.md` |
| Validate, run, render (tools) | `references/TOOLING.md` |
| Behaviour anchors | `references/EXAMPLES.md` |

Five `references/*.md` symlinks point to the repository's authoritative
documents; the authored files (MODELLING / TOOLING / EXAMPLES) hold the
method.

## Output conventions to remember

- Declare every used set; expect a non-fatal warning if you don't.
- Identifiers: letters, digits, underscore only — no hyphens.
- No headers of any kind on generated files unless the user asks.
- Keep weak facts weak (MODELLING.md "Uncertainty"); name no convention
  that wasn't the user's.