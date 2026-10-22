// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

// Package dialect implements GSL dialect support: loading and validating
// dialect documents, matching graph elements with a restricted GQL subset,
// projecting GSL attribute values into JSON, and checking the conformance of a
// graph against JSON Schema attribute constraints and set declaration policy.
//
// The normative behaviour is defined by spec/dialects/DIALECTS_SPEC.md.
//
// The pipeline is deliberately staged and independently testable:
//
//	Load         parse and validate a dialect document (all-or-nothing)
//	Match        evaluate node/edge rules and select eligible elements
//	Project      convert matched attribute maps into JSON-compatible values
//	Validate     check projected values and set policy
//
// Graph matching and value projection never depend on the JSON Schema library.
package dialect

import (
	"io"
	"os"
	"sort"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	gsl "github.com/dnnrly/gsl-lang/implementations/go"
	"github.com/dnnrly/gsl-lang/implementations/go/query"
)

// Representation identifies the kind of GSL element a rule or set definition
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

// Rule is an entry in the ordered rules sequence of a node or edge section.
// The name is an optional dialect-local label used for readability and
// diagnostics; it does not create a new GSL element type.
type Rule struct {
	// Name is the optional dialect-local name of the rule.
	Name string

	// Representation is the kind of GSL element this rule validates.
	Representation Representation

	// Index is the zero-based position of the rule within its section.
	Index int

	// Description is a human-readable explanation of the rule.
	Description string

	// Guidance is descriptive advice for constructing or interpreting graph
	// elements selected by this rule. It is never an executable constraint.
	Guidance string

	// Query is the raw GQL match expression as written in the dialect.
	Query string

	expr   query.Expression
	schema *jsonschema.Schema
}

// Label returns the rule's diagnostic label: its name when present, otherwise
// a section-and-position identifier such as nodes.rules[0].
func (r *Rule) Label() string {
	if r.Name != "" {
		return r.Name
	}
	return sectionRulesPath(r.Representation) + "[" + itoa(r.Index) + "]"
}

// SetPolicy controls declaration-level set behaviour. The zero values are not
// the defaults; use DefaultSetPolicy.
type SetPolicy struct {
	// AllowUndefined permits sets present in the graph that have no dialect
	// definition.
	AllowUndefined bool

	// RequireAll requires every dialect-defined set to be present in the graph.
	RequireAll bool

	// AllowEmpty permits a set present in the graph to have no members.
	AllowEmpty bool
}

// DefaultSetPolicy returns the permissive defaults defined by the
// specification: undefined sets allowed, definitions need not be present, and
// empty sets allowed.
func DefaultSetPolicy() SetPolicy {
	return SetPolicy{AllowUndefined: true, RequireAll: false, AllowEmpty: true}
}

// SetDefinition is a named set definition. The name is the GSL set name.
type SetDefinition struct {
	// Name is the GSL set identifier this definition describes.
	Name string

	// Description is a human-readable explanation of the set.
	Description string

	// Guidance is descriptive advice for using the set.
	Guidance string

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

	nodes          []*Rule
	edges          []*Rule
	setPolicy      SetPolicy
	setDefinitions []*SetDefinition
}

// NodeRules returns the node rules in document order. The returned slice is a
// copy; mutating it does not affect the dialect.
func (d *Dialect) NodeRules() []*Rule {
	return append([]*Rule(nil), d.nodes...)
}

// EdgeRules returns the edge rules in document order. The returned slice is a
// copy; mutating it does not affect the dialect.
func (d *Dialect) EdgeRules() []*Rule {
	return append([]*Rule(nil), d.edges...)
}

// SetPolicy returns the effective set policy, including defaults.
func (d *Dialect) SetPolicy() SetPolicy {
	return d.setPolicy
}

// SetDefinitions returns the set definitions in name-sorted order. The returned
// slice is a copy; mutating it does not affect the dialect.
func (d *Dialect) SetDefinitions() []*SetDefinition {
	return append([]*SetDefinition(nil), d.setDefinitions...)
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

// Selected is one graph element selected by one rule or set definition,
// together with the element's attributes projected into JSON-compatible values.
//
// Matching and projection are deterministic and independent of JSON Schema
// validation, so Selected can be inspected on its own (for example by the
// `project` command or the conformance suite).
type Selected struct {
	// Definition is the rule label or set name that selected the element.
	Definition string

	// Representation is the kind of the selected element.
	Representation Representation

	// Element identifies the selected graph element (a node id, a set id, or
	// an edge descriptor such as "a->b").
	Element string

	// Attributes is the projected attribute map that would be validated
	// against the applicable schema.
	Attributes map[string]any

	schema *jsonschema.Schema
}

// Match evaluates every node and edge rule's match query, selects defined sets
// that are present in the graph, and returns the selected elements with their
// projected attributes. The order is deterministic: node rules and then edge
// rules in document order, followed by set definitions in name order; nodes
// and sets are sorted by id and edges retain graph order.
//
// A non-nil error indicates that matching or projection could not be completed
// and is a *MatchingError or *ProjectionError; in that case the result MUST NOT
// be treated as conforming.
func (d *Dialect) Match(g *gsl.Graph) ([]Selected, error) {
	var out []Selected

	appendRules := func(rules []*Rule) error {
		for _, r := range rules {
			selected, err := r.selectElements(g)
			if err != nil {
				return &MatchingError{Definition: r.Label(), Query: r.Query, Err: err}
			}
			for _, elem := range selected {
				projected, err := projectAttributes(elem.attributes)
				if err != nil {
					return &ProjectionError{Definition: r.Label(), Element: elem.ref, Err: err}
				}
				out = append(out, Selected{
					Definition:     r.Label(),
					Representation: r.Representation,
					Element:        elem.ref,
					Attributes:     projected,
					schema:         r.schema,
				})
			}
		}
		return nil
	}

	if err := appendRules(d.nodes); err != nil {
		return nil, err
	}
	if err := appendRules(d.edges); err != nil {
		return nil, err
	}

	sets := g.GetSets()
	for _, def := range d.setDefinitions {
		set, ok := sets[def.Name]
		if !ok {
			continue
		}
		projected, err := projectAttributes(set.Attributes)
		if err != nil {
			return nil, &ProjectionError{Definition: def.Name, Element: def.Name, Err: err}
		}
		out = append(out, Selected{
			Definition:     def.Name,
			Representation: RepresentationSet,
			Element:        def.Name,
			Attributes:     projected,
			schema:         def.schema,
		})
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

	report := &Report{Outcome: OutcomeConforming}
	for _, s := range selected {
		if s.schema == nil {
			continue
		}
		if verr := s.schema.Validate(s.Attributes); verr != nil {
			schemaErr, ok := verr.(*jsonschema.ValidationError)
			if !ok {
				return nil, &ProjectionError{Definition: s.Definition, Element: s.Element, Err: verr}
			}
			report.Violations = append(report.Violations, violationsFromJSON(schemaErr, s)...)
		}
	}

	report.Violations = append(report.Violations, d.validateSetPolicy(g)...)

	if len(report.Violations) > 0 {
		report.Outcome = OutcomeNonConforming
	}
	return report, nil
}

// validateSetPolicy applies the three independent set policies. Policy failures
// are graph conformance violations, not invalid-dialect errors.
func (d *Dialect) validateSetPolicy(g *gsl.Graph) []Violation {
	var out []Violation

	sets := g.GetSets()
	definitions := make(map[string]bool, len(d.setDefinitions))
	for _, def := range d.setDefinitions {
		definitions[def.Name] = true
	}

	members := setMembers(g)

	if !d.setPolicy.AllowUndefined {
		names := make([]string, 0, len(sets))
		for name := range sets {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if !definitions[name] {
				out = append(out, Violation{
					Representation: RepresentationSet,
					Element:        name,
					Message:        "set is not defined by the dialect",
				})
			}
		}
	}

	if d.setPolicy.RequireAll {
		for _, def := range d.setDefinitions {
			if _, ok := sets[def.Name]; !ok {
				out = append(out, Violation{
					Definition:     def.Name,
					Representation: RepresentationSet,
					Element:        def.Name,
					Message:        "required set is not present in the graph",
				})
			}
		}
	}

	if !d.setPolicy.AllowEmpty {
		names := make([]string, 0, len(sets))
		for name := range sets {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if members[name] == 0 {
				def := ""
				if definitions[name] {
					def = name
				}
				out = append(out, Violation{
					Definition:     def,
					Representation: RepresentationSet,
					Element:        name,
					Message:        "set has no members",
				})
			}
		}
	}

	return out
}

func setMembers(g *gsl.Graph) map[string]int {
	counts := make(map[string]int)
	for _, n := range g.GetNodes() {
		for name := range n.Sets {
			counts[name]++
		}
	}
	for _, e := range g.GetEdges() {
		for name := range e.Sets {
			counts[name]++
		}
	}
	return counts
}

func sectionRulesPath(rep Representation) string {
	if rep == RepresentationEdge {
		return "edges.rules"
	}
	return "nodes.rules"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
