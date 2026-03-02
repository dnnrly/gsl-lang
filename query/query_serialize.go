package query

import (
	"fmt"
	"strconv"
	"strings"
)

// SerializeQuery converts a Query AST back to query string
func SerializeQuery(q *Query) string {
	if q == nil || q.Root == nil {
		return ""
	}
	return serializeStep(q.Root)
}

func serializeStep(step Step) string {
	switch s := step.(type) {
	case *Pipeline:
		return serializePipeline(s)
	case *StartStep:
		return serializeStartStep(s)
	case *FlowStep:
		return serializeFlowStep(s)
	case *FilterStep:
		return serializeFilterStep(s)
	case *MinusStep:
		return serializeMinusStep(s)
	case *CombinatorExpr:
		return serializeCombinatorExpr(s)
	default:
		return ""
	}
}

func serializePipeline(p *Pipeline) string {
	if p == nil || len(p.Steps) == 0 {
		return ""
	}

	var parts []string
	for _, step := range p.Steps {
		parts = append(parts, serializeStep(step))
	}
	return strings.Join(parts, " | ")
}

func serializeStartStep(s *StartStep) string {
	var ids []string
	for _, id := range s.NodeIDs {
		// Quote if necessary (contains special chars or is a keyword)
		if needsQuoting(id) {
			ids = append(ids, fmt.Sprintf("\"%s\"", escapeQueryString(id)))
		} else {
			ids = append(ids, id)
		}
	}
	return "start " + strings.Join(ids, ", ")
}

func serializeFlowStep(f *FlowStep) string {
	var buf strings.Builder
	buf.WriteString("flow ")
	buf.WriteString(f.Direction)

	if f.EdgeFilter != nil {
		buf.WriteString(" where edge.")
		buf.WriteString(f.EdgeFilter.Attr)
		buf.WriteString(" ")
		buf.WriteString(f.EdgeFilter.Op)
		buf.WriteString(" ")
		buf.WriteString(serializeValue(f.EdgeFilter.Value))
	}

	if f.Recursive {
		buf.WriteString(" recursive")
	}

	return buf.String()
}

func serializeFilterStep(f *FilterStep) string {
	var buf strings.Builder
	buf.WriteString("where ")
	buf.WriteString(f.Filter.Attr)
	buf.WriteString(" ")
	buf.WriteString(f.Filter.Op)
	buf.WriteString(" ")
	buf.WriteString(serializeValue(f.Filter.Value))
	return buf.String()
}

func serializeMinusStep(m *MinusStep) string {
	return "minus (" + serializePipeline(m.Pipeline) + ")"
}

func serializeCombinatorExpr(c *CombinatorExpr) string {
	var buf strings.Builder

	if c.IsParens {
		buf.WriteString("(")
		buf.WriteString(serializePipeline(c.Left))
		buf.WriteString(")")
	} else {
		buf.WriteString(serializePipeline(c.Left))
	}

	buf.WriteString(" ")
	buf.WriteString(c.Type)
	buf.WriteString(" ")

	if c.IsParens {
		buf.WriteString("(")
		buf.WriteString(serializePipeline(c.Right))
		buf.WriteString(")")
	} else {
		buf.WriteString(serializePipeline(c.Right))
	}

	return buf.String()
}

func serializeValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", escapeQueryString(val))
	case float64:
		// Format as integer if it's a whole number
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprint(val)
	}
}

func escapeQueryString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	// Check if it's a keyword
	switch s {
	case "start", "flow", "where", "minus", "union", "intersect",
		"in", "out", "both", "recursive", "edge", "contains", "matches",
		"node", "set", "true", "false":
		return true
	}

	// Check if it starts with valid identifier char
	first := rune(s[0])
	if !((first >= 'A' && first <= 'Z') || (first >= 'a' && first <= 'z') || first == '_') {
		return true
	}

	// Check remaining chars
	for _, r := range s[1:] {
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') || r == '_') {
			return true
		}
	}

	return false
}
