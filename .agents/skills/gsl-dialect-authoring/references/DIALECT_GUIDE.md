---
name: gsl-dialect-guide
type: authoring-guide
language: GSL
version: 0.2.0-draft
description: Authoring guide for GSL dialects - YAML documents that declare how a graph should be shaped and what its projected attributes must look like. Covers the dialect document structure, element definitions, the supported GQL match subset, GSL-to-JSON projection, writing JSON Schema against projected values, and the gsl-dialect command-line tool. Use when writing or reviewing a .dialect.yaml file, validating a model against conventions, or projecting graph attributes to JSON.
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
its attributes must look like. It is the answer to *"our team has conventions —
can the model enforce them?"* A dialect does **not** change GSL syntax. It
selects elements with a GQL query, projects their attributes into JSON, and
validates that JSON against JSON Schema.

The normative definition lives in
[`spec/dialects/DIALECTS_SPEC.md`](../../../../spec/dialects/DIALECTS_SPEC.md). This
guide is the practical companion; where the two disagree, the specification
wins.

- A dialect is a *convention*, not a *schema for GSL*. The graph is still valid
  GSL whether or not it conforms.
- Dialects select and check; they never rewrite or reject the graph.

## 1. What a dialect is (and is not)

| A dialect **is** | A dialect **is not** |
|---|---|
| A YAML document with a stable `id` and `version` | A replacement for the GSL grammar |
| A set of element definitions: match + attributes | A constraint on how you write GSL text |
| A JSON Schema check over projected values | A graph transformation |
| Optional guidance for humans and agents | Executable prose |

Conformance is defined **over selected elements**. A graph conforms even if it
contains no element matching any definition — an empty match is not a failure.

## 2. Document structure

A dialect is a YAML mapping. The root fields are:

| Field | Required | Meaning |
|---|---|---|
| `id` | yes | Stable identifier for the dialect |
| `version` | yes | Version of the dialect definition |
| `title` | no | Human-readable name |
| `description` | no | Purpose of the dialect |
| `guidance` | no | Advice for constructing graphs |
| `elements` | yes | Mapping of definition name → element definition |

Unknown root fields are ignored, so misspelled field names can pass unnoticed —
review your diagnostics.

```yaml
id: software-architecture
version: 0.1.0-draft
title: Software Architecture
description: Conventions for describing services and their dependencies.

elements:
  Service:
    representation: node
    match:
      query: 'subgraph node.kind == "service"'
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

Each **element definition** requires `representation`, `match` and
`attributes`, and may add `description` and `guidance` (both are prose only —
they never alter matching or validation).

## 3. Representations

`representation` is one of `node`, `edge` or `set`. It narrows the elements a
definition validates to that kind:

```yaml
representation: node
```

A match query may return a graph containing nodes, edges and sets; only those
matching the declared representation are validated. The other kinds are
incidental and ignored.

The three representations correspond to the three things a graph is made of:

| Representation | Selects | Element identity in diagnostics |
|---|---|---|
| `node` | nodes | the node id |
| `edge` | edges | the edge label, else `from->to` |
| `set` | named sets | the set name |

> **Known limitation.** The current GQL engine's `subgraph` copies every set
> declaration unconditionally, so a `set` definition selects *all* sets in the
> input graph. See
> [`SUPPORTED_GQL_SUBSET.md`](../../../../spec/dialects/SUPPORTED_GQL_SUBSET.md).

## 4. Matching: the supported GQL subset

`match.query` is a GQL query, but dialects support **only a single `subgraph`
predicate**. That is enough to select nodes, edges and sets by their attributes:

```yaml
match:
  query: 'subgraph node.kind == "service"'
```

Typical selection queries:

```yaml
match:
  query: 'subgraph edge.protocol exists'          # every edge with a protocol
match:
  query: 'subgraph node.replicas > 1'             # nodes with more than one replica
match:
  query: 'subgraph node.tier == "frontend"'       # attribute equality
```

Not supported in this subset: graph traversal, query pipelines, graph algebra,
and transformations. An unsupported or unevaluable query is a **matching
error**, not a conformance violation — the graph must not be reported as
conforming. The precise grammar is documented in
[`SUPPORTED_GQL_SUBSET.md`](../../../../spec/dialects/SUPPORTED_GQL_SUBSET.md); the full
language is in the [Query Guide](../../../../guides/GQL_GUIDE.md).

## 5. Projection: GSL values to JSON

JSON Schema validates JSON, so each selected element's **attributes** are
projected into a JSON object before validation. The projected instance contains
only attributes — never the element's id, representation or endpoints. Absent
attributes are omitted.

The value mappings are normative:

| GSL value | JSON projection | Example GSL | Projected JSON |
|---|---|---|---|
| String | JSON string | `name="Gateway"` | `"Gateway"` |
| Number | JSON number | `replicas=3` | `3` |
| Boolean | JSON Boolean | `active=true` | `true` |
| Empty | object with `$gslEmpty` | `enabled` | `{"$gslEmpty": true}` |
| Node reference | object with `$gslRef` | `owner=platform_team` | `{"$gslRef": "platform_team"}` |

A number is never turned into a string, and a string of digits is never turned
into a number.

## 6. Writing schemas against projected values (authoring patterns)

The single most common mistake is writing a schema against **GSL syntax**
instead of the **projected JSON**. These patterns cover the cases that trip
people up.

### 6.1 A plain scalar

```yaml
properties:
  replicas:
    type: integer
    minimum: 1
```

`replicas=3` projects to `3`, so `type: integer` matches.

### 6.2 A node reference

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

### 6.3 A flag that may be present or absent

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

### 6.4 A value that may be a string or a reference

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

### 6.5 A controlled vocabulary

Use `enum` for a fixed set of allowed values:

```yaml
properties:
  protocol:
    enum: [http, grpc, tcp]
```

### 6.6 The object as a whole

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

## 7. Validation outcomes

`gsl-dialect validate` reports one of three outcomes:

| Outcome | Meaning | CLI exit code |
|---|---|---|
| Conforming | Every selected element matched its schema. | `0` |
| Non-conforming | Validation completed; one or more violations. | `1` |
| Incomplete | Validation could not complete (invalid dialect, matching error, I/O or usage). | `2` |

A violation names the **definition**, the **element** and a **JSON Pointer**
into the projected instance:

```
Non-conforming: 1 violation(s) in model.gsl
  node gateway /replicas: minimum: got 0, want 1
```

## 8. Using the CLI

```bash
# Conforming → exit 0
gsl-dialect validate -d software-architecture.dialect.yaml -i model.gsl

# See the projected JSON each schema validates
gsl-dialect project -d software-architecture.dialect.yaml -i model.gsl

# Machine-readable report (provisional shape)
gsl-dialect validate -d software-architecture.dialect.yaml -i model.gsl --json
```

Both commands read the graph from **stdin** when `-i` is omitted. See
[`cmd/gsl-dialect/README.md`](../../../../cmd/gsl-dialect/README.md).

## 9. A complete example

The reference example lives in
[`examples/dialects/software-architecture/`](../../../../examples/dialects/software-architecture):
a dialect with a `Service` node definition, a `Dependency` edge definition and a
`Classification` set definition, plus a conforming `model.gsl`. Read it end to
end before authoring your own.

## 10. Checklist for authoring a dialect

1. Give the dialect a stable `id` and a `version`.
2. Define one element definition per concept you want to enforce.
3. Choose the `representation` that matches the concept's kind.
4. Write the selection query using the supported `subgraph` subset.
5. Write the `attributes` schema against **projected JSON**, not GSL syntax.
6. Use `gsl-dialect project` to inspect the projected instance while iterating.
7. Validate the reference model, then add mutations to confirm your schemas
   actually reject what they should.
