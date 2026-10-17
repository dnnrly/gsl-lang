// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import type { AttributeValue } from "@gsl-lang/core";
import type { WorkEdge, WorkNode } from "./graph.js";

export type TargetType = "" | "node" | "edge" | "error";
export type QueryValue = string | number | boolean;

export interface Predicate {
  evaluateNode(node: WorkNode): boolean;
  evaluateEdge(edge: WorkEdge): boolean;
  targetType(): TargetType;
}

let globalLabelIndex: Map<string, WorkEdge> | null = null;

export function setGlobalLabelIndex(index: Map<string, WorkEdge> | null): void {
  globalLabelIndex = index;
}

export function buildGlobalLabelIndex(edges: readonly WorkEdge[]): void {
  const index = new Map<string, WorkEdge>();
  for (const edge of edges) {
    if (edge.label) {
      if (!index.has(edge.label)) index.set(edge.label, edge);
    }
  }
  globalLabelIndex = index;
}

export class ExistsPredicate implements Predicate {
  evaluateNode(): boolean {
    return true;
  }
  evaluateEdge(): boolean {
    return true;
  }
  targetType(): TargetType {
    return "";
  }
}

export class AttributeEqualsPredicate implements Predicate {
  constructor(
    private readonly target: string,
    private readonly name: string,
    private readonly value: QueryValue,
  ) {}

  evaluateNode(node: WorkNode): boolean {
    if (this.name === "id") return node.id === this.value;
    const val = node.attributes[this.name];
    if (val === undefined) return false;
    return val === this.value;
  }

  evaluateEdge(edge: WorkEdge): boolean {
    const val = edge.attributes[this.name];
    if (val === undefined) return false;
    return val === this.value;
  }

  targetType(): TargetType {
    return this.target as TargetType;
  }
}

export class AttributeNotEqualsPredicate implements Predicate {
  constructor(
    private readonly target: string,
    private readonly name: string,
    private readonly value: QueryValue,
  ) {}

  evaluateNode(node: WorkNode): boolean {
    if (this.name === "id") return node.id !== this.value;
    const val = node.attributes[this.name];
    if (val === undefined) return false;
    return val !== this.value;
  }

  evaluateEdge(edge: WorkEdge): boolean {
    const val = edge.attributes[this.name];
    if (val === undefined) return false;
    return val !== this.value;
  }

  targetType(): TargetType {
    return this.target as TargetType;
  }
}

export class AttributeExistsPredicate implements Predicate {
  constructor(
    private readonly target: string,
    private readonly name: string,
  ) {}

  evaluateNode(node: WorkNode): boolean {
    return node.attributes[this.name] !== undefined;
  }

  evaluateEdge(edge: WorkEdge): boolean {
    return edge.attributes[this.name] !== undefined;
  }

  targetType(): TargetType {
    return this.target as TargetType;
  }
}

export class AttributeNotExistsPredicate implements Predicate {
  constructor(
    private readonly target: string,
    private readonly name: string,
  ) {}

  evaluateNode(node: WorkNode): boolean {
    return node.attributes[this.name] === undefined;
  }

  evaluateEdge(edge: WorkEdge): boolean {
    return edge.attributes[this.name] === undefined;
  }

  targetType(): TargetType {
    return this.target as TargetType;
  }
}

export class SetMembershipPredicate implements Predicate {
  constructor(
    private readonly target: string,
    private readonly setId: string,
  ) {}

  evaluateNode(node: WorkNode): boolean {
    return node.sets.has(this.setId);
  }

  evaluateEdge(edge: WorkEdge): boolean {
    return edge.sets.has(this.setId);
  }

  targetType(): TargetType {
    return this.target as TargetType;
  }
}

export class SetNotMembershipPredicate implements Predicate {
  constructor(
    private readonly target: string,
    private readonly setId: string,
  ) {}

  evaluateNode(node: WorkNode): boolean {
    return !node.sets.has(this.setId);
  }

  evaluateEdge(edge: WorkEdge): boolean {
    return !edge.sets.has(this.setId);
  }

  targetType(): TargetType {
    return this.target as TargetType;
  }
}

export class ParentExistsPredicate implements Predicate {
  constructor(private readonly target: string) {}

  evaluateNode(): boolean {
    return false;
  }
  evaluateEdge(edge: WorkEdge): boolean {
    return edge.parent !== null && edge.parent !== "";
  }
  targetType(): TargetType {
    return this.target as TargetType;
  }
}

export class ParentNotExistsPredicate implements Predicate {
  constructor(private readonly target: string) {}

  evaluateNode(): boolean {
    return false;
  }
  evaluateEdge(edge: WorkEdge): boolean {
    return edge.parent === null || edge.parent === "";
  }
  targetType(): TargetType {
    return this.target as TargetType;
  }
}

export class DepthPredicate implements Predicate {
  constructor(
    private readonly operator: string,
    private readonly value: number,
  ) {}

  evaluateNode(): boolean {
    return false;
  }

  evaluateEdge(edge: WorkEdge): boolean {
    const depth = this.computeDepth(edge);
    if (this.operator === "==") return depth === this.value;
    if (this.operator === "!=") return depth !== this.value;
    return false;
  }

  private computeDepth(edge: WorkEdge): number {
    if (!edge.parent) return 0;
    if (globalLabelIndex !== null) {
      return this.walkDepth(edge, 0, new Set<string>());
    }
    return 1;
  }

  private walkDepth(edge: WorkEdge, depth: number, visited: Set<string>): number {
    if (!edge.parent) return depth;
    if (edge.label) {
      if (visited.has(edge.label)) return depth;
      visited.add(edge.label);
    }
    const parent = globalLabelIndex?.get(edge.parent);
    if (!parent) return depth + 1;
    return this.walkDepth(parent, depth + 1, visited);
  }

  targetType(): TargetType {
    return "edge";
  }
}

export class ParentPredicate implements Predicate {
  allEdges: readonly WorkEdge[] | null = null;

  constructor(
    private readonly target: string,
    private readonly inner: Predicate,
  ) {}

  evaluateNode(): boolean {
    return false;
  }

  evaluateEdge(edge: WorkEdge): boolean {
    if (!edge.parent) return false;
    const parent = this.findParent(edge);
    if (!parent) return false;
    return this.inner.evaluateEdge(parent);
  }

  private findParent(edge: WorkEdge): WorkEdge | null {
    if (!this.allEdges) return null;
    for (const e of this.allEdges) {
      if (e.label === edge.parent) return e;
    }
    return null;
  }

  targetType(): TargetType {
    return this.target as TargetType;
  }
}

export class AndPredicate implements Predicate {
  constructor(
    private readonly left: Predicate,
    private readonly right: Predicate,
  ) {}

  evaluateNode(node: WorkNode): boolean {
    return this.left.evaluateNode(node) && this.right.evaluateNode(node);
  }

  evaluateEdge(edge: WorkEdge): boolean {
    return this.left.evaluateEdge(edge) && this.right.evaluateEdge(edge);
  }

  targetType(): TargetType {
    const left = this.left.targetType();
    const right = this.right.targetType();
    if (left === "") return right;
    if (right === "") return left;
    if (left !== right) return "error";
    return left;
  }
}

export class NotPredicate implements Predicate {
  constructor(private readonly inner: Predicate) {}

  evaluateNode(node: WorkNode): boolean {
    return !this.inner.evaluateNode(node);
  }
  evaluateEdge(edge: WorkEdge): boolean {
    return !this.inner.evaluateEdge(edge);
  }
  targetType(): TargetType {
    return this.inner.targetType();
  }
}

export function isValidGraphName(name: string): boolean {
  if (name.length === 0) return false;
  if (name[0]! < "A" || name[0]! > "Z") return false;
  for (let i = 1; i < name.length; i++) {
    const ch = name[i]!;
    const ok =
      (ch >= "A" && ch <= "Z") ||
      (ch >= "0" && ch <= "9") ||
      ch === "_";
    if (!ok) return false;
  }
  return true;
}

export function isValidGraphRef(ref: string): boolean {
  return ref === "*" || isValidGraphName(ref);
}
