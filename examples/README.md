# GSL Examples

This directory contains example GSL graph definitions organized as a learning path, along with a demonstration of using the GSL library in a Go program.

---

## 01-basics — Core Graph Patterns

Simple graph definitions demonstrating fundamental GSL concepts.

### `simple_workflow.gsl`

A basic directed graph representing a sequential process flow.

**Demonstrates:**
- Simple node declarations
- Text attributes using the shorthand syntax
- Sequential edge definitions

**Use case:** Process workflows, state machines, pipelines.

### `microservices.gsl`

A production-like microservices architecture graph.

**Demonstrates:**
- Multiple sets for categorization (services, critical, internal)
- Nodes with multiple attributes (port, replicas, version, etc.)
- Nodes belonging to multiple sets
- Edges with metadata (method, protocol, timeout)
- Real-world naming conventions and metadata

**Use case:** System architecture, service dependencies, deployment configurations.

### `hierarchical_system.gsl`

A graph with parent-child relationships representing a hierarchical structure.

**Demonstrates:**
- Set declarations with attributes
- Node attributes
- Set membership (`@frontend`, `@backend`, `@database`)
- Parent-child relationships via explicit parent attributes
- Edges between different hierarchies

**Use case:** Component hierarchies, organizational structures, system layers.

### `data_pipeline.gsl`

An ETL (Extract-Transform-Load) data pipeline.

**Demonstrates:**
- Sets for pipeline stages (intake, processing, output)
- Nodes representing data processing stages
- Attributes describing processing characteristics
- Sequential data flow with transformations
- Compression and optimization attributes

**Use case:** Data engineering, ETL workflows, analytics pipelines.

---

## 02-algorithms — Graph Algorithm Examples

Graphs designed for algorithmic analysis and traversal.

### `task_scheduling.gsl`

A build system with task dependencies showing the execution flow of a CI/CD pipeline.

**Demonstrates:**
- Task dependency graphs
- Multiple dependency paths
- Nodes with no dependencies (entry points)
- Suitable for topological sorting

**Use case:** Build systems, CI/CD pipelines, task scheduling, dependency resolution.

### `task_dependencies.gsl`

A demonstration of edge labels and scoping for task dependency graphs.

**Demonstrates:**
- Edge labels for dependency targeting (`E1: Setup -> UnitTests`)
- Scoped edges for implicit dependencies
- Explicit `parent` attributes
- Cross-branch dependencies between different pipeline stages

**Use case:** Build systems, workflow orchestration, dependency graphs, task scheduling.

### `circular_dependencies.gsl`

A graph with circular dependencies between services.

**Demonstrates:**
- Cyclic relationships in graphs
- Multiple components in a cycle
- Dependencies both within and outside the cycle

**Use case:** Detecting problematic circular dependencies, validating acyclic graphs.

### `social_network.gsl`

A social network showing connections between people.

**Demonstrates:**
- Multiple paths between nodes
- Graph with branching and merging paths
- Unidirectional connections

**Use case:** Path finding, shortest path algorithms, reachability analysis, social network analysis.

---

## 03-edge-cases — Parsing Edge Cases

Graphs that demonstrate parser warnings and edge-case handling.

### `implicit_sets.gsl`

Demonstrates implicit set creation warnings.

### `name_collision.gsl`

Demonstrates node/set name collision warnings.

### `parent_override.gsl`

Demonstrates parent override inside block warnings.

---

## 04-serialization — Serialization Round-Trip

Graphs that verify canonical serialization produces parseable output.

### `grouped_edges.gsl`

A labeled grouped edge with duplicate nodes that must be serialized as a grouped declaration to preserve round-trip correctness.

**Demonstrates:**
- Labeled grouped edges
- Serialization round-trip with duplicate edges sharing a label
- Avoiding duplicate edge label error in re-parsed output
---

## 05-sequence — Sequence Diagrams

GSL graphs that produce sequence diagrams when converted with `-t sequence`. These demonstrate edge scoping for activation/deactivation, auto-detected return arrows, and attribute-based arrow overrides.

### `user_login.gsl`

A user login flow with nested 2FA verification. Demonstrates the core sequence pattern: parent edge activates target, child edges execute within that context.

**Demonstrates:**
- Nested scoped blocks for multi-level interaction
- Auto-detected return arrows (child edge pointing back to parent's source)
- Node display names via `text` attribute

### `simple_request_response.gsl`

A basic client-server-database interaction with auto-detected request and return arrows.

**Demonstrates:**
- Simplest possible sequence pattern
- Auto-detected dashed return arrow from DB to Server
- Multi-step request/response chain

### `nested_activations.gsl`

Three levels of nesting: client → auth → database, with a deeply scoped session creation.

**Demonstrates:**
- Stacked activation/deactivation (each scope level adds a frame)
- Mixed auto-detected and overridden arrow styles inside scopes

### `arrow_overrides.gsl`

Shows how to force solid or dashed arrows using the `arrow` attribute, overriding auto-detection.

**Demonstrates:**
- `arrow = "dashed"` for async events and logging
- `arrow = "solid"` for forward-direction edges that would otherwise be auto-detected as returns

### `flat_edges_parent.gsl`

Builds a sequence diagram using explicit `parent` attributes on child edges instead of scoped blocks.

**Demonstrates:**
- Labeled parent edges with `E1:` syntax
- Child edges referencing parent via `[parent=E1]`
- Same activation/deactivation behavior as scoped blocks

### `error_handling.gsl`

An online checkout flow with both happy path and fraud-decline error path.

**Demonstrates:**
- Multiple scoped blocks from the same parent edge
- Different outcomes based on child edge flow

### `api_orchestration.gsl`

Fan-out orchestration where a single API Gateway call triggers multiple downstream services.

**Demonstrates:**
- Multiple child edges from different participants within one scope
- Rich multi-service interaction diagrams

---

## Example Tests

`example_test.go` contains runnable documentation examples that demonstrate common patterns:

### Core Patterns

- **Example_parseSimpleWorkflow** - Parsing a basic workflow and accessing nodes
- **Example_parseHierarchicalSystem** - Working with sets and parent-child relationships
- **Example_microservicesArchitecture** - Finding nodes in specific sets
- **Example_queryNodeDependencies** - Querying outbound edges from a node
- **Example_parseDataPipeline** - Grouping nodes by set membership
- **Example_serializeGraph** - Round-tripping: parse → serialize → parse
- **Example_graphStatistics** - Computing graph statistics

### Algorithm Examples

- **Example_topologicalSort** - Demonstrates topological sort on a task dependency graph using Kahn's algorithm (`02-algorithms/task_scheduling.gsl`)
- **Example_cycleDetection** - Demonstrates cycle detection using DFS with color marking (`02-algorithms/circular_dependencies.gsl`)
- **Example_pathFinding** - Demonstrates finding all paths between two nodes using DFS (`02-algorithms/social_network.gsl`)

### Warning Examples

- **Example_implicitSets** - Demonstrates implicit set creation warnings (`03-edge-cases/implicit_sets.gsl`)
- **Example_nameCollision** - Demonstrates node/set name collision warnings (`03-edge-cases/name_collision.gsl`)
- **Example_parentOverride** - Demonstrates parent override inside block warnings (`03-edge-cases/parent_override.gsl`)

These are Go's standard Example pattern — they run as part of the test suite and also appear in godoc documentation. Warnings are non-fatal and don't prevent parsing; they're informational linter messages.

---

## Running the Tests

```bash
go test ./examples -v
```

---

## Quick Reference

| Feature | Example |
|---------|---------|
| Basic nodes & edges | `01-basics/simple_workflow.gsl` |
| Text attributes | All files |
| Set membership | `01-basics/microservices.gsl`, `01-basics/data_pipeline.gsl` |
| Parent-child relationships | `01-basics/hierarchical_system.gsl` |
| Multi-valued attributes | `01-basics/microservices.gsl` |
| Edge attributes | `01-basics/microservices.gsl`, `01-basics/data_pipeline.gsl` |
| Topological sort | `02-algorithms/task_scheduling.gsl` |
| Cycle detection | `02-algorithms/circular_dependencies.gsl` |
| Path finding | `02-algorithms/social_network.gsl` |
| Edge labels & dependencies | `02-algorithms/task_dependencies.gsl` |
| Implicit sets warning | `03-edge-cases/implicit_sets.gsl` |
| Name collision warning | `03-edge-cases/name_collision.gsl` |
| Parent override warning | `03-edge-cases/parent_override.gsl` |
| Labeled grouped edges | `04-serialization/grouped_edges.gsl` |
| Sequence diagrams | `05-sequence/user_login.gsl` |
| Simple request-response | `05-sequence/simple_request_response.gsl` |
| Nested activations | `05-sequence/nested_activations.gsl` |
| Arrow style overrides | `05-sequence/arrow_overrides.gsl` |
| Flat edges with explicit parent | `05-sequence/flat_edges_parent.gsl` |
| Error handling flow | `05-sequence/error_handling.gsl` |
| API orchestration | `05-sequence/api_orchestration.gsl` |

## Graph Sizes

| File | Nodes | Edges | Sets |
|------|-------|-------|------|
| `01-basics/simple_workflow.gsl` | 6 | 5 | 0 |
| `01-basics/microservices.gsl` | 7 | 9 | 3 |
| `01-basics/hierarchical_system.gsl` | 9 | 4 | 3 |
| `01-basics/data_pipeline.gsl` | 8 | 7 | 3 |
| `02-algorithms/task_scheduling.gsl` | 7 | 7 | 0 |
| `02-algorithms/circular_dependencies.gsl` | 4 | 4 | 0 |
| `02-algorithms/social_network.gsl` | 6 | 7 | 0 |
| `03-edge-cases/implicit_sets.gsl` | 4 | 1 | 2 |
| `03-edge-cases/name_collision.gsl` | 3 | 2 | 1 |
| `03-edge-cases/parent_override.gsl` | 4 | 0 | 0 |
| `04-serialization/grouped_edges.gsl` | 1 | 2 | 0 |
| `05-sequence/user_login.gsl` | 5 | 8 | 0 |
| `05-sequence/simple_request_response.gsl` | 3 | 4 | 0 |
| `05-sequence/nested_activations.gsl` | 3 | 5 | 0 |
| `05-sequence/arrow_overrides.gsl` | 4 | 5 | 0 |
| `05-sequence/flat_edges_parent.gsl` | 4 | 5 | 0 |
| `05-sequence/error_handling.gsl` | 6 | 11 | 0 |
| `05-sequence/api_orchestration.gsl` | 7 | 11 | 0 |

## More Information

- See [SPEC.md](../SPEC.md) for the language specification
- See [GRAMMAR.md](../GRAMMAR.md) for the formal grammar
- See [README.md](../README.md) for an overview of GSL
