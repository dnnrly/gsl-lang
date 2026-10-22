<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Dialect example: software architecture

This directory is the reference, end-to-end example of a GSL **dialect**: a
document that says how a graph should be shaped and what its projected
attributes must look like. It is the worked example from
[`spec/dialects/DIALECTS_SPEC.md`](../../../spec/dialects/DIALECTS_SPEC.md)
(section 13), extended with an edge definition and a set definition.

| File | Purpose |
|------|---------|
| `software-architecture.dialect.yaml` | The dialect: two node rules, one edge rule and two set definitions. |
| `model.gsl` | A small architecture model that conforms to the dialect. |

## The dialect

`software-architecture.dialect.yaml` declares:

- `nodes.rules[Service]` — nodes with `kind = "service"`. Required attributes
  `kind`, `name`, `owner` (a node reference) and `replicas` (`1..1000`).
- `edges.rules[Dependency]` — edges carrying a `protocol`, drawn from the
  controlled vocabulary `http`, `grpc`, `tcp`.
- `sets.definitions` — named definitions for the `critical` and `deprecated`
  classifications. Sets are selected by name, not by a query.

The `sets.policy` makes declaration strict: undefined sets are rejected, every
defined set must be present, and a set must have at least one member. Set
definitions here carry descriptions and guidance but no attribute schema;
attributes are optional.

A dialect never changes the GSL source syntax. It validates the **projected**
attributes of each matched element.

## Run it

```
# Conforming (exit code 0)
go run ../../cmd/gsl-dialect validate \
  -d software-architecture.dialect.yaml -i model.gsl

# The JSON instance each element's schema validates
go run ../../cmd/gsl-dialect project \
  -d software-architecture.dialect.yaml -i model.gsl

# Machine-readable report (provisional shape)
go run ../../cmd/gsl-dialect validate \
  -d software-architecture.dialect.yaml -i model.gsl --json
```

A non-conforming graph exits with code `1`; an unusable dialect (or any
incomplete validation) exits with code `2`. See
[`cmd/gsl-dialect/README.md`](../../../cmd/gsl-dialect/README.md).

## Tests

`examples/dialect_test.go` keeps the example honest:

- the committed model parses without warnings and validates as conforming;
- the projected `gateway` service, the `gateway->payments` dependency and both
  set definitions are asserted;
- in-memory mutations (a missing name, `replicas: 0`, a string owner, an
  unknown protocol) each produce the expected single violation.
