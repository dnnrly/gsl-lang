// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

package gsl_test

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/dnnrly/gsl-lang/implementations/go"
	"github.com/dnnrly/gsl-lang/implementations/go/dialect"
)

const dialectExampleDir = "dialects/software-architecture"

func loadDialectExample(t *testing.T) (*dialect.Dialect, gsl.Graph) {
	t.Helper()
	d, err := dialect.LoadFile(filepath.Join(dialectExampleDir, "software-architecture.dialect.yaml"))
	if err != nil {
		t.Fatalf("failed to load reference dialect: %v", err)
	}
	model, err := os.ReadFile(filepath.Join(dialectExampleDir, "model.gsl"))
	if err != nil {
		t.Fatalf("failed to read model.gsl: %v", err)
	}
	graph, parseErr := gsl.Parse(bytes.NewReader(model))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("model.gsl failed to parse: %v", parseErr)
	}
	if parseErr != nil && parseErr.HasWarnings() {
		t.Errorf("model.gsl produced warnings: %v", parseErr.Warnings)
	}
	return d, *graph
}

func TestDialectExampleConforming(t *testing.T) {
	d, graph := loadDialectExample(t)

	report, err := d.Validate(&graph)
	if err != nil {
		t.Fatalf("Validate returned incomplete validation: %v", err)
	}
	if !report.Conforms() {
		t.Fatalf("expected the reference model to conform, got violations: %v", report.Violations)
	}
}

func TestDialectExampleProjection(t *testing.T) {
	d, graph := loadDialectExample(t)

	selected, err := d.Match(&graph)
	if err != nil {
		t.Fatalf("Match: %v", err)
	}

	byDef := map[string]map[string]any{}
	for _, s := range selected {
		if byDef[s.Definition] == nil {
			byDef[s.Definition] = map[string]any{}
		}
		byDef[s.Definition][s.Element] = s.Attributes
	}

	wantGateway := map[string]any{
		"kind":        "service",
		"name":        "Gateway",
		"owner":       map[string]any{"$gslRef": "platform_team"},
		"replicas":    float64(3),
		"description": "Public entry point",
	}
	if got := byDef["Service"]["gateway"]; !reflect.DeepEqual(got, wantGateway) {
		t.Errorf("gateway projection mismatch:\n got: %#v\nwant: %#v", got, wantGateway)
	}

	// Edge definitions are keyed by from->to when no label is present.
	dep, ok := byDef["Dependency"]["gateway->payments"]
	if !ok {
		t.Errorf("missing projected gateway->payments dependency; got %v", byDef["Dependency"])
	} else if dep.(map[string]any)["protocol"] != "grpc" {
		t.Errorf("gateway->payments protocol = %v, want grpc", dep.(map[string]any)["protocol"])
	}

	// Defined, present sets are selected by their GSL set name.
	if got := byDef["critical"]; len(got) != 1 {
		t.Errorf("expected the critical set to be selected, got %v", got)
	}
	if got := byDef["deprecated"]; len(got) != 1 {
		t.Errorf("expected the deprecated set to be selected, got %v", got)
	}
}

func TestDialectExampleMutations(t *testing.T) {
	d, _ := loadDialectExample(t)

	tests := []struct {
		name       string
		model      string
		violations []string // "definition\x00element\x00location"
	}{
		{
			name: "missing_required_name",
			model: `set critical
set deprecated
node platform_team [kind="team"] @deprecated
node gateway [kind="service", owner=platform_team, replicas=1] @critical
`,
			violations: []string{"Service\x00gateway\x00"},
		},
		{
			name: "replicas_below_minimum",
			model: `set critical
set deprecated
node platform_team [kind="team"] @deprecated
node gateway [kind="service", name="Gateway", owner=platform_team, replicas=0] @critical
`,
			violations: []string{"Service\x00gateway\x00/replicas"},
		},
		{
			name: "owner_is_a_string_not_a_reference",
			model: `set critical
set deprecated
node platform_team [kind="team"] @deprecated
node gateway [kind="service", name="Gateway", owner="platform_team", replicas=1] @critical
`,
			violations: []string{"Service\x00gateway\x00/owner"},
		},
		{
			name: "protocol_outside_vocabulary",
			model: `set critical
set deprecated
node platform_team [kind="team"] @deprecated
node gateway [kind="service", name="Gateway", owner=platform_team, replicas=1] @critical

gateway -> platform_team [protocol="amqp"]
`,
			violations: []string{"Dependency\x00gateway->platform_team\x00/protocol"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			graph, parseErr := gsl.Parse(strings.NewReader(tc.model))
			if parseErr != nil && parseErr.HasError() {
				t.Fatalf("failed to parse mutation model: %v", parseErr)
			}

			report, err := d.Validate(graph)
			if err != nil {
				t.Fatalf("Validate returned incomplete validation: %v", err)
			}
			if report.Conforms() {
				t.Fatalf("expected non-conformance, got conforming")
			}

			var got []string
			for _, v := range report.Violations {
				got = append(got, v.Definition+"\x00"+v.Element+"\x00"+v.InstanceLocation)
			}
			sort.Strings(got)
			want := append([]string(nil), tc.violations...)
			sort.Strings(want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("violations mismatch\n got: %v\nwant: %v\nraw: %v", got, want, report.Violations)
			}
		})
	}
}
