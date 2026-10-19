// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import assert from "node:assert/strict";
import { test } from "node:test";
import { existsSync, readFileSync, readdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

import { parse } from "@gsl-lang/core";
import { parse as parseYaml } from "yaml";

import { LoadError, load } from "../src/index.js";

interface ExpectedViolation {
  definition: string;
  element: string;
  location: string;
}

interface Expectation {
  outcome: string;
  projection?: Record<string, Record<string, unknown>>;
  violations?: ExpectedViolation[];
}

// The shared, language-neutral dialect fixtures live in the repository's
// tests/conformance/dialects directory. Walk up from this compiled test file
// until we find it.
function findConformanceDir(): string {
  let dir = dirname(fileURLToPath(import.meta.url));
  while (true) {
    const candidate = join(dir, "tests", "conformance", "dialects");
    if (existsSync(candidate)) return candidate;
    const parent = dirname(dir);
    if (parent === dir) throw new Error("dialect conformance corpus not found");
    dir = parent;
  }
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
      if (existsSync(join(fixture, "dialect.yaml"))) results.push(fixture);
    }
  }
  return results.sort();
}

const conformanceDir = findConformanceDir();
const fixtures = findFixtures(conformanceDir);

test("dialect conformance fixtures discovered", () => {
  assert.ok(fixtures.length > 0, "no fixtures found");
});

for (const fixture of fixtures) {
  const rel = fixture.slice(conformanceDir.length + 1);
  test(`dialect fixture: ${rel}`, () => {
    const dialectSrc = readFileSync(join(fixture, "dialect.yaml"), "utf8");
    const expected = parseYaml(readFileSync(join(fixture, "expected.yaml"), "utf8")) as Expectation;
    assert.ok(expected.outcome, "expected.yaml is missing an outcome");

    let d;
    try {
      d = load(dialectSrc);
    } catch (err) {
      if (expected.outcome === "invalid-dialect") {
        assert.ok(err instanceof LoadError, `expected a LoadError, got ${String(err)}`);
        return;
      }
      throw err;
    }
    assert.notEqual(expected.outcome, "invalid-dialect", "expected the dialect to be rejected");

    const graphSrc = readFileSync(join(fixture, "graph.gsl"), "utf8");
    const graph = parse(graphSrc).graph;

    if (expected.projection !== undefined) {
      assertProjection(d, graph, expected.projection);
    }

    const report = d.validate(graph);
    assert.equal(report.outcome, expected.outcome, JSON.stringify(report.violations));

    const got = report.violations
      .map((v) => `${v.definition}\u0000${v.element}\u0000${v.instanceLocation}`)
      .sort();
    const want = (expected.violations ?? [])
      .map((v) => `${v.definition}\u0000${v.element}\u0000${v.location}`)
      .sort();
    assert.deepEqual(got, want, JSON.stringify(report.violations));
  });
}

function assertProjection(
  d: ReturnType<typeof load>,
  graph: ReturnType<typeof parse>["graph"],
  want: Record<string, Record<string, unknown>>,
): void {
  const got: Record<string, Record<string, unknown>> = {};
  for (const s of d.match(graph)) {
    got[s.definition] ??= {};
    got[s.definition]![s.element] = s.attributes;
  }
  assert.deepEqual(got, want);
}
