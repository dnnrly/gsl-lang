package formats

import (
	"fmt"
	"strings"

	gsl "github.com/dnnrly/gsl-lang"
)

type mermaidSequenceConverter struct{}

func (c *mermaidSequenceConverter) Convert(graph *gsl.Graph) string {
	var sb strings.Builder

	sb.WriteString("sequenceDiagram\n")

	orderedNodes := sequenceNodeOrder(graph)

	for _, node := range orderedNodes {
		keyword := "participant"
		var stereotype string
		if shape, ok := node.Attributes["shape"]; ok {
			s := fmt.Sprintf("%v", shape)
			switch s {
			case "actor":
				keyword = "actor"
			case "boundary", "control", "entity", "database", "collections", "queue":
				stereotype = s
			}
		}

		if text, ok := node.Attributes["text"]; ok {
			cleanText := strings.ReplaceAll(fmt.Sprintf("%v", text), "\n", " ")
			if stereotype != "" {
				sb.WriteString(fmt.Sprintf("    %s %s as %s <<%s>>\n", keyword, node.ID, cleanText, stereotype))
			} else {
				sb.WriteString(fmt.Sprintf("    %s %s as %s\n", keyword, node.ID, cleanText))
			}
		} else {
			if stereotype != "" {
				sb.WriteString(fmt.Sprintf("    %s %s as \"%s\" <<%s>>\n", keyword, node.ID, node.ID, stereotype))
			} else {
				sb.WriteString(fmt.Sprintf("    %s %s\n", keyword, node.ID))
			}
		}
	}

	edges := graph.GetEdges()
	emitted := make(map[int]bool)

	for i := 0; i < len(edges); i++ {
		if emitted[i] {
			continue
		}

		edge := edges[i]

		switch {
		case edge.Label != "":
			mermaidEmitLabeledScope(&sb, edges, i, emitted, 0)
		case edge.Parent != "":
			continue
		case isScopeRoot(edges, i):
			mermaidEmitUnlabeledScope(&sb, edges, i, emitted, 0)
		case hasActivateAttr(edge):
			mermaidEmitActivationAt(&sb, edge, 0)
			sb.WriteString(fmt.Sprintf("    deactivate %s\n", edge.To))
			emitted[i] = true
		default:
			mermaidEmitEdge(&sb, edge)
			emitted[i] = true
		}
	}

	result := sb.String()
	result = strings.TrimRight(result, "\n")
	result += "\n"

	return result
}

func mermaidEmitLabeledScope(sb *strings.Builder, edges []*gsl.Edge, startIdx int, emitted map[int]bool, indent int) {
	edge := edges[startIdx]
	emitted[startIdx] = true

	mermaidEmitActivationAt(sb, edge, indent)

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
			mermaidEmitUnlabeledScope(sb, edges, j, emitted, indent+1)
		} else {
			mermaidEmitIndentedEdgeAt(sb, child, indent+1)
			emitted[j] = true
		}
	}

	for _, child := range edge.Children {
		mermaidEmitIndentedEdgeAt(sb, child, indent+1)
	}

	indentStr := mermaidIndentPrefix(indent)
	sb.WriteString(fmt.Sprintf("%sdeactivate %s\n", indentStr, edge.To))
}

func mermaidEmitUnlabeledScope(sb *strings.Builder, edges []*gsl.Edge, startIdx int, emitted map[int]bool, indent int) {
	edge := edges[startIdx]
	emitted[startIdx] = true

	mermaidEmitActivationAt(sb, edge, indent)

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
			mermaidEmitUnlabeledScope(sb, edges, j, emitted, indent+1)
		} else {
			mermaidEmitIndentedEdgeAt(sb, child, indent+1)
			emitted[j] = true
		}
	}

	indentStr := mermaidIndentPrefix(indent)
	sb.WriteString(fmt.Sprintf("%sdeactivate %s\n", indentStr, edge.To))
}

// mermaidEmitActivationAt emits a Mermaid activation line with + suffix on arrow.
func mermaidEmitActivationAt(sb *strings.Builder, edge *gsl.Edge, indent int) {
	prefix := mermaidIndentPrefix(indent)
	arrow := mermaidGetArrowStyle(edge)
	activateArrow := arrow + "+"
	if text, ok := edge.Attributes["text"]; ok {
		sb.WriteString(fmt.Sprintf("%s%s %s %s: %v\n", prefix, edge.From, activateArrow, edge.To, text))
	} else {
		sb.WriteString(fmt.Sprintf("%s%s %s %s:\n", prefix, edge.From, activateArrow, edge.To))
	}
}

func mermaidEmitIndentedEdgeAt(sb *strings.Builder, edge *gsl.Edge, indent int) {
	prefix := mermaidIndentPrefix(indent)
	arrow := mermaidGetArrowStyle(edge)
	if text, ok := edge.Attributes["text"]; ok {
		sb.WriteString(fmt.Sprintf("%s%s %s %s: %v\n", prefix, edge.From, arrow, edge.To, text))
	} else {
		sb.WriteString(fmt.Sprintf("%s%s %s %s:\n", prefix, edge.From, arrow, edge.To))
	}
}

func mermaidEmitEdge(sb *strings.Builder, edge *gsl.Edge) {
	arrow := mermaidGetArrowStyle(edge)
	prefix := mermaidIndentPrefix(0)
	if text, ok := edge.Attributes["text"]; ok {
		sb.WriteString(fmt.Sprintf("%s%s %s %s: %v\n", prefix, edge.From, arrow, edge.To, text))
	} else {
		sb.WriteString(fmt.Sprintf("%s%s %s %s:\n", prefix, edge.From, arrow, edge.To))
	}
}

// mermaidGetArrowStyle returns the Mermaid arrow style for the edge.
// Maps semantic arrow names to Mermaid syntax:
//
//	sync     ->>   (solid, arrowhead)          - synchronous call (default)
//	async    -)    (solid, open arrow)         - asynchronous message
//	return   -->>  (dotted, arrowhead)         - return/reply
//	dependency -.-> (dotted, open arrow)       - weak dependency
//	strong    ->>   (solid, arrowhead)         - strong coupling (no double line in Mermaid)
func mermaidGetArrowStyle(edge *gsl.Edge) string {
	arrow, ok := edge.Attributes["arrow"]
	if !ok {
		return "->>"
	}
	switch fmt.Sprintf("%v", arrow) {
	case "async":
		return "-)"
	case "return":
		return "-->>"
	case "dependency":
		return "-.->"
	case "strong":
		return "->>"
	default:
		return "->>"
	}
}

const mermaidBaseIndent = "    "

func mermaidIndentPrefix(level int) string {
	return mermaidBaseIndent + strings.Repeat("    ", level)
}
