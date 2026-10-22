---
name: gsl-dialect-authoring-workflow
type: skill-reference
audience: agent
description: The authoring method for GSL dialects — turning stated conventions into node/edge rules, set definitions and policies, and projected-value JSON Schemas, then checking a model and proving the schemas reject what they should.
---

<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Dialect Authoring Workflow

This is the method the skill follows. It assumes the normative rules in
[`DIALECTS_SPEC.md`](DIALECTS_SPEC.md), the projection rules in
[`DIALECT_GUIDE.md`](DIALECT_GUIDE.md) and the query limits in
[`SUPPORTED_GQL_SUBSET.md`](SUPPORTED_GQL_SUBSET.md).

## Step 1 — Turn each convention into a rule or set definition

A convention about nodes or edges becomes one entry in `nodes.rules` or
`edges.rules` with three parts:

| Part | Question | Example |
|---|---|---|
| section | Is the convention about a node or an edge? | `nodes` |
| `match` | How do you recognise the element? | `kind == "service"` |
| `attributes` | What must its projected attributes look like? | required `name`, `replicas >= 1` |

A convention about sets becomes a key under `sets.definitions` (the key is the
GSL set name) with optional `description`, `guidance` and `attributes`. A
declaration convention such as "every set must be declared" or "every declared
set must be used" belongs in `sets.policy`.

If a convention cannot answer the relevant parts, it is not yet a dialect rule.
Say so and ask rather than guessing the missing part.

## Step 2 — Choose the selection query

Selection is a single **bare attribute predicate** — no `subgraph` keyword and
no element prefix; the enclosing section supplies both. Prefer a stable
discriminator attribute over a name:

```text
kind == "service"
```

```text
protocol exists
```

`kind` (or another explicit type attribute) is usually the right selector; it
does not change when a node is renamed. Do not write the `subgraph` wrapper or
an explicit `node.`/`edge.` prefix — that is an invalid dialect. Do not try
traversal or pipelines — they are outside the subset.

## Step 3 — Write the schema against projected values

The schema validates the **projected JSON**, never the GSL text. The two
mappings that cause almost every authoring bug:

- a node reference (`owner=platform_team`) projects to an object:
  `{"$gslRef": "platform_team"}` — so the schema must describe an object, not a
  string;
- an explicitly present Empty (`enabled`) projects to `{"$gslEmpty": true}` —
  so require it as an object with a `$gslEmpty` constant if it matters.

Full patterns (scalars, references, optional flags, `oneOf`, enums) are in
[`DIALECT_GUIDE.md`](DIALECT_GUIDE.md), section 7.

## Step 4 — Check a model, then break it

Validate the reference model, then confirm each rule actually rejects a
mutation. A schema nobody can violate is not doing anything.

```bash
gsl-dialect validate -d software-architecture.dialect.yaml -i model.gsl
```

```bash
gsl-dialect project -d software-architecture.dialect.yaml -i model.gsl
```

Given a model with a service:

```gsl
node platform_team [kind="team", name="Platform Team"]

node gateway [
  kind="service",
  name="Gateway",
  owner=platform_team,
  replicas=3
]
```

and a dialect that requires a service name and at least one replica:

```yaml
id: software-architecture
version: "1"
nodes:
  rules:
    - name: Service
      match: 'kind == "service"'
      attributes:
        type: object
        required: [kind, name, owner, replicas]
        properties:
          kind:
            const: service
          name:
            type: string
            minLength: 1
          owner:
            type: object
            required: ["$gslRef"]
            properties:
              "$gslRef":
                type: string
          replicas:
            type: integer
            minimum: 1
```

the check reports conformance, and setting `replicas=0` produces:

```text
Non-conforming: 1 violation(s) in model.gsl
  node gateway /replicas: minimum: got 0, want 1
```

## Step 5 — Report honestly

- Quote the actual `gsl-dialect` output; never claim a validation you did not
  run.
- A clean run means the model conforms to *this* dialect — not that the model
  is faithful, complete, or correct.
- A non-conforming run is information about the model, not a licence to edit
  it. Present the violations and let the user decide.
- An `incomplete` result (exit `2`) means validation could not complete — for
  example an unsupported match query or an invalid schema. Fix the dialect, not
  the model.

## Checklist

1. Stable `id` and `version` on the dialect.
2. One rule per node/edge convention in the correct section; one keyed
   definition per set convention.
3. `match` is a bare attribute predicate with no prefix and no `subgraph`.
4. Schema is written against projected values (references and Empty are
   objects); set `attributes` is optional.
5. Set policy is used only when the source states a declaration convention.
6. The reference model validates, and each rule is shown to reject a mutation.
7. Reported results are exactly what the tool printed.
