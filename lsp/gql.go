package lsp

import (
	"fmt"
	"strings"

	"go.lsp.dev/protocol"
)

var gqlKeywords = []struct {
	label  string
	detail string
	kind   protocol.CompletionItemKind
}{
	{"subgraph", "Filter nodes/edges by predicate", protocol.CompletionItemKindKeyword},
	{"from", "Select input graph", protocol.CompletionItemKindKeyword},
	{"make", "Set attributes matching predicate", protocol.CompletionItemKindKeyword},
	{"remove", "Remove edges, attributes, or orphans", protocol.CompletionItemKindKeyword},
	{"collapse", "Merge nodes matching predicate", protocol.CompletionItemKindKeyword},
	{"into", "Target node ID for collapse", protocol.CompletionItemKindKeyword},
	{"where", "Predicate clause", protocol.CompletionItemKindKeyword},
	{"traverse", "Graph traversal", protocol.CompletionItemKindKeyword},
	{"scope", "Sugar for traverse down all", protocol.CompletionItemKindKeyword},
	{"node", "Element type for predicate", protocol.CompletionItemKindKeyword},
	{"edge", "Element type for predicate", protocol.CompletionItemKindKeyword},
	{"exists", "Attribute existence check", protocol.CompletionItemKindKeyword},
	{"not", "Negation", protocol.CompletionItemKindKeyword},
	{"AND", "Logical AND", protocol.CompletionItemKindKeyword},
	{"in", "Set membership / traversal direction", protocol.CompletionItemKindKeyword},
	{"out", "Traversal direction", protocol.CompletionItemKindKeyword},
	{"both", "Traversal direction", protocol.CompletionItemKindKeyword},
	{"up", "Traversal direction (dependency)", protocol.CompletionItemKindKeyword},
	{"down", "Traversal direction (dependency)", protocol.CompletionItemKindKeyword},
	{"all", "Unlimited traversal depth", protocol.CompletionItemKindKeyword},
	{"orphans", "Remove isolated nodes", protocol.CompletionItemKindKeyword},
	{"parent", "Edge parent predicate", protocol.CompletionItemKindKeyword},
	{"depth", "Edge dependency depth", protocol.CompletionItemKindKeyword},
	{"depends", "Edge dependency predicate", protocol.CompletionItemKindKeyword},
	{"on", "Part of 'depends on'", protocol.CompletionItemKindKeyword},
	{"true", "Boolean true", protocol.CompletionItemKindKeyword},
	{"false", "Boolean false", protocol.CompletionItemKindKeyword},
	{"as", "Named graph binding", protocol.CompletionItemKindKeyword},
}

func completeGQL(content string, line, character int) []protocol.CompletionItem {
	lines := strings.Split(content, "\n")
	if line >= len(lines) {
		return keywordCompletionsGQL("")
	}
	lineStr := lines[line]
	prefix := lineStr[:min(character, len(lineStr))]
	trimmed := strings.TrimSpace(prefix)

	lastWord := trimmed
	if idx := strings.LastIndexAny(trimmed, " \t|("); idx >= 0 {
		lastWord = trimmed[idx+1:]
	}

	return keywordCompletionsGQL(lastWord)
}

func keywordCompletionsGQL(prefix string) []protocol.CompletionItem {
	var items []protocol.CompletionItem
	for _, kw := range gqlKeywords {
		if prefix == "" || strings.HasPrefix(kw.label, prefix) {
			items = append(items, protocol.CompletionItem{
				Label:      kw.label,
				Detail:     protocol.NewOptional(kw.detail),
				Kind:       kw.kind,
				InsertText: protocol.NewOptional(kw.label),
			})
		}
	}
	return items
}

func hoverGQL(doc *document, line, character int) *protocol.Hover {
	word := wordAt(doc.content, line, character)
	if word == nil {
		return nil
	}

	for _, kw := range gqlKeywords {
		if kw.label == word.word {
			rnj := protocol.Range{
				Start: protocol.Position{Line: uint32(word.line), Character: uint32(word.colS)},
				End:   protocol.Position{Line: uint32(word.line), Character: uint32(word.colE)},
			}
			return &protocol.Hover{
				Contents: &protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: fmt.Sprintf("**`%s`** — %s", kw.label, kw.detail),
				},
				Range: &rnj,
			}
		}
	}
	return nil
}

func gqlDocumentSymbols(content string) []protocol.DocumentSymbol {
	lines := strings.Split(content, "\n")
	var symbols []protocol.DocumentSymbol

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		parts := strings.Split(trimmed, "|")
		col := 0
		for _, part := range parts {
			exp := strings.TrimSpace(part)
			if exp == "" {
				col += len(part) + 1
				continue
			}

			expType := ""
			switch {
			case strings.HasPrefix(exp, "subgraph"):
				expType = "subgraph"
			case strings.HasPrefix(exp, "from"):
				expType = "from"
			case strings.HasPrefix(exp, "make"):
				expType = "make"
			case strings.HasPrefix(exp, "remove"):
				expType = "remove"
			case strings.HasPrefix(exp, "collapse"):
				expType = "collapse"
			}

			if expType != "" {
				symbols = append(symbols, protocol.DocumentSymbol{
					Name:   expType,
					Detail: &exp,
					Kind:   protocol.SymbolKindFunction,
					Range: protocol.Range{
						Start: protocol.Position{Line: uint32(i), Character: uint32(col)},
						End:   protocol.Position{Line: uint32(i), Character: uint32(col + len(exp))},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: uint32(i), Character: uint32(col)},
						End:   protocol.Position{Line: uint32(i), Character: uint32(col + len(expType))},
					},
				})
			}
			col += len(part) + 1
		}
	}
	return symbols
}

func gqlSemanticTokens(content string) []uint32 {
	type gqlTok struct {
		line    int
		col     int
		len     int
		typeIdx uint32
	}
	var tokens []gqlTok

	lines := strings.Split(content, "\n")
	for lineIdx, line := range lines {
		i := 0
		for i < len(line) {
			if line[i] == ' ' || line[i] == '\t' {
				i++
				continue
			}

			if line[i] == '#' {
				break
			}

			if line[i] == '"' {
				end := i + 1
				for end < len(line) && line[end] != '"' {
					end++
				}
				if end < len(line) {
					end++
				}
				tokens = append(tokens, gqlTok{lineIdx, i, end - i, 1})
				i = end
				continue
			}

			if line[i] >= '0' && line[i] <= '9' || line[i] == '-' && i+1 < len(line) && line[i+1] >= '0' && line[i+1] <= '9' {
				end := i
				if line[end] == '-' {
					end++
				}
				for end < len(line) && (line[end] >= '0' && line[end] <= '9' || line[end] == '.') {
					end++
				}
				tokens = append(tokens, gqlTok{lineIdx, i, end - i, 2})
				i = end
				continue
			}

			if (line[i] >= 'a' && line[i] <= 'z') || (line[i] >= 'A' && line[i] <= 'Z') || line[i] == '_' {
				end := i
				for end < len(line) && ((line[end] >= 'a' && line[end] <= 'z') || (line[end] >= 'A' && line[end] <= 'Z') || (line[end] >= '0' && line[end] <= '9') || line[end] == '_') {
					end++
				}
				word := line[i:end]
				ti := uint32(0)
				for _, kw := range gqlKeywords {
					if kw.label == word {
						ti = 0
						break
					}
				}
				if word == "true" || word == "false" {
					ti = 3
				}
				tokens = append(tokens, gqlTok{lineIdx, i, end - i, ti})
				i = end
				continue
			}

			switch line[i] {
			case '|', '@', '.', '(', ')', '+', '&', '-', '^':
				tokens = append(tokens, gqlTok{lineIdx, i, 1, 6})
			case '=', '!':
				if i+1 < len(line) && line[i+1] == '=' {
					tokens = append(tokens, gqlTok{lineIdx, i, 2, 6})
					i += 2
					continue
				}
				tokens = append(tokens, gqlTok{lineIdx, i, 1, 6})
			}
			i++
		}
	}

	if len(tokens) == 0 {
		return []uint32{}
	}

	var data []uint32
	prevLine := uint32(0)
	prevChar := uint32(0)
	for _, tok := range tokens {
		deltaLine := uint32(tok.line) - prevLine
		deltaChar := uint32(tok.col)
		if deltaLine == 0 {
			deltaChar = uint32(tok.col) - prevChar
		}
		data = append(data, deltaLine, deltaChar, uint32(tok.len), tok.typeIdx, 0)
		prevLine = uint32(tok.line)
		prevChar = uint32(tok.col)
	}
	return data
}
