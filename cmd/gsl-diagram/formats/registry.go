package formats

import (
	"fmt"

	"github.com/dnnrly/gsl-lang/cmd/gsl-diagram/converter"
)

// GetFactory returns a converter factory for the specified format
func GetFactory(format string) (converter.Factory, error) {
	switch format {
	case "mermaid":
		return newMermaidFactory(), nil
	case "plantuml":
		return newPlantUMLFactory(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s (supported: mermaid, plantuml)", format)
	}
}

// ValidateDiagramType reports whether the given diagram type is supported for
// the format. Unsupported combinations (e.g. PlantUML "graph", which silently
// fell back to component output) fail loudly instead of emitting a view the
// user did not ask for.
func ValidateDiagramType(format, diagramType string) error {
	switch format {
	case "mermaid":
		switch diagramType {
		case "component", "graph", "sequence":
			return nil
		}
		return fmt.Errorf("unsupported diagram type for mermaid: %s (supported: component, graph, sequence)", diagramType)
	case "plantuml":
		switch diagramType {
		case "component", "sequence":
			return nil
		}
		return fmt.Errorf("unsupported diagram type for plantuml: %s (supported: component, sequence)", diagramType)
	}
	return fmt.Errorf("unsupported format: %s (supported: mermaid, plantuml)", format)
}

// Mermaid factory
func newMermaidFactory() converter.Factory {
	return func(diagramType string) converter.Converter {
		switch diagramType {
		case "sequence":
			return &mermaidSequenceConverter{}
		case "graph":
			return &mermaidGraphConverter{}
		case "component":
			fallthrough
		default:
			return &mermaidComponentConverter{}
		}
	}
}

// PlantUML factory
func newPlantUMLFactory() converter.Factory {
	return func(diagramType string) converter.Converter {
		switch diagramType {
		case "sequence":
			return &plantUMLSequenceConverter{}
		default:
			return &plantUMLComponentConverter{}
		}
	}
}
