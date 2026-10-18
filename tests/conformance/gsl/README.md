<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

---
description: GSL parse/serialize conformance fixtures. Language-neutral canonical-form corpus any implementation can run.
---

# GSL Conformance Fixtures

This directory holds the **shared, language-neutral GSL conformance suite**:
fixtures that pin the canonical serialized form of a parsed graph. They are the
observable-behaviour contract for the GSL parser/serializer, independent of any
particular implementation.

## Structure

```
tests/conformance/gsl/
  canonical/      → parse input.gsl, serialize, compare with expected.gsl
```

Each canonical fixture directory contains two files:

| File | Role |
|------|------|
| `input.gsl` | Source graph to parse. |
| `expected.gsl` | The exact canonical serialization of the parsed graph. |

Unlike the GQL suite, comparison here is on the **literal expected text**, not on
a re-serialization of `expected.gsl`. This is deliberate: it is what catches
canonical-form differences such as number formatting (e.g. `0.0000001` rather
than `1e-7`) and unquoted node references.

## Fixtures

Group `canonical/` — the canonical serialized form of a parsed graph.

| Fixture | Tests |
|---------|-------|
| `01-number-formatting` | Float/integer formatting matches the spec (no exponent notation, exact integer digits) |
| `02-node-references` | Node-reference attribute values serialize unquoted (`friend=B`) |
| `03-nested-structure` | Grouping and nested blocks re-serialize with canonical indentation |

## Running the suite

- Go reference implementation: `go test -run TestGSLConformance ./implementations/go`.
- TypeScript implementation:
  `npm test -w @gsl-lang/core` (or `make test-typescript`), which runs every
  fixture in `packages/core/test/conformance.test.ts`.
- Another implementation runs the suite by walking
  `canonical/**/{input.gsl,expected.gsl}`, serializing `parse(input.gsl)`, and
  comparing with `expected.gsl` (ignoring a trailing newline).

## Adding fixtures

1. Create a numbered fixture directory whose name describes the canonical-form
   rule under test.
2. Add `input.gsl` (must parse with no errors) and the exact `expected.gsl`
   canonical output.
3. Add the fixture to the table above.

## Normative references

- [`../../spec/gsl/SPEC.md`](../../spec/gsl/SPEC.md) — GSL language, including
  the canonical form.
- [`../../spec/gsl/GRAMMAR.md`](../../spec/gsl/GRAMMAR.md) — GSL grammar.
