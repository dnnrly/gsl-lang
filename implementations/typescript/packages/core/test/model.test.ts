// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import assert from "node:assert/strict";
import { test } from "node:test";

import {
  SPEC_VERSION,
  isAttributeValue,
  type GslGraph,
} from "../src/model.js";

test("SPEC_VERSION tracks the targeted specification", () => {
  assert.equal(SPEC_VERSION, "1.0.0");
});

test("isAttributeValue accepts GSL attribute values", () => {
  assert.equal(isAttributeValue("blue"), true);
  assert.equal(isAttributeValue(42), true);
  assert.equal(isAttributeValue(true), true);
});

test("isAttributeValue rejects non-attributes", () => {
  assert.equal(isAttributeValue(null), false);
  assert.equal(isAttributeValue(undefined), false);
  assert.equal(isAttributeValue({ name: "A" }), false);
  assert.equal(isAttributeValue(["A"]), false);
});

test("the model can describe a graph", () => {
  const graph: GslGraph = {
    nodes: [
      { id: "A", attributes: { team: "orders" }, sets: new Set(["critical"]) },
      { id: "B", attributes: {}, sets: new Set() },
    ],
    edges: [{ from: "A", to: "B", attributes: {}, sets: new Set() }],
  };
  assert.equal(graph.nodes.length, 2);
  assert.equal(graph.edges[0]?.from, "A");
});
