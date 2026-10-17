// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import assert from "node:assert/strict";
import { test } from "node:test";

import { parse, serialize, ParseError } from "../src/index.js";

test("grouped edges on both sides is an error", () => {
  assert.throws(() => parse("A,B -> C,D"), ParseError);
});

test("duplicate edge label is an error", () => {
  assert.throws(() => parse("E1: A->B\nE1: C->D"), ParseError);
});

test("node attributes merge last-write-wins", () => {
  const { graph } = parse("node A [x=1]\nnode A [x=2]");
  const a = graph.nodes.find((n) => n.id === "A");
  assert.equal(a?.attributes["x"], 2);
});

test("edge order and duplicates are preserved", () => {
  const { graph } = parse("A->B\nA->B\nB->A");
  assert.equal(graph.edges.length, 3);
  assert.deepEqual(
    graph.edges.map((e) => `${e.from}->${e.to}`),
    ["A->B", "A->B", "B->A"],
  );
});

test("implicit set membership warns", () => {
  const { warnings } = parse("node A @critical");
  assert.ok(warnings && warnings.length > 0);
});

test("explicit set declaration does not warn", () => {
  const { warnings } = parse("set critical\n\nnode A @critical");
  assert.equal(warnings, undefined);
});

test("node block sets parent relationship", () => {
  const { graph } = parse("node A {\n    node B\n}");
  const b = graph.nodes.find((n) => n.id === "B");
  assert.equal(b?.parent, "A");
});

test("scoped edge block sets implicit parent", () => {
  const { graph } = parse("E1: a->b {\n    b->c\n}");
  const child = graph.edges.find((e) => e.from === "b" && e.to === "c");
  assert.equal(child?.parent, "E1");
});

test("unknown parent label is an error", () => {
  assert.throws(() => parse("A->B [parent=NOPE]"), ParseError);
});

test("explicit parent inside scoped block is an error", () => {
  assert.throws(() => parse("E1: a->b {\n    b->c [parent=E1]\n}"), ParseError);
});

test("parse error carries diagnostics", () => {
  try {
    parse("A,B -> C,D");
    assert.fail("expected ParseError");
  } catch (err) {
    assert.ok(err instanceof ParseError);
    assert.ok((err as ParseError).errors.length > 0);
  }
});

test("empty input produces an empty graph", () => {
  const { graph } = parse("");
  assert.equal(graph.nodes.length, 0);
  assert.equal(graph.edges.length, 0);
  assert.equal(serialize(graph), "");
});
