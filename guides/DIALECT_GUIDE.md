---
name: gsl-dialect-guide
type: authoring-guide
language: GSL
version: 0.4.0-draft
description: Authoring guide for GSL dialects - YAML documents that declare conventions for nodes, edges and sets and validate projected attributes with JSON Schema. Covers the document structure, node and edge rules, set definitions and policy, the supported GQL match subset, GSL-to-JSON projection, writing schemas against projected values, and the gsl-dialect command-line tool.
keywords: [gsl, dialect, json-schema, validation, conventions, projection, gql]
spec: DIALECTS_SPEC.md
implementation: Go
---

<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# GSL Dialect Authoring Guide

A **dialect** is a YAML document that says how a graph should be shaped and what
its attributes must look like. It answers *"our team has conventions — can the
model enforce them?"* A dialect does **not** change GSL syntax. It selects nodes
and edges with GQL match rules, validates their projected attributes with JSON
Schema, and checks set declarations against a declaration policy.

The normative definition lives in
[`spec/dialects/DIALECTS_SPEC.md`](../spec/dialects/DIALECTS_SPEC.md). This
guide is the practical companion; where the two disagree, the specification
wins.

- A dialect is a *convention*, not a *schema for GSL*. The graph is still valid
  GSL whether or not it conforms.
- Dialects select and check; they never rewrite or reject the graph.

## 1. What a dialect is (and is not)

| A dialect **is** | A dialect **is not** |
|---|---|
| A YAML document with a stable `id` and `version` | A replacement for the GSL grammar |
| Ordered node and edge rules: match + attributes | A constraint on how you write GSL text |
| Named set definitions plus a declaration policy | A graph transformation |
| A JSON Schema check over projected values | Executable prose |
| Optional guidance for humans and agents | A schema for the whole graph |

Conformance is defined **over selected nodes and edges and declared sets**. A
graph conforms even if it contains no element matching any rule — an empty match
is not a failure.

## 2. Document structure

A dialect is a YAML mapping. The root fields are:

| Field | Required | Meaning |
|---|---|---|
| `id` | yes | Stable identifier for the dialect |
| `version` | yes | Version of the dialect definition |
| `title` | no | Human-readable name |
| `description` | no | Purpose of the dialect |
| `guidance` | no | Advice for constructing graphs |
| `nodes` | no | Node rules (`nodes.rules`) |
| `edges` | no | Edge rules (`edges.rules`) |
| `sets` | no | Set policy (`sets.policy`) and definitions (`sets.definitions`) |

The three sections are independent and optional. A dialect with no sections is
valid and conforms to every graph. Unknown fields are ignored, so misspelled
field names can pass unnoticed — review your diagnostics.

```yaml
id: software-architecture
version: 0.4.0-draft
title: Software Architecture
description: Conventions for describing services and their dependencies.

nodes:
  rules:
    - name: Service
      match: 'kind == "service"'
      attributes:
        type: object
        required: [kind, name, replicas]
        properties:
          kind:
            const: service
          name:
            type: string
            minLength: 1
          replicas:
            type: integer
            minimum: 1
```

## 3. Node and edge rules

`nodes.rules` and `edges.rules` are **ordered sequences**. Each rule has:

| Field | Required | Meaning |
|---|---|---|
| `name` | no | Dialect-local label; unnamed rules are `nodes.rules[0]`, `edges.rules[1]`, … |
| `match` | yes | A bare GQL attribute predicate (Section 5) |
| `attributes` | yes | JSON Schema for the projected attribute map |
| `description` | no | Prose explaining the rule |
| `guidance` | no | Advice for authoring matching elements |

Rules are ordered for readability, but **every matching rule applies
independently**. There is no first-match-wins behaviour: an element may be
selected and validated by several rules, and every applicable violation is
reported.

The section supplies the element kind and the `node.`/`edge.` prefix:

| Section | Selects | Element identity in diagnostics |
|---|---|---|
| `nodes.rules` | nodes | the node id |
| `edges.rules` | edges | the edge label, else `from->to` |

An edge rule is written exactly like a node rule:

```yaml
edges:
  rules:
    - name: Dependency
      match: 'protocol exists'
      attributes:
        type: object
        required: [protocol]
        properties:
          protocol:
            enum: [http, grpc, tcp]
```

## 4. Sets: definitions and policy

Sets are **not** selected by a query. They are described by name under
`sets.definitions`, keyed by the GSL set name:

```yaml
sets:
  definitions:
    critical:
      description: Elements whose failure has immediate impact.
      guidance: Apply `@critical` to high-priority services.
      attributes:
        type: object
        properties:
          tier:
            type: string
    deprecated:
      description: Elements that should not gain new dependants.
```

A set definition's `attributes` is optional. When present, it validates the
projected attributes of that set; when absent, the definition only documents the
set. Set definitions are never matched, and the GSL set name is both the
definition key and the element identity in diagnostics.

`sets.policy` controls declaration-level conformance. All values default to the
permissive setting:

| Policy | Default | When `false` | When `true` |
|---|---|---|---|
| `allow_undefined` | `true` | Every set in the graph must have a definition | — |
| `require_all` | `false` | — | Every defined set must be present in the graph |
| `allow_empty` | `true` | Every present set must have at least one member | — |

For example, this policy rejects undeclared classifications and requires both
defined sets to be present and non-empty:

```yaml
sets:
  policy:
    allow_undefined: false
    require_all: true
    allow_empty: false
  definitions:
    critical:
      description: High-priority elements.
    deprecated:
      description: Elements being retired.
```

An implicitly created set — one introduced by `@name` without a `set name`
declaration — counts as present and declared for policy purposes. Requiring a
set to be present does not require an explicit `set` statement; requiring it to
be non-empty does require at least one member.

A policy violation is a **graph** conformance violation, not an invalid dialect.
The diagnostic's definition is the set name for `require_all` and
`allow_empty` failures; an undefined-set violation has no definition.

## 5. Matching: the supported GQL subset

`match` is written as a **bare attribute predicate** — no `subgraph` keyword and
no `node.`/`edge.` prefix. Both are supplied by the enclosing section. The
predicate is expanded to a full `subgraph` query before evaluation, so a node
rule with the match below means `subgraph node.kind == "service"`:

```yaml
match: 'kind == "service"'
```

Typical match expressions:

```yaml
match: 'protocol exists'          # every edge with a protocol
match: 'replicas != 0'            # nodes whose replicas are not zero
match: 'tier == "frontend"'       # attribute equality
match: 'zone == "B" AND active exists'
```

Writing an explicit `subgraph` wrapper or element prefix (`node.`, `edge.` or
`set.`) is an invalid dialect — the prefix is already implied by the section,
and stating it twice invites disagreement.

Not supported in this subset: graph traversal, `scope`, query pipelines, graph
algebra, transformations, and range comparisons — the GQL grammar has no `>`,
`>=`, `<` or `<=` operators, so `replicas > 1` is rejected at load. Select with
equality, inequality, `exists`, set membership or depth instead. An unsupported
or malformed match expression is an **invalid dialect**, rejected before graph
validation begins. The precise grammar is documented in
[`SUPPORTED_GQL_SUBSET.md`](../spec/dialects/SUPPORTED_GQL_SUBSET.md); the full
language is in the [Query Guide](GQL_GUIDE.md).

## 6. Projection: GSL values to JSON

JSON Schema validates JSON, so each selected element's **attributes** are
projected into a JSON object before validation. The projected instance contains
only attributes — never the element's id, kind or endpoints. Absent attributes
are omitted.

The value mappings are normative:

| GSL value | JSON projection | Example GSL | Projected JSON |
|---|---|---|---|
| String | JSON string | `name="Gateway"` | `"Gateway"` |
| Number | JSON number | `replicas=3` | `3` |
| Boolean | JSON Boolean | `active=true` | `true` |
| Empty | object with `$gslEmpty` | `enabled` | `{"$gslEmpty": true}` |
| Node reference | object with `$gslRef` | `owner=platform_team` | `{"$gslRef": "platform_team"}` |

A number is never turned into a string, and a string of digits is never turned
into a number. Set attributes cannot contain node references in GSL, so a set
schema should not expect `$gslRef`.

## 7. Writing schemas against projected values (authoring patterns)

The single most common mistake is writing a schema against **GSL syntax**
instead of the **projected JSON**. These patterns cover the cases that trip
people up.

### 7.1 A plain scalar

```yaml
properties:
  replicas:
    type: integer
    minimum: 1
```

`replicas=3` projects to `3`, so `type: integer` matches.

### 7.2 A node reference

A reference is an **object**, not a string. To accept `owner=platform_team`:

```yaml
properties:
  owner:
    type: object
    required: ["$gslRef"]
    properties:
      "$gslRef":
        type: string
    additionalProperties: false
```

A schema that says `type: string` will **reject** a reference — a frequent
surprise.

### 7.3 A flag that may be present or absent

An explicitly present Empty value (`enabled`) projects to
`{"$gslEmpty": true}`. Require it explicitly:

```yaml
properties:
  enabled:
    type: object
    required: ["$gslEmpty"]
    properties:
      "$gslEmpty":
        const: true
    additionalProperties: false
```

If the flag is optional, leave it out of `required` — an absent attribute is
simply omitted, which `required` already handles.

### 7.4 A value that may be a string or a reference

Use `oneOf` when either form is acceptable:

```yaml
properties:
  owner:
    oneOf:
      - type: string
      - type: object
        required: ["$gslRef"]
        properties:
          "$gslRef":
            type: string
```

### 7.5 A controlled vocabulary

Use `enum` for a fixed set of allowed values:

```yaml
properties:
  protocol:
    enum: [http, grpc, tcp]
```

### 7.6 The object as a whole

`attributes` validates the whole projected map, so start from `type: object`
and list `required` keys:

```yaml
attributes:
  type: object
  required: [kind, name]
  properties:
    kind:
      const: service
    name:
      type: string
  additionalProperties: true
```

Set `additionalProperties: false` only when you want to forbid attributes you
have not declared.

### 7.7 Internal references

A schema may use `$ref` to its own `$defs`, but each rule's `attributes` schema
is a separate resource. A reference cannot resolve into another rule's schema,
and external `https://` or `file:` references are rejected at load. Inline the
shared definition or place it in the same `attributes` object.

## 8. Validation outcomes

`gsl-dialect validate` reports one of three outcomes:

| Outcome | Meaning | CLI exit code |
|---|---|---|
| Conforming | Every selected element and set declaration satisfied the dialect. | `0` |
| Non-conforming | Validation completed; one or more violations. | `1` |
| Incomplete | Validation could not complete (invalid dialect, matching error, I/O or usage). | `2` |

A violation names the **rule or set definition**, the **element** and a **JSON
Pointer** into the projected instance:

```
Non-conforming: 1 violation(s) in model.gsl
  node gateway /replicas: minimum: got 0, want 1
```

A policy violation has an empty JSON Pointer because it is about declaration,
not attribute values:

```
  set beta /: required set is not present in the graph
```

## 9. Using the CLI

```bash
# Conforming → exit 0
gsl-dialect validate -d software-architecture.dialect.yaml -i model.gsl

# See the projected JSON each schema validates
gsl-dialect project -d software-architecture.dialect.yaml -i model.gsl

# Machine-readable report (provisional shape)
gsl-dialect validate -d software-architecture.dialect.yaml -i model.gsl --json
```

Both commands read the graph from **stdin** when `-i` is omitted. See
[`cmd/gsl-dialect/README.md`](../cmd/gsl-dialect/README.md).

## 10. A complete example

The reference example lives in
[`examples/dialects/software-architecture/`](../examples/dialects/software-architecture/):
a dialect with a `Service` node rule, a `Dependency` edge rule and named
`critical`/`deprecated` set definitions under a strict declaration policy, plus
a conforming `model.gsl`. Read it end to end before authoring your own.

## 11. Checklist for authoring a dialect

1. Give the dialect a stable `id` and a `version`.
2. Put node conventions under `nodes.rules` and edge conventions under
   `edges.rules`; each rule has a bare `match` and an `attributes` schema.
3. Name rules when diagnostics should be readable; otherwise use the positional
   label (`nodes.rules[0]`).
4. Define sets by GSL set name under `sets.definitions`; select a policy only
   when you need declaration-level enforcement.
5. Write `match` as a bare attribute predicate — never add a prefix or
   `subgraph`.
6. Write `attributes` against **projected JSON**, not GSL syntax.
7. Use `gsl-dialect project` to inspect the projected instance while iterating.
8. Validate the reference model, then add mutations to confirm your schemas
   actually reject what they should.
