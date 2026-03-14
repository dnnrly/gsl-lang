---
name: gsl-query-ai-guide
description: Complete GSL Query Language reference for AI agents and LLMs. Self-contained guide covering syntax, semantics, all expression types, predicates, common patterns, and edge cases. Use as agent context when implementing or analyzing queries.
---

# GSL Query Language — AI Agent Guide

**Target audience:** AI agents, LLMs, and autonomous tools that need to understand, implement, or reason about GSL queries.

This guide is self-contained. Copy it entirely into your agent context when working with queries.

---

## 1. Core Concepts

### 1.1 Pipeline Model

A query is a **sequence of expressions** connected by `|` (pipe), evaluated **left-to-right**:

```
input_graph → expr₁ → expr₂ → … → result_graph
```

Every expression receives a graph and produces a graph. The value type never changes.

### 1.2 Query State

During evaluation, the system maintains:

```
QueryState:
  input_graph     — the original graph provided to the query
  working_value   — the current graph being transformed
  named_graphs    — map of intermediate results (bound with `as NAME`)
```

### 1.3 Node and Edge Identity

**Nodes:** Identified **only by ID**. Two nodes are equal iff `node_a.id == node_b.id`. Attributes do not affect identity.

**Edges:** Have **no identity**. Two edges are equal iff:
- Source node ID matches
- Target node ID matches  
- Attribute set matches

Graphs contain a **multiset of edges** — duplicates are allowed and preserved (except during `collapse`, which deduplicates).

### 1.4 Attribute Equality

Type-sensitive. Examples:
- `"42" ≠ 42`
- `true ≠ "true"`
- `0 ≠ false`

Implementations must preserve types during query evaluation.

---

## 2. Expression Types (Complete Reference)

### 2.1 Source Selection: `from`

**Syntax:**
```
from *        # Reset to input graph
from NAME     # Switch to named graph
```

**Semantics:** Changes the working graph without modifying named graphs.

**Examples:**
```
subgraph node.team == "payments" | from *
```
After filtering, reset to the input graph.

**Notes:**
- If `from NAME` references an undefined named graph, error
- `from *` always succeeds

---

### 2.2 Filtering: `subgraph`

**Syntax:**
```
subgraph <predicate>
subgraph <predicate> traverse <direction> <depth>
```

**Semantics (Node Predicate):**

When predicate targets `node.*`:
1. Select all nodes matching the predicate
2. Include only edges where **both** source and target are selected
3. If `traverse` is present, expand selection per rules below

**Semantics (Edge Predicate):**

When predicate targets `edge.*`:
1. Select all edges matching the predicate
2. Include source and target nodes of matched edges
3. If `traverse` is present, expand selection per rules below

**Traverse Details:**

After node/edge matching, optionally follow edges:

```
traverse <direction> <depth>
```

**Direction:**
- `in` — follow incoming edges
- `out` — follow outgoing edges
- `both` — follow edges in both directions

**Depth:**
- `1`, `2`, `N` — exact number of hops
- `all` — unlimited hops (until frontier exhausted)

For each hop, apply these rules:
- **Node predicate:** Discover nodes via edges. Keep only edges where both endpoints are in selection.
- **Edge predicate:** Discover edges via direction filter. Keep only matching edges; include endpoints.

**Examples:**

```
# All nodes in "payments" team
subgraph node.team == "payments"

# All edges using gRPC protocol
subgraph edge.protocol == "grpc"

# Payments team and everything they depend on (all hops outward)
subgraph node.team == "payments" traverse out all

# Services that depend on the DB, up to 3 hops
subgraph node.id == "db" traverse in 3
```

**Important:** Cannot combine node and edge predicates in traverse. The direction applies to the predicate type (node or edge).

---

### 2.3 Attribute Assignment: `make`

**Syntax:**
```
make <path> = <value> where <predicate>
```

Where `<path>` is:
- `node.<attr>` — node attribute
- `edge.<attr>` — edge attribute

**Semantics:**
1. Find all elements (nodes or edges) matching the predicate
2. Set the attribute to the value on matching elements
3. Return the modified graph

**Value Types:**
- Strings: `"value"`, `"hello world"`
- Numbers: `42`, `3.14`, `-5`
- Booleans: `true`, `false`

**Examples:**

```
# Mark all payment nodes as "reviewed"
make node.status = "reviewed" where node.team == "payments"

# Set priority on critical edges
make edge.priority = 1 where edge in @critical
```

**Notes:**
- Overwrites existing attributes
- Does not create/remove nodes or edges
- Returns the graph unchanged in structure

---

### 2.4 Edge Removal: `remove edge`

**Syntax:**
```
remove edge where <predicate>
```

**Semantics:**
1. Find all edges matching the predicate
2. Delete those edges
3. Return graph with matching edges removed
4. **Nodes remain** (even if they become orphaned)

**Examples:**

```
# Remove all TCP connections
remove edge where edge.protocol == "tcp"

# Remove deprecated edges
remove edge where edge in @deprecated
```

**Notes:**
- Only affects edges; nodes stay
- Use `remove orphans` afterward to clean up disconnected nodes

---

### 2.5 Attribute Removal: `remove <path>`

**Syntax:**
```
remove node.<attr> where <predicate>
remove edge.<attr> where <predicate>
```

**Semantics:**
1. Find all matching elements
2. Delete the specified attribute
3. Return graph with attributes removed

**Examples:**

```
# Remove temporary flags from old nodes
remove node.tmp where node.tmp exists

# Clear debug info from edges
remove edge.debug where edge.debug exists
```

**Notes:**
- Only deletes attributes, not the elements themselves

---

### 2.6 Orphan Removal: `remove orphans`

**Syntax:**
```
remove orphans
```

**Semantics:**
1. Find all nodes with zero incident edges
2. Delete those nodes
3. Return modified graph

**Incident edges:** Incoming, outgoing, or self-loops. A self-loop counts.

**Examples:**

```
# Clean up after removing edges
subgraph node.team == "payments" | remove edge where edge.protocol == "tcp" | remove orphans
```

**Notes:**
- A node with a self-loop is **not** an orphan
- No predicate — always removes all orphans

---

### 2.7 Node Collapse: `collapse`

**Syntax:**
```
collapse into <id> where <predicate>
```

**Semantics:**

1. Find all nodes matching the predicate
2. Create a new node with ID `<id>`
3. Merge attributes from matched nodes (last-write-wins)
4. Redirect all incoming/outgoing edges to `<id>`
5. Remove original nodes
6. Remove edges between collapsed nodes (internal edges)
7. **Deduplicate edges created by merge**

**Attribute Merging:**

When multiple nodes are collapsed, iterate in order and apply attributes. Last value wins.

Example:
```
Node A [team="payments", zone="B"]
Node B [team="payments", zone="C"]
collapse into GROUP → Node GROUP [team="payments", zone="C"]
```

**Edge Rewriting Example:**

Before:
```
A -> X
B -> Y
C -> A
D -> B
```

After `collapse into GROUP where ...` (collapsing A and B):
```
GROUP -> X
GROUP -> Y
C -> GROUP
D -> GROUP
```

**Deduplication:**

If collapse produces duplicate edges (e.g., both A and B pointed to X), the duplicate is removed. This is the **only** operation that deduplicates edges.

**Examples:**

```
# Merge all platform team nodes into one
collapse into platform_cluster where node.team == "platform"

# Merge all nodes in zone C
collapse into zone_c where node.zone == "C"
```

**Notes:**
- The target ID must be a valid node identifier
- Edges between collapsed nodes are always removed
- This is the only operation that modifies edge cardinality

---

### 2.8 Named Graph Binding: `as`

**Syntax:**
```
(<pipeline>) as NAME
```

Where `NAME` matches `[A-Z][A-Z0-9_]*` (uppercase identifiers).

**Semantics:**
1. Evaluate the pipeline in parentheses
2. Store result in `named_graphs[NAME]`
3. Continue with the graph (not the stored graph)
4. If NAME is already bound, error (immutable)

**Examples:**

```
(subgraph node.team == "payments") as PAY
| from *
| (subgraph node.team == "identity") as ID
| PAY + ID
```

Result: Union of payments and identity teams.

**Notes:**
- Names are case-sensitive and must be uppercase
- Cannot rebind a name
- Binding does not change the working graph; it saves intermediate results

---

### 2.9 Graph Algebra

**Syntax:**
```
NAME1 + NAME2      # Union
NAME1 & NAME2      # Intersection
NAME1 - NAME2      # Difference
NAME1 ^ NAME2      # Symmetric difference
```

**Semantics (Union: `+`):**

The result contains:
- All nodes from both graphs
- All edges from both graphs
- Duplicate edges preserved

When same node appears in both:
- Left attributes applied first
- Right attributes overwrite conflicts

**Example:**
```
node A [team="gateway", zone="A"]  +  node A [team="fraud"]
→ node A [team="fraud", zone="A"]
```

**Semantics (Intersection: `&`):**

The result contains:
- Only nodes in **both** graphs
- Only edges in **both** graphs (exact match: source, target, attributes)

**Semantics (Difference: `-`):**

The result contains:
- Nodes in left but not in right
- Edges in left but not in right

**Semantics (Symmetric Difference: `^`):**

The result contains:
- Nodes in exactly one graph
- Edges in exactly one graph

**Examples:**

```
# Union: all critical and deprecated
(subgraph node in @critical) as CRIT
| from *
| (subgraph node in @deprecated) as DEP
| CRIT + DEP

# Difference: critical but not deprecated
CRIT - DEP

# Intersection: both critical AND deprecated
CRIT & DEP
```

**Notes:**
- Both operands must be named graphs
- Edge equality includes attributes

---

## 3. Predicates (Complete Reference)

Predicates appear in `subgraph`, `make`, `remove`, and `collapse` expressions.

### 3.1 Predicate Structure

Predicates have two forms:

**Single condition:**
```
<target> <condition>
```

**Compound (AND only):**
```
<target> <condition1> AND <target> <condition2>
```

Where `<target>` is:
- `node` — targets nodes (no attribute name needed for `exists` checks)
- `node.<attr>` — targets node attribute
- `edge` — targets edges
- `edge.<attr>` — targets edge attribute

### 3.2 Condition Types

#### Equality

```
<target> == <value>
```

True if attribute equals value (type-sensitive).

Examples:
```
node.team == "payments"
edge.protocol == "grpc"
node.zone == "A"
```

#### Inequality

```
<target> != <value>
```

True if attribute exists **and** differs from value. Missing attributes evaluate **false**.

Example:
```
node.zone != "C"
```

If node has no `zone` attribute, this is false.

#### Existence

```
<target> exists
<target> not exists
```

True if attribute is present/absent. Works with `node.<attr>` and `edge.<attr>`.

Examples:
```
node.team exists
edge.debug not exists
```

#### Set Membership

```
<element> in @<setname>
<element> not in @<setname>
```

Where `<element>` is `node` or `edge`.

True if element belongs to the set.

**Important:** If set doesn't exist:
- `in @missing` → **false**
- `not in @missing` → **true**

Examples:
```
node in @critical
edge not in @deprecated
```

#### Compound Predicates (AND only)

```
<cond1> AND <cond2>
```

Both conditions must be true. Cannot mix node and edge predicates.

Examples:
```
node.team == "payments" AND node.zone == "B"
edge.protocol == "grpc" AND edge in @critical
```

**Limitation:** No `OR` support.

### 3.3 Value Types

- **Strings (quoted):** `"value"`, `"hello world"`
- **Numbers:** `42`, `3.14`, `-1`
- **Booleans:** `true`, `false`

---

## 4. Lexical Rules

### 4.1 Keywords

Reserved words (case-sensitive):
```
subgraph, from, make, remove, collapse, into, traverse, where, exists, not, in, and, or, all, both, out, as
```

### 4.2 Named Graph Identifiers

Format: `[A-Z][A-Z0-9_]*`

Must be uppercase. Examples: `PAY`, `CRITICAL_NODES`, `ZONE_A`.

### 4.3 Node/Edge/Set Identifiers

Format: `[a-zA-Z_][a-zA-Z0-9_]*`

Examples: `api`, `payments_db`, `_internal`.

### 4.4 Comments

Lines starting with `#`:
```
# This is a comment
```

### 4.5 Pipe Operator

`|` separates expressions. Whitespace around pipes is ignored.

```
expr1 | expr2 | expr3
```

---

## 5. Common Patterns

### Pattern 1: Filter by Attribute

```
subgraph node.team == "payments"
```

Result: All nodes in the payments team with edges between them.

### Pattern 2: Filter + Expand

```
subgraph node.team == "payments" traverse out 1
```

Result: Payments team plus their direct dependencies.

### Pattern 3: Filter + Cleanup

```
subgraph node.team == "payments" | remove orphans
```

Result: Payments team, with any isolated nodes removed.

### Pattern 4: Multi-Step Transformation

```
subgraph node.zone == "A" traverse out all
| remove edge where edge.protocol == "tcp"
| remove orphans
| collapse into zone_a where node.team == "platform"
```

1. Find zone A and all descendants
2. Remove TCP edges
3. Remove orphaned nodes
4. Merge platform team into single node

### Pattern 5: Graph Union

```
(subgraph node.team == "payments") as PAY
| from *
| (subgraph node.team == "identity") as ID
| PAY + ID
```

Result: All nodes from both teams.

### Pattern 6: Difference

```
(subgraph node in @critical) as CRIT
| from *
| (subgraph node in @deprecated) as DEP
| CRIT - DEP
```

Result: Critical nodes that are not deprecated.

---

## 6. Edge Cases and Subtleties

### 6.1 Traverse with No Matching Nodes/Edges

If the initial subgraph matches no elements, traverse discovers nothing. Result is empty graph.

### 6.2 Cyclic Graphs

Traversal must detect visited nodes to terminate. The implementation should maintain a visited set per traversal operation.

### 6.3 Missing Attributes in Inequality

```
node.zone != "C"
```

If a node has no `zone` attribute, this predicate is **false**.

### 6.4 Empty Named Graphs

Named graphs can be empty (no nodes, no edges). Algebra operations handle empty graphs (union with empty returns original, intersection with empty returns empty, etc.).

### 6.5 Collapse with No Matching Nodes

If the predicate matches no nodes, collapse is a no-op. No error.

### 6.6 Duplicate Edges After Non-Collapse Operations

All operations except `collapse` preserve duplicate edges exactly. If your input has two `A -> B` edges, they remain after any operation except `collapse`.

### 6.7 Self-Loops in Collapse

If a collapsed node had a self-loop, it's removed (internal edge). The new collapsed node does not automatically gain a self-loop.

### 6.8 Attribute Ordering in Collapse

Attributes merge in order, with last-write-wins. If nodes are:
```
A [x=1, y=2]
B [x=3, z=4]
C [x=5, y=6]
collapse into G where ...
```

Attributes apply: `A` then `B` then `C`. Result: `G [x=5, y=6, z=4]`.

### 6.9 Set Membership is Preserved

Unless explicitly removed, set membership is preserved through all operations.

### 6.10 Predicate Target Consistency

In a single compound predicate, all conditions must target the same element type:

```
# Valid
node.team == "payments" AND node.zone == "A"

# Invalid
node.team == "payments" AND edge.protocol == "grpc"
```

---

## 7. Error Conditions

An implementation MUST produce an error for:

- **Rebinding named graphs:** Attempting `(expr) as NAME` when `NAME` already exists
- **Undefined named graphs:** Using `from NAME` or `NAME + NAME2` with undefined `NAME`
- **Mixed predicate targets:** `node.x == "a" AND edge.y == "b"` in same condition
- **Invalid collapse ID:** `<id>` is not a valid identifier
- **Syntax violations:** Malformed expressions
- **Invalid operators:** Unknown keywords or operators

---

## 8. Grammar (Simplified EBNF)

```
query := expression ('|' expression)*

expression :=
      subgraph_expr
    | make_expr
    | remove_expr
    | collapse_expr
    | from_expr
    | binding_expr
    | algebra_expr

subgraph_expr :=
    'subgraph' predicate ('traverse' direction depth)?

direction := 'in' | 'out' | 'both'
depth := INTEGER | 'all'

make_expr :=
    'make' attr_path '=' value 'where' predicate

remove_expr :=
      'remove edge where' predicate
    | 'remove' attr_path 'where' predicate
    | 'remove orphans'

collapse_expr :=
    'collapse into' IDENT 'where' predicate

from_expr :=
    'from' ('*' | UPPER_IDENT)

binding_expr :=
    '(' query ')' 'as' UPPER_IDENT

algebra_expr :=
    UPPER_IDENT ('+' | '&' | '-' | '^') UPPER_IDENT

predicate := condition ('AND' condition)*

condition :=
      element_ref '==' value
    | element_ref '!=' value
    | element_ref 'exists'
    | element_ref 'not exists'
    | element_ref 'in' '@' IDENT
    | element_ref 'not in' '@' IDENT

element_ref := 'node' ('.' IDENT)? | 'edge' ('.' IDENT)?

attr_path := ('node' | 'edge') '.' IDENT

value := STRING | NUMBER | BOOLEAN
```

---

## 9. Implementation Checklist

When implementing a query parser/evaluator:

- [ ] Preserve edge ordering (duplicates allowed)
- [ ] Preserve duplicate edges in all operations except `collapse`
- [ ] Maintain visited set during `traverse` for cycle detection
- [ ] Type-sensitive attribute equality
- [ ] Last-write-wins for attribute merges in `collapse`
- [ ] Cannot rebind named graphs (immutable)
- [ ] `from NAME` is an error if NAME undefined
- [ ] Predicate targets must match (no mixing `node.*` and `edge.*` in AND)
- [ ] Missing attributes in inequality predicates evaluate false
- [ ] `in @missing_set` is false; `not in @missing_set` is true
- [ ] Edges with all attributes matching are equal
- [ ] Self-loops are incident edges (not orphaned)

---

## 10. Examples from Real Queries

### Example 1: Service Dependencies

**Query:**
```
subgraph node.type == "service" traverse out all
```

**What it does:** Find all services and everything they depend on, recursively.

### Example 2: Critical Path

**Query:**
```
subgraph node in @critical traverse in all
```

**What it does:** Find all nodes that (directly or indirectly) feed into critical nodes.

### Example 3: Network Boundary

**Query:**
```
subgraph node.zone == "A" | remove edge where edge.zone != "A" | remove orphans
```

**What it does:** Keep only nodes and edges within zone A, removing cross-zone edges and orphans.

### Example 4: Team Ownership

**Query:**
```
(subgraph node.team == "payments") as PAYMENTS
| from *
| (subgraph node.team == "fraud") as FRAUD
| PAYMENTS & FRAUD
```

**What it does:** Find services owned by both payments and fraud teams (shared ownership).

### Example 5: Remove Stale Connections

**Query:**
```
remove edge where edge.status == "deprecated"
| remove edge where edge.protocol == "http"
| remove orphans
```

**What it does:** Remove deprecated and HTTP edges, then clean up orphaned services.

---

## 11. Quick Syntax Reference

| Task | Syntax |
|------|--------|
| Filter nodes | `subgraph node.<attr> <cond>` |
| Filter edges | `subgraph edge.<attr> <cond>` |
| Expand search | `traverse <dir> <depth>` |
| Follow outbound | `traverse out all` |
| Follow inbound | `traverse in all` |
| Assign attribute | `make node.<attr> = <val> where <pred>` |
| Delete edges | `remove edge where <pred>` |
| Delete attribute | `remove node.<attr> where <pred>` |
| Clean orphans | `remove orphans` |
| Merge nodes | `collapse into <id> where <pred>` |
| Save result | `(<pipeline>) as NAME` |
| Union graphs | `A + B` |
| Intersection | `A & B` |
| Difference | `A - B` |
| Symmetric diff | `A ^ B` |
| Use saved result | `from NAME` |
| Reset to input | `from *` |
| Equality test | `node.x == "value"` |
| Inequality test | `node.x != "value"` |
| Attribute exists | `node.x exists` |
| Set membership | `node in @setname` |
| Compound test | `node.x == "a" AND node.y == "b"` |

