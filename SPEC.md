# GSL-Lang Specification

Version 1.0.0 (Draft)

## Status

This document defines the normative specification for GSL-Lang v1.0.

The key words **MUST**, **MUST NOT**, **REQUIRED**, **SHALL**, **SHALL NOT**, **SHOULD**, **SHOULD NOT**, **RECOMMENDED**, **MAY**, and **OPTIONAL** in this document are to be interpreted as described in RFC 2119 and RFC 8174.

If any conflict exists between this document and other documentation (including README), this document is authoritative.

---

## Table of Contents

1. Scope
2. Lexical Structure
3. Core Model
4. Core Invariants
5. Nodes
6. Edges
7. Edge Scoping
8. Sets
9. Attribute Semantics
10. Multiple Declarations
11. Canonicalisation
12. Errors
13. Linter Warnings (Non-Normative)
14. Conformance

---

# 1. Scope

GSL defines a textual representation of directed graphs composed of:

* Nodes
* Directed edges (multiset)
* Named sets
* Attribute maps attached to nodes, edges, and sets

This specification defines:

* Lexical grammar
* Syntactic grammar
* Semantic constraints
* Canonicalisation requirements
* Error conditions

This specification does NOT define:

* Graph layout
* Schema validation
* Graph-theoretic correctness (e.g. acyclicity)

---

# 2. Lexical Structure

## 2.1 Identifiers

An identifier:

```
[A-Za-z_][A-Za-z0-9_]*
```

* MUST be case-sensitive
* MUST NOT match a reserved keyword

## 2.2 Reserved Keywords

The following identifiers are reserved and MUST NOT be used:

* `node`
* `set`
* `true`
* `false`

---

## 2.3 Literals

### 2.3.1 Strings

A string:

* MUST be enclosed in double quotes (`"`).
* MAY contain escape sequences.
* MUST support round-trip serialisation.

### 2.3.2 Numbers

A number:

* MUST be a decimal integer or decimal floating-point value.
* MUST NOT use exponent notation.
* MUST NOT include a sign (`+` or `-`).

### 2.3.3 Booleans

Boolean literals:

* MUST be either `true` or `false`.

---

## 2.4 Comments

* A comment MUST begin with `#`.
* A comment MUST continue to end-of-line.
* A parser MUST ignore comments entirely.

---

## 2.5 Whitespace

Whitespace:

* MUST be treated as insignificant except as token separator.
* MAY appear between any two tokens.

---

# 3. Core Model

A conforming implementation MUST construct a graph model consisting of:

* A set of nodes (unique by identifier)
* A multiset of edges
* A set of sets (unique by identifier)

Node identifiers and set identifiers:

* MUST exist in separate namespaces.
* MAY share the same name without error.
* SHOULD generate a linter warning if identical names are used in both namespaces.

---

# 4. Core Invariants

### 4.1 Edge Instance Independence

Edges are distinct instances.

Multiple edges may exist with identical:

* source
* target
* attributes

They MUST NOT be merged or deduplicated.

---

### 4.2 Identity Opacity

Edge identity is not observable.

* Implicit identities MUST NOT be referenced
* Programs MUST NOT depend on identity
* Implementations MAY assign identities arbitrarily

---

### 4.3 Lexical Dependency Only

Dependencies between edges are introduced only via:

* explicit labels
* lexical scoping

No other mechanism is allowed.

---

### 4.4 Parent Existence Guarantee

If an edge depends on another edge:

* The parent edge MUST exist

This includes edges introduced via scoped blocks.

---

### 4.5 Non-Observability of Unlabeled Dependencies

Dependencies targeting unlabeled edges:

* Are not directly observable
* Cannot be referenced

---

# 5. Nodes

## 5.1 Declaration

A node MUST be declared using:

```
node <IDENT>
```

If a node is declared multiple times:

* The declarations MUST be merged.
* Attribute conflicts MUST be resolved using last-write-wins semantics.
* Set membership MUST accumulate.

---

## 5.2 Attributes

Node attributes:

* MUST be declared inside `[ ... ]`.
* MUST consist of key-value pairs or keys without values.
* MUST NOT contain duplicate keys within a single declaration (this is a syntax error).
* MAY contain values of type:

    * String
    * Number
    * Boolean
    * NodeRef (identifier)
    * Empty

Empty attribute:

```
flag
```

* MUST be interpreted as an attribute with an explicit Empty value.
* MUST NOT be interpreted as `true`.

---

## 5.3 Text Shorthand

```
node A: "text"
```

* MUST be semantically equivalent to:

  ```
  node A [text="text"]
  ```

If both shorthand and explicit `text` attribute are provided in the same declaration, this MUST be treated as a duplicate attribute error.

Text shorthand and an attribute list MUST NOT appear in the same declaration. A node requiring both a `text` attribute and other attributes MUST use the attribute list form:

```
node A [text="Hello", flag]
```

---

## 5.4 Parent Attribute

`parent`:

* MUST be treated as a normal attribute.
* MUST accept only NodeRef values.
* MAY reference a forward-declared node.

---

## 5.5 Block Syntax

```
node C {
    node D
}
```

Block syntax:

* MUST be treated as syntactic sugar.
* Each node declared inside the block MUST implicitly receive `parent=<enclosing node>`.
* Block structure MUST NOT be preserved in canonical form.

If a node inside a block explicitly declares a conflicting parent:

* The explicit parent MUST override the implicit parent.
* An implementation SHOULD emit a linter warning.

Blocks MAY be nested.

---

# 6. Edges

## 6.1 Declaration

```
<node_list> -> <node_list>
```

Nodes MAY be forward-declared.

---

## 6.2 Grouped Edges

Implementations MUST:

* Expand grouped edges into individual edges.

The following form is invalid:

```
A,B -> C,D
```

If both left and right node lists contain more than one identifier:

* This MUST be treated as a syntax error.

---

## 6.3 Attributes

* Same syntax as nodes
* MUST NOT contain NodeRef
* MUST NOT duplicate keys

---

## 6.4 Text Shorthand

```
A->B: "Next"
```

* MUST be equivalent to:

  ```
  A->B [text="Next"]
  ```

Text shorthand and an attribute list MUST NOT appear in the same edge declaration. An edge requiring both MUST use the attribute list form:

```
A->B [text="Next", weight=1.2]
```

---

## 6.5 Duplicate Edges

* MUST be preserved
* MUST allow duplicates
* MUST be treated as a multiset.

---

## 6.6 Edge Membership

Edges MAY declare set membership via:

```
@setName
```

Membership:

* MUST accumulate.
* MAY cause implicit set creation.

---

## 6.7 Labels

Edges MAY be labeled:

```
E1: A -> B
```

Labels:

* MUST uniquely identify an edge within scope
* MAY be used as dependency targets

---

# 7. Edge Scoping

## 7.1 Overview

```
A -> B {
  B -> C
}
```

or:

```
E1: A -> B {
  B -> C
}
```

---

## 7.2 Semantics

1. Parent edge is constructed
2. Child edges are constructed
3. Each child depends on parent

Equivalent to:

```
E1: A -> B
B -> C [depends_on = E1]
```

---

## 7.3 Dependency Scope Rule

Each edge depends on exactly one parent:

* nearest enclosing edge

---

## 7.4 Nested Scopes

```
A: a -> b {
  B: b -> c {
    c -> d
  }
}
```

* B depends on A
* c->d depends on B

---

## 7.5 Scope and Labels

Labels become dependency targets.

Unlabeled edges:

* still act as parents
* cannot be referenced

---

## 7.6 Constraints

Invalid:

```
e = A -> B
e { ... }
```

Invalid:

```
A -> B {
  C -> D [depends_on = X]
}
```

Rules:

* Scoped edges MUST depend on exactly one parent
* Dependencies MUST NOT be transitive

---

## 7.7 Observability

* Unlabeled edges cannot be referenced
* Dependencies on them are not inspectable

---

## 7.8 Non-Expression Semantics

Scoped edges:

* Are not values
* Cannot be assigned or reused

---

# 8. Sets

## 8.1 Declaration

A set MUST be declared using:

```
set <IDENT>
```

Multiple declarations:

* MUST merge.
* MUST use last-write-wins for attribute conflicts.

---

## 8.2 Attributes

Set attributes:

* MUST follow attribute syntax.
* MUST NOT contain NodeRef values.
* MUST NOT contain duplicate keys within a single declaration.

---

## 8.3 Membership

Sets:

* MUST NOT declare membership in other sets.
* MUST NOT contain membership expressions.

---

## 8.4 Implicit Sets

If a set is referenced via `@` but not declared:

* The set MUST be implicitly created.
* The set MUST initially have no attributes.
* Implementations SHOULD emit a linter warning.

---

# 9. Attribute Semantics

Attribute values are dynamically typed.

Contextual restrictions:

| Context | NodeRef Allowed |
| ------- | --------------- |
| Node    | YES             |
| Edge    | NO              |
| Set     | NO              |

If a NodeRef appears in an edge or set attribute:

* The parser MUST reject the input.

---

## 9.1 Reserved Attributes

Reserved:

* `depends_on`

---

## 9.2 depends_on

Defines edge dependency.

* MUST resolve to exactly one edge

* MUST be:

  * labeled edge OR
  * lexical parent

* Multiple dependencies NOT supported

---

# 10. Multiple Declarations

Nodes and sets:

* MUST support multiple declarations.
* MUST merge attributes.
* MUST use last-write-wins for conflicts.
* MUST accumulate membership.

---

# 11. Canonicalisation

A conforming implementation MUST ensure:

```
parse(serialize(parse(input))) == parse(input)
```

Canonicalisation MUST:

1. Expand grouped edges.
2. Convert block syntax to explicit `parent` attributes.
3. Preserve duplicate edges.
4. Preserve empty attributes.
5. Materialise implicitly created sets.
6. Merge multiple declarations.

Ordering:

* Ordering of nodes and sets in serialisation MAY be implementation-defined.
* Canonical form MUST parse into an identical internal representation.

---

## 11.1 Scoped Evaluation

When evaluating scoped edges:

1. Construct parent
2. Evaluate body with parent context
3. Attach dependency to each child

Order-independent except for parent-child relation.

---

# 12. Errors

The following conditions MUST produce a syntax error:

* Duplicate attribute keys within a single declaration.
* Grouped edges on both sides of `->`.
* NodeRef values in edge attributes.
* NodeRef values in set attributes.
* Use of reserved keywords as identifiers.
* Explicit `depends_on` inside scoped edges.

---

## 12.1 Examples

```invalid-gsl
node A [color="red", color="blue"]
```

Duplicate attribute key `color`.

```invalid-gsl
A,B -> C,D
```

Both sides of edge declaration are grouped (invalid).

```invalid-gsl
A->B [target=C]
```

NodeRef value (`C`) in edge attribute (invalid).

```invalid-gsl
set colors [primary=Red]
```

NodeRef value (`Red`) in set attribute (invalid).

```invalid-gsl
node node
```

Reserved keyword `node` used as identifier (invalid).

---

# 13. Linter Warnings (Non-Normative)

Implementations SHOULD provide warnings for:

* Implicit set creation.
* Parent override inside a block.
* NodeRef type override across declarations.
* Node and set name collision.

Warnings MUST NOT prevent parsing.

---

# 14. Conformance

An implementation conforms if it:

* Accepts all valid input
* Rejects all invalid input
* Produces canonical output
