package lsp

import (
	"context"
	"strings"

	"github.com/dnnrly/gsl-lang"
	"go.lsp.dev/protocol"
)

func (s *Server) Completion(ctx context.Context, params *protocol.CompletionParams) (protocol.CompletionResult, error) {
	s.mu.Lock()
	doc, ok := s.documents[params.TextDocument.URI]
	s.mu.Unlock()
	if !ok {
		return protocol.CompletionItemSlice{}, nil
	}

	if isGQLFile(doc.docURI) {
		items := completeGQL(doc.content, int(params.Position.Line), int(params.Position.Character))
		return protocol.CompletionItemSlice(items), nil
	}

	pos := params.Position
	items := s.complete(doc, int(pos.Line), int(pos.Character))
	return protocol.CompletionItemSlice(items), nil
}

func (s *Server) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	s.mu.Lock()
	doc, ok := s.documents[params.TextDocument.URI]
	s.mu.Unlock()
	if !ok {
		return nil, nil
	}

	if isGQLFile(doc.docURI) {
		return hoverGQL(doc, int(params.Position.Line), int(params.Position.Character)), nil
	}

	if doc.graph == nil {
		return nil, nil
	}

	pos := params.Position
	word := wordAt(doc.content, int(pos.Line), int(pos.Character))
	if word == nil {
		return nil, nil
	}

	info := s.hoverFor(doc.graph, word.word)
	if info == "" {
		return nil, nil
	}

	rnj := protocol.Range{
		Start: protocol.Position{Line: uint32(word.line), Character: uint32(word.colS)},
		End:   protocol.Position{Line: uint32(word.line), Character: uint32(word.colE)},
	}

	return &protocol.Hover{
		Contents: &protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: info,
		},
		Range: &rnj,
	}, nil
}

func (s *Server) Definition(ctx context.Context, params *protocol.DefinitionParams) (protocol.DefinitionResult, error) {
	s.mu.Lock()
	doc, ok := s.documents[params.TextDocument.URI]
	s.mu.Unlock()
	if !ok || isGQLFile(doc.docURI) {
		return nil, nil
	}

	pos := params.Position
	word := wordAt(doc.content, int(pos.Line), int(pos.Character))
	if word == nil {
		return nil, nil
	}

	tokens := tokenize(doc.content)
	loc := findDeclPosition(tokens, word.word, doc.docURI)
	if loc == nil {
		return nil, nil
	}
	return loc, nil
}

func (s *Server) References(ctx context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	s.mu.Lock()
	doc, ok := s.documents[params.TextDocument.URI]
	s.mu.Unlock()
	if !ok || isGQLFile(doc.docURI) {
		return nil, nil
	}

	pos := params.Position
	word := wordAt(doc.content, int(pos.Line), int(pos.Character))
	if word == nil {
		return nil, nil
	}

	tokens := tokenize(doc.content)
	return findReferences(tokens, word.word, doc.docURI), nil
}

func (s *Server) DocumentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) (protocol.DocumentSymbolResult, error) {
	s.mu.Lock()
	doc, ok := s.documents[params.TextDocument.URI]
	s.mu.Unlock()
	if !ok {
		return protocol.DocumentSymbolSlice{}, nil
	}

	if isGQLFile(doc.docURI) {
		symbols := gqlDocumentSymbols(doc.content)
		return protocol.DocumentSymbolSlice(symbols), nil
	}

	tokens := tokenize(doc.content)
	symbols := buildSymbols(tokens)
	return protocol.DocumentSymbolSlice(symbols), nil
}

func (s *Server) Formatting(ctx context.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	s.mu.Lock()
	doc, ok := s.documents[params.TextDocument.URI]
	s.mu.Unlock()
	if !ok || doc.graph == nil || isGQLFile(doc.docURI) {
		return nil, nil
	}

	formatted := gsl.Serialize(doc.graph)
	lines := strings.Split(doc.content, "\n")
	lastLine := uint32(max(0, len(lines)-1))
	lastChar := uint32(0)
	if len(lines) > 0 {
		lastChar = uint32(len(lines[len(lines)-1]))
	}

	return []protocol.TextEdit{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: lastLine, Character: lastChar},
			},
			NewText: formatted,
		},
	}, nil
}

func (s *Server) SemanticTokensFull(ctx context.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	s.mu.Lock()
	doc, ok := s.documents[params.TextDocument.URI]
	s.mu.Unlock()
	if !ok {
		return &protocol.SemanticTokens{Data: []uint32{}}, nil
	}

	if isGQLFile(doc.docURI) {
		data := gqlSemanticTokens(doc.content)
		return &protocol.SemanticTokens{Data: data}, nil
	}

	tokens := tokenize(doc.content)
	data := encodeSemanticTokens(tokens)
	return &protocol.SemanticTokens{Data: data}, nil
}
