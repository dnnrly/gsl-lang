// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import assert from "node:assert/strict";
import { existsSync, readFileSync, readdirSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";
import { test } from "node:test";

import { parse, serialize } from "../src/index.js";

// The shared, language-neutral GSL canonical-form fixtures live in the
// repository's tests/conformance/gsl/canonical directory. Walk up from this
// compiled test file until we find it.
function findConformanceDir(): string {
  let dir = dirname(fileURLToPath(import.meta.url));
  while (true) {
    const candidate = join(dir, "tests", "conformance", "gsl", "canonical");
    if (existsSync(candidate)) return candidate;
    const parent = dirname(dir);
    if (parent === dir) throw new Error("gsl conformance corpus not found");
    dir = parent;
  }
}

const conformanceDir = findConformanceDir();
const fixtures = readdirSync(conformanceDir, { withFileTypes: true })
  .filter((entry) => entry.isDirectory())
  .map((entry) => entry.name)
  .sort();

test("gsl conformance fixtures discovered", () => {
  assert.ok(fixtures.length > 0, "no fixtures found");
});

for (const name of fixtures) {
  test(`gsl conformance: ${name}`, () => {
    const dir = join(conformanceDir, name);
    const inputSrc = readFileSync(join(dir, "input.gsl"), "utf8");
    const expectedSrc = readFileSync(join(dir, "expected.gsl"), "utf8");

    const actual = serialize(parse(inputSrc).graph).replace(/\n+$/, "");
    const expected = expectedSrc.replace(/\n+$/, "");

    assert.equal(actual, expected);
  });
}
