package lsp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestGQLDidOpen_ValidQuery(t *testing.T) {
	s, mc := newTestServer()
	docURI := uri.File("/tmp/test.gql")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: `subgraph edge.timeout exists`,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	diags := mc.diags[docURI]
	for _, d := range diags {
		t.Logf("diagnostic: %v - %s", d.Range, d.Message)
	}
	if len(diags) > 0 {
		t.Fatalf("expected no diagnostics for valid query, got %d", len(diags))
	}

	doc, ok := s.documents[docURI]
	if !ok {
		t.Fatal("document not found")
	}
	if doc.gqlQuery == nil {
		t.Fatal("expected gqlQuery to be set")
	}
	if doc.graph != nil {
		t.Fatal("expected graph to be nil for GQL file")
	}
	if doc.gqlErr != nil {
		t.Fatalf("expected no gqlErr, got %v", doc.gqlErr)
	}
}

func TestGQLDidOpen_InvalidQuery(t *testing.T) {
	s, mc := newTestServer()
	docURI := uri.File("/tmp/test.gql")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: `invalid syntax here`,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	diags := mc.diags[docURI]
	if len(diags) == 0 {
		t.Fatal("expected diagnostics for invalid query")
	}
	t.Logf("got %d diagnostics", len(diags))
	for _, d := range diags {
		t.Logf("  %v: %s", d.Range, d.Message)
	}

	doc, ok := s.documents[docURI]
	if !ok {
		t.Fatal("document not found")
	}
	if doc.gqlQuery != nil {
		t.Fatal("expected gqlQuery to be nil for invalid input")
	}
	if doc.gqlErr == nil {
		t.Fatal("expected gqlErr to be set")
	}
}

func TestGQLDidOpen_EmptyQuery(t *testing.T) {
	s, mc := newTestServer()
	docURI := uri.File("/tmp/test.gql")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: "",
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	diags := mc.diags[docURI]
	if len(diags) > 0 {
		t.Fatalf("expected no diagnostics for empty query, got %d", len(diags))
	}
}

func TestGQLDidOpen_MultilineQuery(t *testing.T) {
	s, mc := newTestServer()
	docURI := uri.File("/tmp/test.gql")

	content := `subgraph node.team == "payments"
| from *
| (subgraph node.zone == "b") as ZONES
| ZONES + *`

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: content,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	diags := mc.diags[docURI]
	if len(diags) > 0 {
		t.Fatalf("expected no diagnostics, got %d: %v", len(diags), diags)
	}
}

func TestGQLDidChange_PreservesGQLParsing(t *testing.T) {
	s, mc := newTestServer()
	docURI := uri.File("/tmp/test.gql")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: `subgraph edge.timeout exists`,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	if len(mc.diags[docURI]) > 0 {
		t.Fatal("expected no diagnostics after open")
	}

	err = s.DidChange(context.Background(), &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: docURI},
			Version:                2,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			&protocol.TextDocumentContentChangeWholeDocument{
				Text: `subgraph node.team == "eng"`,
			},
		},
	})
	if err != nil {
		t.Fatalf("DidChange failed: %v", err)
	}

	if len(mc.diags[docURI]) > 0 {
		t.Fatalf("expected no diagnostics after change, got %d", len(mc.diags[docURI]))
	}
}

func TestGQLCompletion_Keywords(t *testing.T) {
	s, _ := newTestServer()
	docURI := uri.File("/tmp/test.gql")

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

	hasSubgraph := false
	hasMake := false
	hasRemove := false
	hasCollapse := false
	hasFrom := false
	for _, item := range items {
		switch item.Label {
		case "subgraph":
			hasSubgraph = true
		case "make":
			hasMake = true
		case "remove":
			hasRemove = true
		case "collapse":
			hasCollapse = true
		case "from":
			hasFrom = true
		}
	}
	if !hasSubgraph {
		t.Fatal("expected 'subgraph' keyword")
	}
	if !hasMake {
		t.Fatal("expected 'make' keyword")
	}
	if !hasRemove {
		t.Fatal("expected 'remove' keyword")
	}
	if !hasCollapse {
		t.Fatal("expected 'collapse' keyword")
	}
	if !hasFrom {
		t.Fatal("expected 'from' keyword")
	}
}

func TestGQLCompletion_FiltersByPrefix(t *testing.T) {
	s, _ := newTestServer()
	docURI := uri.File("/tmp/test.gql")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: "su",
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	result, err := s.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 0, Character: 2},
		},
	})
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	items, ok := result.(protocol.CompletionItemSlice)
	if !ok {
		t.Fatalf("expected CompletionItemSlice, got %T", result)
	}

	for _, item := range items {
		if !strings.HasPrefix(item.Label, "su") {
			t.Fatalf("expected all completions to start with 'su', got %q", item.Label)
		}
	}
}

func TestGQLHover_KnownKeyword(t *testing.T) {
	s, _ := newTestServer()
	docURI := uri.File("/tmp/test.gql")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: `subgraph node.team == "eng"`,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	hover, err := s.Hover(context.Background(), &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 0, Character: 2},
		},
	})
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}
	if hover == nil {
		t.Fatal("expected hover result for 'subgraph' keyword")
	}

	content, ok := hover.Contents.(*protocol.MarkupContent)
	if !ok {
		t.Fatalf("expected MarkupContent, got %T", hover.Contents)
	}
	if !strings.Contains(content.Value, "Filter nodes/edges by predicate") {
		t.Fatalf("expected hover to describe 'subgraph', got %q", content.Value)
	}
}

func TestGQLHover_NonKeyword(t *testing.T) {
	s, _ := newTestServer()
	docURI := uri.File("/tmp/test.gql")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: `subgraph foo.bar == "eng"`,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	hover, err := s.Hover(context.Background(), &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 0, Character: 10}},
	})
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}
	if hover != nil {
		t.Fatal("expected nil hover for non-keyword identifier")
	}
}

func TestGQLDocumentSymbols_Pipeline(t *testing.T) {
	s, _ := newTestServer()
	docURI := uri.File("/tmp/test.gql")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: `subgraph node.team == "eng" | make node.reviewed = true where node.team == "eng"`,
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

	if len(symbols) < 2 {
		t.Fatalf("expected at least 2 symbols, got %d", len(symbols))
	}

	hasSubgraph := false
	hasMake := false
	for _, s := range symbols {
		if s.Name == "subgraph" {
			hasSubgraph = true
		}
		if s.Name == "make" {
			hasMake = true
		}
	}
	if !hasSubgraph {
		t.Fatal("expected 'subgraph' symbol")
	}
	if !hasMake {
		t.Fatal("expected 'make' symbol")
	}
}

func TestGQLDocumentSymbols_RemoveOrphans(t *testing.T) {
	s, _ := newTestServer()
	docURI := uri.File("/tmp/test.gql")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: `remove orphans`,
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

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}
	if symbols[0].Name != "remove" {
		t.Fatalf("expected 'remove' symbol, got %q", symbols[0].Name)
	}
}

func TestGQLSemanticTokens_BasicQuery(t *testing.T) {
	s, _ := newTestServer()
	docURI := uri.File("/tmp/test.gql")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: `subgraph edge.timeout exists`,
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

	// Tokens are [deltaLine, deltaChar, length, typeIdx, modifiers] quintuples
	if len(tokens.Data)%5 != 0 {
		t.Fatalf("token data length must be multiple of 5, got %d", len(tokens.Data))
	}
}

func TestGQLSemanticTokens_StringAndNumber(t *testing.T) {
	s, _ := newTestServer()
	docURI := uri.File("/tmp/test.gql")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: `subgraph node.x == "hello" AND node.y == 42`,
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

func TestGQLDefinition_Unsupported(t *testing.T) {
	s, _ := newTestServer()
	docURI := uri.File("/tmp/test.gql")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: `subgraph node.team == "eng"`,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	def, err := s.Definition(context.Background(), &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 0, Character: 0},
		},
	})
	if err != nil {
		t.Fatalf("Definition failed: %v", err)
	}
	if def != nil {
		t.Fatal("expected nil definition for GQL file")
	}
}

func TestGQLReferences_Unsupported(t *testing.T) {
	s, _ := newTestServer()
	docURI := uri.File("/tmp/test.gql")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: `subgraph node.team == "eng"`,
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
	if len(refs) != 0 {
		t.Fatal("expected empty references for GQL file")
	}
}

func TestGQLFormatting_Unsupported(t *testing.T) {
	s, _ := newTestServer()
	docURI := uri.File("/tmp/test.gql")

	err := s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  docURI,
			Text: `subgraph node.team == "eng"`,
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
	if edits != nil {
		t.Fatal("expected nil formatting edits for GQL file")
	}
}

// TestGQLAllFixturesValid loads every query.gql in testdata and verifies
// the LSP reports zero diagnostics (all fixture queries are valid GQL).
func findQueryFixtures() []string {
	// Test may run from package dir (lsp/) or module root
	roots := []string{"../query/testdata", "query/testdata"}
	for _, root := range roots {
		var matches []string
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() && filepath.Base(path) == "query.gql" {
				matches = append(matches, path)
			}
			return nil
		})
		if err == nil && len(matches) > 0 {
			return matches
		}
	}
	return nil
}

func TestGQLAllFixturesValid(t *testing.T) {
	matches := findQueryFixtures()
	if len(matches) == 0 {
		t.Fatal("no query.gql fixtures found")
	}

	seen := make(map[string]bool)
	for _, path := range matches {
		if seen[path] {
			continue
		}
		seen[path] = true

		t.Run(filepath.Base(filepath.Dir(path)), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile failed: %v", err)
			}
			content := string(data)

			s, mc := newTestServer()
			docURI := uri.File(path)

			err = s.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
				TextDocument: protocol.TextDocumentItem{
					URI:  docURI,
					Text: content,
				},
			})
			if err != nil {
				t.Fatalf("DidOpen failed: %v", err)
			}

			diags := mc.diags[docURI]
			if len(diags) > 0 {
				for _, d := range diags {
					t.Logf("  diagnostic: %v %s", d.Range, d.Message)
				}
				t.Fatalf("expected no diagnostics for %q, got %d", path, len(diags))
			}
		})
	}
}

// TestGSLStillParsesAsGSL confirms GSL files are not accidentally
// caught by the GQL dispatch path.
func TestGSLStillParsesAsGSL(t *testing.T) {
	s, mc := newTestServer()
	docURI := uri.File("/tmp/test.gsl")

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
	if len(diags) > 0 {
		t.Fatalf("expected no diagnostics, got %d", len(diags))
	}

	doc, ok := s.documents[docURI]
	if !ok {
		t.Fatal("document not found")
	}
	if doc.graph == nil {
		t.Fatal("expected graph to be set for GSL file")
	}
	if doc.gqlQuery != nil {
		t.Fatal("expected gqlQuery to be nil for GSL file")
	}
}

func TestGSLCompletion_NotGQL(t *testing.T) {
	s, _ := newTestServer()
	docURI := uri.File("/tmp/test.gsl")

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
	hasSubgraph := false
	for _, item := range items {
		switch item.Label {
		case "node":
			hasNode = true
		case "set":
			hasSet = true
		case "subgraph":
			hasSubgraph = true
		}
	}
	if !hasNode || !hasSet {
		t.Fatal("expected GSL keywords 'node' and 'set'")
	}
	if hasSubgraph {
		t.Fatal("GSL completions should not include GQL-only keyword 'subgraph'")
	}
}
