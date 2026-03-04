package converter

import (
	"fmt"

	gsl "github.com/dnnrly/gsl-lang"
)

// GetNodeLabel returns the display label for a node
// Prefers "text" attribute, falls back to node ID
func GetNodeLabel(node *gsl.Node) string {
	if text, ok := node.Attributes["text"]; ok {
		return fmt.Sprintf("%v", text)
	}
	return node.ID
}

// GetEdgeLabel returns the display label for an edge
// Checks for "label", "name", or "method" attributes
func GetEdgeLabel(edge *gsl.Edge) string {
	for _, attrName := range []string{"label", "name", "method"} {
		if val, ok := edge.Attributes[attrName]; ok {
			return fmt.Sprintf("%v", val)
		}
	}
	return ""
}
