# GSL-QL EBNF Grammar (Spec v0.3)

## 1. Query Structure

```ebnf
query
    = graph_source , pipeline ;

pipeline
    = { stage } ;

stage
    = "|" , match_stage
    | "|" , traverse_stage
    | "|" , collapse_stage
    | "|" , remove_stage
    | "|" , subgraph_stage ;
```

---

# 2. Graph Source

```ebnf
graph_source
    = "graph" , identifier ;
```

---

# 3. Match Stage

```ebnf
match_stage
    = "match" , predicate ;
```

---

# 4. Predicate

```ebnf
predicate
    = node_predicate
    | edge_predicate ;
```

---

# 5. Node Predicate

```ebnf
node_predicate
    = "node" ,
      [ identifier ] ,
      [ attribute_block ] ,
      [ membership_clause ] ;
```

Examples parsed by this rule:

```
node
node A
node[team="payments"]
node A[zone="A"]
node A[team="payments"] in graph G
```

---

# 6. Edge Predicate

```ebnf
edge_predicate
    = "edge" ,
      [ edge_filter_block ] ;
```

Examples:

```
edge
edge[type="async"]
edge[source=A]
edge[target=B]
```

---

# 7. Traversal

```ebnf
traverse_stage
    = "traverse" , direction ;

direction
    = "out"
    | "in"
    | "all" ;
```

Examples:

```
traverse out
traverse in
traverse all
```

---

# 8. Collapse

```ebnf
collapse_stage
    = "collapse" , "by" , identifier ;
```

Example:

```
collapse by team
```

---

# 9. Remove

```ebnf
remove_stage
    = "remove" , remove_target ;

remove_target
    = "node"
    | "edge"
    | "orphans" ;
```

Examples:

```
remove node
remove edge
remove orphans
```

---

# 10. Subgraph

```ebnf
subgraph_stage
    = "subgraph" , identifier ;
```

Example:

```
subgraph payments
```

---

# 11. Membership Clause

Used in node predicates.

```ebnf
membership_clause
    = "in" , "graph" , identifier ;
```

Example:

```
node in graph A
```

---

# 12. Attribute Block

```ebnf
attribute_block
    = "[" , attribute_list , "]" ;

attribute_list
    = attribute , { "," , attribute } ;

attribute
    = identifier , "=" , value ;
```

Examples:

```
[team="payments"]
[zone="A",team="fraud"]
```

---

# 13. Edge Filter Block

Edge filters allow both attributes and structural filters.

```ebnf
edge_filter_block
    = "[" , edge_filter_list , "]" ;

edge_filter_list
    = edge_filter , { "," , edge_filter } ;

edge_filter
    = attribute
    | "source" , "=" , identifier
    | "target" , "=" , identifier ;
```

Examples:

```
[source=A]
[target=B]
[type="async"]
[source=A,target=B]
```

---

# 14. Values

```ebnf
value
    = string
    | number
    | boolean
    | identifier ;
```

---

# 15. Identifiers

```ebnf
identifier
    = letter , { letter | digit | "_" } ;
```

---

# 16. Literals

```ebnf
string
    = '"' , { character } , '"' ;

number
    = digit , { digit } ;

boolean
    = "true"
    | "false" ;
```

---

# 17. Character Classes

```ebnf
letter
    = "A" | "B" | ... | "Z"
    | "a" | "b" | ... | "z" ;

digit
    = "0" | "1" | "2" | "3" | "4"
    | "5" | "6" | "7" | "8" | "9" ;
```

---

# 18. Reserved Keywords

Identifiers **must not match these**:

```
graph
match
node
edge
traverse
collapse
by
remove
orphans
subgraph
in
source
target
true
false
```

---

# 19. Example Queries Parsed by This Grammar

### Basic

```
graph prod
| match node
```

---

### Attribute filter

```
graph prod
| match node[team="payments"]
```

---

### Node + identifier

```
graph prod
| match node A
```

---

### Node + attributes + membership

```
graph prod
| match node A[team="payments"] in graph baseline
```

---

### Traversal

```
graph prod
| match node[team="payments"]
| traverse out
```

---

### Collapse

```
graph prod
| collapse by team
```

---

### Removal

```
graph prod
| remove orphans
```

---

### Subgraph extraction

```
graph prod
| match node[team="payments"]
| subgraph payments
```

---

# 20. Intentional Grammar Constraints

The grammar **intentionally forbids**:

### Boolean expressions

Not allowed:

```
node[team="payments" AND zone="A"]
```

---

### Nested predicates

Not allowed:

```
node[node[...]]
```

---

### Path expressions

Not allowed:

```
A -> B -> C
```

---

### Multiple predicates

Not allowed:

```
match node ... edge ...
```

This keeps the parser simple and deterministic.

---

# 21. Parser Simplicity (Important)

This grammar is intentionally structured so that:

* **node_predicate always starts with `node`**
* **edge_predicate always starts with `edge`**
* **stages always start with a keyword**

This means a parser can decide **what rule to use from the first token**.

No backtracking required.

---

# 22. Implementation Hint (Go)

The AST shape implied by this grammar is roughly:

```
Query
 ├─ GraphSource
 └─ []Stage

Stage
 ├─ MatchStage
 ├─ TraverseStage
 ├─ CollapseStage
 ├─ RemoveStage
 └─ SubgraphStage
```

Predicates:

```
Predicate
 ├─ NodePredicate
 └─ EdgePredicate
```
