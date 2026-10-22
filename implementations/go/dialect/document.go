// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

package dialect

import (
	"errors"
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
	Nodes       yaml.Node `yaml:"nodes"`
	Edges       yaml.Node `yaml:"edges"`
	Sets        yaml.Node `yaml:"sets"`
}

// rawSection mirrors the supported fields of a nodes or edges section.
type rawSection struct {
	Rules yaml.Node `yaml:"rules"`
}

// rawRule mirrors the supported fields of a node or edge rule.
type rawRule struct {
	Name        yaml.Node `yaml:"name"`
	Match       yaml.Node `yaml:"match"`
	Attributes  yaml.Node `yaml:"attributes"`
	Description string    `yaml:"description"`
	Guidance    string    `yaml:"guidance"`
}

// rawSets mirrors the supported fields of the sets section.
type rawSets struct {
	Policy      yaml.Node `yaml:"policy"`
	Definitions yaml.Node `yaml:"definitions"`
}

// rawSetPolicy mirrors the supported policy keys. yaml.Node preserves the
// written YAML kind so non-boolean values are rejected rather than coerced.
type rawSetPolicy struct {
	AllowUndefined yaml.Node `yaml:"allow_undefined"`
	RequireAll     yaml.Node `yaml:"require_all"`
	AllowEmpty     yaml.Node `yaml:"allow_empty"`
}

// rawSetDefinition mirrors the supported fields of a set definition.
type rawSetDefinition struct {
	Description string    `yaml:"description"`
	Guidance    string    `yaml:"guidance"`
	Attributes  yaml.Node `yaml:"attributes"`
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

	d := &Dialect{
		ID:          id,
		Version:     version,
		Title:       raw.Title,
		Description: raw.Description,
		Guidance:    raw.Guidance,
		setPolicy:   DefaultSetPolicy(),
	}

	d.nodes, err = parseRules("nodes", RepresentationNode, &raw.Nodes)
	if err != nil {
		return nil, err
	}
	d.edges, err = parseRules("edges", RepresentationEdge, &raw.Edges)
	if err != nil {
		return nil, err
	}
	if err := parseSets(d, &raw.Sets); err != nil {
		return nil, err
	}

	return d, nil
}

// requiredString validates a required scalar string field such as `id`,
// `version`, `match` or a rule name.
//
// The YAML tag is inspected rather than the value: a plain scalar like
// `version: 1` resolves to !!int and is rejected instead of being silently
// coerced to "1".
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

func parseRules(section string, rep Representation, node *yaml.Node) ([]*Rule, error) {
	if node.Kind == 0 {
		return nil, nil
	}
	if node.Kind != yaml.MappingNode {
		return nil, loadError(section, "must be a mapping")
	}

	var raw rawSection
	if err := node.Decode(&raw); err != nil {
		return nil, loadError(section, "invalid section: %v", err)
	}
	if raw.Rules.Kind == 0 {
		return nil, nil
	}
	if raw.Rules.Kind != yaml.SequenceNode {
		return nil, loadError(section+".rules", "must be a sequence")
	}

	out := make([]*Rule, 0, len(raw.Rules.Content))
	for i, item := range raw.Rules.Content {
		rule, err := parseRule(section, rep, i, item)
		if err != nil {
			return nil, err
		}
		out = append(out, rule)
	}
	return out, nil
}

func parseRule(section string, rep Representation, index int, node *yaml.Node) (*Rule, error) {
	path := section + ".rules[" + itoa(index) + "]"

	if node.Kind != yaml.MappingNode {
		return nil, loadError(path, "must be a mapping")
	}

	var raw rawRule
	if err := node.Decode(&raw); err != nil {
		return nil, loadError(path, "invalid rule: %v", err)
	}

	name := ""
	if raw.Name.Kind != 0 {
		var err error
		name, err = requiredString(&raw.Name, path+".name")
		if err != nil {
			return nil, err
		}
	}

	q, err := requiredString(&raw.Match, path+".match")
	if err != nil {
		return nil, err
	}

	expr, err := parseMatchQuery(path+".match", string(rep), q)
	if err != nil {
		return nil, err
	}

	schema, err := compileSchema(path+".attributes", section+"-"+itoa(index), &raw.Attributes)
	if err != nil {
		return nil, err
	}

	return &Rule{
		Name:           name,
		Representation: rep,
		Index:          index,
		Description:    raw.Description,
		Guidance:       raw.Guidance,
		Query:          q,
		expr:           expr,
		schema:         schema,
	}, nil
}

// parseMatchQuery expands a bare attribute predicate into a full subgraph
// query using the enclosing section's element prefix, then enforces the
// supported GQL subset: a single `subgraph` predicate with no traversal or
// scope and no pipeline, transformation or graph-algebra stage. The written
// form must not itself carry an element prefix or the `subgraph` keyword.
func parseMatchQuery(path, rep, q string) (query.Expression, error) {
	expanded, err := query.ExpandMatchQuery(rep, q)
	if err != nil {
		return nil, loadError(path, "%v", err)
	}

	parsed, err := query.NewQueryParser(expanded).Parse()
	if err != nil {
		return nil, loadError(path, "invalid match query: %v", err)
	}
	if len(parsed.Expressions) != 1 {
		return nil, loadError(path, "unsupported match query: only a single subgraph expression is supported")
	}
	sub, ok := parsed.Expressions[0].(*query.SubgraphExpr)
	if !ok {
		return nil, loadError(path, "unsupported match query: only the subgraph form is supported")
	}
	if sub.Traversal != nil {
		return nil, loadError(path, "unsupported match query: traversal and scope are not supported")
	}
	return sub, nil
}

func parseSets(d *Dialect, node *yaml.Node) error {
	if node.Kind == 0 {
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return loadError("sets", "must be a mapping")
	}

	var raw rawSets
	if err := node.Decode(&raw); err != nil {
		return loadError("sets", "invalid section: %v", err)
	}

	policy, err := parseSetPolicy(&raw.Policy)
	if err != nil {
		return err
	}
	d.setPolicy = policy

	definitions, err := parseSetDefinitions(&raw.Definitions)
	if err != nil {
		return err
	}
	d.setDefinitions = definitions
	return nil
}

func parseSetPolicy(node *yaml.Node) (SetPolicy, error) {
	policy := DefaultSetPolicy()
	if node.Kind == 0 {
		return policy, nil
	}
	if node.Kind != yaml.MappingNode {
		return policy, loadError("sets.policy", "must be a mapping")
	}

	var raw rawSetPolicy
	if err := node.Decode(&raw); err != nil {
		return policy, loadError("sets.policy", "invalid policy: %v", err)
	}

	var err error
	policy.AllowUndefined, err = optionalBool(&raw.AllowUndefined, "sets.policy.allow_undefined", policy.AllowUndefined)
	if err != nil {
		return policy, err
	}
	policy.RequireAll, err = optionalBool(&raw.RequireAll, "sets.policy.require_all", policy.RequireAll)
	if err != nil {
		return policy, err
	}
	policy.AllowEmpty, err = optionalBool(&raw.AllowEmpty, "sets.policy.allow_empty", policy.AllowEmpty)
	if err != nil {
		return policy, err
	}
	return policy, nil
}

func optionalBool(node *yaml.Node, path string, def bool) (bool, error) {
	if node.Kind == 0 {
		return def, nil
	}
	if node.Kind != yaml.ScalarNode || node.Tag != "!!bool" {
		return def, loadError(path, "must be a boolean (got %s)", nodeTag(node))
	}
	var value bool
	if err := node.Decode(&value); err != nil {
		return def, loadError(path, "must be a boolean: %v", err)
	}
	return value, nil
}

func parseSetDefinitions(node *yaml.Node) ([]*SetDefinition, error) {
	if node.Kind == 0 {
		return nil, nil
	}
	if node.Kind != yaml.MappingNode {
		return nil, loadError("sets.definitions", "must be a mapping")
	}

	content := node.Content
	out := make([]*SetDefinition, 0, len(content)/2)
	seen := make(map[string]bool, len(content)/2)
	for i := 0; i+1 < len(content); i += 2 {
		key := content[i]
		if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || !validSetName(key.Value) {
			return nil, loadError("sets.definitions", "definition key %q must be a GSL set identifier", key.Value)
		}
		if seen[key.Value] {
			return nil, loadError("sets.definitions["+key.Value+"]", "duplicate set definition")
		}
		seen[key.Value] = true

		def, err := parseSetDefinition(key.Value, content[i+1])
		if err != nil {
			return nil, err
		}
		out = append(out, def)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func parseSetDefinition(name string, node *yaml.Node) (*SetDefinition, error) {
	path := "sets.definitions[" + name + "]"

	if node.Kind != yaml.MappingNode {
		return nil, loadError(path, "must be a mapping")
	}

	var raw rawSetDefinition
	if err := node.Decode(&raw); err != nil {
		return nil, loadError(path, "invalid set definition: %v", err)
	}

	def := &SetDefinition{
		Name:        name,
		Description: raw.Description,
		Guidance:    raw.Guidance,
	}

	if raw.Attributes.Kind != 0 {
		schema, err := compileSchema(path+".attributes", "sets-"+name, &raw.Attributes)
		if err != nil {
			return nil, err
		}
		def.schema = schema
	}

	return def, nil
}

func validSetName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			continue
		}
		if i > 0 && r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return true
}

func compileSchema(path, locID string, node *yaml.Node) (*jsonschema.Schema, error) {
	if node.Kind == 0 {
		return nil, loadError(path, "required JSON Schema")
	}

	var doc any
	if err := node.Decode(&doc); err != nil {
		return nil, loadError(path, "invalid JSON Schema: %v", err)
	}

	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2020)
	// No external resource loading: an attributes schema may reference its own
	// $defs, but must not escape to the network or the filesystem.
	c.UseLoader(jsonschema.SchemeURLLoader{})

	loc := "dialect://" + url.PathEscape(locID) + "/attributes.json"
	if err := c.AddResource(loc, doc); err != nil {
		return nil, loadError(path, "invalid JSON Schema: %v", err)
	}
	schema, err := c.Compile(loc)
	if err != nil {
		return nil, loadError(path, "invalid JSON Schema: %v", err)
	}
	return schema, nil
}
