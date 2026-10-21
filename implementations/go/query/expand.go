// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.
package query

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
)

// matchKeywords are the identifiers that carry grammatical meaning inside a
// dialect match expression and MUST NOT be treated as bare attribute names.
var matchKeywords = map[string]bool{
	"AND":     true,
	"parent":  true,
	"exists":  true,
	"not":     true,
	"in":      true,
	"depends": true,
	"on":      true,
	"depth":   true,
	"true":    true,
	"false":   true,
}

// Token types resolved against the shared query lexer, so this file does not
// duplicate the lexer definition.
var (
	identToken    = queryLexer.Symbols()["Ident"]
	stringToken   = queryLexer.Symbols()["String"]
	numberToken   = queryLexer.Symbols()["Number"]
	operatorToken = queryLexer.Symbols()["Operator"]
	assignToken   = queryLexer.Symbols()["Assign"]
	atToken       = queryLexer.Symbols()["At"]
)

// ExpandMatchQuery expands a dialect match expression into a full GQL query.
//
// The expression is a bare attribute predicate: a single `subgraph` predicate
// written without the `subgraph` keyword or an element prefix. Each bare
// attribute reference is prefixed with element (the definition's
// representation) and the result is wrapped in `subgraph`.
//
// For example, with element "node", `kind == "service" AND active exists`
// expands to `subgraph node.kind == "service" AND node.active exists`.
//
// The rewrite is token-aware: string literals and numbers are never modified,
// and grammatical keywords (`AND`, `exists`, `not`, `in`, `parent`, `depends`,
// `on`) are preserved. `depth` is a derived edge attribute and always expands
// to `edge.depth`, regardless of element.
//
// ExpandMatchQuery returns an error if the expression contains an explicit
// element prefix, an explicit `subgraph` keyword, a pipeline, or another token
// that cannot appear in a bare match predicate. Callers MUST NOT fall back to
// evaluating the unexpanded input.
func ExpandMatchQuery(element, expr string) (string, error) {
	if element != "node" && element != "edge" {
		return "", fmt.Errorf("unsupported element prefix %q", element)
	}

	lex, err := queryLexer.Lex("", strings.NewReader(expr))
	if err != nil {
		return "", fmt.Errorf("invalid match expression: %w", err)
	}
	raw, err := lexer.ConsumeAll(lex)
	if err != nil {
		return "", fmt.Errorf("invalid match expression: %w", err)
	}

	// Drop comments and whitespace.
	tokens := raw[:0]
	for _, tok := range raw {
		if tok.Type == lexer.EOF {
			continue
		}
		if name := tokenTypeName(tok.Type); name == "Comment" || name == "Whitespace" {
			continue
		}
		tokens = append(tokens, tok)
	}
	if len(tokens) == 0 {
		return "", fmt.Errorf("empty match expression")
	}

	var out []string
	out = append(out, "subgraph")

	sawTerm := false
	expectTerm := true // at the start of a predicate term
	afterAt := false   // the next ident is a set reference, not an attribute

	emit := func(s string) { out = append(out, s) }

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		switch tok.Type {
		case identToken:
			word := tok.Value

			if afterAt {
				// A set reference following `@`; never an attribute. Attach it
				// to the `@` already emitted so the set ref stays one token.
				out[len(out)-1] += word
				afterAt = false
				expectTerm = false
				continue
			}

			// Reject the expanded forms explicitly.
			switch word {
			case "subgraph":
				return "", fmt.Errorf("the subgraph keyword is implied and must not be written")
			case "node", "edge", "set":
				return "", fmt.Errorf("element prefix %q must not be written; the representation supplies it", word)
			}

			if !matchKeywords[word] {
				// A bare attribute name, only valid at the start of a term.
				if !expectTerm {
					return "", fmt.Errorf("unexpected identifier %q in match expression", word)
				}
				emit(element + "." + word)
				expectTerm = false
				sawTerm = true
				continue
			}

			// Grammatical keywords.
			switch word {
			case "AND":
				emit("AND")
				expectTerm = true
			case "depth":
				// A derived edge attribute, independent of the element.
				emit("edge.depth")
				expectTerm = false
				sawTerm = true
			case "parent":
				emit(element)
				emit("parent")
				expectTerm = false
				sawTerm = true
			case "depends":
				emit(element)
				emit("depends")
				expectTerm = false
				sawTerm = true
			case "on":
				emit("on")
				// The inner predicate of `depends on` is a fresh term.
				expectTerm = true
			case "in":
				if expectTerm {
					emit(element)
				}
				emit("in")
				expectTerm = false
				sawTerm = true
			case "not":
				// `not in` (bare set non-membership) versus `attr not exists`.
				if expectTerm && i+1 < len(tokens) && tokens[i+1].Value == "in" {
					emit(element)
					emit("not")
					emit("in")
					i++ // consume the `in` token
					expectTerm = false
					sawTerm = true
					continue
				}
				emit("not")
			case "exists":
				emit("exists")
				expectTerm = false
			case "true", "false":
				emit(word)
				expectTerm = false
			default:
				emit(word)
				expectTerm = false
			}

		case atToken:
			emit(tok.Value)
			afterAt = true

		case stringToken, numberToken, operatorToken, assignToken:
			emit(tok.Value)
			expectTerm = false

		default:
			return "", fmt.Errorf("unsupported token %q in match expression", tok.Value)
		}
	}

	if !sawTerm {
		return "", fmt.Errorf("match expression contains no attribute predicate")
	}
	return strings.Join(out, " "), nil
}

// tokenTypeName maps a lexer token type back to its symbol name.
func tokenTypeName(tt lexer.TokenType) string {
	for name, t := range queryLexer.Symbols() {
		if t == tt {
			return name
		}
	}
	return ""
}
