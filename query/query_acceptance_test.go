package query

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/dnnrly/gsl-lang"
)

// AcceptanceTest represents a complete end-to-end test case for the query executor
// Follows the pattern: Input Graph → Query → Expected Output Graph
type AcceptanceTest struct {
	// Test metadata
	Name    string // Test case name
	Section string // SPEC section (e.g., "Section 4: Start Step")

	// Inputs
	InputGraph string // GSL defining the graph to query
	Query      string // Query in GQL v1.0 syntax

	// Expected output
	ExpectedGraph string // GSL defining the expected result

	// Optional error expectation
	ExpectError *QueryError
}

// Run executes the acceptance test
func (at *AcceptanceTest) Run(t *testing.T) {
	// Parse input graph
	inputGraph, warns, err := gsl.Parse(strings.NewReader(at.InputGraph))
	if err != nil {
		t.Fatalf("failed to parse input graph: %v", err)
	}
	if len(warns) > 0 {
		t.Logf("warnings parsing input graph: %v", warns)
	}

	// Parse query
	q, queryErrs := ParseQuery(at.Query)
	if len(queryErrs) > 0 {
		t.Fatalf("failed to parse query: %v", queryErrs)
	}

	// Execute query
	result, err := EvaluateQuery(inputGraph, q)

	// Check error expectation
	if at.ExpectError != nil {
		if err == nil {
			t.Errorf("expected error %s, got nil", at.ExpectError.Type)
		} else if queryErr, ok := err.(*QueryError); !ok {
			t.Errorf("expected QueryError, got %T", err)
		} else if queryErr.Type != at.ExpectError.Type {
			t.Errorf("expected error %s, got %s", at.ExpectError.Type, queryErr.Type)
		}
		return
	}

	if err != nil {
		t.Fatalf("execution error: %v", err)
	}

	// Parse expected output graph
	expectedGraph, _, err := gsl.Parse(strings.NewReader(at.ExpectedGraph))
	if err != nil {
		t.Fatalf("failed to parse expected graph: %v", err)
	}

	// Compare result against expected
	if !graphsEqual(result, expectedGraph) {
		resultGSL := result.ToGSL()
		t.Errorf("result mismatch\nGot:\n%s\nExpected:\n%s", resultGSL, at.ExpectedGraph)
	}
}

// graphsEqual compares two query results for equality (set semantics)
func graphsEqual(result *QueryResult, expected *gsl.Graph) bool {
	// Check node count
	if len(result.Nodes) != len(expected.Nodes) {
		return false
	}

	// Check all nodes present
	for id := range expected.Nodes {
		if _, ok := result.Nodes[id]; !ok {
			return false
		}
	}

	// Check edge count
	if len(result.Edges) != len(expected.Edges) {
		return false
	}

	// Check all edges present
	type edgeKey struct{ from, to string }
	expectedEdges := make(map[edgeKey]bool)
	for _, e := range expected.Edges {
		expectedEdges[edgeKey{e.From, e.To}] = true
	}

	for _, e := range result.Edges {
		if !expectedEdges[edgeKey{e.From, e.To}] {
			return false
		}
	}

	return true
}

// ============================================================================
// Shared Query Executor Utilities
// ============================================================================

// QueryResult represents the output of query evaluation
type QueryResult struct {
	Nodes map[string]*gsl.Node // All nodes in the result subgraph
	Edges []*gsl.Edge           // All edges in the result subgraph
}

// NodeIDs returns the sorted list of node IDs
func (r *QueryResult) NodeIDs() []string {
	ids := make([]string, 0, len(r.Nodes))
	for id := range r.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// ToGSL serializes the query result to GSL format
// This allows visual inspection of results and round-trip validation
func (r *QueryResult) ToGSL() string {
	var buf strings.Builder

	// Write nodes (sorted for determinism)
	nodeIDs := r.NodeIDs()
	for _, id := range nodeIDs {
		buf.WriteString("node ")
		buf.WriteString(id)

		node := r.Nodes[id]
		if len(node.Attributes) > 0 {
			buf.WriteString(" [")
			first := true
			for k, v := range node.Attributes {
				if !first {
					buf.WriteString(", ")
				}
				first = false
				buf.WriteString(k)
				buf.WriteString("=")
				serializeAttrValue(&buf, v)
			}
			buf.WriteString("]")
		}
		buf.WriteString("\n")
	}

	// Write edges (sorted for determinism)
	type edgeKey struct{ from, to string }
	edgeMap := make(map[edgeKey]*gsl.Edge)
	for _, e := range r.Edges {
		edgeMap[edgeKey{e.From, e.To}] = e
	}

	var sortedEdges []edgeKey
	for k := range edgeMap {
		sortedEdges = append(sortedEdges, k)
	}
	sort.Slice(sortedEdges, func(i, j int) bool {
		if sortedEdges[i].from != sortedEdges[j].from {
			return sortedEdges[i].from < sortedEdges[j].from
		}
		return sortedEdges[i].to < sortedEdges[j].to
	})

	for _, k := range sortedEdges {
		e := edgeMap[k]
		buf.WriteString(e.From)
		buf.WriteString(" -> ")
		buf.WriteString(e.To)

		if len(e.Attributes) > 0 {
			buf.WriteString(" [")
			first := true
			for k, v := range e.Attributes {
				if !first {
					buf.WriteString(", ")
				}
				first = false
				buf.WriteString(k)
				buf.WriteString("=")
				serializeAttrValue(&buf, v)
			}
			buf.WriteString("]")
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

func serializeAttrValue(buf *strings.Builder, v interface{}) {
	switch val := v.(type) {
	case string:
		buf.WriteString(`"`)
		buf.WriteString(val)
		buf.WriteString(`"`)
	case float64:
		buf.WriteString(fmt.Sprintf("%g", val))
	case bool:
		buf.WriteString(fmt.Sprintf("%v", val))
	default:
		buf.WriteString(fmt.Sprintf("%v", val))
	}
}

// EvaluateQuery evaluates a query against a graph
// TODO: Implement per SPEC Section 3-9
func EvaluateQuery(g *gsl.Graph, q *Query) (*QueryResult, *QueryError) {
	panic("EvaluateQuery not yet implemented")
}

// ============================================================================
// Acceptance Tests - SPEC Section 4: Start Step
// ============================================================================

func TestAcceptanceStartSingleNode(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Single node start",
		Section: "Section 4: Start Step",
		InputGraph: `
node A
node B
node C
node D
A -> B
B -> C
C -> D
`,
		Query: `start "A"`,
		ExpectedGraph: `
node A
`,
	}
	test.Run(t)
}

func TestAcceptanceStartMultipleNodes(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Multiple nodes start",
		Section: "Section 4: Start Step",
		InputGraph: `
node A
node B
node C
node D
A -> B
B -> C
C -> D
`,
		Query: `start "A", "B", "C"`,
		ExpectedGraph: `
node A
node B
node C
A -> B
B -> C
`,
	}
	test.Run(t)
}

func TestAcceptanceStartNonConsecutive(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Non-consecutive nodes",
		Section: "Section 4: Start Step",
		InputGraph: `
node A
node B
node C
node D
A -> B
B -> C
C -> D
`,
		Query: `start "A", "C"`,
		ExpectedGraph: `
node A
node C
`,
	}
	test.Run(t)
}

func TestAcceptanceStartUnknownNode(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Unknown node error",
		Section: "Section 4: Start Step",
		InputGraph: `
node A
node B
`,
		Query: `start "A", "UNKNOWN"`,
		ExpectError: &QueryError{
			Type:    ErrorUnknownNodeID,
			Message: "node UNKNOWN not found",
		},
	}
	test.Run(t)
}

// ============================================================================
// Acceptance Tests - SPEC Section 5: Flow Step
// ============================================================================

func TestAcceptanceFlowOut(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Flow out (non-recursive)",
		Section: "Section 5.3: Non-Recursive Flow",
		InputGraph: `
node A
node B
node C
node D
A -> B
B -> C
C -> D
`,
		Query: `start "A" | flow out`,
		ExpectedGraph: `
node A
node B
A -> B
`,
	}
	test.Run(t)
}

func TestAcceptanceFlowIn(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Flow in (non-recursive)",
		Section: "Section 5.3: Non-Recursive Flow",
		InputGraph: `
node A
node B
node C
node D
A -> B
B -> C
C -> D
`,
		Query: `start "D" | flow in`,
		ExpectedGraph: `
node C
node D
C -> D
`,
	}
	test.Run(t)
}

func TestAcceptanceFlowBoth(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Flow both directions",
		Section: "Section 5.3: Non-Recursive Flow",
		InputGraph: `
node A
node B
node C
node D
A -> B
B -> C
C -> D
`,
		Query: `start "B" | flow both`,
		ExpectedGraph: `
node A
node B
node C
A -> B
B -> C
`,
	}
	test.Run(t)
}

func TestAcceptanceFlowOutRecursive(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Flow out recursive - transitive closure",
		Section: "Section 5.4: Recursive Flow",
		InputGraph: `
node A
node B
node C
node D
A -> B
B -> C
C -> D
`,
		Query: `start "A" | flow out recursive`,
		ExpectedGraph: `
node A
node B
node C
node D
A -> B
B -> C
C -> D
`,
	}
	test.Run(t)
}

func TestAcceptanceFlowInRecursive(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Flow in recursive - all predecessors",
		Section: "Section 5.4: Recursive Flow",
		InputGraph: `
node A
node B
node C
node D
A -> B
B -> C
C -> D
`,
		Query: `start "D" | flow in recursive`,
		ExpectedGraph: `
node A
node B
node C
node D
A -> B
B -> C
C -> D
`,
	}
	test.Run(t)
}

func TestAcceptanceFlowDiamond(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Flow recursive - diamond pattern",
		Section: "Section 5.4: Recursive Flow",
		InputGraph: `
node A
node B
node C
node D
node E
A -> B
A -> C
B -> D
C -> D
D -> E
`,
		Query: `start "A" | flow out recursive`,
		ExpectedGraph: `
node A
node B
node C
node D
node E
A -> B
A -> C
B -> D
C -> D
D -> E
`,
	}
	test.Run(t)
}

// ============================================================================
// Acceptance Tests - SPEC Section 5.2: Edge Filtering
// ============================================================================

func TestAcceptanceEdgeFilter(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Edge filter - type=http",
		Section: "Section 5.2: Edge Filtering",
		InputGraph: `
node A
node B
node C
node D
A -> B [type="http"]
A -> C [type="db"]
B -> D [type="db"]
C -> D [type="http"]
`,
		Query: `start "A" | flow out where edge.type = "http"`,
		ExpectedGraph: `
node A
node B
A -> B [type="http"]
`,
	}
	test.Run(t)
}

func TestAcceptanceEdgeFilterRecursive(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Recursive flow with edge filter",
		Section: "Section 5.2: Edge Filtering",
		InputGraph: `
node A
node B
node C
node D
A -> B [type="http"]
A -> C [type="db"]
B -> D [type="db"]
C -> D [type="http"]
`,
		Query: `start "A" | flow out recursive where edge.type = "db"`,
		ExpectedGraph: `
node A
node C
node D
A -> C [type="db"]
C -> D [type="http"]
`,
	}
	test.Run(t)
}

// ============================================================================
// Acceptance Tests - SPEC Section 6-7: Node Filtering
// ============================================================================

func TestAcceptanceNodeFilter(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Node filter - status=active",
		Section: "Section 6: Node Filter Step",
		InputGraph: `
node A [status="active"]
node B [status="inactive"]
node C [status="active"]
node D [status="inactive"]
A -> B
B -> C
A -> D
`,
		Query: `start "A", "B", "C", "D" | where status = "active"`,
		ExpectedGraph: `
node A [status="active"]
node C [status="active"]
`,
	}
	test.Run(t)
}

func TestAcceptanceNodeFilterAfterFlow(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Flow then filter",
		Section: "Section 6: Node Filter Step",
		InputGraph: `
node A [status="active"]
node B [status="inactive"]
node C [status="active"]
node D [status="inactive"]
A -> B
B -> C
A -> D
`,
		Query: `start "A" | flow out recursive | where status = "active"`,
		ExpectedGraph: `
node A [status="active"]
node C [status="active"]
A -> B
B -> C
`,
	}
	test.Run(t)
}

// ============================================================================
// Acceptance Tests - SPEC Section 8: Combinators
// ============================================================================

func TestAcceptanceUnion(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Union combinator",
		Section: "Section 8.1: Union",
		InputGraph: `
node A
node B
node C
node D
node E
A -> B
A -> C
B -> D
C -> D
D -> E
`,
		Query: `(start "A" | flow out) union (start "D" | flow out)`,
		ExpectedGraph: `
node A
node B
node C
node D
node E
A -> B
A -> C
D -> E
`,
	}
	test.Run(t)
}

func TestAcceptanceIntersect(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Intersect combinator",
		Section: "Section 8.2: Intersect",
		InputGraph: `
node A
node B
node C
node D
A -> B
B -> C
C -> D
`,
		Query: `(start "A", "B", "C") intersect (start "B", "C", "D")`,
		ExpectedGraph: `
node B
node C
B -> C
`,
	}
	test.Run(t)
}

func TestAcceptanceMinus(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Minus combinator",
		Section: "Section 8.3: Minus",
		InputGraph: `
node A
node B
node C
node D
A -> B
B -> C
C -> D
`,
		Query: `(start "A" | flow out recursive) minus (start "D")`,
		ExpectedGraph: `
node A
node B
node C
A -> B
B -> C
`,
	}
	test.Run(t)
}

// ============================================================================
// Acceptance Tests - SPEC Section 3.3: Strict Subgraph Construction
// ============================================================================

func TestAcceptanceSubgraphConstruction(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name: "Strict subgraph construction - all edges included",
		Section: "Section 3.3: Result Subgraph Construction",
		InputGraph: `
node A
node B
node C
A -> B
B -> C
A -> C
`,
		Query: `(start "A") union (start "B", "C")`,
		ExpectedGraph: `
node A
node B
node C
A -> B
B -> C
A -> C
`,
	}
	test.Run(t)
}

// ============================================================================
// Real-World Acceptance Tests
// ============================================================================

func TestAcceptanceAPIDependencies(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Find all dependencies of a service",
		Section: "Real-world: Service dependency graph",
		InputGraph: `
node consumer
node gateway
node userService [critical=true]
node authService [critical=true]
node db [type="postgres"]
node cache [type="redis"]
consumer -> gateway
gateway -> userService
gateway -> authService
userService -> db
userService -> cache
authService -> db
`,
		Query: `start "userService" | flow out recursive`,
		ExpectedGraph: `
node userService [critical=true]
node db [type="postgres"]
node cache [type="redis"]
userService -> db
userService -> cache
`,
	}
	test.Run(t)
}

func TestAcceptanceFindCriticalPath(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Find all critical services in dependency tree",
		Section: "Real-world: Filter for critical services",
		InputGraph: `
node userApi [critical=true]
node authService [critical=true]
node userDb [critical=true]
node cache [critical=false]
node logger [critical=false]
userApi -> authService
userApi -> cache
authService -> userDb
userApi -> logger
`,
		Query: `start "userApi" | flow out recursive | where critical = true`,
		ExpectedGraph: `
node userApi [critical=true]
node authService [critical=true]
node userDb [critical=true]
userApi -> authService
authService -> userDb
`,
	}
	test.Run(t)
}

func TestAcceptanceFindDatabaseAccess(t *testing.T) {
	t.Skip("EvaluateQuery executor not yet implemented")

	test := AcceptanceTest{
		Name:    "Find services that directly access the database",
		Section: "Real-world: Edge filtering for database calls",
		InputGraph: `
node userService
node orderService
node cache
node userDb
node orderDb
userService -> cache [type="cache"]
userService -> userDb [type="db"]
orderService -> orderDb [type="db"]
orderService -> cache [type="cache"]
`,
		Query: `start "userService", "orderService" | flow out where edge.type = "db"`,
		ExpectedGraph: `
node userService
node orderService
node userDb
node orderDb
userService -> userDb [type="db"]
orderService -> orderDb [type="db"]
`,
	}
	test.Run(t)
}
