// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import assert from "node:assert/strict";
import { test } from "node:test";

import { MatchExpandError, expandMatchQuery, parseQuery } from "../src/index.js";

const cases: ReadonlyArray<readonly [string, string, string]> = [
  ["node", 'kind == "service"', 'subgraph node.kind == "service"'],
  ["node", 'kind != "service"', 'subgraph node.kind != "service"'],
  ["node", "active exists", "subgraph node.active exists"],
  ["node", "active not exists", "subgraph node.active not exists"],
  ["node", 'zone == "B" AND team == "payments"', 'subgraph node.zone == "B" AND node.team == "payments"'],
  ["node", "in @critical", "subgraph node in @critical"],
  ["node", "not in @deprecated", "subgraph node not in @deprecated"],
  ["edge", 'protocol == "http"', 'subgraph edge.protocol == "http"'],
  ["edge", "parent exists", "subgraph edge parent exists"],
  ["edge", "parent not exists", "subgraph edge parent not exists"],
  ["edge", 'depends on protocol == "http"', 'subgraph edge depends on edge.protocol == "http"'],
  ["edge", "depth == 0", "subgraph edge.depth == 0"],
  ["edge", "depth != 0", "subgraph edge.depth != 0"],
  ["node", "replicas == 3", "subgraph node.replicas == 3"],
  ["node", 'kind == "a AND b"', 'subgraph node.kind == "a AND b"'],
];

test("expandMatchQuery expands bare predicates", () => {
  for (const [element, input, want] of cases) {
    const got = expandMatchQuery(element, input);
    assert.equal(got, want, `${element}: ${input}`);
    assert.doesNotThrow(() => parseQuery(got), `expanded ${got} should parse`);
  }
});

test("expandMatchQuery rejects expanded forms and out-of-subset input", () => {
  const rejects: ReadonlyArray<readonly [string, string]> = [
    ["node", 'node.kind == "service"'],
    ["node", 'edge.protocol == "http"'],
    ["node", 'set.kind == "service"'],
    ["node", 'subgraph node.kind == "service"'],
    ["node", 'subgraph kind == "service"'],
    ["node", ""],
    ["node", "   "],
    ["node", 'kind == "service" | remove orphans'],
    ["edge", 'kind == "service" traverse out 1'],
    ["set", 'kind == "service"'],
  ];
  for (const [element, input] of rejects) {
    assert.throws(
      () => expandMatchQuery(element, input),
      MatchExpandError,
      `expected ${element}: ${input} to be rejected`,
    );
  }
});
