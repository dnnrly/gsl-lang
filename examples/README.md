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
- Explicit `depends_on` attributes
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

## More Information

- See [SPEC.md](../SPEC.md) for the language specification
- See [GRAMMAR.md](../GRAMMAR.md) for the formal grammar
- See [README.md](../README.md) for an overview of GSL
