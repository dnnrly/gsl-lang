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
	sets := g.GetSets()
	if len(sets) > 0 {
		setIDs := make([]string, 0, len(sets))
		for id := range sets {
			setIDs = append(setIDs, id)
		}
		sort.Strings(setIDs)

		var lines []string
		for _, id := range setIDs {
			lines = append(lines, serializeSet(sets[id]))
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}

	// Nodes, sorted by ID
	nodes := g.GetNodes()
	if len(nodes) > 0 {
		nodeIDs := make([]string, 0, len(nodes))
		for id := range nodes {
			nodeIDs = append(nodeIDs, id)
		}
		sort.Strings(nodeIDs)

		var lines []string
		for _, id := range nodeIDs {
			lines = append(lines, serializeNode(nodes[id]))
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}

	// Edges, in slice order
	edges := g.GetEdges()
	if len(edges) > 0 {
		var lines []string
		for _, e := range edges {
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
	b.WriteString(serializeSetMemberships(n.Sets))
	return b.String()
}

func serializeEdge(e *Edge) string {
	var b strings.Builder

	// Output edge label if present
	if e.Label != "" {
		b.WriteString(e.Label)
		b.WriteString(": ")
	}

	b.WriteString(e.From)
	b.WriteString("->")
	b.WriteString(e.To)

	// Build attributes including depends_on
	attrs := make(map[string]interface{})
	for k, v := range e.Attributes {
		attrs[k] = v
	}
	if e.DependsOn != "" {
		attrs["depends_on"] = e.DependsOn
	}

	if len(attrs) > 0 {
		b.WriteString(" ")
		b.WriteString(serializeAttrs(attrs))
	}
	b.WriteString(serializeSetMemberships(e.Sets))
	return b.String()
}

func serializeSetMemberships(sets map[string]struct{}) string {
	if len(sets) == 0 {
		return ""
	}
	setNames := make([]string, 0, len(sets))
	for name := range sets {
		setNames = append(setNames, name)
	}
	sort.Strings(setNames)
	var b strings.Builder
	for _, name := range setNames {
		b.WriteString(" @")
		b.WriteString(name)
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
		// Special case: depends_on outputs as identifier (no quotes)
		if key == "depends_on" {
			return key + "=" + v
		}
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
