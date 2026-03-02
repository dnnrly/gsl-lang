package query

import (
	"testing"
)

func TestParseQueryStartStep(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
		check   func(t *testing.T, q *Query)
	}{
		{
			name:    "simple start",
			query:   "start A",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				if len(pipe.Steps) != 1 {
					t.Fatalf("expected 1 step, got %d", len(pipe.Steps))
				}
				start := pipe.Steps[0].(*StartStep)
				if len(start.NodeIDs) != 1 || start.NodeIDs[0] != "A" {
					t.Errorf("expected NodeIDs=[A], got %v", start.NodeIDs)
				}
			},
		},
		{
			name:    "start with quoted string",
			query:   `start "AuthService"`,
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				start := pipe.Steps[0].(*StartStep)
				if len(start.NodeIDs) != 1 || start.NodeIDs[0] != "AuthService" {
					t.Errorf("expected NodeIDs=[AuthService], got %v", start.NodeIDs)
				}
			},
		},
		{
			name:    "start with multiple nodes",
			query:   "start A, B, C",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				start := pipe.Steps[0].(*StartStep)
				if len(start.NodeIDs) != 3 {
					t.Fatalf("expected 3 node IDs, got %d", len(start.NodeIDs))
				}
				if start.NodeIDs[0] != "A" || start.NodeIDs[1] != "B" || start.NodeIDs[2] != "C" {
					t.Errorf("got %v", start.NodeIDs)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, errs := ParseQuery(tt.query)
			if (len(errs) > 0) != tt.wantErr {
				t.Fatalf("wantErr=%v, got errs: %v", tt.wantErr, errs)
			}
			if tt.check != nil && !tt.wantErr {
				tt.check(t, q)
			}
		})
	}
}

func TestParseQueryFlowStep(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
		check   func(t *testing.T, q *Query)
	}{
		{
			name:    "simple flow out",
			query:   "start A | flow out",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				if len(pipe.Steps) != 2 {
					t.Fatalf("expected 2 steps, got %d", len(pipe.Steps))
				}
				flow := pipe.Steps[1].(*FlowStep)
				if flow.Direction != "out" {
					t.Errorf("expected direction=out, got %s", flow.Direction)
				}
				if flow.Recursive {
					t.Error("expected Recursive=false")
				}
			},
		},
		{
			name:    "flow with recursive",
			query:   "start A | flow out recursive",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				flow := pipe.Steps[1].(*FlowStep)
				if !flow.Recursive {
					t.Error("expected Recursive=true")
				}
			},
		},
		{
			name:    "flow with star",
			query:   "start A | flow in *",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				flow := pipe.Steps[1].(*FlowStep)
				if flow.Direction != "in" {
					t.Errorf("expected direction=in, got %s", flow.Direction)
				}
				if !flow.Recursive {
					t.Error("expected Recursive=true")
				}
			},
		},
		{
			name:    "flow with edge filter",
			query:   `start A | flow out where edge.color = "Blue"`,
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				flow := pipe.Steps[1].(*FlowStep)
				if flow.EdgeFilter == nil {
					t.Fatal("expected EdgeFilter, got nil")
				}
				if flow.EdgeFilter.Attr != "color" {
					t.Errorf("expected attr=color, got %s", flow.EdgeFilter.Attr)
				}
				if flow.EdgeFilter.Op != "=" {
					t.Errorf("expected op==, got %s", flow.EdgeFilter.Op)
				}
				if flow.EdgeFilter.Value != "Blue" {
					t.Errorf("expected value=Blue, got %v", flow.EdgeFilter.Value)
				}
			},
		},
		{
			name:    "flow both directions",
			query:   "start A | flow both",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				flow := pipe.Steps[1].(*FlowStep)
				if flow.Direction != "both" {
					t.Errorf("expected direction=both, got %s", flow.Direction)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, errs := ParseQuery(tt.query)
			if (len(errs) > 0) != tt.wantErr {
				t.Fatalf("wantErr=%v, got errs: %v", tt.wantErr, errs)
			}
			if tt.check != nil && !tt.wantErr {
				tt.check(t, q)
			}
		})
	}
}

func TestParseQueryFilterStep(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
		check   func(t *testing.T, q *Query)
	}{
		{
			name:    "filter with string equality",
			query:   `start A | where status = "active"`,
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				filter := pipe.Steps[1].(*FilterStep)
				if filter.Filter.Attr != "status" {
					t.Errorf("expected attr=status, got %s", filter.Filter.Attr)
				}
				if filter.Filter.Op != "=" {
					t.Errorf("expected op==, got %s", filter.Filter.Op)
				}
				if filter.Filter.Value != "active" {
					t.Errorf("expected value=active, got %v", filter.Filter.Value)
				}
			},
		},
		{
			name:    "filter with number",
			query:   "start A | where count = 5",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				filter := pipe.Steps[1].(*FilterStep)
				if v, ok := filter.Filter.Value.(float64); !ok || v != 5 {
					t.Errorf("expected value=5, got %v", filter.Filter.Value)
				}
			},
		},
		{
			name:    "filter with boolean",
			query:   "start A | where critical = true",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				filter := pipe.Steps[1].(*FilterStep)
				if v, ok := filter.Filter.Value.(bool); !ok || !v {
					t.Errorf("expected value=true, got %v", filter.Filter.Value)
				}
			},
		},
		{
			name:    "filter with !=",
			query:   `start A | where type != "archived"`,
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				filter := pipe.Steps[1].(*FilterStep)
				if filter.Filter.Op != "!=" {
					t.Errorf("expected op!=, got %s", filter.Filter.Op)
				}
			},
		},
		{
			name:    "filter with less than",
			query:   `start A | where count < 10`,
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				filter := pipe.Steps[1].(*FilterStep)
				if filter.Filter.Op != "<" {
					t.Errorf("expected op=<, got %s", filter.Filter.Op)
				}
				if filter.Filter.Value != 10.0 {
					t.Errorf("expected value=10, got %v", filter.Filter.Value)
				}
			},
		},
		{
			name:    "filter with greater than or equal",
			query:   `start A | where priority >= 5`,
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				filter := pipe.Steps[1].(*FilterStep)
				if filter.Filter.Op != ">=" {
					t.Errorf("expected op=>=, got %s", filter.Filter.Op)
				}
				if filter.Filter.Value != 5.0 {
					t.Errorf("expected value=5, got %v", filter.Filter.Value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, errs := ParseQuery(tt.query)
			if (len(errs) > 0) != tt.wantErr {
				t.Fatalf("wantErr=%v, got errs: %v", tt.wantErr, errs)
			}
			if tt.check != nil && !tt.wantErr {
				tt.check(t, q)
			}
		})
	}
}

func TestParseQueryMinusStep(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
		check   func(t *testing.T, q *Query)
	}{
		{
			name:    "simple minus",
			query:   "start A | minus (start B)",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				if len(pipe.Steps) != 2 {
					t.Fatalf("expected 2 steps, got %d", len(pipe.Steps))
				}
				minus := pipe.Steps[1].(*MinusStep)
				if minus.Pipeline == nil {
					t.Fatal("expected Pipeline, got nil")
				}
				if len(minus.Pipeline.Steps) != 1 {
					t.Fatalf("expected 1 step in minus pipeline, got %d", len(minus.Pipeline.Steps))
				}
			},
		},
		{
			name:    "minus with complex pipeline",
			query:   "start A | minus (start B | flow out | where type = \"test\")",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				minus := pipe.Steps[1].(*MinusStep)
				if len(minus.Pipeline.Steps) != 3 {
					t.Fatalf("expected 3 steps in minus pipeline, got %d", len(minus.Pipeline.Steps))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, errs := ParseQuery(tt.query)
			if (len(errs) > 0) != tt.wantErr {
				t.Fatalf("wantErr=%v, got errs: %v", tt.wantErr, errs)
			}
			if tt.check != nil && !tt.wantErr {
				tt.check(t, q)
			}
		})
	}
}

func TestParseQueryCombinators(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
		check   func(t *testing.T, q *Query)
	}{
		{
			name:    "union",
			query:   "(start A | flow out) union (start B | flow out)",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				comb := q.Root.(*CombinatorExpr)
				if comb.Type != "union" {
					t.Errorf("expected type=union, got %s", comb.Type)
				}
				if comb.Left == nil || comb.Right == nil {
					t.Fatal("expected Left and Right pipelines")
				}
			},
		},
		{
			name:    "intersect",
			query:   "start A union start B intersect start C",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				// Parse result depends on operator precedence
				// Both union and intersect have same precedence, left-associative
				// So: (start A union start B) intersect start C
				comb := q.Root.(*CombinatorExpr)
				// The root should be an intersect with left being union result
				if comb.Type != "intersect" {
					t.Errorf("expected root type=intersect, got %s", comb.Type)
				}
			},
		},
		{
			name:    "minus combinator",
			query:   "(start A | flow out) union (start B | flow out) minus (start C)",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				comb := q.Root.(*CombinatorExpr)
				if comb.Type != "minus" {
					t.Errorf("expected root type=minus, got %s", comb.Type)
				}
			},
		},
		{
			name:    "parenthesized pipelines",
			query:   "(start AuthService | flow out recursive) union (start BillingService | flow out recursive)",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				comb := q.Root.(*CombinatorExpr)
				if comb.Type != "union" {
					t.Errorf("expected type=union, got %s", comb.Type)
				}
			},
		},
		{
			name:    "parenthesized single pipeline",
			query:   "(start A)",
			wantErr: false,
			check: func(t *testing.T, q *Query) {
				pipe := q.Root.(*Pipeline)
				if len(pipe.Steps) != 1 {
					t.Fatalf("expected 1 step, got %d", len(pipe.Steps))
				}
				start := pipe.Steps[0].(*StartStep)
				if len(start.NodeIDs) != 1 || start.NodeIDs[0] != "A" {
					t.Errorf("expected NodeIDs=[A], got %v", start.NodeIDs)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, errs := ParseQuery(tt.query)
			if (len(errs) > 0) != tt.wantErr {
				t.Fatalf("wantErr=%v, got errs: %v", tt.wantErr, errs)
			}
			if tt.check != nil && !tt.wantErr {
				tt.check(t, q)
			}
		})
	}
}

func TestParseQueryRoundTrip(t *testing.T) {
	tests := []string{
		"start A",
		"start A | flow out",
		"start A | flow out recursive",
		"start A | flow in *",
		"start A | flow both",
		`start A | flow out where edge.color = "Blue"`,
		"start A | where status = \"active\"",
		"start A | flow out | where critical = true",
		"start A | flow out | where count != 5",
		"start A | flow out | where count < 10",
		"start A | flow out | where priority >= 5",
		"(start A | flow out) union (start B | flow out)",
		"start A | flow out recursive | where critical = true",
		"(start A)",
	}

	for _, query := range tests {
		t.Run(query, func(t *testing.T) {
			q1, errs := ParseQuery(query)
			if len(errs) > 0 {
				t.Fatalf("failed to parse: %v", errs)
			}

			serialized := SerializeQuery(q1)
			if serialized == "" {
				t.Fatal("serialization produced empty string")
			}

			q2, errs := ParseQuery(serialized)
			if len(errs) > 0 {
				t.Fatalf("failed to parse serialized query: %v", errs)
			}

			reserialize := SerializeQuery(q2)
			if serialized != reserialize {
				t.Errorf("round-trip mismatch:\noriginal serialized:  %q\nreserialized:         %q", serialized, reserialize)
			}
		})
	}
}

func TestParseQueryComplexExamples(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "spec example 1",
			query:   `start "AuthService" | flow out recursive | where critical = true`,
			wantErr: false,
		},
		{
			name:    "spec example 2",
			query:   `(start "AuthService" | flow out recursive) union (start "BillingService" | flow out recursive)`,
			wantErr: false,
		},
		{
			name:    "spec example 3",
			query:   `start "D" | flow out where edge.color = "Blue" recursive`,
			wantErr: false,
		},
		{
			name:    "complex pipeline",
			query:   "start A, B, C | flow out recursive | where type != \"archived\" | minus (start X | flow both)",
			wantErr: false,
		},
		{
			name:    "nested union",
			query:   "(start A | flow out) union (start B | flow in) union (start C | flow both)",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, errs := ParseQuery(tt.query)
			if (len(errs) > 0) != tt.wantErr {
				t.Fatalf("wantErr=%v, got errs: %v", tt.wantErr, errs)
			}
			if !tt.wantErr && q == nil {
				t.Fatal("expected non-nil query")
			}
		})
	}
}

func TestParseQueryEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "empty query",
			query:   "",
			wantErr: true,
		},
		{
			name:    "just whitespace",
			query:   "   \t\n  ",
			wantErr: true,
		},
		{
			name:    "missing direction in flow",
			query:   "start A | flow",
			wantErr: true,
		},
		{
			name:    "invalid direction",
			query:   "start A | flow sideways",
			wantErr: true,
		},
		{
			name:    "filter without attribute",
			query:   "start A | where = 5",
			wantErr: true,
		},
		{
			name:    "trailing tokens after pipeline",
			query:   "start A foo",
			wantErr: true,
		},
		{
			name:    "trailing tokens after parenthesized pipeline",
			query:   "(start A) bar",
			wantErr: true,
		},
		{
			name:    "trailing tokens after combinator",
			query:   "start A union start B extra",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := ParseQuery(tt.query)
			if (len(errs) > 0) != tt.wantErr {
				t.Errorf("wantErr=%v, got errs: %v", tt.wantErr, errs)
			}
		})
	}
}

func TestSerializeValueTypes(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected string
	}{
		{"hello", `"hello"`},
		{"with\"quote", `"with\"quote"`},
		{3.14, "3.14"},
		{5.0, "5"},
		{true, "true"},
		{false, "false"},
	}

	for _, tt := range tests {
		result := serializeValue(tt.value)
		if result != tt.expected {
			t.Errorf("serializeValue(%v) = %q, want %q", tt.value, result, tt.expected)
		}
	}
}

func BenchmarkParseQuery(b *testing.B) {
	query := `start "AuthService" | flow out recursive | where critical = true`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseQuery(query)
	}
}

func BenchmarkSerializeQuery(b *testing.B) {
	query := `start "AuthService" | flow out recursive | where critical = true`
	q, _ := ParseQuery(query)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SerializeQuery(q)
	}
}
