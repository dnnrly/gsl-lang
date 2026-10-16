<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Conformance Suite

This directory holds the **shared, language-neutral conformance fixtures** for
GSL and GQL. They are the observable-behaviour contract: an implementation is
conformant when it agrees with the expected output for every fixture, and the
fixtures are independent of any particular implementation.

## Layout

```
tests/conformance/
  gql/        # GQL query fixtures (graph.gsl + query.gql + result.gsl)
    README.md # catalog, learning path and semantic notes
```

GSL parse/serialize conformance is currently covered by the Go parser and
round-trip tests (`implementations/go/*_test.go`), which assert the canonical
form guarantee `parse(serialize(parse(x))) == parse(x)` over the examples and
corpus. A dedicated `gsl/` fixture set may be added here as further
implementations appear; the intent is that any implementation can run the whole
suite from this directory alone.

## Fixture contract (GQL)

Every fixture is a directory containing exactly three files:

| File | Role |
|------|------|
| `graph.gsl` | Input graph (must parse with no errors or warnings). |
| `query.gql` | GQL query to run against the input graph. |
| `result.gsl` | Expected result, in canonical serialized form. |

Comparison is on **canonical serialized GSL**: an implementation serializes its
result graph and compares it byte-for-byte with `result.gsl`. Attribute values,
node and edge ordering, and set expansion must therefore match the specification
exactly.

## Running the suite

- Go reference implementation: `go test -v -run TestFixtures ./implementations/go/query`.
- The same fixtures seed the GQL fuzz targets
  (`implementations/go/query/fuzz_test.go`).
- The LSP server also loads them as a validity corpus
  (`lsp/gql_test.go`).

An implementation in another language runs the suite by walking the group
directories under `gql/`, applying each `query.gql` to its `graph.gsl`, and
comparing canonical output with `result.gsl`.

## Adding or changing fixtures

See [`gql/README.md`](gql/README.md) for the catalog, the numbered learning
path, per-group semantic notes, and the maintenance checklist. Keep the
`gql/README.md` catalog in sync whenever fixtures are added or changed.

## Normative references

- [`../../spec/gsl/SPEC.md`](../../spec/gsl/SPEC.md) — GSL language.
- [`../../spec/gql/QUERY_SPEC.md`](../../spec/gql/QUERY_SPEC.md) — GQL language.
