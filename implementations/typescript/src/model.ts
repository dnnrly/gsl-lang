// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

/**
 * Core, implementation-neutral GSL model types.
 *
 * This module contains only the shapes described by the normative
 * specification (see `spec/gsl/SPEC.md`). It deliberately contains no parser:
 * parsing, serialization and query execution are out of scope for this
 * workspace and will be added by a later phase.
 */

/** The GSL specification version this workspace targets. */
export const SPEC_VERSION = "1.0.0";

/**
 * A GSL attribute value. Attributes are untyped in GSL; the reference
 * serialization distinguishes strings, numbers and booleans.
 */
export type AttributeValue = string | number | boolean;

/** A node identifier. In GSL identifiers are bare words. */
export type NodeId = string;

/** A named collection of nodes and edges (see the specification's sets). */
export type SetName = string;

/** Attribute storage: name to value. Absent names are simply not present. */
export type Attributes = Readonly<Record<string, AttributeValue>>;

/** A graph node: an identifier, its attributes, and its set membership. */
export interface GslNode {
  readonly id: NodeId;
  readonly attributes: Attributes;
  readonly sets: ReadonlySet<SetName>;
}

/**
 * A directed graph edge. `parent`, when present, records the edge this one
 * depends on (the edge-dependency hierarchy).
 */
export interface GslEdge {
  readonly from: NodeId;
  readonly to: NodeId;
  readonly attributes: Attributes;
  readonly sets: ReadonlySet<SetName>;
  readonly parent?: GslEdge;
}

/** A materialised graph: nodes and directed edges. */
export interface GslGraph {
  readonly nodes: readonly GslNode[];
  readonly edges: readonly GslEdge[];
}

/** True when `value` is a legal GSL attribute value. */
export function isAttributeValue(value: unknown): value is AttributeValue {
  return (
    typeof value === "string" ||
    typeof value === "number" ||
    typeof value === "boolean"
  );
}
