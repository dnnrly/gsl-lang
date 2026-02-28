# GSL (GSL-Lang)

**GSL** is a small, declarative language for describing directed graphs with attributes and set-based grouping.

It is designed to be:

* Human-readable
* Deterministic
* Canonicalisable
* Easy to parse
* Easy to diff

GSL is not a visual graph language.
It is a textual graph representation designed for tooling, transformation, and programmatic analysis.

---

## What GSL Describes

A GSL document defines:

* Nodes
* Directed edges
* Sets (groupings)
* Arbitrary attributes on all of the above

---

## Quick Example

```gsl
# Declare sets
set flow [color="blue"]

# Declare nodes
node A: "Start" @flow
node B [flag]

# Declare edges
A->B [weight=1.2] @flow
```

This defines:

* 2 nodes
* 1 directed edge
* 1 set
* Attributes on nodes, edges, and sets

---

## Core Features

### Nodes

```gsl
node A
node B [flag, weight=2]
node C: "Hello"
```

### Edges

```gsl
A->B
A,B->C
C->D,E
```

Grouped edges expand automatically.

Duplicate edges are allowed.

---

### Sets and Membership

```gsl
set cluster [visible]

node A @cluster
A->B @cluster
```

Sets are named groupings.
Membership accumulates across declarations.

---

### Parent Relationships

```gsl
node C {
    node D
}
```

This is syntactic sugar for:

```gsl
node D [parent=C]
```

`parent` is treated as a normal attribute.

---

## Design Goals

GSL is designed to:

* Round-trip cleanly
* Produce a canonical internal representation
* Merge repeated declarations
* Preserve duplicate edges
* Keep semantics simple and explicit

It intentionally avoids:

* Edge identity
* Schema enforcement
* Graph correctness constraints (acyclicity, tree validity, etc.)

---

## Canonical Behaviour

A compliant parser must ensure:

```
parse(serialize(parse(input))) == parse(input)
```

Grouped edges expand.
Blocks become explicit `parent` attributes.
Implicit sets are materialised.

---

## Status

The formal language specification is defined in:

* [the specification](SPEC.md)
* [the grammar](GRAMMAR.md)

Parser implementation details will be added as development progresses.
