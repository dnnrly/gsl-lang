package lsp

import (
	"context"
	"testing"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

type mockClient struct {
	diags map[uri.URI][]protocol.Diagnostic
}

func (m *mockClient) PublishDiagnostics(ctx context.Context, params *protocol.PublishDiagnosticsParams) error {
	if m.diags == nil {
		m.diags = make(map[uri.URI][]protocol.Diagnostic)
	}
	m.diags[params.URI] = params.Diagnostics
	return nil
}

func (m *mockClient) Progress(ctx context.Context, params *protocol.ProgressParams) error { return nil }
func (m *mockClient) LogTrace(ctx context.Context, params *protocol.LogTraceParams) error { return nil }
func (m *mockClient) RegisterCapability(ctx context.Context, params *protocol.RegistrationParams) error {
	return nil
}
func (m *mockClient) UnregisterCapability(ctx context.Context, params *protocol.UnregistrationParams) error {
	return nil
}
func (m *mockClient) ShowMessage(ctx context.Context, params *protocol.ShowMessageParams) error {
	return nil
}
func (m *mockClient) ShowMessageRequest(ctx context.Context, params *protocol.ShowMessageRequestParams) (*protocol.MessageActionItem, error) {
	return nil, nil
}
func (m *mockClient) LogMessage(ctx context.Context, params *protocol.LogMessageParams) error {
	return nil
}
func (m *mockClient) ShowDocument(ctx context.Context, params *protocol.ShowDocumentParams) (*protocol.ShowDocumentResult, error) {
	return nil, nil
}
func (m *mockClient) WorkDoneProgressCreate(ctx context.Context, params *protocol.WorkDoneProgressCreateParams) error {
	return nil
}
func (m *mockClient) Telemetry(ctx context.Context, params protocol.LSPAny) error { return nil }
func (m *mockClient) Configuration(ctx context.Context, params *protocol.ConfigurationParams) ([]protocol.LSPAny, error) {
	return nil, nil
}
func (m *mockClient) WorkspaceFolders(ctx context.Context) ([]protocol.WorkspaceFolder, error) {
	return nil, nil
}
func (m *mockClient) ApplyEdit(ctx context.Context, params *protocol.ApplyWorkspaceEditParams) (*protocol.ApplyWorkspaceEditResult, error) {
	return nil, nil
}
func (m *mockClient) CodeLensRefresh(ctx context.Context) error       { return nil }
func (m *mockClient) FoldingRangeRefresh(ctx context.Context) error   { return nil }
func (m *mockClient) SemanticTokensRefresh(ctx context.Context) error { return nil }
func (m *mockClient) InlineValueRefresh(ctx context.Context) error    { return nil }
func (m *mockClient) InlayHintRefresh(ctx context.Context) error      { return nil }
func (m *mockClient) DiagnosticRefresh(ctx context.Context) error     { return nil }
func (m *mockClient) TextDocumentContentRefresh(ctx context.Context, params *protocol.TextDocumentContentRefreshParams) error {
	return nil
}

func newTestServer() (*Server, *mockClient) {
	mc := &mockClient{}
	s := NewServer(mc)
	return s, mc
}

func mustURI(s string) uri.URI {
	u, _ := uri.Parse(s)
	return u
}

func TestInitialize(t *testing.T) {
	s, _ := newTestServer()
	result, err := s.Initialize(context.Background(), &protocol.InitializeParams{})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	if result.Capabilities.TextDocumentSync == nil {
		t.Fatal("expected TextDocumentSync capabilities")
	}
	if result.Capabilities.CompletionProvider == nil {
		t.Fatal("expected CompletionProvider capabilities")
	}
	if result.Capabilities.HoverProvider == nil {
		t.Fatal("expected HoverProvider capabilities")
	}
	if result.ServerInfo.Name != "gsl-lsp" {
		t.Fatalf("expected server name gsl-lsp, got %s", result.ServerInfo.Name)
	}
}

func TestDidOpen_ValidDocument(t *testing.T) {
	s, mc := newTestServer()
	docURI := mustURI("file:///test.gsl")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: "node A\nnode B\nA -> B",
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	diags := mc.diags[docURI]
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for valid document, got %d: %v", len(diags), diags)
	}
}

func TestDidOpen_InvalidDocument(t *testing.T) {
	s, mc := newTestServer()
	docURI := mustURI("file:///test.gsl")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: "node A\nA -> \n",
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	diags := mc.diags[docURI]
	if len(diags) == 0 {
		t.Fatal("expected diagnostics for invalid document")
	}

	found := false
	for _, d := range diags {
		if d.Severity == protocol.DiagnosticSeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected at least one error diagnostic")
	}
}

func TestDidOpen_DocumentWithWarning(t *testing.T) {
	s, mc := newTestServer()
	docURI := mustURI("file:///test.gsl")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: "node A @implicitSet",
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	diags := mc.diags[docURI]
	found := false
	for _, d := range diags {
		if d.Severity == protocol.DiagnosticSeverityWarning {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected at least one warning diagnostic for implicit set")
	}
}

func TestDidClose(t *testing.T) {
	s, mc := newTestServer()
	docURI := mustURI("file:///test.gsl")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: "node A",
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	err = s.DidClose(context.Background(), &protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: docURI,
		},
	})
	if err != nil {
		t.Fatalf("DidClose failed: %v", err)
	}

	diags := mc.diags[docURI]
	if len(diags) != 0 {
		t.Fatal("expected empty diagnostics after close")
	}
}

func TestCompletion_Keywords(t *testing.T) {
	s, _ := newTestServer()
	docURI := mustURI("file:///test.gsl")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: "",
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 0, Character: 0},
		},
	})
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	items, ok := result.(protocol.CompletionItemSlice)
	if !ok {
		t.Fatalf("expected CompletionItemSlice, got %T", result)
	}

	hasNode := false
	hasSet := false
	for _, item := range items {
		if item.Label == "node" {
			hasNode = true
		}
		if item.Label == "set" {
			hasSet = true
		}
	}
	if !hasNode || !hasSet {
		t.Fatal("expected 'node' and 'set' completion items")
	}
}

func TestCompletion_NodeNames(t *testing.T) {
	s, _ := newTestServer()
	docURI := mustURI("file:///test.gsl")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: "node Alpha\nnode Beta\nA -> ",
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 2, Character: 5},
		},
	})
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	items, ok := result.(protocol.CompletionItemSlice)
	if !ok {
		t.Fatalf("expected CompletionItemSlice, got %T", result)
	}

	hasAlpha := false
	hasBeta := false
	for _, item := range items {
		if item.Label == "Alpha" {
			hasAlpha = true
		}
		if item.Label == "Beta" {
			hasBeta = true
		}
	}
	if !hasAlpha || !hasBeta {
		t.Fatal("expected node name completions for Alpha and Beta")
	}
}

func TestHover_Node(t *testing.T) {
	s, _ := newTestServer()
	docURI := mustURI("file:///test.gsl")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: "node A [color=red, weight=2]\nnode B",
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	hover, err := s.Hover(context.Background(), &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 0, Character: 5},
		},
	})
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}
	if hover == nil {
		t.Fatal("expected hover result for node A")
	}

	content, ok := hover.Contents.(*protocol.MarkupContent)
	if !ok {
		t.Fatalf("expected MarkupContent, got %T", hover.Contents)
	}

	if content.Value == "" {
		t.Fatal("expected non-empty hover content")
	}
}

func TestDefinition_NodeDeclaration(t *testing.T) {
	s, _ := newTestServer()
	docURI := mustURI("file:///test.gsl")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: "node A\nnode B\nA -> B",
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	def, err := s.Definition(context.Background(), &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 2, Character: 0},
		},
	})
	if err != nil {
		t.Fatalf("Definition failed: %v", err)
	}
	if def == nil {
		t.Fatal("expected definition result")
	}
}

func TestDocumentSymbols(t *testing.T) {
	s, _ := newTestServer()
	docURI := mustURI("file:///test.gsl")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: "node A\nset myset\nnode B\nA -> B\nnode C",
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	result, err := s.DocumentSymbol(context.Background(), &protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
	})
	if err != nil {
		t.Fatalf("DocumentSymbol failed: %v", err)
	}

	symbols, ok := result.(protocol.DocumentSymbolSlice)
	if !ok {
		t.Fatalf("expected DocumentSymbolSlice, got %T", result)
	}

	if len(symbols) < 3 {
		t.Fatalf("expected at least 3 symbols, got %d", len(symbols))
	}
}

func TestFormatting(t *testing.T) {
	s, _ := newTestServer()
	docURI := mustURI("file:///test.gsl")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: "node B\nnode A\n",
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	edits, err := s.Formatting(context.Background(), &protocol.DocumentFormattingParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
	})
	if err != nil {
		t.Fatalf("Formatting failed: %v", err)
	}

	if len(edits) == 0 {
		t.Fatal("expected formatting edits")
	}
}

func TestSemanticTokens(t *testing.T) {
	s, _ := newTestServer()
	docURI := mustURI("file:///test.gsl")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: "node A",
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	tokens, err := s.SemanticTokensFull(context.Background(), &protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
	})
	if err != nil {
		t.Fatalf("SemanticTokensFull failed: %v", err)
	}

	if len(tokens.Data) == 0 {
		t.Fatal("expected semantic token data")
	}
}

func TestReferences(t *testing.T) {
	s, _ := newTestServer()
	docURI := mustURI("file:///test.gsl")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: "node A\nnode B\nA -> B",
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	refs, err := s.References(context.Background(), &protocol.ReferenceParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 0, Character: 5},
		},
	})
	if err != nil {
		t.Fatalf("References failed: %v", err)
	}

	if len(refs) == 0 {
		t.Fatal("expected at least one reference")
	}
}

func TestWordAt(t *testing.T) {
	tests := []struct {
		content   string
		line      int
		character int
		expected  string
	}{
		{"node A", 0, 5, "A"},
		{"node A", 0, 6, "A"},
		{"A -> B", 0, 0, "A"},
		{"A -> B", 0, 5, "B"},
		{"@myset", 0, 3, "myset"},
		{"hello world", 0, 0, "hello"},
		{"  ", 0, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := wordAt(tt.content, tt.line, tt.character)
			if tt.expected == "" {
				if result != nil {
					t.Fatalf("expected nil, got %v", result)
				}
				return
			}
			if result == nil {
				t.Fatalf("expected word %q, got nil", tt.expected)
			}
			if result.word != tt.expected {
				t.Fatalf("expected word %q, got %q", tt.expected, result.word)
			}
		})
	}
}

func TestContextAt(t *testing.T) {
	tests := []struct {
		content   string
		line      int
		character int
		expected  completionContext
	}{
		{"", 0, 0, contextKeyword},
		{"node", 0, 4, contextNodeName},
		{"set", 0, 3, contextNodeName},
		{"A @", 0, 3, contextSetName},
		{"A [", 0, 2, contextAttribute},
		{"A ->", 0, 4, contextEdgeTarget},
		{"A -> B [x=", 0, 9, contextValue},
	}

	for _, tt := range tests {
		tokens := tokenize(tt.content)
		got := contextAt(tokens, tt.line, tt.character)
		if got != tt.expected {
			t.Errorf("contextAt(%q, %d, %d) = %d, want %d", tt.content, tt.line, tt.character, got, tt.expected)
		}
	}
}

func TestParseErrorToDiagnostic(t *testing.T) {
	content := "node A\nnode B [color=\n"
	msg := "expected value, got EOF () at 2:16"

	d := parseErrorToDiagnostic(msg, content)
	if d == nil {
		t.Fatal("expected diagnostic")
	}

	if d.Range.Start.Line != 1 {
		t.Fatalf("expected line 1 (0-based), got %d", d.Range.Start.Line)
	}
	if d.Range.Start.Character != 15 {
		t.Fatalf("expected character 15 (0-based), got %d", d.Range.Start.Character)
	}
}
