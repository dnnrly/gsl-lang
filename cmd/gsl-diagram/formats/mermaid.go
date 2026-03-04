package formats

import (
	"fmt"
	"strings"

	gsl "github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/cmd/gsl-diagram/converter"
)

type mermaidComponentConverter struct{}

func (c *mermaidComponentConverter) Convert(graph *gsl.Graph) string {
	var sb strings.Builder

	sb.WriteString("graph TB\n")

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
			parentLabel := converter.GetNodeLabel(parentNode)
			sb.WriteString(fmt.Sprintf("  subgraph %s[\"%s\"]\n", parentID, parentLabel))
			for _, child := range children {
				childLabel := converter.GetNodeLabel(child)
				sb.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", child.ID, childLabel))
			}
			sb.WriteString("  end\n")
		}
	}

	for _, node := range orphanNodes {
		label := converter.GetNodeLabel(node)
		sb.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", node.ID, label))
	}

	sb.WriteString("\n")

	for _, edge := range graph.Edges {
		label := converter.GetEdgeLabel(edge)
		if label != "" {
			sb.WriteString(fmt.Sprintf("  %s -->|%s| %s\n", edge.From, label, edge.To))
		} else {
			sb.WriteString(fmt.Sprintf("  %s --> %s\n", edge.From, edge.To))
		}
	}

	return sb.String()
}

type mermaidGraphConverter struct{}

func (g *mermaidGraphConverter) Convert(graph *gsl.Graph) string {
	var sb strings.Builder

	sb.WriteString("graph TD\n")

	for _, node := range graph.Nodes {
		label := converter.GetNodeLabel(node)
		sb.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", node.ID, label))
	}

	sb.WriteString("\n")

	for _, edge := range graph.Edges {
		label := converter.GetEdgeLabel(edge)
		if label != "" {
			sb.WriteString(fmt.Sprintf("  %s -->|%s| %s\n", edge.From, label, edge.To))
		} else {
			sb.WriteString(fmt.Sprintf("  %s --> %s\n", edge.From, edge.To))
		}
	}

	return sb.String()
}
