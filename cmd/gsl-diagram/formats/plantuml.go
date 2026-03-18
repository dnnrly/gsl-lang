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
	sb.WriteString("\n")

	parentGroups := make(map[string][]*gsl.Node)
	usedParents := make(map[string]bool)
	orphanNodes := []*gsl.Node{}

	nodes := graph.GetNodes()
	for _, node := range nodes {
		parent, hasParent := node.Attributes["parent"]
		if hasParent {
			parentID := fmt.Sprintf("%v", parent)
			parentGroups[parentID] = append(parentGroups[parentID], node)
			usedParents[parentID] = true
		} else {
			orphanNodes = append(orphanNodes, node)
		}
	}

	for parentID, children := range parentGroups {
		parentNode := nodes[parentID]
		if parentNode != nil {
			sb.WriteString(fmt.Sprintf("package \"%s\" as %s {\n", converter.GetNodeLabel(parentNode), parentID))
			for _, child := range children {
				sb.WriteString(fmt.Sprintf("  component %s\n", child.ID))
			}
			sb.WriteString("}\n")
		}
	}

	for _, node := range orphanNodes {
		// Skip nodes that are used as parents
		if !usedParents[node.ID] {
			sb.WriteString(fmt.Sprintf("component %s\n", node.ID))
		}
	}

	sb.WriteString("\n")

	edges := graph.GetEdges()
	for _, edge := range edges {
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
