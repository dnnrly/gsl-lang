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
  - definition: <rule label or set name>
    element: <element reference>
    location: <JSON Pointer into the projected attributes>
```

- **`outcome`** is the result of loading the dialect and, when it loads,
  validating the graph. `invalid-dialect` means the dialect is rejected
  before graph validation begins (spec §11.1). This suite has no
  `incomplete` outcome because the supported profile rejects unsupported
  match queries at load time; matching errors therefore surface as
  `invalid-dialect`.
- **`projection`** pins the deterministic result of matching and projecting
  (spec §8), keyed by rule label or set name, then by element reference. It is
  checked **independently of JSON Schema validation**, so a projection failure
  is attributable to projection and not to a schema (spec §10.3).
- **`violations`** pins the set of violations, as
  `(definition, element, location)` triples. `location` is the JSON Pointer into
  the projected attribute map (`""` for a violation at the root, such as a
  missing `required` property, and for every set policy violation). The order is
  not significant.
- Element references: a node is its id, a set is its name, and an edge is its
  label if present, otherwise `from->to` (for example `api->payments`).
- Rule labels: a named rule uses its `name`; an unnamed rule uses its section
  position, for example `nodes.rules[0]` or `edges.rules[1]`. Set definitions
  use the GSL set name.

Projection values follow the normative mapping in spec §9: numbers are JSON
numbers, strings are JSON strings (never coerced to numbers), Empty values are
`{"$gslEmpty": true}` and node references are `{"$gslRef": "<node-id>"}`.

## Layout

```
tests/conformance/dialects/
  01-document/              → dialect document structure and rejection
  02-matching/              → node and edge matching and selection
  03-projection/            → GSL-to-JSON value projection
  04-attribute-validation/  → JSON Schema constraints on projected values
  05-outcomes/              → conforming / non-conforming semantics
  06-sets/                  → set definitions and declaration policy
```

## Fixtures

### 01-document — dialect document (spec §5, §11)

| Fixture | Tests |
|---------|-------|
| `01_missing_id` | Missing required root field `id` |
| `02_missing_version` | Missing required root field `version` |
| `03_sections_optional` | Omitted `nodes`, `edges` and `sets` sections are valid |
| `04_nodes_not_mapping` | `nodes` is a sequence, not a mapping |
| `05_rule_missing_match` | Missing rule field `match` |
| `06_rules_not_sequence` | `rules` is a mapping, not a sequence |
| `07_rule_missing_attributes` | Missing rule field `attributes` |
| `08_rule_match_not_string` | `match` given as a non-string scalar is rejected |
| `09_match_explicit_prefix` | An explicit element prefix is rejected |
| `10_invalid_json_schema` | `attributes` is not a valid JSON Schema |
| `11_unknown_fields_ignored` | Unknown root, section and rule fields are ignored |
| `12_empty_sections` | Empty sections load and conform |
| `13_not_a_mapping` | The document is not a YAML mapping |
| `14_id_not_a_string` | `id` given as a non-string scalar (`!!int`) is rejected, not coerced |
| `15_version_not_a_string` | `version` given as a non-string scalar (`!!int`) is rejected, not coerced |
| `16_sets_not_mapping` | `sets` is a sequence, not a mapping |
| `17_policy_not_boolean` | A policy value is a string, not a boolean |
| `18_definitions_not_mapping` | `sets.definitions` is a sequence, not a mapping |
| `19_invalid_set_name` | A definition key is not a GSL set identifier |
| `20_duplicate_set_definition` | The same set is defined twice |

All `01-document` fixtures except `03`, `11` and `12` expect `invalid-dialect`.

### 02-matching — node and edge matching (spec §8)

| Fixture | Tests |
|---------|-------|
| `01_single_node_match` | A node rule selecting one node |
| `02_no_match_conforming` | A valid query that selects nothing |
| `03_multiple_matches` | A query selecting several nodes |
| `04_incidental_elements` | Edges in a node-query result are ignored by a node rule |
| `05_representation_edge` | An edge rule, keyed by `from->to` |
| `06_set_definitions` | Defined, present sets are selected by name |
| `07_unsupported_traversal` | Traversal is rejected at load |
| `08_invalid_syntax` | A malformed query is rejected at load |
| `09_pipeline_rejected` | A pipeline is rejected at load |
| `10_explicit_subgraph_rejected` | The `subgraph` wrapper is rejected at load |

### 03-projection — GSL-to-JSON projection (spec §9)

| Fixture | Tests |
|---------|-------|
| `01_string_values` | Strings stay strings (including numeric-looking strings) |
| `02_number_values` | Numbers stay numbers without value changes |
| `03_boolean_values` | Booleans stay booleans |
| `04_empty_value` | Empty uses `{"$gslEmpty": true}` |
| `05_node_reference` | References use `{"$gslRef": "<id>"}` |
| `06_missing_omitted` | Absent attributes are omitted |
| `07_distinctions_preserved` | String `"true"`, boolean `true`, Empty and a reference stay distinct |

### 04-attribute-validation — JSON Schema constraints (spec §10)

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
| `14_multiple_definitions_same_element` | Two rules validate one element independently |

### 05-outcomes — outcome semantics (spec §10, §11)

| Fixture | Tests |
|---------|-------|
| `01_empty_match_no_violations` | An empty match is not a violation |
| `02_violation_non_conforming` | An attribute violation → non-conforming |
| `03_conforming_does_not_imply_completeness` | Unmatched elements are not validated |
| `04_matching_error_rejected` | An unsupported match prevents a conforming result (here: load-time rejection) |

### 06-sets — set definitions and policy (spec §7)

| Fixture | Tests |
|---------|-------|
| `01_defaults_permit_undefined_and_empty` | Defaults allow undefined and empty sets |
| `02_undefined_set_rejected` | `allow_undefined: false` rejects undeclared sets |
| `03_require_all` | `require_all: true` requires every definition to be present |
| `04_allow_empty_false` | `allow_empty: false` rejects a present set with no members |
| `05_implicit_set_counts_as_present` | An implicitly created, non-empty set satisfies strict policy |
| `06_set_attributes_validated` | A set definition's `attributes` schema is validated |
| `07_optional_attributes` | A set definition without `attributes` still documents the set |

## Known limitations

- **Matching errors are folded into dialect rejection.** The supported profile
  rejects unsupported match queries when the dialect is loaded
  ([`SUPPORTED_GQL_SUBSET.md`](../../spec/dialects/SUPPORTED_GQL_SUBSET.md)), so
  there is no reachable `incomplete` outcome in this suite. An implementation
  that instead defers the failure to validation must still avoid reporting the
  graph as conforming (spec §11.4).
- **Negative numbers cannot appear in a fixture graph.** GSL numbers MUST NOT
  include a sign (GSL spec §2.3.2), so projection fixtures use non-negative
  values only.
- **Implicit-set fixtures produce a GSL warning.** Implicit set creation is a
  warning in GSL, not an error; fixture runners MUST ignore warnings when
  parsing `graph.gsl`.

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
