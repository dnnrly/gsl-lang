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
elements:
  Thing:
    representation: node
    match:
      query: 'kind == "thing"'
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
elements:
  Thing:
    representation: node
    match:
      query: 'kind == "thing"'
    attributes:
      $ref: 'https://example.com/never-loaded.json'
`),
    LoadError,
  );
});

test("unknown root and element fields are ignored", () => {
  const d = dialect(`
id: unknown-fields
version: "1"
futureRootField: ignored
elements:
  Service:
    representation: node
    futureElementField: ignored
    match:
      query: 'kind == "service"'
    attributes:
      type: object
`);
  assert.equal(d.validate(graph('node a [kind="service"]')).outcome, "conforming");
});

test("empty elements mapping conforms", () => {
  const d = dialect('id: empty\nversion: "1"\nelements: {}');
  assert.equal(d.validate(graph('node a [kind="service"]')).outcome, "conforming");
});

test("edge representation selects matching edges", () => {
  const d = dialect(`
id: edges
version: "1"
elements:
  HttpEdge:
    representation: edge
    match:
      query: 'protocol == "http"'
    attributes:
      type: object
`);
  const selection = d.match(graph('node a\nnode b\na -> b [protocol="http"]'));
  assert.deepEqual(
    selection.map((s) => s.element),
    ["a->b"],
  );
});

test("set representation selects all sets", () => {
  const d = dialect(`
id: sets
version: "1"
elements:
  Classification:
    representation: set
    match:
      query: 'kind == "service"'
    attributes:
      type: object
`);
  const selection = d.match(graph('set alpha\nset beta\nnode a [kind="service"] @alpha'));
  assert.deepEqual(
    selection.map((s) => s.element),
    ["alpha", "beta"],
  );
});

test("multiple definitions validate independently", () => {
  const d = dialect(`
id: multi
version: "1"
elements:
  Service:
    representation: node
    match:
      query: 'kind == "service"'
    attributes:
      type: object
      required: [kind]
      properties:
        kind:
          const: service
  Replicas:
    representation: node
    match:
      query: 'kind == "service"'
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
elements:
  Service:
    representation: node
    match:
      query: 'kind == "service" traverse out 1'
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
elements:
  Service:
    representation: node
    match:
      query: 'kind == "service" | remove orphans'
    attributes:
      type: object
`),
    LoadError,
  );
});

test("missing representation is rejected at load", () => {
  assert.throws(
    () =>
      dialect(`
id: missing-representation
version: "1"
elements:
  Service:
    match:
      query: 'kind == "service"'
    attributes:
      type: object
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
elements:
  Service:
    representation: node
    match:
      query: 'kind == "service"'
    attributes:
      type: object
`);
  // Corrupt the compiled expression to force an evaluation failure.
  const expr = (d.elements[0] as unknown as { expr: { apply: () => never } }).expr;
  expr.apply = () => {
    throw new Error("boom");
  };
  assert.throws(() => d.match(graph('node a [kind="service"]')), MatchingError);
});

test("match query is a bare predicate expanded with the representation", () => {
  const d = dialect(`
id: bare
version: "1"
elements:
  Service:
    representation: node
    match:
      query: 'kind == "service" AND active exists'
    attributes:
      type: object
`);
  assert.equal(d.elements[0]!.query, 'kind == "service" AND active exists');

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
elements:
  T:
    representation: node
    match:
      query: '${query}'
    attributes:
      type: object
`),
      LoadError,
      `expected ${query} to be rejected`,
    );
  }
});

test("a set match query is validated but not evaluated", () => {
  const d = dialect(`
id: sets
version: "1"
elements:
  Group:
    representation: set
    match:
      query: 'kind == "service"'
    attributes:
      type: object
`);
  const selected = d.match(graph('set alpha\nset beta\nnode a [kind="service"] @alpha'));
  assert.deepEqual(
    selected.map((s) => s.element),
    ["alpha", "beta"],
  );

  assert.throws(
    () =>
      dialect(`
id: x
version: "1"
elements:
  G:
    representation: set
    match:
      query: 'kind =='
    attributes:
      type: object
`),
    LoadError,
  );
});
