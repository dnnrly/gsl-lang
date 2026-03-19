// +build integration

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	gsl "github.com/dnnrly/gsl-lang"
)

// getBinaryPath returns the path to the gsl-query binary
func getBinaryPath(t *testing.T) string {
	// Try common locations
	paths := []string{
		"./gsl-query",
		"./tmp/gsl-query",
		"../../tmp/gsl-query",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// If not found, try to build it
	t.Logf("Binary not found in expected locations, building...")
	cmd := exec.Command("go", "build", "-o", "tmp/gsl-query", "./cmd/gsl-query")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build gsl-query: %v\nOutput: %s", err, string(output))
	}

	return "tmp/gsl-query"
}

func TestCLIIntegrationStdin(t *testing.T) {
	// Create temp file for query
	queryFile, err := os.CreateTemp("", "query*.txt")
	if err != nil {
		t.Fatalf("failed to create temp query file: %v", err)
	}
	defer os.Remove(queryFile.Name())

	queryFile.WriteString("subgraph in critical")
	queryFile.Close()

	// Create temp input GSL file
	inputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp input file: %v", err)
	}
	defer os.Remove(inputFile.Name())

	inputFile.WriteString(`set critical
node API @critical
node DB @critical
node Cache

API -> DB
API -> Cache
DB -> Cache
`)
	inputFile.Close()

	// Create temp output file
	outputFile, err := os.CreateTemp("", "output*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp output file: %v", err)
	}
	defer os.Remove(outputFile.Name())
	outputFile.Close()

	// Run CLI command
	binaryPath := getBinaryPath(t)
	cmd := exec.Command(binaryPath, "-i", inputFile.Name(), "-o", outputFile.Name(), "-f", queryFile.Name())
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, string(output))
	}

	// Verify output
	content, err := os.ReadFile(outputFile.Name())
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	if len(content) == 0 {
		t.Fatalf("output file is empty")
	}

	// Parse output to verify it's valid GSL
	result, pErr := gsl.Parse(bytes.NewReader(content))
	if pErr != nil {
		t.Fatalf("failed to parse query result as GSL: %v", pErr)
	}

	// Should have at least API and DB nodes
	if len(result.GetNodes()) < 2 {
		t.Errorf("expected at least 2 nodes, got %d", len(result.GetNodes()))
	}

	// Should have edges between nodes
	if len(result.GetEdges()) == 0 {
		t.Errorf("expected edges in result, got 0")
	}
}

func TestCLIIntegrationQueryArg(t *testing.T) {
	// Create temp input file
	inputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp input file: %v", err)
	}
	defer os.Remove(inputFile.Name())

	inputFile.WriteString(`set payments
node PaymentSvc [team="payments"] @payments
node LegacySvc [team="legacy"]

PaymentSvc -> LegacySvc
`)
	inputFile.Close()

	// Create temp output file
	outputFile, err := os.CreateTemp("", "output*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp output file: %v", err)
	}
	defer os.Remove(outputFile.Name())
	outputFile.Close()

	// Run CLI with query argument
	binaryPath := getBinaryPath(t)
	cmd := exec.Command(binaryPath, "-i", inputFile.Name(), "-o", outputFile.Name(), "subgraph in payments")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, string(output))
	}

	// Verify output
	content, _ := os.ReadFile(outputFile.Name())
	result, _ := gsl.Parse(bytes.NewReader(content))

	// Should have PaymentSvc
	if len(result.GetNodes()) == 0 {
		t.Errorf("expected nodes in result, got 0")
	}

	foundPaymentSvc := false
	for _, node := range result.GetNodes() {
		if node.ID == "PaymentSvc" {
			foundPaymentSvc = true
			break
		}
	}

	if !foundPaymentSvc {
		t.Errorf("expected PaymentSvc in result")
	}
}

func TestCLIIntegrationPipeline(t *testing.T) {
	// Create temp input file
	inputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp input file: %v", err)
	}
	defer os.Remove(inputFile.Name())

	inputFile.WriteString(`set critical
node API [status="active"] @critical
node DB [status="active"] @critical
node Cache [status="deprecated"]

API -> DB [status="active"]
API -> Cache [status="deprecated"]
DB -> Cache [status="deprecated"]
`)
	inputFile.Close()

	// Create temp output file
	outputFile, err := os.CreateTemp("", "output*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp output file: %v", err)
	}
	defer os.Remove(outputFile.Name())
	outputFile.Close()

	// Complex pipeline: get critical nodes and remove deprecated edges
	query := "subgraph in critical | remove edge where edge.status = \"deprecated\""
	binaryPath := getBinaryPath(t)
	cmd := exec.Command(binaryPath, "-i", inputFile.Name(), "-o", outputFile.Name(), query)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, string(output))
	}

	content, _ := os.ReadFile(outputFile.Name())
	result, _ := gsl.Parse(bytes.NewReader(content))

	// Should only have API->DB edge (active)
	if len(result.GetEdges()) != 1 {
		t.Errorf("expected 1 edge after pipeline, got %d", len(result.GetEdges()))
	}

	// Should have API and DB
	if len(result.GetNodes()) < 2 {
		t.Errorf("expected at least 2 nodes, got %d", len(result.GetNodes()))
	}
}

func TestExampleFilesWithQueries(t *testing.T) {
	examplesDir := "../../examples"

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("failed to read examples directory: %v", err)
	}

	// Test queries for each example
	testQueries := []string{
		"subgraph exists",          // Select all nodes and edges
		"remove orphans",           // Remove isolated nodes
		"subgraph in payments",     // Test set-based queries
	}

	for _, entry := range entries {
		if entry.IsDir() || !bytes.HasSuffix([]byte(entry.Name()), []byte(".gsl")) {
			continue
		}

		for i, query := range testQueries {
			t.Run(entry.Name()+"_query_"+string(rune(48+i)), func(t *testing.T) {
				filePath := filepath.Join(examplesDir, entry.Name())
				_, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("failed to read example file: %v", err)
				}

				// Create temp output file
				outputFile, err := os.CreateTemp("", "output*.gsl")
				if err != nil {
					t.Fatalf("failed to create temp output file: %v", err)
				}
				defer os.Remove(outputFile.Name())
				outputFile.Close()

				// Run query
				binaryPath := getBinaryPath(t)
				cmd := exec.Command(binaryPath, "-i", filePath, "-o", outputFile.Name(), query)
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("query failed: %v\nOutput: %s", err, string(output))
				}

				// Verify result is valid GSL
				content, _ := os.ReadFile(outputFile.Name())
				_, pErr := gsl.Parse(bytes.NewReader(content))
				if pErr != nil && pErr.HasError() {
					t.Fatalf("result failed to parse as GSL: %v", pErr)
				}
			})
		}
	}
}

func TestCLIIntegrationErrorHandling(t *testing.T) {
	// Test invalid query
	binaryPath := getBinaryPath(t)
	cmd := exec.Command(binaryPath, "invalid query syntax !!!!")
	err := cmd.Run()

	// Should fail
	if err == nil {
		t.Errorf("expected invalid query to fail")
	}
}

func TestCLIIntegrationRoundTrip(t *testing.T) {
	// Test that parse -> query -> parse -> serialize -> parse is consistent
	gslInput := `set critical [priority=1]
set payments

node API [team="payments", status="active"] @critical @payments
node DB [team="platform", status="active"] @critical
node Cache [team="platform"] 

API -> DB [weight=10]
API -> Cache [weight=5]
DB -> Cache
`

	// Create temp input file
	inputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp input file: %v", err)
	}
	defer os.Remove(inputFile.Name())

	inputFile.WriteString(gslInput)
	inputFile.Close()

	// Create temp output file
	outputFile, err := os.CreateTemp("", "output*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp output file: %v", err)
	}
	defer os.Remove(outputFile.Name())
	outputFile.Close()

	// Run query that selects all
	binaryPath := getBinaryPath(t)
	cmd := exec.Command(binaryPath, "-i", inputFile.Name(), "-o", outputFile.Name(), "subgraph exists")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, string(output))
	}

	// Parse both original and result
	originalInput, _ := os.ReadFile(inputFile.Name())
	originalGraph, _ := gsl.Parse(bytes.NewReader(originalInput))

	resultOutput, _ := os.ReadFile(outputFile.Name())
	resultGraph, _ := gsl.Parse(bytes.NewReader(resultOutput))

	// Should have same number of nodes and edges
	if len(originalGraph.GetNodes()) != len(resultGraph.GetNodes()) {
		t.Errorf("node count mismatch: original=%d, result=%d", len(originalGraph.GetNodes()), len(resultGraph.GetNodes()))
	}

	if len(originalGraph.GetEdges()) != len(resultGraph.GetEdges()) {
		t.Errorf("edge count mismatch: original=%d, result=%d", len(originalGraph.GetEdges()), len(resultGraph.GetEdges()))
	}
}
