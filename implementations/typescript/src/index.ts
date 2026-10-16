// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

/**
 * `@gsl-lang/core` — a work-in-progress TypeScript implementation of GSL.
 *
 * Status: model types only. Parsing, serialization and query execution are
 * not implemented yet. See `README.md`.
 */

export {
  SPEC_VERSION,
  isAttributeValue,
  type AttributeValue,
  type Attributes,
  type GslEdge,
  type GslGraph,
  type GslNode,
  type NodeId,
  type SetName,
} from "./model.js";
