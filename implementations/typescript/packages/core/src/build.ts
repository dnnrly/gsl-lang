// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import type { Program, Statement, NodeDecl, EdgeDecl, SetDecl } from "./ast.js";
import { makeNodeRef, type GslGraph, type GslNode, type GslEdge, type GslSet } from "./model.js";
import { ParseError } from "./errors.js";

interface EdgeScope {
  labels: Map<string, GslEdgeInternal>;
  parent: string | undefined;
  outer?: EdgeScope;
}

interface GslNodeInternal {
  id: string;
  attributes: Record<string, any>;
  sets: Set<string>;
  parent?: string | null;
}

interface GslEdgeInternal {
  from: string;
  to: string;
  label: string | undefined;
  parent: string | undefined;
  attributes: Record<string, any>;
  sets: Set<string>;
  scopeDepth: number;
  children: GslEdgeInternal[];
}

class GraphBuilder {
  nodes = new Map<string, GslNodeInternal>();
  edges: GslEdgeInternal[] = [];
  sets = new Map<string, { id: string; attributes: Record<string, any>; declared: boolean }>();
  warnings: any[] = [];
  errors: any[] = [];

  scope: EdgeScope = { labels: new Map<string, GslEdgeInternal>(), parent: undefined, outer: undefined as any };
  scopeDepth = 0;

  ensureNode(name: string): GslNodeInternal {
    let n = this.nodes.get(name);
    if (!n) {
      n = {
        id: name,
        attributes: {},
        sets: new Set(),
      };
      this.nodes.set(name, n);
    }
    return n;
  }

  ensureSet(name: string, implicit: boolean): void {
    if (!this.sets.has(name)) {
      this.sets.set(name, { id: name, attributes: {}, declared: !implicit });
      if (implicit) {
        this.warnings.push({ message: `implicit set creation: ${name}` });
      }
    }
  }

  setDeclared(name: string): boolean {
    return this.sets.get(name)?.declared === true;
  }
}

export function buildGraph(program: Program): { graph: GslGraph; errors?: any[]; warnings?: any[] } {
  const b = new GraphBuilder();
  for (const stmt of program.statements) {
    switch (stmt.kind) {
      case "node":
        processNodeDecl(b, stmt);
        break;
      case "edge":
        processEdgeDecl(b, stmt);
        break;
      case "set":
        processSetDecl(b, stmt);
        break;
    }
  }
  populateChildren(b);
  validateDeps(b);

  const graph = toGraph(b);
  if (b.errors.length > 0) {
    throw new ParseError("build failed", b.errors, b.warnings);
  }
  return { graph, warnings: b.warnings.length > 0 ? b.warnings : undefined } as any;
}

function processNodeDecl(b: GraphBuilder, nd: NodeDecl, enclosingParent?: string): void {
  const node = b.ensureNode(nd.name);

  if (nd.textValue !== undefined) {
    node.attributes["text"] = nd.textValue;
  }

  for (const attr of nd.attrs) {
    node.attributes[attr.key] = convertAttrValue(attr.value);
  }

  if (enclosingParent !== undefined) {
    if (!("parent" in node.attributes)) {
      node.attributes["parent"] = makeNodeRef(enclosingParent);
    } else {
      b.warnings.push({ message: `parent override inside block`, line: nd.line, column: nd.col });
    }
  }

  const ref = node.attributes["parent"];
  if (ref && typeof ref === "string") {
    node.parent = ref;
  } else if (ref && typeof ref === "object" && (ref as any).__brand) {
    node.parent = String(ref);
  }

  for (const setName of nd.memberships) {
    b.ensureSet(setName, !b.setDeclared(setName));
    node.sets.add(setName);
  }

  if (b.sets.has(nd.name)) {
    b.warnings.push({ message: `node and set name collision: ${nd.name}` });
  }

  for (const child of nd.block) {
    processNodeDecl(b, child, nd.name);
  }
}

function convertAttrValue(v: any): any {
  if (!v) return v;
  switch (v.kind) {
    case "string":
      return v.strVal;
    case "number":
      return v.numVal;
    case "bool":
      return v.boolVal;
    case "noderef":
      return makeNodeRef(v.refVal);
    default:
      return undefined;
  }
}

function processSetDecl(b: GraphBuilder, sd: SetDecl): void {
  if (b.nodes.has(sd.name)) {
    b.warnings.push({ message: `node and set name collision: ${sd.name}` });
  }
  b.ensureSet(sd.name, false);
  const s = b.sets.get(sd.name)!;
  s.declared = true;
  for (const attr of sd.attrs) {
    const v = convertAttrValue(attr.value);
    if (v && typeof v === "object" && (v as any).__brand) {
      b.errors.push({ message: `NodeRef not allowed in set attribute ${attr.key}`, line: attr.line, column: attr.col });
      continue;
    }
    s.attributes[attr.key] = v;
  }
}

function processEdgeDecl(b: GraphBuilder, ed: EdgeDecl, insideScope = false): void {
  if (ed.left.length > 1 && ed.right.length > 1) {
    b.errors.push({ message: "grouped edges on both sides", line: ed.line, column: ed.col });
    return;
  }

  let edgeLabel = ed.label;
  if (edgeLabel) {
    for (let scope: EdgeScope | undefined = b.scope; scope; scope = scope.outer) {
      if (scope.labels.has(edgeLabel)) {
        b.errors.push({ message: `duplicate edge label ${edgeLabel}`, line: ed.line, column: ed.col });
        return;
      }
    }
  }

  if (insideScope) {
    for (const attr of ed.attrs) {
      if (attr.key === "parent") {
        b.errors.push({ message: "explicit parent not allowed inside scoped edge", line: attr.line, column: attr.col });
        return;
      }
    }
  }

  let parentEdge: GslEdgeInternal | undefined;
  for (const from of ed.left) {
    for (const to of ed.right) {
      b.ensureNode(from);
      b.ensureNode(to);

      const edge: GslEdgeInternal = { parent: undefined,
        from,
        to,
        label: edgeLabel,
        attributes: {},
        sets: new Set(),
        children: [],
        scopeDepth: b.scopeDepth,
      };

      if (ed.textValue !== undefined) {
        edge.attributes["text"] = ed.textValue;
      }

      for (const attr of ed.attrs) {
        if (attr.key === "parent") {
          if (attr.value?.strVal) {
            edge.parent = attr.value.strVal;
          }
          continue;
        }
        const v = convertAttrValue(attr.value);
        if (v && typeof v === "object" && (v as any).__brand) {
          b.errors.push({ message: `NodeRef not allowed in edge attribute ${attr.key}`, line: attr.line, column: attr.col });
          continue;
        }
        edge.attributes[attr.key] = v;
      }

      if (insideScope && b.scope.parent) {
        edge.parent = b.scope.parent;
      }

      for (const setName of ed.memberships) {
        b.ensureSet(setName, !b.setDeclared(setName));
        edge.sets.add(setName);
      }

      b.edges.push(edge);
      if (!parentEdge) parentEdge = edge;
    }
  }

  if (parentEdge && edgeLabel) {
    b.scope.labels.set(edgeLabel, parentEdge);
  }

  if (ed.block.length > 0) {
    processScopedBlock(b, ed.block, edgeLabel);
  }
}

function processScopedBlock(b: GraphBuilder, stmts: Statement[], parentLabel?: string): void {
  const newScope: EdgeScope = {
    labels: new Map<string, GslEdgeInternal>(),
    parent: parentLabel,
    outer: b.scope,
  };
  const oldScope = b.scope;
  b.scope = newScope;
  b.scopeDepth++;

  for (const stmt of stmts) {
    switch (stmt.kind) {
      case "node":
        processNodeDecl(b, stmt, undefined);
        break;
      case "edge":
        processEdgeDecl(b, stmt, true);
        break;
      case "set":
        processSetDecl(b, stmt);
        break;
    }
  }

  b.scopeDepth--;
  b.scope = oldScope;
}

function populateChildren(b: GraphBuilder): void {
  const labelIndex = new Map<string, GslEdgeInternal>();
  for (const edge of b.edges) {
    if (edge.label && !labelIndex.has(edge.label)) {
      labelIndex.set(edge.label, edge);
    }
  }
  for (const edge of b.edges) {
    if (edge.parent) {
      const parent = labelIndex.get(edge.parent);
      if (parent) {
        if (parent.children) parent.children.push(edge);
      }
    }
  }
}

function validateDeps(b: GraphBuilder): void {
  const allLabels = new Set<string>();
  for (const edge of b.edges) {
    if (edge.label) allLabels.add(edge.label);
  }
  for (const edge of b.edges) {
    if (edge.parent && !allLabels.has(edge.parent)) {
      b.errors.push({ message: `edge parent references unknown label ${edge.parent}` });
    }
  }
}

function toGraph(b: GraphBuilder): GslGraph {
  const nodes: GslNode[] = [];
  for (const n of b.nodes.values()) {
    nodes.push({
      id: n.id,
      attributes: { ...n.attributes } as any,
      sets: new Set(n.sets),
      parent: n.parent ?? null,
    });
  }
  const edges: GslEdge[] = b.edges.map((e) => ({
    from: e.from,
    to: e.to,
    label: e.label ?? null,
    parent: e.parent ?? null,
    attributes: { ...e.attributes } as any,
    sets: new Set(e.sets),
    scopeDepth: e.scopeDepth,
  }));
  const sets: GslSet[] = [];
  for (const s of b.sets.values()) {
    sets.push({ id: s.id, attributes: { ...s.attributes } as any });
  }
  return { nodes, edges, sets };
}
