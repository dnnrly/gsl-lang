package lsp

import (
	"context"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/query"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func isGQLFile(docURI uri.URI) bool {
	return strings.HasSuffix(strings.ToLower(string(docURI)), ".gql")
}

func (s *Server) parseAndDiagnose(ctx context.Context, doc *document) {
	if isGQLFile(doc.docURI) {
		log.Printf("gsl-lsp: parsing GQL file %s", doc.docURI)
		s.parseGQLAndDiagnose(ctx, doc)
		return
	}
	log.Printf("gsl-lsp: parsing GSL file %s", doc.docURI)

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

func (s *Server) parseGQLAndDiagnose(ctx context.Context, doc *document) {
	parser := query.NewQueryParser(doc.content)
	q, err := parser.Parse()
	doc.gqlQuery = q
	doc.gqlErr = err

	var diags []protocol.Diagnostic
	if err != nil {
		d := gqlParseErrorToDiagnostic(err, doc.content)
		if d != nil {
			diags = append(diags, *d)
		} else {
			diags = append(diags, protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: 0, Character: 0},
					End:   protocol.Position{Line: 0, Character: 0},
				},
				Severity: protocol.DiagnosticSeverityError,
				Source:   protocol.NewOptional("gql"),
				Message:  protocol.String(err.Error()),
			})
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
var gqlErrPosRe = regexp.MustCompile(`(\d+):(\d+):`)

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

func gqlParseErrorToDiagnostic(err error, content string) *protocol.Diagnostic {
	msg := err.Error()
	m := gqlErrPosRe.FindStringSubmatch(msg)
	if m != nil {
		line, _ := strconv.Atoi(m[1])
		col, _ := strconv.Atoi(m[2])
		start := protocol.Position{
			Line:      uint32(max(0, line-1)),
			Character: uint32(max(0, col-1)),
		}
		lines := strings.Split(content, "\n")
		end := start
		if int(start.Line) < len(lines) {
			end.Character = uint32(len(lines[start.Line]))
		}

		cleanMsg := gqlErrPosRe.ReplaceAllString(msg, "")
		cleanMsg = strings.TrimSpace(cleanMsg)

		return &protocol.Diagnostic{
			Range:    protocol.Range{Start: start, End: end},
			Severity: protocol.DiagnosticSeverityError,
			Source:   protocol.NewOptional("gql"),
			Message:  protocol.String(cleanMsg),
		}
	}

	return nil
}
