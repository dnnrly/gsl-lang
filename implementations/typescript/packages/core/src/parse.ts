// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import { parseProgram } from "./parser.js";
import { buildGraph } from "./build.js";
import type { GslGraph } from "./model.js";
import { ParseError } from "./errors.js";

export interface ParseResult {
  graph: GslGraph;
  warnings?: any[];
}

export function parse(input: string): ParseResult {
  const { program, errors } = parseProgram(input);
  if (errors.length > 0) {
    throw new ParseError("parse failed", errors);
  }
  const { graph, warnings } = buildGraph(program as any);
  const result: ParseResult = { graph };
  if (warnings && warnings.length > 0) {
    (result as any).warnings = warnings;
  }
  return result;
}
