// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import { Ajv2020 } from "ajv/dist/2020.js";
import type { AnySchema, ValidateFunction } from "ajv";
import { parseAllDocuments } from "yaml";

import type { GslGraph } from "@gsl-lang/core";
import { expandMatchQuery, parseQuery, SubgraphExpr } from "@gsl-lang/gql";

import { selectElements, type SelectedElement } from "./match.js";
import { projectAttributes } from "./project.js";
import {
  LoadError,
  MatchingError,
  ProjectionError,
  Report,
  type Representation,
  type Violation,
} from "./result.js";

/**
 * A named definition within a dialect, describing a category of graph elements,
 * its matching rule and its attribute constraints.
 *
 * The name is a dialect-local identifier used for readability and diagnostics;
 * it does not create a new GSL element type.
 */
export class ElementDefinition {
  constructor(
    readonly name: string,
    readonly representation: Representation,
    readonly description: string,
    readonly guidance: string,
    readonly query: string,
    private readonly expr: SubgraphExpr,
    private readonly validate: ValidateFunction,
  ) {}

  /** Select the graph elements this definition applies to. */
  selectElements(graph: GslGraph): SelectedElement[] {
    return selectElements(this.representation, this.expr, graph);
  }

  /** Check one matched element's projected attributes against the schema. */
  validateAttributes(element: string, attributes: Record<string, unknown>): Violation[] {
    if (this.validate(attributes)) return [];
    return (this.validate.errors ?? []).map((e) => ({
      definition: this.name,
      representation: this.representation,
      element,
      instanceLocation: e.instancePath ?? "",
      message: e.message ?? "",
    }));
  }
}

/**
 * A validated dialect document.
 *
 * Loading is all-or-nothing: when {@link load} throws, no dialect is returned
 * and no part of the document should be applied.
 */
export class Dialect {
  private readonly byName: Map<string, ElementDefinition>;

  constructor(
    readonly id: string,
    readonly version: string,
    readonly title: string,
    readonly description: string,
    readonly guidance: string,
    private readonly elementDefinitions: readonly ElementDefinition[],
  ) {
    this.byName = new Map(elementDefinitions.map((el) => [el.name, el]));
  }

  /** The element definitions in deterministic (name-sorted) order. */
  get elements(): readonly ElementDefinition[] {
    return this.elementDefinitions;
  }

  /**
   * Evaluate every element definition's match query against `graph` and return
   * the selected elements with their projected attributes, in a deterministic
   * order (definitions by name, then nodes and sets by id, then edges in graph
   * order).
   *
   * Throws a {@link MatchingError} or {@link ProjectionError} when matching or
   * projection could not be completed; in that case the result MUST NOT be
   * treated as conforming.
   */
  match(graph: GslGraph): Selection[] {
    const out: Selection[] = [];
    for (const el of this.elementDefinitions) {
      let selected: SelectedElement[];
      try {
        selected = el.selectElements(graph);
      } catch (err) {
        throw new MatchingError(el.name, el.query, err);
      }
      for (const elem of selected) {
        let attributes: Record<string, unknown>;
        try {
          attributes = projectAttributes(elem.attributes);
        } catch (err) {
          throw new ProjectionError(el.name, elem.ref, err);
        }
        out.push({
          definition: el.name,
          representation: el.representation,
          element: elem.ref,
          attributes,
        });
      }
    }
    return out;
  }

  /**
   * Check `graph` for conformance to the dialect.
   *
   * The returned {@link Report} reports either conformance or one or more
   * violations. A thrown {@link MatchingError} or {@link ProjectionError} means
   * validation could not be completed and the graph MUST NOT be reported as
   * conforming.
   */
  validate(graph: GslGraph): Report {
    const selected = this.match(graph);
    const violations: Violation[] = [];
    for (const s of selected) {
      const el = this.byName.get(s.definition)!;
      violations.push(...el.validateAttributes(s.element, s.attributes));
    }
    return new Report(violations.length === 0 ? "conforming" : "non-conforming", violations);
  }
}

/** One graph element selected by one element definition. */
export interface Selection {
  /** The name of the element definition that selected the element. */
  readonly definition: string;

  /** The kind of the selected element. */
  readonly representation: Representation;

  /** Identifies the selected graph element. */
  readonly element: string;

  /** The projected attribute map that would be validated against the schema. */
  readonly attributes: Record<string, unknown>;
}

/**
 * Parse and validate a dialect document from YAML.
 *
 * Throws a {@link LoadError} if the document is not a valid dialect.
 */
export function load(source: string): Dialect {
  if (source.trim() === "") throw new LoadError("", "empty document");

  let docs;
  try {
    docs = parseAllDocuments(source);
  } catch (err) {
    throw new LoadError("", `not a valid YAML mapping: ${message(err)}`);
  }
  if (docs.length === 0) throw new LoadError("", "empty document");
  if (docs.length > 1) throw new LoadError("", "must be a single YAML document");

  const root = docs[0]!.toJS();
  if (root === null || typeof root !== "object" || Array.isArray(root)) {
    throw new LoadError("", "must be a mapping");
  }
  const obj = root as Record<string, unknown>;

  const id = obj["id"];
  if (typeof id !== "string" || id.trim() === "") {
    throw new LoadError("id", "required non-empty string");
  }
  const version = obj["version"];
  if (typeof version !== "string" || version.trim() === "") {
    throw new LoadError("version", "required non-empty string");
  }

  const elements = obj["elements"];
  if (elements === undefined) {
    throw new LoadError("elements", "required mapping (may be empty)");
  }
  if (elements === null || typeof elements !== "object" || Array.isArray(elements)) {
    throw new LoadError("elements", "must be a mapping");
  }

  const definitions: ElementDefinition[] = [];
  for (const [name, value] of Object.entries(elements as Record<string, unknown>)) {
    definitions.push(parseElement(name, value));
  }
  definitions.sort((a, b) => (a.name < b.name ? -1 : a.name > b.name ? 1 : 0));

  return new Dialect(
    id,
    version,
    asString(obj["title"]),
    asString(obj["description"]),
    asString(obj["guidance"]),
    definitions,
  );
}

function parseElement(name: string, value: unknown): ElementDefinition {
  const path = `elements[${name}]`;

  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    throw new LoadError(path, "must be a mapping");
  }
  const el = value as Record<string, unknown>;

  const representation = el["representation"];
  if (representation === undefined) {
    throw new LoadError(`${path}.representation`, "required");
  }
  if (representation !== "node" && representation !== "edge" && representation !== "set") {
    throw new LoadError(
      `${path}.representation`,
      `must be one of node, edge, set (got ${JSON.stringify(representation)})`,
    );
  }

  const { expr, query } = parseMatch(path, representation, el["match"]);
  const schema = compileSchema(path, el["attributes"]);

  return new ElementDefinition(
    name,
    representation,
    asString(el["description"]),
    asString(el["guidance"]),
    query,
    expr,
    schema,
  );
}

/** Parse a match query, enforcing the supported GQL subset. */
function parseMatch(
  path: string,
  representation: Representation,
  match: unknown,
): { expr: SubgraphExpr; query: string } {
  const matchPath = `${path}.match`;
  if (match === undefined) {
    throw new LoadError(matchPath, "required mapping with a query field");
  }
  if (match === null || typeof match !== "object" || Array.isArray(match)) {
    throw new LoadError(matchPath, "must be a mapping");
  }
  const query = (match as Record<string, unknown>)["query"];
  const queryPath = `${matchPath}.query`;
  if (typeof query !== "string" || query.trim() === "") {
    throw new LoadError(queryPath, "required non-empty string");
  }

  // A `representation: set` definition has no `set.` element prefix in the GQL
  // grammar, so its predicate is validated by expanding with the `node` prefix
  // but is never evaluated; set matching selects all sets.
  const prefix = representation === "set" ? "node" : representation;
  let expanded: string;
  try {
    expanded = expandMatchQuery(prefix, query);
  } catch (err) {
    throw new LoadError(queryPath, message(err));
  }

  let parsed;
  try {
    parsed = parseQuery(expanded);
  } catch (err) {
    throw new LoadError(queryPath, `invalid match query: ${message(err)}`);
  }
  if (parsed.expressions.length !== 1) {
    throw new LoadError(
      queryPath,
      "unsupported match query: only a single subgraph expression is supported",
    );
  }
  const expr = parsed.expressions[0]!;
  if (!(expr instanceof SubgraphExpr)) {
    throw new LoadError(queryPath, "unsupported match query: only the subgraph form is supported");
  }
  if (expr.traversal !== null) {
    throw new LoadError(
      queryPath,
      "unsupported match query: traversal and scope are not supported",
    );
  }
  return { expr, query };
}

/** Compile a dialect element's attribute JSON Schema, with no external loading. */
function compileSchema(path: string, attributes: unknown): ValidateFunction {
  const attrsPath = `${path}.attributes`;
  if (attributes === undefined) {
    throw new LoadError(attrsPath, "required JSON Schema");
  }
  try {
    const ajv = new Ajv2020({ strict: false, allErrors: true });
    return ajv.compile(attributes as AnySchema);
  } catch (err) {
    throw new LoadError(attrsPath, `invalid JSON Schema: ${message(err)}`);
  }
}

function asString(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function message(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}
