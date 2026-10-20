// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

package dialect

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"

	"github.com/dnnrly/gsl-lang/implementations/go/query"
)

// rawDialect mirrors the supported root fields of a dialect document. Unknown
// fields are ignored (and therefore tolerated) by the YAML decoder.
type rawDialect struct {
	ID          yaml.Node `yaml:"id"`
	Version     yaml.Node `yaml:"version"`
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Guidance    string    `yaml:"guidance"`
	Elements    yaml.Node `yaml:"elements"`
}

// rawElement mirrors the supported fields of an element definition.
type rawElement struct {
	Representation *string   `yaml:"representation"`
	Match          yaml.Node `yaml:"match"`
	Attributes     yaml.Node `yaml:"attributes"`
	Description    string    `yaml:"description"`
	Guidance       string    `yaml:"guidance"`
}

// rawMatch mirrors the `match` mapping.
type rawMatch struct {
	Query *string `yaml:"query"`
}

func load(r io.Reader) (*Dialect, error) {
	dec := yaml.NewDecoder(r)

	var raw rawDialect
	if err := dec.Decode(&raw); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, loadError("", "empty document")
		}
		return nil, loadError("", "not a valid YAML mapping: %v", err)
	}

	// A dialect document is a single YAML document. Reject additional
	// documents rather than silently ignoring them.
	var extra yaml.Node
	if err := dec.Decode(&extra); err == nil {
		return nil, loadError("", "must be a single YAML document")
	} else if !errors.Is(err, io.EOF) {
		return nil, loadError("", "not a valid YAML mapping: %v", err)
	}

	id, err := requiredString(&raw.ID, "id")
	if err != nil {
		return nil, err
	}
	version, err := requiredString(&raw.Version, "version")
	if err != nil {
		return nil, err
	}
	if raw.Elements.Kind == 0 {
		return nil, loadError("elements", "required mapping (may be empty)")
	}
	if raw.Elements.Kind != yaml.MappingNode {
		return nil, loadError("elements", "must be a mapping")
	}

	d := &Dialect{
		ID:          id,
		Version:     version,
		Title:       raw.Title,
		Description: raw.Description,
		Guidance:    raw.Guidance,
	}

	content := raw.Elements.Content
	for i := 0; i+1 < len(content); i += 2 {
		name := content[i].Value
		el, err := parseElement(name, content[i+1])
		if err != nil {
			return nil, err
		}
		d.elements = append(d.elements, el)
	}

	sort.Slice(d.elements, func(i, j int) bool {
		return d.elements[i].Name < d.elements[j].Name
	})

	return d, nil
}

// requiredString validates a required scalar string field such as `id` or
// `version`.
//
// The YAML tag is inspected rather than the value: a plain scalar like
// `version: 1` resolves to !!int and is rejected instead of being silently
// coerced to "1". Spec §5.1/§5.2 require both fields to be strings, and the
// TypeScript loader applies the same rule, so the two implementations agree.
func requiredString(node *yaml.Node, path string) (string, error) {
	if node.Kind == 0 {
		return "", loadError(path, "required non-empty string")
	}
	if node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
		return "", loadError(path, "must be a non-empty string (got %s)", nodeTag(node))
	}
	if strings.TrimSpace(node.Value) == "" {
		return "", loadError(path, "required non-empty string")
	}
	return node.Value, nil
}

// nodeTag reports a parsed node's YAML tag, falling back to its kind so that a
// diagnostic never names an empty tag.
func nodeTag(node *yaml.Node) string {
	if node.Tag != "" {
		return node.Tag
	}
	switch node.Kind {
	case yaml.MappingNode:
		return "!!map"
	case yaml.SequenceNode:
		return "!!seq"
	}
	return "!!null"
}

func parseElement(name string, node *yaml.Node) (*ElementDefinition, error) {
	path := "elements[" + name + "]"

	if node.Kind != yaml.MappingNode {
		return nil, loadError(path, "must be a mapping")
	}

	var raw rawElement
	if err := node.Decode(&raw); err != nil {
		return nil, loadError(path, "invalid element definition: %v", err)
	}

	if raw.Representation == nil {
		return nil, loadError(path+".representation", "required")
	}
	rep := Representation(*raw.Representation)
	if !rep.valid() {
		return nil, loadError(path+".representation",
			"must be one of node, edge, set (got %q)", *raw.Representation)
	}

	expr, matchQuery, err := parseMatch(path, &raw.Match)
	if err != nil {
		return nil, err
	}

	schema, err := compileSchema(name, path, &raw.Attributes)
	if err != nil {
		return nil, err
	}

	return &ElementDefinition{
		Name:           name,
		Representation: rep,
		Description:    raw.Description,
		Guidance:       raw.Guidance,
		Query:          matchQuery,
		expr:           expr,
		schema:         schema,
	}, nil
}

func parseMatch(elementPath string, node *yaml.Node) (query.Expression, string, error) {
	path := elementPath + ".match"

	if node.Kind == 0 {
		return nil, "", loadError(path, "required mapping with a query field")
	}
	if node.Kind != yaml.MappingNode {
		return nil, "", loadError(path, "must be a mapping")
	}

	var raw rawMatch
	if err := node.Decode(&raw); err != nil {
		return nil, "", loadError(path, "invalid match: %v", err)
	}
	if raw.Query == nil || strings.TrimSpace(*raw.Query) == "" {
		return nil, "", loadError(path+".query", "required non-empty string")
	}

	expr, err := parseMatchQuery(*raw.Query)
	if err != nil {
		return nil, "", loadError(path+".query", "%v", err)
	}
	return expr, *raw.Query, nil
}

// parseMatchQuery parses a match query and enforces the supported GQL subset:
// a single `subgraph` predicate with no traversal or scope and no pipeline,
// transformation or graph-algebra stage.
func parseMatchQuery(q string) (query.Expression, error) {
	parsed, err := query.NewQueryParser(q).Parse()
	if err != nil {
		return nil, fmt.Errorf("invalid match query: %v", err)
	}
	if len(parsed.Expressions) != 1 {
		return nil, errors.New("unsupported match query: only a single subgraph expression is supported")
	}
	sub, ok := parsed.Expressions[0].(*query.SubgraphExpr)
	if !ok {
		return nil, errors.New("unsupported match query: only the subgraph form is supported")
	}
	if sub.Traversal != nil {
		return nil, errors.New("unsupported match query: traversal and scope are not supported")
	}
	return sub, nil
}

func compileSchema(name, elementPath string, node *yaml.Node) (*jsonschema.Schema, error) {
	path := elementPath + ".attributes"

	if node.Kind == 0 {
		return nil, loadError(path, "required JSON Schema")
	}

	var doc any
	if err := node.Decode(&doc); err != nil {
		return nil, loadError(path, "invalid JSON Schema: %v", err)
	}

	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2020)
	// No external resource loading: a dialect's attribute schema may reference
	// its own $defs, but must not escape to the network or the filesystem.
	c.UseLoader(jsonschema.SchemeURLLoader{})

	loc := "dialect://" + url.PathEscape(name) + "/attributes.json"
	if err := c.AddResource(loc, doc); err != nil {
		return nil, loadError(path, "invalid JSON Schema: %v", err)
	}
	schema, err := c.Compile(loc)
	if err != nil {
		return nil, loadError(path, "invalid JSON Schema: %v", err)
	}
	return schema, nil
}
