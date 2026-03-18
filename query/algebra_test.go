package query

import (
	"testing"

	"github.com/dnnrly/gsl-lang"
)

func TestGraphAlgebraUnion(t *testing.T) {
	tests := []struct {
		name     string
		left     *gsl.Graph
		right    *gsl.Graph
		want     *gsl.Graph
		wantErr  bool
		errMsg   string
	}{
		{
			name: "union of two disjoint graphs",
			left: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{"type": "service"}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			right: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"B": {ID: "B", Attributes: map[string]interface{}{"type": "database"}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			want: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{"type": "service"}, Sets: make(map[string]struct{})},
					"B": {ID: "B", Attributes: map[string]interface{}{"type": "database"}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
		},
		{
			name: "union with overlapping nodes (right overwrites)",
			left: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{"team": "payments", "zone": "A"}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			right: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{"team": "fraud"}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			want: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{"team": "fraud", "zone": "A"}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
		},
		{
			name: "union with edges from both graphs",
			left: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{
					{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Sets: make(map[string]*gsl.Set),
			}),
			right: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{
					{From: "B", To: "C", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Sets: make(map[string]*gsl.Set),
			}),
			want: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{
					{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					{From: "B", To: "C", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Sets: make(map[string]*gsl.Set),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &GraphAlgebraExpr{Operator: "+"}
			got := expr.union(tt.left, tt.right)

			if !graphsEqual(got, tt.want) {
				t.Errorf("union() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGraphAlgebraIntersection(t *testing.T) {
	tests := []struct {
		name string
		left *gsl.Graph
		right *gsl.Graph
		want *gsl.Graph
	}{
		{
			name: "intersection of graphs with one common node",
			left: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			right: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			want: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
		},
		{
			name: "intersection with shared edges",
			left: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{
					{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					{From: "B", To: "C", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Sets: make(map[string]*gsl.Set),
			}),
			right: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{
					{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Sets: make(map[string]*gsl.Set),
			}),
			want: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{
					{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Sets: make(map[string]*gsl.Set),
			}),
		},
		{
			name: "intersection with disjoint graphs",
			left: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			right: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			want: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &GraphAlgebraExpr{Operator: "&"}
			got := expr.intersection(tt.left, tt.right)

			if !graphsEqual(got, tt.want) {
				t.Errorf("intersection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGraphAlgebraDifference(t *testing.T) {
	tests := []struct {
		name string
		left *gsl.Graph
		right *gsl.Graph
		want *gsl.Graph
	}{
		{
			name: "difference of overlapping graphs",
			left: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			right: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			want: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
		},
		{
			name: "difference with edges",
			left: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{
					{From: "A", To: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					{From: "B", To: "C", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Sets: make(map[string]*gsl.Set),
			}),
			right: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			want: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
		},
		{
			name: "difference with disjoint graphs",
			left: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			right: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			want: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &GraphAlgebraExpr{Operator: "-"}
			got := expr.difference(tt.left, tt.right)

			if !graphsEqual(got, tt.want) {
				t.Errorf("difference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGraphAlgebraSymmetricDifference(t *testing.T) {
	tests := []struct {
		name string
		left *gsl.Graph
		right *gsl.Graph
		want *gsl.Graph
	}{
		{
			name: "symmetric difference of overlapping graphs",
			left: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			right: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			want: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"C": {ID: "C", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
		},
		{
			name: "symmetric difference with disjoint graphs",
			left: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			right: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			want: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &GraphAlgebraExpr{Operator: "^"}
			got := expr.symmetricDifference(tt.left, tt.right)

			if !graphsEqual(got, tt.want) {
				t.Errorf("symmetricDifference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseGraphAlgebra(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *GraphAlgebraExpr
		wantErr bool
		errMsg  string
	}{
		{
			name:  "simple union",
			input: "* + G1",
			want: &GraphAlgebraExpr{
				LeftRef:  "*",
				RightRef: "G1",
				Operator: "+",
			},
		},
		{
			name:  "intersection of named graphs",
			input: "G1 & G2",
			want: &GraphAlgebraExpr{
				LeftRef:  "G1",
				RightRef: "G2",
				Operator: "&",
			},
		},
		{
			name:  "difference",
			input: "G1 - G2",
			want: &GraphAlgebraExpr{
				LeftRef:  "G1",
				RightRef: "G2",
				Operator: "-",
			},
		},
		{
			name:  "symmetric difference",
			input: "G1 ^ G2",
			want: &GraphAlgebraExpr{
				LeftRef:  "G1",
				RightRef: "G2",
				Operator: "^",
			},
		},
		{
			name:    "invalid: no operator",
			input:   "G1 G2",
			wantErr: true,
		},
		{
			name:    "invalid left reference",
			input:   "invalid_name + G1",
			wantErr: true,
		},
		{
			name:    "invalid right reference",
			input:   "G1 + invalid_name",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := newExpressionParser(tt.input)
			got, err := parser.parseGraphAlgebra()

			if (err != nil) != tt.wantErr {
				t.Errorf("parseGraphAlgebra() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				expr := got.(*GraphAlgebraExpr)
				if expr.LeftRef != tt.want.LeftRef || expr.RightRef != tt.want.RightRef || expr.Operator != tt.want.Operator {
					t.Errorf("parseGraphAlgebra() = %v, want %v", expr, tt.want)
				}
			}
		})
	}
}

func TestGraphAlgebraIntegration(t *testing.T) {
	// Test full pipeline execution with algebra
	// Helper function to build bindings map
	buildBindings := func(graphs ...struct {
		key string
		g   *gsl.Graph
	}) map[string]*gsl.Graph {
		result := make(map[string]*gsl.Graph)
		for _, item := range graphs {
			result[item.key] = item.g
		}
		return result
	}

	tests := []struct {
		name     string
		query    string
		input    *gsl.Graph
		bindings map[string]*gsl.Graph
		wantErr  bool
	}{
		{
			name:  "union via pipeline",
			query: "* + G1",
			input: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			bindings: buildBindings(
				struct {
					key string
					g   *gsl.Graph
				}{
					key: "G1",
					g: newTestGraph(testGraphInput{
						Nodes: map[string]*gsl.Node{
							"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
						},
						Edges: []*gsl.Edge{},
						Sets:  make(map[string]*gsl.Set),
					}),
				},
			),
		},
		{
			name:  "intersection via pipeline",
			query: "G1 & G2",
			input: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{},
				Edges: []*gsl.Edge{},
				Sets:  make(map[string]*gsl.Set),
			}),
			bindings: buildBindings(
				struct {
					key string
					g   *gsl.Graph
				}{
					key: "G1",
					g: newTestGraph(testGraphInput{
						Nodes: map[string]*gsl.Node{
							"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
						},
						Edges: []*gsl.Edge{},
						Sets:  make(map[string]*gsl.Set),
					}),
				},
				struct {
					key string
					g   *gsl.Graph
				}{
					key: "G2",
					g: newTestGraph(testGraphInput{
						Nodes: map[string]*gsl.Node{
							"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
						},
						Edges: []*gsl.Edge{},
						Sets:  make(map[string]*gsl.Set),
					}),
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewQueryParser(tt.query)
			q, err := parser.Parse()
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			ctx := &QueryContext{
				InputGraph:  tt.input,
				NamedGraphs: tt.bindings,
			}

			result, err := q.Execute(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if _, ok := result.(GraphValue); !ok {
					t.Errorf("Execute() returned non-Graph value")
				}
			}
		})
	}
}

// Helper to compare graphs structurally
func graphsEqual(left *gsl.Graph, right *gsl.Graph) bool {
	// Compare node count
	leftNodes := left.GetNodes()
	rightNodes := right.GetNodes()
	if len(leftNodes) != len(rightNodes) {
		return false
	}

	// Compare nodes
	for id, node := range leftNodes {
		rightNode, exists := rightNodes[id]
		if !exists {
			return false
		}
		if !nodesEqual(node, rightNode) {
			return false
		}
	}

	// Compare edge count
	leftEdges := left.GetEdges()
	rightEdges := right.GetEdges()
	if len(leftEdges) != len(rightEdges) {
		return false
	}

	// Compare edges
	for i, edge := range leftEdges {
		if !edgesEqual(edge, rightEdges[i]) {
			return false
		}
	}

	// Compare sets
	leftSets := left.GetSets()
	rightSets := right.GetSets()
	return len(leftSets) == len(rightSets)
}

func nodesEqual(left *gsl.Node, right *gsl.Node) bool {
	if left.ID != right.ID {
		return false
	}

	if len(left.Attributes) != len(right.Attributes) {
		return false
	}

	for k, v := range left.Attributes {
		rv, exists := right.Attributes[k]
		if !exists || v != rv {
			return false
		}
	}

	return true
}

func edgesEqual(left *gsl.Edge, right *gsl.Edge) bool {
	if left.From != right.From || left.To != right.To {
		return false
	}

	if len(left.Attributes) != len(right.Attributes) {
		return false
	}

	for k, v := range left.Attributes {
		rv, exists := right.Attributes[k]
		if !exists || v != rv {
			return false
		}
	}

	return true
}

func TestGraphAlgebraWithSets(t *testing.T) {
	tests := []struct {
		name string
		op   string
		left *gsl.Graph
		right *gsl.Graph
		want *gsl.Graph
	}{
		{
			name: "union preserves sets from both graphs",
			op:   "+",
			left: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets: map[string]*gsl.Set{
					"S1": {ID: "S1", Attributes: map[string]interface{}{"color": "red"}},
				},
			}),
			right: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets: map[string]*gsl.Set{
					"S2": {ID: "S2", Attributes: map[string]interface{}{"size": "10"}},
				},
			}),
			want: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets: map[string]*gsl.Set{
					"S1": {ID: "S1", Attributes: map[string]interface{}{"color": "red"}},
					"S2": {ID: "S2", Attributes: map[string]interface{}{"size": "10"}},
				},
			}),
		},
		{
			name: "intersection only includes shared sets",
			op:   "&",
			left: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets: map[string]*gsl.Set{
					"S1": {ID: "S1", Attributes: map[string]interface{}{}},
					"S2": {ID: "S2", Attributes: map[string]interface{}{}},
				},
			}),
			right: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets: map[string]*gsl.Set{
					"S1": {ID: "S1", Attributes: map[string]interface{}{}},
					"S3": {ID: "S3", Attributes: map[string]interface{}{}},
				},
			}),
			want: newTestGraph(testGraphInput{
				Nodes: map[string]*gsl.Node{
					"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
				},
				Edges: []*gsl.Edge{},
				Sets: map[string]*gsl.Set{
					"S1": {ID: "S1", Attributes: map[string]interface{}{}},
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &GraphAlgebraExpr{Operator: tt.op}
			var got *gsl.Graph
			switch tt.op {
			case "+":
				got = expr.union(tt.left, tt.right)
			case "&":
				got = expr.intersection(tt.left, tt.right)
			case "-":
				got = expr.difference(tt.left, tt.right)
			case "^":
				got = expr.symmetricDifference(tt.left, tt.right)
			}

			if !graphsEqual(got, tt.want) {
				t.Errorf("%s() = %v, want %v", tt.op, got, tt.want)
			}
		})
	}
}

func TestGraphAlgebraResolveGraph(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		ctx     *QueryContext
		wantErr bool
	}{
		{
			name: "resolve wildcard to input graph",
			ref:  "*",
			ctx: &QueryContext{
				InputGraph: newTestGraph(testGraphInput{
					Nodes: map[string]*gsl.Node{
						"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					},
					Edges: []*gsl.Edge{},
					Sets:  make(map[string]*gsl.Set),
				}),
				NamedGraphs: make(map[string]*gsl.Graph),
			},
		},
		{
			name: "resolve named graph",
			ref:  "G1",
			ctx: &QueryContext{
				InputGraph: newTestGraph(testGraphInput{
					Nodes: map[string]*gsl.Node{},
					Edges: []*gsl.Edge{},
					Sets:  make(map[string]*gsl.Set),
				}),
				NamedGraphs: map[string]*gsl.Graph{
					"G1": newTestGraph(testGraphInput{
						Nodes: map[string]*gsl.Node{
							"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
						},
						Edges: []*gsl.Edge{},
						Sets:  make(map[string]*gsl.Set),
					}),
				},
			},
		},
		{
			name:    "error: graph not found",
			ref:     "G_MISSING",
			ctx:     &QueryContext{InputGraph: newTestGraph(testGraphInput{}), NamedGraphs: make(map[string]*gsl.Graph)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &GraphAlgebraExpr{}
			got, err := expr.resolveGraph(tt.ctx, tt.ref)

			if (err != nil) != tt.wantErr {
				t.Errorf("resolveGraph() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Errorf("resolveGraph() returned nil graph")
			}
		})
	}
}

func TestGraphAlgebraApply(t *testing.T) {
	tests := []struct {
		name    string
		expr    *GraphAlgebraExpr
		ctx     *QueryContext
		wantErr bool
	}{
		{
			name: "apply union expression",
			expr: &GraphAlgebraExpr{
				LeftRef:  "*",
				RightRef: "G1",
				Operator: "+",
			},
			ctx: &QueryContext{
				InputGraph: newTestGraph(testGraphInput{
					Nodes: map[string]*gsl.Node{
						"A": {ID: "A", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
					},
					Edges: []*gsl.Edge{},
					Sets:  make(map[string]*gsl.Set),
				}),
				NamedGraphs: map[string]*gsl.Graph{
					"G1": newTestGraph(testGraphInput{
						Nodes: map[string]*gsl.Node{
							"B": {ID: "B", Attributes: map[string]interface{}{}, Sets: make(map[string]struct{})},
						},
						Edges: []*gsl.Edge{},
						Sets:  make(map[string]*gsl.Set),
					}),
				},
			},
		},
		{
			name: "error: unknown operator",
			expr: &GraphAlgebraExpr{
				LeftRef:  "*",
				RightRef: "G1",
				Operator: "INVALID",
			},
			ctx: &QueryContext{
				InputGraph:  newTestGraph(testGraphInput{}),
				NamedGraphs: make(map[string]*gsl.Graph),
			},
			wantErr: true,
		},
		{
			name: "error: left graph not found",
			expr: &GraphAlgebraExpr{
				LeftRef:  "G_MISSING",
				RightRef: "G1",
				Operator: "+",
			},
			ctx: &QueryContext{
				InputGraph:  newTestGraph(testGraphInput{}),
				NamedGraphs: make(map[string]*gsl.Graph),
			},
			wantErr: true,
		},
		{
			name: "error: right graph not found",
			expr: &GraphAlgebraExpr{
				LeftRef:  "*",
				RightRef: "G_MISSING",
				Operator: "+",
			},
			ctx: &QueryContext{
				InputGraph:  newTestGraph(testGraphInput{}),
				NamedGraphs: make(map[string]*gsl.Graph),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.expr.Apply(tt.ctx, GraphValue{tt.ctx.InputGraph})

			if (err != nil) != tt.wantErr {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidGraphRef(t *testing.T) {
	tests := []struct {
		ref  string
		want bool
	}{
		{ref: "*", want: true},
		{ref: "G1", want: true},
		{ref: "G_NAME", want: true},
		{ref: "ABC123", want: true},
		{ref: "g1", want: false}, // lowercase
		{ref: "invalid-name", want: false}, // dash
		{ref: "1ABC", want: false}, // starts with digit
		{ref: "", want: false}, // empty
	}

	for _, tt := range tests {
		got := isValidGraphRef(tt.ref)
		if got != tt.want {
			t.Errorf("isValidGraphRef(%q) = %v, want %v", tt.ref, got, tt.want)
		}
	}
}
