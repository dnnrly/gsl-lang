package formats

import (
	"bytes"
	"strings"
	"testing"

	gsl "github.com/dnnrly/gsl-lang"
)

func TestGetFactory(t *testing.T) {
	tests := []struct {
		format    string
		wantError bool
	}{
		{"mermaid", false},
		{"plantuml", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			_, err := GetFactory(tt.format)
			if (err != nil) != tt.wantError {
				t.Errorf("GetFactory(%s) error = %v, wantError %v", tt.format, err, tt.wantError)
			}
		})
	}
}

func TestMermaidComponentConverter(t *testing.T) {
	gslInput := `
node API
node DB
API -> DB
`

	graph, parseErr := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("failed to parse GSL: %v", parseErr)
	}

	factory, _ := GetFactory("mermaid")
	conv := factory("component")
	output := conv.Convert(graph)

	if !strings.Contains(output, "graph TB") {
		t.Errorf("mermaid component: missing graph directive")
	}
	if !strings.Contains(output, "API") {
		t.Errorf("mermaid component: missing API node")
	}
	if !strings.Contains(output, "-->") {
		t.Errorf("mermaid component: missing edge")
	}
}

func TestMermaidGraphConverter(t *testing.T) {
	gslInput := `
node Start
node End
Start -> End
`

	graph, parseErr := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("failed to parse GSL: %v", parseErr)
	}

	factory, _ := GetFactory("mermaid")
	conv := factory("graph")
	output := conv.Convert(graph)

	if !strings.Contains(output, "graph TD") {
		t.Errorf("mermaid graph: missing graph TD directive")
	}
	if !strings.Contains(output, "Start") {
		t.Errorf("mermaid graph: missing Start node")
	}
}

func TestPlantUMLComponentConverter(t *testing.T) {
	gslInput := `
node API
node DB
API -> DB
`

	graph, parseErr := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("failed to parse GSL: %v", parseErr)
	}

	factory, _ := GetFactory("plantuml")
	conv := factory("component")
	output := conv.Convert(graph)

	if !strings.Contains(output, "@startuml") {
		t.Errorf("plantuml: missing @startuml")
	}
	if !strings.Contains(output, "@enduml") {
		t.Errorf("plantuml: missing @enduml")
	}
	if !strings.Contains(output, "component API") {
		t.Errorf("plantuml: missing API component")
	}
	if !strings.Contains(output, "-->") {
		t.Errorf("plantuml: missing edge")
	}
}

func TestParentNodes(t *testing.T) {
	gslInput := `
node Parent
node Child {
  node GrandChild
}
Parent -> GrandChild
`

	graph, parseErr := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("failed to parse GSL: %v", parseErr)
	}

	// Test Mermaid
	mermaidFactory, _ := GetFactory("mermaid")
	mermaidOutput := mermaidFactory("component").Convert(graph)
	if !strings.Contains(mermaidOutput, "subgraph") {
		t.Errorf("mermaid: missing subgraph for parent")
	}
	if !strings.Contains(mermaidOutput, "GrandChild") {
		t.Errorf("mermaid: missing child node in subgraph")
	}

	// Test PlantUML
	pumlFactory, _ := GetFactory("plantuml")
	pumlOutput := pumlFactory("component").Convert(graph)
	if !strings.Contains(pumlOutput, "package") {
		t.Errorf("plantuml: missing package for parent")
	}
	if !strings.Contains(pumlOutput, "component GrandChild") {
		t.Errorf("plantuml: missing child component in package")
	}
}

func TestMermaidSequenceConverter(t *testing.T) {
	gslInput := `
node Client
node Server
node Database

Client -> Server [method = "GET /api/users"] {
    Server -> Database [method = "query"]
    Database -> Server [method = "return results"]
    Server -> Client [method = "response 200"]
}
`

	graph, parseErr := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("failed to parse GSL: %v", parseErr)
	}

	factory, _ := GetFactory("mermaid")
	conv := factory("sequence")
	output := conv.Convert(graph)

	if !strings.Contains(output, "sequenceDiagram") {
		t.Errorf("mermaid sequence: missing sequenceDiagram header")
	}
	if !strings.Contains(output, "participant Client") {
		t.Errorf("mermaid sequence: missing Client participant")
	}
	if !strings.Contains(output, "participant Server") {
		t.Errorf("mermaid sequence: missing Server participant")
	}
	if !strings.Contains(output, "participant Database") {
		t.Errorf("mermaid sequence: missing Database participant")
	}
	if !strings.Contains(output, "Client->>Server: GET /api/users") {
		t.Errorf("mermaid sequence: missing Client->Server arrow with label")
	}
	if !strings.Contains(output, "Server->>Database: query") {
		t.Errorf("mermaid sequence: missing Server->Database arrow")
	}
	if !strings.Contains(output, "activate Server") {
		t.Errorf("mermaid sequence: missing activate Server")
	}
	if !strings.Contains(output, "deactivate Server") {
		t.Errorf("mermaid sequence: missing deactivate Server")
	}
}

func TestPlantUMLSequenceConverter(t *testing.T) {
	gslInput := `
node Client
node Server

Client -> Server [method = "GET /api/users"] {
    Server -> Client [method = "response"]
}
`

	graph, parseErr := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("failed to parse GSL: %v", parseErr)
	}

	factory, _ := GetFactory("plantuml")
	conv := factory("sequence")
	output := conv.Convert(graph)

	if !strings.Contains(output, "@startuml") {
		t.Errorf("plantuml sequence: missing @startuml")
	}
	if !strings.Contains(output, "@enduml") {
		t.Errorf("plantuml sequence: missing @enduml")
	}
	if !strings.Contains(output, "participant Client") {
		t.Errorf("plantuml sequence: missing Client participant")
	}
	if !strings.Contains(output, "participant Server") {
		t.Errorf("plantuml sequence: missing Server participant")
	}
	if !strings.Contains(output, "Client -> Server : GET /api/users") {
		t.Errorf("plantuml sequence: missing Client->Server arrow")
	}
	if !strings.Contains(output, "activate Server") {
		t.Errorf("plantuml sequence: missing activate")
	}
	if !strings.Contains(output, "deactivate Server") {
		t.Errorf("plantuml sequence: missing deactivate")
	}
}

func TestPlantUMLSequenceUnlabeledEdge(t *testing.T) {
	gslInput := `
node A
node B
node C

E1: A -> B {
    B -> C
}
`

	graph, parseErr := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("failed to parse GSL: %v", parseErr)
	}

	factory, _ := GetFactory("plantuml")
	conv := factory("sequence")
	output := conv.Convert(graph)

	if !strings.Contains(output, "A -> B") {
		t.Errorf("plantuml sequence: missing A->B arrow")
	}
	if !strings.Contains(output, "B -> C") {
		t.Errorf("plantuml sequence: missing B->C arrow")
	}
	if !strings.Contains(output, "activate B") {
		t.Errorf("plantuml sequence: missing activate B")
	}
	if !strings.Contains(output, "deactivate B") {
		t.Errorf("plantuml sequence: missing deactivate B")
	}
}

func TestSequenceParticipants(t *testing.T) {
	gslInput := `
node Client [text = "Web Browser"]
node Server [text = "API Server"]

Client -> Server
Server -> Client
`

	graph, parseErr := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("failed to parse GSL: %v", parseErr)
	}

	factory, _ := GetFactory("mermaid")
	conv := factory("sequence")
	output := conv.Convert(graph)

	if !strings.Contains(output, `participant Client as "Web Browser"`) {
		t.Errorf("mermaid sequence: should use text attribute for participant display")
	}
	if !strings.Contains(output, `participant Server as "API Server"`) {
		t.Errorf("mermaid sequence: should use text attribute for Server display")
	}
}

func TestSequenceArrowStyles(t *testing.T) {
	tests := []struct {
		name      string
		gsl       string
		wantArrow string // substring to check in mermaid output
	}{
		{
			name: "return arrow auto-detected",
			gsl: `
node A
node B
node C

E1: A -> B {
    B -> C
    C -> A
}
`,
			wantArrow: "C-->>A",
		},
		{
			name: "request arrow solid",
			gsl: `
node A
node B
node C

E1: A -> B {
    B -> C
}
`,
			wantArrow: "B->>C",
		},
		{
			name: "arrow attribute override to dashed",
			gsl: `
node A
node B

A -> B [arrow = "dashed"]
`,
			wantArrow: "-->>",
		},
		{
			name: "arrow attribute override to solid",
			gsl: `
node A
node B

E1: A -> B [arrow = "solid"]
B -> A [arrow = "solid"]
`,
			wantArrow: "->>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, parseErr := gsl.Parse(bytes.NewReader([]byte(tt.gsl)))
			if parseErr != nil && parseErr.HasError() {
				t.Fatalf("failed to parse GSL: %v", parseErr)
			}

			factory, _ := GetFactory("mermaid")
			conv := factory("sequence")
			output := conv.Convert(graph)

			if !strings.Contains(output, tt.wantArrow) {
				t.Errorf("mermaid sequence: expected arrow %q in output", tt.wantArrow)
			}
		})
	}
}

func TestSequenceActivation(t *testing.T) {
	gslInput := `
node A
node B
node C

A -> B {
    B -> C
}
`

	graph, parseErr := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("failed to parse GSL: %v", parseErr)
	}

	factory, _ := GetFactory("mermaid")
	conv := factory("sequence")
	output := conv.Convert(graph)

	if !strings.Contains(output, "activate B") {
		t.Errorf("mermaid sequence: B should be activated (has children)")
	}
	if !strings.Contains(output, "deactivate B") {
		t.Errorf("mermaid sequence: B should be deactivated")
	}
}

func TestSequenceNestedScopes(t *testing.T) {
	gslInput := `
node A
node B
node C
node D

A -> B {
    B -> C {
        C -> D
    }
}
`

	graph, parseErr := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("failed to parse GSL: %v", parseErr)
	}

	factory, _ := GetFactory("mermaid")
	conv := factory("sequence")
	output := conv.Convert(graph)

	if !strings.Contains(output, "activate B") {
		t.Errorf("mermaid sequence: B should be activated")
	}
	if !strings.Contains(output, "activate C") {
		t.Errorf("mermaid sequence: C should be activated")
	}
	if !strings.Contains(output, "deactivate C") {
		t.Errorf("mermaid sequence: C should be deactivated")
	}
	if !strings.Contains(output, "deactivate B") {
		t.Errorf("mermaid sequence: B should be deactivated")
	}

	// Verify nesting order: A->B, activate B, B->C, activate C, C->D, deactivate C, deactivate B
	lines := strings.Split(output, "\n")
	aToBIdx := -1
	activateBIdx := -1
	bToCIdx := -1
	activateCIdx := -1
	deactivateCIdx := -1
	deactivateBIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case "A->>B":
			aToBIdx = i
		case "activate B":
			activateBIdx = i
		case "B->>C":
			bToCIdx = i
		case "activate C":
			activateCIdx = i
		case "deactivate C":
			deactivateCIdx = i
		case "deactivate B":
			deactivateBIdx = i
		}
	}

	if !(aToBIdx < activateBIdx && activateBIdx < bToCIdx &&
		bToCIdx < activateCIdx && activateCIdx < deactivateCIdx &&
		deactivateCIdx < deactivateBIdx) {
		t.Errorf("mermaid sequence: activation nesting order is incorrect")
	}
}

func TestSequenceFlatEdges(t *testing.T) {
	gslInput := `
node A
node B
node C

A -> B
B -> C
`

	graph, parseErr := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("failed to parse GSL: %v", parseErr)
	}

	factory, _ := GetFactory("mermaid")
	conv := factory("sequence")
	output := conv.Convert(graph)

	if !strings.Contains(output, "A->>B") {
		t.Errorf("mermaid sequence: missing edge A->B")
	}
	if !strings.Contains(output, "B->>C") {
		t.Errorf("mermaid sequence: missing edge B->C")
	}
	if strings.Contains(output, "activate") {
		t.Errorf("mermaid sequence: flat edges should not have activation")
	}
	if strings.Contains(output, "deactivate") {
		t.Errorf("mermaid sequence: flat edges should not have deactivation")
	}
}

func TestSequenceParticipantOrder(t *testing.T) {
	gslInput := `
node Z
node A
node M

Z -> A
A -> M
`

	graph, parseErr := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("failed to parse GSL: %v", parseErr)
	}

	factory, _ := GetFactory("mermaid")
	conv := factory("sequence")
	output := conv.Convert(graph)

	// Participants should be in first-appearance order: Z, A, M (not alphabetical)
	lines := strings.Split(output, "\n")
	participants := make([]string, 0)
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "participant ") {
			participants = append(participants, strings.TrimSpace(line))
		}
	}

	if len(participants) < 3 {
		t.Fatalf("expected 3 participants, got %d", len(participants))
	}

	expected := []string{"Z", "A", "M"}
	for i, name := range expected {
		if !strings.Contains(participants[i], name) {
			t.Errorf("participant %d should be %s, got %s", i, name, participants[i])
		}
	}
}

func TestEdgeLabels(t *testing.T) {
	gslInput := `
node A
node B
A -> B [label="connects"]
`

	graph, parseErr := gsl.Parse(bytes.NewReader([]byte(gslInput)))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("failed to parse GSL: %v", parseErr)
	}

	mermaidFactory, _ := GetFactory("mermaid")
	output := mermaidFactory("component").Convert(graph)
	if !strings.Contains(output, "connects") {
		t.Errorf("mermaid: edge label not included")
	}
}
