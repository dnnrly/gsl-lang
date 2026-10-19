<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# TypeScript implementation

This directory is an npm workspace containing TypeScript implementations of the
GSL language family:

- **`@gsl-lang/core`** — GSL parsing, graph construction and canonical
  serialization (`packages/core`).
- **`@gsl-lang/gql`** — GQL query parsing and execution, depending on
  `@gsl-lang/core` (`packages/gql`).
- **`@gsl-lang/dialect`** — dialect loading, element matching, GSL-to-JSON
  projection and JSON Schema conformance validation, depending on
  `@gsl-lang/core` and `@gsl-lang/gql` (`packages/dialect`).

All packages are ESM, written with strict TypeScript, and build to `dist/`.
They are validated against the shared, language-neutral conformance fixtures
under [`tests/conformance/`](../../tests/conformance/README.md); serialization
matches the Go reference implementation byte-for-byte.

- **Status:** functional (parser, serializer, query engine, dialect
  validator). No CLI or LSP. Not published.
- **Licence:** Apache-2.0.

## What is here

```
implementations/typescript/
  package.json        # npm workspace: build, typecheck, test
  tsconfig.base.json  # strict, NodeNext
  packages/core/      # @gsl-lang/core
  packages/gql/       # @gsl-lang/gql
  packages/dialect/   # @gsl-lang/dialect
```

## Commands

```bash
npm install
npm run build       # tsc for every package
npm run typecheck   # tsc --noEmit for every package
npm test            # build, then node --test for every package
```

## GSL example

```gsl
set critical

node api [team="platform"] @critical
node db

api -> db
```

## GQL example

```gql
subgraph node in @critical traverse out 1 | make node.reviewed = true where node.team == "platform"
```

## Relationship to the specification

[`spec/gsl/SPEC.md`](../../spec/gsl/SPEC.md) and
[`spec/gql/QUERY_SPEC.md`](../../spec/gql/QUERY_SPEC.md) are authoritative.
