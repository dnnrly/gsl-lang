---
name: gsl-modelling-examples
type: skill-reference
audience: agent
description: Three short, neutral worked examples that pin down intended behaviour — faithful modelling of weak facts, by-example adaptation to a user's conventions, and sets vs labels. Invalid blocks show what the parser rejects.
---

<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Examples

Three short cases. They show **behaviour**, not style: what to keep, what
to adopt, and what the parser will reject. Nothing here establishes a
convention — the models use generic names and the attribute choices in
each case are the author's choice for that case, not a template.

## 1. Faithful modelling with a weak fact

**Source (prose):** *"The ordering service submits orders to the payment
service over a message queue. Order failures are dropped. Jane suspects
the early-warning service reads those drops, but she is not sure."*

**Model (what the source supports):**

```gsl
node ordering_service
node payment_service
node early_warning_service

ordering_service->payment_service [transport="queue"]
payment_service->early_warning_service
```

- The first edge is directly supported ("submits orders over a queue").
- The second edge is *only* Jane's suspicion. It is kept because it is a
  fact about the system the user asked to model — but it is **weak**, and
  the right way to say so here is in the accompanying note, not by
  inventing an attribute convention:

*"payment → early_warning is Jane's suspicion only (not confirmed). It is
included so the question can be verified; the direction models the flow of
drop records (producer to consumer) — invert it if your team's dependency
convention is consumer→producer."*

Nothing was added, corrected, or upgraded to a fact.

## 2. Follow a user-supplied model's conventions

**User-supplied existing model (this is the user's vocabulary):**

```gsl
set services [env="prod"]
node checkout_service [team="storefront"] @services
node inventory_service [team="storefront"] @services
checkout_service->inventory_service [protocol="grpc"]
```

**Source:** *"The payments service is also run by the storefront team."*

**New content — added by example, adopting only the user's own
conventions:**

```gsl
node payments_service [team="storefront"] @services
```

The `team`/`env`/`@services` vocabulary came from the supplied model, not
invented here. If the source had named only an attribute the user's model
has no place for, the skill would say so and surface it in notes rather
than quietly adding a new attribute kind.

## 3. Sets are boundaries, not labels

A set earns its place when membership explains shared fate or scope:

```gsl
set rollout_candidate
node checkout_service @rollout_candidate
node payments_service @rollout_candidate
```

Membership in one set joins the nodes for a reason — "both are in the
next rollout" — that a query (`subgraph node in @rollout_candidate`) can
act on.

Using a set only to repeat a status already carried as an attribute
adds a second, weaker source of truth. The following parses fine — the
parser cannot see this — but it is a modelling smell: `@deprecated`
restates `status="deprecated"` instead of grouping anything:

```gsl
set deprecated
node legacy_service [status="deprecated"] @deprecated
```

Prefer one home for a status fact: an attribute *or* a set, chosen by
whether there is a boundary (a group of things sharing that status) worth
querying — not both by reflex.

## What the parser rejects (structural errors to avoid)

These are real drafting mistakes — invented syntax and illegal identifiers:

```invalid-gsl
@critical: checkout_service, payments_service   # list-set syntax does not exist
node catalog-service [team="storefront"]         # hyphens are illegal in identifiers
```

The parser stops at the first: it catches structure, never semantics. A
valid parse says nothing about whether the model is faithful — that is
the review's job (MODELLING.md).

## Further reading

- The full messy-prose → reviewed-model trajectory lives in
  `examples/flagships/05-llm-assisted-modelling/` (source, drafts,
  canonical diff, queries). Its attribute choices are **that example's
  choices, not conventions** — the skill does not adopt them unless the
  user supplies a model that uses them.
- `examples/advanced/architecture-archaeology/` shows provenance-style
  modelling from failure evidence. Same caveat: read it as illustration,
  not as a convention to reproduce.