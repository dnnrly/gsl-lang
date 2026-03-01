# GSL-Lang Specification

Version 1.0.0 (Draft)

## Status

This document defines the normative specification for GSL-Lang v1.0.

The key words **MUST**, **MUST NOT**, **REQUIRED**, **SHALL**, **SHALL NOT**, **SHOULD**, **SHOULD NOT**, **RECOMMENDED**, **MAY**, and **OPTIONAL** in this document are to be interpreted as described in RFC 2119 and RFC 8174.

If any conflict exists between this document and other documentation (including README), this document is authoritative.

---

## Table of Contents

1. [Scope](#1-scope)
2. [Lexical Structure](#2-lexical-structure)
3. [Core Model](#3-core-model)
4. [Nodes](#4-nodes)
5. [Edges](#5-edges)
6. [Sets](#6-sets)
7. [Attribute Semantics](#7-attribute-semantics)
8. [Multiple Declarations](#8-multiple-declarations)
9. [Canonicalisation](#9-canonicalisation)
10. [Errors](#10-errors)
11. [Linter Warnings](#11-linter-warnings-non-normative)
12. [Conformance](#12-conformance)

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

* MUST match the pattern:

```
[A-Za-z_][A-Za-z0-9_]*
```

* MUST be case-sensitive.
* MUST NOT match a reserved keyword.

## 2.2 Reserved Keywords

The following identifiers are reserved and MUST NOT be used as identifiers:

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

# 4. Nodes

## 4.1 Declaration

A node MUST be declared using:

```
node <IDENT>
```

If a node is declared multiple times:

* The declarations MUST be merged.
* Attribute conflicts MUST be resolved using last-write-wins semantics.
* Set membership MUST accumulate.

---

## 4.2 Attributes

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

An attribute without a value:

```
flag
```

* MUST be interpreted as an attribute with an explicit Empty value.
* MUST NOT be interpreted as `true`.

---

## 4.3 Text Shorthand

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

## 4.4 Parent Attribute

`parent`:

* MUST be treated as a normal attribute.
* MUST accept only NodeRef values.
* MAY reference a forward-declared node.

---

## 4.5 Block Syntax

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

# 5. Edges

## 5.1 Declaration

An edge MUST use:

```
<node_list> -> <node_list>
```

Nodes MAY be forward-declared.

---

## 5.2 Grouped Edges

A node list MAY contain multiple identifiers separated by commas.

Implementations MUST:

* Expand grouped edges into individual edges.

The following form is invalid:

```
A,B -> C,D
```

If both left and right node lists contain more than one identifier:

* This MUST be treated as a syntax error.

---

## 5.3 Attributes

Edge attributes:

* MUST follow the same syntax as node attributes.
* MUST NOT contain NodeRef values.
* MUST NOT contain duplicate keys within a single declaration.

---

## 5.4 Text Shorthand

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

## 5.5 Duplicate Edges

Edges:

* MUST be preserved exactly as declared.
* MUST allow duplicates.
* MUST be treated as a multiset.

Edges MUST NOT have intrinsic identity.

---

## 5.6 Edge Membership

Edges MAY declare set membership via:

```
@setName
```

Membership:

* MUST accumulate.
* MAY cause implicit set creation.

---

# 6. Sets

## 6.1 Declaration

A set MUST be declared using:

```
set <IDENT>
```

Multiple declarations:

* MUST merge.
* MUST use last-write-wins for attribute conflicts.

---

## 6.2 Attributes

Set attributes:

* MUST follow attribute syntax.
* MUST NOT contain NodeRef values.
* MUST NOT contain duplicate keys within a single declaration.

---

## 6.3 Membership

Sets:

* MUST NOT declare membership in other sets.
* MUST NOT contain membership expressions.

---

## 6.4 Implicit Sets

If a set is referenced via `@` but not declared:

* The set MUST be implicitly created.
* The set MUST initially have no attributes.
* Implementations SHOULD emit a linter warning.

---

# 7. Attribute Semantics

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

# 8. Multiple Declarations

Nodes and sets:

* MUST support multiple declarations.
* MUST merge attributes.
* MUST use last-write-wins for conflicts.
* MUST accumulate membership.

---

# 9. Canonicalisation

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

* Ordering of nodes, edges, and sets in serialisation MAY be implementation-defined.
* Canonical form MUST parse into an identical internal representation.

---

# 10. Errors

The following conditions MUST produce a syntax error:

* Duplicate attribute keys within a single declaration.
* Grouped edges on both sides of `->`.
* NodeRef values in edge attributes.
* NodeRef values in set attributes.
* Use of reserved keywords as identifiers.

## 10.1 Example Invalid Syntax

The following are examples of syntax errors that MUST be rejected:

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

# 11. Linter Warnings (Non-Normative)

Implementations SHOULD provide warnings for:

* Implicit set creation.
* Parent override inside a block.
* NodeRef type override across declarations.
* Node and set name collision.

Warnings MUST NOT prevent parsing.

---

# 12. Conformance

An implementation conforms to GSL-Lang v1.0 if it:

* Accepts all valid input defined by this specification.
* Rejects all invalid input defined by this specification.
* Produces canonical output meeting Section 9 requirements.
