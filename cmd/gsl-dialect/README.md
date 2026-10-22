<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# gsl-dialect

Validate GSL graphs against [dialect documents](../../spec/dialects/DIALECTS_SPEC.md),
and project graph attributes into the JSON instances a dialect validates.

## Usage

```
gsl-dialect validate -d <dialect.yaml> [-i <graph.gsl>] [--json]
gsl-dialect project  -d <dialect.yaml> [-i <graph.gsl>]
gsl-dialect version
```

- `-d, --dialect` — the dialect document (required).
- `-i, --input` — the input GSL graph; read from stdin when omitted.
- `--json` — `validate` only: emit a provisional machine-readable report.

Both commands read the graph from stdin when `-i` is omitted:

```
cat model.gsl | gsl-dialect validate -d software-architecture.dialect.yaml
```

## `validate`

Loads the dialect, matches and projects the graph's attributes, validates them
against the dialect's JSON Schemas, and reports the outcome.

Text output on success:

```
Conforming: dialect "software-architecture" version "0.4.0-draft" (model.gsl)
```

Text output on violations:

```
Non-conforming: 1 violation(s) in model.gsl
  node gateway /replicas: minimum: got 0, want 1
```

The `--json` report is a provisional shape pending a standardised diagnostic
contract (see `spec/dialects/DIALECTS_SPEC.md` §16.4):

```json
{
  "dialect": { "id": "software-architecture", "version": "0.4.0-draft" },
  "graph": "model.gsl",
  "outcome": "non-conforming",
  "violations": [
    {
      "definition": "Service",
      "representation": "node",
      "element": "gateway",
      "location": "/replicas",
      "message": "minimum: got 0, want 1"
    }
  ]
}
```

## `project`

Loads the dialect and prints the projected attribute map for every matched
element, keyed by definition and element reference. Each reference maps to a
**list** of projected instances, so two elements that share a reference (for
example two unlabelled edges between the same endpoints) are both reported
rather than one overwriting the other. This is the JSON instance a dialect's
schemas validate; it is useful for debugging projection independently of JSON
Schema validation.

```json
{
  "Service": {
    "gateway": [
      {
        "kind": "service",
        "owner": { "$gslRef": "platform_team" },
        "replicas": 3,
        "enabled": { "$gslEmpty": true }
      }
    ]
  }
}
```

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | The graph conforms to the dialect. |
| `1` | The graph does not conform: validation completed and found violations. |
| `2` | Validation could not be completed: invalid dialect, matching error, or an I/O/usage error. |

The distinction between codes `1` and `2` matches the specification's
separation of a non-conforming graph from incomplete validation
(`DIALECTS_SPEC.md` §10.2, §10.4).

## Building

```
make build      # writes ./tmp/gsl-dialect
go build ./cmd/gsl-dialect
```

## Tests

```
go test ./cmd/gsl-dialect
make test-integration
```
