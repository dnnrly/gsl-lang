package converter

import (
	"testing"

	gsl "github.com/dnnrly/gsl-lang"
)

func TestGetNodeLabel(t *testing.T) {
	tests := []struct {
		name     string
		node     *gsl.Node
		expected string
	}{
		{
			name: "node with text attribute",
			node: &gsl.Node{
				ID: "mynode",
				Attributes: map[string]interface{}{
					"text": "My Label",
				},
			},
			expected: "My Label",
		},
		{
			name: "node without text attribute",
			node: &gsl.Node{
				ID: "mynode",
				Attributes: map[string]interface{}{},
			},
			expected: "mynode",
		},
		{
			name: "node with non-string text attribute",
			node: &gsl.Node{
				ID: "mynode",
				Attributes: map[string]interface{}{
					"text": 123,
				},
			},
			expected: "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetNodeLabel(tt.node)
			if got != tt.expected {
				t.Errorf("GetNodeLabel() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetEdgeLabel(t *testing.T) {
	tests := []struct {
		name     string
		edge     *gsl.Edge
		expected string
	}{
		{
			name: "edge with label attribute",
			edge: &gsl.Edge{
				From: "a",
				To:   "b",
				Attributes: map[string]interface{}{
					"label": "connects",
				},
			},
			expected: "connects",
		},
		{
			name: "edge without label attribute",
			edge: &gsl.Edge{
				From: "a",
				To:   "b",
				Attributes: map[string]interface{}{},
			},
			expected: "",
		},
		{
			name: "edge with non-string label attribute",
			edge: &gsl.Edge{
				From: "a",
				To:   "b",
				Attributes: map[string]interface{}{
					"label": 456,
				},
			},
			expected: "456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetEdgeLabel(tt.edge)
			if got != tt.expected {
				t.Errorf("GetEdgeLabel() = %q, want %q", got, tt.expected)
			}
		})
	}
}
