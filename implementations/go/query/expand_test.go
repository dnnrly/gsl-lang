// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.
package query

import "testing"

func TestExpandMatchQuery(t *testing.T) {
	tests := []struct {
		element string
		in      string
		want    string
	}{
		{"node", `kind == "service"`, `subgraph node.kind == "service"`},
		{"node", `kind != "service"`, `subgraph node.kind != "service"`},
		{"node", `active exists`, `subgraph node.active exists`},
		{"node", `active not exists`, `subgraph node.active not exists`},
		{"node", `zone == "B" AND team == "payments"`, `subgraph node.zone == "B" AND node.team == "payments"`},
		{"node", `in @critical`, `subgraph node in @critical`},
		{"node", `not in @deprecated`, `subgraph node not in @deprecated`},
		{"edge", `protocol == "http"`, `subgraph edge.protocol == "http"`},
		{"edge", `parent exists`, `subgraph edge parent exists`},
		{"edge", `parent not exists`, `subgraph edge parent not exists`},
		{"edge", `depends on protocol == "http"`, `subgraph edge depends on edge.protocol == "http"`},
		{"edge", `depth == 0`, `subgraph edge.depth == 0`},
		{"edge", `depth != 0`, `subgraph edge.depth != 0`},
		{"node", `replicas == 3`, `subgraph node.replicas == 3`},
		{"node", `kind == "a AND b"`, `subgraph node.kind == "a AND b"`},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ExpandMatchQuery(tt.element, tt.in)
			if err != nil {
				t.Fatalf("ExpandMatchQuery(%q, %q): %v", tt.element, tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ExpandMatchQuery(%q, %q) = %q, want %q", tt.element, tt.in, got, tt.want)
			}
			if _, perr := NewQueryParser(got).Parse(); perr != nil {
				t.Errorf("expanded query %q did not parse: %v", got, perr)
			}
		})
	}
}

func TestExpandMatchQueryRejects(t *testing.T) {
	tests := []struct {
		element string
		in      string
	}{
		{"node", `node.kind == "service"`},
		{"node", `edge.protocol == "http"`},
		{"node", `set.kind == "service"`},
		{"node", `subgraph node.kind == "service"`},
		{"node", `subgraph kind == "service"`},
		{"node", ``},
		{"node", `   `},
		{"node", `kind == "service" | remove orphans`},
		{"edge", `kind == "service" traverse out 1`},
		{"set", `kind == "service"`}, // no set element prefix exists
	}
	for _, tt := range tests {
		t.Run(tt.element+"/"+tt.in, func(t *testing.T) {
			if _, err := ExpandMatchQuery(tt.element, tt.in); err == nil {
				t.Errorf("ExpandMatchQuery(%q, %q) = nil error, want rejection", tt.element, tt.in)
			}
		})
	}
}
