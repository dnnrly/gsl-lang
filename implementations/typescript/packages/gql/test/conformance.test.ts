// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import assert from "node:assert/strict";
import { test } from "node:test";
import { existsSync, readFileSync, readdirSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { parse, serialize } from "@gsl-lang/core";
import { runQuery } from "../src/index.js";

function findConformanceDir(): string {
  let dir = dirname(fileURLToPath(import.meta.url));
  while (dir !== dirname(dir)) {
    const candidate = join(dir, "tests", "conformance", "gql");
    if (existsSync(candidate)) return candidate;
    dir = dirname(dir);
  }
  throw new Error("could not locate tests/conformance/gql");
}

function findFixtures(root: string): string[] {
  const results: string[] = [];
  for (const group of readdirSync(root)) {
    const groupPath = join(root, group);
    let entries: string[] = [];
    try {
      entries = readdirSync(groupPath);
    } catch {
      continue;
    }
    for (const name of entries) {
      const fixture = join(groupPath, name);
      if (existsSync(join(fixture, "graph.gsl"))) {
        results.push(fixture);
      }
    }
  }
  return results.sort();
}

const conformanceDir = findConformanceDir();
const fixtures = findFixtures(conformanceDir);

test("conformance fixtures discovered", () => {
  assert.ok(fixtures.length > 0, "no fixtures found");
});

for (const fixture of fixtures) {
  const rel = fixture.slice(conformanceDir.length + 1);
  test(`fixture: ${rel}`, () => {
    const graphSrc = readFileSync(join(fixture, "graph.gsl"), "utf8");
    const querySrc = readFileSync(join(fixture, "query.gql"), "utf8");
    const resultSrc = readFileSync(join(fixture, "result.gsl"), "utf8");

    const graph = parse(graphSrc).graph;
    const actualGraph = runQuery(graph, querySrc);
    const actual = serialize(actualGraph);
    const expected = serialize(parse(resultSrc).graph);

    assert.equal(actual, expected);
  });
}
