<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# GSL and GQL Specifications

This directory holds the **language-neutral, normative** specifications for
GSL and its query language GQL. They describe observable behaviour only; they
are independent of any implementation (Go, TypeScript, or otherwise).

If a document here conflicts with any other material (implementation, guide,
README), the specification is authoritative.

## GSL — the graph specification language

| Document | Contents |
|----------|----------|
| [`gsl/SPEC.md`](gsl/SPEC.md) | Normative specification: lexical structure, core model, nodes, edges, edge scoping, sets, attribute semantics, declaration merging, canonicalisation, errors, conformance. |
| [`gsl/GRAMMAR.md`](gsl/GRAMMAR.md) | Formal EBNF grammar for GSL. |

## GQL — the query language

| Document | Contents |
|----------|----------|
| [`gql/QUERY_SPEC.md`](gql/QUERY_SPEC.md) | Normative specification: subgraph extraction, predicates, traversal, `make`/`remove`/`collapse`, named graphs, graph algebra, pipeline composition, result semantics. |
| [`gql/QUERY_GRAMMAR.md`](gql/QUERY_GRAMMAR.md) | Formal EBNF grammar for GQL. |

## Dialects — domain conventions on top of GSL

| Document | Contents |
|----------|----------|
| [`dialects/DIALECTS_SPEC.md`](dialects/DIALECTS_SPEC.md) | Draft specification: dialect documents (YAML), element definitions, GQL-subset matching, GSL-to-JSON projection, JSON Schema attribute validation, conformance outcomes and diagnostics. |
| [`dialects/SUPPORTED_GQL_SUBSET.md`](dialects/SUPPORTED_GQL_SUBSET.md) | Companion profile: the exact GQL query subset permitted in a dialect `match.query`, and the forms that MUST be rejected. |

## Normative vs explanatory

The specifications above are **normative**: they define what an implementation
must accept, reject, and produce. Everything else in the repository is
explanatory:

- [`../guides/`](../guides/) — self-contained, LLM/agent-oriented language
  guides and the query tutorial (also embedded by the reference CLI `ai`
  command).
- [`../docs/`](../docs/) — the human learning journey (getting started,
  concepts, tutorials, cookbook).
- [`../README.md`](../README.md) — the project overview.

## Status

Both specifications are drafts while the project is pre-1.0. See
[`../CONTRIBUTING.md`](../CONTRIBUTING.md) for the versioning and
compatibility policy.

## Conformance

Observable behaviour described here is exercised by the shared, language-neutral
fixtures under [`../tests/conformance/`](../tests/conformance/). See
[`../tests/conformance/README.md`](../tests/conformance/README.md) for the
fixture format and how a new implementation can run the suite.
