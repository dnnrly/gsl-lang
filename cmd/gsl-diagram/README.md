# gsl-diagram

Convert GSL (Graph Specification Language) documents to various diagram formats.

## Installation

```bash
go build -o gsl-diagram ./cmd/gsl-diagram
```

## Usage

### Basic usage
```bash
gsl-diagram -i graph.gsl -o diagram.mmd
```

### Specify output format
```bash
gsl-diagram -i graph.gsl -o diagram.mmd --format mermaid
gsl-diagram -i graph.gsl -o diagram.puml --format plantuml
```

### Specify diagram type
```bash
gsl-diagram -i graph.gsl -t graph --format mermaid
```

### From stdin/stdout
```bash
cat graph.gsl | gsl-diagram --format mermaid > diagram.mmd
```

## Flags

- `-i, --input` - Input GSL file (reads from stdin if not provided)
- `-o, --output` - Output diagram file (writes to stdout if not provided)
- `-f, --format` - Output format: `mermaid` (default), `plantuml`
- `-t, --type` - Diagram type: `component` (default), `graph`, `sequence`

## Supported Formats

### Mermaid
- **component**: Hierarchical architecture diagrams with subgraphs
- **graph**: Flowcharts and general directed graphs
- **sequence**: Sequence diagrams with activation blocks from edge scopes

```bash
gsl-diagram -i arch.gsl -f mermaid -t component
gsl-diagram -i workflow.gsl -f mermaid -t graph
gsl-diagram -i sequence.gsl -f mermaid -t sequence
```

### PlantUML
- **component**: Component diagrams with packages
- **sequence**: Sequence diagrams with activation blocks from edge scopes

```bash
gsl-diagram -i arch.gsl -f plantuml
gsl-diagram -i sequence.gsl -f plantuml -t sequence
```

## Examples

### Architecture Diagram (Mermaid)
```bash
gsl-diagram -i microservices.gsl -f mermaid -t component
```

### Workflow Diagram (Mermaid)
```bash
gsl-diagram -i process.gsl -f mermaid -t graph
```

### Component Diagram (PlantUML)
```bash
gsl-diagram -i system.gsl -f plantuml
```

## GSL to Diagram Mapping

### Nodes → Components/Nodes
- Nodes become diagram components or nodes
- `text` attribute is used for display labels
- Falls back to node ID if no text attribute

### Edges → Relationships
- Edges become diagram relationships
- `label`, `name`, or `method` attributes become relationship labels
- Checked in order of precedence: `label` → `name` → `method`

### Parent Relationships
- Nodes with `parent` attribute group into:
  - Mermaid: subgraphs
  - PlantUML: packages

### Sequence Diagrams
- Edge scopes (`A -> B { ... }`) become activation blocks
- The parent edge's target participant is activated while children execute
- Arrow style auto-detects return messages: dashed when `edge.To == parentEdge.From`
- Override arrow style with the `arrow` attribute: `arrow = "solid"` or `arrow = "dashed"`
- Each unique node referenced in an edge becomes a participant
- The `text` attribute on nodes sets the participant display name
- Edge labels use `label`, `name`, or `method` attributes (in priority order)

```gsl
# Example: login flow with nested interactions
Client -> API [method = "POST /login"] {
    API -> Auth [method = "validate credentials"] {
        Auth -> DB [method = "query user"]
        DB -> Auth [method = "user data"]
    }
    API -> Client [method = "200 OK"]
}
```
