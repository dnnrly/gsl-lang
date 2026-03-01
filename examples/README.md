# GSL Examples

This directory contains example GSL graph definitions and a demonstration of using the GSL library in a Go program.

## Graph Definition Examples

### 1. Simple Workflow (`simple_workflow.gsl`)

A basic directed graph representing a sequential process flow.

**Demonstrates:**
- Simple node declarations
- Text attributes using the shorthand syntax
- Sequential edge definitions

**Use case:** Process workflows, state machines, pipelines.

### 2. Hierarchical System (`hierarchical_system.gsl`)

A graph with parent-child relationships representing a hierarchical structure.

**Demonstrates:**
- Set declarations with attributes
- Node attributes
- Set membership (`@frontend`, `@backend`, `@database`)
- Parent-child relationships via explicit parent attributes
- Edges between different hierarchies

**Use case:** Component hierarchies, organizational structures, system layers.

### 3. Microservices (`microservices.gsl`)

A production-like microservices architecture graph.

**Demonstrates:**
- Multiple sets for categorization (services, critical, internal)
- Nodes with multiple attributes (port, replicas, version, etc.)
- Nodes belonging to multiple sets
- Edges with metadata (method, protocol, timeout)
- Real-world naming conventions and metadata

**Use case:** System architecture, service dependencies, deployment configurations.

### 4. Data Pipeline (`data_pipeline.gsl`)

An ETL (Extract-Transform-Load) data pipeline.

**Demonstrates:**
- Sets for pipeline stages (intake, processing, output)
- Nodes representing data processing stages
- Attributes describing processing characteristics
- Sequential data flow with transformations
- Compression and optimization attributes

**Use case:** Data engineering, ETL workflows, analytics pipelines.

### 5. Task Scheduling (`task_scheduling.gsl`)

A build system with task dependencies showing the execution flow of a CI/CD pipeline.

**Demonstrates:**
- Task dependency graphs
- Multiple dependency paths
- Nodes with no dependencies (entry points)
- Suitable for topological sorting

**Use case:** Build systems, CI/CD pipelines, task scheduling, dependency resolution.

### 6. Circular Dependencies (`circular_dependencies.gsl`)

A graph with circular dependencies between services.

**Demonstrates:**
- Cyclic relationships in graphs
- Multiple components in a cycle
- Dependencies both within and outside the cycle

**Use case:** Detecting problematic circular dependencies, validating acyclic graphs.

### 7. Social Network (`social_network.gsl`)

A social network showing connections between people.

**Demonstrates:**
- Multiple paths between nodes
- Graph with branching and merging paths
- Unidirectional connections

**Use case:** Path finding, shortest path algorithms, reachability analysis, social network analysis.

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

- **Example_topologicalSort** - Demonstrates topological sort on a task dependency graph using Kahn's algorithm (`task_scheduling.gsl`)
- **Example_cycleDetection** - Demonstrates cycle detection using DFS with color marking (`circular_dependencies.gsl`)
- **Example_pathFinding** - Demonstrates finding all paths between two nodes using DFS (`social_network.gsl`)

### Warning Examples

- **Example_implicitSets** - Demonstrates implicit set creation warnings (`implicit_sets.gsl`)
- **Example_nameCollision** - Demonstrates node/set name collision warnings (`name_collision.gsl`)
- **Example_parentOverride** - Demonstrates parent override inside block warnings (`parent_override.gsl`)

These are Go's standard Example pattern - they run as part of the test suite and also appear in godoc documentation. Warnings are non-fatal and don't prevent parsing; they're informational linter messages.

## Running the Tests

To run all example tests:

```bash
go test ./examples -v
```

To run the full test suite including examples:

```bash
go test ./...
```

## Key Features Demonstrated

| Feature | Example File |
|---------|--------------|
| Basic nodes & edges | `simple_workflow.gsl` |
| Text attributes | All files |
| Set membership | `microservices.gsl`, `data_pipeline.gsl` |
| Parent-child relationships | `hierarchical_system.gsl` |
| Multi-valued attributes | `microservices.gsl` |
| Edge attributes | `microservices.gsl`, `data_pipeline.gsl` |
| Canonical serialization | `main.go` |
| Graph traversal | `main.go` |

## Graph Sizes and Examples

| File | Nodes | Edges | Sets | Algorithm Use |
|------|-------|-------|------|---------------|
| `simple_workflow.gsl` | 6 | 5 | 0 | Basic graph structure |
| `hierarchical_system.gsl` | 9 | 4 | 3 | Set membership queries |
| `microservices.gsl` | 7 | 9 | 3 | Set membership queries |
| `data_pipeline.gsl` | 8 | 7 | 3 | Edge traversal |
| `task_scheduling.gsl` | 7 | 7 | 0 | Topological sort |
| `circular_dependencies.gsl` | 4 | 4 | 0 | Cycle detection |
| `social_network.gsl` | 6 | 7 | 0 | Path finding |
| `implicit_sets.gsl` | 4 | 1 | 2 | Warning handling |
| `name_collision.gsl` | 3 | 2 | 1 | Warning handling |
| `parent_override.gsl` | 4 | 0 | 0 | Warning handling |

## More Information

- See [SPEC.md](../SPEC.md) for the language specification
- See [GRAMMAR.md](../GRAMMAR.md) for the formal grammar
- See [README.md](../README.md) for an overview of GSL
