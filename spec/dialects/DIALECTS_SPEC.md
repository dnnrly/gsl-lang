# GSL Dialect Specification

**Status:** Design draft
**Version:** `0.3.0-draft`
**Target:** GSL v1.0 draft
**Schema standard:** JSON Schema Draft 2020-12
**Normative companion:** [SUPPORTED_GQL_SUBSET.md](SUPPORTED_GQL_SUBSET.md)
**Reference implementation:** `implementations/go/dialect` (Go); conformance fixtures in `tests/conformance/dialects/`

## 1. Purpose

A GSL dialect defines conventions for representing a particular domain using the Graph Specification Language (GSL).

GSL provides a general-purpose graph representation. A dialect describes how graph elements should be used to represent concepts in a domain, which elements it recognises, and what constraints apply to their attributes.

Examples of potential dialects include:

- Software architecture: services, databases, APIs and their relationships.
- State machines: states, transitions, events and conditions.
- Financial relationships: accounts, obligations, transactions and ownership.

A dialect serves three purposes:

1. **Guidance:** help humans and language models construct and interpret GSL consistently.
2. **Validation:** enable tools to check whether selected graph elements conform to the dialect.
3. **Shared semantics:** establish predictable conventions for interpreting domain-specific graph representations.

A dialect does not change the underlying GSL language or restrict which graphs GSL can represent.

## 2. Design principles

The specification follows these principles, in priority order:

1. **Predictable interpretation:** dialects should make the intended meaning of graph structures clear.
2. **Consistent representation:** equivalent domain concepts should be represented consistently.
3. **Faithful representation:** the graph should preserve the distinctions and relationships that matter in the domain.

Additional principles:

- Reuse JSON Schema for attribute constraints rather than inventing a parallel validation language.
- Reuse GQL for selecting graph elements rather than introducing a separate matching language.
- Separate normative constraints from descriptive guidance.
- Make GSL-specific behaviour explicit and small in scope.
- Report validation failures clearly and distinguish them from failures to execute validation.
- Do not infer missing domain intent from graph structure.
- Do not confuse conformance with completeness, correctness of real-world facts, or overall graph quality.

## 3. Scope

### 3.1 Included

This version defines:

- The structure of a dialect document.
- Definitions of domain-specific graph element types.
- Matching graph elements using a restricted GQL subset.
- Validation of attributes on matched elements.
- Projection of GSL values into JSON values.
- Compilation or assembly of JSON Schemas for validation.
- Conformance outcomes and diagnostic requirements.
- Dialect versioning and conformance testing expectations.

### 3.2 Excluded

This version does not define:

- Graph-wide constraints.
- Validation of relationships between different matched elements.
- Validation that node references point to nodes of a particular dialect type.
- Automatic classification of unmatched graph elements.
- Inheritance or composition between dialects.
- Dialect dependency resolution.
- Arbitrary graph transformations.
- General-purpose extensions to JSON Schema.
- A new query language independent of GQL.

These exclusions may be revisited in later versions where concrete requirements justify them.

## 4. Terminology

**GSL graph:** A graph expressed using GSL.

**Dialect:** A document defining conventions and constraints for a domain.

**Element definition:** A named definition within a dialect describing a category of graph elements, its matching rule and its attribute constraints.

**Representation:** The GSL element kind to which an element definition applies: `node`, `edge` or `set`.

**Match query:** A GQL query used to select graph elements for validation.

**Projected attribute map:** A JSON object containing the attributes of one matched GSL element after conversion into JSON-compatible values.

**Dialect compiler:** An implementation component that validates and interprets a dialect document and prepares the JSON Schemas used for attribute validation.

**Conformance:** The result of checking the elements selected by a dialect against its declared constraints.

**Guidance:** Descriptive advice for humans and language models. Guidance is not, by itself, a machine-enforceable constraint.

## 5. Dialect document structure

A dialect document MUST be valid YAML and MUST have a mapping as its root value.

The following fields are defined at the root:

| Field | Required | Description |
|---|---|---|
| `id` | Yes | Stable identifier for the dialect |
| `version` | Yes | Version of the dialect definition |
| `title` | No | Human-readable name |
| `description` | No | Description of the dialect's purpose |
| `guidance` | No | General advice for constructing and interpreting graphs |
| `elements` | Yes | Mapping of element-definition names to definitions |

The fields `id`, `version` and `elements` MUST be present and valid.

Unknown fields MUST be ignored by implementations of this version. Implementations MUST NOT reject an otherwise valid dialect solely because it contains an unknown field.

This permissive behaviour allows dialect documents to carry additional information without immediately breaking older implementations. It also means that misspelled field names may go unnoticed; authors should validate dialect documents and review diagnostics carefully.

### 5.1 Identifier

`id` MUST be a non-empty string that identifies the dialect.

The identifier SHOULD remain stable across compatible revisions of the dialect. It SHOULD change only when the dialect is intentionally treated as a different dialect.

The precise identifier syntax and namespace policy remain to be standardised.

### 5.2 Version

`version` MUST be a non-empty string identifying the version of the dialect definition.

Versioning applies to the dialect, not to GSL or GQL.

A version change does not itself establish compatibility. The meaning of compatibility between dialect versions is specified in Section 12.

The exact version syntax and compatibility policy remain provisional.

### 5.3 Elements

`elements` MUST be a mapping containing zero or more named element definitions.

Each entry maps an element-definition name to its definition. The name is a dialect-local identifier used for readability and diagnostics; it does not create a new GSL element type.

An implementation MUST reject a dialect if `elements` is missing or is not a mapping.

An empty `elements` mapping is valid and selects no elements.

## 6. Element definitions

Each element definition MUST contain the following fields:

- `representation`
- `match`
- `attributes`

It MAY also contain:

- `description`
- `guidance`

Unknown fields MUST be ignored.

### 6.1 Representation

`representation` MUST be one of:

- `node`
- `edge`
- `set`

It identifies which kind of GSL element the definition describes and which elements from a match result are eligible for validation.

For example:

```yaml
representation: node
```

A match query may return a graph containing several kinds of elements. Only elements of the declared representation are selected for validation by that definition. Other kinds of elements in the result are incidental and MUST be ignored.

The representation field MUST NOT be interpreted as a claim that GSL itself has a domain-specific type system.

### 6.2 Match

`match` MUST be a mapping containing a `query` field.

The value of `query` MUST be a non-empty string containing a supported match expression.

A match expression is written as an **attribute predicate without an element prefix**. The `subgraph` wrapper and the element namespace are both implicit and are supplied by the definition's `representation` (Section 6.1). An implementation MUST expand the expression by prefixing every attribute reference with the representation before evaluation: the two queries below are equivalent, and the second is the canonical expansion of the first.

```yaml
representation: node
match:
  query: 'kind == "service"'
```

```yaml
representation: node
match:
  query: 'subgraph node.kind == "service"'
```

This keeps dialect match expressions focused on the question dialects actually ask — *which elements does this rule apply to?* — and removes the possibility of a mismatch between the namespace in the query and the definition's representation.

The expansion is exactly the `subgraph` form of the full GQL grammar with the representation used as the element prefix. Because the expansion is defined this way, the match expression is a strict subset of GQL and introduces no new syntax; later revisions may accept the explicit `subgraph` form unchanged.

The GQL `subgraph` grammar has no `set.` element prefix: it admits only `node.` and `edge.` (Section 7.5). A `representation: set` definition therefore cannot be expanded with its representation as the prefix. For such a definition the `query` is still REQUIRED to be a well-formed in-subset attribute predicate, but it is **not evaluated**: a `set` definition selects every set in the input graph, and the predicate has no filtering effect (Section 7.5). An implementation MUST validate the predicate's syntax and subset membership and MUST NOT branch on its value when selecting sets.

An implementation MUST reject, as an invalid dialect, a `query` that:

- is empty;
- is not a single attribute predicate (for example it contains a pipeline `|`, traversal, graph algebra, a graph transformation, or more than one top-level predicate);
- contains an explicit element prefix (`node.`, `edge.` or `set.`) rather than a bare attribute name; or
- otherwise does not conform to the supported subset in [`SUPPORTED_GQL_SUBSET.md`](SUPPORTED_GQL_SUBSET.md).

This example is illustrative; the exact expression MUST conform to the supported subset implemented by the validator.

Matching determines which graph elements are subject to the definition's constraints. It does not establish that those elements are valid.

A matched element MUST remain selected for validation even if its attributes violate the associated JSON Schema. Implementations MUST NOT silently exclude malformed elements from validation.

The supported query subset is defined in Section 7.

### 6.3 Attributes

`attributes` MUST contain a valid JSON Schema Draft 2020-12 schema.

The schema MUST validate the projected attribute map of one matched element. It MUST NOT be applied to an object containing the element's identity, representation, endpoints, or other graph metadata.

For example:

```yaml
attributes:
  type: object
  required:
    - name
    - replicas
  properties:
    name:
      type: string
      minLength: 1
    replicas:
      type: integer
      minimum: 1
```

The schema is ordinary JSON Schema. Dialect authors SHOULD use standard JSON Schema keywords for constraints whenever possible.

Implementations MUST reject a dialect if an `attributes` schema is invalid under Draft 2020-12 or cannot be compiled for validation.

The following `$ref` policy applies. A schema MAY use `$ref` to targets defined within the same dialect document (for example `#/$defs/...`); such internal references MUST resolve during compilation. A schema MUST NOT use `$ref` to external `http` or `file` targets; an implementation MUST reject such a dialect at load time and MUST NOT silently resolve references using arbitrary network or filesystem access. The `$ref` policy remains provisional and may be extended.

### 6.4 Description and guidance

`description` is a human-readable explanation of the element definition.

`guidance` provides advice for constructing or interpreting graph elements of that definition.

These fields MUST NOT alter matching, projection, or validation behaviour. Implementations MUST NOT treat prose in either field as an executable constraint.

Guidance may explain, for example, when a concept should be represented as a node rather than an edge, or how a language model should choose meaningful attribute names.

## 7. Element matching

### 7.1 Query language

GQL is the query language for selecting elements. This specification does not define a separate dialect matching language: a `match.query` is expanded into a full GQL `subgraph` expression (Section 6.2) and evaluated with GQL semantics.

A match expression is a single **attribute predicate** whose element prefix is supplied by the definition's `representation`. This is the initial supported subset.

The initial subset MUST NOT include:

- Graph traversal.
- Query pipelines.
- Graph algebra.
- Graph transformations.
- An explicit element prefix in the match expression.
- Other unsupported query forms.

The exact grammar and expression semantics of the supported subset are documented normatively in the companion profile [`SUPPORTED_GQL_SUBSET.md`](SUPPORTED_GQL_SUBSET.md), which defines the accepted predicate forms, the expansion rule, and the observable distinction between an out-of-subset query and a syntax error.

Implementations MUST reject unsupported query forms. They MUST NOT approximate an unsupported query with different semantics.

### 7.2 Matching process

For each element definition, an implementation MUST:

1. Expand its match expression into a `subgraph` predicate using the definition's representation as the element prefix (Section 6.2).
2. Evaluate the expanded query against the input graph.
3. Identify the elements returned by the query.
4. Select only elements whose representation matches the definition's `representation`.
5. Validate the attributes of each selected element against the definition's schema.

Because the expansion fixes the element prefix to the representation, steps 1 and 4 agree by construction: a match expression cannot select elements of a kind other than its definition's representation.

An element may be selected by more than one element definition. There is no one-to-one relationship between graph elements and dialect definitions.

If the same element is selected by multiple definitions, each applicable definition MUST be validated independently.

### 7.3 Empty match results

A valid query that selects no eligible elements is not an error.

The implementation MUST continue validation of other definitions and MUST NOT report a conformance violation solely because a query has no matches.

Consequently, a graph may conform to a dialect even when none of its elements match any definition. Conformance is defined over selected elements, not over an expectation that the graph contains particular domain concepts.

### 7.4 Matching errors

An invalid, unsupported or unevaluable query MUST produce a matching error.

A matching error is distinct from an attribute constraint violation. If a matching error prevents the validator from completing the required checks, the graph MUST NOT be reported as conforming.

### 7.5 Set selection in the initial subset

The initial `subgraph` subset operates on nodes and edges; it has no predicate that targets a specific named set, and the `subgraph` grammar admits only `node.` and `edge.` element prefixes, not `set.`. A `subgraph` expression therefore copies every declared set of the input graph into its result, regardless of attributes.

As a consequence, an element definition with `representation: set` selects **all** sets present in the input graph. Its `match.query` MUST still be a well-formed in-subset attribute predicate (Section 6.2), but the predicate is validated for syntax and subset membership only and MUST NOT be evaluated to filter sets. Dialect authors SHOULD write permissive set schemas under the initial subset, and MUST NOT rely on set selection to discriminate one set from another. This restriction is a property of the current GQL engine, not of dialects, and may be relaxed if GQL gains set-targeting predicates (see [`SUPPORTED_GQL_SUBSET.md`](SUPPORTED_GQL_SUBSET.md)).

## 8. GSL-to-JSON projection

### 8.1 Purpose

JSON Schema validates JSON instances, not GSL syntax or graph structures directly.

Implementations MUST therefore project the attributes of each selected element into a JSON object before validating them.

Projection MUST be deterministic and MUST preserve the semantic distinctions defined by GSL. It MUST NOT silently coerce one GSL value type into another.

The projection is semantic, not a reconstruction of the original GSL source text.

### 8.2 Attribute map

The projected instance MUST be a JSON object whose properties correspond to the attributes of the selected element.

An attribute that is absent in GSL MUST be omitted from the JSON object.

An explicitly present GSL `Empty` value MUST NOT be treated as an absent attribute, JSON `null`, or Boolean `true`.

### 8.3 Value mappings

The following mappings are normative:

| GSL value type | JSON projection |
|---|---|
| String | JSON string |
| Number | JSON number |
| Boolean | JSON Boolean |
| Empty | `{"$gslEmpty": true}` |
| Node reference | `{"$gslRef": "<node-id>"}` |

The node-reference representation is deliberately explicit so that a node reference is distinguishable from an ordinary string.

#### String

A GSL string MUST be projected as a JSON string.

Implementations MUST NOT convert a string containing numeric characters into a JSON number.

#### Number

A GSL number MUST be projected as a JSON number, preserving its numeric value.

Implementations MUST NOT silently round a value or change its meaning during projection. GSL numbers project to JSON numbers; because JSON Schema's `integer` type accepts any number with a zero fractional part, a dialect MAY use `type: integer` for attributes that are always whole. The broader numeric interoperability and precision requirements remain to be specified, particularly for implementations using different numeric representations.

#### Boolean

A GSL Boolean MUST be projected as the corresponding JSON Boolean, `true` or `false`.

#### Empty

An explicitly present GSL `Empty` value MUST be projected as:

```json
{"$gslEmpty": true}
```

This representation preserves the distinction between an empty attribute value and a Boolean value. JSON Schema can constrain it using ordinary object, property, constant and required-property constraints.

#### Node reference

A GSL node reference MUST be projected into an explicit JSON representation that preserves the referenced node's identifier.

The proposed representation is:

```json
{"$gslRef": "platform_team"}
```

The projection MUST NOT replace a node reference with the referenced node's attributes or otherwise expand it into a nested graph.

GSL node references are available only where permitted by the GSL specification. This dialect specification does not introduce edge or set references.

### 8.4 Example

Given the following illustrative GSL:

```gsl
node platform_team [kind="team", active=true]

node gateway [
  kind="service",
  owner=platform_team,
  replicas=3,
  enabled,
  description="Public entry point"
]
```

The projected attribute map for `gateway` is:

```json
{
  "kind": "service",
  "owner": {
    "$gslRef": "platform_team"
  },
  "replicas": 3,
  "enabled": {
    "$gslEmpty": true
  },
  "description": "Public entry point"
}
```

The `owner` value remains a node reference, `enabled` remains an explicitly present Empty value, and `description` remains a string.

### 8.5 Schema constraints on projected values

Dialect authors MUST write their attribute schemas against the projected JSON representation, not against GSL source syntax.

For example, a schema that accepts only strings will reject a projected node reference because a node reference is represented by an object. A schema that intends to accept references MUST explicitly describe the reference representation.

Implementations SHOULD provide documentation and examples for common GSL value types so that dialect authors do not have to rediscover these mappings.

## 9. JSON Schema compilation and validation

### 9.1 Architecture

Implementations MUST use JSON Schema Draft 2020-12 for attribute validation.

The implementation SHOULD use an established JSON Schema validator rather than implement JSON Schema validation independently.

The implementation is responsible for GSL-specific behaviour, including:

- Parsing and validating the dialect document.
- Interpreting element definitions.
- Evaluating supported match queries.
- Selecting eligible graph elements.
- Projecting GSL attribute values into JSON.
- Preparing the applicable JSON Schemas.
- Mapping validation failures into dialect conformance diagnostics.

The JSON Schema library is responsible for evaluating the compiled schemas against projected JSON instances.

### 9.2 Compilation

For each valid element definition, the implementation MUST prepare a schema that validates the definition's `attributes` schema against projected attribute maps.

Because `attributes` is already JSON Schema, compilation SHOULD preserve standard JSON Schema semantics. It need not translate each ordinary constraint into a new representation.

The implementation MAY assemble, wrap or compile the schema as required by its chosen validator, but it MUST NOT change the meaning of valid JSON Schema constraints.

The implementation MUST NOT invent custom JSON Schema keywords for constraints that standard JSON Schema can express.

### 9.3 Validation isolation

Each selected element MUST be validated independently against the applicable element definition.

The validator MUST NOT implicitly combine the schemas of different definitions or assume that an element can match only one definition.

The schema instance MUST contain only the selected element's projected attributes. Graph-wide information MUST NOT be added to the instance unless a future version explicitly defines such behaviour.

### 9.4 Additional attributes

The behaviour for attributes not explicitly listed in a schema is controlled by JSON Schema.

For example:

- `additionalProperties: false` rejects undeclared properties.
- `additionalProperties: true` permits them.
- `patternProperties` can constrain properties matching a pattern.

The dialect specification MUST NOT impose a separate default policy for additional attributes.

### 9.5 Standard JSON Schema semantics

Implementations MUST follow the semantics of JSON Schema Draft 2020-12 for supported keywords.

A dialect is invalid if its schema is invalid under that specification or if the implementation cannot support a required feature within its declared conformance profile.

Implementations MUST report such a failure rather than silently ignore a recognised constraint.

## 10. Conformance outcomes

Validation MUST distinguish three principal outcomes.

### 10.1 Invalid dialect

The dialect is invalid if its structure, required fields, recognised field values, match queries or attribute schemas are invalid.

An invalid dialect MUST be rejected before graph validation begins.

This version uses all-or-nothing dialect loading. Implementations MUST NOT partially apply a malformed dialect and report the resulting graph as conforming.

### 10.2 Non-conforming graph

A graph is non-conforming if the dialect is valid, the required validation completes, and one or more matched elements violate their applicable constraints.

Examples include:

- A required attribute is missing.
- An attribute has the wrong JSON type after projection.
- A numeric value falls outside a permitted range.
- A property is rejected by `additionalProperties: false`.

A non-conforming graph may still be a valid GSL graph. Dialect conformance is separate from GSL language validity.

### 10.3 Conforming graph

A graph is conforming if the dialect is valid, all required matching and validation completes, and no applicable constraint violations are found.

Conformance means only that the matched elements satisfy the constraints defined by the dialect.

It does not establish that:

- The graph is complete.
- Every graph element belongs to the dialect.
- Every domain concept that should be represented is present.
- The graph accurately reflects the real world.
- The graph is free from contradictions not expressed by the dialect.

### 10.4 Incomplete validation

If matching or validation fails in a way that prevents required checks from completing, the implementation MUST NOT report the graph as conforming.

Implementations SHOULD distinguish incomplete validation from a completed validation that found violations.

The precise external result model for incomplete validation remains to be finalised. The reference command-line implementation distinguishes the three principal outcomes by exit code — `0` for a conforming graph, `1` for a non-conforming graph and `2` for incomplete validation — as a provisional convention pending standardisation.

## 11. Diagnostics

Implementations SHOULD provide actionable diagnostics for invalid dialects, matching errors and conformance violations.

At minimum, a diagnostic SHOULD identify:

- The category of failure.
- The affected element definition, where applicable.
- The affected graph element, where applicable.
- The attribute or schema location, where applicable.
- A human-readable explanation.

For attribute violations, implementations SHOULD include the JSON Schema validation path or equivalent location information. They SHOULD identify the GSL element and attribute so that users can locate the problem in the source graph.

For matching errors, diagnostics SHOULD distinguish unsupported syntax, invalid syntax and evaluation failures where the implementation can determine the difference.

Implementations SHOULD avoid reporting a matching error as an attribute violation, or an invalid dialect as a non-conforming graph.

Stable diagnostic codes, required fields and machine-readable output formats remain open for standardisation.

## 12. Versioning and compatibility

The dialect's `version` field identifies the version of the dialect document.

Implementations MUST NOT assume that two versions of a dialect are compatible merely because their identifiers match.

Changes that alter matching, projection, or the accepted set of graph elements can change conformance results. Changes to JSON Schema constraints can also make a previously conforming graph non-conforming.

The versioning scheme, compatibility rules and treatment of changes to the projection contract remain provisional.

This specification version fixes JSON Schema Draft 2020-12. Dialect documents MUST NOT select a different JSON Schema draft through an additional dialect-level version field.

## 13. Worked example: software architecture

The following example defines a dialect for software architecture. It is illustrative: the query expression and node-reference schema must be checked against the final GQL subset and projection contract.

```yaml
id: software-architecture
version: 0.1.0-draft
title: Software Architecture
description: |
  Conventions for describing software architecture.
guidance: |
  Represent independently identifiable software services
  as nodes. Use directed edges for relationships and sets
  for orthogonal classifications.

elements:
  Service:
    representation: node
    description: |
      An independently identifiable software service.
    guidance: |
      Use a service node for a separately identifiable
      runtime or deployable component.
    match:
      query: 'subgraph node.kind == "service"'
    attributes:
      type: object
      required:
        - kind
        - name
        - owner
        - replicas
      properties:
        kind:
          const: service
        name:
          type: string
          minLength: 1
          maxLength: 120
        owner:
          type: object
          properties:
            "$gslRef":
              type: string
          required:
            - "$gslRef"
          additionalProperties: false
        replicas:
          type: integer
          minimum: 1
          maximum: 1000
      additionalProperties: true
```

The `attributes` schema validates the projected attribute map for each matched service node.

In this example:

- `kind` MUST equal `service`.
- `name` MUST be a non-empty string of at most 120 characters.
- `owner` MUST use the projected node-reference representation.
- `replicas` MUST be an integer between 1 and 1000 inclusive.
- Other attributes are permitted.

The `owner` schema checks the shape of the reference value only. It does not verify that the referenced node exists, is a service, or belongs to a particular dialect definition.

The example deliberately does not impose graph-wide constraints.

## 14. Conformance tests

A conforming implementation MUST be tested against fixtures covering dialect parsing, matching, projection and attribute validation.

The test suite SHOULD include the following categories.

### 14.1 Dialect document tests

- Missing required root fields.
- Invalid root field types.
- Missing required element-definition fields.
- Invalid representation values.
- Invalid `match` structure.
- Invalid JSON Schema.
- Unknown fields at supported levels.
- Empty `elements` mapping.

Unknown fields MUST be ignored. Invalid recognised fields MUST cause dialect rejection.

### 14.2 Matching tests

- Valid queries in the supported subset.
- Queries that return no elements.
- Queries that return multiple elements.
- Queries whose results include incidental element kinds.
- Unsupported query syntax.
- Invalid or unevaluable queries.
- Elements selected by more than one definition.

### 14.3 Projection tests

- String values remain strings.
- Numbers remain numbers without unintended value changes.
- Booleans remain Booleans.
- Empty values use the explicit Empty representation.
- Node references use the defined reference representation.
- Missing attributes are omitted.
- Projection preserves distinctions between strings, Empty values, Booleans and node references.

### 14.4 Attribute validation tests

- Required attributes present and missing.
- Correct and incorrect JSON types.
- Numeric minimum and maximum constraints.
- String length and pattern constraints.
- Enumerated and constant values.
- Explicitly allowed and forbidden additional properties.
- Empty-value schemas.
- Node-reference schemas.
- Multiple definitions applying to the same element.

### 14.5 Outcome tests

- Invalid dialects are rejected before graph validation.
- Attribute violations produce non-conformance.
- Empty match results do not produce violations by themselves.
- Matching errors prevent a conforming result.
- Incomplete validation is never reported as conformance.
- A conforming result does not imply graph completeness.

Tests SHOULD use concrete GSL fixtures with expected match sets, projected JSON instances, validation results and diagnostics. Projection fixtures should be tested independently from JSON Schema validation so that failures can be attributed to the correct stage.

## 15. Implementation sequence

A practical implementation should proceed in the following order:

1. Finalise the dialect document structure and its validation rules.
2. Define the exact supported GQL grammar and evaluation behaviour.
3. Finalise the GSL-to-JSON projection contract.
4. Implement dialect loading and validation.
5. Implement matching and representation filtering.
6. Implement deterministic attribute projection.
7. Integrate a JSON Schema Draft 2020-12 validator.
8. Implement conformance outcomes and diagnostics.
9. Build fixture-based tests covering each stage.
10. Validate the worked example and publish a minimal reference dialect.

Each stage should have independent tests. In particular, the implementation should not depend on the JSON Schema library to interpret GSL values or perform graph matching.

## 16. Decisions and open questions

The following items record the design decisions taken for this draft and the questions that remain open before it should be treated as implementation-ready.

### 16.1 GSL-to-JSON projection

- The node-reference representation is `{"$gslRef": "<node-id>"}` and the explicitly present Empty value is `{"$gslEmpty": true}` (Section 8.3).
- GSL numbers project to JSON numbers; a whole value may satisfy `type: integer`. Full precision and interoperability requirements for differing numeric representations remain open.
- Specify behaviour for any GSL value types added in future versions.

### 16.2 GQL subset

- The supported `subgraph` predicate grammar and semantics are documented in [`SUPPORTED_GQL_SUBSET.md`](SUPPORTED_GQL_SUBSET.md), including the observable distinction between an out-of-subset query and a syntax error.
- Failures to evaluate a match query are reported as a matching error that yields the incomplete outcome (Sections 7.4 and 10.4), distinct from a conformance violation.
- The `subgraph` wrapper and the element namespace are implicit; a `match.query` is a bare attribute predicate expanded with the definition's `representation` as the element prefix (Sections 6.2 and 7.1). Explicit `node.`/`edge.`/`set.` prefixes are rejected.
- Whether later revisions broaden the accepted predicate operators beyond attribute value matching (for example traversal-derived selection or set targeting) remains open.

### 16.3 JSON Schema reference resolution

- `$ref` to targets within the same dialect document is permitted and MUST resolve at compile time; external `http`/`file` `$ref` targets MUST be rejected at dialect load (Section 6.3).
- Whether the permitted internal reference scope extends beyond the single document remains open.

### 16.4 Diagnostics and validation results

- Define stable machine-readable diagnostic codes.
- Define the result representation for invalid dialects, non-conforming graphs, conforming graphs and incomplete validation.
- Define which diagnostic fields are mandatory.

### 16.5 Versioning

- Choose the dialect version syntax.
- Define compatibility expectations.
- Define how future changes to projection semantics affect compatibility.

### 16.6 Dialect document validation

- Publish a JSON Schema for the dialect document itself, or an equivalent normative structural definition.
- Ensure that this validation respects the policy of ignoring unknown fields while rejecting malformed recognised fields.

## 17. Status of this specification

This document consolidates the agreed design and provides a basis for implementation and review.

**Implementation note:** the Go reference implementation expands bare attribute predicates, rejects explicit element prefixes and the explicit `subgraph` wrapper, and validates but does not evaluate a `representation: set` predicate (Section 6.2). The TypeScript implementation is being brought to parity; until it lands, this document describes the target form and the TypeScript package describes the current form.

The following architectural decisions are treated as settled for this draft:

- YAML dialect documents.
- JSON Schema Draft 2020-12 for attribute constraints.
- GQL for matching, with a restricted initial subset (`SUPPORTED_GQL_SUBSET.md`).
- Match expressions written as bare attribute predicates, with the `subgraph` wrapper and element namespace implied by the definition's `representation`.
- Explicit projection of GSL attribute values into JSON.
- The compact `{"$gslRef": ...}` and `{"$gslEmpty": true}` projections for node references and explicit Empty values.
- Internal-document `$ref` permitted; external `$ref` rejected at load.
- Validation of projected attribute maps only.
- No graph-wide constraints or inferred classifications.
- Independent validation of each matched element definition.
- All-or-nothing rejection of malformed dialects.
- Explicit distinction between conformance violations and failures to complete validation.
- Permissive handling of unknown fields.

The diagnostic contract, dialect-document schema and full versioning rules still require final decisions. These should be resolved before declaring the specification implementation-ready.
