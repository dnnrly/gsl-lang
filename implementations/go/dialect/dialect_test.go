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

func nodeRuleDoc(match, attributes string) string {
	return "id: x\nversion: \"1\"\nnodes:\n  rules:\n    - match: '" + strings.ReplaceAll(match, "'", "''") + "'\n      attributes:\n" + attributes
}

func TestLoadWorkedExample(t *testing.T) {
	d := mustLoad(t, readFile(t, "testdata/software-architecture.dialect.yaml"))

	if d.ID != "software-architecture" {
		t.Errorf("ID = %q", d.ID)
	}
	if d.Version != "0.4.0-draft" {
		t.Errorf("Version = %q", d.Version)
	}
	if d.Title != "Software Architecture" {
		t.Errorf("Title = %q", d.Title)
	}

	rules := d.NodeRules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 node rule, got %d", len(rules))
	}
	rule := rules[0]
	if rule.Name != "Service" || rule.Label() != "Service" {
		t.Errorf("Name/Label = %q/%q", rule.Name, rule.Label())
	}
	if rule.Representation != RepresentationNode {
		t.Errorf("Representation = %q", rule.Representation)
	}
	if rule.Query != `kind == "service"` {
		t.Errorf("Query = %q", rule.Query)
	}
	if len(d.EdgeRules()) != 0 {
		t.Errorf("expected no edge rules, got %d", len(d.EdgeRules()))
	}
	if got := d.SetPolicy(); got != DefaultSetPolicy() {
		t.Errorf("SetPolicy = %+v, want defaults", got)
	}
}

// TestProjectionWorkedExample checks the projection example exactly: node
// references, Empty values, numbers, booleans and strings are all
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
nodes:
  rules:
    - name: Service
      match: 'kind == "service"'
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

func TestValidateEmptySectionsConforming(t *testing.T) {
	tests := map[string]string{
		"omitted":  "id: empty\nversion: \"1\"\n",
		"empty":    "id: empty\nversion: \"1\"\nnodes: {}\nedges: {}\nsets: {}\n",
		"no rules": "id: empty\nversion: \"1\"\nnodes:\n  rules: []\nedges:\n  rules: []\n",
	}
	for name, doc := range tests {
		t.Run(name, func(t *testing.T) {
			d := mustLoad(t, doc)
			report, err := d.Validate(mustGraph(t, "node a [x=1]"))
			if err != nil {
				t.Fatalf("Validate: %v", err)
			}
			if !report.Conforms() {
				t.Fatalf("dialect with no rules must conform, got %v", report.Violations)
			}
		})
	}
}

func TestUnknownFieldsIgnored(t *testing.T) {
	d := mustLoad(t, `
id: tolerant
version: "1"
something_extra: true
nodes:
  future_section: whatever
  rules:
    - future_field: whatever
      match: 'kind == "thing"'
      attributes:
        type: object
sets:
  future_policy: whatever
  definitions:
    thing:
      future_field: whatever
`)
	if len(d.NodeRules()) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(d.NodeRules()))
	}
	if len(d.SetDefinitions()) != 1 {
		t.Fatalf("expected 1 set definition, got %d", len(d.SetDefinitions()))
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
			yaml:     "version: \"1\"\n",
			wantPath: "id",
		},
		{
			name:     "empty id",
			yaml:     "id: \"\"\nversion: \"1\"\n",
			wantPath: "id",
		},
		{
			name:     "missing version",
			yaml:     "id: x\n",
			wantPath: "version",
		},
		{
			// A plain `version: 1` resolves to !!int; it must be rejected
			// rather than coerced, so Go and TypeScript agree.
			name:     "version not a string",
			yaml:     "id: x\nversion: 1\n",
			wantPath: "version",
		},
		{
			name:     "id not a string",
			yaml:     "id: 123\nversion: \"1\"\n",
			wantPath: "id",
		},
		{
			name:     "id is a mapping",
			yaml:     "id: {a: 1}\nversion: \"1\"\n",
			wantPath: "id",
		},
		{
			name:     "nodes not a mapping",
			yaml:     "id: x\nversion: \"1\"\nnodes:\n  - a\n",
			wantPath: "nodes",
		},
		{
			name:     "rules not a sequence",
			yaml:     "id: x\nversion: \"1\"\nnodes:\n  rules: {}\n",
			wantPath: "nodes.rules",
		},
		{
			name:     "rule not a mapping",
			yaml:     "id: x\nversion: \"1\"\nnodes:\n  rules:\n    - a\n",
			wantPath: "nodes.rules[0]",
		},
		{
			name:     "missing match",
			yaml:     "id: x\nversion: \"1\"\nnodes:\n  rules:\n    - attributes:\n        type: object\n",
			wantPath: "nodes.rules[0].match",
		},
		{
			name:     "match not a string",
			yaml:     "id: x\nversion: \"1\"\nnodes:\n  rules:\n    - match: 7\n      attributes:\n        type: object\n",
			wantPath: "nodes.rules[0].match",
		},
		{
			name:     "empty name",
			yaml:     "id: x\nversion: \"1\"\nedges:\n  rules:\n    - name: \"\"\n      match: 'protocol exists'\n      attributes:\n        type: object\n",
			wantPath: "edges.rules[0].name",
		},
		{
			name:     "missing attributes",
			yaml:     "id: x\nversion: \"1\"\nnodes:\n  rules:\n    - match: 'kind exists'\n",
			wantPath: "nodes.rules[0].attributes",
		},
		{
			name:     "invalid json schema",
			yaml:     "id: x\nversion: \"1\"\nnodes:\n  rules:\n    - match: 'kind exists'\n      attributes:\n        type: 7\n",
			wantPath: "nodes.rules[0].attributes",
		},
		{
			name:     "sets not a mapping",
			yaml:     "id: x\nversion: \"1\"\nsets:\n  - a\n",
			wantPath: "sets",
		},
		{
			name:     "policy not a mapping",
			yaml:     "id: x\nversion: \"1\"\nsets:\n  policy: true\n",
			wantPath: "sets.policy",
		},
		{
			name:     "policy value not a boolean",
			yaml:     "id: x\nversion: \"1\"\nsets:\n  policy:\n    allow_undefined: \"false\"\n",
			wantPath: "sets.policy.allow_undefined",
		},
		{
			name:     "definitions not a mapping",
			yaml:     "id: x\nversion: \"1\"\nsets:\n  definitions:\n    - a\n",
			wantPath: "sets.definitions",
		},
		{
			name:     "invalid set name",
			yaml:     "id: x\nversion: \"1\"\nsets:\n  definitions:\n    critical-group: {}\n",
			wantPath: "sets.definitions",
		},
		{
			name:     "definition not a mapping",
			yaml:     "id: x\nversion: \"1\"\nsets:\n  definitions:\n    critical: true\n",
			wantPath: "sets.definitions[critical]",
		},
		{
			name:     "invalid set attributes",
			yaml:     "id: x\nversion: \"1\"\nsets:\n  definitions:\n    critical:\n      attributes:\n        type: 7\n",
			wantPath: "sets.definitions[critical].attributes",
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
// supported subset are rejected at dialect-load time.
func TestUnsupportedQueriesRejected(t *testing.T) {
	queries := []string{
		`kind == "service" traverse out 1`,
		`kind == "service" scope`,
		`kind == "service" | remove edge where x == "y"`,
		`from API`,
		`kind == "service" | make x = 1 where exists`,
		`subgraph`,
		`kind ==`,
	}
	for _, q := range queries {
		t.Run(q, func(t *testing.T) {
			doc := nodeRuleDoc(q, "        type: object\n")
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

func TestEdgeRules(t *testing.T) {
	d := mustLoad(t, `
id: edges
version: "1"
edges:
  rules:
    - name: HttpEdge
      match: 'protocol == "http"'
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
edges:
  rules:
    - match: 'protocol == "https"'
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
	if report2.Violations[0].Definition != "edges.rules[0]" {
		t.Errorf("unnamed rule label = %q", report2.Violations[0].Definition)
	}
}

func TestSetDefinitionsValidatePresentDefinedSets(t *testing.T) {
	d := mustLoad(t, `
id: sets
version: "1"
sets:
  definitions:
    alpha:
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
	if !report.Conforms() {
		t.Fatalf("undefined beta is allowed by default, got %v", report.Violations)
	}

	bad := mustGraph(t, `
set alpha
node a @alpha
`)
	report2, err := d.Validate(bad)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if report2.Outcome != OutcomeNonConforming {
		t.Fatalf("expected non-conforming, got %s", report2.Outcome)
	}
	if len(report2.Violations) != 1 || report2.Violations[0].Definition != "alpha" || report2.Violations[0].Element != "alpha" {
		t.Errorf("unexpected violations: %+v", report2.Violations)
	}
}

func TestOptionalSetAttributes(t *testing.T) {
	d := mustLoad(t, `
id: sets
version: "1"
sets:
  definitions:
    critical:
      description: Important elements.
`)
	g := mustGraph(t, "set critical\nnode a @critical\n")
	selected, err := d.Match(g)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if len(selected) != 1 || selected[0].Definition != "critical" || selected[0].Representation != RepresentationSet {
		t.Fatalf("unexpected selected sets: %+v", selected)
	}
	report, err := d.Validate(g)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !report.Conforms() {
		t.Fatalf("definition without attributes must not fail validation: %v", report.Violations)
	}
}

func TestSetPolicyDefaultsPermitAnything(t *testing.T) {
	d := mustLoad(t, `
id: policy
version: "1"
sets:
  definitions:
    unused: {}
`)
	g := mustGraph(t, "set empty\nnode a @implicit\n")
	report, err := d.Validate(g)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !report.Conforms() {
		t.Fatalf("default permissive policy must conform, got %v", report.Violations)
	}
}

func TestSetPolicyAllowUndefined(t *testing.T) {
	d := mustLoad(t, `
id: policy
version: "1"
sets:
  policy:
    allow_undefined: false
  definitions:
    alpha: {}
`)
	g := mustGraph(t, "set alpha\nset beta\nnode a @implicit\n")

	report, err := d.Validate(g)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if report.Outcome != OutcomeNonConforming {
		t.Fatalf("expected non-conforming, got %s", report.Outcome)
	}
	got := map[string]bool{}
	for _, v := range report.Violations {
		got[v.Element] = true
		if v.Definition != "" {
			t.Errorf("undefined set violation should have no definition: %+v", v)
		}
	}
	if !got["beta"] || !got["implicit"] {
		t.Errorf("expected beta and implicit violations, got %v", report.Violations)
	}
}

func TestSetPolicyRequireAll(t *testing.T) {
	d := mustLoad(t, `
id: policy
version: "1"
sets:
  policy:
    require_all: true
  definitions:
    alpha: {}
    beta: {}
`)
	report, err := d.Validate(mustGraph(t, "set alpha\n"))
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if report.Outcome != OutcomeNonConforming {
		t.Fatalf("expected non-conforming, got %s", report.Outcome)
	}
	if len(report.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %v", report.Violations)
	}
	v := report.Violations[0]
	if v.Definition != "beta" || v.Element != "beta" {
		t.Errorf("unexpected violation: %+v", v)
	}
}

func TestSetPolicyAllowEmpty(t *testing.T) {
	d := mustLoad(t, `
id: policy
version: "1"
sets:
  policy:
    allow_empty: false
  definitions:
    alpha: {}
    implicit: {}
`)
	report, err := d.Validate(mustGraph(t, "set alpha\nnode a @implicit\n"))
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if report.Outcome != OutcomeNonConforming {
		t.Fatalf("expected non-conforming, got %s", report.Outcome)
	}
	if len(report.Violations) != 1 || report.Violations[0].Element != "alpha" {
		t.Fatalf("expected only empty alpha to violate, got %v", report.Violations)
	}
}

func TestImplicitSetCountsAsPresent(t *testing.T) {
	d := mustLoad(t, `
id: policy
version: "1"
sets:
  policy:
    require_all: true
    allow_empty: false
  definitions:
    implicit: {}
`)
	report, err := d.Validate(mustGraph(t, "node a @implicit\n"))
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !report.Conforms() {
		t.Fatalf("implicit set is present and non-empty, got %v", report.Violations)
	}
}

func TestMultipleRulesIndependent(t *testing.T) {
	d := mustLoad(t, `
id: multi
version: "1"
nodes:
  rules:
    - name: WithName
      match: 'kind == "service"'
      attributes:
        type: object
        required: [name]
    - name: Replicas
      match: 'kind == "service"'
      attributes:
        type: object
        required: [replicas]
`)
	// Missing both name and replicas: each rule must report independently.
	g := mustGraph(t, `node a [kind="service"]`)

	report, err := d.Validate(g)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(report.Violations) != 2 {
		t.Fatalf("expected 2 violations (one per rule), got %d: %v", len(report.Violations), report.Violations)
	}
	defs := map[string]bool{}
	for _, v := range report.Violations {
		defs[v.Definition] = true
	}
	if !defs["WithName"] || !defs["Replicas"] {
		t.Errorf("expected violations from both rules, got %v", defs)
	}
}

// TestInternalRefAllowed verifies that a schema may reference its own
// definitions without escaping the document.
func TestInternalRefAllowed(t *testing.T) {
	d := mustLoad(t, `
id: refs
version: "1"
nodes:
  rules:
    - match: 'kind == "thing"'
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

// TestRefIsScopedToContainingSchema verifies that each rule's schema has its
// own resource: an internal $ref cannot resolve in another rule.
func TestRefIsScopedToContainingSchema(t *testing.T) {
	_, err := Load(strings.NewReader(`
id: refs
version: "1"
nodes:
  rules:
    - match: 'kind == "a"'
      attributes:
        type: object
        properties:
          owner:
            $ref: '#/$defs/ref'
    - match: 'kind == "b"'
      attributes:
        type: object
        $defs:
          ref:
            type: string
`))
	if err == nil {
		t.Fatal("expected a cross-schema internal $ref to be rejected")
	}
}

// TestExternalRefRejected verifies that attribute schemas cannot resolve
// external resources (no silent network resolution).
func TestExternalRefRejected(t *testing.T) {
	_, err := Load(strings.NewReader(`
id: refs
version: "1"
nodes:
  rules:
    - match: 'kind == "thing"'
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
// attribute predicate and expanded with the section's element prefix.
func TestMatchQueryIsBarePredicate(t *testing.T) {
	d := mustLoad(t, `
id: bare
version: "1"
nodes:
  rules:
    - match: 'kind == "service" AND active exists'
      attributes:
        type: object
        properties:
          kind:
            const: service
`)

	// The stored Query is the author's bare text, not the expanded form.
	rules := d.NodeRules()
	if got := rules[0].Query; got != `kind == "service" AND active exists` {
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
// the subgraph wrapper is an invalid dialect.
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
			doc := nodeRuleDoc(tt.query, "        type: object\n")
			_, err := Load(strings.NewReader(doc))
			if err == nil {
				t.Fatalf("expected %q to be rejected", tt.query)
			}
			if !strings.Contains(err.Error(), "nodes.rules[0].match") {
				t.Errorf("error should reference nodes.rules[0].match, got %v", err)
			}
		})
	}
}

func TestRuleOrderAndLabels(t *testing.T) {
	d := mustLoad(t, `
id: order
version: "1"
nodes:
  rules:
    - name: First
      match: 'kind exists'
      attributes:
        type: object
    - match: 'kind exists'
      attributes:
        type: object
edges:
  rules:
    - name: Relation
      match: 'protocol exists'
      attributes:
        type: object
    - match: 'protocol exists'
      attributes:
        type: object
sets:
  definitions:
    zeta: {}
    alpha: {}
`)
	if got := d.NodeRules()[0].Label(); got != "First" {
		t.Errorf("named node rule label = %q", got)
	}
	if got := d.NodeRules()[1].Label(); got != "nodes.rules[1]" {
		t.Errorf("unnamed node rule label = %q", got)
	}
	if got := d.EdgeRules()[0].Label(); got != "Relation" {
		t.Errorf("named edge rule label = %q", got)
	}
	if got := d.EdgeRules()[1].Label(); got != "edges.rules[1]" {
		t.Errorf("unnamed edge rule label = %q", got)
	}
	defs := d.SetDefinitions()
	if len(defs) != 2 || defs[0].Name != "alpha" || defs[1].Name != "zeta" {
		t.Errorf("set definitions not name-sorted: %+v", defs)
	}
}
