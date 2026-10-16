<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# GSL for TypeScript (work in progress)

This directory is a **minimal TypeScript workspace** for a future GSL
implementation. It is deliberately a skeleton: it contains the implementation-
neutral model *types* and a build/typecheck/test setup, and nothing that
pretends to parse GSL.

- **Status:** model types only. There is **no parser, serializer, query
  engine, or CLI** here yet.
- **Package name:** `@gsl-lang/core` (provisional; not published, and not
  verified against any registry).
- **Licence:** Apache-2.0.

## What is here

```
implementations/typescript/
  package.json      # scripts: build, typecheck, test
  tsconfig.json     # strict, NodeNext, emits to dist/
  src/model.ts      # model types transcribed from spec/gsl/SPEC.md
  src/index.ts      # public entry point
  test/*.test.ts    # node:test checks for the model
```

## Commands

```bash
npm install
npm run typecheck   # tsc --noEmit
npm run build       # tsc -> dist/
npm test            # build, then node --test dist/test
```

## Relationship to the specification

The model types mirror the shapes in
[`spec/gsl/SPEC.md`](../../spec/gsl/SPEC.md); the specification is
authoritative. When a real parser is added, it should be validated against the
shared, language-neutral fixtures under
[`tests/conformance/`](../../tests/conformance/README.md).
