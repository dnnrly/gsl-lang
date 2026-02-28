package gsl

import "fmt"

type builder struct {
	graph    *Graph
	warnings []error
	errors   []error
}

func buildGraph(prog *program) (*Graph, []error, error) {
	b := &builder{
		graph: &Graph{
			Nodes: make(map[string]*Node),
			Sets:  make(map[string]*Set),
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

	if len(b.errors) > 0 {
		return b.graph, b.warnings, b.errors[0]
	}
	return b.graph, b.warnings, nil
}

func (b *builder) ensureNode(name string) *Node {
	if n, ok := b.graph.Nodes[name]; ok {
		return n
	}
	n := &Node{
		ID:         name,
		Attributes: make(map[string]interface{}),
		Sets:       make(map[string]struct{}),
	}
	b.graph.Nodes[name] = n
	return n
}

func (b *builder) ensureSet(name string, implicit bool) *Set {
	if s, ok := b.graph.Sets[name]; ok {
		return s
	}
	s := &Set{
		ID:         name,
		Attributes: make(map[string]interface{}),
	}
	b.graph.Sets[name] = s
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

	// Cache Parent field
	if p, ok := node.Attributes["parent"]; ok {
		if ref, isRef := p.(NodeRef); isRef {
			s := string(ref)
			node.Parent = &s
		}
	}

	// Memberships
	for _, setName := range nd.memberships {
		b.ensureSet(setName, !b.setDeclared(setName))
		node.Sets[setName] = struct{}{}
	}

	// Check node/set name collision
	if _, ok := b.graph.Sets[nd.name]; ok {
		b.warnings = append(b.warnings, fmt.Errorf("node and set name collision: %q", nd.name))
	}

	// Process block children
	parentName := nd.name
	for i := range nd.block {
		b.processNodeDecl(&nd.block[i], &parentName)
	}
}

func (b *builder) processEdgeDecl(ed *edgeDecl) {
	if len(ed.left) > 1 && len(ed.right) > 1 {
		b.errors = append(b.errors, fmt.Errorf("%d:%d: grouped edges on both sides", ed.line, ed.col))
		return
	}

	for _, from := range ed.left {
		for _, to := range ed.right {
			b.ensureNode(from)
			b.ensureNode(to)

			edge := &Edge{
				From:       from,
				To:         to,
				Attributes: make(map[string]interface{}),
				Sets:       make(map[string]struct{}),
			}

			// Text shorthand
			if ed.textValue != nil {
				edge.Attributes["text"] = *ed.textValue
			}

			// Attributes
			for _, attr := range ed.attrs {
				v := convertAttrValue(attr.value)
				if _, isRef := v.(NodeRef); isRef {
					b.errors = append(b.errors, fmt.Errorf("%d:%d: NodeRef not allowed in edge attribute %q", attr.line, attr.col, attr.key))
					continue
				}
				edge.Attributes[attr.key] = v
			}

			// Memberships
			for _, setName := range ed.memberships {
				b.ensureSet(setName, !b.setDeclared(setName))
				edge.Sets[setName] = struct{}{}
			}

			b.graph.Edges = append(b.graph.Edges, edge)
		}
	}
}

func (b *builder) processSetDecl(sd *setDecl) {
	// Check node/set name collision
	if _, ok := b.graph.Nodes[sd.name]; ok {
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
	if s, ok := b.graph.Sets[name]; ok {
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
