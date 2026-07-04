# Learning the GSL Query Language

A step-by-step guide to understanding GSL-QL — the pipeline-oriented query and transformation language for GSL graphs.

---

## Prerequisites

You should be familiar with GSL graph definitions: nodes, edges, attributes, and sets. If not, read `README.md` and `SPEC.md` first.

Throughout this guide we'll use the following example graph:

```gsl
node api [team="gateway", zone="A"] @critical
node payments [team="payments", zone="B"] @critical
node fraud [team="payments", zone="B"]
node users [team="identity", zone="A"]
node db [team="platform", zone="C"] @critical
node cache [team="platform", zone="A"] @deprecated

api -> payments [protocol="http"]
api -> users [protocol="http"]
payments -> fraud [protocol="grpc"]
payments -> db [protocol="tcp"]
users -> db [protocol="tcp"]
users -> cache [protocol="tcp"]
fraud -> db [protocol="tcp"]
```

---

## Step 1: The Mental Model — Everything Is a Pipeline

The core idea: a query is a **sequence of expressions** where each expression takes a graph and produces a new graph.

```
input graph → expr₁ → expr₂ → … → result graph
```

Expressions are separated by `|` (the pipe operator), and the pipeline flows left to right — just like a Unix shell pipeline, but for graphs instead of text.

```
subgraph node.team == "payments" | remove orphans
```

**Key insight:** Every expression receives a graph, does something to it, and outputs a graph. The output type never changes — it's always a valid GSL graph.

### Questions to consider

- *Does the pipeline metaphor feel natural for graph operations?*
- *Is left-to-right evaluation intuitive enough, or would you expect something different?*

---

## Step 2: Starting a Query — Choosing Your Input

By default, a query operates on the **input graph** (whatever graph you supply). You can also switch to a different working graph mid-pipeline using `from`:

```
from *
```

`from *` resets to the original input graph. `from NAME` switches to a named graph (more on this in Step 8).

If you don't specify a source, the input graph is used implicitly.

### Questions to consider

- *Is the implicit source clear enough, or would you prefer an explicit `from` always?*

---

## Step 3: Selecting Nodes — Your First Subgraph

The `subgraph` expression filters the working graph using a predicate. Predicates use **dot-path** syntax to indicate whether they target nodes or edges.

The simplest node filter:

```
subgraph node.team == "payments"
```

This returns only nodes where `team` equals `"payments"` — in our example, that's `payments` and `fraud`. Edges between matched nodes are included (`payments -> fraud`). Edges to nodes *outside* the result are excluded.

### How node subgraphs work

When you match nodes:
1. All nodes satisfying the predicate are selected
2. Only edges where **both** source and target are selected are kept

This prevents the subgraph from growing beyond what you asked for.

### Questions to consider

- *Is it clear that matching nodes also filters edges?*
- *Does the "both endpoints must match" rule feel right, or would you sometimes want edges where only one end matches?*

---

## Step 4: Selecting Edges

You can target edges instead of nodes by using `edge.` in the predicate:

```
subgraph edge.protocol == "grpc"
```

This selects all edges where `protocol` is `"grpc"` (here: `payments -> fraud`) and automatically includes their source and target nodes.

### How edge subgraphs work

When you match edges:
1. All edges satisfying the predicate are selected
2. The source and target nodes of those edges are included
3. No other edges are included

### Node match vs. edge match — the key difference

| Predicate type | What's selected first | What's included automatically |
|---|---|---|
| `node.*` | Nodes | Edges between matched nodes |
| `edge.*` | Edges | Source and target nodes of matched edges |

**Important:** You cannot mix `node.` and `edge.` in a single predicate. This is an error:

```
node.team == "payments" AND edge.protocol == "http"
```

### Edge Dependency Predicates

Edges can have parent-child relationships through the `parent` attribute:

```gsl
E1: api -> users [protocol="http"]
E2: payments -> fraud [protocol="grpc", parent=E1]
E3: payments -> db [protocol="tcp", parent=E1]
```

Here `E2` and `E3` both depend on `E1`.

**Parent existence:**

```
subgraph edge parent exists
```

Selects edges with a parent — `E2` and `E3`.

```
subgraph edge parent not exists
```

Selects edges with no parent — `E1`.

**Edge depth:**

```
subgraph edge.depth == 0
```

Selects root edges (`E1`). Depth is `0` for edges with no parent, and `parent.depth + 1` otherwise. So `E2` and `E3` have depth 1.

**Parent dependency predicate:**

```
subgraph edge depends on edge.protocol == "http"
```

Selects edges whose **parent** satisfies the inner predicate. Here `E2` and `E3` both match because their parent `E1` uses `protocol="http"`.

**Putting it together:**

```
subgraph edge depends on edge.protocol == "http" scope
| make edge.status = "reviewed"
```

This selects edges whose parent uses HTTP, expands to all descendants via `scope`, then marks all matched edges as `reviewed`.

### Questions to consider

- *Is the distinction between node predicates and edge predicates clear?*
- *Is forbidding mixed predicates too restrictive, or does it keep things simple?*
- *Do the dependency predicates (`edge parent exists`, `edge.depth`, `edge depends on`) feel intuitive for navigating edge hierarchies?*

---

## Step 5: Richer Predicates

Beyond equality (`==`), predicates support several other forms:

### Inequality

```
subgraph node.zone != "C"
```

True only if the attribute exists **and** the value differs. Missing attributes evaluate false.

### Attribute existence

```
subgraph node.team exists
subgraph edge.debug not exists
```

### Set membership

```
subgraph node in @critical
subgraph edge not in @deprecated
```

If the set doesn't exist: `in` → false, `not in` → true.

### Compound predicates (AND)

```
subgraph node.team == "payments" AND node.zone == "B"
```

Both sides must target the same element type. Only `AND` is supported — no `OR`.

### Questions to consider

- *Is `AND`-only sufficient? How often would you need `OR`?*
- *Is the "missing attribute → false" rule for inequality intuitive?*
- *Is `@` prefix for sets clear enough?*

---

## Step 6: Expanding Your Selection — Traversal

After matching, you often want to explore the neighbourhood. Traversal is a **suffix** on a `subgraph` expression:

```
subgraph node.team == "payments" traverse out 1
```

Starting from `payments` and `fraud` (the matched nodes), this follows outgoing edges one hop. It discovers `db` (via `payments -> db` and `fraud -> db`).

### Directions

| Direction | Meaning |
|-----------|---|
| `out`     | Follow outgoing edges |
| `in`      | Follow incoming edges |
| `both`    | Follow edges in both directions |
| `up`      | Follow parent dependency chain (`parent`) |
| `down`    | Follow child dependency chain (`Children`) |

Directions can be combined: `traverse out up 2` follows both graph edges and dependency chains.

**Scope** is shorthand for `traverse down all`:

```
subgraph edge.protocol == "http" scope
```

This is equivalent to:

```
subgraph edge.protocol == "http" traverse down all
```

### Depth

| Depth | Meaning |
|---|---|
| `1` | One hop |
| `N` | N hops |
| `all` | Unlimited (until frontier exhausted) |

### Examples

- *"What does the payments team depend on?"* → `subgraph node.team == "payments" traverse out 1`
- *"What depends on the database?"* → `subgraph node.team == "platform" traverse in all`

### Questions to consider

- *Is traversal as a suffix on `subgraph` natural, or would a standalone `| traverse out 1` expression feel better?*
- *Is unbounded traversal (`all`) risky on large graphs?*

---

## Step 7: Transforming and Removing

### Assigning attributes with `make`

```
make node.status = "reviewed" where node.team == "payments"
```

This creates or overwrites the `status` attribute on matching nodes. The graph structure doesn't change.

### Removing edges

```
remove edge where edge.protocol == "tcp"
```

Matching edges are removed; nodes remain.

### Removing attributes

```
remove node.tmp where node.tmp exists
```

### Removing orphans

After removing edges, some nodes may have no connections:

```
subgraph node.team exists
| remove edge where edge.protocol == "tcp"
| remove orphans
```

A node with a self-loop is **not** an orphan.

### Questions to consider

- *Is `make` a good verb for attribute assignment? Alternatives: `set`, `assign`, `tag`.*
- *Is the interaction between removal stages and the pipeline clear?*

---

## Step 8: Collapsing Nodes

`collapse` merges multiple nodes into one:

```
collapse into platform_group where node.team == "platform"
```

This merges `db` and `cache` into a single `platform_group` node.

### What happens during collapse

1. Nodes matching the predicate are selected
2. A new node `platform_group` is created
3. Attributes from collapsed nodes are merged
4. Edges are redirected to the new node
5. Original nodes are removed
6. Internal edges (between collapsed nodes) are removed
7. Duplicate edges created by the merge are **deduplicated**

**Important:** Edge deduplication happens **only** during collapse. All other operations preserve duplicate edges.

### Questions to consider

- *Is the explicit target ID (`into <id>`) useful, or would grouping by attribute (`collapse by team`) be more convenient?*
- *Is it surprising that deduplication only happens here?*

---

## Step 9: Named Graphs and Graph Algebra

You can save pipeline results as named graphs and combine them.

### Binding a named graph

```
(subgraph node.team == "payments") as PAY
```

Wrap a pipeline in parentheses and bind it with `as`. Names must be **uppercase** (`[A-Z][A-Z0-9_]*`) and are immutable — you can't rebind.

### Combining graphs

| Operator | Meaning |
|---|---|
| `PAY + ID` | Union — all nodes and edges from both |
| `PAY & ID` | Intersection — only shared elements |
| `PAY - ID` | Difference — in PAY but not ID |
| `PAY ^ ID` | Symmetric difference — in exactly one |

### Attribute conflicts

When the same node exists in both graphs, right-hand side wins:

```
node api [team="gateway", zone="A"]  +  node api [team="platform"]
=
node api [team="platform", zone="A"]
```

### Full example

```
(subgraph node.team == "payments") as PAY
| from *
| (subgraph node.team == "identity") as ID
| PAY + ID
```

### Questions to consider

- *Is the uppercase naming convention clear enough to distinguish named graphs from regular identifiers?*
- *Is parenthesized binding intuitive?*
- *Would you need more than two operands in graph algebra (e.g., `A + B + C`)?*

---

## Step 10: Putting It All Together

**"Show me what the payments team depends on, excluding TCP connections, and collapse the platform nodes."**

```
subgraph node.team == "payments" traverse out 1
| remove edge where edge.protocol == "tcp"
| remove orphans
| collapse into platform where node.team == "platform"
```

Reading left to right:

1. Select payments team nodes and follow outgoing edges one hop
2. Remove all TCP edges from that result
3. Remove any nodes left with no connections
4. Merge remaining platform nodes into a single `platform` node

---

## Quick Reference

### Expression summary

| Expression | Syntax | Purpose |
|---|---|---|
| Source | `from *` / `from NAME` | Switch working graph |
| Subgraph | `subgraph <predicate> [traverse <dir> <depth>]` | Filter and optionally traverse |
| Make | `make <path> = <value> where <predicate>` | Assign attributes |
| Remove | `remove edge where <predicate>` | Delete matching edges |
| Remove | `remove <path> where <predicate>` | Delete attributes |
| Remove | `remove orphans` | Delete isolated nodes |
| Collapse | `collapse into <id> where <predicate>` | Merge nodes |
| Binding | `(<pipeline>) as NAME` | Save intermediate result |
| Algebra | `A + B`, `A & B`, `A - B`, `A ^ B` | Combine named graphs |

### Predicate quick reference

| Form | Example |
|---|---|
| Equality | `node.team == "payments"` |
| Inequality | `node.zone != "C"` |
| Exists | `node.team exists` |
| Not exists | `edge.debug not exists` |
| Set membership | `node in @critical` |
| Set non-membership | `edge not in @deprecated` |
| Compound | `node.team == "payments" AND node.zone == "B"` |
| Parent exists | `edge parent exists` |
| Parent not exists | `edge parent not exists` |
| Edge depth | `edge.depth == 1` |
| Parent depends on | `edge depends on edge.protocol == "http"` |
