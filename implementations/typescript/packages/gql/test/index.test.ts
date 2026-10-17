// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import assert from "node:assert/strict";
import { test } from "node:test";

import { parse, serialize } from "@gsl-lang/core";
import { runQuery } from "../src/index.js";

const SOURCE = [
  'node api [team="platform"]',
  'node db [team="data"]',
  'node cache [team="data"]',
  'api -> db [protocol="http"]',
  "api -> cache",
  "db -> cache",
].join("\n");

function graph() {
  return parse(SOURCE).graph;
}

test("subgraph by attribute equality keeps matching nodes and their edges", () => {
  const result = runQuery(graph(), 'subgraph node.team == "data"');
  assert.equal(
    serialize(result),
    'node cache [team="data"]\nnode db [team="data"]\n\ndb->cache',
  );
});

test("missing attribute is not an error", () => {
  const result = runQuery(graph(), 'subgraph node.nope == "x"');
  assert.equal(serialize(result), "");
});

test("traversal keeps the reachable subgraph", () => {
  const result = runQuery(graph(), 'subgraph node.id == "api" traverse out 1');
  assert.equal(
    serialize(result),
    [
      'node api [team="platform"]',
      'node db [team="data"]',
      'node cache [team="data"]',
      "",
      "api->db [protocol=\"http\"]",
      "api->cache",
      "db->cache",
    ].join("\n"),
  );
});

test("make sets an attribute on matching nodes", () => {
  const result = runQuery(graph(), 'make node.reviewed = true where node.team == "data"');
  assert.equal(
    serialize(result),
    [
      'node api [team="platform"]',
      'node db [reviewed=true, team="data"]',
      'node cache [reviewed=true, team="data"]',
      "",
      "api->db [protocol=\"http\"]",
      "api->cache",
      "db->cache",
    ].join("\n"),
  );
});

test("empty query returns the input graph unchanged", () => {
  const result = runQuery(graph(), "");
  assert.equal(serialize(result), serialize(graph()));
});

test("executing a query against an unbound named graph fails", () => {
  assert.throws(() => runQuery(graph(), "from SOMEGRAPH"), /named graph not found/);
});
