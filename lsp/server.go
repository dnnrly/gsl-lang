// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.
package lsp

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/query"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

type document struct {
	docURI  uri.URI
	content string
	version int32

	graph *gsl.Graph
	pErrs []protocol.Diagnostic

	gqlQuery *query.Query
	gqlErr   error
}

type wordAtPos struct {
	word string
	line int
	colS int
	colE int
}

type Server struct {
	protocol.UnimplementedServer

	mu        sync.Mutex
	client    protocol.Client
	documents map[uri.URI]*document
}

func NewServer(client protocol.Client) *Server {
	return &Server{
		client:    client,
		documents: make(map[uri.URI]*document),
	}
}

func (s *Server) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	openClose := true
	kind := protocol.TextDocumentSyncKindFull

	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				OpenClose: &openClose,
				Change:    &kind,
			},
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{"@", ".", ":", "["},
			},
			HoverProvider:              protocol.Boolean(true),
			DefinitionProvider:         protocol.Boolean(true),
			ReferencesProvider:         protocol.Boolean(true),
			DocumentSymbolProvider:     &protocol.DocumentSymbolOptions{},
			DocumentFormattingProvider: &protocol.DocumentFormattingOptions{},
			SemanticTokensProvider: &protocol.SemanticTokensOptions{
				Legend: protocol.SemanticTokensLegend{
					TokenTypes:     semanticTokenTypes(),
					TokenModifiers: []string{},
				},
				Range: protocol.Boolean(false),
				Full:  protocol.Boolean(true),
			},
		},
		ServerInfo: protocol.ServerInfo{
			Name:    "gsl-lsp",
			Version: protocol.NewOptional("0.1.0"),
		},
	}, nil
}

func (s *Server) Initialized(ctx context.Context, params *protocol.InitializedParams) error {
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.documents = nil
	return nil
}

func (s *Server) Exit(ctx context.Context) error {
	return nil
}

func (s *Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc := &document{
		docURI:  params.TextDocument.URI,
		content: params.TextDocument.Text,
		version: params.TextDocument.Version,
	}
	s.documents[doc.docURI] = doc
	s.parseAndDiagnose(ctx, doc)
	return nil
}

func (s *Server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, ok := s.documents[params.TextDocument.URI]
	if !ok {
		return nil
	}

	if len(params.ContentChanges) > 0 {
		change := params.ContentChanges[len(params.ContentChanges)-1]
		switch c := change.(type) {
		case *protocol.TextDocumentContentChangeWholeDocument:
			doc.content = c.Text
		case *protocol.TextDocumentContentChangePartial:
			var err error
			doc.content, err = applyPartialChange(doc.content, c)
			if err != nil {
				return err
			}
		}
	}
	doc.version = params.TextDocument.Version
	s.parseAndDiagnose(ctx, doc)
	return nil
}

// applyPartialChange applies an incremental text change to the current
// document content. The server advertises full-document sync, but it must
// not silently replace the whole document with a single changed snippet if a
// partial change is received.
func applyPartialChange(content string, c *protocol.TextDocumentContentChangePartial) (string, error) {
	start, err := positionToOffset(content, c.Range.Start)
	if err != nil {
		return "", err
	}
	end, err := positionToOffset(content, c.Range.End)
	if err != nil {
		return "", err
	}
	return content[:start] + c.Text + content[end:], nil
}

// positionToOffset converts a zero-based line/character position into a byte
// offset within content. Offsets are computed on the document as UTF-16-like
// LSP positions, but since GSL sources are ASCII in practice, characters are
// treated as bytes here.
func positionToOffset(content string, pos protocol.Position) (int, error) {
	line := int(pos.Line)
	col := int(pos.Character)
	offset := 0
	for currentLine := 0; currentLine < line; currentLine++ {
		idx := strings.IndexByte(content[offset:], '\n')
		if idx < 0 {
			return 0, fmt.Errorf("line %d out of range", line)
		}
		offset += idx + 1
	}
	if offset+col > len(content) {
		return 0, fmt.Errorf("column %d out of range on line %d", col, line)
	}
	return offset + col, nil
}

func (s *Server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.documents, params.TextDocument.URI)

	_ = s.client.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
		URI:         params.TextDocument.URI,
		Diagnostics: []protocol.Diagnostic{},
	})
	return nil
}

func (s *Server) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, ok := s.documents[params.TextDocument.URI]
	if ok {
		s.parseAndDiagnose(ctx, doc)
	}
	return nil
}

func (s *Server) DidChangeConfiguration(ctx context.Context, params *protocol.DidChangeConfigurationParams) error {
	return nil
}

func (s *Server) DidChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
	return nil
}
