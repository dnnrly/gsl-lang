# Phase 1.5 Graph Algebra Test Fixtures

Created 5 comprehensive test fixtures for Phase 1.5 Graph Algebra operations.

## Fixtures Created

### 1. named_graph_difference (-)
**Location:** `query/testdata/named_graph_difference/`

- **Operator:** Difference (-)
- **Semantics:** A - B returns nodes in A but not in B
- **Graph:** 
  - Nodes: a, b, c (team="api"), d (team="fraud")
  - Edges: a→b, b→c
- **Query:** `(subgraph node.team == "api") as SET1 | from * | (subgraph node.team == "fraud") as SET2 | SET1 - SET2`
- **Expected Result:** Nodes {a, b, c} with edges a→b, b→c (all api nodes, excluding fraud)

### 2. named_graph_symmetric_difference (^)
**Location:** `query/testdata/named_graph_symmetric_difference/`

- **Operator:** Symmetric Difference (^)
- **Semantics:** A ^ B returns nodes in either A or B but not both (XOR)
- **Graph:**
  - Nodes: api1, api2 (team="api"), crit1, crit2 (team="fraud")
  - Attributes: critical boolean flag
  - Edges: api1→api2, crit1→crit2
- **Query:** `(subgraph node.team == "api") as API | from * | (subgraph node.critical == true) as CRIT | API ^ CRIT`
- **Expected Result:** Nodes {api1 (api but not critical), crit1 (critical but not api)}
  - Note: Edges to excluded nodes are removed (api2, crit2)

### 3. named_graph_chained_algebra (Multiple operations)
**Location:** `query/testdata/named_graph_chained_algebra/`

- **Operators:** Union (+) in sequence
- **Semantics:** Create multiple named graphs and chain algebra operations
- **Graph:**
  - Nodes: a (zone="a"), b (zone="b"), c (zone="c")
  - Edges: a→b, b→c
- **Query:** `(subgraph node.zone == "a") as A | from * | (subgraph node.zone == "b") as B | from * | (subgraph node.zone == "c") as C | A + B`
- **Expected Result:** Nodes {a, b} (union of A and B)

### 4. named_graph_from_named (`from NAME` switching)
**Location:** `query/testdata/named_graph_from_named/`

- **Feature:** Switching working graph with `from NAME`
- **Semantics:** `from NAME` changes the working graph to a named graph, operations apply only to that graph
- **Graph:**
  - Nodes: api (team="api"), fraud (team="fraud"), gateway (team="gateway")
  - All nodes have priority=0
  - Edges: api→fraud→gateway
- **Query:** `(subgraph node.team == "api") as API | from API | make node.priority = 1 where node.team exists`
- **Expected Result:** Only api node with priority updated to 1
  - Note: `from API` switches working graph; make only affects nodes in API

### 5. named_graph_attribute_merge_rules (Right-wins attribute merging)
**Location:** `query/testdata/named_graph_attribute_merge_rules/`

- **Feature:** Attribute merge semantics in union (right overwrites left)
- **Semantics:** In A + B, if a node exists in both, right's attributes overwrite left's
- **Graph:**
  - Node x with team="api", env="prod"
- **Query:** `(subgraph node.team == "api") as LEFT | from * | (subgraph node.env == "prod") as RIGHT | LEFT + RIGHT`
- **Expected Result:** Node x with both team="api" and env="prod" (both graphs match, attributes unchanged)

## Test Execution

All fixtures pass the test suite:

```bash
go test -v -run "TestFixtures/named_graph" ./query
# Output: 7 PASS (5 new + 2 existing union/intersection)
```

## Key Design Decisions

1. **Simple Graph Structures:** Each fixture uses minimal nodes/edges to clearly demonstrate the algebra operation
2. **Distinct Predicates:** Graphs use different attributes (team, zone, critical, env) to allow clear separation via filters
3. **Semantics Testing:** 
   - Difference: Tests node exclusion and edge cleanup
   - Symmetric Difference: Tests XOR semantics 
   - Chaining: Tests pipeline composition
   - From Named: Tests working graph switching
   - Attribute Merge: Tests attribute precedence rules

## Integration Notes

These fixtures integrate with the `fixtures_test.go` harness which:
- Loads graph.gsl and parses it
- Parses query.gql and executes it on the parsed graph
- Serializes the result and compares with expected result.gsl
- Validates canonical form round-tripping

All fixtures follow GSL syntax standards and use valid query language features.
