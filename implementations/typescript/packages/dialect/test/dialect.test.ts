// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import assert from "node:assert/strict";
import { test } from "node:test";

import { parse } from "@gsl-lang/core";
import { NodeRef } from "@gsl-lang/core";

import {
  LoadError,
  MatchingError,
  load,
  projectAttributes,
  type Dialect,
} from "../src/index.js";

function dialect(yaml: string): Dialect {
  return load(yaml);
}

function graph(source: string) {
  return parse(source).graph;
}

test("internal $ref is allowed", () => {
  const d = dialect(`
id: refs
version: "1"
nodes:
  rules:
    - match: 'kind == "thing"'
      attributes:
        type: object
        required: [owner]
        properties:
          owner:
            $ref: '#/$defs/ref'
        $defs:
          ref:
            type: object
            properties:
              "$gslRef":
                type: string
            required: ["$gslRef"]
`);
  const report = d.validate(graph('node team\nnode a [kind="thing", owner=team]'));
  assert.equal(report.outcome, "conforming", JSON.stringify(report.violations));
});

test("external $ref is rejected at load", () => {
  assert.throws(
    () =>
      dialect(`
id: refs
version: "1"
nodes:
  rules:
    - match: 'kind == "thing"'
      attributes:
        $ref: 'https://example.com/never-loaded.json'
`),
    LoadError,
  );
});

test("unknown root and section fields are ignored", () => {
  const d = dialect(`
id: unknown-fields
version: "1"
futureRootField: ignored
nodes:
  futureSectionField: ignored
  rules:
    - futureRuleField: ignored
      match: 'kind == "service"'
      attributes:
        type: object
sets:
  futurePolicyField: ignored
  definitions:
    critical:
      futureDefinitionField: ignored
`);
  assert.equal(d.validate(graph('node a [kind="service"]')).outcome, "conforming");
  assert.equal(d.setDefinitions.length, 1);
});

test("omitted and empty sections conform", () => {
  for (const source of [
    'id: empty\nversion: "1"',
    'id: empty\nversion: "1"\nnodes: {}\nedges: {}\nsets: {}',
    'id: empty\nversion: "1"\nnodes:\n  rules: []\nedges:\n  rules: []',
  ]) {
    const d = dialect(source);
    assert.equal(d.validate(graph('node a [kind="service"]')).outcome, "conforming");
  }
});

test("edge rules select matching edges", () => {
  const d = dialect(`
id: edges
version: "1"
edges:
  rules:
    - name: HttpEdge
      match: 'protocol == "http"'
      attributes:
        type: object
`);
  const selection = d.match(graph('node a\nnode b\na -> b [protocol="http"]\nb -> a [protocol="tcp"]'));
  assert.deepEqual(
    selection.map((s) => s.element),
    ["a->b"],
  );
  assert.equal(selection[0]!.definition, "HttpEdge");
});

test("unnamed rules use their section position as the label", () => {
  const d = dialect(`
id: labels
version: "1"
nodes:
  rules:
    - name: Named
      match: 'kind exists'
      attributes:
        type: object
    - match: 'kind exists'
      attributes:
        type: object
edges:
  rules:
    - match: 'protocol exists'
      attributes:
        type: object
`);
  assert.equal(d.nodeRules[0]!.label, "Named");
  assert.equal(d.nodeRules[1]!.label, "nodes.rules[1]");
  assert.equal(d.edgeRules[0]!.label, "edges.rules[0]");
});

test("set definitions select only present defined sets", () => {
  const d = dialect(`
id: sets
version: "1"
sets:
  definitions:
    alpha: {}
    missing: {}
`);
  const selected = d.match(graph('set alpha\nset beta\nnode a @alpha'));
  assert.deepEqual(
    selected.map((s) => s.element),
    ["alpha"],
  );
});

test("multiple rules validate independently", () => {
  const d = dialect(`
id: multi
version: "1"
nodes:
  rules:
    - name: Service
      match: 'kind == "service"'
      attributes:
        type: object
        required: [kind]
        properties:
          kind:
            const: service
    - name: Replicas
      match: 'kind == "service"'
      attributes:
        type: object
        required: [replicas]
        properties:
          replicas:
            type: integer
`);
  const report = d.validate(graph('node a [kind="service"]'));
  assert.equal(report.outcome, "non-conforming");
  assert.deepEqual(
    report.violations.map((v) => v.definition),
    ["Replicas"],
  );
});

test("projection maps each documented GSL value kind", () => {
  const got = projectAttributes({
    str: "hello",
    num: 3.5,
    bool: true,
    empty: null,
    ref: new NodeRef("platform_team"),
  });
  assert.deepEqual(got, {
    str: "hello",
    num: 3.5,
    bool: true,
    empty: { $gslEmpty: true },
    ref: { $gslRef: "platform_team" },
  });
});

test("traversal match query is rejected at load", () => {
  assert.throws(
    () =>
      dialect(`
id: traversal
version: "1"
nodes:
  rules:
    - match: 'kind == "service" traverse out 1'
      attributes:
        type: object
`),
    LoadError,
  );
});

test("pipeline match query is rejected at load", () => {
  assert.throws(
    () =>
      dialect(`
id: pipeline
version: "1"
nodes:
  rules:
    - match: 'kind == "service" | remove orphans'
      attributes:
        type: object
`),
    LoadError,
  );
});

test("missing match is rejected at load", () => {
  assert.throws(
    () =>
      dialect(`
id: missing-match
version: "1"
nodes:
  rules:
    - attributes:
        type: object
`),
    LoadError,
  );
});

test("missing attributes is rejected at load", () => {
  assert.throws(
    () =>
      dialect(`
id: missing-attributes
version: "1"
nodes:
  rules:
    - match: 'kind == "service"'
`),
    LoadError,
  );
});

test("non-mapping document is rejected at load", () => {
  assert.throws(() => dialect('- id: nope\n- version: "1"'), LoadError);
});

test("empty document is rejected at load", () => {
  assert.throws(() => dialect(""), LoadError);
});

test("a matching error surfaces as a MatchingError", () => {
  const d = dialect(`
id: matching
version: "1"
nodes:
  rules:
    - match: 'kind == "service"'
      attributes:
        type: object
`);
  const expr = (d.nodeRules[0] as unknown as { expr: { apply: () => never } }).expr;
  expr.apply = () => {
    throw new Error("boom");
  };
  assert.throws(() => d.match(graph('node a [kind="service"]')), MatchingError);
});

test("match query is a bare predicate expanded with the section prefix", () => {
  const d = dialect(`
id: bare
version: "1"
nodes:
  rules:
    - match: 'kind == "service" AND active exists'
      attributes:
        type: object
`);
  assert.equal(d.nodeRules[0]!.query, 'kind == "service" AND active exists');

  const selected = d.match(
    graph('node a [kind="service", active=true]\nnode b [kind="team", active=true]\nnode c [kind="service"]'),
  );
  assert.equal(selected.length, 1);
  assert.equal(selected[0]!.element, "a");
});

test("explicit prefixes and the subgraph wrapper are rejected at load", () => {
  const queries = [
    'node.kind == "service"',
    'edge.protocol == "http"',
    'set.kind == "service"',
    'subgraph node.kind == "service"',
    'subgraph kind == "service"',
  ];
  for (const query of queries) {
    assert.throws(
      () =>
        dialect(`
id: x
version: "1"
nodes:
  rules:
    - match: '${query}'
      attributes:
        type: object
`),
      LoadError,
      `expected ${query} to be rejected`,
    );
  }
});

test("set attributes are validated when defined", () => {
  const d = dialect(`
id: sets
version: "1"
sets:
  definitions:
    critical:
      attributes:
        type: object
        required: [label]
        properties:
          label:
            type: string
`);
  const report = d.validate(graph('set critical\nset ok [label="OK"]'));
  assert.equal(report.outcome, "non-conforming");
  assert.deepEqual(
    report.violations.map((v) => [v.definition, v.element, v.instanceLocation]),
    [["critical", "critical", ""]],
  );
});

test("set definitions without attributes impose no constraint", () => {
  const d = dialect(`
id: sets
version: "1"
sets:
  definitions:
    critical:
      description: Important elements.
`);
  assert.equal(d.validate(graph("set critical\n")).outcome, "conforming");
});

test("default set policy is permissive", () => {
  const d = dialect(`
id: policy
version: "1"
sets:
  definitions:
    unused: {}
`);
  assert.deepEqual(d.setPolicy, {
    allowUndefined: true,
    requireAll: false,
    allowEmpty: true,
  });
  assert.equal(d.validate(graph("set empty\nnode a @implicit\n")).outcome, "conforming");
});

test("undefined sets are rejected when disallowed", () => {
  const d = dialect(`
id: policy
version: "1"
sets:
  policy:
    allow_undefined: false
  definitions:
    alpha: {}
`);
  const report = d.validate(graph("set alpha\nset beta\nnode a @implicit\n"));
  assert.equal(report.outcome, "non-conforming");
  assert.deepEqual(
    report.violations.map((v) => [v.definition, v.element]),
    [
      ["", "beta"],
      ["", "implicit"],
    ],
  );
});

test("required sets must be present", () => {
  const d = dialect(`
id: policy
version: "1"
sets:
  policy:
    require_all: true
  definitions:
    alpha: {}
    beta: {}
`);
  const report = d.validate(graph("set alpha\n"));
  assert.equal(report.outcome, "non-conforming");
  assert.deepEqual(
    report.violations.map((v) => [v.definition, v.element]),
    [["beta", "beta"]],
  );
});

test("empty sets are rejected when disallowed", () => {
  const d = dialect(`
id: policy
version: "1"
sets:
  policy:
    allow_empty: false
  definitions:
    alpha: {}
`);
  const report = d.validate(graph("set alpha\n"));
  assert.equal(report.outcome, "non-conforming");
  assert.deepEqual(
    report.violations.map((v) => [v.definition, v.element]),
    [["alpha", "alpha"]],
  );
});

test("implicit sets count as present and non-empty", () => {
  const d = dialect(`
id: policy
version: "1"
sets:
  policy:
    require_all: true
    allow_empty: false
  definitions:
    implicit: {}
`);
  assert.equal(d.validate(graph("node a @implicit\n")).outcome, "conforming");
});

test("cross-schema internal $ref is rejected", () => {
  assert.throws(
    () =>
      dialect(`
id: refs
version: "1"
nodes:
  rules:
    - match: 'kind == "a"'
      attributes:
        type: object
        properties:
          owner:
            $ref: '#/$defs/ref'
    - match: 'kind == "b"'
      attributes:
        type: object
        $defs:
          ref:
            type: string
`),
    LoadError,
  );
});
