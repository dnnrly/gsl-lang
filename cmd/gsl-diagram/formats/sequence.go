package formats

import (
	"fmt"
	"strings"

	gsl "github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/cmd/gsl-diagram/converter"
)

const (
	arrowSolid  = "solid"
	arrowDashed = "dashed"
)

func collectParticipants(edges []*gsl.Edge, nodes map[string]*gsl.Node) []string {
	seen := make(map[string]bool)
	order := make([]string, 0, len(nodes))
	for _, edge := range edges {
		for _, name := range []string{edge.From, edge.To} {
			if !seen[name] {
				seen[name] = true
				order = append(order, name)
			}
		}
	}
	return order
}

func participantDisplayName(name string, nodes map[string]*gsl.Node) string {
	if node, ok := nodes[name]; ok {
		if text, ok := node.Attributes["text"]; ok {
			return fmt.Sprintf("%v", text)
		}
	}
	return name
}

func arrowStyle(edge *gsl.Edge, parentEdge *gsl.Edge) string {
	if attr, ok := edge.Attributes["arrow"]; ok {
		switch fmt.Sprintf("%v", attr) {
		case "dashed":
			return arrowDashed
		case "solid":
			return arrowSolid
		}
	}
	if parentEdge != nil && edge.To == parentEdge.From {
		return arrowDashed
	}
	return arrowSolid
}

func buildChildSet(edges []*gsl.Edge) map[*gsl.Edge]struct{} {
	childSet := make(map[*gsl.Edge]struct{})
	for _, edge := range edges {
		for _, child := range edge.Children {
			childSet[child] = struct{}{}
		}
	}
	return childSet
}

func renderSequenceEdgeMermaid(sb *strings.Builder, edge *gsl.Edge, parentEdge *gsl.Edge, nodes map[string]*gsl.Node, indent string) {
	style := arrowStyle(edge, parentEdge)
	arrow := "->>"
	if style == arrowDashed {
		arrow = "-->>"
	}

	label := converter.GetEdgeLabel(edge)
	if label != "" {
		sb.WriteString(fmt.Sprintf("%s%s%s%s: %s\n", indent, edge.From, arrow, edge.To, label))
	} else {
		sb.WriteString(fmt.Sprintf("%s%s%s%s\n", indent, edge.From, arrow, edge.To))
	}

	if len(edge.Children) > 0 {
		sb.WriteString(fmt.Sprintf("%sactivate %s\n", indent, edge.To))
		for _, child := range edge.Children {
			renderSequenceEdgeMermaid(sb, child, edge, nodes, indent+"    ")
		}
		sb.WriteString(fmt.Sprintf("%sdeactivate %s\n", indent, edge.To))
	}
}

func renderSequenceEdgePlantUML(sb *strings.Builder, edge *gsl.Edge, parentEdge *gsl.Edge, nodes map[string]*gsl.Node, indent string) {
	style := arrowStyle(edge, parentEdge)
	arrow := "->"
	if style == arrowDashed {
		arrow = "-->"
	}

	label := converter.GetEdgeLabel(edge)
	if label != "" {
		sb.WriteString(fmt.Sprintf("%s%s %s %s : %s\n", indent, edge.From, arrow, edge.To, label))
	} else {
		sb.WriteString(fmt.Sprintf("%s%s %s %s\n", indent, edge.From, arrow, edge.To))
	}

	if len(edge.Children) > 0 {
		sb.WriteString(fmt.Sprintf("%sactivate %s\n", indent, edge.To))
		for _, child := range edge.Children {
			renderSequenceEdgePlantUML(sb, child, edge, nodes, indent+"    ")
		}
		sb.WriteString(fmt.Sprintf("%sdeactivate %s\n", indent, edge.To))
	}
}
