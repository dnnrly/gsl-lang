package formats

import (
	"fmt"
	"strings"

	gsl "github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/cmd/gsl-diagram/converter"
)

type plantUMLComponentConverter struct{}

func (c *plantUMLComponentConverter) Convert(graph *gsl.Graph) string {
	var sb strings.Builder

	sb.WriteString("@startuml\n")
	sb.WriteString("!define COMPONENT []\n")
	sb.WriteString("scale max 1024 width\n")
	sb.WriteString("\n")

	parentGroups := make(map[string][]*gsl.Node)
	orphanNodes := []*gsl.Node{}

	for _, node := range graph.Nodes {
		parent, hasParent := node.Attributes["parent"]
		if hasParent {
			parentID := fmt.Sprintf("%v", parent)
			parentGroups[parentID] = append(parentGroups[parentID], node)
		} else {
			orphanNodes = append(orphanNodes, node)
		}
	}

	for parentID, children := range parentGroups {
		parentNode := graph.Nodes[parentID]
		if parentNode != nil {
			sb.WriteString(fmt.Sprintf("package \"%s\" as %s {\n", converter.GetNodeLabel(parentNode), parentID))
			for _, child := range children {
				sb.WriteString(fmt.Sprintf("  [%s]\n", converter.GetNodeLabel(child)))
			}
			sb.WriteString("}\n")
		}
	}

	for _, node := range orphanNodes {
		sb.WriteString(fmt.Sprintf("[%s]\n", converter.GetNodeLabel(node)))
	}

	sb.WriteString("\n")

	for _, edge := range graph.Edges {
		label := converter.GetEdgeLabel(edge)
		if label != "" {
			sb.WriteString(fmt.Sprintf("%s --> %s : %s\n", edge.From, edge.To, label))
		} else {
			sb.WriteString(fmt.Sprintf("%s --> %s\n", edge.From, edge.To))
		}
	}

	sb.WriteString("\n")
	sb.WriteString("@enduml\n")

	return sb.String()
}
