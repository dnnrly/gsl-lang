---
name: gsl-modelling-method
type: skill-reference
audience: agent
description: Operational method for turning source material into a faithful GSL model — fidelity priorities, source analysis, ambiguity handling, mechanism-neutral uncertainty, review checklists, and delegated review. Companion to SKILL.md; for syntax see GSL_GUIDE.md.
---

<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Modelling Method

This is the "how" of the skill: how to turn source material into a
faithful GSL model, and how to know you did not quietly alter it.

## Fidelity priorities

When two goals conflict, resolve in this order:

1. **Faithful** — the model says what the source says.
2. **Structurally appropriate** — the shape of the model (nodes, edges,
   sets, nesting) follows the source, using the modelling guidance in
   `modelling-with-gsl.md`.
3. **Consistent** — the model is internally coherent (naming, structure,
   attribute usage).
4. **Explicit about uncertainty** — weak or unsupported facts are
   distinguishable from strong ones.
5. **Complete** — nothing material in the source is missing.
6. **Traceable** — where useful, the user can see why a fact is in the
   model.

Priority 1 beats everything. A "cleaner" model that does not say what the
source says is not a better model.

## The pipeline

Analyse, review, construct. The gatekeepers are honesty, not speed.

### 1. Analyse

Read the whole source before writing anything. Identify:

- **Loci** — things with their own relationships and life span (nodes).
- **Relationships** — connections between loci, including direction,
  multiplicity, and the mechanism or medium.
- **Attributes** — single-value facts that describe a node or edge, not
  relationships between two things.
- **Boundaries / sets** — group membership implied by shared fate,
  lifecycle, or rollout scope, not by shared adjectives.
- **Contradictions** — places where the source disagrees with itself.
- **Unsupported or weak claims** — asserted without evidence, hedged
  ("I think", "probably", "someone said"), or from a low-trust source.
- **Silent-knowledge traps** — statements your background would make you
  want to "correct" or "improve". Flag them; the source wins.

Make no model decisions yet. Just inventory and notes.

### 2. Fidelity triage

Apply the ambiguity protocol (SKILL.md, §Fidelity guardrails) to each
finding:

- **Local ambiguity** (a single weak fact, a spelling variant, minor
  uncertainty) — proceed, and keep the faithful representation.
- **Consequential ambiguity** (would change the model's shape or set a
  convention the whole model follows) — pause and ask.

**Correction temptation.** A claim that contradicts well-known external
facts is still modelled as the source states it. The user wants a model
of their source, not a correction of it. If they want the discrepancy
recorded, that is their call and is best surfaced in the accompanying
notes — never silently fixed in the graph.

### 3. Review

Run at least the fidelity review checklist (below) before constructing.
Delegated reviewers may help; the constructing agent remains the owner.
See "Delegated review" at the end of this file.

### 4. Construct

Draft the GSL following:

- Syntax and rules: `references/GSL_GUIDE.md` and `references/GRAMMAR.md`.
- Modelling decisions (what goes where): `references/modelling-with-gsl.md`.
- The fidelity guardrails and authority hierarchy in `SKILL.md`.

Structural discipline while drafting:

- Declare every set you use (`@name` on something without a `set name`
  declaration is only a parser warning — but it is a signal you are
  hiding a grouping that should be explicit).
- Identifiers: `[A-Za-z_][A-Za-z0-9_]*` — no hyphens, no spaces.
- Resolve contradictions by keeping both facts, not by deleting one, when
  both carry information (see flagship 05: the "double write" was kept
  because it was the risk, not an error).

## Uncertainty, without prescribing a vocabulary

The skill establishes **no attribute conventions** — no required
attribute names, no required sets, no required headers. This keeps the
skill from inventing dialects and from importing conventions of one
project into another.

For a weak or unsupported fact, choose among these, in order of
preference:

1. **Follow the user's own model conventions** if one was explicitly
   supplied (by-example mode). If their model marks weakness with an
   attribute, use that attribute. Do not invent a new one.
2. **Omit** the fact if it is unsupported material detail with no
   modelling value.
3. **Surface it in the accompanying notes** — a short sentence next to
   the model ("the early-warning read is asserted only by Jane; it is
   included as an edge on her word"). This is always safe and always
   honest.
4. Keep it with a **plain-language attribute** whose name the user can
   recognise and rename; say you did so and that it is not a convention.

The default is 2 or 3. Reach for 4 only when the fact matters structurally
and the user wants traceability inside the graph — and never present the
attribute name as anything other than a suggestion.

## Granularity

Match detail to the questions the model must answer. Too much detail is
noise; too little loses information. When unsure, preserve the fact
(`modelling-with-gsl.md` states this bias) — but a fact is only worth
preserving if it answers a question the user cares about. State your
granularity choice in the analysis so the user can push back.

## Review checklists

### Fidelity review (the non-negotiables)

- [ ] Every node, edge, attribute, and set is supported somewhere in the
      source — nothing interpolated from general knowledge.
- [ ] No fact corrected or "improved" to make it more sensible or true.
- [ ] Contradictions preserved, both sides findable.
- [ ] Weak claims distinguishable from strong ones (see Uncertainty above).
- [ ] "Confidence" judgements come from the source (or the user), never
      from the agent's priors.

### Correctness review (structural)

- [ ] Parses: `gsl-query ""` (or the self-check in TOOLING.md) exits
      clean, ignoring trailing non-fatal warnings you have consciously
      accepted.
- [ ] All sets used are declared.
- [ ] Valid identifiers; attributes valid; no invented syntax.
- [ ] Canonical form round-trips: `parse(serialize(parse(x))) == parse(x)`.

### Consistency review

- [ ] Naming is uniform; one thing has one name.
- [ ] The same fact appears once, not twice in different shapes.
- [ ] Attribute usage does not vary meaninglessly between similar items.

### Correction-temptation review (run last, deliberately)

- [ ] I modelled the source's claim, not "the truth", wherever they differ.
- [ ] I recorded what the user's source *said*, even when I know better.

## Delegated review

If the host can run sub-agents, use them as **reviewers — not independent
authorities**:

- A **fidelity reviewer** reads the source and the model together and
  lists every unsupported entity, edge, attribute, and set, plus every
  place the model seems to have "improved" on the source.
- A **consistency reviewer** checks naming, duplication, and attribute
  uniformity (correctness + consistency checklists).
- A **correction-temptation reviewer** specifically hunts for places
  where well-known facts leaked into the model.

Reviewers report findings and a recommended change. **The constructing
agent owns the model and the final decision**: it weighs each finding,
keeps or rejects it, and must be able to justify the outcome against the
fidelity priorities. Reviewers never write the model directly. When the
host does not support delegation, the constructing agent runs the same
checklists sequentially on its own work — the reviews are the quality
bar; the delegation is just a way to reach it.

## Authority

When guidance conflicts, in this order:

1. **The GSL language itself** (semantics in `GSL_GUIDE.md` /
   `GRAMMAR.md`) is authoritative over every other rule here.
2. **Explicit user instruction** beats skill defaults.
3. **An explicitly supplied existing model's conventions** are followed
   for new content by example.
4. **This skill's defaults** apply where nothing else does.

Never let an example file in the repository be mistaken for step 3: those
models' attributes and styles are one project's choices, not conventions.