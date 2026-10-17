// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import assert from "node:assert/strict";
import { test } from "node:test";
import { existsSync, readFileSync, readdirSync } from "node:fs";
import { dirname, join, relative } from "node:path";
import { fileURLToPath } from "node:url";

import { parse, serialize } from "@gsl-lang/core";
import { runQuery } from "../src/index.js";

/**
 * Cross-implementation behaviour checks over the shared exemplar fixtures.
 *
 * For every flagship and advanced-study model:
 *  - `model.gsl` must parse with no errors and no warnings (sets are declared
 *    explicitly);
 *  - every `qN-*.gql` must execute and serialize byte-for-byte equal to the
 *    committed `qN-*.result.gsl` (which the Go reference implementation
 *    generated);
 *  - queries and results must correspond one-to-one.
 */

function findRepoRoot(): string {
  let dir = dirname(fileURLToPath(import.meta.url));
  while (dir !== dirname(dir)) {
    if (existsSync(join(dir, "tests", "conformance", "gql"))) return dir;
    dir = dirname(dir);
  }
  throw new Error("could not locate repository root");
}

function dirsWithModel(dir: string, out: string[] = []): string[] {
  if (!existsSync(dir)) return out;
  if (existsSync(join(dir, "model.gsl"))) {
    out.push(dir);
    return out;
  }
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    if (entry.isDirectory()) dirsWithModel(join(dir, entry.name), out);
  }
  return out;
}

function trim(s: string): string {
  return s.replace(/\n+$/, "");
}

const repoRoot = findRepoRoot();
const modelDirs = [
  ...dirsWithModel(join(repoRoot, "examples", "flagships")),
  ...dirsWithModel(join(repoRoot, "examples", "advanced")),
].sort();

test("flagship/advanced models discovered", () => {
  assert.ok(modelDirs.length >= 5, `expected >=5 exemplar models, found ${modelDirs.length}`);
});

for (const dir of modelDirs) {
  const rel = relative(repoRoot, dir).split("\\").join("/");
  const model = parse(readFileSync(join(dir, "model.gsl"), "utf8"));

  test(`model parses cleanly: ${rel}`, () => {
    assert.equal(model.warnings?.length ?? 0, 0, "model.gsl produced warnings");
    const count = model.graph.nodes.length;
    assert.ok(count >= 5 && count <= 40, `unexpected node count ${count}`);
  });

  const entries = readdirSync(dir);
  const queries = entries.filter((n) => /^q.*\.gql$/.test(n)).sort();
  const results = entries.filter((n) => /^q.*\.result\.gsl$/.test(n)).sort();

  test(`queries and results correspond: ${rel}`, () => {
    assert.ok(queries.length > 0, "no q*.gql query files found");
    assert.deepEqual(
      queries.map((q) => q.replace(/\.gql$/, "")),
      results.map((r) => r.replace(/\.result\.gsl$/, "")),
    );
  });

  for (const queryFile of queries) {
    const base = queryFile.replace(/\.gql$/, "");

    test(`query: ${rel}/${queryFile}`, () => {
      const expected = trim(readFileSync(join(dir, `${base}.result.gsl`), "utf8"));
      const actual = trim(serialize(runQuery(model.graph, readFileSync(join(dir, queryFile), "utf8"))));
      assert.equal(actual, expected);
    });
  }
}
