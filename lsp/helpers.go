package lsp

import (
	"strings"

	"github.com/dnnrly/gsl-lang"
	"go.lsp.dev/protocol"
)

func wordAt(content string, line, character int) *wordAtPos {
	lines := strings.Split(content, "\n")
	if line >= len(lines) {
		return nil
	}
	lineStr := lines[line]
	if character > len(lineStr) {
		character = len(lineStr)
	}

	start := character
	for start > 0 && isIdentRune(rune(lineStr[start-1])) {
		start--
	}
	end := character
	for end < len(lineStr) && isIdentRune(rune(lineStr[end])) {
		end++
	}

	if start >= end {
		if character < len(lineStr) && lineStr[character] == '@' {
			return &wordAtPos{word: "@", line: line, colS: character, colE: character + 1}
		}
		return nil
	}

	return &wordAtPos{
		word: lineStr[start:end],
		line: line,
		colS: start,
		colE: end,
	}
}

func isIdentRune(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_'
}

func tokenize(content string) []gsl.Token {
	l, err := gsl.NewLexer(strings.NewReader(content))
	if err != nil {
		return nil
	}
	return l.Tokenize()
}

func lineColToPos(tok gsl.Token) protocol.Position {
	return protocol.Position{
		Line:      uint32(max(0, tok.Line-1)),
		Character: uint32(max(0, tok.Column-1)),
	}
}
