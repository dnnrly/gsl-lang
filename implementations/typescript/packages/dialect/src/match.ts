// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import type { Attributes, GslEdge, GslGraph } from "@gsl-lang/core";
import type { QueryContext, SubgraphExpr } from "@gsl-lang/gql";

import type { Representation } from "./result.js";

/** A graph element selected for validation, with its raw GSL attributes. */
export interface SelectedElement {
  readonly ref: string;
  readonly attributes: Attributes;
}

/**
 * Evaluate the definition's match query against `graph` and return the eligible
 * elements for its representation, in a deterministic order (nodes and sets by
 * id, edges in graph order).
 *
 * Matching determines which elements are subject to the definition's
 * constraints; it does not establish that they are valid. Elements whose
 * attributes later fail validation remain selected.
 */
export function selectElements(
  representation: Representation,
  expr: SubgraphExpr,
  graph: GslGraph,
): SelectedElement[] {
  const ctx: QueryContext = { inputGraph: graph, namedGraphs: new Map() };
  const matched = expr.apply(ctx, { graph }).graph;

  switch (representation) {
    case "node": {
      const nodes = [...matched.nodes].sort((a, b) => (a.id < b.id ? -1 : a.id > b.id ? 1 : 0));
      return nodes.map((n) => ({ ref: n.id, attributes: n.attributes }));
    }
    case "edge": {
      return matched.edges.map((e: GslEdge) => ({ ref: edgeRef(e), attributes: e.attributes }));
    }
    case "set": {
      const sets = [...(matched.sets ?? [])].sort((a, b) =>
        a.id < b.id ? -1 : a.id > b.id ? 1 : 0,
      );
      return sets.map((s) => ({ ref: s.id, attributes: s.attributes }));
    }
  }
}

function edgeRef(e: GslEdge): string {
  if (e.label !== undefined && e.label !== null && e.label !== "") return e.label;
  return `${e.from}->${e.to}`;
}
