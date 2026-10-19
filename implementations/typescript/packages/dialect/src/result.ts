// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

/**
 * Result types for dialect conformance validation.
 *
 * The normative behaviour is defined by `spec/dialects/DIALECTS_SPEC.md`.
 */

/** The kind of GSL element an element definition applies to. */
export type Representation = "node" | "edge" | "set";

/** The overall conformance outcome of a validation. */
export type Outcome = "conforming" | "non-conforming";

/** A single attribute constraint violation. */
export interface Violation {
  /** The name of the element definition that was violated. */
  readonly definition: string;

  /** The kind of the offending element. */
  readonly representation: Representation;

  /**
   * Identifies the offending graph element: a node id, a set id, or an edge
   * descriptor such as `a->b`.
   */
  readonly element: string;

  /**
   * The JSON Pointer path within the projected attribute map (for example
   * `/replicas`). Empty for a violation at the root.
   */
  readonly instanceLocation: string;

  /** A human-readable explanation of the violation. */
  readonly message: string;
}

/**
 * The result of validating a graph against a dialect. A report is produced only
 * when validation completed; an incomplete validation is reported by throwing a
 * {@link MatchingError} or {@link ProjectionError} instead.
 */
export class Report {
  constructor(
    readonly outcome: Outcome,
    readonly violations: readonly Violation[],
  ) {}

  /** Whether the graph conforms to the dialect. */
  get conforms(): boolean {
    return this.outcome === "conforming";
  }
}

/** Base class for dialect errors. */
export class DialectError extends Error {}

/**
 * A dialect document that is invalid and was rejected before graph validation
 * began. `path` locates the offending field, for example
 * `elements[Service].match`.
 */
export class LoadError extends DialectError {
  constructor(
    readonly path: string,
    message: string,
  ) {
    super(path === "" ? `invalid dialect: ${message}` : `invalid dialect: ${path}: ${message}`);
    this.name = "LoadError";
  }
}

/**
 * A failure to evaluate a match query, indicating incomplete validation rather
 * than a conformance violation.
 */
export class MatchingError extends DialectError {
  constructor(
    readonly definition: string,
    readonly query: string,
    override readonly cause: unknown,
  ) {
    super(
      `matching error in definition ${JSON.stringify(definition)} (query ${JSON.stringify(query)}): ${cause instanceof Error ? cause.message : String(cause)}`,
    );
    this.name = "MatchingError";
  }
}

/**
 * A failure to project an element's attribute values into JSON, indicating
 * incomplete validation.
 */
export class ProjectionError extends DialectError {
  constructor(
    readonly definition: string,
    readonly element: string,
    override readonly cause: unknown,
  ) {
    super(
      `projection error for ${element} in definition ${JSON.stringify(definition)}: ${cause instanceof Error ? cause.message : String(cause)}`,
    );
    this.name = "ProjectionError";
  }
}
