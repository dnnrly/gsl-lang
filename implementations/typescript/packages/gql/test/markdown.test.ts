// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import assert from "node:assert/strict";
import { test } from "node:test";
import { existsSync, readFileSync, readdirSync } from "node:fs";
import { dirname, join, relative } from "node:path";
import { fileURLToPath } from "node:url";

import { parse } from "@gsl-lang/core";
import { parseQuery } from "../src/index.js";

/**
 * Mirrors the Go `TestMarkdownCodeBlocks` corpus: every fenced `gsl` and `gql`
 * block in the specs, guides and docs must parse, and every `invalid-gsl` /
 * `invalid-gql` block must fail. This keeps the two implementations agreeing on
 * the language's acceptance boundary as documented.
 */
const TREE_ROOTS = ["docs", "spec", "guides", "tests", "implementations", "examples"];

const FENCE_RE = /^```([a-z-]*)\n([\s\S]*?)\n```/gm;

function findRepoRoot(): string {
  let dir = dirname(fileURLToPath(import.meta.url));
  while (dir !== dirname(dir)) {
    if (existsSync(join(dir, "tests", "conformance", "gql"))) return dir;
    dir = dirname(dir);
  }
  throw new Error("could not locate repository root");
}

function collectMarkdown(dir: string, out: string[] = []): string[] {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    if (entry.name === "node_modules") continue;
    const path = join(dir, entry.name);
    if (entry.isDirectory()) {
      collectMarkdown(path, out);
    } else if (path.endsWith(".md")) {
      out.push(path);
    }
  }
  return out;
}

const repoRoot = findRepoRoot();
const files = readdirSync(repoRoot)
  .filter((name) => name.endsWith(".md"))
  .map((name) => join(repoRoot, name));

for (const tree of TREE_ROOTS) {
  const root = join(repoRoot, tree);
  if (existsSync(root)) collectMarkdown(root, files);
}

const counts = { gsl: 0, invalidGsl: 0, gql: 0, invalidGql: 0 };

for (const file of files) {
  const rel = relative(repoRoot, file).split("\\").join("/");
  const blocks = [...readFileSync(file, "utf8").matchAll(FENCE_RE)];

  for (const block of blocks) {
    const language = block[1];
    const code = block[2] ?? "";

    if (language === "gsl") {
      counts.gsl++;
      test(`valid gsl block: ${rel}`, () => {
        parse(code);
      });
    } else if (language === "invalid-gsl") {
      counts.invalidGsl++;
      test(`invalid gsl block: ${rel}`, () => {
        assert.throws(() => parse(code));
      });
    } else if (language === "gql") {
      counts.gql++;
      test(`valid gql block: ${rel}`, () => {
        parseQuery(code);
      });
    } else if (language === "invalid-gql") {
      counts.invalidGql++;
      test(`invalid gql block: ${rel}`, () => {
        assert.throws(() => parseQuery(code));
      });
    }
  }
}

test("markdown corpus discovered", () => {
  assert.ok(counts.gsl > 0 && counts.gql > 0, `unexpected block counts: ${JSON.stringify(counts)}`);
});
