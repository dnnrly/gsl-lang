<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# @gsl-lang/dialect

TypeScript implementation of GSL dialect support, as specified in
[`spec/dialects/DIALECTS_SPEC.md`](../../../spec/dialects/DIALECTS_SPEC.md).
Depends on [`@gsl-lang/core`](../core/README.md) and
[`@gsl-lang/gql`](../gql/README.md); uses `ajv` for JSON Schema validation and
`yaml` to read dialect documents.

- **Status:** functional. Loading, matching, projection and validation are
  implemented and validated against every shared conformance fixture. Not
  published.
- **Licence:** Apache-2.0.

## Usage

```ts
import { parse } from "@gsl-lang/core";
import { load } from "@gsl-lang/dialect";

const dialect = load(`
id: software-architecture
version: "1"
elements:
  Service:
    representation: node
    match:
      query: 'subgraph node.kind == "service"'
    attributes:
      type: object
      required: [kind]
      properties:
        kind:
          const: service
`);

const { graph } = parse('node api [kind="service"]');
const report = dialect.validate(graph);
report.conforms; // true
```

## What is here

```
packages/dialect/
  src/document.ts     # load: YAML document parsing and schema compilation
  src/match.ts        # element matching via the supported GQL subset
  src/project.ts      # GSL-to-JSON attribute projection
  src/result.ts       # outcomes, violations and error types
  test/conformance.test.ts  # runs the shared fixtures
```

## Conformance

The package executes each fixture directory under
[`tests/conformance/dialects`](../../../tests/conformance/dialects/README.md):
it loads `dialect.yaml` (unless the expected outcome is `invalid-dialect`),
parses `graph.gsl`, and compares the projected instances and violation set with
`expected.yaml`, matching the Go reference implementation exactly.

The supported match-query subset is defined in
[`spec/dialects/SUPPORTED_GQL_SUBSET.md`](../../../spec/dialects/SUPPORTED_GQL_SUBSET.md):
a single `subgraph` predicate with no traversal, scope, pipeline or
graph-algebra stage.
