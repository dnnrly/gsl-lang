package gsl

import "fmt"

type builder struct {
	graph     *Graph
	warnings  []error
	errors    []error
	edgeScope *edgeScope // current edge scope for label resolution
}

// edgeScope tracks edge labels and parent relationships for dependency resolution
type edgeScope struct {
	labels map[string]*Edge // label -> edge mapping for current scope
	parent string           // label of parent edge (for implicit dependencies)
	outer  *edgeScope       // outer scope for nested blocks
}

func buildGraph(prog *program) (*Graph, []error, error) {
	b := &builder{
		graph: &Graph{
			nodes: make(map[string]*Node),
			sets:  make(map[string]*Set),
		},
		edgeScope: &edgeScope{
			labels: make(map[string]*Edge),
			parent: "",
			outer:  nil,
		},
	}

	for _, stmt := range prog.statements {
		switch s := stmt.(type) {
		case *nodeDecl:
			b.processNodeDecl(s, nil)
		case *edgeDecl:
			b.processEdgeDecl(s)
		case *setDecl:
			b.processSetDecl(s)
		}
	}

	// Validate all edge dependencies (labels must exist)
	b.validateEdgeDependencies()

	if len(b.errors) > 0 {
		return b.graph, b.warnings, b.errors[0]
	}
	return b.graph, b.warnings, nil
}

func (b *builder) ensureNode(name string) *Node {
	if n, ok := b.graph.nodes[name]; ok {
		return n
	}
	n := &Node{
		ID:         name,
		Attributes: make(AttributeMap),
		Sets:       make(map[string]struct{}),
	}
	b.graph.nodes[name] = n
	return n
}

func (b *builder) ensureSet(name string, implicit bool) *Set {
	if s, ok := b.graph.sets[name]; ok {
		return s
	}
	s := &Set{
		ID:         name,
		Attributes: make(AttributeMap),
	}
	b.graph.sets[name] = s
	if implicit {
		b.warnings = append(b.warnings, fmt.Errorf("implicit set creation: %q", name))
	}
	return s
}

func (b *builder) processNodeDecl(nd *nodeDecl, enclosingParent *string) {
	node := b.ensureNode(nd.name)

	// Text shorthand
	if nd.textValue != nil {
		node.Attributes["text"] = *nd.textValue
	}

	// Attributes
	for _, attr := range nd.attrs {
		node.Attributes[attr.key] = convertAttrValue(attr.value)
	}

	// Implicit parent from block
	if enclosingParent != nil {
		if _, hasExplicit := node.Attributes["parent"]; !hasExplicit {
			node.Attributes["parent"] = NodeRef(*enclosingParent)
		} else {
			b.warnings = append(b.warnings, fmt.Errorf("%d:%d: parent override inside block", nd.line, nd.col))
		}
	}

	// Cache Parent field using typed accessor
	if ref, ok := node.GetRef("parent"); ok {
		s := string(*ref)
		node.Parent = &s
	}

	// Memberships
	for _, setName := range nd.memberships {
		b.ensureSet(setName, !b.setDeclared(setName))
		node.Sets[setName] = struct{}{}
	}

	// Check node/set name collision
	if _, ok := b.graph.sets[nd.name]; ok {
		b.warnings = append(b.warnings, fmt.Errorf("node and set name collision: %q", nd.name))
	}

	// Process block children
	parentName := nd.name
	for i := range nd.block {
		b.processNodeDecl(&nd.block[i], &parentName)
	}
}

func (b *builder) processEdgeDecl(ed *edgeDecl) {
	b.processEdgeDeclWithScope(ed, false)
}

func (b *builder) processEdgeDeclWithScope(ed *edgeDecl, insideScope bool) {
	if len(ed.left) > 1 && len(ed.right) > 1 {
		b.errors = append(b.errors, fmt.Errorf("%d:%d: grouped edges on both sides", ed.line, ed.col))
		return
	}

	// Check for label uniqueness within current scope
	var edgeLabel string
	if ed.label != nil {
		edgeLabel = *ed.label
		if b.edgeScope != nil {
			if _, exists := b.edgeScope.labels[edgeLabel]; exists {
				b.errors = append(b.errors, fmt.Errorf("%d:%d: duplicate edge label %q", ed.line, ed.col, edgeLabel))
				return
			}
		}
	}

	// Check for explicit depends_on inside scoped block (invalid per spec)
	if insideScope {
		for _, attr := range ed.attrs {
			if attr.key == "depends_on" {
				b.errors = append(b.errors, fmt.Errorf("%d:%d: explicit depends_on not allowed inside scoped edge", attr.line, attr.col))
				return
			}
		}
	}

	// Create edges
	var parentEdge *Edge
	for _, from := range ed.left {
		for _, to := range ed.right {
			b.ensureNode(from)
			b.ensureNode(to)

			edge := &Edge{
				From:       from,
				To:         to,
				Label:      edgeLabel,
				Attributes: make(AttributeMap),
				Sets:       make(map[string]struct{}),
			}

			// Text shorthand
			if ed.textValue != nil {
				edge.Attributes["text"] = *ed.textValue
			}

			// Attributes (excluding depends_on which is handled separately)
			for _, attr := range ed.attrs {
				if attr.key == "depends_on" {
					// Store the dependency reference
					if attr.value != nil {
						edge.DependsOn = attr.value.strVal
					}
					continue
				}
				v := convertAttrValue(attr.value)
				if _, isRef := v.(NodeRef); isRef {
					b.errors = append(b.errors, fmt.Errorf("%d:%d: NodeRef not allowed in edge attribute %q", attr.line, attr.col, attr.key))
					continue
				}
				edge.Attributes[attr.key] = v
			}

			// If inside a scope, set implicit dependency on parent
			if insideScope && b.edgeScope != nil && b.edgeScope.parent != "" {
				edge.DependsOn = b.edgeScope.parent
			}

			// Memberships
			for _, setName := range ed.memberships {
				b.ensureSet(setName, !b.setDeclared(setName))
				edge.Sets[setName] = struct{}{}
			}

			b.graph.edges = append(b.graph.edges, edge)

			// Track the first edge as parent for scope purposes
			if parentEdge == nil {
				parentEdge = edge
			}
		}
	}

	// Register label for this edge (if labeled) in current scope
	if parentEdge != nil && edgeLabel != "" {
		if b.edgeScope != nil {
			b.edgeScope.labels[edgeLabel] = parentEdge
		}
	}

	// Process scoped block if present
	if len(ed.block) > 0 {
		b.processScopedBlock(ed.block, edgeLabel)
	}
}

// processScopedBlock processes statements inside an edge scope
func (b *builder) processScopedBlock(stmts []statement, parentLabel string) {
	// Create new scope
	newScope := &edgeScope{
		labels: make(map[string]*Edge),
		parent: parentLabel,
		outer:  b.edgeScope,
	}

	// Save old scope and set new one
	oldScope := b.edgeScope
	b.edgeScope = newScope

	// Process all statements in the block
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *nodeDecl:
			b.processNodeDecl(s, nil)
		case *edgeDecl:
			b.processEdgeDeclWithScope(s, true) // true = inside scope
		case *setDecl:
			b.processSetDecl(s)
		}
	}

	// Restore outer scope
	b.edgeScope = oldScope
}

// validateEdgeDependencies checks that all depends_on references resolve to valid labels
func (b *builder) validateEdgeDependencies() {
	// Collect all edge labels in the graph
	allLabels := make(map[string]bool)
	for _, edge := range b.graph.edges {
		if edge.Label != "" {
			allLabels[edge.Label] = true
		}
	}

	// Validate dependencies
	for _, edge := range b.graph.edges {
		if edge.DependsOn != "" {
			if !allLabels[edge.DependsOn] {
				// Find the edge's position for error reporting
				// We report on the edge that has the bad dependency
				b.errors = append(b.errors, fmt.Errorf("edge depends_on references unknown label %q", edge.DependsOn))
			}
		}
	}
}

func (b *builder) processSetDecl(sd *setDecl) {
	// Check node/set name collision
	if _, ok := b.graph.nodes[sd.name]; ok {
		b.warnings = append(b.warnings, fmt.Errorf("node and set name collision: %q", sd.name))
	}

	s := b.ensureSet(sd.name, false)
	s.declared = true

	for _, attr := range sd.attrs {
		v := convertAttrValue(attr.value)
		if _, isRef := v.(NodeRef); isRef {
			b.errors = append(b.errors, fmt.Errorf("%d:%d: NodeRef not allowed in set attribute %q", attr.line, attr.col, attr.key))
			continue
		}
		s.Attributes[attr.key] = v
	}
}

// setDeclared checks whether a set was explicitly declared (not just implicitly created).
func (b *builder) setDeclared(name string) bool {
	if s, ok := b.graph.sets[name]; ok {
		return s.declared
	}
	return false
}

func convertAttrValue(v *attrValue) interface{} {
	if v == nil {
		return nil
	}
	switch v.kind {
	case valueString:
		return v.strVal
	case valueNumber:
		return v.numVal
	case valueBool:
		return v.boolVal
	case valueNodeRef:
		return NodeRef(v.refVal)
	}
	return nil
}
