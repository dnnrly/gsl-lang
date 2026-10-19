<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Dialect Conformance Fixtures

This directory is the **shared, language-neutral dialect conformance suite** —
fixture-based tests for the dialect layer described by
[`../../spec/dialects/DIALECTS_SPEC.md`](../../spec/dialects/DIALECTS_SPEC.md).

Each fixture is a self-contained directory:

| File | Role |
|------|------|
| `dialect.yaml` | The dialect document under test (required). |
| `graph.gsl` | The input graph (required except when the dialect is rejected). |
| `expected.yaml` | The expected outcome, projection and violations. |

An implementation is conformant for a fixture when it agrees with
`expected.yaml`.

## `expected.yaml` contract

`expected.yaml` is deliberately minimal so that an implementation in any
language can satisfy it. Only `outcome` is required; the other fields are
asserted **only when present**, which lets a fixture pin exactly the behaviour
it is about.

```yaml
outcome: conforming | non-conforming | invalid-dialect
projection:                 # optional
  <Definition>:
    <element>: <projected attribute object>
violations:                 # optional
  - definition: <definition name>
    element: <element reference>
    location: <JSON Pointer into the projected attributes>
```

- **`outcome`** is the result of loading the dialect and, when it loads,
  validating the graph. `invalid-dialect` means the dialect is rejected
  before graph validation begins (spec §10.1). This suite has no
  `incomplete` outcome because the supported profile rejects unsupported
  match queries at load time; matching errors therefore surface as
  `invalid-dialect` (see "Known limitations").
- **`projection`** pins the deterministic result of matching and projecting
  (spec §7.2, §8), keyed by definition then by element reference. It is
  checked **independently of JSON Schema validation**, so a projection failure
  is attributable to projection and not to a schema (spec §14.3).
- **`violations`** pins the set of attribute constraint violations, as
  `(definition, element, location)` triples. `location` is the JSON Pointer
  into the projected attribute map (`""` for a violation at the root, such as
  a missing `required` property). The order is not significant.
- Element references: a node is its id, a set is its name, and an edge is its
  label if present, otherwise `from->to` (for example `api->payments`).

Projection values follow the normative mapping in spec §8.3: numbers are JSON
numbers, strings are JSON strings (never coerced to numbers), Empty values are
`{"$gslEmpty": true}` and node references are `{"$gslRef": "<node-id>"}`.

## Layout

```
tests/conformance/dialects/
  01-document/              → dialect document structure and rejection
  02-matching/              → match queries, representations and selection
  03-projection/            → GSL-to-JSON value projection
  04-attribute-validation/  → JSON Schema constraints on projected values
  05-outcomes/              → conforming / non-conforming semantics
```

## Fixtures

### 01-document — dialect document (spec §14.1)

| Fixture | Tests |
|---------|-------|
| `01_missing_id` | Missing required root field `id` |
| `02_missing_version` | Missing required root field `version` |
| `03_missing_elements` | Missing required root field `elements` |
| `04_elements_not_mapping` | `elements` is a sequence, not a mapping |
| `05_element_missing_representation` | Missing element field `representation` |
| `06_invalid_representation` | `representation` is not `node`/`edge`/`set` |
| `07_element_missing_match` | Missing element field `match` |
| `08_match_missing_query` | `match` present but has no `query` |
| `09_element_missing_attributes` | Missing element field `attributes` |
| `10_invalid_json_schema` | `attributes` is not a valid JSON Schema |
| `11_unknown_fields_ignored` | Unknown root and element fields are ignored |
| `12_empty_elements_mapping` | Empty `elements` mapping loads and conforms |
| `13_not_a_mapping` | The document is not a YAML mapping |

All `01-document` fixtures except `11` and `12` expect `invalid-dialect`.

### 02-matching — element matching (spec §14.2)

| Fixture | Tests |
|---------|-------|
| `01_single_node_match` | A `node` definition selecting one node |
| `02_no_match_conforming` | A valid query that selects nothing |
| `03_multiple_matches` | A query selecting several elements |
| `04_incidental_elements` | Edges in a node-query result are ignored by a node definition |
| `05_representation_edge` | An `edge` definition, keyed by `from->to` |
| `06_representation_set` | A `set` definition selects all declared sets |
| `07_unsupported_traversal` | Traversal is rejected at load |
| `08_invalid_syntax` | A malformed query is rejected at load |
| `09_pipeline_rejected` | A pipeline is rejected at load |

### 03-projection — GSL-to-JSON projection (spec §14.3)

| Fixture | Tests |
|---------|-------|
| `01_string_values` | Strings stay strings (including numeric-looking strings) |
| `02_number_values` | Numbers stay numbers without value changes |
| `03_boolean_values` | Booleans stay booleans |
| `04_empty_value` | Empty uses `{"$gslEmpty": true}` |
| `05_node_reference` | References use `{"$gslRef": "<id>"}` |
| `06_missing_omitted` | Absent attributes are omitted |
| `07_distinctions_preserved` | String `"true"`, boolean `true`, Empty and a reference stay distinct |

### 04-attribute-validation — JSON Schema constraints (spec §14.4)

| Fixture | Tests |
|---------|-------|
| `01_required_present` | Required attributes present → conforming |
| `02_required_missing` | Required attribute missing → violation at root |
| `03_wrong_type` | Wrong JSON type after projection |
| `04_numeric_minimum` | `minimum` lower bound |
| `05_numeric_maximum` | `maximum` upper bound (boundary value conforms) |
| `06_string_length` | `minLength` |
| `07_string_pattern` | `pattern` |
| `08_enum` | `enum` |
| `09_const` | `const` |
| `10_additional_properties_forbidden` | `additionalProperties: false` |
| `11_additional_properties_allowed` | `additionalProperties: true` |
| `12_empty_value_schema` | A schema constraining the Empty representation |
| `13_node_reference_schema` | A schema constraining the reference representation |
| `14_multiple_definitions_same_element` | Two definitions validate one element independently |

### 05-outcomes — outcome semantics (spec §14.5)

| Fixture | Tests |
|---------|-------|
| `01_empty_match_no_violations` | An empty match is not a violation |
| `02_violation_non_conforming` | An attribute violation → non-conforming |
| `03_conforming_does_not_imply_completeness` | Unmatched elements are not validated |
| `04_matching_error_rejected` | A matching error prevents a conforming result (here: load-time rejection) |

## Known limitations

- **Matching errors are folded into dialect rejection.** The supported profile
  rejects unsupported match queries when the dialect is loaded
  ([`SUPPORTED_GQL_SUBSET.md`](../../spec/dialects/SUPPORTED_GQL_SUBSET.md)), so
  there is no reachable `incomplete` outcome in this suite. An implementation
  that instead defers the failure to validation must still avoid reporting the
  graph as conforming (spec §10.4).
- **`set` definitions select every declared set.** Because a GQL `subgraph`
  carries all set declarations, the match predicate does not filter sets
  (`06_representation_set`). This is a known limitation recorded in the
  planning log, pending set-targeting predicates in GQL.
- **Negative numbers cannot appear in a fixture graph.** GSL numbers MUST NOT
  include a sign (GSL spec §2.3.2), so projection fixtures use non-negative
  values only.

## Running the suite

- Go reference implementation: `go test -v -run TestConformanceFixtures ./implementations/go/dialect`.
- TypeScript implementation:
  `npm test -w @gsl-lang/dialect` (or `make test-typescript`), which runs every
  fixture in `packages/dialect/test/conformance.test.ts`.
- Any other implementation walks the group directories under `dialects/`, and
  for each fixture loads `dialect.yaml`, (unless the outcome is
  `invalid-dialect`) parses `graph.gsl`, and compares the projected instances
  and violation set with `expected.yaml`.

## Adding or changing fixtures

1. Create a directory under the appropriate numbered group with
   `dialect.yaml`, `graph.gsl` (omit only for `invalid-dialect` fixtures) and
   `expected.yaml`.
2. Keep `expected.yaml` minimal — pin `outcome` always, and add only the
   `projection` and `violations` the fixture is about.
3. Verify it passes: `go test -run TestConformanceFixtures ./implementations/go/dialect`.
4. Add a row to the group table above with a one-line description.
5. If the fixture exposes a spec ambiguity, record it in the programme log
   rather than silently encoding a choice.

## Normative references

- [`../../spec/dialects/DIALECTS_SPEC.md`](../../spec/dialects/DIALECTS_SPEC.md)
  — the dialect specification.
- [`../../spec/dialects/SUPPORTED_GQL_SUBSET.md`](../../spec/dialects/SUPPORTED_GQL_SUBSET.md)
  — the supported match-query subset.
- [`../README.md`](../README.md) — the shared conformance suite contract.
