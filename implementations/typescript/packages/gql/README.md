<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# @gsl-lang/gql

TypeScript implementation of the GSL Query Language, as specified in
[`spec/gql/QUERY_SPEC.md`](../../../spec/gql/QUERY_SPEC.md). Depends on
[`@gsl-lang/core`](../core/README.md).

- **Status:** functional. Parser and query engine are implemented and validated
  against every shared conformance fixture. Not published.
- **Licence:** Apache-2.0.

## Usage

```ts
import { parse } from "@gsl-lang/core";
import { runQuery } from "@gsl-lang/gql";

const { graph } = parse('node api [team="payments"]\\nnode db\\napi -> db');
const result = runQuery(graph, 'subgraph node.team == "payments"');
```

## What is here

```
packages/gql/
  src/lexer.ts        # tokenizer
  src/parser.ts       # hand-written recursive-descent parser
  src/predicates.ts   # predicate evaluation (attribute, set, parent, depth)
  src/expressions.ts  # subgraph, traversal, make, remove, collapse, algebra
  src/graph.ts        # mutable working graph used during query execution
  test/conformance.test.ts  # runs the shared fixtures
```

## Conformance

The package executes each fixture directory under
[`tests/conformance/gql`](../../../tests/conformance/README.md) — `graph.gsl`
plus `query.gql` — and compares the canonical serialization of the result with
`result.gsl`, matching the Go reference implementation exactly.
