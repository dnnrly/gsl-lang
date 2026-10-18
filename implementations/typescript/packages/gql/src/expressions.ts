// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import type { AttributeValue, GslGraph } from "@gsl-lang/core";
import {
  AndPredicate,
  AttributeNotExistsPredicate,
  buildGlobalLabelIndex,
  NotPredicate,
  ParentPredicate,
  Predicate,
  setGlobalLabelIndex,
} from "./predicates.js";
import { GraphBuilder, WorkEdge, WorkNode, toWorkGraph } from "./graph.js";

export interface Value {
  graph: GslGraph;
}

export interface QueryContext {
  inputGraph: GslGraph;
  namedGraphs: Map<string, GslGraph>;
}

export interface Expression {
  apply(ctx: QueryContext, input: Value): Value;
}

function cloneNode(node: WorkNode): WorkNode {
  return {
    id: node.id,
    attributes: { ...node.attributes },
    sets: new Set(node.sets),
    parent: node.parent,
  };
}

function cloneEdge(edge: WorkEdge): WorkEdge {
  return {
    from: edge.from,
    to: edge.to,
    label: edge.label,
    parent: edge.parent,
    attributes: { ...edge.attributes },
    sets: new Set(edge.sets),
    scopeDepth: edge.scopeDepth,
  };
}

export function validateTargetType(pred: Predicate): string {
  const t = pred.targetType();
  if (t === "error") throw new Error("predicate mixes node and edge targets");
  return t;
}

export class Query {
  constructor(readonly expressions: Expression[]) {}

  execute(ctx: QueryContext): Value {
    let value: Value = { graph: ctx.inputGraph };
    for (const expr of this.expressions) {
      value = expr.apply(ctx, value);
    }
    return value;
  }
}

export class IdentityExpr implements Expression {
  apply(_ctx: QueryContext, input: Value): Value {
    return input;
  }
}

export class FromExpr implements Expression {
  constructor(
    private readonly isWildcard: boolean,
    private readonly name: string,
  ) {}

  apply(ctx: QueryContext): Value {
    if (this.isWildcard) return { graph: ctx.inputGraph };
    const graph = ctx.namedGraphs.get(this.name);
    if (!graph) throw new Error(`named graph not found: ${this.name}`);
    return { graph };
  }
}

export class BindExpr implements Expression {
  constructor(
    private readonly pipeline: Query,
    private readonly name: string,
  ) {}

  apply(ctx: QueryContext, input: Value): Value {
    if (ctx.namedGraphs.has(this.name)) {
      throw new Error(`named graph already bound: ${this.name}`);
    }
    const result = this.pipeline.execute(ctx);
    ctx.namedGraphs.set(this.name, result.graph);
    return input;
  }
}

export interface TraversalConfig {
  directions: string[];
  depth: number;
}

export class SubgraphExpr implements Expression {
  constructor(
    private readonly pred: Predicate,
    private readonly traversal: TraversalConfig | null,
  ) {}

  apply(_ctx: QueryContext, input: Value): Value {
    const graph = input.graph;
    const targetType = validateTargetType(this.pred);

    const work = toWorkGraph(graph);
    const nodes = work.nodes;
    const edges = work.edges;

    buildGlobalLabelIndex(edges);

    const { nodeSet, edgeSet } = this.buildSubgraph(nodes, edges, targetType);

    let baseNodes = nodeSet;
    let baseEdges = edgeSet;
    if (this.traversal && this.traversal.depth > 0) {
      baseNodes = this.traverse(edges, baseNodes, this.traversal);
      const rebuilt = new Set<number>();
      for (let i = 0; i < edges.length; i++) {
        const edge = edges[i]!;
        if (baseNodes.has(edge.from) && baseNodes.has(edge.to)) rebuilt.add(i);
      }
      baseEdges = rebuilt;
    }

    setGlobalLabelIndex(null);

    const result = new GraphBuilder();
    for (const set of work.sets.values()) {
      result.addSet({ id: set.id, attributes: { ...set.attributes } });
    }
    for (const id of baseNodes) {
      const node = nodes.get(id);
      if (node) result.addNode(cloneNode(node));
    }
    if (targetType === "edge") {
      const indices = Array.from(baseEdges).sort((a, b) => a - b);
      for (const idx of indices) {
        const edge = edges[idx];
        if (edge) result.addEdge(cloneEdge(edge));
      }
    } else {
      for (const edge of edges) {
        if (baseNodes.has(edge.from) && baseNodes.has(edge.to)) {
          result.addEdge(cloneEdge(edge));
        }
      }
    }
    return { graph: result.finalize() };
  }

  private buildSubgraph(
    nodes: Map<string, WorkNode>,
    edges: WorkEdge[],
    targetType: string,
  ): { nodeSet: Set<string>; edgeSet: Set<number> } {
    const nodeSet = new Set<string>();
    const edgeSet = new Set<number>();

    if (targetType === "node") {
      for (const [id, node] of nodes) {
        if (this.pred.evaluateNode(node)) nodeSet.add(id);
      }
    } else if (targetType === "edge") {
      this.injectEdges(edges);
      for (let i = 0; i < edges.length; i++) {
        const edge = edges[i]!;
        if (this.pred.evaluateEdge(edge)) {
          nodeSet.add(edge.from);
          nodeSet.add(edge.to);
          edgeSet.add(i);
        }
      }
    } else {
      for (const [id, node] of nodes) {
        if (this.pred.evaluateNode(node)) nodeSet.add(id);
      }
      if (nodeSet.size === 0) {
        this.injectEdges(edges);
        for (let i = 0; i < edges.length; i++) {
          const edge = edges[i]!;
          if (this.pred.evaluateEdge(edge)) {
            nodeSet.add(edge.from);
            nodeSet.add(edge.to);
            edgeSet.add(i);
          }
        }
      }
      if (nodeSet.size === 0) {
        for (const id of nodes.keys()) nodeSet.add(id);
      }
    }

    return { nodeSet, edgeSet };
  }

  private injectEdges(edges: WorkEdge[]): void {
    if (this.pred instanceof ParentPredicate) {
      this.pred.allEdges = edges;
    }
  }

  private traverse(edges: readonly WorkEdge[], startNodes: Set<string>, cfg: TraversalConfig): Set<string> {
    const result = new Set<string>(startNodes);
    const visited = new Set<string>(startNodes);
    let frontier = Array.from(startNodes);

    for (let depth = 0; depth < cfg.depth && frontier.length > 0; depth++) {
      const nextFrontier: string[] = [];
      for (const nodeId of frontier) {
        for (const neighbor of getNeighbors(edges, nodeId, cfg.directions)) {
          if (!visited.has(neighbor)) {
            visited.add(neighbor);
            result.add(neighbor);
            nextFrontier.push(neighbor);
          }
        }
      }
      frontier = nextFrontier;
    }

    if (hasDependencyDirection(cfg.directions)) {
      traverseDependencies(edges, result, cfg);
    }

    return result;
  }
}

function getNeighbors(edges: readonly WorkEdge[], nodeId: string, directions: string[]): string[] {
  const neighbors = new Set<string>();
  for (const direction of directions) {
    if (direction === "out") {
      for (const edge of edges) if (edge.from === nodeId) neighbors.add(edge.to);
    } else if (direction === "in") {
      for (const edge of edges) if (edge.to === nodeId) neighbors.add(edge.from);
    } else if (direction === "both") {
      for (const edge of edges) {
        if (edge.from === nodeId) neighbors.add(edge.to);
        if (edge.to === nodeId) neighbors.add(edge.from);
      }
    }
  }
  return Array.from(neighbors);
}

function hasDependencyDirection(dirs: string[]): boolean {
  return dirs.some((d) => d === "up" || d === "down");
}

function traverseDependencies(
  edges: readonly WorkEdge[],
  result: Set<string>,
  cfg: TraversalConfig,
): void {
  const labelIndex = new Map<string, WorkEdge>();
  const childIndex = new Map<string, WorkEdge[]>();
  for (const edge of edges) {
    if (edge.label) {
      if (!labelIndex.has(edge.label)) labelIndex.set(edge.label, edge);
    }
    if (edge.parent) {
      if (!childIndex.has(edge.parent)) childIndex.set(edge.parent, []);
      childIndex.get(edge.parent)!.push(edge);
    }
  }

  setGlobalLabelIndex(labelIndex);

  let edgeQueue: WorkEdge[] = [];
  const edgeVisited = new Set<WorkEdge>();
  for (const edge of edges) {
    if (result.has(edge.from) || result.has(edge.to)) {
      if (!edgeVisited.has(edge)) {
        edgeVisited.add(edge);
        edgeQueue.push(edge);
      }
      result.add(edge.from);
      result.add(edge.to);
    }
  }

  for (const dir of cfg.directions) {
    if (dir === "up") {
      for (let depth = 0; depth < cfg.depth && edgeQueue.length > 0; depth++) {
        const nextQueue: WorkEdge[] = [];
        for (const edge of edgeQueue) {
          if (edge.parent) {
            const parent = labelIndex.get(edge.parent);
            if (parent) {
              result.add(parent.from);
              result.add(parent.to);
              if (!edgeVisited.has(parent)) {
                edgeVisited.add(parent);
                nextQueue.push(parent);
              }
            }
          }
        }
        edgeQueue = nextQueue;
      }
    } else if (dir === "down") {
      for (let depth = 0; depth < cfg.depth && edgeQueue.length > 0; depth++) {
        const nextQueue: WorkEdge[] = [];
        for (const edge of edgeQueue) {
          for (const child of childIndex.get(edge.label ?? "") ?? []) {
            result.add(child.from);
            result.add(child.to);
            if (!edgeVisited.has(child)) {
              edgeVisited.add(child);
              nextQueue.push(child);
            }
          }
        }
        edgeQueue = nextQueue;
      }
    }
  }
}

export class RemoveEdgeExpr implements Expression {
  constructor(private readonly pred: Predicate) {}

  apply(_ctx: QueryContext, input: Value): Value {
    const graph = input.graph;
    const targetType = validateTargetType(this.pred);
    const work = toWorkGraph(graph);
    const result = new GraphBuilder();
    for (const node of work.nodes.values()) result.addNode(cloneNode(node));
    for (const set of work.sets.values()) {
      result.addSet({ id: set.id, attributes: { ...set.attributes } });
    }
    for (const edge of work.edges) {
      if (targetType === "edge" || targetType === "") {
        if (!this.pred.evaluateEdge(edge)) result.addEdge(cloneEdge(edge));
      } else {
        result.addEdge(cloneEdge(edge));
      }
    }
    return { graph: result.finalize() };
  }
}

export class RemoveAttributeExpr implements Expression {
  constructor(
    private readonly target: string,
    private readonly attr: string,
    private readonly pred: Predicate,
  ) {}

  apply(_ctx: QueryContext, input: Value): Value {
    const graph = input.graph;
    const targetType = validateTargetType(this.pred);
    const work = toWorkGraph(graph);
    const result = new GraphBuilder();
    for (const set of work.sets.values()) {
      result.addSet({ id: set.id, attributes: { ...set.attributes } });
    }
    for (const node of work.nodes.values()) {
      const copy = cloneNode(node);
      if (this.target === "node" && (targetType === "" || targetType === "node")) {
        if (this.pred.evaluateNode(node)) delete copy.attributes[this.attr];
      }
      result.addNode(copy);
    }
    for (const edge of work.edges) {
      const copy = cloneEdge(edge);
      if (this.target === "edge" && (targetType === "" || targetType === "edge")) {
        if (this.pred.evaluateEdge(edge)) delete copy.attributes[this.attr];
      }
      result.addEdge(copy);
    }
    return { graph: result.finalize() };
  }
}

export class RemoveOrphansExpr implements Expression {
  apply(_ctx: QueryContext, input: Value): Value {
    const graph = input.graph;
    const work = toWorkGraph(graph);
    const result = new GraphBuilder();
    for (const set of work.sets.values()) {
      result.addSet({ id: set.id, attributes: { ...set.attributes } });
    }
    const incident = new Set<string>();
    for (const edge of work.edges) {
      incident.add(edge.from);
      incident.add(edge.to);
    }
    for (const node of work.nodes.values()) {
      if (incident.has(node.id)) result.addNode(cloneNode(node));
    }
    for (const edge of work.edges) result.addEdge(cloneEdge(edge));
    return { graph: result.finalize() };
  }
}

export class CollapseExpr implements Expression {
  constructor(
    private readonly nodeId: string,
    private readonly pred: Predicate,
  ) {}

  apply(_ctx: QueryContext, input: Value): Value {
    const graph = input.graph;
    const targetType = validateTargetType(this.pred);
    if (targetType === "edge") {
      throw new Error("collapse only works with node predicates");
    }
    const work = toWorkGraph(graph);
    const collapsed = new Set<string>();
    for (const [id, node] of work.nodes) {
      if (this.pred.evaluateNode(node)) collapsed.add(id);
    }
    if (collapsed.size === 0) return { graph };

    const result = new GraphBuilder();
    for (const set of work.sets.values()) {
      result.addSet({ id: set.id, attributes: { ...set.attributes } });
    }
    for (const node of work.nodes.values()) {
      if (!collapsed.has(node.id)) result.addNode(cloneNode(node));
    }

    const merged: Record<string, AttributeValue> = {};
    for (const id of Array.from(collapsed).sort()) {
      const node = work.nodes.get(id);
      if (node) {
        for (const k of Object.keys(node.attributes)) merged[k] = node.attributes[k]!;
      }
    }
    result.addNode({ id: this.nodeId, attributes: merged, sets: new Set(), parent: null });

    const seen = new Set<string>();
    for (const edge of work.edges) {
      const fromCollapsed = collapsed.has(edge.from);
      const toCollapsed = collapsed.has(edge.to);
      if (fromCollapsed && toCollapsed) continue;
      const newEdge: WorkEdge = {
        from: fromCollapsed ? this.nodeId : edge.from,
        to: toCollapsed ? this.nodeId : edge.to,
        label: null,
        parent: null,
        attributes: { ...edge.attributes },
        sets: new Set(edge.sets),
        scopeDepth: 0,
      };
      const key = edgeKey(newEdge);
      if (!seen.has(key)) {
        seen.add(key);
        result.addEdge(newEdge);
      }
    }
    return { graph: result.finalize() };
  }
}

function edgeKey(edge: WorkEdge): string {
  const keys = Object.keys(edge.attributes).sort();
  let attrs = "";
  for (const k of keys) {
    if (attrs !== "") attrs += ",";
    attrs += `${k}:${String(edge.attributes[k])}`;
  }
  return `${edge.from}|${edge.to}|${attrs}`;
}

export class MakeExpr implements Expression {
  constructor(
    private readonly target: string,
    private readonly attr: string,
    private readonly value: AttributeValue,
    private readonly pred: Predicate,
  ) {}

  apply(_ctx: QueryContext, input: Value): Value {
    const graph = input.graph;
    const targetType = validateTargetType(this.pred);
    const work = toWorkGraph(graph);
    const result = new GraphBuilder();
    for (const set of work.sets.values()) {
      result.addSet({ id: set.id, attributes: { ...set.attributes } });
    }
    for (const node of work.nodes.values()) {
      const copy = cloneNode(node);
      if (this.target === "node" && (targetType === "" || targetType === "node")) {
        if (this.pred.evaluateNode(node)) copy.attributes[this.attr] = this.value;
      }
      result.addNode(copy);
    }
    for (const edge of work.edges) {
      const copy = cloneEdge(edge);
      if (this.target === "edge" && (targetType === "" || targetType === "edge")) {
        if (this.pred.evaluateEdge(edge)) copy.attributes[this.attr] = this.value;
      }
      result.addEdge(copy);
    }
    return { graph: result.finalize() };
  }
}

export class GraphAlgebraExpr implements Expression {
  constructor(
    private readonly leftRef: string,
    private readonly operator: string,
    private readonly rightRef: string,
  ) {}

  apply(ctx: QueryContext): Value {
    const left = this.resolveGraph(ctx, this.leftRef);
    const right = this.resolveGraph(ctx, this.rightRef);
    switch (this.operator) {
      case "+":
        return { graph: this.union(left, right) };
      case "&":
        return { graph: this.intersection(left, right) };
      case "-":
        return { graph: this.difference(left, right) };
      case "^":
        return { graph: this.symmetricDifference(left, right) };
      default:
        throw new Error(`unknown algebra operator: ${this.operator}`);
    }
  }

  private resolveGraph(ctx: QueryContext, ref: string): GslGraph {
    if (ref === "*") return ctx.inputGraph;
    const graph = ctx.namedGraphs.get(ref);
    if (!graph) throw new Error(`named graph not found: ${ref}`);
    return graph;
  }

  private union(left: GslGraph, right: GslGraph): GslGraph {
    const lb = toWorkGraph(left);
    const rb = toWorkGraph(right);
    const result = new GraphBuilder();
    for (const node of lb.nodes.values()) result.addNode(cloneNode(node));
    for (const edge of lb.edges) result.addEdge(cloneEdge(edge));
    for (const set of lb.sets.values()) {
      result.addSet({ id: set.id, attributes: { ...set.attributes } });
    }
    for (const node of rb.nodes.values()) {
      const existing = result.nodes.get(node.id);
      if (existing) {
        for (const k of Object.keys(node.attributes)) existing.attributes[k] = node.attributes[k]!;
        for (const s of node.sets) existing.sets.add(s);
      } else {
        result.addNode(cloneNode(node));
      }
    }
    for (const edge of rb.edges) result.addEdge(cloneEdge(edge));
    for (const set of rb.sets.values()) {
      const existing = result.sets.get(set.id);
      if (existing) {
        for (const k of Object.keys(set.attributes)) existing.attributes[k] = set.attributes[k]!;
      } else {
        result.addSet({ id: set.id, attributes: { ...set.attributes } });
      }
    }
    return result.finalize();
  }

  private intersection(left: GslGraph, right: GslGraph): GslGraph {
    const lb = toWorkGraph(left);
    const rb = toWorkGraph(right);
    const result = new GraphBuilder();
    const resultNodes = new Set<string>();
    for (const [id, node] of lb.nodes) {
      if (rb.nodes.has(id)) {
        result.addNode(cloneNode(node));
        resultNodes.add(id);
      }
    }
    const rightEdgeSet = new Set<string>();
    for (const edge of rb.edges) rightEdgeSet.add(`${edge.from}|${edge.to}`);
    for (const edge of lb.edges) {
      if (
        rightEdgeSet.has(`${edge.from}|${edge.to}`) &&
        resultNodes.has(edge.from) &&
        resultNodes.has(edge.to)
      ) {
        result.addEdge(cloneEdge(edge));
      }
    }
    for (const [id, set] of lb.sets) {
      if (rb.sets.has(id)) result.addSet({ id: set.id, attributes: { ...set.attributes } });
    }
    return result.finalize();
  }

  private difference(left: GslGraph, right: GslGraph): GslGraph {
    const lb = toWorkGraph(left);
    const rb = toWorkGraph(right);
    const result = new GraphBuilder();
    const resultNodes = new Set<string>();
    for (const [id, node] of lb.nodes) {
      if (!rb.nodes.has(id)) {
        result.addNode(cloneNode(node));
        resultNodes.add(id);
      }
    }
    const rightEdgeSet = new Set<string>();
    for (const edge of rb.edges) rightEdgeSet.add(`${edge.from}|${edge.to}`);
    for (const edge of lb.edges) {
      if (!rightEdgeSet.has(`${edge.from}|${edge.to}`) && resultNodes.has(edge.from) && resultNodes.has(edge.to)) {
        result.addEdge(cloneEdge(edge));
      }
    }
    for (const [id, set] of lb.sets) {
      if (!rb.sets.has(id)) result.addSet({ id: set.id, attributes: { ...set.attributes } });
    }
    return result.finalize();
  }

  private symmetricDifference(left: GslGraph, right: GslGraph): GslGraph {
    const lb = toWorkGraph(left);
    const rb = toWorkGraph(right);
    const result = new GraphBuilder();
    const resultNodes = new Set<string>();
    for (const [id, node] of lb.nodes) {
      if (!rb.nodes.has(id)) {
        result.addNode(cloneNode(node));
        resultNodes.add(id);
      }
    }
    for (const [id, node] of rb.nodes) {
      if (!lb.nodes.has(id)) {
        result.addNode(cloneNode(node));
        resultNodes.add(id);
      }
    }
    const leftEdgeSet = new Set<string>();
    const rightEdgeSet = new Set<string>();
    for (const edge of lb.edges) leftEdgeSet.add(`${edge.from}|${edge.to}`);
    for (const edge of rb.edges) rightEdgeSet.add(`${edge.from}|${edge.to}`);
    for (const edge of lb.edges) {
      if (!rightEdgeSet.has(`${edge.from}|${edge.to}`) && resultNodes.has(edge.from) && resultNodes.has(edge.to)) {
        result.addEdge(cloneEdge(edge));
      }
    }
    for (const edge of rb.edges) {
      if (!leftEdgeSet.has(`${edge.from}|${edge.to}`) && resultNodes.has(edge.from) && resultNodes.has(edge.to)) {
        result.addEdge(cloneEdge(edge));
      }
    }
    for (const [id, set] of lb.sets) {
      if (!rb.sets.has(id)) result.addSet({ id: set.id, attributes: { ...set.attributes } });
    }
    for (const [id, set] of rb.sets) {
      if (!lb.sets.has(id)) result.addSet({ id: set.id, attributes: { ...set.attributes } });
    }
    return result.finalize();
  }
}
