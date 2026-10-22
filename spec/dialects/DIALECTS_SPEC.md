# GSL Dialect Specification

**Status:** Design draft
**Version:** `0.4.0-draft`
**Target:** GSL v1.0 draft
**Schema standard:** JSON Schema Draft 2020-12
**Normative companion:** [SUPPORTED_GQL_SUBSET.md](SUPPORTED_GQL_SUBSET.md)
**Reference implementation:** `implementations/go/dialect` (Go); conformance fixtures in `tests/conformance/dialects/`

## 1. Purpose

A GSL dialect defines conventions for representing a particular domain using the Graph Specification Language (GSL).

GSL provides a general-purpose graph representation. A dialect describes how graph elements should be used to represent concepts in a domain, which elements it recognises, and what constraints apply to their attributes and set declarations.

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
- Reuse GQL for selecting nodes and edges rather than introducing a separate matching language.
- Separate normative constraints from descriptive guidance.
- Keep policy configuration separate from definitions and rules.
- Make GSL-specific behaviour explicit and small in scope.
- Report validation failures clearly and distinguish them from failures to execute validation.
- Do not infer missing domain intent from graph structure.
- Do not confuse conformance with completeness, correctness of real-world facts, or overall graph quality.

## 3. Scope

### 3.1 Included

This version defines:

- The structure of a dialect document.
- Ordered matching rules for nodes and edges.
- Validation of attributes on matched nodes and edges.
- Named set definitions and declaration-level set policies.
- Validation of attributes on defined sets.
- Matching nodes and edges using a restricted GQL subset.
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
- Set matching, set-querying operations, or membership constraints beyond emptiness.
- Inheritance or composition between dialects.
- Dialect dependency resolution.
- Arbitrary graph transformations.
- General-purpose extensions to JSON Schema.
- A new query language independent of GQL.

These exclusions may be revisited in later versions where concrete requirements justify them.

## 4. Terminology

**GSL graph:** A graph expressed using GSL.

**Dialect:** A document defining conventions and constraints for a domain.

**Node rule / edge rule:** An entry in the ordered `rules` sequence of the `nodes` or `edges` section. A rule selects graph elements of its section's kind and constrains their attributes.

**Set definition:** A named entry in `sets.definitions` describing a set that the dialect recognises.

**Policy:** Configuration that controls declaration-level behaviour. Policies are separate from rules and definitions.

**Match query:** A GQL expression used to select nodes or edges for validation.

**Projected attribute map:** A JSON object containing the attributes of one matched GSL element after conversion into JSON-compatible values.

**Dialect compiler:** An implementation component that validates and interprets a dialect document and prepares the JSON Schemas used for attribute validation.

**Conformance:** The result of checking the elements and set declarations selected by a dialect against its declared constraints.

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
| `nodes` | No | Node rules |
| `edges` | No | Edge rules |
| `sets` | No | Set policy and named set definitions |

The fields `id` and `version` MUST be present and valid.

The sections `nodes`, `edges` and `sets` are OPTIONAL. An omitted section is equivalent to a present section with no rules, no definitions and the default policy. A dialect with no sections is valid and selects nothing.

Unknown fields MUST be ignored by implementations of this version. Implementations MUST NOT reject an otherwise valid dialect solely because it contains an unknown field.

This permissive behaviour allows dialect documents to carry additional information without immediately breaking older implementations. It also means that misspelled field names may go unnoticed; authors should validate dialect documents and review diagnostics carefully.

### 5.1 Identifier

`id` MUST be a non-empty string that identifies the dialect.

The identifier SHOULD remain stable across compatible revisions of the dialect. It SHOULD change only when the dialect is intentionally treated as a different dialect.

The precise identifier syntax and namespace policy remain to be standardised.

### 5.2 Version

`version` MUST be a non-empty string identifying the version of the dialect definition.

Versioning applies to the dialect, not to GSL or GQL.

A version change does not itself establish compatibility. The meaning of compatibility between dialect versions is specified in Section 13.

The exact version syntax and compatibility policy remain provisional.

### 5.3 Sections

Each of `nodes`, `edges` and `sets` MUST be a mapping if present.

- `nodes` MAY contain `rules`, an ordered sequence of node rules (Section 6).
- `edges` MAY contain `rules`, an ordered sequence of edge rules (Section 6).
- `sets` MAY contain `policy` and `definitions` (Section 7).

An implementation MUST reject a dialect if a recognised section is present but is not a mapping, or if a recognised member has the wrong kind.

## 6. Nodes and edges

The `nodes` and `edges` sections share the same rule structure. They differ only in the kind of graph element their rules select.

```yaml
nodes:
  rules:
    - name: Service
      description: ...
      guidance: ...
      match: 'kind == "service"'
      attributes:
        type: object
        properties: {}

edges:
  rules:
    - name: Dependency
      match: 'protocol exists'
      attributes:
        type: object
        properties: {}
```

A `rules` member MUST be a sequence if present. Each entry MUST be a mapping.

### 6.1 Rule ordering and identity

Rules within a section are ordered. Implementations MUST preserve their document order when processing and reporting them.

Ordering does not create first-match-wins behaviour. Every rule that matches an element applies independently (Section 8.4).

A rule MAY contain `name`, a non-empty string used as a dialect-local label for readability and diagnostics. It does not create a new GSL element type. A rule name need not be unique. When a rule has no name, implementations SHOULD identify it by its section and zero-based position, for example `nodes.rules[0]`.

### 6.2 Match

`match` MUST be a non-empty string containing a supported match expression.

A match expression is written as an **attribute predicate without an element prefix**. The `subgraph` wrapper and the element namespace are both implicit and are supplied by the enclosing section: `nodes` supplies `node.` and `edges` supplies `edge.`. An implementation MUST expand the expression by prefixing every attribute reference with that element prefix before evaluation.

For a node rule, the two queries below are equivalent, and the second is the canonical expansion of the first.

```yaml
nodes:
  rules:
    - match: 'kind == "service"'
      attributes:
        type: object
```

```yaml
nodes:
  rules:
    - match: 'subgraph node.kind == "service"'
      attributes:
        type: object
```

The second form is shown only to describe the expansion. The accepted input form is the bare predicate; an explicit `subgraph` wrapper or element prefix MUST be rejected (Section 8.1).

The expansion is exactly the `subgraph` form of the full GQL grammar with the section's element prefix. Because the expansion is defined this way, the match expression is a strict subset of GQL and introduces no new syntax.

An implementation MUST reject, as an invalid dialect, a `match` that:

- is empty or not a string;
- is not a single attribute predicate (for example it contains a pipeline `|`, traversal, graph algebra, a graph transformation, or more than one top-level predicate);
- contains an explicit element prefix (`node.`, `edge.` or `set.`) rather than a bare attribute name;
- contains an explicit `subgraph` wrapper; or
- otherwise does not conform to the supported subset in [`SUPPORTED_GQL_SUBSET.md`](SUPPORTED_GQL_SUBSET.md).

Matching determines which graph elements are subject to the rule's constraints. It does not establish that those elements are valid.

A matched element MUST remain selected for validation even if its attributes violate the associated JSON Schema. Implementations MUST NOT silently exclude malformed elements from validation.

The supported query subset is defined in Section 8.

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

The following `$ref` policy applies. A schema MAY use `$ref` to targets defined within the same `attributes` schema (for example `#/$defs/...`); such internal references MUST resolve during compilation. The schema is a JSON Schema document embedded in the dialect; this version does not define dialect-level shared `$defs` across rules. A schema MUST NOT use `$ref` to external `http` or `file` targets; an implementation MUST reject such a dialect at load time and MUST NOT silently resolve references using arbitrary network or filesystem access. The `$ref` policy remains provisional and may be extended.

### 6.4 Description and guidance

`description` is a human-readable explanation of the rule.

`guidance` provides advice for constructing or interpreting graph elements selected by the rule.

These fields MUST NOT alter matching, projection, or validation behaviour. Implementations MUST NOT treat prose in either field as an executable constraint.

Guidance may explain, for example, when a concept should be represented as a node rather than an edge, or how a language model should choose meaningful attribute names.

## 7. Sets

The `sets` section defines declaration-level policy and named set definitions. It does not select sets with a query.

```yaml
sets:
  policy:
    allow_undefined: true
    require_all: false
    allow_empty: true

  definitions:
    critical:
      description: Sets whose failure is operationally critical.

    deprecated:
      description: Sets scheduled for removal.
```

`sets` MUST be a mapping if present. It MAY contain `policy` and `definitions`. Unknown fields MUST be ignored.

### 7.1 Set policy

`policy` MUST be a mapping if present. The following keys are defined:

| Key | Type | Default | Meaning |
|---|---|---|---|
| `allow_undefined` | Boolean | `true` | Whether sets present in the graph may have no dialect definition |
| `require_all` | Boolean | `false` | Whether every dialect-defined set must be present in the graph |
| `allow_empty` | Boolean | `true` | Whether a set present in the graph may have no members |

Each recognised key MUST be a Boolean if present. Unknown keys MUST be ignored.

These policies are independent:

1. **Recognition:** `allow_undefined` controls whether a set present in the graph must be recognised by a definition.
2. **Requirement:** `require_all` controls whether every definition must be present in the graph.
3. **Membership:** `allow_empty` controls whether a set present in the graph must have at least one member.

A set is **present in the graph** when it appears in the graph's set collection, whether through an explicit `set` declaration or through implicit creation by `@` membership. For dialect policy purposes, both forms count as a declaration; otherwise strictness could be evaded by using membership without an explicit declaration.

A set has a **member** when a node or edge declares membership in that set. Membership is the GSL `@setName` relation; the dialect specification does not introduce a separate membership concept.

Policy failures are graph conformance violations, not invalid-dialect errors (Section 11). For example:

- With `allow_undefined: false`, every set present in the graph MUST have a definition.
- With `require_all: true`, every definition MUST correspond to a set present in the graph.
- With `allow_empty: false`, every set present in the graph MUST have at least one member.

The defaults are deliberately permissive so that adding a `sets` section does not reject graphs that the dialect does not constrain.

### 7.2 Set definitions

`definitions` MUST be a mapping if present. Each entry maps a set name to its definition.

The map key MUST be a non-empty GSL set identifier matching the grammar `[A-Za-z_][A-Za-z0-9_]*`. It is the name of the GSL set the definition describes; there is no separate `name` field.

Each definition MUST be a mapping if present. It MAY contain:

- `description` — a human-readable explanation of the set.
- `guidance` — advice for using the set.
- `attributes` — a valid JSON Schema Draft 2020-12 schema for the set's projected attributes.

A definition with no `attributes` still participates in declaration policy (`allow_undefined`, `require_all`); it simply imposes no attribute constraint.

If a definition supplies `attributes`, an implementation MUST compile it under the same rules as node and edge attributes (Section 6.3) and validate the projected attributes of the corresponding set when that set is present in the graph.

Set definitions are named GSL set definitions, not categories determined by a match predicate. This version defines no set matchers and no constraint on set membership other than `allow_empty`.

### 7.3 Description and guidance

Set definition `description` and `guidance` are prose only. They MUST NOT alter set policy, matching, projection, or validation behaviour.

## 8. Matching

### 8.1 Query language

GQL is the query language for selecting nodes and edges. This specification does not define a separate dialect matching language: a rule's `match` is expanded into a full GQL `subgraph` expression (Section 6.2) and evaluated with GQL semantics.

A match expression is a single **attribute predicate** whose element prefix is supplied by the enclosing section. This is the initial supported subset.

The initial subset MUST NOT include:

- Graph traversal.
- Query pipelines.
- Graph algebra.
- Graph transformations.
- An explicit element prefix in the match expression.
- An explicit `subgraph` wrapper.
- Other unsupported query forms.

The exact grammar and expression semantics of the supported subset are documented normatively in the companion profile [`SUPPORTED_GQL_SUBSET.md`](SUPPORTED_GQL_SUBSET.md), which defines the accepted predicate forms, the expansion rule, and the observable distinction between an out-of-subset query and a syntax error.

Implementations MUST reject unsupported query forms. They MUST NOT approximate an unsupported query with different semantics.

### 8.2 Matching process

For each node or edge rule, an implementation MUST:

1. Expand its match expression into a `subgraph` predicate using the enclosing section's element prefix (Section 6.2).
2. Evaluate the expanded query against the input graph.
3. Identify the elements returned by the query.
4. Select only elements whose kind matches the enclosing section: nodes for `nodes.rules`, edges for `edges.rules`.
5. Validate the attributes of each selected element against the rule's schema.

A match query may return a graph containing several kinds of elements. Elements of a kind other than the rule's section are incidental and MUST be ignored by that rule. They MUST NOT become subject to validation merely because they appear in the result.

### 8.3 Empty match results

A valid query that selects no eligible elements is not an error.

The implementation MUST continue validation of other rules and set policy, and MUST NOT report a conformance violation solely because a query has no matches.

Consequently, a graph may conform to a dialect even when none of its nodes or edges match any rule. Conformance is defined over selected elements and declared sets, not over an expectation that the graph contains particular domain concepts.

### 8.4 Overlapping matches

An element may be selected by more than one rule. There is no one-to-one relationship between graph elements and rules.

If the same element is selected by multiple rules, each applicable rule MUST be validated independently. Implementations MUST NOT stop after the first match, merge the applicable schemas, or treat rule order as a precedence order.

Rule order affects only deterministic processing and reporting order.

### 8.5 Matching errors

An invalid, unsupported or unevaluable query MUST produce a matching error.

A matching error is distinct from an attribute constraint violation or set policy violation. If a matching error prevents the validator from completing the required checks, the graph MUST NOT be reported as conforming.

## 9. GSL-to-JSON projection

### 9.1 Purpose

JSON Schema validates JSON instances, not GSL syntax or graph structures directly.

Implementations MUST therefore project the attributes of each selected element into a JSON object before validating them.

Projection MUST be deterministic and MUST preserve the semantic distinctions defined by GSL. It MUST NOT silently coerce one GSL value type into another.

The projection is semantic, not a reconstruction of the original GSL source text.

### 9.2 Attribute map

The projected instance MUST be a JSON object whose properties correspond to the attributes of the selected element.

An attribute that is absent in GSL MUST be omitted from the JSON object.

An explicitly present GSL `Empty` value MUST NOT be treated as an absent attribute, JSON `null`, or Boolean `true`.

### 9.3 Value mappings

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

GSL node references are available only where permitted by the GSL specification. This dialect specification does not introduce edge or set references. Set attributes cannot contain node references (GSL specification Section 9), so a set definition schema should not expect `$gslRef`.

### 9.4 Example

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

### 9.5 Schema constraints on projected values

Dialect authors MUST write their attribute schemas against the projected JSON representation, not against GSL source syntax.

For example, a schema that accepts only strings will reject a projected node reference because a node reference is represented by an object. A schema that intends to accept references MUST explicitly describe the reference representation.

Implementations SHOULD provide documentation and examples for common GSL value types so that dialect authors do not have to rediscover these mappings.

## 10. JSON Schema compilation and validation

### 10.1 Architecture

Implementations MUST use JSON Schema Draft 2020-12 for attribute validation.

The implementation SHOULD use an established JSON Schema validator rather than implement JSON Schema validation independently.

The implementation is responsible for GSL-specific behaviour, including:

- Parsing and validating the dialect document.
- Interpreting node and edge rules and set definitions.
- Evaluating supported match queries.
- Selecting eligible graph elements and sets.
- Projecting GSL attribute values into JSON.
- Preparing the applicable JSON Schemas.
- Applying set declaration policy.
- Mapping validation failures into dialect conformance diagnostics.

The JSON Schema library is responsible for evaluating the compiled schemas against projected JSON instances.

### 10.2 Compilation

For each node or edge rule, and for each set definition that supplies `attributes`, the implementation MUST prepare a schema that validates the projected attribute map.

Because `attributes` is already JSON Schema, compilation SHOULD preserve standard JSON Schema semantics. It need not translate each ordinary constraint into a new representation.

The implementation MAY assemble, wrap or compile the schema as required by its chosen validator, but it MUST NOT change the meaning of valid JSON Schema constraints.

The implementation MUST NOT invent custom JSON Schema keywords for constraints that standard JSON Schema can express.

### 10.3 Validation isolation

Each selected element MUST be validated independently against the applicable rule or set definition.

The validator MUST NOT implicitly combine the schemas of different rules or assume that an element can match only one rule.

The schema instance MUST contain only the selected element's projected attributes. Graph-wide information MUST NOT be added to the instance unless a future version explicitly defines such behaviour.

### 10.4 Additional attributes

The behaviour for attributes not explicitly listed in a schema is controlled by JSON Schema.

For example:

- `additionalProperties: false` rejects undeclared properties.
- `additionalProperties: true` permits them.
- `patternProperties` can constrain properties matching a pattern.

The dialect specification MUST NOT impose a separate default policy for additional attributes.

### 10.5 Standard JSON Schema semantics

Implementations MUST follow the semantics of JSON Schema Draft 2020-12 for supported keywords.

A dialect is invalid if its schema is invalid under that specification or if the implementation cannot support a required feature within its declared conformance profile.

Implementations MUST report such a failure rather than silently ignore a recognised constraint.

## 11. Conformance outcomes

Validation MUST distinguish the principal outcomes below.

### 11.1 Invalid dialect

The dialect is invalid if its structure, required fields, recognised field values, match queries or attribute schemas are invalid.

An invalid dialect MUST be rejected before graph validation begins.

This version uses all-or-nothing dialect loading. Implementations MUST NOT partially apply a malformed dialect and report the resulting graph as conforming.

### 11.2 Non-conforming graph

A graph is non-conforming if the dialect is valid, the required validation completes, and one or more applicable constraints are violated.

Examples include:

- A required node or edge attribute is missing.
- An attribute has the wrong JSON type after projection.
- A numeric value falls outside a permitted range.
- A property is rejected by `additionalProperties: false`.
- A set present in the graph has no definition while `allow_undefined` is `false`.
- A dialect-defined set is absent from the graph while `require_all` is `true`.
- A set present in the graph has no members while `allow_empty` is `false`.
- A defined set's projected attributes violate its schema.

A non-conforming graph may still be a valid GSL graph. Dialect conformance is separate from GSL language validity.

### 11.3 Conforming graph

A graph is conforming if the dialect is valid, all required matching, set policy checks and validation complete, and no applicable constraint violations are found.

Conformance means only that the selected elements and set declarations satisfy the constraints defined by the dialect.

It does not establish that:

- The graph is complete.
- Every graph element belongs to the dialect.
- Every domain concept that should be represented is present.
- The graph accurately reflects the real world.
- The graph is free from contradictions not expressed by the dialect.

### 11.4 Incomplete validation

If matching or validation fails in a way that prevents required checks from completing, the implementation MUST NOT report the graph as conforming.

Implementations SHOULD distinguish incomplete validation from a completed validation that found violations.

The precise external result model for incomplete validation remains to be finalised. The reference command-line implementation distinguishes the three principal outcomes by exit code — `0` for a conforming graph, `1` for a non-conforming graph and `2` for incomplete validation — as a provisional convention pending standardisation.

## 12. Diagnostics

Implementations SHOULD provide actionable diagnostics for invalid dialects, matching errors, set policy violations and conformance violations.

At minimum, a diagnostic SHOULD identify:

- The category of failure.
- The affected rule or set definition, where applicable.
- The affected graph element or set, where applicable.
- The attribute or schema location, where applicable.
- A human-readable explanation.

For attribute violations, implementations SHOULD include the JSON Schema validation path or equivalent location information. They SHOULD identify the GSL element and attribute so that users can locate the problem in the source graph.

For set policy violations, implementations SHOULD identify the set name and the policy that failed.

For matching errors, diagnostics SHOULD distinguish unsupported syntax, invalid syntax and evaluation failures where the implementation can determine the difference.

Implementations SHOULD avoid reporting a matching error as an attribute violation, or an invalid dialect as a non-conforming graph.

Stable diagnostic codes, required fields and machine-readable output formats remain open for standardisation.

## 13. Versioning and compatibility

The dialect's `version` field identifies the version of the dialect document.

Implementations MUST NOT assume that two versions of a dialect are compatible merely because their identifiers match.

Changes that alter matching, projection, set policy, or the accepted set of graph elements can change conformance results. Changes to JSON Schema constraints can also make a previously conforming graph non-conforming.

The versioning scheme, compatibility rules and treatment of changes to the projection contract remain provisional.

This specification version fixes JSON Schema Draft 2020-12. Dialect documents MUST NOT select a different JSON Schema draft through an additional dialect-level version field.

## 14. Worked example: software architecture

The following example defines a dialect for software architecture. It is illustrative: the query expressions and node-reference schema must be checked against the final GQL subset and projection contract.

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

nodes:
  rules:
    - name: Service
      description: |
        An independently identifiable software service.
      guidance: |
        Use a service node for a separately identifiable
        runtime or deployable component.
      match: 'kind == "service"'
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

edges:
  rules:
    - name: Dependency
      match: 'protocol exists'
      attributes:
        type: object
        required:
          - protocol
        properties:
          protocol:
            enum:
              - http
              - grpc
              - tcp

sets:
  policy:
    allow_undefined: true
    require_all: false
    allow_empty: true
  definitions:
    critical:
      description: |
        Sets whose failure is operationally critical.
    deprecated:
      description: |
        Sets scheduled for removal.
```

The `attributes` schema validates the projected attribute map for each matched service node. In this example:

- `kind` MUST equal `service`.
- `name` MUST be a non-empty string of at most 120 characters.
- `owner` MUST use the projected node-reference representation.
- `replicas` MUST be an integer between 1 and 1000 inclusive.
- Other attributes are permitted.

The `owner` schema checks the shape of the reference value only. It does not verify that the referenced node exists, is a service, or belongs to a particular dialect rule.

The example deliberately does not impose graph-wide constraints. It recognises the `critical` and `deprecated` sets but does not require them to exist or reject other sets.

## 15. Conformance tests

A conforming implementation MUST be tested against fixtures covering dialect parsing, matching, set policy, projection and attribute validation.

The test suite SHOULD include the following categories.

### 15.1 Dialect document tests

- Missing required root fields.
- Invalid root field types.
- Invalid section types.
- Invalid rule structure.
- Invalid `match` structure.
- Invalid policy values.
- Invalid set-definition keys or structures.
- Invalid JSON Schema.
- Unknown fields at supported levels.
- Empty sections and an empty dialect.

Unknown fields MUST be ignored. Invalid recognised fields MUST cause dialect rejection.

### 15.2 Matching tests

- Valid queries in the supported subset.
- Queries that return no elements.
- Queries that return multiple elements.
- Queries whose results include incidental element kinds.
- Unsupported query syntax.
- Invalid or unevaluable queries.
- Elements selected by more than one rule.
- Rule order does not create first-match-wins behaviour.

### 15.3 Set policy tests

- Default policy permits undefined, missing and empty sets.
- `allow_undefined: false` rejects sets with no definition.
- `require_all: true` rejects definitions absent from the graph.
- `allow_empty: false` rejects sets with no members.
- Membership on nodes and edges counts when checking emptiness.
- Implicitly created sets participate in policy.
- A defined set's attributes are validated when it is present.
- A definition without `attributes` imposes no attribute constraint.

### 15.4 Projection tests

- String values remain strings.
- Numbers remain numbers without unintended value changes.
- Booleans remain Booleans.
- Empty values use the explicit Empty representation.
- Node references use the defined reference representation.
- Missing attributes are omitted.
- Projection preserves distinctions between strings, Empty values, Booleans and node references.

### 15.5 Attribute validation tests

- Required attributes present and missing.
- Correct and incorrect JSON types.
- Numeric minimum and maximum constraints.
- String length and pattern constraints.
- Enumerated and constant values.
- Explicitly allowed and forbidden additional properties.
- Empty-value schemas.
- Node-reference schemas.
- Multiple rules applying to the same element.
- Attributes on defined sets.

### 15.6 Outcome tests

- Invalid dialects are rejected before graph validation.
- Attribute violations produce non-conformance.
- Set policy violations produce non-conformance.
- Empty match results do not produce violations by themselves.
- Matching errors prevent a conforming result.
- Incomplete validation is never reported as conformance.
- A conforming result does not imply graph completeness.

Tests SHOULD use concrete GSL fixtures with expected match sets, projected JSON instances, validation results and diagnostics. Projection fixtures should be tested independently from JSON Schema validation so that failures can be attributed to the correct stage.

## 16. Implementation sequence

A practical implementation should proceed in the following order:

1. Finalise the dialect document structure and its validation rules.
2. Define the exact supported GQL grammar and evaluation behaviour.
3. Finalise the GSL-to-JSON projection contract.
4. Implement dialect loading and validation.
5. Implement node and edge rule matching.
6. Implement set declaration policy.
7. Implement deterministic attribute projection.
8. Integrate a JSON Schema Draft 2020-12 validator.
9. Implement conformance outcomes and diagnostics.
10. Build fixture-based tests covering each stage.
11. Validate the worked example and publish a minimal reference dialect.

Each stage should have independent tests. In particular, the implementation should not depend on the JSON Schema library to interpret GSL values or perform graph matching.

## 17. Decisions and open questions

The following items record the design decisions taken for this draft and the questions that remain open before it should be treated as implementation-ready.

### 17.1 GSL-to-JSON projection

- The node-reference representation is `{"$gslRef": "<node-id>"}` and the explicitly present Empty value is `{"$gslEmpty": true}` (Section 9.3).
- GSL numbers project to JSON numbers; a whole value may satisfy `type: integer`. Full precision and interoperability requirements for differing numeric representations remain open.
- Specify behaviour for any GSL value types added in future versions.

### 17.2 GQL subset

- The supported `subgraph` predicate grammar and semantics are documented in [`SUPPORTED_GQL_SUBSET.md`](SUPPORTED_GQL_SUBSET.md), including the observable distinction between an out-of-subset query and a syntax error.
- Failures to evaluate a match query are reported as a matching error that yields the incomplete outcome (Sections 8.5 and 11.4), distinct from a conformance violation.
- The `subgraph` wrapper and the element namespace are implicit; a rule's `match` is a bare attribute predicate expanded with the enclosing section's element prefix (Sections 6.2 and 8.1). Explicit `node.`/`edge.`/`set.` prefixes and an explicit `subgraph` wrapper are rejected.
- Whether later revisions broaden the accepted predicate operators beyond attribute value matching (for example traversal-derived selection) remains open.

### 17.3 Set semantics

- Set definitions are keyed by GSL set name and do not use match queries (Section 7.2).
- The set policies and their defaults are defined in Section 7.1: `allow_undefined: true`, `require_all: false`, `allow_empty: true`.
- Implicitly created sets are treated as set declarations for policy purposes (Section 7.1).
- Whether later revisions add set-targeting GQL predicates, membership constraints beyond emptiness, or category-style set matching remains open; none is required by the current contract.

### 17.4 Rule identity

- Node and edge rules may carry an optional `name` used for readability and diagnostics (Section 6.1).
- Unnamed rules are identified by section and position, for example `nodes.rules[0]`.
- Stable machine-readable rule identifiers remain open.

### 17.5 JSON Schema reference resolution

- `$ref` to targets within the same `attributes` schema is permitted and MUST resolve at compile time; external `http`/`file` `$ref` targets MUST be rejected at dialect load (Section 6.3).
- Whether the permitted internal reference scope extends beyond a single `attributes` schema remains open.

### 17.6 Diagnostics and validation results

- Define stable machine-readable diagnostic codes.
- Define the result representation for invalid dialects, non-conforming graphs, conforming graphs and incomplete validation.
- Define which diagnostic fields are mandatory.

### 17.7 Versioning

- Choose the dialect version syntax.
- Define compatibility expectations.
- Define how future changes to projection semantics affect compatibility.

### 17.8 Dialect document validation

- Publish a JSON Schema for the dialect document itself, or an equivalent normative structural definition.
- Ensure that this validation respects the policy of ignoring unknown fields while rejecting malformed recognised fields.

## 18. Status of this specification

This document consolidates the agreed design and provides a basis for implementation and review.

**Implementation note:** the Go reference implementation and the TypeScript implementation both parse the `nodes`, `edges` and `sets` sections, preserve node and edge rule order, and apply set declaration policy. Section optionality and permissive set-policy defaults are consistent across both implementations.

The following architectural decisions are treated as settled for this draft:

- YAML dialect documents.
- Explicit `nodes`, `edges` and `sets` sections instead of a generic `elements` map and a `representation` selector.
- Ordered node and edge rule sequences; all matching rules apply independently.
- Optional rule names for readability and diagnostics.
- JSON Schema Draft 2020-12 for attribute constraints.
- GQL for matching, with a restricted initial subset (`SUPPORTED_GQL_SUBSET.md`).
- Match expressions written as bare attribute predicates, with the `subgraph` wrapper and element namespace implied by the enclosing section.
- Explicit projection of GSL attribute values into JSON.
- The compact `{"$gslRef": ...}` and `{"$gslEmpty": true}` projections for node references and explicit Empty values.
- Internal-document `$ref` permitted within an `attributes` schema; external `$ref` rejected at load.
- Validation of projected attribute maps only.
- No graph-wide constraints or inferred classifications.
- Independent validation of every matching rule.
- Set definitions keyed by set name, with separate declaration policy and no set matching.
- All-or-nothing rejection of malformed dialects.
- Explicit distinction between conformance violations and failures to complete validation.
- Permissive handling of unknown fields.

The diagnostic contract, dialect-document schema and full versioning rules still require final decisions. These should be resolved before declaring the specification implementation-ready.
