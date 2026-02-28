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

| File | Nodes | Edges | Sets | Demonstrates |
|------|-------|-------|------|--------------|
| `simple_workflow.gsl` | 6 | 5 | 0 | Basic parsing and node access |
| `hierarchical_system.gsl` | 9 | 4 | 3 | Parent-child relationships |
| `microservices.gsl` | 7 | 9 | 3 | Set membership, edge attributes |
| `data_pipeline.gsl` | 8 | 7 | 3 | Node grouping by stage |
| `implicit_sets.gsl` | 4 | 1 | 2 | Implicit set creation warnings |
| `name_collision.gsl` | 3 | 2 | 1 | Name collision warnings |
| `parent_override.gsl` | 4 | 0 | 0 | Parent override warnings |

## More Information

- See [SPEC.md](../SPEC.md) for the language specification
- See [GRAMMAR.md](../GRAMMAR.md) for the formal grammar
- See [README.md](../README.md) for an overview of GSL
