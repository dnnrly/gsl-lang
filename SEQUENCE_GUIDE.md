---
name: sequence-diagram-guide
type: dialect-reference
language: GSL
version: 1.0.0
description: Reference guide for the PlantUML sequence diagram dialect in gsl-diagram. Covers how GSL graphs map to UML sequence diagrams, including participant declarations, scoped activations, arrow styles, labeled scopes, and nesting patterns. Use when converting GSL to sequence diagrams or understanding how GSL edges become PlantUML messages.
keywords: [gsl, sequence-diagram, plantuml, uml, activations, arrow-styles, lifelines, gsl-diagram]
---

# GSL to PlantUML Sequence Diagrams - LLM Guide

This document explains how `gsl-diagram` converts GSL graphs into PlantUML sequence diagrams using the `--format plantuml --type sequence` option.

## Overview

A GSL graph describes **components** (nodes) and **interactions** (edges). The sequence dialect renders these as a UML sequence diagram where:

- **Nodes** become **participants** (lifelines)
- **Edges** become **messages** (arrows between participants)
- **Scoped blocks** become **activations** (lifeline highlights with implicit return)
- **Edge text** becomes **message labels**
- **Edge attributes** control **arrow style** and other visual properties

```bash
gsl-diagram -i interactions.gsl -f plantuml -t sequence > diagram.puml
```

## Participant Declarations

Nodes are declared with `node` and become PlantUML participants:

```gsl
node Client
node Server
node Database
```

Renders as:

```plantuml
participant Client
participant Server
participant Database
```

### Participant Types

Use the `shape` attribute to change the participant shape. Supported PlantUML types:

```gsl
node User [shape=actor]
node API [shape=boundary]
node Core [shape=control]
node Entity [shape=entity]
node DB [shape=database]
node Cache [shape=collections]
node Queue [shape=queue]
```

Renders as:

```plantuml
actor User
boundary API
control Core
entity Entity
database DB
collections Cache
queue Queue
```

### Participant Labels

Use the `text` attribute to set a display label:

```gsl
node A: "REST API Gateway"
node B [text="Authentication Service"]
```

Renders as:

```plantuml
participant A as "REST API Gateway"
participant B as "Authentication Service"
```

Participants are discovered from edges. If a node is referenced in an edge but not declared, it still appears as a participant:

```gsl
Client->Server: "Hello"
```

Renders as:

```plantuml
participant Client
participant Server

Client -> Server: Hello
```

## Messages (Edges)

Edges become messages between participants. The arrow style defaults to solid (`->`) for synchronous calls.

```gsl
Client->Server: "Request"
Server->Database: "Query"
```

Renders as:

```plantuml
Client -> Server: Request
Server -> Database: Query
```

### Self-Messages

An edge where `from == to` becomes a self-message:

```gsl
Server->Server: "Internal processing"
```

## Arrow Styles

The `arrow` attribute controls the message style using UML-meaningful names:

| `arrow` value | UML meaning | PlantUML | Visual |
|---|---|---|---|
| `sync` (default) | Synchronous call | `->` | Solid line, filled arrowhead |
| `async` | Asynchronous message | `->>` | Solid line, open arrowhead |
| `return` | Return/reply | `-->` | Dashed line, open arrowhead |
| `dependency` | Weak dependency | `..>` | Dotted line, open arrowhead |
| `strong` | Strong coupling | `==>` | Double line, filled arrowhead |

```gsl
Client->Server: "Login"
Server->Client [arrow="return"]: "Token"
Client->Server [arrow="async"]: "Fire event"
Client->Cache [arrow="dependency"]: "Check cache"
Server->Database [arrow="strong"]: "Write record"
```

Renders as:

```plantuml
Client -> Server: Login
Server --> Client: Token
Client ->> Server: Fire event
Client ..> Cache: Check cache
Server ==> Database: Write record
```

### When to Use Each Style

- **sync**: Blocking request-response. The sender waits for a reply. Use for function calls, RPC, HTTP requests.
- **async**: Fire-and-forget message. The sender does not wait. Use for events, signals, message queues.
- **return**: Reply from a synchronous call. Typically paired with a sync call earlier.
- **dependency**: Weak "uses" relationship. The interaction is optional or indirect.
- **strong**: Tight coupling. The components are tightly bound (composition, ownership).

## Scoped Blocks (Activations)

A scoped block on an edge creates a PlantUML activation — a highlighted lifeline region with an implicit `return` at the end:

```gsl
Client->Server: "Login" {
    Server->Database: "Lookup"
    Server->Client: "Token"
}
```

Renders as:

```plantuml
Client -> Server ++: Login
    Server -> Database: Lookup
    Server -> Client: Token
return
```

The `++` activation marker and `return` are added automatically. The opening edge becomes the activation trigger; edges inside the block are indented.

### Multiple Children

A scoped block can contain multiple edges:

```gsl
A->B: "Process" {
    B->C: "Step 1"
    B->D: "Step 2"
    C->D: "Handoff"
}
```

Renders as:

```plantuml
A -> B ++: Process
    B -> C: Step 1
    B -> D: Step 2
    C -> D: Handoff
return
```

### Self-Reference in Scope

A scoped block where the edge is a self-message:

```gsl
A->A: "Initialize" {
    A->B: "Notify"
}
```

Renders as:

```plantuml
A -> A ++: Initialize
    A -> B: Notify
return
```

### Nested Scopes

Scoped blocks can be nested. Each level gets its own activation and return:

```gsl
A->B: "Outer" {
    B->C: "Inner" {
        C->D: "Deep"
    }
}
```

Renders as:

```plantuml
A -> B ++: Outer
    B -> C ++: Inner
        C -> D: Deep
    return
return
```

Inner scopes are indented by one level (4 spaces).

### Standalone Activations

The `activate` attribute creates a standalone activation without a scoped block:

```gsl
A->B [activate=true]
```

Renders as:

```plantuml
A -> B ++
return
```

Use this when you need an activation without nested children.

## Labeled Scopes

A label on an edge creates a named scope. Child edges reference the label with `[parent=label]`:

```gsl
my_scope: A->B
B->C
A->D [parent=my_scope]
```

Renders as:

```plantuml
A -> B ++: my_scope
    B -> C
    A -> D
return
```

### How Labeled Scopes Work

1. The labeled edge opens the scope and becomes an activation
2. All subsequent edges (until a new labeled edge, an edge with text, or end of input) are treated as children
3. Edges with `[parent=label]` explicitly reference the scope as their parent
4. The scope closes at the next labeled edge or end of input

### Mixing Scope Patterns

Labeled scopes can mix different activation patterns inside:

```gsl
workflow: A->B
B->C
D->E {
    E->F
}
C->G [parent=workflow]
```

Renders as:

```plantuml
A -> B ++: workflow
    B -> C
    D -> E ++
        E -> F
    return
    C -> G
return
```

## Indentation Rules

- Top-level messages: no indent
- First-level scope children: 4 spaces
- Nested scope children: +4 spaces per level

```gsl
A->B: "L1" {
    B->C: "L2" {
        C->D: "L3"
    }
}
```

Renders as:

```plantuml
A -> B ++: L1
    B -> C ++: L2
        C -> D: L3
    return
return
```

## Edge Ordering

Edges are rendered in declaration order. The participant order is determined by:

1. First appearance in edge declarations (from → to → next edge)
2. Alphabetically for nodes not referenced in any edge

```gsl
node Z
node A
A->B
B->C
```

Renders as:

```plantuml
participant A
participant B
participant C
participant Z

A -> B
B -> C
```

`Z` appears last because it is not referenced in any edge.

## Complete Example

```gsl
node Client
node Gateway [shape=boundary]
node Auth [shape=control]
node DB [shape=database]

Client->Gateway: "Login request"
Gateway->Auth [arrow="async"]: "Authenticate"
Auth->DB: "Lookup user"
DB->Auth [arrow="return"]: "User record"
Auth->Gateway [arrow="return"]: "Token"
Gateway->Client [arrow="return"]: "Session"

Client->Gateway: "Fetch data" {
    Gateway->Auth: "Validate token"
    Auth->Gateway [arrow="return"]: "OK"
    Gateway->DB: "Query data"
    DB->Gateway [arrow="return"]: "Result set"
    Gateway->Client [arrow="return"]: "Response"
}
```

Renders as:

```plantuml
participant Client
boundary Gateway
control Auth
database DB

Client -> Gateway: Login request
Gateway ->> Auth: Authenticate
Auth -> DB: Lookup user
DB --> Auth: User record
Auth --> Gateway: Token
Gateway --> Client: Session
Client -> Gateway ++: Fetch data
    Gateway -> Auth: Validate token
    Auth --> Gateway: OK
    Gateway -> DB: Query data
    DB --> Gateway: Result set
    Gateway --> Client: Response
return
```

## Usage

```bash
gsl-diagram -i interactions.gsl -f plantuml -t sequence
gsl-diagram -i interactions.gsl -f plantuml -t sequence -o diagram.puml
cat interactions.gsl | gsl-diagram -f plantuml -t sequence
```

## Limitations

- **No notes**: PlantUML notes are not supported in the sequence dialect
- **No groups/fragments**: `alt`, `opt`, `loop`, `par` fragments are not mapped from GSL
- **No creation/destruction**: `create` and `destroy` lifeline operations are not supported
- **No message numbering**: `autonumber` is not supported
- **No direction control**: Arrow direction (up/down/left/right) is not supported
- **Sequential only**: Messages are rendered in declaration order; no way to reorder visually
