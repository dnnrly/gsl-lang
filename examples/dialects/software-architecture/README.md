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
| `software-architecture.dialect.yaml` | The dialect: three element definitions (Service, Dependency, Classification). |
| `model.gsl` | A small architecture model that conforms to the dialect. |

## The dialect

`software-architecture.dialect.yaml` declares:

- **Service** (`node`) — nodes with `kind = "service"`. Required attributes
  `kind`, `name`, `owner` (a node reference) and `replicas` (`1..1000`).
- **Dependency** (`edge`) — edges carrying a `protocol`, drawn from the
  controlled vocabulary `http`, `grpc`, `tcp`.
- **Classification** (`set`) — orthogonal tags such as `critical` and
  `deprecated`. Set selection is permissive; see
  [`SUPPORTED_GQL_SUBSET.md`](../../../spec/dialects/SUPPORTED_GQL_SUBSET.md).

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
  classifications are asserted;
- in-memory mutations (a missing name, `replicas: 0`, a string owner, an
  unknown protocol) each produce the expected single violation.
