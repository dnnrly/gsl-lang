package gsl

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

// Serialize converts a Graph to canonical GSL text.
// Output uses nested block syntax for parent-child relationships.
// Deterministic: sets first (sorted by ID), root nodes next (ordered by
// first appearance in edge list, then by ID), root edges last (in slice order).
// Children are nested under their parent using block syntax.
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

	// Nodes, with nesting
	nodes := g.GetNodes()
	if len(nodes) > 0 {
		sections = append(sections, serializeNodes(nodes, g.GetEdges()))
	}

	// Edges, with nesting built from Parent field at serialization time
	edges := g.GetEdges()
	if len(edges) > 0 {
		sections = append(sections, serializeEdges(edges))
	}

	return strings.Join(sections, "\n\n")
}

// serializeNodes builds the parent-child tree and serializes with recursive nesting.
func serializeNodes(nodes map[string]*Node, edges []*Edge) string {
	positions := buildNodePositions(edges)

	// Build parent->children map, detecting cycles
	children := make(map[string][]*Node)
	for _, n := range nodes {
		if n.Parent != nil {
			if _, ok := nodes[*n.Parent]; ok && !isCyclic(n, nodes) {
				children[*n.Parent] = append(children[*n.Parent], n)
			}
		}
	}

	// Sort children within each parent
	for id := range children {
		sortNodesByPosition(children[id], positions)
	}

	// Determine roots: no parent, parent doesn't exist, or in a cycle
	roots := []*Node{}
	for _, n := range nodes {
		if n.Parent == nil || nodes[*n.Parent] == nil || isCyclic(n, nodes) {
			roots = append(roots, n)
		}
	}
	sortNodesByPosition(roots, positions)

	var lines []string
	for _, n := range roots {
		lines = append(lines, serializeNodeNested(n, children[n.ID], children, make(map[string]bool), false))
	}
	return strings.Join(lines, "\n")
}

// serializeEdges identifies root edges and serializes children inside blocks.
// Parent-child relationships are built from the Parent field at serialization
// time (consistent with how node nesting works), not from the pre-populated
// Children field which may not be set in all code paths.
// Edges in parent cycles are treated as roots with explicit parent attribute.
func serializeEdges(edges []*Edge) string {
	// Build parent label → children map from Parent field
	// Index edges by label for cycle detection
	labelIndex := make(map[string]*Edge)
	for _, e := range edges {
		if e.Label != "" {
			labelIndex[e.Label] = e
		}
	}

	// Build parent label → children map, excluding cyclic edges
	// (cyclic edges are output at root level with explicit parent attribute)
	childMap := make(map[string][]*Edge)
	childSet := make(map[*Edge]bool)
	for _, e := range edges {
		if e.Parent != "" {
			parentExists := labelIndex[e.Parent] != nil
			cyclic := parentExists && isEdgeCyclic(e, labelIndex)
			if parentExists && !cyclic {
				childMap[e.Parent] = append(childMap[e.Parent], e)
				childSet[e] = true
			}
		}
	}

	// Group labeled root edges by label for grouped-edge serialization.
	// Unlabeled root edges are serialized individually via serializeEdgeNested.
	labelGroups := make(map[string][]*Edge)
	for _, e := range edges {
		if !childSet[e] && e.Label != "" {
			labelGroups[e.Label] = append(labelGroups[e.Label], e)
		}
	}

	var lines []string
	visiting := make(map[string]bool)
	seenLabels := make(map[string]bool)

	for _, e := range edges {
		if childSet[e] {
			continue
		}
		if e.Label != "" {
			if !seenLabels[e.Label] {
				seenLabels[e.Label] = true
				lines = append(lines, serializeLabeledEdgeGroup(labelGroups[e.Label], childMap, visiting))
			}
		} else {
			lines = append(lines, serializeEdgeNested(e, childMap, visiting))
		}
	}
	return strings.Join(lines, "\n")
}

// isEdgeCyclic walks the edge parent chain to detect cycles.
func isEdgeCyclic(e *Edge, labelIndex map[string]*Edge) bool {
	visited := make(map[string]bool)
	current := e
	for current != nil && current.Parent != "" {
		if current.Label != "" {
			if visited[current.Label] {
				return true
			}
			visited[current.Label] = true
		}
		parent, ok := labelIndex[current.Parent]
		if !ok {
			return false
		}
		current = parent
	}
	return false
}

// buildNodePositions returns node ID to first edge index referencing it.
func buildNodePositions(edges []*Edge) map[string]int {
	positions := make(map[string]int)
	for i, e := range edges {
		if _, ok := positions[e.From]; !ok {
			positions[e.From] = i
		}
		if _, ok := positions[e.To]; !ok {
			positions[e.To] = i
		}
	}
	return positions
}

// sortNodesByPosition sorts nodes by first edge appearance, then ID.
func sortNodesByPosition(nodes []*Node, positions map[string]int) {
	sort.SliceStable(nodes, func(i, j int) bool {
		pi, hasI := positions[nodes[i].ID]
		pj, hasJ := positions[nodes[j].ID]
		if hasI && hasJ {
			if pi != pj {
				return pi < pj
			}
			return nodes[i].ID < nodes[j].ID
		}
		if hasI {
			return true
		}
		if hasJ {
			return false
		}
		return nodes[i].ID < nodes[j].ID
	})
}

// isCyclic walks the parent chain to detect cycles.
func isCyclic(n *Node, allNodes map[string]*Node) bool {
	visited := make(map[string]bool)
	current := n
	for current != nil && current.Parent != nil {
		if visited[current.ID] {
			return true
		}
		visited[current.ID] = true
		parent, ok := allNodes[*current.Parent]
		if !ok {
			return false
		}
		current = parent
	}
	return false
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

// serializeNodeNested serializes a node with optional block children.
// stripParent controls whether the "parent" attribute is omitted (true when
// inside a block where parent is implicit).
func serializeNodeNested(n *Node, children []*Node, childMap map[string][]*Node, visiting map[string]bool, stripParent bool) string {
	var b strings.Builder
	b.WriteString("node ")
	b.WriteString(n.ID)

	// Build attrs, omitting parent when inside a block
	var attrs map[string]interface{}
	if stripParent {
		attrs = make(map[string]interface{})
		for k, v := range n.Attributes {
			if k != "parent" {
				attrs[k] = v
			}
		}
	} else {
		attrs = n.Attributes
	}
	if len(attrs) > 0 {
		b.WriteString(" ")
		b.WriteString(serializeAttrs(attrs))
	}

	// Block children (if any and no cycle)
	if len(children) > 0 && !visiting[n.ID] {
		visiting[n.ID] = true
		b.WriteString(" {\n")
		for _, child := range children {
			grandchildren := childMap[child.ID]
			childStr := serializeNodeNested(child, grandchildren, childMap, visiting, true)
			b.WriteString(indentBlock(childStr, "    "))
			b.WriteString("\n")
		}
		b.WriteString("}")
		visiting[n.ID] = false
	}

	// Memberships must come after block (parser expects this order)
	b.WriteString(serializeSetMemberships(n.Sets))

	return b.String()
}

func serializeEdgeNested(e *Edge, childMap map[string][]*Edge, visiting map[string]bool) string {
	var b strings.Builder

	if e.Label != "" {
		b.WriteString(e.Label)
		b.WriteString(": ")
	}

	b.WriteString(e.From)
	b.WriteString("->")
	b.WriteString(e.To)

	if len(e.Attributes) > 0 {
		b.WriteString(" ")
		b.WriteString(serializeAttrs(e.Attributes))
	}
	b.WriteString(serializeSetMemberships(e.Sets))

	if children, ok := childMap[e.Label]; ok && len(children) > 0 && !visiting[e.Label] {
		visiting[e.Label] = true
		b.WriteString(" {\n")
		for _, child := range children {
			childStr := serializeEdgeNested(child, childMap, visiting)
			b.WriteString(indentBlock(childStr, "    "))
			b.WriteString("\n")
		}
		b.WriteString("}")
		visiting[e.Label] = false
	}

	return b.String()
}

// serializeLabeledEdgeGroup serializes a group of edges sharing the same label
// as a single grouped-edge declaration. It detects whether the edges all share
// To (left-grouped: label: from1,from2,...->to) or all share From (right-grouped:
// label: from->to1,to2,...). This ensures the output can be re-parsed without
// triggering the duplicate edge label error.
func serializeLabeledEdgeGroup(edges []*Edge, childMap map[string][]*Edge, visiting map[string]bool) string {
	if len(edges) == 0 {
		return ""
	}

	var b strings.Builder
	label := edges[0].Label

	// Detect grouping pattern: all share To (left-grouped) or all share From (right-grouped)
	allShareTo := true
	for _, e := range edges[1:] {
		if e.To != edges[0].To {
			allShareTo = false
			break
		}
	}

	allShareFrom := true
	for _, e := range edges[1:] {
		if e.From != edges[0].From {
			allShareFrom = false
			break
		}
	}

	b.WriteString(label)
	b.WriteString(": ")

	if allShareTo {
		// Left-grouped: label: from1,from2,...->to
		for i, e := range edges {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(e.From)
		}
		b.WriteString("->")
		b.WriteString(edges[0].To)
	} else if allShareFrom {
		// Right-grouped: label: from->to1,to2,...
		b.WriteString(edges[0].From)
		b.WriteString("->")
		for i, e := range edges {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(e.To)
		}
	} else {
		// Fallback: shouldn't happen for parser-produced graphs
		// (valid GSL produces edges that share From or To within a label)
		// Serialize individually; re-parse may fail for API-constructed graphs.
		var lines []string
		for _, e := range edges {
			lines = append(lines, serializeEdgeNested(e, childMap, visiting))
		}
		return strings.Join(lines, "\n")
	}

	// Attributes (all edges from the same declaration share attributes)
	if len(edges[0].Attributes) > 0 {
		b.WriteString(" ")
		b.WriteString(serializeAttrs(edges[0].Attributes))
	}

	// Set memberships
	b.WriteString(serializeSetMemberships(edges[0].Sets))

	// Children (scoped block) — looked up by label
	if children, ok := childMap[label]; ok && len(children) > 0 && !visiting[label] {
		visiting[label] = true
		b.WriteString(" {\n")
		for _, child := range children {
			childStr := serializeEdgeNested(child, childMap, visiting)
			b.WriteString(indentBlock(childStr, "    "))
			b.WriteString("\n")
		}
		b.WriteString("}")
		visiting[label] = false
	}

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
		if key == "parent" {
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

func indentBlock(s string, indent string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}
