# GSL-Lang Grammar

This is the formal grammar for GSL-Lang v1.0, expressed in Extended Backus-Naur Form (EBNF). It defines the syntax of valid GSL documents.

```ebnf
program      ::= statement*

statement    ::= node_decl
               | edge_decl
               | scoped_edge_decl
               | set_decl

node_decl    ::= "node" IDENT node_suffix? membership*

node_suffix  ::= attribute_list
               | ":" STRING
               | block

block        ::= "{" statement* "}"

edge_decl    ::= edge_label? edge_expr edge_suffix? membership*

edge_expr    ::= IDENT "->" node_list
               | node_list "->" IDENT

scoped_edge_decl ::= edge_label? edge_expr block

edge_label   ::= IDENT ":"

node_list    ::= IDENT ("," IDENT)*

edge_suffix  ::= attribute_list
               | ":" STRING
               | attribute_list? ":" STRING

set_decl     ::= "set" IDENT attribute_list?

membership   ::= "@" IDENT

attribute_list ::= "[" (attribute ("," attribute)*)? "]"

attribute    ::= IDENT ("=" value)?
               | "depends_on" "=" IDENT

value        ::= STRING
               | NUMBER
               | BOOLEAN
               | IDENT  (NodeRef only in node context)
```