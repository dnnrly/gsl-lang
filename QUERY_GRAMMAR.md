# GSL-QL EBNF Grammar (Spec v0.3)

## 1. Query Structure

```ebnf
query
    = [ graph_source ] , pipeline ;

pipeline
    = stage , { "|" , stage } ;

stage
    = subgraph_stage
    | make_stage
    | remove_stage
    | collapse_stage
    | from_stage
    | binding
    | graph_algebra ;
```

---

# 2. Graph Source

```ebnf
graph_source
    = "from" , ( identifier | "*" ) ;
```

`*` resets the working graph to the original input graph.

If no `graph_source` is given, the working graph is the input graph.

---

# 3. Subgraph Stage

```ebnf
subgraph_stage
    = "subgraph" , predicate , [ traverse_clause ] ;

traverse_clause
    = "traverse" , direction , depth ;

direction
    = "in"
    | "out"
    | "both" ;

depth
    = INTEGER
    | "all" ;
```

Examples:

```
subgraph node.team == "payments"
subgraph node.team == "payments" traverse out 1
subgraph edge.protocol == "http" traverse all all
```

---

# 4. Predicate

```ebnf
predicate
    = predicate_term , { "AND" , predicate_term } ;

predicate_term
    = equality_predicate
    | inequality_predicate
    | exists_predicate
    | not_exists_predicate
    | set_membership_predicate
    | set_non_membership_predicate ;
```

All predicate terms in a compound predicate MUST target the same element type (`node` or `edge`). Mixed references are an error.

---

# 5. Equality Predicate

```ebnf
equality_predicate
    = attribute_path , "==" , value ;
```

Examples:

```
node.team == "payments"
edge.protocol == "http"
```

---

# 6. Inequality Predicate

```ebnf
inequality_predicate
    = attribute_path , "!=" , value ;
```

Examples:

```
node.zone != "C"
edge.protocol != "grpc"
```

---

# 7. Exists Predicate

```ebnf
exists_predicate
    = attribute_path , "exists" ;
```

Example:

```
node.team exists
```

---

# 8. Not Exists Predicate

```ebnf
not_exists_predicate
    = attribute_path , "not" , "exists" ;
```

Example:

```
edge.protocol not exists
```

---

# 9. Set Membership Predicate

```ebnf
set_membership_predicate
    = element_type , "in" , set_ref ;

set_non_membership_predicate
    = element_type , "not" , "in" , set_ref ;

element_type
    = "node"
    | "edge" ;

set_ref
    = "@" , identifier ;
```

If the set does not exist: `in` evaluates **false**, `not in` evaluates **true**.

Examples:

```
node in @critical
edge not in @deprecated
```

---

# 10. Attribute Path

```ebnf
attribute_path
    = element_type , "." , identifier ;
```

The prefix determines whether the predicate targets nodes or edges.

Examples:

```
node.team
edge.protocol
```

---

# 11. Make Stage

```ebnf
make_stage
    = "make" , attribute_path , "=" , value , "where" , predicate ;
```

Creates or overwrites the attribute on matching elements. Graph structure is unchanged.

Examples:

```
make node.status = "reviewed" where node.team == "payments"
make edge.encrypted = true where edge.protocol == "http"
```

---

# 12. Remove Stage

```ebnf
remove_stage
    = "remove" , "edge" , "where" , predicate
    | "remove" , attribute_path , "where" , predicate
    | "remove" , "orphans" ;
```

Examples:

```
remove edge where edge.protocol == "tcp"
remove node.tmp where node.tmp exists
remove edge.debug where edge.debug exists
remove orphans
```

---

# 13. Collapse Stage

```ebnf
collapse_stage
    = "collapse" , "into" , identifier , "where" , predicate ;
```

Collapses all nodes matching the predicate into a single node with the given identifier.

Example:

```
collapse into platform_group where node.team == "platform"
```

---

# 14. From Stage

```ebnf
from_stage
    = "from" , ( identifier | "*" ) ;
```

Switches the working graph mid-pipeline.

Examples:

```
from *
from PAYMENTS
```

---

# 15. Named Graph Binding

```ebnf
binding
    = "(" , pipeline , ")" , "as" , named_graph_id ;

named_graph_id
    = uppercase_letter , { uppercase_letter | digit | "_" } ;
```

Named graph identifiers MUST be uppercase: `[A-Z][A-Z0-9_]*`.

Named graphs are immutable — rebinding a name is an error.

Example:

```
(subgraph node.team == "payments") as PAY
```

---

# 16. Graph Algebra

```ebnf
graph_algebra
    = named_graph_id , algebra_op , named_graph_id ;

algebra_op
    = "+"
    | "&"
    | "-"
    | "^" ;
```

| Operator | Meaning              |
|----------|----------------------|
| `+`      | union                |
| `&`      | intersection         |
| `-`      | difference           |
| `^`      | symmetric difference |

Node attribute conflicts resolved by right-hand side wins. Duplicate edges are preserved.

Example:

```
PAY + ID
PAY - PLATFORM
```

---

# 17. Values

```ebnf
value
    = string
    | number
    | boolean
    | identifier ;
```

---

# 18. Identifiers

```ebnf
identifier
    = letter , { letter | digit | "_" } ;
```

---

# 19. Literals

```ebnf
string
    = '"' , { character } , '"' ;

number
    = digit , { digit } ;

boolean
    = "true"
    | "false" ;

integer
    = digit , { digit } ;
```

---

# 20. Character Classes

```ebnf
letter
    = "A" | "B" | ... | "Z"
    | "a" | "b" | ... | "z" ;

uppercase_letter
    = "A" | "B" | ... | "Z" ;

digit
    = "0" | "1" | "2" | "3" | "4"
    | "5" | "6" | "7" | "8" | "9" ;
```

---

# 21. Comments

Comments begin with `#` and continue to end of line.

```ebnf
comment
    = "#" , { character } , NEWLINE ;
```

---

# 22. Reserved Keywords

Identifiers **must not match these**:

```
subgraph
from
as
traverse
make
remove
collapse
into
where
AND
in
out
exists
not
orphans
all
node
edge
true
false
```

---

# 23. Example Queries Parsed by This Grammar

### Basic subgraph

```
subgraph node.team == "payments"
```

---

### Subgraph with traversal

```
subgraph node.team == "payments" traverse out 1
```

---

### Compound predicate

```
subgraph node.team == "payments" AND node.zone == "B"
```

---

### Edge subgraph

```
subgraph edge.protocol == "http"
```

---

### Attribute assignment

```
subgraph node.team == "payments"
| make node.reviewed = true where node.zone == "B"
```

---

### Conditional edge removal

```
subgraph node.team exists
| remove edge where edge.protocol == "tcp"
| remove orphans
```

---

### Collapse

```
subgraph node.team exists
| collapse into platform where node.team == "platform"
```

---

### Named graphs and algebra

```
(subgraph node.team == "payments") as PAY
| from *
| (subgraph node.team == "identity") as ID
| PAY + ID
```

---

### Set membership

```
subgraph node in @critical
```

---

### Unlimited traversal

```
subgraph node.team == "payments" traverse out all
```

---

### Switch working graph

```
(subgraph node.team == "payments") as PAY
| from PAY
| remove orphans
```

---

# 24. Intentional Grammar Constraints

The grammar **intentionally forbids**:

### OR expressions

Not allowed:

```
node.team == "payments" OR node.team == "fraud"
```

---

### Nested predicates

Not allowed:

```
subgraph (subgraph ...)
```

---

### Path expressions

Not allowed:

```
A -> B -> C
```

---

### Mixed predicate targets

Not allowed:

```
node.team == "payments" AND edge.protocol == "http"
```

This keeps the parser simple and deterministic.

---

# 25. Parser Simplicity (Important)

This grammar is intentionally structured so that:

* **Predicates are typed by their attribute path prefix** (`node.` or `edge.`)
* **Stages always start with a keyword** (`subgraph`, `make`, `remove`, `collapse`, `from`)
* **Named graph references are uppercase**, distinguishing them from regular identifiers

This means a parser can decide **what rule to use from the first token**.

No backtracking required.

---

# 26. Implementation Hint (Go)

The AST shape implied by this grammar is roughly:

```
Query
 ├─ GraphSource (optional)
 └─ []Stage

Stage
 ├─ SubgraphStage
 │    ├─ Predicate
 │    └─ TraverseClause (optional)
 ├─ MakeStage
 │    ├─ AttributePath
 │    ├─ Value
 │    └─ Predicate
 ├─ RemoveStage
 │    ├─ RemoveEdge + Predicate
 │    ├─ RemoveAttribute + Predicate
 │    └─ RemoveOrphans
 ├─ CollapseStage
 │    ├─ Identifier
 │    └─ Predicate
 ├─ FromStage
 ├─ Binding
 │    ├─ Pipeline
 │    └─ NamedGraphId
 └─ GraphAlgebra
      ├─ NamedGraphId (left)
      ├─ Operator
      └─ NamedGraphId (right)
```

Predicates:

```
Predicate
 └─ []PredicateTerm (joined by AND)

PredicateTerm
 ├─ EqualityPredicate (path == value)
 ├─ InequalityPredicate (path != value)
 ├─ ExistsPredicate (path exists)
 ├─ NotExistsPredicate (path not exists)
 ├─ SetMembershipPredicate (element in @set)
 └─ SetNonMembershipPredicate (element not in @set)
```
