// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import assert from "node:assert/strict";
import { test } from "node:test";

import { parse, serialize } from "../src/index.js";

test("parse single node", () => {
  const res = parse("node A");
  assert.equal(res.graph.nodes.length, 1);
  assert.equal(res.graph.nodes[0]!.id, "A");
});

test("parse edge", () => {
  const res = parse("A->B");
  assert.equal(res.graph.edges.length, 1);
  assert.equal(res.graph.edges[0]!.from, "A");
  assert.equal(res.graph.edges[0]!.to, "B");
});
