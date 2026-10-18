// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

/**
 * `@gsl-lang/core` — TypeScript implementation of GSL.
 */

export { parse } from "./parse.js";
export { serialize } from "./serialize.js";
export { ParseError, type ParseErrorDetail, type Diagnostic } from "./errors.js";
export { SPEC_VERSION, NodeRef, isAttributeValue, isNodeRef, makeNodeRef } from "./model.js";
export type {
  AttributeValue,
  Attributes,
  GslEdge,
  GslGraph,
  GslNode,
  GslSet,
  NodeId,
  SetName,
} from "./model.js";
