// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

export const SPEC_VERSION = "1.0.0";

/**
 * A reference to another node, produced by a bare identifier in a node
 * attribute list (e.g. `node A [friend=B]`). Distinct from a plain string at
 * runtime so serialization can emit it unquoted (`friend=B`) and so queries can
 * tell a reference from a string value.
 */
export class NodeRef {
  readonly id: string;

  constructor(id: string) {
    this.id = id;
  }

  toString(): string {
    return this.id;
  }

  toJSON(): string {
    return this.id;
  }
}

export function makeNodeRef(id: string): NodeRef {
  return new NodeRef(id);
}

export function isNodeRef(value: unknown): value is NodeRef {
  return value instanceof NodeRef;
}

export type AttributeValue = string | number | boolean | NodeRef | null | undefined;
export type NodeId = string;
export type SetName = string;
export type Attributes = Readonly<Record<string, AttributeValue>>;

export interface GslNode {
  readonly id: NodeId;
  readonly attributes: Attributes;
  readonly sets: ReadonlySet<SetName>;
  readonly parent?: NodeId | null;
}

export interface GslEdge {
  readonly from: NodeId;
  readonly to: NodeId;
  readonly label?: string | null;
  readonly parent?: string | null; // label of parent edge
  readonly attributes: Attributes;
  readonly sets: ReadonlySet<SetName>;
  readonly scopeDepth?: number;
  readonly children?: GslEdge[];
}

export interface GslSet {
  readonly id: SetName;
  readonly attributes: Attributes;
}

export interface GslGraph {
  readonly nodes: readonly GslNode[];
  readonly edges: readonly GslEdge[];
  readonly sets?: readonly GslSet[];
}

export function isAttributeValue(value: unknown): value is AttributeValue {
  return (
    typeof value === "string" ||
    typeof value === "number" ||
    typeof value === "boolean" ||
    value instanceof NodeRef
  );
}
