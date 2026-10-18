// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import type {
  AttributeValue,
  Attributes,
  GslEdge,
  GslGraph,
  GslNode,
  GslSet,
} from "@gsl-lang/core";

export interface WorkNode {
  id: string;
  attributes: Record<string, AttributeValue>;
  sets: Set<string>;
  parent: string | null;
}

export interface WorkEdge {
  from: string;
  to: string;
  label: string | null;
  parent: string | null;
  attributes: Record<string, AttributeValue>;
  sets: Set<string>;
  scopeDepth: number;
}

export interface WorkSet {
  id: string;
  attributes: Record<string, AttributeValue>;
}

export class GraphBuilder {
  readonly nodes = new Map<string, WorkNode>();
  readonly edges: WorkEdge[] = [];
  readonly sets = new Map<string, WorkSet>();

  addNode(node: WorkNode): void {
    this.nodes.set(node.id, node);
  }

  addEdge(edge: WorkEdge): void {
    this.edges.push(edge);
  }

  addSet(set: WorkSet): void {
    this.sets.set(set.id, set);
  }

  finalize(): GslGraph {
    const nodes: GslNode[] = [];
    for (const n of this.nodes.values()) {
      nodes.push({
        id: n.id,
        attributes: { ...n.attributes } as Attributes,
        sets: new Set(n.sets),
        parent: n.parent,
      });
    }
    const edges: GslEdge[] = this.edges.map((e) => ({
      from: e.from,
      to: e.to,
      label: e.label,
      parent: e.parent,
      attributes: { ...e.attributes } as Attributes,
      sets: new Set(e.sets),
      scopeDepth: e.scopeDepth,
    }));
    const sets: GslSet[] = [];
    for (const s of this.sets.values()) {
      sets.push({ id: s.id, attributes: { ...s.attributes } as Attributes });
    }
    return { nodes, edges, sets };
  }
}

export function toWorkGraph(graph: GslGraph): GraphBuilder {
  const b = new GraphBuilder();
  for (const n of graph.nodes) {
    b.addNode({
      id: n.id,
      attributes: { ...n.attributes },
      sets: new Set(n.sets),
      parent: n.parent ?? null,
    });
  }
  for (const e of graph.edges) {
    b.addEdge({
      from: e.from,
      to: e.to,
      label: e.label ?? null,
      parent: e.parent ?? null,
      attributes: { ...e.attributes },
      sets: new Set(e.sets),
      scopeDepth: e.scopeDepth ?? 0,
    });
  }
  for (const s of graph.sets ?? []) {
    b.addSet({ id: s.id, attributes: { ...s.attributes } });
  }
  return b;
}
