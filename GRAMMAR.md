# GSL-Lang Grammar

This is the formal grammar for GSL-Lang v0.1, expressed in Extended Backus-Naur Form (EBNF). It defines the syntax of valid GSL documents.

```ebnf
program      ::= statement*

statement    ::= node_decl
               | edge_decl
               | set_decl

node_decl    ::= "node" IDENT node_suffix? membership*

node_suffix  ::= attribute_list
               | ":" STRING
               | block

block        ::= "{" node_decl* "}"

edge_decl    ::= IDENT "->" node_list edge_suffix? membership*
               | node_list "->" IDENT edge_suffix? membership*

node_list    ::= IDENT ("," IDENT)*

edge_suffix  ::= attribute_list
               | ":" STRING

set_decl     ::= "set" IDENT attribute_list?

membership   ::= "@" IDENT

attribute_list ::= "[" (attribute ("," attribute)*)? "]"

attribute    ::= IDENT ("=" value)?

value        ::= STRING
               | NUMBER
               | BOOLEAN
               | IDENT  (NodeRef only in node context)
```