---
description: Query language fixture test catalog. Organized as a learning journey — each group builds on the previous. Quick reference for LLMs and developers.
---

# Query Language Test Fixtures

This directory contains **71 fixture-based integration tests** for the GSL Query Language, organized as a **learning path** from simple to complex concepts.

## Structure

Fixtures are nested under numbered group directories:

```
testdata/
  01-basics/              → simple subgraph extraction
  02-predicates/          → filtering conditions (exists, in, !=, AND)
  03-make/                → attribute assignment
  04-remove/              → deletion operations
  05-traversal/           → graph structure traversal (out, in, both)
   06-edge-dependencies/   → edge hierarchy (parent, depth, parent chain)
  07-collapse/            → node merging
  08-named-graphs/        → saving and reusing pipeline results
  09-pipelines/           → multi-stage query composition
  10-edge-cases/          → boundary conditions
```

Each fixture directory contains:
- `graph.gsl` — input graph
- `query.gql` — query to execute
- `result.gsl` — expected output

Tests are loaded recursively by `fixtures_test.go`.

---

## 01-basics — Getting Started

Foundational subgraph operations.

| Fixture | Tests |
|---------|-------|
| `01_single_node_no_edges` | Minimal graph: single node, no edges |
| `02_empty_graph_subgraph` | Subgraph on empty graph returns empty |
| `03_subgraph_node_filter` | `subgraph node.attr = value` extracts matching nodes |
| `04_subgraph_edge_filter` | `subgraph edge.attr = value` extracts matching edges |
| `05_example_basic` | `subgraph node.exists = true` |

---

## 02-predicates — Filtering Conditions

Attribute existence, inequality, type-sensitive equality, and set membership.

| Fixture | Tests |
|---------|-------|
| `01_subgraph_exists_attribute` | `subgraph node.attr exists` matches nodes with attribute |
| `02_subgraph_edge_exists_attribute` | `subgraph edge.attr exists` matches edges with attribute |
| `03_subgraph_not_exists_attribute` | `subgraph node.attr not exists` matches without attribute |
| `04_subgraph_node_inequality` | `subgraph node.attr != value` |
| `05_subgraph_edge_inequality` | `subgraph edge.attr != value` |
| `06_predicate_string_equality` | `node.attr = "value"` string comparison |
| `07_predicate_numeric_equality` | `node.attr = 42` numeric comparison |
| `08_predicate_boolean_equality` | `node.attr = true` boolean comparison |
| `09_predicate_type_sensitive` | `"42" != 42` different types, not equal |
| `10_node_in_set` | `node in @set` set membership |
| `11_edge_in_set` | `edge in @set` set membership |
| `12_node_not_in_set` | `node not in @set` set non-membership |
| `13_edge_not_in_set` | `edge not in @set` set non-membership |
| `14_node_in_missing_set` | `in @missing_set` is false; `not in @missing_set` is true |

**Key Semantic Notes:**
- Predicates are type-sensitive: `"42" != 42`
- Missing attributes: `!= value` returns false (per spec 7.2)
- `exists` requires attribute presence
- `in @missing_set` → false, `not in @missing_set` → true

---

## 03-make — Attribute Assignment

Add or update node/edge attributes conditionally.

| Fixture | Tests |
|---------|-------|
| `01_make_assign` | `make node.attr = value where predicate` |
| `02_make_boolean_value` | `make node.flag = true` boolean assignment |
| `03_make_numeric_value` | `make node.count = 42` numeric assignment |
| `04_make_edge_attribute` | `make edge.attr = value where predicate` |
| `05_make_multiple_attributes` | Multiple make operations in pipeline |

**Key Semantic Notes:**
- Attributes are untyped; type stored as provided
- `where` predicate filters which nodes/edges to update
- Operations are cumulative in pipeline

---

## 04-remove — Deletion Operations

Delete edges, attributes, or orphan nodes.

| Fixture | Tests |
|---------|-------|
| `01_remove_edge_filter` | `remove edge where predicate` removes matching edges |
| `02_remove_orphans` | `remove orphans` deletes nodes with no incident edges |
| `03_remove_orphans_with_self_loop` | Self-loop counts as incident edge (not orphan) |
| `04_remove_attribute` | `remove node.attr where predicate` clears attribute |
| `05_remove_edge_attribute` | `remove edge.attr where predicate` clears attribute |
| `06_remove_multiple_operations` | Multiple removes in sequence |
| `07_single_node_remove_orphans` | Single node with no edges is orphan |

**Key Semantic Notes:**
- Remove edge: nodes remain, edges deleted
- Remove attribute: node/edge remains, property cleared
- Self-loop prevents orphan status

---

## 05-traversal — Graph Structure Traversal

Follow edges from a start node up to a depth limit (`out`, `in`, `both`).

| Fixture | Tests |
|---------|-------|
| `01_subgraph_traverse` | `traverse out 1` one level out |
| `02_subgraph_traverse_in` | `traverse in 2` two levels incoming |
| `03_subgraph_traverse_both` | `traverse both 2` bidirectional |
| `04_subgraph_traverse_depth_3` | `traverse out 3` three levels |
| `05_subgraph_traverse_all_depth` | `traverse out` (unbounded, handles cycles) |
| `06_cyclic_graph_traversal` | Cycles don't infinite loop (visited set) |
| `07_wide_fanout` | High fan-out correctly traversed |

**Key Semantic Notes:**
- `traverse` requires matching predicate first
- Visited set prevents cycles from infinite loops
- `out`, `in`, `both` control graph-structure direction
- Unbounded traversal (`traverse out`) safe on cyclic graphs

---

## 06-edge-dependencies — Edge Hierarchy

Query and traverse edges based on their position in the dependency tree (`parent`).

### Predicates

| Fixture | Tests |
|---------|-------|
| `01_edge_parent_exists` | `subgraph edge parent exists` selects edges with a parent |
| `02_edge_parent_not_exists` | `subgraph edge parent not exists` selects root edges |
| `03_edge_depth` | `subgraph edge.depth == 0` matches edges by dependency depth |

### Dependency Traversal

| Fixture | Tests |
|---------|-------|
| `04_traverse_up` | `traverse up 1` follows Parent chain upward |
| `05_traverse_down` | `traverse down 1` follows Children chain downward |
| `06_traverse_out_up` | `traverse out up 1` combines graph and dependency directions |
| `07_subgraph_scope` | `scope` sugar for `traverse down all` on edge predicates |

### Negative Tests (Boundary Conditions)

| Fixture | Tests |
|---------|-------|
| `08_edge_parent_exists_no_parents` | `edge parent exists` on graph with only root edges → empty |
| `09_edge_depth_no_edges` | `edge.depth == 0` on graph with no edges → empty |
| `10_scope_no_matching_edges` | `scope` on edge predicate with no matches → empty |
| `11_traverse_up_from_root` | `traverse up 1` from root edge → no-op (same result) |
| `12_traverse_down_no_children` | `traverse down 1` from leaf node → no-op (same result) |

**Key Semantic Notes:**
- `parent exists` = edge has `parent` set
- `parent not exists` = edge is a root (no `parent`)
- `depth` is computed by walking the `parent` chain
- `depth == 0` are root edges with no parent
- `scope` ≡ `traverse down all`
- Directions can be combined: `traverse out up 2`
- `edge depends on <predicate>` is tested via unit tests only (`TestDependsOnPredicate`)

---

## 07-collapse — Node Merging

Merge multiple nodes into a single target node.

| Fixture | Tests |
|---------|-------|
| `01_collapse_nodes` | `collapse into ID where predicate` merges nodes |
| `02_collapse_attribute_merge` | Attributes merged (last-write-wins) during collapse |
| `03_collapse_edge_redirect` | Edges to collapsed nodes redirect to target |
| `04_collapse_internal_edges_removed` | Edges between collapsed nodes deleted |
| `05_collapse_deduplication` | Duplicate edges deduplicated **only** during collapse |
| `06_collapse_multiple_targets` | Multiple collapse operations in sequence |
| `07_invalid_collapse_target_not_in_graph` | Error if collapse target not in graph |

**Key Semantic Notes:**
- Edge rewriting: `A → B` becomes `ID → B` if A or B collapsed
- Internal edges (collapsed→collapsed) removed
- Deduplication happens only during collapse, not other operations
- Target node must already exist in graph

---

## 08-named-graphs — Saving and Reusing Results

Bind intermediate pipeline results and combine them with graph algebra.

| Fixture | Tests |
|---------|-------|
| `01_named_graph_union` | `A + B` union combines all nodes/edges |
| `02_named_graph_intersection` | `A & B` intersection keeps only common nodes |
| `03_named_graph_difference` | `A - B` difference removes nodes in B from A |
| `04_named_graph_symmetric_difference` | `A ^ B` symmetric difference (XOR on node sets) |
| `05_named_graph_attribute_merge_rules` | `A + B` right-wins attribute merging |
| `06_named_graph_chained_algebra` | Multiple algebra ops in sequence |
| `07_named_graph_from_named` | `from NAME` switches working graph |

**Key Semantic Notes:**
- Union merges nodes; last-write-wins for attributes
- Intersection requires node IDs to match
- Difference removes matching nodes and incident edges
- Symmetric difference is XOR on node sets
- Named graph scope is single query (session-local)
- `from *` resets to input graph
- Cannot rebind a name (error if attempted twice)

---

## 09-pipelines — Multi-Stage Composition

Chain multiple expressions with the pipe operator.

| Fixture | Tests |
|---------|-------|
| `01_from_clause` | `from *` resets working graph to input |
| `02_pipeline_subgraph_traverse_make` | `subgraph \| traverse \| make` three stages |
| `03_pipeline_three_stages` | Three different expression types |
| `04_pipeline_binding_and_algebra` | Bind + algebra + make in sequence |

**Key Semantic Notes:**
- Stages separated by `|` (pipe)
- Each stage receives output of previous
- Working graph persists across stages unless `from` used

---

## 10-edge-cases — Boundary Conditions

Minimal and unusual graph structures.

| Fixture | Tests |
|---------|-------|
| `01_self_loop_only` | Node with only self-loop, no other edges |
| `02_disconnected_components` | Multiple disconnected subgraphs |
| `03_duplicate_edges_preserved` | Multiple edges between same nodes |

**Key Semantic Notes:**
- Empty graph is valid result
- Self-loop counts as incident edge
- Duplicate edges preserved except during collapse
- Disconnected components behave independently

---

## Maintenance Checklist

**For any agent adding new tests or modifying existing ones:**

- [ ] **Add new fixture?** Place it in the appropriate numbered group directory
- [ ] **New concept that doesn't fit?** Add a new numbered group and update this README
- [ ] **Change predicate syntax?** Update predicate examples and semantic notes
- [ ] **Change algebra semantics?** Update graph algebra section and merge rules
- [ ] **Change collapse behavior?** Update collapse section, especially deduplication rules
- [ ] **Add new operator/keyword?** Add new group section and document semantics
- [ ] **Run all tests?** Verify `go test -v ./query` passes before commit
- [ ] **Document in QUERY_SPEC.md?** Core language changes belong in spec, not just here

**Maintaining context efficiency:**

This README is designed so LLMs can:
1. Follow the learning path from basics → edge cases
2. Find relevant fixtures by group (10 groups, 3-14 entries each)
3. Reference semantic notes for correct behavior
4. Know when to update this file (checklist above)

Keep entries brief: test name + one-line description + key semantic note.
