package converter

import (
	gsl "github.com/dnnrly/gsl-lang"
)

// Converter converts a GSL graph to diagram syntax
type Converter interface {
	Convert(*gsl.Graph) string
}

// Factory creates converters based on diagram type
type Factory func(diagramType string) Converter
