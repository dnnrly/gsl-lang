package lsp

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/dnnrly/gsl-lang"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

type document struct {
	docURI  uri.URI
	content string
	version int32

	graph *gsl.Graph
	pErrs []protocol.Diagnostic
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
			doc.content = c.Text
		}
	}
	doc.version = params.TextDocument.Version
	s.parseAndDiagnose(ctx, doc)
	return nil
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

func (s *Server) parseAndDiagnose(ctx context.Context, doc *document) {
	graph, pErr := gsl.Parse(strings.NewReader(doc.content))
	doc.graph = graph

	var diags []protocol.Diagnostic

	if pErr != nil && pErr.HasError() {
		d := parseErrorToDiagnostic(pErr.Err.Error(), doc.content)
		if d != nil {
			diags = append(diags, *d)
		} else {
			diags = append(diags, protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: 0, Character: 0},
					End:   protocol.Position{Line: 0, Character: 0},
				},
				Severity: protocol.DiagnosticSeverityError,
				Source:   protocol.NewOptional("gsl"),
				Message:  protocol.String(pErr.Err.Error()),
			})
		}
	}

	if pErr != nil && pErr.HasWarnings() {
		for _, w := range pErr.Warnings {
			d := parseErrorToDiagnostic(w.Error(), doc.content)
			if d != nil {
				d.Severity = protocol.DiagnosticSeverityWarning
				diags = append(diags, *d)
			} else {
				diags = append(diags, protocol.Diagnostic{
					Range:    protocol.Range{Start: protocol.Position{Line: 0, Character: 0}, End: protocol.Position{Line: 0, Character: 0}},
					Severity: protocol.DiagnosticSeverityWarning,
					Source:   protocol.NewOptional("gsl"),
					Message:  protocol.String(w.Error()),
				})
			}
		}
	}

	doc.pErrs = diags

	_ = s.client.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
		URI:         doc.docURI,
		Version:     protocol.NewOptional(doc.version),
		Diagnostics: diags,
	})
}

var errPosRe = regexp.MustCompile(`at (\d+):(\d+)$`)

func parseErrorToDiagnostic(msg, content string) *protocol.Diagnostic {
	m := errPosRe.FindStringSubmatch(msg)
	if m == nil {
		return nil
	}

	line, _ := strconv.Atoi(m[1])
	col, _ := strconv.Atoi(m[2])

	start := protocol.Position{
		Line:      uint32(max(0, line-1)),
		Character: uint32(max(0, col-1)),
	}

	lines := strings.Split(content, "\n")
	end := start
	if int(start.Line) < len(lines) {
		lineStr := lines[start.Line]
		end.Character = uint32(len(lineStr))
	}

	cleanMsg := errPosRe.ReplaceAllString(msg, "")
	cleanMsg = strings.TrimSpace(cleanMsg)

	return &protocol.Diagnostic{
		Range:    protocol.Range{Start: start, End: end},
		Severity: protocol.DiagnosticSeverityError,
		Source:   protocol.NewOptional("gsl"),
		Message:  protocol.String(cleanMsg),
	}
}

func (s *Server) Completion(ctx context.Context, params *protocol.CompletionParams) (protocol.CompletionResult, error) {
	s.mu.Lock()
	doc, ok := s.documents[params.TextDocument.URI]
	s.mu.Unlock()
	if !ok {
		return protocol.CompletionItemSlice{}, nil
	}

	pos := params.Position
	items := s.complete(doc, int(pos.Line), int(pos.Character))
	return protocol.CompletionItemSlice(items), nil
}

func (s *Server) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	s.mu.Lock()
	doc, ok := s.documents[params.TextDocument.URI]
	s.mu.Unlock()
	if !ok || doc.graph == nil {
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
	if !ok {
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
	if !ok {
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

	tokens := tokenize(doc.content)
	symbols := buildSymbols(tokens)
	return protocol.DocumentSymbolSlice(symbols), nil
}

func (s *Server) Formatting(ctx context.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	s.mu.Lock()
	doc, ok := s.documents[params.TextDocument.URI]
	s.mu.Unlock()
	if !ok || doc.graph == nil {
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

	tokens := tokenize(doc.content)
	data := encodeSemanticTokens(tokens)
	return &protocol.SemanticTokens{Data: data}, nil
}

func (s *Server) DidChangeConfiguration(ctx context.Context, params *protocol.DidChangeConfigurationParams) error {
	return nil
}

func (s *Server) DidChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
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

func (s *Server) complete(doc *document, line, character int) []protocol.CompletionItem {
	if doc == nil {
		return nil
	}

	lines := strings.Split(doc.content, "\n")
	if line >= len(lines) {
		return nil
	}

	lineStr := lines[line]
	prefix := lineStr[:min(character, len(lineStr))]
	trimmed := strings.TrimSpace(prefix)

	tokens := tokenize(doc.content)
	ctx := contextAt(tokens, line, character)

	switch ctx {
	case contextKeyword:
		return keywordCompletions(trimmed)
	case contextNodeName:
		return nodeNameCompletions(doc)
	case contextSetName:
		return setNameCompletions(doc)
	case contextAttribute:
		return attributeCompletions(trimmed)
	case contextEdgeTarget:
		return nodeNameCompletions(doc)
	case contextValue:
		return valueCompletions(trimmed)
	default:
		return keywordCompletions(trimmed)
	}
}

type completionContext int

const (
	contextUnknown completionContext = iota
	contextKeyword
	contextNodeName
	contextSetName
	contextAttribute
	contextEdgeTarget
	contextValue
)

func contextAt(tokens []gsl.Token, line, character int) completionContext {
	var bestTok *gsl.Token
	bestIdx := -1

	for i, tok := range tokens {
		if tok.Type == gsl.TOKEN_EOF {
			continue
		}
		if int(tok.Line-1) != line {
			continue
		}
		if int(tok.Column-1) > character {
			continue
		}
		if bestTok == nil || int(tok.Column-1) >= int(bestTok.Column-1) {
			t := tok
			bestTok = &t
			bestIdx = i
		}
	}

	if bestTok == nil {
		return contextKeyword
	}

	switch bestTok.Type {
	case gsl.TOKEN_AT:
		return contextSetName
	case gsl.TOKEN_LBRACKET:
		return contextAttribute
	case gsl.TOKEN_COMMA, gsl.TOKEN_ARROW:
		return contextEdgeTarget
	case gsl.TOKEN_EQUALS:
		return contextValue
	case gsl.TOKEN_NODE:
		return contextNodeName
	case gsl.TOKEN_SET:
		return contextNodeName
	case gsl.TOKEN_COLON:
		if bestIdx > 0 && tokens[bestIdx-1].Type == gsl.TOKEN_IDENT {
			return contextEdgeTarget
		}
	case gsl.TOKEN_LBRACE:
		if bestIdx > 0 {
			prev := tokens[bestIdx-1]
			if prev.Type == gsl.TOKEN_RBRACKET || prev.Type == gsl.TOKEN_IDENT || prev.Type == gsl.TOKEN_STRING {
				return contextKeyword
			}
		}
	}

	return contextKeyword
}

func keywordCompletions(prefix string) []protocol.CompletionItem {
	keywords := []struct {
		label  string
		detail string
		kind   protocol.CompletionItemKind
	}{
		{"node", "Declare a node", protocol.CompletionItemKindKeyword},
		{"set", "Declare a set", protocol.CompletionItemKindKeyword},
		{"true", "Boolean true", protocol.CompletionItemKindKeyword},
		{"false", "Boolean false", protocol.CompletionItemKindKeyword},
	}

	var items []protocol.CompletionItem
	for _, kw := range keywords {
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

func nodeNameCompletions(doc *document) []protocol.CompletionItem {
	names := make(map[string]struct{})

	if doc.graph != nil {
		for name := range doc.graph.GetNodes() {
			names[name] = struct{}{}
		}
	}

	extractNodeNames(doc.content, names)

	var items []protocol.CompletionItem
	for name := range names {
		items = append(items, protocol.CompletionItem{
			Label:  name,
			Kind:   protocol.CompletionItemKindVariable,
			Detail: protocol.NewOptional("node"),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })
	return items
}

func extractNodeNames(content string, names map[string]struct{}) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "node ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				name := parts[1]
				if strings.HasPrefix(name, "[") || strings.HasPrefix(name, ":") {
					continue
				}
				names[name] = struct{}{}
			}
		}
		// Extract node names from edge declarations: "A -> B" or "A, B -> C"
		if strings.Contains(trimmed, "->") {
			fields := strings.Fields(trimmed)
			for _, f := range fields {
				if f == "->" {
					continue
				}
				if strings.HasPrefix(f, "@") || strings.HasPrefix(f, "[") || f == ":" {
					continue
				}
				// Clean trailing punctuation
				f = strings.TrimRight(f, ",;:")
				if isIdent(f) && !isKeyword(f) {
					names[f] = struct{}{}
				}
			}
		}
	}
}

func isIdent(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

func isKeyword(s string) bool {
	switch s {
	case "node", "set", "true", "false":
		return true
	}
	return false
}

func setNameCompletions(doc *document) []protocol.CompletionItem {
	names := make(map[string]struct{})

	if doc.graph != nil {
		for name := range doc.graph.GetSets() {
			names[name] = struct{}{}
		}
	}

	var items []protocol.CompletionItem
	for name := range names {
		items = append(items, protocol.CompletionItem{
			Label:  name,
			Kind:   protocol.CompletionItemKindVariable,
			Detail: protocol.NewOptional("set"),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })
	return items
}

func attributeCompletions(prefix string) []protocol.CompletionItem {
	known := []struct {
		label  string
		detail string
		kind   protocol.CompletionItemKind
	}{
		{"text", "Display text", protocol.CompletionItemKindProperty},
		{"parent", "Parent/reference", protocol.CompletionItemKindProperty},
		{"color", "Color attribute", protocol.CompletionItemKindProperty},
		{"weight", "Weight attribute", protocol.CompletionItemKindProperty},
		{"style", "Style attribute", protocol.CompletionItemKindProperty},
		{"shape", "Shape attribute", protocol.CompletionItemKindProperty},
	}

	var items []protocol.CompletionItem
	for _, a := range known {
		if prefix == "" || strings.HasPrefix(a.label, prefix) {
			items = append(items, protocol.CompletionItem{
				Label:  a.label,
				Kind:   a.kind,
				Detail: protocol.NewOptional(a.detail),
			})
		}
	}
	return items
}

func valueCompletions(prefix string) []protocol.CompletionItem {
	var items []protocol.CompletionItem
	if prefix == "" || strings.HasPrefix("true", prefix) {
		items = append(items, protocol.CompletionItem{
			Label:  "true",
			Kind:   protocol.CompletionItemKindKeyword,
			Detail: protocol.NewOptional("boolean"),
		})
	}
	if prefix == "" || strings.HasPrefix("false", prefix) {
		items = append(items, protocol.CompletionItem{
			Label:  "false",
			Kind:   protocol.CompletionItemKindKeyword,
			Detail: protocol.NewOptional("boolean"),
		})
	}
	return items
}

func (s *Server) hoverFor(graph *gsl.Graph, word string) string {
	if graph == nil {
		return ""
	}

	var parts []string

	if node := graph.GetNode(word); node != nil {
		parts = append(parts, fmt.Sprintf("**Node** `%s`", node.ID))
		if len(node.Attributes) > 0 {
			parts = append(parts, "\n\n**Attributes:**")
			keys := make([]string, 0, len(node.Attributes))
			for k := range node.Attributes {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				parts = append(parts, fmt.Sprintf("\n- `%s` = `%v`", k, node.Attributes[k]))
			}
		}
		if len(node.Sets) > 0 {
			setNames := make([]string, 0, len(node.Sets))
			for s := range node.Sets {
				setNames = append(setNames, s)
			}
			sort.Strings(setNames)
			parts = append(parts, "\n\n**Sets:**"+fmt.Sprintf(" @%s", strings.Join(setNames, ", @")))
		}
	}

	if sets := graph.GetSets(); sets != nil {
		if s, ok := sets[word]; ok {
			if len(parts) > 0 {
				parts = append(parts, "\n---\n")
			}
			parts = append(parts, fmt.Sprintf("**Set** `%s`", s.ID))
			if len(s.Attributes) > 0 {
				parts = append(parts, "\n\n**Attributes:**")
				keys := make([]string, 0, len(s.Attributes))
				for k := range s.Attributes {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					parts = append(parts, fmt.Sprintf("\n- `%s` = `%v`", k, s.Attributes[k]))
				}
			}
		}
	}

	for _, e := range graph.GetEdges() {
		if word == e.From || word == e.To || (e.Label != "" && word == e.Label) {
			if len(parts) > 0 {
				parts = append(parts, "\n---\n")
			}
			parts = append(parts, fmt.Sprintf("**Edge** `%s` → `%s`", e.From, e.To))
			if e.Label != "" {
				parts = append(parts, fmt.Sprintf("\n\nLabel: `%s`", e.Label))
			}
			if len(e.Attributes) > 0 {
				parts = append(parts, "\n\n**Attributes:**")
				keys := make([]string, 0, len(e.Attributes))
				for k := range e.Attributes {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					parts = append(parts, fmt.Sprintf("\n- `%s` = `%v`", k, e.Attributes[k]))
				}
			}
			break
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "")
}

func findDeclPosition(tokens []gsl.Token, name string, docURI uri.URI) *protocol.Location {
	for i, tok := range tokens {
		if tok.Literal != name || tok.Type != gsl.TOKEN_IDENT {
			continue
		}
		if i >= 1 {
			prev := tokens[i-1]
			if prev.Type == gsl.TOKEN_NODE || prev.Type == gsl.TOKEN_SET {
				pos := lineColToPos(tok)
				return &protocol.Location{
					URI: docURI,
					Range: protocol.Range{
						Start: pos,
						End:   pos,
					},
				}
			}
		}
	}
	return nil
}

func findReferences(tokens []gsl.Token, name string, docURI uri.URI) []protocol.Location {
	var locs []protocol.Location
	for _, tok := range tokens {
		if tok.Literal == name && tok.Type == gsl.TOKEN_IDENT {
			pos := lineColToPos(tok)
			locs = append(locs, protocol.Location{
				URI: docURI,
				Range: protocol.Range{
					Start: pos,
					End:   pos,
				},
			})
		}
	}
	return locs
}

func buildSymbols(tokens []gsl.Token) []protocol.DocumentSymbol {
	var symbols []protocol.DocumentSymbol

	for i, tok := range tokens {
		if tok.Type == gsl.TOKEN_EOF {
			break
		}

		if tok.Type == gsl.TOKEN_NODE && i+1 < len(tokens) && tokens[i+1].Type == gsl.TOKEN_IDENT {
			nameTok := tokens[i+1]
			endPos := lineColToPos(nameTok)
			endPos.Character += uint32(len(nameTok.Literal))

			symbols = append(symbols, protocol.DocumentSymbol{
				Name:           nameTok.Literal,
				Kind:           protocol.SymbolKindVariable,
				Range:          protocol.Range{Start: lineColToPos(tok), End: endPos},
				SelectionRange: protocol.Range{Start: lineColToPos(nameTok), End: endPos},
			})
		}

		if tok.Type == gsl.TOKEN_SET && i+1 < len(tokens) && tokens[i+1].Type == gsl.TOKEN_IDENT {
			nameTok := tokens[i+1]
			endPos := lineColToPos(nameTok)
			endPos.Character += uint32(len(nameTok.Literal))

			symbols = append(symbols, protocol.DocumentSymbol{
				Name:           nameTok.Literal,
				Kind:           protocol.SymbolKindNamespace,
				Range:          protocol.Range{Start: lineColToPos(tok), End: endPos},
				SelectionRange: protocol.Range{Start: lineColToPos(nameTok), End: endPos},
			})
		}
	}

	return symbols
}

func semanticTokenTypes() []string {
	return []string{
		"keyword",
		"string",
		"number",
		"boolean",
		"type",
		"property",
		"operator",
		"comment",
	}
}

func tokenTypeToSemanticIndex(tt gsl.TokenType) uint32 {
	switch tt {
	case gsl.TOKEN_NODE, gsl.TOKEN_SET:
		return 0
	case gsl.TOKEN_TRUE, gsl.TOKEN_FALSE:
		return 3
	case gsl.TOKEN_STRING:
		return 1
	case gsl.TOKEN_NUMBER:
		return 2
	case gsl.TOKEN_IDENT:
		return 4
	case gsl.TOKEN_EQUALS, gsl.TOKEN_ARROW:
		return 6
	case gsl.TOKEN_AT, gsl.TOKEN_COLON, gsl.TOKEN_COMMA:
		return 6
	case gsl.TOKEN_LBRACKET, gsl.TOKEN_RBRACKET, gsl.TOKEN_LBRACE, gsl.TOKEN_RBRACE:
		return 6
	default:
		return 0
	}
}

func encodeSemanticTokens(tokens []gsl.Token) []uint32 {
	if len(tokens) == 0 {
		return []uint32{}
	}

	var data []uint32
	prevLine := uint32(0)
	prevChar := uint32(0)

	hasCommentPrefix := false
	commentLine := uint32(0)
	commentChar := uint32(0)

	for _, tok := range tokens {
		if tok.Type == gsl.TOKEN_EOF {
			break
		}

		pos := lineColToPos(tok)

		if tok.Type == gsl.TOKEN_ILLEGAL {
			if tok.Literal == "#" {
				hasCommentPrefix = true
				commentLine = pos.Line
				commentChar = pos.Character
				continue
			}
			if hasCommentPrefix && pos.Line == commentLine && pos.Character >= commentChar {
				continue
			}
			hasCommentPrefix = false
		} else {
			if hasCommentPrefix {
				if pos.Line == commentLine {
					continue
				}
				hasCommentPrefix = false
			}
		}

		tokLen := uint32(len(tok.Literal))

		typeIdx := tokenTypeToSemanticIndex(tok.Type)

		deltaLine := pos.Line - prevLine
		deltaChar := pos.Character
		if deltaLine == 0 {
			deltaChar = pos.Character - prevChar
		}

		data = append(data, deltaLine, deltaChar, tokLen, typeIdx, 0)
		prevLine = pos.Line
		prevChar = pos.Character
	}

	return data
}

var (
	_ = fmt.Sprintf
	_ = regexp.Compile
)
