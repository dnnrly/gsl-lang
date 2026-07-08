package formats

import (
	"fmt"
	"strings"

	gsl "github.com/dnnrly/gsl-lang"
	"github.com/dnnrly/gsl-lang/cmd/gsl-diagram/converter"
)

type mermaidSequenceConverter struct{}

func (c *mermaidSequenceConverter) Convert(graph *gsl.Graph) string {
	var sb strings.Builder

	sb.WriteString("sequenceDiagram\n")

	nodes := graph.GetNodes()
	edges := graph.GetEdges()

	participants := collectParticipants(edges, nodes)
	for _, p := range participants {
		displayName := participantDisplayName(p, nodes)
		if displayName != p {
			sb.WriteString(fmt.Sprintf("    participant %s as \"%s\"\n", p, displayName))
		} else {
			sb.WriteString(fmt.Sprintf("    participant %s\n", p))
		}
	}
	sb.WriteString("\n")

	childSet := buildChildSet(edges)
	for _, edge := range edges {
		if _, isChild := childSet[edge]; !isChild {
			renderSequenceEdgeMermaid(&sb, edge, nil, nodes, "    ")
		}
	}

	return sb.String()
}

type mermaidComponentConverter struct{}

func (c *mermaidComponentConverter) Convert(graph *gsl.Graph) string {
	var sb strings.Builder

	sb.WriteString("graph TB\n")

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
		// Skip nodes that are used as parents
		if !usedParents[node.ID] {
			label := converter.GetNodeLabel(node)
			sb.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", node.ID, label))
		}
	}

	sb.WriteString("\n")

	edges := graph.GetEdges()
	for _, edge := range edges {
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

	nodes := graph.GetNodes()
	for _, node := range nodes {
		label := converter.GetNodeLabel(node)
		sb.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", node.ID, label))
	}

	sb.WriteString("\n")

	edges := graph.GetEdges()
	for _, edge := range edges {
		label := converter.GetEdgeLabel(edge)
		if label != "" {
			sb.WriteString(fmt.Sprintf("  %s -->|%s| %s\n", edge.From, label, edge.To))
		} else {
			sb.WriteString(fmt.Sprintf("  %s --> %s\n", edge.From, edge.To))
		}
	}

	return sb.String()
}
