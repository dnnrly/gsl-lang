// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

/**
 * `@gsl-lang/dialect` — TypeScript implementation of GSL dialect support:
 * loading and validating dialect documents, matching graph elements with a
 * restricted GQL subset, projecting GSL attribute values into JSON, and
 * checking the conformance of a graph against JSON Schema attribute
 * constraints.
 *
 * The normative behaviour is defined by `spec/dialects/DIALECTS_SPEC.md`.
 */

export { load, Dialect, ElementDefinition, type Selection } from "./document.js";
export { projectAttributes } from "./project.js";
export { selectElements, type SelectedElement } from "./match.js";
export {
  Report,
  DialectError,
  LoadError,
  MatchingError,
  ProjectionError,
} from "./result.js";
export type { Outcome, Representation, Violation } from "./result.js";
