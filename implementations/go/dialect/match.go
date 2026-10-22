// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

package dialect

import (
	"fmt"
	"sort"

	gsl "github.com/dnnrly/gsl-lang/implementations/go"
	"github.com/dnnrly/gsl-lang/implementations/go/query"
)

// selectedElement is a graph element selected for validation.
type selectedElement struct {
	ref        string
	attributes gsl.AttributeMap
}

// selectElements evaluates the rule's match query against g and returns the
// eligible elements for its representation.
//
// Matching determines which elements are subject to the rule's constraints; it
// does not establish that they are valid. Elements whose attributes later fail
// validation remain selected.
func (r *Rule) selectElements(g *gsl.Graph) ([]selectedElement, error) {
	ctx := &query.QueryContext{
		InputGraph:  g,
		NamedGraphs: map[string]*gsl.Graph{},
	}

	result, err := r.expr.Apply(ctx, query.GraphValue{Graph: g})
	if err != nil {
		return nil, err
	}
	gv, ok := result.(query.GraphValue)
	if !ok {
		return nil, fmt.Errorf("unexpected match result type %T", result)
	}
	matched := gv.Graph

	switch r.Representation {
	case RepresentationNode:
		nodes := matched.GetNodes()
		ids := make([]string, 0, len(nodes))
		for id := range nodes {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		out := make([]selectedElement, 0, len(ids))
		for _, id := range ids {
			n := nodes[id]
			out = append(out, selectedElement{ref: n.ID, attributes: n.Attributes})
		}
		return out, nil

	case RepresentationEdge:
		edges := matched.GetEdges()
		out := make([]selectedElement, 0, len(edges))
		for _, e := range edges {
			out = append(out, selectedElement{ref: edgeRef(e), attributes: e.Attributes})
		}
		return out, nil
	}

	return nil, fmt.Errorf("unsupported representation %q", r.Representation)
}

func edgeRef(e *gsl.Edge) string {
	if e.Label != "" {
		return e.Label
	}
	return e.From + "->" + e.To
}
