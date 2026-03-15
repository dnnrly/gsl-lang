---
description: Query language fixture test catalog. Quick reference for LLMs and developers to understand test coverage without scanning individual fixtures. Updated for context efficiency.
---

# Query Language Test Fixtures

This directory contains **59 fixture-based integration tests** for the GSL Query Language. Each fixture is a directory containing:
- `graph.gsl` — input graph
- `query.gql` — query to execute
- `result.gsl` — expected output

Tests are loaded and executed by `fixtures_test.go`.

---

## Quick Navigation

| Category | Count | Purpose |
|----------|-------|---------|
| [Subgraph Filtering](#subgraph-filtering) | 12 | Extract nodes/edges matching predicates |
| [Graph Algebra](#graph-algebra) | 7 | Union, intersection, difference, XOR |
| [Make (Assign)](#make-assign) | 5 | Add/update node/edge attributes |
| [Remove](#remove) | 7 | Delete edges, attributes, orphan nodes |
| [Collapse](#collapse) | 7 | Merge nodes with predicate matching |
| [Traversal](#traversal) | 7 | Follow edges in/out/both with depth control |
| [Predicates](#predicates) | 4 | Type-sensitive equality, existence checks |
| [Named Graphs](#named-graphs) | 6 | Binding, from clause, algebra chains |
| [Pipelines](#pipelines) | 3 | Multi-stage query composition |
| [Edge Cases](#edge-cases) | 5 | Empty graphs, self-loops, orphans |

---

## Subgraph Filtering

Extract nodes/edges where a predicate is true.

| Fixture | Tests |
|---------|-------|
| `subgraph_node_filter` | `subgraph node.attr = value` extracts matching nodes |
| `subgraph_edge_filter` | `subgraph edge.attr = value` extracts matching edges |
| `subgraph_node_inequality` | `subgraph node.attr != value` (NEW: inequality operator) |
| `subgraph_edge_inequality` | `subgraph edge.attr != value` (NEW: inequality operator) |
| `subgraph_exists_attribute` | `subgraph node.attr exists` matches nodes with attribute |
| `subgraph_edge_exists_attribute` | `subgraph edge.attr exists` matches edges with attribute |
| `subgraph_not_exists_attribute` | `subgraph node.attr not exists` matches without attribute |
| `subgraph_traverse` | `traverse out 1` follows outgoing edges one level |
| `subgraph_traverse_in` | `traverse in 2` follows incoming edges two levels |
| `subgraph_traverse_both` | `traverse both 2` follows both directions |
| `subgraph_traverse_depth_3` | `traverse out 3` three-level depth |
| `subgraph_traverse_all_depth` | `traverse out` (no depth limit, acyclic) |

**Key Semantic Notes:**
- Predicates are type-sensitive: `"42" != 42`
- Missing attributes: `!= value` returns false (per spec 7.2)
- `exists` requires attribute presence
- `traverse` maintains visited set to avoid infinite loops on cycles

---

## Graph Algebra

Combine named graphs with set operations.

| Fixture | Tests |
|---------|-------|
| `named_graph_union` | `A + B` combines all nodes/edges |
| `named_graph_intersection` | `A & B` keeps only common nodes |
| `named_graph_difference` | `A - B` removes nodes in B from A |
| `named_graph_symmetric_difference` | `A ^ B` nodes in either but not both |
| `named_graph_attribute_merge_rules` | `A + B` right-wins attribute merging |
| `named_graph_chained_algebra` | Multiple algebra ops in sequence |
| `named_graph_from_named` | `from NAME` switches working graph |

**Key Semantic Notes:**
- Union merges nodes; last-write-wins for attributes
- Intersection requires node IDs to match
- Difference removes matching nodes and incident edges
- Symmetric difference is XOR on node sets
- Edges deduplicated only during **collapse**, not algebra

---

## Make (Assign)

Add or update node/edge attributes conditionally.

| Fixture | Tests |
|---------|-------|
| `make_assign` | `make node.attr = value where predicate` |
| `make_boolean_value` | `make node.flag = true` boolean assignment |
| `make_numeric_value` | `make node.count = 42` numeric assignment |
| `make_edge_attribute` | `make edge.attr = value where predicate` |
| `make_multiple_attributes` | Multiple make operations in pipeline |

**Key Semantic Notes:**
- Attributes are untyped; type stored as provided
- `where` predicate filters which nodes/edges to update
- Operations are cumulative in pipeline

---

## Remove

Delete nodes, edges, or attributes.

| Fixture | Tests |
|---------|-------|
| `remove_edge_filter` | `remove edge where predicate` removes matching edges |
| `remove_edge_attribute` | `remove edge.attr where predicate` clears attribute |
| `remove_attribute` | `remove node.attr where predicate` clears attribute |
| `remove_orphans` | `remove orphans` deletes nodes with no incident edges |
| `remove_orphans_with_self_loop` | Self-loop counts as incident edge (not orphan) |
| `remove_multiple_operations` | Multiple removes in sequence |
| `single_node_remove_orphans` | Single node with no edges is orphan |

**Key Semantic Notes:**
- Remove edge: nodes remain, edges deleted
- Remove attribute: node/edge remains, property cleared
- Self-loop prevents orphan status
- Orphans pass preserved sets during removal

---

## Collapse

Merge multiple nodes into a single target node.

| Fixture | Tests |
|---------|-------|
| `collapse_nodes` | `collapse into ID where predicate` merges nodes |
| `collapse_attribute_merge` | Attributes merged (last-write-wins) during collapse |
| `collapse_edge_redirect` | Edges to collapsed nodes redirect to target |
| `collapse_internal_edges_removed` | Edges between collapsed nodes deleted |
| `collapse_deduplication` | Duplicate edges deduplicated **only** during collapse |
| `collapse_multiple_targets` | Multiple collapse operations in sequence |
| `invalid_collapse_target_not_in_graph` | Error if collapse target not in graph |

**Key Semantic Notes:**
- Edge rewriting: `A → B` becomes `ID → B` if A or B collapsed
- Internal edges (collapsed→collapsed) removed
- Deduplication happens only during collapse, not other operations
- Target node must already exist in graph

---

## Traversal

Follow edges from a start node up to a depth limit.

| Fixture | Tests |
|---------|-------|
| `subgraph_traverse` | `traverse out 1` one level out |
| `subgraph_traverse_in` | `traverse in 2` two levels incoming |
| `subgraph_traverse_both` | `traverse both 2` bidirectional |
| `subgraph_traverse_depth_3` | `traverse out 3` three levels |
| `subgraph_traverse_all_depth` | `traverse out` (unbounded, handles cycles) |
| `cyclic_graph_traversal` | Cycles don't infinite loop (visited set) |
| `wide_fanout` | High fan-out correctly traversed |

**Key Semantic Notes:**
- `traverse` requires matching predicate first
- Visited set prevents cycles from infinite loops
- `out`, `in`, `both` control direction
- Unbounded traversal (`traverse out`) safe on cyclic graphs

---

## Predicates

Type-sensitive comparisons and existence checks.

| Fixture | Tests |
|---------|-------|
| `predicate_string_equality` | `node.attr = "value"` string comparison |
| `predicate_numeric_equality` | `node.attr = 42` numeric comparison |
| `predicate_boolean_equality` | `node.attr = true` boolean comparison |
| `predicate_type_sensitive` | `"42" != 42` different types, not equal |

**Key Semantic Notes:**
- Attributes stored as provided (no type coercion)
- Equality is `==` (double equals) or ` = ` (space-equals-space)
- Inequality `!=` returns false for missing attributes
- No implicit string→int conversion

---

## Named Graphs

Bind intermediate results and reference them in algebra.

| Fixture | Tests |
|---------|-------|
| `named_graph_union` | `(pipeline) as NAME` binds graph |
| `named_graph_intersection` | Reference `NAME` in algebra expression |
| `named_graph_difference` | `A - B` where A, B are named |
| `named_graph_symmetric_difference` | `A ^ B` where A, B are named |
| `named_graph_attribute_merge_rules` | Attribute merge semantics in `A + B` |
| `named_graph_chained_algebra` | Multiple named graphs, chained operations |
| `named_graph_from_named` | `from NAME` changes working graph |

**Key Semantic Notes:**
- Named graph scope is single query (session-local)
- `from *` resets to input graph
- `from NAME` changes working graph to named graph
- Cannot rebind a name (error if attempted twice)

---

## Pipelines

Multi-stage composition of expressions.

| Fixture | Tests |
|---------|-------|
| `pipeline_subgraph_traverse_make` | `subgraph \| traverse \| make` three stages |
| `pipeline_three_stages` | Three different expression types |
| `pipeline_binding_and_algebra` | Bind + algebra + make in sequence |

**Key Semantic Notes:**
- Stages separated by `|` (pipe)
- Each stage receives output of previous
- Working graph persists across stages unless `from` used

---

## Edge Cases

Minimal/boundary graph structures.

| Fixture | Tests |
|---------|-------|
| `single_node_no_edges` | Single node, no edges |
| `empty_graph_subgraph` | Subgraph on empty graph returns empty |
| `self_loop_only` | Node with only self-loop, no other edges |
| `disconnected_components` | Multiple disconnected subgraphs |
| `duplicate_edges_preserved` | Multiple edges between same nodes |

**Key Semantic Notes:**
- Empty graph is valid result
- Self-loop counts as incident edge
- Duplicate edges preserved except during collapse
- Disconnected components behave independently

---

## Maintenance Checklist

**For any agent adding new tests or modifying existing ones:**

- [ ] **Add new fixture?** Update this README with entry in appropriate category
- [ ] **Change predicate syntax?** Update predicate examples and semantic notes
- [ ] **Change algebra semantics?** Update graph algebra section and merge rules
- [ ] **Change collapse behavior?** Update collapse section, especially deduplication rules
- [ ] **Add new operator/keyword?** Add new category section and document semantics
- [ ] **Run all tests?** Verify `go test -v ./query` passes before commit
- [ ] **Document in QUERY_SPEC.md?** Core language changes belong in spec, not just here

**Maintaining context efficiency:**

This README is designed so LLMs can:
1. Quickly understand test coverage without scanning 59 directories
2. Find relevant fixtures by category (15 categories, 1-12 entries each)
3. Reference semantic notes for correct behavior
4. Know when to update this file (checklist above)

Keep entries brief: test name + one-line description + key semantic note.

