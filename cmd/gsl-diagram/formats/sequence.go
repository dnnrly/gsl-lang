package formats

import (
	"fmt"
	"sort"
	"strings"

	gsl "github.com/dnnrly/gsl-lang"
)

type plantUMLSequenceConverter struct{}

func (c *plantUMLSequenceConverter) Convert(graph *gsl.Graph) string {
	var sb strings.Builder

	sb.WriteString("@startuml\n")
	sb.WriteString("\n")

	orderedNodes := sequenceNodeOrder(graph)

	for _, node := range orderedNodes {
		keyword := "participant"
		if shape, ok := node.Attributes["shape"]; ok {
			keyword = fmt.Sprintf("%v", shape)
		}

		if text, ok := node.Attributes["text"]; ok {
			sb.WriteString(fmt.Sprintf("%s %s as \"%s\"\n", keyword, node.ID, escapeNewlines(fmt.Sprintf("%v", text))))
		} else {
			sb.WriteString(fmt.Sprintf("%s %s\n", keyword, node.ID))
		}
	}

	sb.WriteString("\n")

	for _, edge := range graph.GetEdges() {
		if text, ok := edge.Attributes["text"]; ok {
			sb.WriteString(fmt.Sprintf("%s -> %s: %v\n", edge.From, edge.To, text))
		} else {
			sb.WriteString(fmt.Sprintf("%s -> %s\n", edge.From, edge.To))
		}
	}

	sb.WriteString("\n")
	sb.WriteString("@enduml\n")

	return sb.String()
}

func sequenceNodeOrder(graph *gsl.Graph) []*gsl.Node {
	nodes := graph.GetNodes()
	edges := graph.GetEdges()

	seen := make(map[string]bool)
	var order []*gsl.Node

	for _, edge := range edges {
		for _, id := range []string{edge.From, edge.To} {
			if !seen[id] && nodes[id] != nil {
				seen[id] = true
				order = append(order, nodes[id])
			}
		}
	}

	var remaining []string
	for id := range nodes {
		if !seen[id] {
			remaining = append(remaining, id)
		}
	}
	sort.Strings(remaining)
	for _, id := range remaining {
		order = append(order, nodes[id])
	}

	return order
}

func escapeNewlines(s string) string {
	return strings.ReplaceAll(s, "\n", "\\n")
}
