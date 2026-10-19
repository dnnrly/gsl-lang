// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

// Package dialect implements GSL dialect support: loading and validating
// dialect documents, matching graph elements with a restricted GQL subset,
// projecting GSL attribute values into JSON, and checking the conformance of a
// graph against JSON Schema attribute constraints.
//
// The normative behaviour is defined by spec/dialects/DIALECTS_SPEC.md.
//
// The pipeline is deliberately staged and independently testable:
//
//	Load         parse and validate a dialect document (all-or-nothing)
//	Match        evaluate the definition's GQL query and select eligible elements
//	Project      convert matched attribute maps into JSON-compatible values
//	Validate     check projected values against the definition's JSON Schema
//
// Graph matching and value projection never depend on the JSON Schema library.
package dialect

import (
	"io"
	"os"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	gsl "github.com/dnnrly/gsl-lang/implementations/go"
	"github.com/dnnrly/gsl-lang/implementations/go/query"
)

// Representation identifies the kind of GSL element an element definition
// applies to.
type Representation string

const (
	// RepresentationNode selects nodes as the elements to validate.
	RepresentationNode Representation = "node"
	// RepresentationEdge selects edges as the elements to validate.
	RepresentationEdge Representation = "edge"
	// RepresentationSet selects sets as the elements to validate.
	RepresentationSet Representation = "set"
)

func (r Representation) valid() bool {
	switch r {
	case RepresentationNode, RepresentationEdge, RepresentationSet:
		return true
	}
	return false
}

// ElementDefinition is a named definition within a dialect describing a
// category of graph elements, its matching rule and its attribute constraints.
// The name is a dialect-local identifier used for readability and diagnostics;
// it does not create a new GSL element type.
type ElementDefinition struct {
	// Name is the dialect-local name of the definition.
	Name string

	// Representation is the kind of GSL element this definition validates.
	Representation Representation

	// Description is a human-readable explanation of the definition.
	Description string

	// Guidance is descriptive advice for constructing or interpreting graph
	// elements of this definition. It is never an executable constraint.
	Guidance string

	// Query is the raw GQL match query as written in the dialect document.
	Query string

	expr   query.Expression
	schema *jsonschema.Schema
}

// Dialect is a validated dialect document.
type Dialect struct {
	// ID is the stable identifier of the dialect.
	ID string

	// Version is the version of the dialect definition.
	Version string

	// Title is a human-readable name for the dialect.
	Title string

	// Description describes the dialect's purpose.
	Description string

	// Guidance is general advice for constructing and interpreting graphs.
	Guidance string

	elements []*ElementDefinition
}

// Elements returns the element definitions in deterministic (name-sorted)
// order. The returned slice is a copy; mutating it does not affect the dialect.
func (d *Dialect) Elements() []*ElementDefinition {
	out := make([]*ElementDefinition, len(d.elements))
	copy(out, d.elements)
	return out
}

// Load parses and validates a dialect document from YAML. It returns a
// *LoadError if the document is not a valid dialect.
//
// Loading is all-or-nothing: when Load returns an error, no dialect is
// returned and no part of the document should be applied.
func Load(r io.Reader) (*Dialect, error) {
	return load(r)
}

// LoadFile reads and validates a dialect document from the named file.
func LoadFile(path string) (*Dialect, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return load(f)
}

// Selected is one graph element selected by one element definition, together
// with the element's attributes projected into JSON-compatible values.
//
// Matching and projection are deterministic and independent of JSON Schema
// validation, so Selected can be inspected on its own (for example by the
// `project` command or the conformance suite).
type Selected struct {
	// Definition is the name of the element definition that selected the
	// element.
	Definition string

	// Representation is the kind of the selected element.
	Representation Representation

	// Element identifies the selected graph element (a node id, a set id, or
	// an edge descriptor such as "a->b").
	Element string

	// Attributes is the projected attribute map that would be validated
	// against the definition's schema.
	Attributes map[string]any
}

// Match evaluates every element definition's match query against g and returns
// the selected elements with their projected attributes, in a deterministic
// order (definitions by name, then nodes and sets by id, then edges in graph
// order).
//
// A non-nil error indicates that matching or projection could not be completed
// and is a *MatchingError or *ProjectionError; in that case the result MUST NOT
// be treated as conforming.
func (d *Dialect) Match(g *gsl.Graph) ([]Selected, error) {
	var out []Selected
	for _, el := range d.elements {
		selected, err := el.selectElements(g)
		if err != nil {
			return nil, &MatchingError{Definition: el.Name, Query: el.Query, Err: err}
		}
		for _, elem := range selected {
			projected, err := projectAttributes(elem.attributes)
			if err != nil {
				return nil, &ProjectionError{Definition: el.Name, Element: elem.ref, Err: err}
			}
			out = append(out, Selected{
				Definition:     el.Name,
				Representation: el.Representation,
				Element:        elem.ref,
				Attributes:     projected,
			})
		}
	}
	return out, nil
}

// Validate checks graph g for conformance to the dialect.
//
// A nil error means validation completed: the returned Report reports either
// conformance (Report.Outcome == OutcomeConforming) or one or more violations
// (OutcomeNonConforming). A non-nil error means validation could not be
// completed, for example because a match query failed to evaluate; in that case
// the graph MUST NOT be reported as conforming.
func (d *Dialect) Validate(g *gsl.Graph) (*Report, error) {
	selected, err := d.Match(g)
	if err != nil {
		return nil, err
	}

	byName := make(map[string]*ElementDefinition, len(d.elements))
	for _, el := range d.elements {
		byName[el.Name] = el
	}

	report := &Report{Outcome: OutcomeConforming}
	for _, s := range selected {
		el := byName[s.Definition]
		if verr := el.schema.Validate(s.Attributes); verr != nil {
			schemaErr, ok := verr.(*jsonschema.ValidationError)
			if !ok {
				return nil, &ProjectionError{Definition: el.Name, Element: s.Element, Err: verr}
			}
			report.Violations = append(report.Violations, violationsFromJSON(schemaErr, el, s.Element)...)
		}
	}

	if len(report.Violations) > 0 {
		report.Outcome = OutcomeNonConforming
	}
	return report, nil
}
