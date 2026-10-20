<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Recipe: Validate a Model Against Team Conventions

**Problem:** your team has conventions — services must have an owner and at
least one replica, dependencies must name a protocol — and you want the *model*
to hold you to them, without changing how you write GSL.

**At scale:** the reference example
[`examples/dialects/software-architecture/`](../../examples/dialects/software-architecture/)
does the same thing with node, edge and set definitions together.

## The model

A deliberately minimal architecture, saved as `model.gsl`:

```gsl
node platform_team [kind="team", name="Platform Team"]

node gateway [
  kind="service",
  name="Gateway",
  owner=platform_team,
  replicas=3
]
```

## The dialect

The conventions, written as a **dialect** (`software-architecture.dialect.yaml`).
It selects service nodes with a GQL query and validates their projected
attributes with JSON Schema:

```yaml
id: software-architecture
version: "1"
elements:
  Service:
    representation: node
    match:
      query: 'subgraph node.kind == "service"'
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

Note the authoring rule that trips people up: `owner=platform_team` is a **node
reference**, so the schema describes an object with a `$gslRef` — not a string.

## Run it

```bash
gsl-dialect validate -d software-architecture.dialect.yaml -i model.gsl
```

## The answer

```text
Conforming: dialect "software-architecture" version "1" (model.gsl)
```

Exit code `0`. Now break a convention — set `replicas=0` — and rerun:

```bash
gsl-dialect validate -d software-architecture.dialect.yaml -i model.gsl
```

```text
Non-conforming: 1 violation(s) in model.gsl
  node gateway /replicas: minimum: got 0, want 1
```

Exit code `1`. The diagnostic names the definition (`Service`), the element
(`gateway`) and a JSON Pointer into the projected attributes (`/replicas`).
Because `replicas=0` still *matches*, it is selected and then rejected — a
malformed element is never silently skipped.

## Projection, on demand

To see exactly what the schemas validate, print the projected JSON:

```bash
gsl-dialect project -d software-architecture.dialect.yaml -i model.gsl
```

```json
{
  "Service": {
    "gateway": {
      "kind": "service",
      "name": "Gateway",
      "owner": { "$gslRef": "platform_team" },
      "replicas": 3
    }
  }
}
```

## Next

- The full authoring guide: [Dialect Guide](../../guides/DIALECT_GUIDE.md).
- The reference dialect with node, edge and set definitions:
  [examples/dialects/software-architecture](../../examples/dialects/software-architecture/).
- The command reference: [gsl-dialect](../../cmd/gsl-dialect/README.md).
