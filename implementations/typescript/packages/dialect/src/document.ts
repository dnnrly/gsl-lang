// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import { Ajv2020 } from "ajv/dist/2020.js";
import type { AnySchema, ValidateFunction } from "ajv";
import { parseAllDocuments } from "yaml";

import type { GslGraph, GslSet } from "@gsl-lang/core";
import { expandMatchQuery, parseQuery, SubgraphExpr } from "@gsl-lang/gql";

import { selectElements, type SelectedElement } from "./match.js";
import { projectAttributes } from "./project.js";
import {
  LoadError,
  MatchingError,
  ProjectionError,
  Report,
  defaultSetPolicy,
  type Representation,
  type SetPolicy,
  type Violation,
} from "./result.js";

/**
 * A node or edge rule.
 *
 * The name is an optional dialect-local label used for readability and
 * diagnostics; it does not create a new GSL element type. Unnamed rules are
 * identified by their section position, for example `nodes.rules[0]`.
 */
export class Rule {
  constructor(
    readonly name: string,
    readonly representation: "node" | "edge",
    readonly index: number,
    readonly description: string,
    readonly guidance: string,
    readonly query: string,
    private readonly expr: SubgraphExpr,
    readonly validate: ValidateFunction,
  ) {}

  /** The diagnostic label: the name when present, otherwise the section path. */
  get label(): string {
    if (this.name !== "") return this.name;
    const section = this.representation === "edge" ? "edges.rules" : "nodes.rules";
    return `${section}[${this.index}]`;
  }

  /** Select the graph elements this rule applies to. */
  selectElements(graph: GslGraph): SelectedElement[] {
    return selectElements(this.representation, this.expr, graph);
  }

  /** Check one matched element's projected attributes against the schema. */
  validateAttributes(element: string, attributes: Record<string, unknown>): Violation[] {
    if (this.validate(attributes)) return [];
    return (this.validate.errors ?? []).map((e) => ({
      definition: this.label,
      representation: this.representation,
      element,
      instanceLocation: e.instancePath ?? "",
      message: e.message ?? "",
    }));
  }
}

/** A set definition keyed by its GSL set name. */
export class SetDefinition {
  constructor(
    readonly name: string,
    readonly description: string,
    readonly guidance: string,
    readonly validate?: ValidateFunction,
  ) {}

  /** Check one present set's projected attributes against the schema, if any. */
  validateAttributes(attributes: Record<string, unknown>): Violation[] {
    if (this.validate === undefined) return [];
    if (this.validate(attributes)) return [];
    return (this.validate.errors ?? []).map((e) => ({
      definition: this.name,
      representation: "set" as const,
      element: this.name,
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
  constructor(
    readonly id: string,
    readonly version: string,
    readonly title: string,
    readonly description: string,
    readonly guidance: string,
    private readonly nodeRuleList: readonly Rule[],
    private readonly edgeRuleList: readonly Rule[],
    readonly setPolicy: SetPolicy,
    private readonly setDefinitionList: readonly SetDefinition[],
  ) {}

  /** Node rules in document order. */
  get nodeRules(): readonly Rule[] {
    return this.nodeRuleList;
  }

  /** Edge rules in document order. */
  get edgeRules(): readonly Rule[] {
    return this.edgeRuleList;
  }

  /** Set definitions in name-sorted order. */
  get setDefinitions(): readonly SetDefinition[] {
    return this.setDefinitionList;
  }

  /**
   * Evaluate every node and edge rule's match query, select defined sets that
   * are present in the graph, and return the selected elements with their
   * projected attributes.
   *
   * The order is deterministic: node rules and then edge rules in document
   * order, followed by set definitions in name order; nodes and sets are sorted
   * by id and edges retain graph order. Sets are never selected by a query.
   *
   * Throws a {@link MatchingError} or {@link ProjectionError} when matching or
   * projection could not be completed; in that case the result MUST NOT be
   * treated as conforming.
   */
  match(graph: GslGraph): Selection[] {
    return this.select(graph).map((entry) => entry.selection);
  }

  private select(graph: GslGraph): Array<{ selection: Selection; validate: ValidateFunction | undefined }> {
    const out: Array<{ selection: Selection; validate: ValidateFunction | undefined }> = [];

    const appendRules = (rules: readonly Rule[]): void => {
      for (const rule of rules) {
        let selected: SelectedElement[];
        try {
          selected = rule.selectElements(graph);
        } catch (err) {
          throw new MatchingError(rule.label, rule.query, err);
        }
        for (const elem of selected) {
          let attributes: Record<string, unknown>;
          try {
            attributes = projectAttributes(elem.attributes);
          } catch (err) {
            throw new ProjectionError(rule.label, elem.ref, err);
          }
          out.push({
            selection: {
              definition: rule.label,
              representation: rule.representation,
              element: elem.ref,
              attributes,
            },
            validate: rule.validate,
          });
        }
      }
    };

    appendRules(this.nodeRuleList);
    appendRules(this.edgeRuleList);

    const sets = new Map<string, GslSet>();
    for (const set of graph.sets ?? []) sets.set(set.id, set);
    for (const def of this.setDefinitionList) {
      const set = sets.get(def.name);
      if (set === undefined) continue;
      let attributes: Record<string, unknown>;
      try {
        attributes = projectAttributes(set.attributes);
      } catch (err) {
        throw new ProjectionError(def.name, def.name, err);
      }
      out.push({
        selection: {
          definition: def.name,
          representation: "set",
          element: def.name,
          attributes,
        },
        validate: def.validate,
      });
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
    const violations: Violation[] = [];
    for (const entry of this.select(graph)) {
      const validate = entry.validate;
      if (validate === undefined) continue;
      if (!validate(entry.selection.attributes)) {
        for (const e of validate.errors ?? []) {
          violations.push({
            definition: entry.selection.definition,
            representation: entry.selection.representation,
            element: entry.selection.element,
            instanceLocation: e.instancePath ?? "",
            message: e.message ?? "",
          });
        }
      }
    }

    violations.push(...validateSetPolicy(this.setDefinitionList, this.setPolicy, graph));

    return new Report(violations.length === 0 ? "conforming" : "non-conforming", violations);
  }
}

/** One graph element selected by one rule or set definition. */
export interface Selection {
  /** The rule label or set name that selected the element. */
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
  if (docs[0]!.errors.length > 0) {
    throw new LoadError("", `not a valid YAML mapping: ${docs[0]!.errors[0]!.message}`);
  }

  const root = docs[0]!.toJS();
  if (!isMapping(root)) throw new LoadError("", "must be a mapping");

  const id = root["id"];
  if (typeof id !== "string" || id.trim() === "") {
    throw new LoadError("id", "required non-empty string");
  }
  const version = root["version"];
  if (typeof version !== "string" || version.trim() === "") {
    throw new LoadError("version", "required non-empty string");
  }

  const nodeRules = parseRules("nodes", "node", root["nodes"]);
  const edgeRules = parseRules("edges", "edge", root["edges"]);
  const { policy, definitions } = parseSets(root["sets"]);

  return new Dialect(
    id,
    version,
    asString(root["title"]),
    asString(root["description"]),
    asString(root["guidance"]),
    nodeRules,
    edgeRules,
    policy,
    definitions,
  );
}

function parseRules(
  section: "nodes" | "edges",
  representation: "node" | "edge",
  value: unknown,
): Rule[] {
  if (value === undefined) return [];
  if (!isMapping(value)) throw new LoadError(section, "must be a mapping");
  const rules = value["rules"];
  if (rules === undefined) return [];
  if (!Array.isArray(rules)) throw new LoadError(`${section}.rules`, "must be a sequence");

  return rules.map((rule, index) => parseRule(section, representation, index, rule));
}

function parseRule(
  section: "nodes" | "edges",
  representation: "node" | "edge",
  index: number,
  value: unknown,
): Rule {
  const path = `${section}.rules[${index}]`;
  if (!isMapping(value)) throw new LoadError(path, "must be a mapping");

  let name = "";
  if (value["name"] !== undefined) {
    const raw = value["name"];
    if (typeof raw !== "string" || raw.trim() === "") {
      throw new LoadError(`${path}.name`, "must be a non-empty string");
    }
    name = raw;
  }

  const query = value["match"];
  if (typeof query !== "string" || query.trim() === "") {
    throw new LoadError(`${path}.match`, "required non-empty string");
  }

  const { expr } = parseMatch(`${path}.match`, representation, query);
  const schema = compileSchema(`${path}.attributes`, value["attributes"]);

  return new Rule(
    name,
    representation,
    index,
    asString(value["description"]),
    asString(value["guidance"]),
    query,
    expr,
    schema,
  );
}

/** Parse a match query, enforcing the supported GQL subset. */
function parseMatch(
  path: string,
  representation: "node" | "edge",
  query: string,
): { expr: SubgraphExpr } {
  let expanded: string;
  try {
    expanded = expandMatchQuery(representation, query);
  } catch (err) {
    throw new LoadError(path, message(err));
  }

  let parsed;
  try {
    parsed = parseQuery(expanded);
  } catch (err) {
    throw new LoadError(path, `invalid match query: ${message(err)}`);
  }
  if (parsed.expressions.length !== 1) {
    throw new LoadError(
      path,
      "unsupported match query: only a single subgraph expression is supported",
    );
  }
  const expr = parsed.expressions[0]!;
  if (!(expr instanceof SubgraphExpr)) {
    throw new LoadError(path, "unsupported match query: only the subgraph form is supported");
  }
  if (expr.traversal !== null) {
    throw new LoadError(path, "unsupported match query: traversal and scope are not supported");
  }
  return { expr };
}

function parseSets(value: unknown): { policy: SetPolicy; definitions: SetDefinition[] } {
  if (value === undefined) {
    return { policy: defaultSetPolicy, definitions: [] };
  }
  if (!isMapping(value)) throw new LoadError("sets", "must be a mapping");

  const policy = parseSetPolicy(value["policy"]);
  const definitions = parseSetDefinitions(value["definitions"]);
  return { policy, definitions };
}

function parseSetPolicy(value: unknown): SetPolicy {
  if (value === undefined) return defaultSetPolicy;
  if (!isMapping(value)) throw new LoadError("sets.policy", "must be a mapping");

  return {
    allowUndefined: optionalBoolean(
      value["allow_undefined"],
      "sets.policy.allow_undefined",
      defaultSetPolicy.allowUndefined,
    ),
    requireAll: optionalBoolean(
      value["require_all"],
      "sets.policy.require_all",
      defaultSetPolicy.requireAll,
    ),
    allowEmpty: optionalBoolean(
      value["allow_empty"],
      "sets.policy.allow_empty",
      defaultSetPolicy.allowEmpty,
    ),
  };
}

function optionalBoolean(value: unknown, path: string, fallback: boolean): boolean {
  if (value === undefined) return fallback;
  if (typeof value !== "boolean") {
    throw new LoadError(path, `must be a boolean (got ${JSON.stringify(value)})`);
  }
  return value;
}

function parseSetDefinitions(value: unknown): SetDefinition[] {
  if (value === undefined) return [];
  if (!isMapping(value)) throw new LoadError("sets.definitions", "must be a mapping");

  const definitions: SetDefinition[] = [];
  for (const [name, raw] of Object.entries(value)) {
    if (!validSetName(name)) {
      throw new LoadError(
        "sets.definitions",
        `definition key ${JSON.stringify(name)} must be a GSL set identifier`,
      );
    }
    definitions.push(parseSetDefinition(name, raw));
  }
  definitions.sort((a, b) => (a.name < b.name ? -1 : a.name > b.name ? 1 : 0));
  return definitions;
}

function parseSetDefinition(name: string, value: unknown): SetDefinition {
  const path = `sets.definitions[${name}]`;
  if (!isMapping(value)) throw new LoadError(path, "must be a mapping");

  const validate =
    value["attributes"] === undefined
      ? undefined
      : compileSchema(`${path}.attributes`, value["attributes"]);

  return new SetDefinition(
    name,
    asString(value["description"]),
    asString(value["guidance"]),
    validate,
  );
}

function validSetName(name: string): boolean {
  return /^[A-Za-z_][A-Za-z0-9_]*$/.test(name);
}

/** Compile an attribute JSON Schema, with no external loading. */
function compileSchema(path: string, attributes: unknown): ValidateFunction {
  if (attributes === undefined) {
    throw new LoadError(path, "required JSON Schema");
  }
  try {
    const ajv = new Ajv2020({ strict: false, allErrors: true });
    return ajv.compile(attributes as AnySchema);
  } catch (err) {
    throw new LoadError(path, `invalid JSON Schema: ${message(err)}`);
  }
}

function validateSetPolicy(
  definitions: readonly SetDefinition[],
  policy: SetPolicy,
  graph: GslGraph,
): Violation[] {
  const violations: Violation[] = [];
  const sets = new Map<string, GslSet>();
  for (const set of graph.sets ?? []) sets.set(set.id, set);

  const defined = new Set(definitions.map((def) => def.name));

  if (!policy.allowUndefined) {
    for (const name of [...sets.keys()].sort()) {
      if (!defined.has(name)) {
        violations.push({
          definition: "",
          representation: "set",
          element: name,
          instanceLocation: "",
          message: "set is not defined by the dialect",
        });
      }
    }
  }

  if (policy.requireAll) {
    for (const def of definitions) {
      if (!sets.has(def.name)) {
        violations.push({
          definition: def.name,
          representation: "set",
          element: def.name,
          instanceLocation: "",
          message: "required set is not present in the graph",
        });
      }
    }
  }

  if (!policy.allowEmpty) {
    const members = new Map<string, number>();
    for (const node of graph.nodes) {
      for (const name of node.sets) members.set(name, (members.get(name) ?? 0) + 1);
    }
    for (const edge of graph.edges) {
      for (const name of edge.sets) members.set(name, (members.get(name) ?? 0) + 1);
    }
    for (const name of [...sets.keys()].sort()) {
      if ((members.get(name) ?? 0) === 0) {
        violations.push({
          definition: defined.has(name) ? name : "",
          representation: "set",
          element: name,
          instanceLocation: "",
          message: "set has no members",
        });
      }
    }
  }

  return violations;
}

function isMapping(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function asString(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function message(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}
