// +build integration

package main

import (
	"bytes"
	"os"
	"testing"

	gsl "github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/cmd/gsl-diagram/formats"
)

func TestMermaidComponentDiagramOutput(t *testing.T) {
	gslInput := `
node API: "REST API"
node DB: "Database"
node Cache: "Redis"

API -> DB [label="query"]
API -> Cache [label="cache"]
`

	graph, _, err := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if err != nil {
		t.Fatalf("failed to parse GSL: %v", err)
	}

	factory, err := formats.GetFactory("mermaid")
	if err != nil {
		t.Fatalf("failed to get factory: %v", err)
	}

	conv := factory("component")
	output := conv.Convert(graph)

	// Validate basic structure
	if !contains(output, "graph TB") {
		t.Errorf("missing graph directive")
	}
	if !contains(output, "API[") {
		t.Errorf("missing API node")
	}
	if !contains(output, "DB[") {
		t.Errorf("missing DB node")
	}
	if !contains(output, "Cache[") {
		t.Errorf("missing Cache node")
	}
	if !contains(output, "-->") {
		t.Errorf("missing edge connector")
	}
}

func TestMermaidGraphDiagramOutput(t *testing.T) {
	gslInput := `
node Start: "Begin"
node Process: "Processing"
node End: "Complete"

Start -> Process [label="execute"]
Process -> End
`

	graph, _, err := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if err != nil {
		t.Fatalf("failed to parse GSL: %v", err)
	}

	factory, err := formats.GetFactory("mermaid")
	if err != nil {
		t.Fatalf("failed to get factory: %v", err)
	}

	conv := factory("graph")
	output := conv.Convert(graph)

	// Validate basic structure
	if !contains(output, "graph TD") {
		t.Errorf("missing graph TD directive")
	}
	if !contains(output, "Start[") {
		t.Errorf("missing Start node")
	}
	if !contains(output, "Process[") {
		t.Errorf("missing Process node")
	}
	if !contains(output, "execute") {
		t.Errorf("missing edge label")
	}
}

func TestPlantUMLComponentDiagramOutput(t *testing.T) {
	gslInput := `
node API: "REST API"
node DB: "Database"
node Cache: "Redis"

API -> DB [label="query"]
API -> Cache [label="cache"]
`

	graph, _, err := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if err != nil {
		t.Fatalf("failed to parse GSL: %v", err)
	}

	factory, err := formats.GetFactory("plantuml")
	if err != nil {
		t.Fatalf("failed to get factory: %v", err)
	}

	conv := factory("component")
	output := conv.Convert(graph)

	// Validate PlantUML structure
	if !contains(output, "@startuml") {
		t.Errorf("missing @startuml directive")
	}
	if !contains(output, "@enduml") {
		t.Errorf("missing @enduml directive")
	}
	if !contains(output, "[REST API]") {
		t.Errorf("missing API component")
	}
	if !contains(output, "[Database]") {
		t.Errorf("missing DB component")
	}
	if !contains(output, "-->") {
		t.Errorf("missing edge connector")
	}
	if !contains(output, "query") {
		t.Errorf("missing edge label")
	}
}

func TestPlantUMLParentRelationships(t *testing.T) {
	gslInput := `
node System
node Backend {
  node Auth
  node DB
}

System -> Auth
System -> DB
`

	graph, _, err := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if err != nil {
		t.Fatalf("failed to parse GSL: %v", err)
	}

	factory, err := formats.GetFactory("plantuml")
	if err != nil {
		t.Fatalf("failed to get factory: %v", err)
	}

	conv := factory("component")
	output := conv.Convert(graph)

	// Validate package structure for parents
	if !contains(output, "package") {
		t.Errorf("missing package directive for parent")
	}
	if !contains(output, "[Auth]") {
		t.Errorf("missing Auth child component")
	}
	if !contains(output, "[DB]") {
		t.Errorf("missing DB child component")
	}
}

func TestMermaidParentRelationships(t *testing.T) {
	gslInput := `
node System
node Backend {
  node Auth
  node DB
}

System -> Auth
System -> DB
`

	graph, _, err := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if err != nil {
		t.Fatalf("failed to parse GSL: %v", err)
	}

	factory, err := formats.GetFactory("mermaid")
	if err != nil {
		t.Fatalf("failed to get factory: %v", err)
	}

	conv := factory("component")
	output := conv.Convert(graph)

	// Validate subgraph structure for parents
	if !contains(output, "subgraph") {
		t.Errorf("missing subgraph directive for parent")
	}
	if !contains(output, "Auth[") {
		t.Errorf("missing Auth node in subgraph")
	}
	if !contains(output, "DB[") {
		t.Errorf("missing DB node in subgraph")
	}
}

func TestCLIIntegrationMermaid(t *testing.T) {
	// Create temp files
	inputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp input file: %v", err)
	}
	defer os.Remove(inputFile.Name())

	outputFile, err := os.CreateTemp("", "test*.mmd")
	if err != nil {
		t.Fatalf("failed to create temp output file: %v", err)
	}
	defer os.Remove(outputFile.Name())

	// Write test GSL
	if _, err := inputFile.WriteString("node A\nnode B\nA -> B\n"); err != nil {
		t.Fatalf("failed to write test input: %v", err)
	}
	inputFile.Close()

	// Execute conversion
	factory, err := formats.GetFactory("mermaid")
	if err != nil {
		t.Fatalf("failed to get mermaid factory: %v", err)
	}

	cfg := &Config{
		InputFile:   inputFile.Name(),
		OutputFile:  outputFile.Name(),
		DiagramType: "component",
		Converter:   factory,
	}

	if err := Execute(cfg); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify output file exists and has content
	content, err := os.ReadFile(outputFile.Name())
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Errorf("output file is empty")
	}

	if !contains(string(content), "graph TB") {
		t.Errorf("output missing graph directive")
	}
}

func TestCLIIntegrationPlantUML(t *testing.T) {
	// Create temp files
	inputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp input file: %v", err)
	}
	defer os.Remove(inputFile.Name())

	outputFile, err := os.CreateTemp("", "test*.puml")
	if err != nil {
		t.Fatalf("failed to create temp output file: %v", err)
	}
	defer os.Remove(outputFile.Name())

	// Write test GSL
	if _, err := inputFile.WriteString("node A\nnode B\nA -> B\n"); err != nil {
		t.Fatalf("failed to write test input: %v", err)
	}
	inputFile.Close()

	// Execute conversion
	factory, err := formats.GetFactory("plantuml")
	if err != nil {
		t.Fatalf("failed to get plantuml factory: %v", err)
	}

	cfg := &Config{
		InputFile:   inputFile.Name(),
		OutputFile:  outputFile.Name(),
		DiagramType: "component",
		Converter:   factory,
	}

	if err := Execute(cfg); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify output file exists and has content
	content, err := os.ReadFile(outputFile.Name())
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Errorf("output file is empty")
	}

	if !contains(string(content), "@startuml") {
		t.Errorf("output missing @startuml directive")
	}
	if !contains(string(content), "@enduml") {
		t.Errorf("output missing @enduml directive")
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
