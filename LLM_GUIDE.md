---
name: gsl-language-guide
description: Complete reference for GSL syntax, semantics, and language design. Covers nodes, edges, sets, attributes, and parsing behavior. Use when learning GSL, writing GSL files, or understanding how the language works.
---

# GSL (Graph Specification Language) - Complete Guide for LLMs

This document contains everything needed to understand and write GSL files and use the language correctly.

## What is GSL?

GSL is a small, declarative language for describing **directed graphs** with attributes and set-based grouping. It is:
- Human-readable and easy to parse
- Deterministic and canonicalisable
- Designed for tooling, transformation, and programmatic analysis
- NOT a visual language—purely textual

## GSL Syntax Reference

### Basic Nodes

```gsl
node A
node B [flag]
node C [weight=2, color="red"]
node D: "Display Text"
```

**Rules:**
- Node IDs must match: `[A-Za-z_][A-Za-z0-9_]*`
- Cannot use reserved keywords: `node`, `set`, `true`, `false`
- Attributes use `[key]` or `[key=value]` syntax
- Text shorthand `: "string"` sets the `text` attribute
- Text shorthand cannot be combined with attributes in same declaration

### Basic Edges

```gsl
A->B
A,B->C
D->E,F
A->B [weight=1.5, color="blue"]
A->B: "label"
A->B [weight=1.5] @setname
```

**Rules:**
- Edges expand grouped nodes automatically: `A,B->C` becomes `A->C` and `B->C`
- Cannot have grouped nodes on both sides: `A,B->C,D` is a syntax error
- Attributes supported same as nodes (no NodeRef types allowed)
- Text shorthand supported
- Duplicate edges are allowed and preserved
- Edges can have text shorthand and attributes but not both in same declaration

### Sets (Groupings)

```gsl
set critical [backup=true]
set services [env="production"]

node ServiceA @critical @services
node ServiceB @critical
A->B @services
```

**Rules:**
- Sets are named groupings that nodes and edges can belong to
- Sets are declared with: `set <name> [attributes]`
- Membership added via `@setname` suffix after node/edge declaration
- Membership accumulates across multiple declarations
- Sets created implicitly when referenced but not declared (generates warning)

### Parent-Child Relationships

```gsl
# Block syntax (syntactic sugar)
node Parent {
  node Child1
  node Child2
}

# Explicit equivalent
node Parent
node Child1 [parent=Parent]
node Child2 [parent=Parent]
```

**Rules:**
- Block syntax is shorthand for implicit `parent` attributes
- Explicit parent in block overrides implicit (generates warning)
- `parent` attribute uses NodeRef type (can reference any node ID)
- Blocks can be nested

### Attribute Types and Values

Attributes can have these value types:

```gsl
node A [
  text="string",
  count=42,
  ratio=3.14,
  enabled=true,
  disabled=false,
  parent=OtherNode,
  flag
]
```

**Value types:**
- **String**: `"anything in quotes"`, supports escape sequences: `\"`, `\\`, `\n`, `\t`
- **Number**: Integer or float, no sign/exponent: `42`, `3.14`
- **Boolean**: `true` or `false` (must be lowercase)
- **NodeRef**: Bare identifier (only allowed in node attributes, not edges/sets): `OtherNode`
- **Empty**: Bare key with no value means empty attribute: `flag`

**Rules:**
- No duplicate keys in single declaration
- Nodes can have `parent` attributes pointing to other nodes
- Edges cannot have NodeRef values
- Sets cannot have NodeRef values

### Comments

```gsl
# This is a comment
node A  # Inline comments work too
# set B  # Commented out
```

## Complete GSL Example

```gsl
# Microservices architecture example
set frontend [color="blue", visible=true]
set backend [color="green"]
set critical [backup=true]

# Frontend services
node WebUI [text="Web Interface"] @frontend
node Dashboard [text="Dashboard"] @frontend

# Backend services
node API [text="API Server"] @backend @critical
node Database [text="PostgreSQL"] @backend @critical
node Cache [text="Redis"] @backend

# API structure
node AuthModule [parent=API, timeout=30] @API
node DataModule [parent=API, timeout=60] @API

# Connections
WebUI -> API [protocol="REST", timeout=5000]
Dashboard -> API
API -> Database [pool_size=20]
API -> Cache [ttl=3600]
Database -> Cache

# Complex edge declaration
AuthModule -> Database, Cache
```

## Language Design Notes

### Parsing Behavior

- **Lenient parsing**: Parse succeeds even with warnings (implicit sets, name collisions)
- **Last-write-wins**: Multiple declarations merge with attribute conflicts resolved by last occurrence
- **Warnings are non-fatal**: Parser returns both graph and warning list; check both

### Graph Properties

- **No schema validation**: GSL doesn't validate graph structure (no acyclicity checking, tree validity, etc.)
- **Duplicate edges allowed**: The same edge can be declared multiple times; all are preserved
- **Set membership is separate**: Nodes and edges have set membership tracked independently
- **Parent is just an attribute**: The `parent` attribute is normal except that it has semantic meaning in parent-child relationships
- **Attributes are untyped**: Values are stored without type information; interpretation is up to the consumer

### Serialization

- **Canonical form**: Serialized output is deterministic and parseable
- **Ordering may differ**: Serialized output may reorder elements but parses to semantically equivalent graph
- **Round-trip guarantee**: Parsing and serializing multiple times produces consistent results
- **No data loss**: All information is preserved (attributes, sets, duplicates, etc.)

### Warning Types

```
- "implicit set creation: %q"        // Set used but never declared
- "%d:%d: parent override inside block"  // Explicit parent differs from block parent
- "node and set name collision: %q"  // Same ID used as both node and set
```

Warnings are informational only—parsing continues.

## Common Gotchas

1. **Text shorthand and attributes clash**: Can't do `node A: "text" [attr=1]` - split into separate declarations
2. **Grouped edges on both sides fails**: `A,B->C,D` is syntax error, must be `A,B->C` or `A->C,D`
3. **Sets accumulate**: Multiple `@setname` on same node adds to existing set membership
4. **Implicit sets create warnings**: Using `@undeclared` without `set undeclared` produces warning
5. **No set-of-sets**: Sets can't contain other sets, only nodes and edges can be in sets
6. **Parent is attribute**: Setting parent doesn't create hierarchical structure—it's just an attribute
7. **NodeRef only in nodes**: Can't put `parent=SomeNode` in edge or set attributes

## Quick Reference

**Write a simple graph:**
```gsl
set critical
node A @critical
node B
A -> B [weight=2]
```

**For programmatic use**, see the appropriate language guide:
- **Go**: See `GO_GUIDE.md`
- **Other languages**: Implement parser following `SPEC.md` and `GRAMMAR.md`

That's everything you need to read and write GSL correctly!
