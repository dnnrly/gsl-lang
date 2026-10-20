// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	gsl "github.com/dnnrly/gsl-lang/implementations/go"
	"github.com/dnnrly/gsl-lang/implementations/go/dialect"
)

// Exit codes for the gsl-dialect CLI.
const (
	// ExitConforming is returned when the graph conforms to the dialect.
	ExitConforming = 0
	// ExitNonConforming is returned when validation completed and found one or
	// more attribute constraint violations.
	ExitNonConforming = 1
	// ExitIncomplete is returned when validation could not be completed, for
	// example because the dialect is invalid or a match query failed.
	ExitIncomplete = 2
)

// exitError carries a requested process exit code up to main.
type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string {
	if e.err == nil {
		return fmt.Sprintf("exit %d", e.code)
	}
	return e.err.Error()
}

func (e *exitError) Unwrap() error { return e.err }

type validateConfig struct {
	DialectFile string
	InputFile   string
	JSON        bool
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
}

type projectConfig struct {
	DialectFile string
	InputFile   string
	Stdin       io.Reader
	Stdout      io.Writer
}

func runValidate(cfg *validateConfig) error {
	d, err := dialect.LoadFile(cfg.DialectFile)
	if err != nil {
		return &exitError{code: ExitIncomplete, err: err}
	}

	graph, name, err := readGraph(cfg.InputFile, cfg.Stdin)
	if err != nil {
		return &exitError{code: ExitIncomplete, err: err}
	}

	report, err := d.Validate(graph)
	if err != nil {
		return &exitError{code: ExitIncomplete, err: err}
	}

	if cfg.JSON {
		if err := writeReportJSON(cfg.Stdout, d, graph, name, report); err != nil {
			return &exitError{code: ExitIncomplete, err: err}
		}
	} else {
		writeReportText(cfg.Stdout, d, name, report)
	}

	if !report.Conforms() {
		return &exitError{code: ExitNonConforming}
	}
	return nil
}

func runProject(cfg *projectConfig) error {
	d, err := dialect.LoadFile(cfg.DialectFile)
	if err != nil {
		return &exitError{code: ExitIncomplete, err: err}
	}

	graph, _, err := readGraph(cfg.InputFile, cfg.Stdin)
	if err != nil {
		return &exitError{code: ExitIncomplete, err: err}
	}

	selected, err := d.Match(graph)
	if err != nil {
		return &exitError{code: ExitIncomplete, err: err}
	}

	// Every selected element is reported, even when two distinct elements
	// share a descriptor (for example duplicate unlabelled edges between the
	// same endpoints). The value for each descriptor is therefore always a
	// list of projected attribute maps, so no selection is silently dropped
	// and the shape does not vary with the number of matches.
	projected := map[string]map[string][]map[string]any{}
	for _, s := range selected {
		if projected[s.Definition] == nil {
			projected[s.Definition] = map[string][]map[string]any{}
		}
		projected[s.Definition][s.Element] = append(projected[s.Definition][s.Element], s.Attributes)
	}

	out, err := json.MarshalIndent(projected, "", "  ")
	if err != nil {
		return &exitError{code: ExitIncomplete, err: err}
	}
	fmt.Fprintln(cfg.Stdout, string(out))
	return nil
}

func readGraph(inputFile string, stdin io.Reader) (*gsl.Graph, string, error) {
	name := inputFile
	var data []byte
	var err error

	if inputFile != "" {
		data, err = os.ReadFile(inputFile)
		if err != nil {
			return nil, name, fmt.Errorf("failed to read input file: %w", err)
		}
	} else {
		name = "<stdin>"
		if stdin == nil {
			stdin = os.Stdin
		}
		data, err = io.ReadAll(stdin)
		if err != nil {
			return nil, name, fmt.Errorf("failed to read from stdin: %w", err)
		}
	}

	graph, perr := gsl.Parse(bytes.NewReader(data))
	if perr != nil && perr.HasError() {
		return nil, name, fmt.Errorf("failed to parse %s: %w", name, perr)
	}
	return graph, name, nil
}

func writeReportText(w io.Writer, d *dialect.Dialect, name string, report *dialect.Report) {
	if report.Conforms() {
		fmt.Fprintf(w, "Conforming: %s (%s)\n", dialectLabel(d), name)
		return
	}

	fmt.Fprintf(w, "Non-conforming: %d violation(s) in %s\n", len(report.Violations), name)
	for _, v := range sortedViolations(report.Violations) {
		loc := v.InstanceLocation
		if loc == "" {
			loc = "/"
		}
		fmt.Fprintf(w, "  %s %s %s: %s\n", v.Representation, v.Element, loc, v.Message)
	}
}

// reportJSON is the provisional machine-readable validation report. Its shape
// is not yet standardised (spec §16.4).
type reportJSON struct {
	Dialect struct {
		ID      string `json:"id"`
		Version string `json:"version"`
	} `json:"dialect"`
	Graph      string          `json:"graph"`
	Outcome    string          `json:"outcome"`
	Violations []violationJSON `json:"violations"`
}

type violationJSON struct {
	Definition     string `json:"definition"`
	Representation string `json:"representation"`
	Element        string `json:"element"`
	Location       string `json:"location"`
	Message        string `json:"message"`
}

func writeReportJSON(w io.Writer, d *dialect.Dialect, graph *gsl.Graph, name string, report *dialect.Report) error {
	var out reportJSON
	out.Dialect.ID = d.ID
	out.Dialect.Version = d.Version
	out.Graph = name
	out.Outcome = report.Outcome.String()
	out.Violations = []violationJSON{}
	for _, v := range sortedViolations(report.Violations) {
		out.Violations = append(out.Violations, violationJSON{
			Definition:     v.Definition,
			Representation: string(v.Representation),
			Element:        v.Element,
			Location:       v.InstanceLocation,
			Message:        v.Message,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func dialectLabel(d *dialect.Dialect) string {
	label := fmt.Sprintf("dialect %q", d.ID)
	if d.Version != "" {
		label += fmt.Sprintf(" version %q", d.Version)
	}
	return label
}

func sortedViolations(violations []dialect.Violation) []dialect.Violation {
	out := make([]dialect.Violation, len(violations))
	copy(out, violations)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Definition != out[j].Definition {
			return out[i].Definition < out[j].Definition
		}
		if out[i].Element != out[j].Element {
			return out[i].Element < out[j].Element
		}
		return out[i].InstanceLocation < out[j].InstanceLocation
	})
	return out
}
