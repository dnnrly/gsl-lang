// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import { TokenType, tokenize } from "./lexer.js";

/**
 * Identifiers that carry grammatical meaning inside a dialect match expression
 * and must not be treated as bare attribute names.
 */
const MATCH_KEYWORDS = new Set([
  "AND",
  "parent",
  "exists",
  "not",
  "in",
  "depends",
  "on",
  "depth",
  "true",
  "false",
]);

export class MatchExpandError extends Error {}

/**
 * Expand a dialect match expression into a full GQL query.
 *
 * The expression is a bare attribute predicate: a single `subgraph` predicate
 * written without the `subgraph` keyword or an element prefix. Each bare
 * attribute reference is prefixed with `element` and the result is wrapped in
 * `subgraph`. For example, with element `node`,
 * `kind == "service" AND active exists` expands to
 * `subgraph node.kind == "service" AND node.active exists`.
 *
 * The rewrite is token-aware: string literals and numbers are never modified,
 * and grammatical keywords are preserved. `depth` is a derived edge attribute
 * and always expands to `edge.depth`.
 *
 * Throws a {@link MatchExpandError} if the expression contains an explicit
 * element prefix, an explicit `subgraph` keyword, a pipeline, or another token
 * that cannot appear in a bare match predicate.
 */
export function expandMatchQuery(element: string, expr: string): string {
  if (element !== "node" && element !== "edge") {
    throw new MatchExpandError(`unsupported element prefix ${JSON.stringify(element)}`);
  }

  const tokens = tokenize(expr).filter((t) => t.type !== "EOF");
  if (tokens.length === 0) {
    throw new MatchExpandError("empty match expression");
  }

  const out: string[] = ["subgraph"];
  const emit = (s: string): void => {
    out.push(s);
  };

  let sawTerm = false;
  let expectTerm = true; // at the start of a predicate term
  let afterAt = false; // the next ident is a set reference, not an attribute

  for (let i = 0; i < tokens.length; i++) {
    const tok = tokens[i]!;

    switch (tok.type as TokenType) {
      case "Ident": {
        const word = tok.value;

        if (afterAt) {
          // A set reference following `@`; never an attribute. Attach it to
          // the `@` already emitted so the set ref stays one token.
          out[out.length - 1] = out[out.length - 1]! + word;
          afterAt = false;
          expectTerm = false;
          break;
        }

        if (word === "subgraph") {
          throw new MatchExpandError("the subgraph keyword is implied and must not be written");
        }
        if (word === "node" || word === "edge" || word === "set") {
          throw new MatchExpandError(
            `element prefix ${JSON.stringify(word)} must not be written; the representation supplies it`,
          );
        }

        if (!MATCH_KEYWORDS.has(word)) {
          if (!expectTerm) {
            throw new MatchExpandError(`unexpected identifier ${JSON.stringify(word)} in match expression`);
          }
          emit(`${element}.${word}`);
          expectTerm = false;
          sawTerm = true;
          break;
        }

        switch (word) {
          case "AND":
            emit("AND");
            expectTerm = true;
            break;
          case "depth":
            emit("edge.depth");
            expectTerm = false;
            sawTerm = true;
            break;
          case "parent":
            emit(element);
            emit("parent");
            expectTerm = false;
            sawTerm = true;
            break;
          case "depends":
            emit(element);
            emit("depends");
            expectTerm = false;
            sawTerm = true;
            break;
          case "on":
            emit("on");
            expectTerm = true;
            break;
          case "in":
            if (expectTerm) emit(element);
            emit("in");
            expectTerm = false;
            sawTerm = true;
            break;
          case "not":
            // `not in` (bare set non-membership) versus `attr not exists`.
            if (expectTerm && i + 1 < tokens.length && tokens[i + 1]!.value === "in") {
              emit(element);
              emit("not");
              emit("in");
              i++; // consume the `in` token
              expectTerm = false;
              sawTerm = true;
              break;
            }
            emit("not");
            break;
          case "exists":
            emit("exists");
            expectTerm = false;
            break;
          default:
            emit(word);
            expectTerm = false;
            break;
        }
        break;
      }

      case "At":
        emit(tok.value);
        afterAt = true;
        break;

      case "String":
      case "Number":
      case "Operator":
      case "Assign":
        emit(tok.value);
        expectTerm = false;
        break;

      default:
        throw new MatchExpandError(`unsupported token ${JSON.stringify(tok.value)} in match expression`);
    }
  }

  if (!sawTerm) {
    throw new MatchExpandError("match expression contains no attribute predicate");
  }
  return out.join(" ");
}
