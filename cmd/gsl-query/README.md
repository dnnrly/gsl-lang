# gsl-query

Execute queries against GSL (Graph Specification Language) graphs and output filtered/transformed results.

## Installation

```bash
make build
```

The binary will be available at `./cmd/gsl-query/gsl-query`.

## Usage

### Basic Syntax

```bash
gsl-query [flags] <query>
```

### Query from Argument

```bash
gsl-query 'subgraph node.team == "payments"' < input.gsl
```

### Query from File

```bash
gsl-query -f query.txt < input.gsl
gsl-query -f queries/find_critical.txt -i graph.gsl -o result.gsl
```

### Flags

- `-i, --input <file>` — Read GSL graph from file (default: stdin)
- `-o, --output <file>` — Write result to file (default: stdout)
- `-f, --query-file <file>` — Read query from file
- `-h, --help` — Show help

## Examples

### Filter by Attribute

```bash
# Find all payment services
gsl-query 'subgraph node.team == "payments"' < services.gsl
```

### Traverse Dependencies

```bash
# Find all services that critical services depend on
gsl-query 'subgraph node in @critical traverse out all' < services.gsl
```

### Remove Edges

```bash
# Remove all deprecated connections
gsl-query 'remove edge where edge.status == "deprecated"' < services.gsl
```

### Complex Pipeline

```bash
# Find critical nodes, then remove deprecated edges, then clean orphans
gsl-query '(subgraph node in @critical) as CRITICAL | from * | remove edge where edge.status == "deprecated" | remove orphans' < services.gsl
```

### Assign Attributes

```bash
# Mark payment services as reviewed
gsl-query 'make node.reviewed = "yes" where node.team == "payments"' < services.gsl
```

## See Also

- [QUERY_AI_GUIDE.md](../../QUERY_AI_GUIDE.md) — Complete query language reference
- [LLM_GUIDE.md](../../LLM_GUIDE.md) — GSL syntax and semantics
