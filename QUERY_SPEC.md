---
name: gsl-query-language-spec
description: RFC specification for the GSL Query Language. Covers subgraph extraction, predicates (equality, existence, set membership), traversal (follow), transformations (make), removal operations, node collapse, named graphs, graph algebra (union, intersection, difference, symmetric difference), and pipeline composition. Use when implementing, extending, or understanding GSL query semantics.
---

# GSL Query Language Specification

Version 0.3.0 (Revised Draft RFC)

---

# 1. Motivation

GSL defines a declarative language for describing directed graphs. However, once a graph is defined, users need the ability to **query, filter, transform, and compose graphs** without resorting to general-purpose programming.

The GSL Query Language provides a **pipeline-oriented transformation language** for performing these operations.

Design goals:

* **Composability** — queries form pipelines of graph transformations
* **Determinism** — identical inputs produce identical outputs
* **Graph preservation** — every stage produces a valid GSL graph
* **Minimality** — small orthogonal language surface

---

# 2. Conceptual Model

The query language is a **graph transformation pipeline**.

Each stage performs:

```
Graph → Graph
```

A query therefore consists of a sequence of graph transformations applied to a **working graph**.

---

# 3. Graph Model

## 3.1 Node Identity

Nodes are identified **only by their identifier**.

Two nodes are equal if and only if:

```
node_a.id == node_b.id
```

Attributes do **not** affect node identity.

---

## 3.2 Edge Representation

Edges have **no identity**.

Edges are considered **equal** when the following values match:

```
source node id
target node id
attribute set
```

Graphs therefore contain a **multiset of edges**.

Duplicate edges are permitted and preserved.

---

## 3.3 Attribute Equality

Attribute equality is **type-sensitive**.

Examples:

```
"42" ≠ 42
true ≠ "true"
```

Implementations MUST retain attribute value types during query evaluation.

---

## 3.4 Set Membership

Nodes and edges may belong to sets.

Set membership is preserved during query operations unless explicitly modified.

If a set referenced in a predicate does not exist, evaluation proceeds with the rules defined in Section 7.

---

# 4. Lexical Structure

## 4.1 File Extension

Query files SHOULD use:

```
.gsql
```

---

## 4.2 Reserved Keywords

```
subgraph
from
as
traverse
make
remove
collapse
into
where
AND
in
out
both
exists
not
orphans
all
```

---

## 4.3 Named Graph Identifiers

Named graph identifiers MUST match:

```
[A-Z][A-Z0-9_]*
```

They MUST be uppercase.

Named graphs share a **single namespace within the query**.

---

## 4.4 Attribute Paths

Attribute references use:

```
node.<attribute>
edge.<attribute>
```

---

## 4.5 Comments

Comments begin with:

```
#
```

and continue to end of line.

---

## 4.6 Pipeline Operator

Pipeline stages are separated by:

```
|
```

Pipelines are evaluated **left to right**.

---

# 5. Query Pipeline Model

A query consists of sequential stages:

```
input_graph → stage₁ → stage₂ → … → stageₙ → result_graph
```

Each stage receives a **working graph** and produces a **new working graph**.

---

## 5.1 Stage Types

| Stage                  | Description           |
| ---------------------- | --------------------- |
| `subgraph`             | Extract subgraph      |
| `make`                 | Assign attributes     |
| `remove edge`          | Remove edges          |
| `remove attribute`     | Remove attributes     |
| `remove orphans`       | Remove isolated nodes |
| `collapse`             | Merge nodes           |
| `from`                 | Change working graph  |
| `(<pipeline>) as NAME` | Bind named graph      |
| graph algebra          | Combine named graphs  |

---

## 5.2 Implicit Source

If a query begins without a `from` stage, the working graph is the **input graph**.

---

# 6. Subgraph Extraction

```
subgraph <predicate>
```

Produces a new graph consisting of elements that match the predicate.

The predicate determines whether the operation targets **nodes** or **edges**.

---

## 6.1 Predicate Target Detection

The target type is determined from attribute references in the predicate.

Examples:

```
node.team == "payments"
```

→ node predicate

```
edge.protocol == "http"
```

→ edge predicate

Mixed references are an **error**.

---

## 6.2 Edge Predicate Subgraph

For an edge predicate:

1. Evaluate predicate against all edges
2. Include all matching edges
3. Include their source and target nodes

No other edges are included.

Duplicate edges are preserved.

---

## 6.3 Node Predicate Subgraph

For a node predicate:

1. Evaluate predicate against all nodes
2. Include all matching nodes
3. Include edges whose **source and target nodes are both included**

Edges connecting to nodes outside the result set are excluded.

This behaviour avoids unintentionally expanding the subgraph.

---

# 7. Predicates

Predicates evaluate to **true or false** for a node or edge.

---

## 7.1 Equality

```
node.<attr> == <value>
edge.<attr> == <value>
```

True only if:

* attribute exists
* values are equal
* types match

---

## 7.2 Inequality

```
node.<attr> != <value>
edge.<attr> != <value>
```

True only if:

* attribute exists
* value differs

Missing attributes evaluate **false**.

---

## 7.3 Attribute Exists

```
node.<attr> exists
edge.<attr> exists
```

---

## 7.4 Attribute Not Exists

```
node.<attr> not exists
edge.<attr> not exists
```

---

## 7.5 Set Membership

```
node in @set
edge in @set
```

If the set does not exist, predicate evaluates **false**.

---

## 7.6 Set Non-Membership

```
node not in @set
edge not in @set
```

If the set does not exist, predicate evaluates **true**.

---

## 7.7 Predicate Composition

```
predicate1 AND predicate2
```

Both predicates MUST target the same element type.

---

# 8. Traversal (`traverse`)

Traversal expands a subgraph structurally.

```
subgraph <predicate> traverse <direction> <depth>
```

Traversal operates **after the subgraph is constructed**.

Traversal is **structure-based**, not predicate-based.

---

## 8.1 Direction

```
in
out
both
```

---

## 8.2 Depth

```
1      → one hop
N      → N hops
all    → unlimited traversal
```

---

## 8.3 Traversal Algorithm

Traversal MUST:

1. Begin from nodes produced by the subgraph stage
2. Use breadth-first traversal
3. Maintain a visited node set
4. Continue until depth limit reached or frontier empty

Edges encountered during traversal MUST be included in the result graph.

Edges MUST be included even if both endpoints were previously visited.

This ensures deterministic edge inclusion.

---

# 9. Transformation Operations

## 9.1 Attribute Assignment

```
make node.<attr> = <value> where <predicate>
make edge.<attr> = <value> where <predicate>
```

The attribute is created or overwritten.

Graph structure is unchanged.

---

# 10. Removal Operations

## 10.1 Remove Edge

```
remove edge where <predicate>
```

Matching edges are removed.

Nodes remain.

---

## 10.2 Remove Attribute

```
remove node.<attr> where <predicate>
remove edge.<attr> where <predicate>
```

---

## 10.3 Remove Orphans

```
remove orphans
```

Removes nodes with **no incident edges**.

A self-loop counts as an incident edge.

Therefore a node with a self-loop is **not an orphan**.

---

# 11. Collapse

Collapse merges multiple nodes into a single node.

```
collapse into <id> where <predicate>
```

`<id>` MUST be a valid node identifier.

---

## 11.1 Collapse Algorithm

1. Select nodes matching predicate
2. Create node `<id>`
3. Merge attributes of collapsed nodes
4. Redirect edges
5. Remove original nodes

---

## 11.2 Edge Rewriting

For every edge:

```
A → B
```

If A or B was collapsed:

```
A → B
```

becomes

```
<id> → B
A → <id>
```

---

## 11.3 Internal Edges

Edges between collapsed nodes are removed.

---

## 11.4 Edge Deduplication

Collapse MAY produce duplicate edges.

During collapse **only**, duplicate edges MUST be deduplicated.

Edge equality is defined as:

```
source id
target id
attribute set
```

This deduplication rule applies **only to collapse**.

All other operations preserve duplicate edges.

---

# 12. Named Graphs

Named graphs store intermediate pipeline results.

---

## 12.1 Binding

```
(<pipeline>) as NAME
```

The result graph is stored under NAME.

---

## 12.2 Immutability

Named graphs cannot be rebound.

Attempting to bind a name twice MUST produce an error.

---

## 12.3 From Clause

```
from *
from NAME
```

`*` resets the working graph to the input graph.

---

# 13. Graph Algebra

Named graphs can be combined.

| Operator | Meaning              |
|----------|----------------------|
| `+`      | union                |
| `&`      | intersection         |
| `-`      | difference           |
| `^`      | symmetric difference |

---

## 13.1 Node Merge Rules

If the same node exists in both graphs:

```
left graph attributes applied first
right graph attributes overwrite conflicts
```

Example:

```
node A [team="payments", zone="A"]
+
node A [team="fraud"]
=
node A [team="fraud", zone="A"]
```

---

## 13.2 Edge Merge Rules

Edges from both graphs are combined.

Duplicate edges are preserved.

---

# 14. Grammar (Simplified)

```
query := stage ('|' stage)*

stage :=
      subgraph_stage
    | make_stage
    | remove_stage
    | collapse_stage
    | from_stage
    | binding
    | graph_algebra

subgraph_stage :=
    'subgraph' predicate
    ('traverse' direction depth)?

make_stage :=
    'make' target '=' value 'where' predicate

remove_stage :=
      'remove edge where' predicate
    | 'remove node.' IDENT 'where' predicate
    | 'remove edge.' IDENT 'where' predicate
    | 'remove orphans'

collapse_stage :=
    'collapse into' IDENT 'where' predicate
```

---

# 15. Error Conditions

An implementation MUST produce an error for:

* rebinding named graph identifiers
* mixed predicate targets
* invalid attribute paths
* invalid collapse identifier
* syntax violations

---

# 16. Deferred Topics

The following areas remain intentionally undefined:

* node deletion semantics
* parent attribute semantics
* predicate OR expressions
* attribute inheritance policies
* result construction customization

---

# Implementation Notes

Implementations should maintain:

```
node set
edge multiset
set memberships
attribute maps
```

Traversal must maintain a **visited node set** to ensure termination on cyclic graphs.