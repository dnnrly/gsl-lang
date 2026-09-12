package formats

import "testing"

func TestValidateDiagramType(t *testing.T) {
	cases := []struct {
		format string
		dtype  string
		wantOk bool
	}{
		{"mermaid", "component", true},
		{"mermaid", "graph", true},
		{"mermaid", "sequence", true},
		{"mermaid", "gibberish", false},
		{"plantuml", "component", true},
		{"plantuml", "sequence", true},
		{"plantuml", "graph", false},
		{"plantuml", "gibberish", false},
		{"unknown", "component", false},
	}

	for _, c := range cases {
		err := ValidateDiagramType(c.format, c.dtype)
		if c.wantOk && err != nil {
			t.Errorf("ValidateDiagramType(%q, %q) = %v, want nil", c.format, c.dtype, err)
		}
		if !c.wantOk && err == nil {
			t.Errorf("ValidateDiagramType(%q, %q) = nil, want error", c.format, c.dtype)
		}
	}
}