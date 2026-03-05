// +build integration

package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	gsl "github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/cmd/gsl-diagram/formats"
)

// checkMermaidCLI checks if mermaid-cli is available.
// If INTEGRATION_STRICT env var is set, fails the test. Otherwise skips.
func checkMermaidCLI(t *testing.T) {
	cmd := exec.Command("mmdc", "--version")
	if err := cmd.Run(); err != nil {
		if os.Getenv("INTEGRATION_STRICT") != "" {
			t.Fatalf("mermaid-cli (mmdc) not available: %v (set by INTEGRATION_STRICT)", err)
		}
		t.Skipf("mermaid-cli (mmdc) not available: %v", err)
	}
}

// checkPlantUMLCLI checks if plantuml is available.
// If INTEGRATION_STRICT env var is set, fails the test. Otherwise skips.
func checkPlantUMLCLI(t *testing.T) {
	cmd := exec.Command("plantuml", "-version")
	if err := cmd.Run(); err != nil {
		if os.Getenv("INTEGRATION_STRICT") != "" {
			t.Fatalf("plantuml not available: %v (set by INTEGRATION_STRICT)", err)
		}
		t.Skipf("plantuml not available: %v", err)
	}
}

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
	if !contains(output, "component API") {
		t.Errorf("missing API component")
	}
	if !contains(output, "component DB") {
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
	if !contains(output, "component Auth") {
		t.Errorf("missing Auth child component")
	}
	if !contains(output, "component DB") {
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
	checkMermaidCLI(t)

	// Create temp files
	inputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp input file: %v", err)
	}
	defer os.Remove(inputFile.Name())

	mmdFile, err := os.CreateTemp("", "test*.mmd")
	if err != nil {
		t.Fatalf("failed to create temp mmd file: %v", err)
	}
	defer os.Remove(mmdFile.Name())

	pngFile := mmdFile.Name() + ".png"
	defer os.Remove(pngFile)

	// Write test GSL
	if _, err := inputFile.WriteString("node API: \"REST API\"\nnode DB: \"Database\"\nAPI -> DB\n"); err != nil {
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
		OutputFile:  mmdFile.Name(),
		DiagramType: "component",
		Converter:   factory,
	}

	if err := Execute(cfg); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify output file exists and has content
	content, err := os.ReadFile(mmdFile.Name())
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Fatalf("output file is empty")
	}

	// Validate with mermaid-cli
	cmd := exec.Command("mmdc", "-i", mmdFile.Name(), "-o", pngFile, "-q")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("mermaid-cli validation failed: %v\nOutput: %s", err, string(output))
	}

	// Verify PNG was created
	if _, err := os.Stat(pngFile); err != nil {
		t.Fatalf("mermaid-cli did not produce output file: %v", err)
	}
}

func TestCLIIntegrationPlantUML(t *testing.T) {
	checkPlantUMLCLI(t)

	// Create temp files
	inputFile, err := os.CreateTemp("", "test*.gsl")
	if err != nil {
		t.Fatalf("failed to create temp input file: %v", err)
	}
	defer os.Remove(inputFile.Name())

	pumlFile, err := os.CreateTemp("", "test*.puml")
	if err != nil {
		t.Fatalf("failed to create temp puml file: %v", err)
	}
	defer os.Remove(pumlFile.Name())

	pngFile := pumlFile.Name() + ".png"
	defer os.Remove(pngFile)

	// Write test GSL
	if _, err := inputFile.WriteString("node API: \"REST API\"\nnode DB: \"Database\"\nAPI -> DB\n"); err != nil {
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
		OutputFile:  pumlFile.Name(),
		DiagramType: "component",
		Converter:   factory,
	}

	if err := Execute(cfg); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify output file exists and has content
	content, err := os.ReadFile(pumlFile.Name())
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Fatalf("output file is empty")
	}

	// Validate with plantuml (output directory must exist)
	tmpDir := os.TempDir()
	cmd := exec.Command("plantuml", "-tpng", "-o", tmpDir, pumlFile.Name())
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("plantuml validation failed: %v\nOutput: %s", err, string(output))
	}

	// Verify PNG was created in temp directory
	baseName := pumlFile.Name()[:len(pumlFile.Name())-5] // remove .puml
	pngPath := baseName + ".png"
	if _, err := os.Stat(pngPath); err != nil {
		t.Fatalf("plantuml did not produce output file at %s: %v", pngPath, err)
	}
	defer os.Remove(pngPath)
}

func TestExampleFilesWithMermaid(t *testing.T) {
	examplesDir := "../../examples"

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("failed to read examples directory: %v", err)
	}

	factory, err := formats.GetFactory("mermaid")
	if err != nil {
		t.Fatalf("failed to get mermaid factory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !bytes.HasSuffix([]byte(entry.Name()), []byte(".gsl")) {
			continue
		}

		t.Run("mermaid_"+entry.Name(), func(t *testing.T) {
			filePath := examplesDir + "/" + entry.Name()
			input, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read example file %s: %v", filePath, err)
			}

			graph, warnings, err := gsl.Parse(bytes.NewReader(input))
			if err != nil {
				t.Fatalf("failed to parse GSL from %s: %v", filePath, err)
			}

			// Log warnings if any
			for _, w := range warnings {
				t.Logf("Warning from %s: %v", filePath, w)
			}

			// Convert to both diagram types
			for _, diagramType := range []string{"component", "graph"} {
				conv := factory(diagramType)
				output := conv.Convert(graph)

				if len(output) == 0 {
					t.Errorf("%s: empty output for diagram type %s", entry.Name(), diagramType)
				}

				if !contains(output, "graph") {
					t.Errorf("%s: missing graph directive for type %s", entry.Name(), diagramType)
				}
			}
		})
	}
}

func TestExampleFilesWithPlantUML(t *testing.T) {
	examplesDir := "../../examples"

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("failed to read examples directory: %v", err)
	}

	factory, err := formats.GetFactory("plantuml")
	if err != nil {
		t.Fatalf("failed to get plantuml factory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !bytes.HasSuffix([]byte(entry.Name()), []byte(".gsl")) {
			continue
		}

		t.Run("plantuml_"+entry.Name(), func(t *testing.T) {
			filePath := examplesDir + "/" + entry.Name()
			input, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read example file %s: %v", filePath, err)
			}

			graph, warnings, err := gsl.Parse(bytes.NewReader(input))
			if err != nil {
				t.Fatalf("failed to parse GSL from %s: %v", filePath, err)
			}

			// Log warnings if any
			for _, w := range warnings {
				t.Logf("Warning from %s: %v", filePath, w)
			}

			conv := factory("component")
			output := conv.Convert(graph)

			if len(output) == 0 {
				t.Errorf("%s: empty output", entry.Name())
			}

			if !contains(output, "@startuml") {
				t.Errorf("%s: missing @startuml directive", entry.Name())
			}

			if !contains(output, "@enduml") {
				t.Errorf("%s: missing @enduml directive", entry.Name())
			}
		})
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
