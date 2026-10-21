// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

/**
 * `@gsl-lang/gql` — TypeScript implementation of the GSL query language (GQL).
 */

export { parseQuery, GqlParseError } from "./parser.js";
export { expandMatchQuery, MatchExpandError } from "./expand.js";
export {
  Query,
  SubgraphExpr,
  type QueryContext,
  type Value,
  type Expression,
  type TraversalConfig,
} from "./expressions.js";
export type { Predicate } from "./predicates.js";
export {
  GraphBuilder,
  toWorkGraph,
  type WorkEdge,
  type WorkNode,
  type WorkSet,
} from "./graph.js";

import type { GslGraph } from "@gsl-lang/core";
import { parseQuery } from "./parser.js";
import { Query, type QueryContext, type Value } from "./expressions.js";

/**
 * Execute a GQL query against an input graph, returning the resulting graph.
 */
export function runQuery(input: GslGraph, query: string): GslGraph {
  const parsed: Query = parseQuery(query);
  const ctx: QueryContext = {
    inputGraph: input,
    namedGraphs: new Map<string, GslGraph>(),
  };
  const result: Value = parsed.execute(ctx);
  return result.graph;
}
