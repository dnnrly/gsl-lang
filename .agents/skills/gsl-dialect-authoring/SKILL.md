---
name: gsl-dialect-authoring
description: >-
  Author and review GSL dialect documents — YAML files that declare conventions for a
  graph (which nodes/edges/sets must exist and what their projected attributes must look
  like) and validate them with JSON Schema. Use when the user asks to "enforce our
  conventions", "validate this model against a dialect", "write a .dialect.yaml", or
  "why does this graph not conform?". Follows the dialect specification; encodes only the
  conventions the source states — never invents rules, vocabularies, or required fields.
license: CC-BY-4.0
compatibility: host-agnostic; optional step-up tool gsl-dialect
metadata:
  version: 2.0.0
---

<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# GSL Dialect Authoring Skill

Turn a set of stated conventions into a **dialect** — a YAML document that
selects nodes and edges with GQL match rules, validates their projected
attributes with JSON Schema, and checks named set declarations against a
policy — and use it to check a GSL model. GSL is the
graph specification language of the gsl-lang project; GQL is its query
language; a dialect is a convention layer over both.

## Role

Conventions → dialect → validation report. A dialect is a *convention*, not a
*change to GSL*. The graph is valid GSL whether or not it conforms; the dialect
selects and checks, and never rewrites or rejects the source.

## Fidelity guardrails (non-negotiable)

1. **Encode only stated conventions.** Never invent a required attribute, an
   allowed vocabulary, a cardinality, or a naming rule the user or source does
   not state. If a convention is implied but unstated, surface it and ask.
2. **Model conventions, not truth.** A dialect captures what the team agrees a
   model *should* look like. It is not a claim about the domain.
3. **No example-specific copying.** Never lift `id`, `version`, attribute
   names, or vocabularies from any example file unless the user supplies them.
4. **Schemas target projection, not syntax.** Write JSON Schema against the
   projected JSON value (`references/DIALECT_GUIDE.md`, section 7) — the most
   common authoring error is describing GSL text instead of the projected value.
5. **Respect the supported subset.** Match queries use a single `subgraph`
   predicate only (`references/SUPPORTED_GQL_SUBSET.md`). Never approximate an
   unsupported query.
6. **Existing dialects are treaty territory.** Never modify one without
   explicit instruction.

## Primary workflow

Steps marked *internal* happen inside your reasoning; the rest are
user-visible.

1. **Scope (user-visible).** Establish the conventions, the graph they apply
   to, and which element kinds (node / edge / set) they cover. If a convention
   is ambiguous or consequential, pause and ask.
2. **Inventory (internal).** For each convention decide the section (`nodes`
   or `edges`), the bare selection query, and the attribute schema. Set
   conventions become keys under `sets.definitions`, with `sets.policy` only
   where declaration-level enforcement is stated. Method:
   `references/WORKFLOW.md`.
3. **Draft (internal).** Write the dialect against
   `references/DIALECT_GUIDE.md`; validate the document against
   `references/DIALECTS_SPEC.md`.
4. **Check (user-visible when tooling is present).** Run `gsl-dialect
   validate`; if it does not conform, report each violation (definition,
   element, JSON Pointer) without "fixing" the model unless asked.
5. **Inspect (internal).** Use `gsl-dialect project` to see the projected JSON
   the schemas actually validate.
6. **Deliver (user-visible).** The dialect, a validation result, and — when a
   convention could be misread — the projected instance that proves the schema
   sees what you intended.
7. **Offer step-up (user-visible).** Offer the exact `gsl-dialect` commands you
   would run. Never claim you validated something you did not.

## Modes

| Mode | When | What it adds |
|---|---|---|
| Quick | one obvious rule | minimal ceremony; draft, check, deliver |
| Standard (default) | a handful of conventions | full document review + projection inspection |
| Thorough | many definitions or shared dialects | per-definition checks; mutation cases proving each schema rejects what it should |

## Authority

1. The dialect specification (`references/DIALECTS_SPEC.md`).
2. The supported GQL subset (`references/SUPPORTED_GQL_SUBSET.md`).
3. Explicit user instruction.
4. This skill's defaults.

## Navigation

| Task | Open |
|---|---|
| Authoring method, selection/schema design, checklists | `references/WORKFLOW.md` |
| Document structure, projection, schema patterns | `references/DIALECT_GUIDE.md` |
| Normative rules, outcomes, exit codes | `references/DIALECTS_SPEC.md` |
| What match queries are allowed | `references/SUPPORTED_GQL_SUBSET.md` |

Three `references/*.md` files are plain copies of the repository's
authoritative documents (kept in sync by `TestAgentSkillMirrorDrift`,
refreshed with `go test -run TestAgentSkillMirrorDrift -update .`); the
authored `WORKFLOW.md` holds the method.

## Output conventions to remember

- A dialect needs a stable `id` and a `version`; both are strings.
- Node rules live under `nodes.rules`; edge rules under `edges.rules`; the
  section supplies the `node.`/`edge.` prefix and the element kind.
- `match` is a bare attribute predicate with no `subgraph` keyword and no
  element prefix.
- Sets are keyed by GSL set name under `sets.definitions`; they are never
  selected by a query. Use `sets.policy` for `allow_undefined`, `require_all`
  and `allow_empty` only when the source states the convention.
- Write schemas against projected values: a node reference is
  `{"$gslRef": "..."}`; an explicit Empty is `{"$gslEmpty": true}`.
- Exit codes: `0` conforming, `1` non-conforming, `2` incomplete.

## Worked material

The reference dialect lives in
`examples/dialects/software-architecture/` in the repository (node rules, an
edge rule, and named set definitions under a strict policy). Read it before
authoring a new dialect.
