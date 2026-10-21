// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

package dialect

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	gsl "github.com/dnnrly/gsl-lang/implementations/go"
)

func mustGraph(t *testing.T, src string) *gsl.Graph {
	t.Helper()
	g, perr := gsl.Parse(strings.NewReader(src))
	if perr != nil && perr.HasError() {
		t.Fatalf("failed to parse GSL fixture: %v", perr)
	}
	return g
}

func mustLoad(t *testing.T, src string) *Dialect {
	t.Helper()
	d, err := Load(strings.NewReader(src))
	if err != nil {
		t.Fatalf("failed to load dialect: %v", err)
	}
	return d
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(b)
}

func TestLoadWorkedExample(t *testing.T) {
	d := mustLoad(t, readFile(t, "testdata/software-architecture.dialect.yaml"))

	if d.ID != "software-architecture" {
		t.Errorf("ID = %q", d.ID)
	}
	if d.Version != "0.1.0-draft" {
		t.Errorf("Version = %q", d.Version)
	}
	if d.Title != "Software Architecture" {
		t.Errorf("Title = %q", d.Title)
	}

	els := d.Elements()
	if len(els) != 1 {
		t.Fatalf("expected 1 element definition, got %d", len(els))
	}
	el := els[0]
	if el.Name != "Service" {
		t.Errorf("Name = %q", el.Name)
	}
	if el.Representation != RepresentationNode {
		t.Errorf("Representation = %q", el.Representation)
	}
	if el.Query != `kind == "service"` {
		t.Errorf("Query = %q", el.Query)
	}
}

// TestProjectionWorkedExample checks the §8.4 projection example exactly:
// node references, Empty values, numbers, booleans and strings are all
// distinguishable after projection.
func TestProjectionWorkedExample(t *testing.T) {
	g := mustGraph(t, readFile(t, "testdata/gateway.gsl"))
	gateway := g.GetNodes()["gateway"]
	if gateway == nil {
		t.Fatal("gateway node missing from fixture")
	}

	got, err := projectAttributes(gateway.Attributes)
	if err != nil {
		t.Fatalf("projectAttributes: %v", err)
	}

	want := map[string]any{
		"kind":        "service",
		"name":        "Gateway",
		"owner":       map[string]any{"$gslRef": "platform_team"},
		"replicas":    float64(3),
		"enabled":     map[string]any{"$gslEmpty": true},
		"description": "Public entry point",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("projection mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

// TestMatchIsIndependentOfValidation checks that Match returns selected
// elements with projected attributes without consulting the JSON Schema.
func TestMatchIsIndependentOfValidation(t *testing.T) {
	d := mustLoad(t, `
id: m
version: "1"
elements:
  Service:
    representation: node
    match:
      query: 'kind == "service"'
    attributes:
      type: object
      properties:
        name:
          type: string
`)
	g := mustGraph(t, `
node platform_team [kind="team"]
node gateway [kind="service", name="Gateway", owner=platform_team, replicas=3]
`)

	selected, err := d.Match(g)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if len(selected) != 1 {
		t.Fatalf("expected 1 selected element, got %d: %+v", len(selected), selected)
	}
	got := selected[0]
	if got.Definition != "Service" || got.Representation != RepresentationNode || got.Element != "gateway" {
		t.Errorf("unexpected selection: %+v", got)
	}
	want := map[string]any{
		"kind":     "service",
		"name":     "Gateway",
		"owner":    map[string]any{"$gslRef": "platform_team"},
		"replicas": float64(3),
	}
	if !reflect.DeepEqual(got.Attributes, want) {
		t.Errorf("attributes mismatch:\n got: %#v\nwant: %#v", got.Attributes, want)
	}
}

func TestValidateWorkedExampleConforming(t *testing.T) {
	d := mustLoad(t, readFile(t, "testdata/software-architecture.dialect.yaml"))
	g := mustGraph(t, readFile(t, "testdata/gateway.gsl"))

	report, err := d.Validate(g)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !report.Conforms() {
		t.Fatalf("expected conforming, got %s: %v", report.Outcome, report.Violations)
	}
}

func TestValidateMissingRequiredAttribute(t *testing.T) {
	d := mustLoad(t, readFile(t, "testdata/software-architecture.dialect.yaml"))
	g := mustGraph(t, `
node platform_team [kind="team"]
node gateway [kind="service", owner=platform_team, replicas=3]
`)

	report, err := d.Validate(g)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if report.Outcome != OutcomeNonConforming {
		t.Fatalf("expected non-conforming, got %s", report.Outcome)
	}
	if len(report.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(report.Violations), report.Violations)
	}
	v := report.Violations[0]
	if v.Definition != "Service" || v.Element != "gateway" {
		t.Errorf("unexpected violation target: %+v", v)
	}
	if !strings.Contains(v.Message, "name") {
		t.Errorf("violation message should mention the missing property: %q", v.Message)
	}
}

func TestValidateNumericMinimum(t *testing.T) {
	d := mustLoad(t, readFile(t, "testdata/software-architecture.dialect.yaml"))
	g := mustGraph(t, `
node platform_team [kind="team"]
node gateway [kind="service", name="Gateway", owner=platform_team, replicas=0]
`)

	report, err := d.Validate(g)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if report.Outcome != OutcomeNonConforming {
		t.Fatalf("expected non-conforming, got %s", report.Outcome)
	}
	if len(report.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(report.Violations), report.Violations)
	}
	if got := report.Violations[0].InstanceLocation; got != "/replicas" {
		t.Errorf("InstanceLocation = %q, want /replicas", got)
	}
}

// TestValidateOwnerNotAReference verifies that a plain string owner fails the
// node-reference schema: the projection preserves the distinction between a
// string and a reference.
func TestValidateOwnerNotAReference(t *testing.T) {
	d := mustLoad(t, readFile(t, "testdata/software-architecture.dialect.yaml"))
	g := mustGraph(t, `
node gateway [kind="service", name="Gateway", owner="platform", replicas=3]
`)

	report, err := d.Validate(g)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if report.Outcome != OutcomeNonConforming {
		t.Fatalf("expected non-conforming, got %s", report.Outcome)
	}
	var found bool
	for _, v := range report.Violations {
		if v.InstanceLocation == "/owner" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected an /owner violation, got %v", report.Violations)
	}
}

func TestValidateEmptyMatchConforming(t *testing.T) {
	d := mustLoad(t, readFile(t, "testdata/software-architecture.dialect.yaml"))
	g := mustGraph(t, `
node a [kind="database"]
node b [kind="queue"]
a -> b
`)

	report, err := d.Validate(g)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !report.Conforms() {
		t.Fatalf("empty match results must not produce violations, got %v", report.Violations)
	}
}

func TestValidateEmptyElementsConforming(t *testing.T) {
	d := mustLoad(t, `
id: empty
version: "1"
elements: {}
`)
	report, err := d.Validate(mustGraph(t, "node a [x=1]"))
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !report.Conforms() {
		t.Fatalf("dialect with no elements must conform, got %v", report.Violations)
	}
}

func TestUnknownFieldsIgnored(t *testing.T) {
	d := mustLoad(t, `
id: tolerant
version: "1"
something_extra: true
elements:
  Thing:
    representation: node
    future_field: whatever
    match:
      query: 'kind == "thing"'
    attributes:
      type: object
`)
	if len(d.Elements()) != 1 {
		t.Fatalf("expected 1 element, got %d", len(d.Elements()))
	}
}

func TestInvalidDialectRejected(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		wantPath string
	}{
		{
			name:     "missing id",
			yaml:     "version: \"1\"\nelements: {}\n",
			wantPath: "id",
		},
		{
			name:     "empty id",
			yaml:     "id: \"\"\nversion: \"1\"\nelements: {}\n",
			wantPath: "id",
		},
		{
			name:     "missing version",
			yaml:     "id: x\nelements: {}\n",
			wantPath: "version",
		},
		{
			// A plain `version: 1` resolves to !!int; it must be rejected
			// rather than coerced, so Go and TypeScript agree (spec §5.2).
			name:     "version not a string",
			yaml:     "id: x\nversion: 1\nelements: {}\n",
			wantPath: "version",
		},
		{
			name:     "id not a string",
			yaml:     "id: 123\nversion: \"1\"\nelements: {}\n",
			wantPath: "id",
		},
		{
			name:     "id is a mapping",
			yaml:     "id: {a: 1}\nversion: \"1\"\nelements: {}\n",
			wantPath: "id",
		},
		{
			name:     "missing elements",
			yaml:     "id: x\nversion: \"1\"\n",
			wantPath: "elements",
		},
		{
			name:     "elements not a mapping",
			yaml:     "id: x\nversion: \"1\"\nelements:\n  - a\n",
			wantPath: "elements",
		},
		{
			name: "missing representation",
			yaml: `id: x
version: "1"
elements:
  T:
    match:
      query: 'kind exists'
    attributes:
      type: object
`,
			wantPath: "elements[T].representation",
		},
		{
			name: "invalid representation",
			yaml: `id: x
version: "1"
elements:
  T:
    representation: vertex
    match:
      query: 'kind exists'
    attributes:
      type: object
`,
			wantPath: "elements[T].representation",
		},
		{
			name: "missing match",
			yaml: `id: x
version: "1"
elements:
  T:
    representation: node
    attributes:
      type: object
`,
			wantPath: "elements[T].match",
		},
		{
			name: "missing query",
			yaml: `id: x
version: "1"
elements:
  T:
    representation: node
    match: {}
    attributes:
      type: object
`,
			wantPath: "elements[T].match.query",
		},
		{
			name: "missing attributes",
			yaml: `id: x
version: "1"
elements:
  T:
    representation: node
    match:
      query: 'kind exists'
`,
			wantPath: "elements[T].attributes",
		},
		{
			name: "invalid json schema",
			yaml: `id: x
version: "1"
elements:
  T:
    representation: node
    match:
      query: 'kind exists'
    attributes:
      type: 7
`,
			wantPath: "elements[T].attributes",
		},
		{
			name:     "not a mapping",
			yaml:     "- a\n- b\n",
			wantPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(strings.NewReader(tt.yaml))
			if err == nil {
				t.Fatal("expected dialect rejection, got nil error")
			}
			var le *LoadError
			if !errors.As(err, &le) {
				t.Fatalf("expected *LoadError, got %T: %v", err, err)
			}
			if tt.wantPath != "" && le.Path != tt.wantPath {
				t.Errorf("Path = %q, want %q (%v)", le.Path, tt.wantPath, err)
			}
		})
	}
}

// TestUnsupportedQueriesRejected verifies that query forms outside the
// supported subset are rejected at dialect-load time (spec §7.1).
func TestUnsupportedQueriesRejected(t *testing.T) {
	queries := []string{
		`kind == "service" traverse out 1`,
		`kind == "service" scope`,
		`kind == "service" | remove edge where edge.x == "y"`,
		`from API`,
		`kind == "service" | make node.x = 1 where exists`,
		`subgraph`,
	}
	for _, q := range queries {
		t.Run(q, func(t *testing.T) {
			doc := "id: x\nversion: \"1\"\nelements:\n  T:\n    representation: node\n    match:\n      query: '" + strings.ReplaceAll(q, "'", "''") + "'\n    attributes:\n      type: object\n"
			_, err := Load(strings.NewReader(doc))
			if err == nil {
				t.Fatalf("expected unsupported query %q to be rejected", q)
			}
			var le *LoadError
			if !errors.As(err, &le) {
				t.Fatalf("expected *LoadError, got %T", err)
			}
		})
	}
}

func TestEdgeRepresentation(t *testing.T) {
	d := mustLoad(t, `
id: edges
version: "1"
elements:
  HttpEdge:
    representation: edge
    match:
      query: 'protocol == "http"'
    attributes:
      type: object
      required: [protocol]
      properties:
        protocol:
          const: http
`)
	g := mustGraph(t, `
a -> b [protocol="http"]
b -> c [protocol="tcp"]
`)

	report, err := d.Validate(g)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !report.Conforms() {
		t.Fatalf("expected conforming, got %v", report.Violations)
	}

	// A violating http edge is still selected and reported.
	g2 := mustGraph(t, `a -> b [protocol="https"]`)
	d2 := mustLoad(t, `
id: edges
version: "1"
elements:
  HttpEdge:
    representation: edge
    match:
      query: 'protocol == "https"'
    attributes:
      type: object
      properties:
        protocol:
          const: http
`)
	report2, err := d2.Validate(g2)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if report2.Outcome != OutcomeNonConforming {
		t.Fatalf("expected non-conforming, got %s", report2.Outcome)
	}
}

// TestSetRepresentationSelectsAllSets documents the known GQL limitation: a
// subgraph result carries every set in the graph, so a set definition validates
// all sets regardless of the match query.
func TestSetRepresentationSelectsAllSets(t *testing.T) {
	d := mustLoad(t, `
id: sets
version: "1"
elements:
  Group:
    representation: set
    match:
      query: 'kind == "service"'
    attributes:
      type: object
      required: [label]
      properties:
        label:
          type: string
`)
	g := mustGraph(t, `
set alpha [label="Alpha"]
set beta
node a [kind="service"] @alpha
`)

	report, err := d.Validate(g)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if report.Outcome != OutcomeNonConforming {
		t.Fatalf("expected non-conforming (beta lacks label), got %s", report.Outcome)
	}
	var betaSeen bool
	for _, v := range report.Violations {
		if v.Element == "beta" {
			betaSeen = true
		}
	}
	if !betaSeen {
		t.Errorf("expected a violation for set beta selected by the definition, got %v", report.Violations)
	}
}

func TestMultipleDefinitionsIndependent(t *testing.T) {
	d := mustLoad(t, `
id: multi
version: "1"
elements:
  WithName:
    representation: node
    match:
      query: 'kind == "service"'
    attributes:
      type: object
      required: [name]
  Replicas:
    representation: node
    match:
      query: 'kind == "service"'
    attributes:
      type: object
      required: [replicas]
`)
	// Missing both name and replicas: each definition must report independently.
	g := mustGraph(t, `node a [kind="service"]`)

	report, err := d.Validate(g)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(report.Violations) != 2 {
		t.Fatalf("expected 2 violations (one per definition), got %d: %v", len(report.Violations), report.Violations)
	}
	defs := map[string]bool{}
	for _, v := range report.Violations {
		defs[v.Definition] = true
	}
	if !defs["WithName"] || !defs["Replicas"] {
		t.Errorf("expected violations from both definitions, got %v", defs)
	}
}

// TestInternalRefAllowed verifies that a schema may reference its own
// definitions without escaping the document.
func TestInternalRefAllowed(t *testing.T) {
	d := mustLoad(t, `
id: refs
version: "1"
elements:
  Thing:
    representation: node
    match:
      query: 'kind == "thing"'
    attributes:
      type: object
      required: [owner]
      properties:
        owner:
          $ref: '#/$defs/ref'
      $defs:
        ref:
          type: object
          properties:
            "$gslRef":
              type: string
          required: ["$gslRef"]
`)
	g := mustGraph(t, `
node team
node a [kind="thing", owner=team]
`)
	report, err := d.Validate(g)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !report.Conforms() {
		t.Fatalf("expected conforming, got %v", report.Violations)
	}
}

// TestExternalRefRejected verifies that attribute schemas cannot resolve
// external resources (spec §6.3: no silent network resolution).
func TestExternalRefRejected(t *testing.T) {
	_, err := Load(strings.NewReader(`
id: refs
version: "1"
elements:
  Thing:
    representation: node
    match:
      query: 'kind == "thing"'
    attributes:
      $ref: 'https://example.com/never-loaded.json'
`))
	if err == nil {
		t.Fatal("expected external $ref to be rejected at dialect load")
	}
	var le *LoadError
	if !errors.As(err, &le) {
		t.Fatalf("expected *LoadError, got %T: %v", err, err)
	}
}

// TestProjectValueKinds covers each documented GSL value mapping in isolation.
func TestProjectValueKinds(t *testing.T) {
	got, err := projectAttributes(gsl.AttributeMap{
		"str":   "hello",
		"num":   float64(3.5),
		"bool":  true,
		"empty": nil,
		"ref":   gsl.NodeRef("platform_team"),
	})
	if err != nil {
		t.Fatalf("projectAttributes: %v", err)
	}
	want := map[string]any{
		"str":   "hello",
		"num":   float64(3.5),
		"bool":  true,
		"empty": map[string]any{"$gslEmpty": true},
		"ref":   map[string]any{"$gslRef": "platform_team"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("projection mismatch:\n got: %#v\nwant: %#v", got, want)
	}

	if _, err := projectAttributes(gsl.AttributeMap{"weird": struct{}{}}); err == nil {
		t.Error("expected an error for an unsupported GSL value type")
	}
}

// TestMatchQueryIsBarePredicate checks that a match query is written as a bare
// attribute predicate and expanded with the definition's representation
// (spec §6.2).
func TestMatchQueryIsBarePredicate(t *testing.T) {
	d := mustLoad(t, `
id: bare
version: "1"
elements:
  Service:
    representation: node
    match:
      query: 'kind == "service" AND active exists'
    attributes:
      type: object
      properties:
        kind:
          const: service
`)

	// The stored Query is the author's bare text, not the expanded form.
	els := d.Elements()
	if got := els[0].Query; got != `kind == "service" AND active exists` {
		t.Errorf("Query = %q, want the bare authored text", got)
	}

	g := mustGraph(t, `
node a [kind="service", active=true]
node b [kind="team", active=true]
node c [kind="service"]
`)
	selected, err := d.Match(g)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if len(selected) != 1 || selected[0].Element != "a" {
		t.Fatalf("expected only node a selected, got %+v", selected)
	}
}

// TestMatchQueryRejectsExpandedForms checks that writing the element prefix or
// the subgraph wrapper is an invalid dialect (spec §6.2).
func TestMatchQueryRejectsExpandedForms(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"explicit node prefix", `node.kind == "service"`},
		{"explicit edge prefix", `edge.protocol == "http"`},
		{"explicit set prefix", `set.kind == "service"`},
		{"explicit subgraph wrapper", `subgraph node.kind == "service"`},
		{"explicit subgraph with bare predicate", `subgraph kind == "service"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := "id: x\nversion: \"1\"\nelements:\n  T:\n    representation: node\n    match:\n      query: '" + tt.query + "'\n    attributes:\n      type: object\n"
			_, err := Load(strings.NewReader(doc))
			if err == nil {
				t.Fatalf("expected %q to be rejected", tt.query)
			}
			if !strings.Contains(err.Error(), "match.query") {
				t.Errorf("error should reference match.query, got %v", err)
			}
		})
	}
}

// TestSetMatchQueryIsValidatedNotEvaluated checks that a set definition's bare
// predicate must be well-formed but never filters sets (spec §6.2, §7.5).
func TestSetMatchQueryIsValidatedNotEvaluated(t *testing.T) {
	// Well-formed: selects all sets even though no set has kind "service".
	d := mustLoad(t, `
id: sets
version: "1"
elements:
  Group:
    representation: set
    match:
      query: 'kind == "service"'
    attributes:
      type: object
`)
	g := mustGraph(t, `
set alpha
set beta
node a [kind="service"] @alpha
`)
	selected, err := d.Match(g)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if len(selected) != 2 {
		t.Fatalf("expected both sets selected, got %+v", selected)
	}

	// Ill-formed: still an invalid dialect.
	bad := "id: x\nversion: \"1\"\nelements:\n  G:\n    representation: set\n    match:\n      query: 'kind =='\n    attributes:\n      type: object\n"
	if _, err := Load(strings.NewReader(bad)); err == nil {
		t.Fatal("expected an ill-formed set predicate to be rejected")
	}
}
