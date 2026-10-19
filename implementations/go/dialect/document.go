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
	ID          *string   `yaml:"id"`
	Version     *string   `yaml:"version"`
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

	if raw.ID == nil || strings.TrimSpace(*raw.ID) == "" {
		return nil, loadError("id", "required non-empty string")
	}
	if raw.Version == nil || strings.TrimSpace(*raw.Version) == "" {
		return nil, loadError("version", "required non-empty string")
	}
	if raw.Elements.Kind == 0 {
		return nil, loadError("elements", "required mapping (may be empty)")
	}
	if raw.Elements.Kind != yaml.MappingNode {
		return nil, loadError("elements", "must be a mapping")
	}

	d := &Dialect{
		ID:          *raw.ID,
		Version:     *raw.Version,
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
