Excellent — this is where we remove all “implementation-defined” wiggle room and lock this into something production-safe and testable.
# GSL Query Language (GQL)

## Version 0.1 — Production Specification

**Status:** Stable
**Scope:** Querying a single in-memory GSL graph
**Output:** Deterministic subgraph

---

# 1. Conformance

An implementation conforms to GQL v0.1 if and only if:

* All semantics in this document are followed exactly.
* No behaviour is left implementation-defined.
* All required error conditions produce defined failures.
* All queries are deterministic.

---

# 2. Data Model

## 2.1 Graph

A graph consists of:

* A finite set of **Nodes**
* A finite set of **Directed Edges**

---

## 2.2 Node

A Node MUST contain:

* `ID` (string, globally unique)
* `Attributes` (map[string]Value)

---

## 2.3 Edge

An Edge MUST contain:

* `From` (Node ID)
* `To` (Node ID)
* `Attributes` (map[string]Value)

Edges are directed.

---

# 3. Query Evaluation Model

## 3.1 Global Rules

1. A query operates on exactly one input graph.
2. Evaluation MUST be deterministic.
3. All pipelines evaluate left-to-right.
4. Parenthesised pipelines are evaluated independently.
5. All intermediate results are NodeSets.

---

## 3.2 NodeSet Definition

A NodeSet is a mathematical set of Node IDs.

Duplicate nodes MUST NOT exist.

---

## 3.3 Result Subgraph Construction (STRICT)

After final NodeSet is computed:

The resulting subgraph MUST contain:

* All nodes in the NodeSet.
* ALL edges from the original graph such that:

  ```
  edge.From ∈ NodeSet AND edge.To ∈ NodeSet
  ```

Traversal history MUST NOT affect final edge inclusion.

This rule removes all ambiguity.

---

# 4. Start Step

## Syntax

```
start "ID" [, "ID"...]
```

## Semantics

1. All specified IDs MUST exist.
2. If any ID does not exist, evaluation MUST fail with error:

   ```
   UnknownNodeID
   ```
3. Result NodeSet = set of specified nodes.

---

# 5. Flow Step

## Syntax

```
flow out|in|both [recursive]
[where edge.<attr> <op> <value>]
```

---

## 5.1 Direction

Given current NodeSet S:

* `out`: traverse edges where edge.From ∈ S
* `in`: traverse edges where edge.To ∈ S
* `both`: union of both directions

---

## 5.2 Edge Filtering

If edge filter exists:

* It MUST be evaluated before traversal.
* Only edges satisfying predicate are eligible.

If attribute does not exist on an edge:

* Predicate evaluates to false.

Invalid attribute reference MUST NOT fail execution.

---

## 5.3 Non-Recursive Flow

Result NodeSet =

```
S ∪ AdjacentNodes
```

Where AdjacentNodes are nodes reachable in one step via eligible edges.

---

## 5.4 Recursive Flow

Recursive traversal MUST:

* Use BFS or DFS (order irrelevant)
* Track visited nodes
* Terminate on cycles
* Visit each node at most once

Result NodeSet =

```
S ∪ AllTransitivelyReachableNodes
```

---

## 5.5 Depth Limiting

Not supported in v0.1.

---

# 6. Node Filter Step

## Syntax

```
where <attr> <op> <value>
```

## Semantics

1. Applied to current NodeSet.
2. If node lacks attribute → predicate evaluates false.
3. Invalid operator or type mismatch MUST produce error:

   ```
   InvalidPredicate
   ```
4. Result NodeSet = nodes satisfying predicate.

---

# 7. Predicate Semantics

## 7.1 Supported Operators

* `=`
* `!=`
* `<`
* `<=`
* `>`
* `>=`

## 7.2 Type Rules

Allowed types:

* string
* number
* boolean

Comparison rules:

* Types MUST match.
* If types differ → error `InvalidPredicate`.

---

# 8. Combinators

Combinators operate on two complete pipelines.

Each side MUST be evaluated independently against the original graph.

---

## 8.1 union

```
Result = Left ∪ Right
```

---

## 8.2 intersect

```
Result = Left ∩ Right
```

---

## 8.3 minus

```
Result = Left − Right
```

---

## 8.4 Deterministic Guarantee

Combinator evaluation MUST NOT depend on:

* Execution order
* Traversal order
* Memory layout

---

# 9. Prune (Inline Minus)

Syntax:

```
<pipeline> minus (<pipeline>)
```

Semantics:

1. Evaluate left → L
2. Evaluate right → R
3. Result = L − R

Equivalent to combinator minus.

No special behaviour.

---

# 10. Error Handling (STRICT)

The following MUST produce runtime errors:

| Condition                  | Error              |
| -------------------------- | ------------------ |
| Unknown node in start      | `UnknownNodeID`    |
| Invalid predicate operator | `InvalidPredicate` |
| Type mismatch in predicate | `InvalidPredicate` |
| Malformed AST              | `InvalidQuery`     |

The engine MUST fail fast and return an error object.

---

# 11. Determinism Guarantees

Given identical:

* Input graph
* Query string

The output MUST be bitwise identical if serialized canonically.

Implementations MUST:

* Use set semantics
* Avoid reliance on map iteration order

---

# 12. Performance Guarantees (Non-Normative)

An implementation SHOULD:

* Represent NodeSet as hash set
* Avoid revisiting nodes during recursion
* Avoid quadratic set operations

---

# 13. Explicit Non-Goals (v0.1)

The following are NOT part of v0.1:

* Aggregations
* Scalar functions
* Edge-only queries
* Ordering guarantees
* Depth limits
* Node collapsing
* Path enumeration

These MAY appear in v2.0.

---

# 14. Canonical Example

Given:

```
A -> B
B -> C
C -> D
```

Query:

```
start "A" | flow out recursive
```

Result NodeSet:

```
{A, B, C, D}
```

Result Edges (strict inclusion rule):

```
A->B
B->C
C->D
```

---

# 15. Canonical Evaluation Algorithm

1. Parse query → AST
2. Evaluate pipeline → final NodeSet
3. Construct result graph:

    * Include nodes in NodeSet
    * Include all edges with endpoints in NodeSet
4. Return subgraph

---

# Final Properties of v0.1

This version:

* Has zero implementation-defined behaviour
* Has no ambiguous edge semantics
* Is fully deterministic
* Is testable via pure set equality
* Is safe for production use
