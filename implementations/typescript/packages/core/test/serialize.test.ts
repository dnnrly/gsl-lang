// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import assert from "node:assert/strict";
import { test } from "node:test";

import { parse, serialize } from "../src/index.js";

function roundTrip(src: string): string {
  return serialize(parse(src).graph);
}

const cases: Array<[string, string]> = [
  ["single node", "node A"],
  ["simple edge", "A->B"],
  ["multiple sources", "A,B -> C"],
  ["node attribute", 'node A [team="orders"]'],
  ["node text", 'node A: "hello"'],
  ["edge text", 'A->B: "hello"'],
  ["edge attribute", "A->B [weight=3]"],
  ["labelled edge", "E1: A->B"],
  ["node block nesting", "node A {\n    node B\n}"],
  ["edge block nesting", "E1: a->b {\n    b->c\n}"],
  [
    "chained parent dependencies",
    "E1: a->b\nE2: b->c [parent = E1]\nc->d [parent = E2]",
  ],
  ["set membership", "set critical\n\nnode A @critical"],
  ["multiple attributes sorted", "node A [z=1, a=2]"],
];

for (const [name, src] of cases) {
  test(`round-trips: ${name}`, () => {
    const once = roundTrip(src);
    const twice = roundTrip(once);
    assert.equal(twice, once);
  });
}

test("canonical order: sets, nodes, edges", () => {
  const out = roundTrip('node A @critical\nA->B');
  assert.equal(out, "set critical\n\nnode A @critical\nnode B\n\nA->B");
});

test("nodes ordered by first edge appearance", () => {
  const out = roundTrip("C->B\nA->D");
  assert.equal(out, "node B\nnode C\nnode A\nnode D\n\nC->B\nA->D");
});

test("parent attribute is implicit inside node blocks", () => {
  const out = roundTrip("node A {\n    node B\n}");
  assert.equal(out, "node A {\n    node B\n}");
});

test("edge attributes are sorted by key", () => {
  const out = roundTrip('A->B [z="1", a="2"]');
  assert.equal(out, 'node A\nnode B\n\nA->B [a="2", z="1"]');
});

test("string escaping round-trips", () => {
  const out = roundTrip('node A [text="line\\nbreak"]');
  assert.equal(out, 'node A [text="line\\nbreak"]');
});
