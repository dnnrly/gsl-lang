// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

package dialect

import (
	"fmt"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

// Outcome is the result of dialect conformance validation.
type Outcome int

const (
	// OutcomeConforming means the dialect is valid, all matching and
	// validation completed, and no applicable constraint violations were found.
	OutcomeConforming Outcome = iota

	// OutcomeNonConforming means the dialect is valid, validation completed,
	// and one or more matched elements violated their constraints.
	OutcomeNonConforming

	// OutcomeIncomplete means matching or validation failed in a way that
	// prevented the required checks from completing. It is never reported as
	// conformance.
	OutcomeIncomplete
)

func (o Outcome) String() string {
	switch o {
	case OutcomeConforming:
		return "conforming"
	case OutcomeNonConforming:
		return "non-conforming"
	case OutcomeIncomplete:
		return "incomplete"
	default:
		return fmt.Sprintf("unknown(%d)", int(o))
	}
}

// Report is the result of validating a graph against a dialect.
type Report struct {
	// Outcome is the overall conformance outcome.
	Outcome Outcome

	// Violations lists the attribute constraint violations found. It is
	// populated only when Outcome is OutcomeNonConforming.
	Violations []Violation
}

// Conforms reports whether the graph conforms to the dialect.
func (r *Report) Conforms() bool {
	return r.Outcome == OutcomeConforming
}

// Violation describes a single conformance violation: either an attribute
// constraint failure or a set policy failure.
type Violation struct {
	// Definition is the rule label or set name that selected the element. It is
	// empty for a policy violation concerning a set that has no definition.
	Definition string

	// Representation is the kind of the offending element.
	Representation Representation

	// Element identifies the offending graph element (a node id, a set id, or
	// an edge descriptor such as "a->b").
	Element string

	// InstanceLocation is the JSON Pointer path within the projected attribute
	// map (for example "/replicas"). It is empty for a violation at the root.
	InstanceLocation string

	// Message is a human-readable explanation of the violation.
	Message string
}

func (v Violation) String() string {
	loc := v.InstanceLocation
	if loc == "" {
		loc = "/"
	}
	return fmt.Sprintf("%s %s %s: %s", v.Representation, v.Element, loc, v.Message)
}

// LoadError describes a dialect document that is invalid and was rejected
// before graph validation began.
type LoadError struct {
	// Path locates the offending field, for example "nodes.rules[0].match" or
	// "sets.definitions[critical].attributes".
	Path string

	// Message explains the problem.
	Message string
}

func (e *LoadError) Error() string {
	if e.Path == "" {
		return "invalid dialect: " + e.Message
	}
	return "invalid dialect: " + e.Path + ": " + e.Message
}

func loadError(path, format string, args ...any) *LoadError {
	return &LoadError{Path: path, Message: fmt.Sprintf(format, args...)}
}

// MatchingError describes a failure to evaluate a match query. It indicates
// incomplete validation rather than a conformance violation.
type MatchingError struct {
	Definition string
	Query      string
	Err        error
}

func (e *MatchingError) Error() string {
	return fmt.Sprintf("matching error in definition %q (query %q): %v", e.Definition, e.Query, e.Err)
}

func (e *MatchingError) Unwrap() error { return e.Err }

// ProjectionError describes a failure to project an element's attribute values
// into JSON. It indicates incomplete validation.
type ProjectionError struct {
	Definition string
	Element    string
	Err        error
}

func (e *ProjectionError) Error() string {
	return fmt.Sprintf("projection error for %s in definition %q: %v", e.Element, e.Definition, e.Err)
}

func (e *ProjectionError) Unwrap() error { return e.Err }

// violationsFromJSON flattens a JSON Schema validation error into dialect
// violations, one per leaf cause.
func violationsFromJSON(verr *jsonschema.ValidationError, selected Selected) []Violation {
	var out []Violation
	var collect func(e *jsonschema.ValidationError, location []string)
	collect = func(e *jsonschema.ValidationError, location []string) {
		loc := e.InstanceLocation
		if len(loc) == 0 {
			loc = location
		}
		if len(e.Causes) == 0 {
			out = append(out, Violation{
				Definition:       selected.Definition,
				Representation:   selected.Representation,
				Element:          selected.Element,
				InstanceLocation: jsonPointer(loc),
				Message:          e.Error(),
			})
			return
		}
		for _, cause := range e.Causes {
			collect(cause, loc)
		}
	}
	collect(verr, nil)
	return out
}

func jsonPointer(tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}
	out := ""
	for _, t := range tokens {
		out += "/" + escapeJSONPointer(t)
	}
	return out
}

func escapeJSONPointer(token string) string {
	var b []byte
	for i := 0; i < len(token); i++ {
		switch token[i] {
		case '~':
			b = append(b, '~', '0')
		case '/':
			b = append(b, '~', '1')
		default:
			b = append(b, token[i])
		}
	}
	return string(b)
}
