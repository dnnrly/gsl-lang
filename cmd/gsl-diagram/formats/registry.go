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

// Mermaid factory
func newMermaidFactory() converter.Factory {
	return func(diagramType string) converter.Converter {
		switch diagramType {
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
		// PlantUML only supports component for now
		return &plantUMLComponentConverter{}
	}
}
