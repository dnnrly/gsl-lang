// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import assert from "node:assert/strict";
import { test } from "node:test";
import { existsSync, readFileSync, readdirSync } from "node:fs";
import { dirname, join, relative } from "node:path";
import { fileURLToPath } from "node:url";

import { parse, serialize } from "../src/index.js";

/**
 * Every `examples/**\/*.gsl` file is expected to parse and survive the
 * canonical round-trip `serialize(parse(x))`, matching the canonical-form
 * guarantee that the Go reference implementation asserts.
 *
 * Draft files that are deliberately not yet valid GSL (used to demonstrate an
 * LLM-modelling workflow) are expected to fail to parse instead.
 */
const INTENTIONALLY_INVALID = new Set<string>([
  "examples/flagships/05-llm-assisted-modelling/draft-01.gsl",
]);

function findRepoRoot(): string {
  let dir = dirname(fileURLToPath(import.meta.url));
  while (dir !== dirname(dir)) {
    if (existsSync(join(dir, "tests", "conformance", "gql"))) return dir;
    dir = dirname(dir);
  }
  throw new Error("could not locate repository root");
}

function collectFiles(dir: string, suffix: string, out: string[] = []): string[] {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    if (entry.name === "node_modules") continue;
    const path = join(dir, entry.name);
    if (entry.isDirectory()) {
      collectFiles(path, suffix, out);
    } else if (path.endsWith(suffix)) {
      out.push(path);
    }
  }
  return out;
}

const repoRoot = findRepoRoot();
const fixtures = collectFiles(join(repoRoot, "examples"), ".gsl").sort();

test("examples corpus discovered", () => {
  assert.ok(fixtures.length > 0, "no example .gsl files found");
});

for (const fixture of fixtures) {
  const rel = relative(repoRoot, fixture).split("\\").join("/");

  if (INTENTIONALLY_INVALID.has(rel)) {
    test(`intentionally invalid: ${rel}`, () => {
      assert.throws(() => parse(readFileSync(fixture, "utf8")));
    });
    continue;
  }

  test(`round-trip: ${rel}`, () => {
    const graph = parse(readFileSync(fixture, "utf8")).graph;
    const once = serialize(graph);
    const twice = serialize(parse(once).graph);
    assert.equal(twice, once);
  });
}
