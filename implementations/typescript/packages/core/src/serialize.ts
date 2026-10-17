// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import type { GslGraph, GslNode, GslEdge, Attributes } from "./model.js";

export function serialize(graph: GslGraph): string {
  if (!graph) return "";

  const sections: string[] = [];

  const sets = collectAllSets(graph);
  if (sets.size > 0) {
    const setIDs = Array.from(sets.keys()).sort();
    const lines: string[] = [];
    for (const id of setIDs) {
      lines.push(serializeSet(sets.get(id)!));
    }
    sections.push(lines.join("\n"));
  }

  if (graph.nodes.length > 0) {
    sections.push(serializeNodes(graph.nodes, graph.edges));
  }

  if (graph.edges.length > 0) {
    sections.push(serializeEdges(graph.edges));
  }

  return sections.join("\n\n");
}

interface SerializeSetEntry {
  id: string;
  attributes: Attributes;
}

function serializeSet(s: SerializeSetEntry): string {
  let out = "set " + s.id;
  if (Object.keys(s.attributes).length > 0) {
    out += " " + serializeAttrs(s.attributes);
  }
  return out;
}

function serializeNodes(nodes: readonly GslNode[], edges: readonly GslEdge[]): string {
  const positions = buildNodePositions(edges);
  const nodeMap = new Map<string, GslNode>();
  for (const n of nodes) {
    nodeMap.set(n.id, n);
  }

  const children = new Map<string, GslNode[]>();
  for (const n of nodes) {
    if (n.parent) {
      if (nodeMap.has(n.parent) && !isNodeCyclic(n, nodeMap)) {
        if (!children.has(n.parent)) {
          children.set(n.parent, []);
        }
        children.get(n.parent)!.push(n);
      }
    }
  }

  for (const id of children.keys()) {
    sortNodesByPosition(children.get(id)!, positions);
  }

  const roots: GslNode[] = [];
  for (const n of nodes) {
    if (!n.parent || !nodeMap.has(n.parent) || isNodeCyclic(n, nodeMap)) {
      roots.push(n);
    }
  }
  sortNodesByPosition(roots, positions);

  const lines: string[] = [];
  for (const n of roots) {
    lines.push(serializeNodeNested(n, children.get(n.id) ?? [], children, new Set<string>(), false));
  }
  return lines.join("\n");
}

function buildNodePositions(edges: readonly GslEdge[]): Map<string, number> {
  const positions = new Map<string, number>();
  for (let i = 0; i < edges.length; i++) {
    const e = edges[i]!;
    if (!positions.has(e.from)) positions.set(e.from, i);
    if (!positions.has(e.to)) positions.set(e.to, i);
  }
  return positions;
}

function sortNodesByPosition(nodes: GslNode[], positions: Map<string, number>): void {
  nodes.sort((a, b) => {
    const pi = positions.get(a.id);
    const pj = positions.get(b.id);
    const hasI = pi !== undefined;
    const hasJ = pj !== undefined;
    if (hasI && hasJ) {
      if (pi !== pj) return pi! - pj!;
      return a.id < b.id ? -1 : a.id > b.id ? 1 : 0;
    }
    if (hasI) return -1;
    if (hasJ) return 1;
    return a.id < b.id ? -1 : a.id > b.id ? 1 : 0;
  });
}

function isNodeCyclic(n: GslNode, allNodes: Map<string, GslNode>): boolean {
  const visited = new Set<string>();
  let current: GslNode | undefined = n;
  while (current && current.parent) {
    if (visited.has(current.id)) return true;
    visited.add(current.id);
    const parent = allNodes.get(current.parent);
    if (!parent) return false;
    current = parent;
  }
  return false;
}

function serializeNodeNested(
  n: GslNode,
  children: readonly GslNode[],
  childMap: Map<string, GslNode[]>,
  visiting: Set<string>,
  stripParent: boolean,
): string {
  let out = "node " + n.id;

  let attrs: Attributes;
  if (stripParent) {
    attrs = {};
    for (const k of Object.keys(n.attributes)) {
      if (k !== "parent") (attrs as any)[k] = (n.attributes as any)[k];
    }
  } else {
    attrs = n.attributes;
  }
  if (Object.keys(attrs).length > 0) {
    out += " " + serializeAttrs(attrs);
  }

  if (children.length > 0 && !visiting.has(n.id)) {
    visiting.add(n.id);
    out += " {\n";
    for (const child of children) {
      const grandchildren = childMap.get(child.id) ?? [];
      const childStr = serializeNodeNested(child, grandchildren, childMap, visiting, true);
      out += indentBlock(childStr, "    ");
      out += "\n";
    }
    out += "}";
    visiting.delete(n.id);
  }

  out += serializeSetMemberships(n.sets);
  return out;
}

function serializeEdges(edges: readonly GslEdge[]): string {
  const labelIndex = new Map<string, GslEdge>();
  for (const e of edges) {
    if (e.label) {
      if (!labelIndex.has(e.label)) {
        labelIndex.set(e.label, e);
      }
    }
  }

  const childMap = new Map<string, GslEdge[]>();
  const childSet = new Set<GslEdge>();
  for (const e of edges) {
    if (e.parent) {
      const parentExists = labelIndex.has(e.parent);
      const cyclic = parentExists && isEdgeCyclic(e, labelIndex);
      if (parentExists && !cyclic) {
        if (!childMap.has(e.parent)) childMap.set(e.parent, []);
        childMap.get(e.parent)!.push(e);
        childSet.add(e);
      }
    }
  }

  const labelGroups = new Map<string, GslEdge[]>();
  for (const e of edges) {
    if (!childSet.has(e) && e.label) {
      if (!labelGroups.has(e.label)) labelGroups.set(e.label, []);
      labelGroups.get(e.label)!.push(e);
    }
  }

  const lines: string[] = [];
  const visiting = new Set<string>();
  const seenLabels = new Set<string>();

  for (const e of edges) {
    if (childSet.has(e)) continue;
    if (e.label) {
      if (!seenLabels.has(e.label)) {
        seenLabels.add(e.label);
        lines.push(serializeLabeledEdgeGroup(labelGroups.get(e.label)!, childMap, visiting));
      }
    } else {
      lines.push(serializeEdgeNested(e, childMap, visiting));
    }
  }
  return lines.join("\n");
}

function isEdgeCyclic(e: GslEdge, labelIndex: Map<string, GslEdge>): boolean {
  const visited = new Set<GslEdge>();
  let current: GslEdge | undefined = e;
  while (current && current.parent) {
    if (visited.has(current)) return true;
    visited.add(current);
    const parent = labelIndex.get(current.parent);
    if (!parent) return false;
    current = parent;
  }
  return false;
}

function serializeEdgeNested(e: GslEdge, childMap: Map<string, GslEdge[]>, visiting: Set<string>): string {
  let out = "";
  if (e.label) {
    out += e.label + ": ";
  }
  out += e.from + "->" + e.to;
  if (Object.keys(e.attributes).length > 0) {
    out += " " + serializeAttrs(e.attributes);
  }
  out += serializeSetMemberships(e.sets);

  const children = childMap.get(e.label ?? "");
  if (children && children.length > 0 && !visiting.has(e.label ?? "")) {
    visiting.add(e.label ?? "");
    out += " {\n";
    for (const child of children) {
      const childStr = serializeEdgeNested(child, childMap, visiting);
      out += indentBlock(childStr, "    ");
      out += "\n";
    }
    out += "}";
    visiting.delete(e.label ?? "");
  }
  return out;
}

function serializeLabeledEdgeGroup(
  edges: readonly GslEdge[],
  childMap: Map<string, GslEdge[]>,
  visiting: Set<string>,
): string {
  if (edges.length === 0) return "";
  const label = edges[0]!.label ?? "";

  let allShareTo = true;
  for (const e of edges.slice(1)) {
    if (e.to !== edges[0]!.to) {
      allShareTo = false;
      break;
    }
  }
  let allShareFrom = true;
  for (const e of edges.slice(1)) {
    if (e.from !== edges[0]!.from) {
      allShareFrom = false;
      break;
    }
  }

  let out = label + ": ";
  if (allShareTo) {
    for (let i = 0; i < edges.length; i++) {
      if (i > 0) out += ",";
      out += edges[i]!.from;
    }
    out += "->" + edges[0]!.to;
  } else if (allShareFrom) {
    out += edges[0]!.from + "->";
    for (let i = 0; i < edges.length; i++) {
      if (i > 0) out += ",";
      out += edges[i]!.to;
    }
  } else {
    const lines: string[] = [];
    for (const e of edges) {
      lines.push(serializeEdgeNested(e, childMap, visiting));
    }
    return lines.join("\n");
  }

  if (Object.keys(edges[0]!.attributes).length > 0) {
    out += " " + serializeAttrs(edges[0]!.attributes);
  }
  out += serializeSetMemberships(edges[0]!.sets);

  const children = childMap.get(label);
  if (children && children.length > 0 && !visiting.has(label)) {
    visiting.add(label);
    out += " {\n";
    for (const child of children) {
      const childStr = serializeEdgeNested(child, childMap, visiting);
      out += indentBlock(childStr, "    ");
      out += "\n";
    }
    out += "}";
    visiting.delete(label);
  }
  return out;
}

function collectAllSets(graph: GslGraph): Map<string, SerializeSetEntry> {
  const sets = new Map<string, SerializeSetEntry>();
  for (const s of graph.sets ?? []) {
    sets.set(s.id, { id: s.id, attributes: s.attributes });
  }
  for (const n of graph.nodes) {
    for (const s of n.sets) {
      if (!sets.has(s)) sets.set(s, { id: s, attributes: {} });
    }
  }
  for (const e of graph.edges) {
    for (const s of e.sets) {
      if (!sets.has(s)) sets.set(s, { id: s, attributes: {} });
    }
  }
  return sets;
}

function serializeSetMemberships(sets: ReadonlySet<string>): string {
  if (sets.size === 0) return "";
  const sorted = Array.from(sets).sort();
  return sorted.map((s) => ` @${s}`).join("");
}

function serializeAttrs(attrs: Attributes): string {
  const keys = Object.keys(attrs).sort();
  const parts: string[] = [];
  for (const k of keys) {
    parts.push(serializeAttr(k, (attrs as any)[k]));
  }
  return "[" + parts.join(", ") + "]";
}

function serializeAttr(key: string, value: unknown): string {
  if (value === null || value === undefined) {
    return key;
  }
  if (typeof value === "string") {
    if (key === "parent") {
      return key + "=" + value;
    }
    return key + "=" + '"' + escapeString(value) + '"';
  }
  if (typeof value === "number") {
    return key + "=" + formatNumber(value);
  }
  if (typeof value === "boolean") {
    return key + "=" + (value ? "true" : "false");
  }
  return key;
}

function escapeString(s: string): string {
  let out = "";
  for (const ch of s) {
    switch (ch) {
      case '"':
        out += '\\"';
        break;
      case "\\":
        out += "\\\\";
        break;
      case "\n":
        out += "\\n";
        break;
      case "\t":
        out += "\\t";
        break;
      default:
        out += ch;
    }
  }
  return out;
}

function formatNumber(f: number): string {
  if (Number.isFinite(f) && Math.floor(f) === f) {
    return f.toFixed(0);
  }
  return String(f);
}

function indentBlock(s: string, indent: string): string {
  return s
    .split("\n")
    .map((line) => indent + line)
    .join("\n");
}
