<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# @gsl-lang/core

TypeScript implementation of [GSL](../../../spec/gsl/SPEC.md): parsing,
graph construction, and canonical serialization.

- **Status:** functional. Parser, serializer and graph model are implemented
  and validated against the shared conformance fixtures. Not published.
- **Licence:** Apache-2.0.

## Usage

```ts
import { parse, serialize } from "@gsl-lang/core";

const { graph } = parse('node api @critical\napi -> db');
const text = serialize(graph);
```

`parse` returns `{ graph, warnings? }` and throws a `ParseError` on fatal
errors (for example, grouped edges on both sides or duplicate edge labels).
Warnings (such as implicit set creation) are returned rather than thrown.

## What is here

```
packages/core/
  src/model.ts      # graph model types (GslGraph, GslNode, GslEdge, GslSet)
  src/lexer.ts      # tokenizer
  src/parser.ts     # hand-written recursive-descent parser
  src/build.ts      # AST -> graph construction (merging, scopes, diagnostics)
  src/parse.ts      # parse() entry point
  src/serialize.ts  # canonical serialization
  test/*.test.ts    # node:test unit and round-trip checks
```

## Relationship to the specification

The model and rules mirror [`spec/gsl/SPEC.md`](../../../spec/gsl/SPEC.md) and
[`spec/gsl/GRAMMAR.md`](../../../spec/gsl/GRAMMAR.md); the specification is
authoritative. Serialization follows the canonical form rules and is verified
byte-for-byte against the Go reference implementation for every
`graph.gsl` file under [`tests/conformance/gql`](../../../tests/conformance/README.md).
