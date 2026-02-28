package gsl

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

// Serialize converts a Graph to canonical GSL text.
// Output is deterministic: sets first (sorted by ID), nodes second (sorted by ID),
// edges last (in slice order). Attribute keys are sorted alphabetically.
func Serialize(g *Graph) string {
	if g == nil {
		return ""
	}

	var sections []string

	// Sets, sorted by ID
	if len(g.Sets) > 0 {
		setIDs := make([]string, 0, len(g.Sets))
		for id := range g.Sets {
			setIDs = append(setIDs, id)
		}
		sort.Strings(setIDs)

		var lines []string
		for _, id := range setIDs {
			lines = append(lines, serializeSet(g.Sets[id]))
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}

	// Nodes, sorted by ID
	if len(g.Nodes) > 0 {
		nodeIDs := make([]string, 0, len(g.Nodes))
		for id := range g.Nodes {
			nodeIDs = append(nodeIDs, id)
		}
		sort.Strings(nodeIDs)

		var lines []string
		for _, id := range nodeIDs {
			lines = append(lines, serializeNode(g.Nodes[id]))
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}

	// Edges, in slice order
	if len(g.Edges) > 0 {
		var lines []string
		for _, e := range g.Edges {
			lines = append(lines, serializeEdge(e))
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}

	return strings.Join(sections, "\n\n")
}

func serializeSet(s *Set) string {
	var b strings.Builder
	b.WriteString("set ")
	b.WriteString(s.ID)
	if len(s.Attributes) > 0 {
		b.WriteString(" ")
		b.WriteString(serializeAttrs(s.Attributes))
	}
	return b.String()
}

func serializeNode(n *Node) string {
	var b strings.Builder
	b.WriteString("node ")
	b.WriteString(n.ID)
	if len(n.Attributes) > 0 {
		b.WriteString(" ")
		b.WriteString(serializeAttrs(n.Attributes))
	}
	if len(n.Sets) > 0 {
		setNames := make([]string, 0, len(n.Sets))
		for name := range n.Sets {
			setNames = append(setNames, name)
		}
		sort.Strings(setNames)
		for _, name := range setNames {
			b.WriteString(" @")
			b.WriteString(name)
		}
	}
	return b.String()
}

func serializeEdge(e *Edge) string {
	var b strings.Builder
	b.WriteString(e.From)
	b.WriteString("->")
	b.WriteString(e.To)
	if len(e.Attributes) > 0 {
		b.WriteString(" ")
		b.WriteString(serializeAttrs(e.Attributes))
	}
	if len(e.Sets) > 0 {
		setNames := make([]string, 0, len(e.Sets))
		for name := range e.Sets {
			setNames = append(setNames, name)
		}
		sort.Strings(setNames)
		for _, name := range setNames {
			b.WriteString(" @")
			b.WriteString(name)
		}
	}
	return b.String()
}

func serializeAttrs(attrs map[string]interface{}) string {
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, serializeAttr(k, attrs[k]))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func serializeAttr(key string, value interface{}) string {
	if value == nil {
		return key
	}
	switch v := value.(type) {
	case string:
		return key + "=" + `"` + escapeString(v) + `"`
	case float64:
		return key + "=" + formatNumber(v)
	case bool:
		if v {
			return key + "=true"
		}
		return key + "=false"
	case NodeRef:
		return key + "=" + string(v)
	default:
		return key
	}
}

func escapeString(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func formatNumber(f float64) string {
	if f == math.Trunc(f) && !math.IsInf(f, 0) && !math.IsNaN(f) {
		return strconv.FormatFloat(f, 'f', 0, 64)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}
