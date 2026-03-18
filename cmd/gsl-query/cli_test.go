package main

import (
	"bytes"
	"os"
	"testing"

	gsl "github.com/dnnrly/gsl-lang"
)

func TestExecuteWithStdin(t *testing.T) {
	gslInput := `set critical
node API @critical
node DB @critical
node Cache

API -> DB
API -> Cache
DB -> Cache
`

	// Save original stdin
	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	// Create pipe for stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	os.Stdin = r

	// Write test data
	go func() {
		_, _ = w.WriteString(gslInput)
		w.Close()
	}()

	// Capture stdout
	origStdout := os.Stdout
	defer func() { os.Stdout = origStdout }()

	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	cfg := &Config{
		InputFile:  "",
		OutputFile: "",
		Query:      "subgraph in critical",
		InputName:  "<stdin>",
	}

	err = Execute(cfg)
	wOut.Close()

	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Read output
	var output bytes.Buffer
	_, _ = output.ReadFrom(rOut)

	result := output.String()

	// Verify output contains critical nodes
	if !contains(result, "node API") {
		t.Errorf("output missing API node")
	}
	if !contains(result, "node DB") {
		t.Errorf("output missing DB node")
	}
	if !contains(result, "@critical") {
		t.Errorf("output missing @critical set membership")
	}
}

func TestExecuteWithFile(t *testing.T) {
	gslInput := `set important
node ServiceA [status="active"] @important
node ServiceB [status="inactive"]
node ServiceC @important

ServiceA -> ServiceB
ServiceB -> ServiceC
`

	// Create temp input file
	inputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp input file: %v", err)
	}
	defer os.Remove(inputFile.Name())

	_, err = inputFile.WriteString(gslInput)
	if err != nil {
		t.Fatalf("failed to write input: %v", err)
	}
	inputFile.Close()

	// Create temp output file
	outputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp output file: %v", err)
	}
	defer os.Remove(outputFile.Name())
	outputFile.Close()

	cfg := &Config{
		InputFile:  inputFile.Name(),
		OutputFile: outputFile.Name(),
		Query:      "subgraph in important",
		InputName:  inputFile.Name(),
	}

	err = Execute(cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Read output file
	output, err := os.ReadFile(outputFile.Name())
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	result := string(output)

	// Verify output
	if !contains(result, "ServiceA") {
		t.Errorf("output missing ServiceA node")
	}
	if !contains(result, "ServiceC") {
		t.Errorf("output missing ServiceC node")
	}
	if !contains(result, "@important") {
		t.Errorf("output missing @important set membership")
	}
}

func TestExecuteSubgraphFilter(t *testing.T) {
	gslInput := `set payments
set critical

node API [team="payments"] @payments @critical
node AuthSvc [team="payments"] @payments
node LegacySvc [team="legacy"]

API -> AuthSvc
API -> LegacySvc
AuthSvc -> LegacySvc
`

	// Create temp input file
	inputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp input file: %v", err)
	}
	defer os.Remove(inputFile.Name())

	_, _ = inputFile.WriteString(gslInput)
	inputFile.Close()

	// Create temp output file
	outputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp output file: %v", err)
	}
	defer os.Remove(outputFile.Name())
	outputFile.Close()

	// Query for payment services
	cfg := &Config{
		InputFile:  inputFile.Name(),
		OutputFile: outputFile.Name(),
		Query:      "subgraph node.team = \"payments\"",
		InputName:  inputFile.Name(),
	}

	err = Execute(cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output, _ := os.ReadFile(outputFile.Name())
	result := string(output)

	// Should include API and AuthSvc
	if !contains(result, "API") {
		t.Errorf("missing API in filtered result")
	}
	if !contains(result, "AuthSvc") {
		t.Errorf("missing AuthSvc in filtered result")
	}

	// Should NOT include LegacySvc (only as orphan, but it will be gone because no edges to it)
	if contains(result, "node LegacySvc") {
		t.Errorf("should not have LegacySvc in output (not in payments team)")
	}
}

func TestExecuteRemoveEdges(t *testing.T) {
	gslInput := `node A
node B
node C

A -> B [status="deprecated"]
A -> C [status="active"]
B -> C [status="deprecated"]
`

	// Create temp input file
	inputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp input file: %v", err)
	}
	defer os.Remove(inputFile.Name())

	_, _ = inputFile.WriteString(gslInput)
	inputFile.Close()

	// Create temp output file
	outputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp output file: %v", err)
	}
	defer os.Remove(outputFile.Name())
	outputFile.Close()

	// Remove deprecated edges
	cfg := &Config{
		InputFile:  inputFile.Name(),
		OutputFile: outputFile.Name(),
		Query:      "remove edge where edge.status = \"deprecated\"",
		InputName:  inputFile.Name(),
	}

	err = Execute(cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output, _ := os.ReadFile(outputFile.Name())
	result := string(output)

	// Parse output to verify edges
	graph, _, err := gsl.Parse(bytes.NewReader(output))
	if err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	// Should have 1 edge (A->C)
	edges := graph.GetEdges()
	if len(edges) != 1 {
		t.Errorf("expected 1 edge after removal, got %d", len(edges))
	}

	// Verify A->C is active
	if contains(result, "A->C") {
		if !contains(result, `status="active"`) {
			t.Errorf("A->C edge missing status attribute")
		}
	}
}

func TestExecuteMakePipeline(t *testing.T) {
	gslInput := `node API [team="platform"]
node DB [team="platform"]

API -> DB
`

	// Create temp input file
	inputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp input file: %v", err)
	}
	defer os.Remove(inputFile.Name())

	_, _ = inputFile.WriteString(gslInput)
	inputFile.Close()

	// Create temp output file
	outputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp output file: %v", err)
	}
	defer os.Remove(outputFile.Name())
	outputFile.Close()

	// Mark platform team as reviewed
	cfg := &Config{
		InputFile:  inputFile.Name(),
		OutputFile: outputFile.Name(),
		Query:      "make node.reviewed = \"true\" where node.team = \"platform\"",
		InputName:  inputFile.Name(),
	}

	err = Execute(cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output, _ := os.ReadFile(outputFile.Name())

	// Parse and verify attributes
	graph, _, _ := gsl.Parse(bytes.NewReader(output))

	// Check that all platform team nodes have reviewed attribute
	nodes := graph.GetNodes()
	for _, node := range nodes {
		if team, ok := node.Attributes["team"]; ok && team == "platform" {
			reviewed, ok := node.Attributes["reviewed"]
			if !ok || reviewed != "true" {
				t.Errorf("node %s missing or incorrect reviewed attribute", node.ID)
			}
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
