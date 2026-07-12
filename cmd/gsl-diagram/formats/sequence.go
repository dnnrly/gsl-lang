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

	edges := graph.GetEdges()
	emitted := make(map[int]bool)

	for i := 0; i < len(edges); i++ {
		if emitted[i] {
			continue
		}

		edge := edges[i]

		switch {
		case edge.Label != "":
			emitLabeledScope(&sb, edges, i, emitted, 0)
		case edge.Parent != "":
			continue
		case isScopeRoot(edges, i):
			emitUnlabeledScope(&sb, edges, i, emitted, 0)
		case hasActivateAttr(edge):
			emitActivationAt(&sb, edge, 0)
			sb.WriteString("return\n\n")
			emitted[i] = true
		default:
			emitEdge(&sb, edge)
			emitted[i] = true
		}
	}

	sb.WriteString("\n@enduml\n")

	return sb.String()
}

// isScopeRoot checks if the edge at idx starts a scoped block.
// An edge is a scope root if its next non-explicit-child edge has a greater ScopeDepth.
func isScopeRoot(edges []*gsl.Edge, idx int) bool {
	depth := edges[idx].ScopeDepth
	for j := idx + 1; j < len(edges); j++ {
		if edges[j].Parent != "" {
			continue
		}
		return edges[j].ScopeDepth > depth
	}
	return false
}

// emitLabeledScope emits a labeled activation scope and all edges within it.
func emitLabeledScope(sb *strings.Builder, edges []*gsl.Edge, startIdx int, emitted map[int]bool, indent int) {
	edge := edges[startIdx]
	emitted[startIdx] = true

	emitActivationAt(sb, edge, indent)

	scopeEnd := findLabelScopeEnd(edges, startIdx)

	for j := startIdx + 1; j < scopeEnd; j++ {
		child := edges[j]
		if emitted[j] {
			continue
		}
		if child.Parent != "" {
			continue
		}

		if isScopeRoot(edges, j) {
			emitUnlabeledScope(sb, edges, j, emitted, indent+1)
		} else {
			emitIndentedEdgeAt(sb, child, indent+1)
			emitted[j] = true
		}
	}

	for _, child := range edge.Children {
		emitIndentedEdgeAt(sb, child, indent+1)
	}

	indentStr := indentPrefix(indent)
	sb.WriteString(fmt.Sprintf("%sreturn\n\n", indentStr))
}

// emitUnlabeledScope emits a scoped-block activation and all edges within it.
func emitUnlabeledScope(sb *strings.Builder, edges []*gsl.Edge, startIdx int, emitted map[int]bool, indent int) {
	edge := edges[startIdx]
	emitted[startIdx] = true

	emitActivationAt(sb, edge, indent)

	scopeEnd := findDepthScopeEnd(edges, startIdx)

	for j := startIdx + 1; j < scopeEnd; j++ {
		child := edges[j]
		if emitted[j] {
			continue
		}
		if child.Parent != "" {
			continue
		}

		if isScopeRoot(edges, j) {
			emitUnlabeledScope(sb, edges, j, emitted, indent+1)
		} else {
			emitIndentedEdgeAt(sb, child, indent+1)
			emitted[j] = true
		}
	}

	indentStr := indentPrefix(indent)
	if indent == 0 {
		sb.WriteString(fmt.Sprintf("%sreturn\n\n", indentStr))
	} else {
		sb.WriteString(fmt.Sprintf("%sreturn\n", indentStr))
	}
}

// findLabelScopeEnd returns the index where a labeled scope ends.
// Scope ends at the next edge with text, the next labeled edge, or end of input.
func findLabelScopeEnd(edges []*gsl.Edge, startIdx int) int {
	for j := startIdx + 1; j < len(edges); j++ {
		child := edges[j]
		if child.Parent != "" {
			continue
		}
		if child.Label != "" {
			return j
		}
		if _, ok := child.Attributes["text"]; ok {
			return j
		}
	}
	return len(edges)
}

// findDepthScopeEnd returns the index where a depth-based scope ends.
// Scope ends at the next edge with ScopeDepth <= the current edge's ScopeDepth.
func findDepthScopeEnd(edges []*gsl.Edge, startIdx int) int {
	depth := edges[startIdx].ScopeDepth
	for j := startIdx + 1; j < len(edges); j++ {
		child := edges[j]
		if child.Parent != "" {
			continue
		}
		if child.ScopeDepth <= depth {
			return j
		}
	}
	return len(edges)
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

func hasActivateAttr(edge *gsl.Edge) bool {
	_, ok := edge.Attributes["activate"]
	return ok
}

func indentPrefix(level int) string {
	return strings.Repeat("    ", level)
}

func emitActivationAt(sb *strings.Builder, edge *gsl.Edge, indent int) {
	prefix := indentPrefix(indent)
	if text, ok := edge.Attributes["text"]; ok {
		sb.WriteString(fmt.Sprintf("%s%s->%s ++: %v\n", prefix, edge.From, edge.To, text))
	} else {
		sb.WriteString(fmt.Sprintf("%s%s->%s ++\n", prefix, edge.From, edge.To))
	}
}

func emitIndentedEdgeAt(sb *strings.Builder, edge *gsl.Edge, indent int) {
	prefix := indentPrefix(indent)
	if text, ok := edge.Attributes["text"]; ok {
		sb.WriteString(fmt.Sprintf("%s%s->%s: %v\n", prefix, edge.From, edge.To, text))
	} else {
		sb.WriteString(fmt.Sprintf("%s%s->%s\n", prefix, edge.From, edge.To))
	}
}

func emitEdge(sb *strings.Builder, edge *gsl.Edge) {
	if text, ok := edge.Attributes["text"]; ok {
		sb.WriteString(fmt.Sprintf("%s -> %s: %v\n", edge.From, edge.To, text))
	} else {
		sb.WriteString(fmt.Sprintf("%s -> %s\n", edge.From, edge.To))
	}
}
