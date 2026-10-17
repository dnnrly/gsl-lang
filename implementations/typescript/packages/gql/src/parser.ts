// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import { Token, tokenize } from "./lexer.js";
import {
  AndPredicate,
  AttributeEqualsPredicate,
  AttributeExistsPredicate,
  AttributeNotEqualsPredicate,
  AttributeNotExistsPredicate,
  DepthPredicate,
  ExistsPredicate,
  isValidGraphName,
  isValidGraphRef,
  NotPredicate,
  ParentExistsPredicate,
  ParentNotExistsPredicate,
  ParentPredicate,
  Predicate,
  QueryValue,
  SetMembershipPredicate,
  SetNotMembershipPredicate,
} from "./predicates.js";
import {
  BindExpr,
  CollapseExpr,
  Expression,
  FromExpr,
  GraphAlgebraExpr,
  IdentityExpr,
  MakeExpr,
  Query,
  RemoveAttributeExpr,
  RemoveEdgeExpr,
  RemoveOrphansExpr,
  SubgraphExpr,
  TraversalConfig,
} from "./expressions.js";

const DIRECTIONS = new Set(["in", "out", "both", "up", "down"]);
const ALL_DEPTH = 999999;

export class GqlParseError extends Error {}

class Parser {
  private pos = 0;

  constructor(private readonly tokens: Token[]) {}

  private peek(offset = 0): Token {
    return this.tokens[Math.min(this.pos + offset, this.tokens.length - 1)]!;
  }

  private next(): Token {
    const token = this.tokens[this.pos]!;
    if (this.pos < this.tokens.length - 1) this.pos++;
    return token;
  }

  private at(type: Token["type"], value?: string): boolean {
    const token = this.peek();
    if (token.type !== type) return false;
    if (value !== undefined && token.value !== value) return false;
    return true;
  }

  private expect(type: Token["type"], value?: string): Token {
    const token = this.peek();
    if (token.type !== type || (value !== undefined && token.value !== value)) {
      throw new GqlParseError(
        `expected ${value ?? type} but found ${JSON.stringify(token.value)}`,
      );
    }
    return this.next();
  }

  private expectKeyword(word: string): void {
    if (!this.at("Ident", word)) {
      throw new GqlParseError(`expected keyword ${word}`);
    }
    this.next();
  }

  parseQuery(): Query {
    if (this.at("EOF")) return new Query([new IdentityExpr()]);
    const expressions: Expression[] = [];
    expressions.push(this.parseExpression());
    while (this.at("Pipe")) {
      this.next();
      expressions.push(this.parseExpression());
    }
    return new Query(expressions);
  }

  private parseExpression(): Expression {
    if (this.at("LParen")) return this.parseBinding();
    if (this.at("Ident", "subgraph")) return this.parseSubgraph();
    if (this.at("Ident", "make")) return this.parseMake();
    if (this.at("Ident", "remove")) return this.parseRemove();
    if (this.at("Ident", "collapse")) return this.parseCollapse();
    if (this.at("Ident", "from")) return this.parseFrom();

    if (this.at("Star") || this.at("Ident")) {
      const nextIsAlgebra = this.peek(1).type === "AlgebraOp";
      if (nextIsAlgebra) return this.parseGraphAlgebra();
    }

    throw new GqlParseError(`unexpected token ${JSON.stringify(this.peek().value)}`);
  }

  private parseBinding(): Expression {
    this.expect("LParen");
    const pipeline = this.parseQuery();
    this.expect("RParen");
    this.expectKeyword("as");
    const name = this.expect("Ident").value;
    if (!isValidGraphName(name)) throw new GqlParseError(`invalid graph name: ${name}`);
    return new BindExpr(pipeline, name);
  }

  private parseSubgraph(): Expression {
    this.expectKeyword("subgraph");
    const pred = this.parsePredicate();

    let traversal: TraversalConfig | null = null;
    if (this.at("Ident", "scope")) {
      this.next();
      traversal = { directions: ["down"], depth: ALL_DEPTH };
    } else if (this.at("Ident", "traverse")) {
      this.next();
      const directions: string[] = [];
      while (this.at("Ident") && DIRECTIONS.has(this.peek().value)) {
        directions.push(this.next().value);
      }
      if (directions.length === 0) throw new GqlParseError("traverse requires a direction");
      let depth: number;
      if (this.at("Ident", "all")) {
        this.next();
        depth = ALL_DEPTH;
      } else {
        const numToken = this.expect("Number");
        depth = Number.parseInt(numToken.value, 10);
      }
      traversal = { directions, depth };
    }

    return new SubgraphExpr(pred, traversal);
  }

  private parseMake(): Expression {
    this.expectKeyword("make");
    const { element, attr } = this.parseAttributePath();
    this.expect("Assign");
    const value = this.parseValue();
    this.expectKeyword("where");
    const pred = this.parsePredicate();
    return new MakeExpr(element, attr, value, pred);
  }

  private parseRemove(): Expression {
    this.expectKeyword("remove");
    if (this.at("Ident", "orphans")) {
      this.next();
      return new RemoveOrphansExpr();
    }
    // "edge where" → edge predicate removal
    if (this.at("Ident", "edge") && this.peek(1).type === "Ident" && this.peek(1).value === "where") {
      this.next();
      this.expectKeyword("where");
      const pred = this.parsePredicate();
      return new RemoveEdgeExpr(pred);
    }
    const { element, attr } = this.parseAttributePath();
    let pred: Predicate = new ExistsPredicate();
    if (this.at("Ident", "where")) {
      this.next();
      pred = this.parsePredicate();
    }
    return new RemoveAttributeExpr(element, attr, pred);
  }

  private parseCollapse(): Expression {
    this.expectKeyword("collapse");
    this.expectKeyword("into");
    const nodeId = this.expect("Ident").value;
    this.expectKeyword("where");
    const pred = this.parsePredicate();
    return new CollapseExpr(nodeId, pred);
  }

  private parseFrom(): Expression {
    this.expectKeyword("from");
    if (this.at("Star")) {
      this.next();
      return new FromExpr(true, "");
    }
    const name = this.expect("Ident").value;
    if (!isValidGraphName(name)) throw new GqlParseError(`invalid graph name: ${name}`);
    return new FromExpr(false, name);
  }

  private parseGraphAlgebra(): Expression {
    const left = this.at("Star") ? "*" : this.next().value;
    if (!isValidGraphRef(left)) throw new GqlParseError(`invalid graph reference: ${left}`);
    this.expect("AlgebraOp");
    const op = this.tokens[this.pos - 1]!.value;
    const right = this.at("Star") ? "*" : this.expect("Ident").value;
    if (!isValidGraphRef(right)) throw new GqlParseError(`invalid graph reference: ${right}`);
    return new GraphAlgebraExpr(left, op, right);
  }

  private parseAttributePath(): { element: string; attr: string } {
    const element = this.expect("Ident").value;
    if (element !== "node" && element !== "edge") {
      throw new GqlParseError(`expected node or edge, found ${element}`);
    }
    this.expect("Dot");
    const attr = this.expect("Ident").value;
    return { element, attr };
  }

  private parsePredicate(): Predicate {
    const first = this.parsePredicateTerm();
    let pred = first;
    while (this.at("Ident", "AND")) {
      this.next();
      const right = this.parsePredicateTerm();
      pred = new AndPredicate(pred, right);
    }
    return pred;
  }

  private parsePredicateTerm(): Predicate {
    // Bare "exists"
    if (this.at("Ident", "exists")) {
      this.next();
      return new ExistsPredicate();
    }

    // Bare "in @set"
    if (this.at("Ident", "in")) {
      this.next();
      this.expect("At");
      const name = this.expect("Ident").value;
      return new SetMembershipPredicate("", name);
    }

    // "not" prefix: "not in @set" or "not <predicate>"
    if (this.at("Ident", "not")) {
      this.next();
      if (this.at("Ident", "in")) {
        this.next();
        this.expect("At");
        const name = this.expect("Ident").value;
        return new SetNotMembershipPredicate("", name);
      }
      return new NotPredicate(this.parsePredicate());
    }

    // element-prefixed predicates
    if (this.at("Ident", "node") || this.at("Ident", "edge")) {
      const element = this.peek().value;
      const after = this.peek(1);
      if (after.type === "Ident" && after.value === "in") {
        this.next();
        this.next();
        this.expect("At");
        const name = this.expect("Ident").value;
        return new SetMembershipPredicate(element, name);
      }
      if (after.type === "Ident" && after.value === "not" && this.peek(2).value === "in") {
        this.next();
        this.next();
        this.next();
        this.expect("At");
        const name = this.expect("Ident").value;
        return new SetNotMembershipPredicate(element, name);
      }
      if (after.type === "Ident" && after.value === "depends") {
        this.next();
        this.expectKeyword("depends");
        this.expectKeyword("on");
        const inner = this.parsePredicate();
        return new ParentPredicate(element, inner);
      }
      if (after.type === "Ident" && after.value === "parent") {
        this.next();
        this.expectKeyword("parent");
        let not = false;
        if (this.at("Ident", "not")) {
          this.next();
          not = true;
        }
        this.expectKeyword("exists");
        return not ? new ParentNotExistsPredicate(element) : new ParentExistsPredicate(element);
      }
      if (after.type === "Dot") {
        this.next();
        return this.parseAttributePredicate(element);
      }
    }

    throw new GqlParseError(`unexpected predicate token ${JSON.stringify(this.peek().value)}`);
  }

  private parseAttributePredicate(element: string): Predicate {
    this.expect("Dot");
    const attr = this.expect("Ident").value;

    if (element === "edge" && attr === "depth") {
      if (this.at("Operator")) {
        const op = this.next().value;
        const num = this.expect("Number").value;
        const value = Number.parseInt(num, 10);
        return new DepthPredicate(op === "=" ? "==" : op, value);
      }
      if (this.at("Ident", "exists")) {
        this.next();
        return new ExistsPredicate();
      }
      if (this.at("Ident", "not")) {
        this.next();
        this.expectKeyword("exists");
        return new NotPredicate(new ExistsPredicate());
      }
      throw new GqlParseError("invalid depth predicate");
    }

    if (this.at("Ident", "exists")) {
      this.next();
      return new AttributeExistsPredicate(element, attr);
    }
    if (this.at("Ident", "not")) {
      this.next();
      this.expectKeyword("exists");
      return new AttributeNotExistsPredicate(element, attr);
    }
    if (this.at("Operator") || this.at("Assign")) {
      const op = this.next().value;
      const value = this.parseValue();
      return op === "!="
        ? new AttributeNotEqualsPredicate(element, attr, value)
        : new AttributeEqualsPredicate(element, attr, value);
    }

    throw new GqlParseError("invalid attribute predicate");
  }

  private parseValue(): QueryValue {
    const token = this.peek();
    if (token.type === "String") {
      this.next();
      return token.value.slice(1, -1);
    }
    if (token.type === "Ident" && (token.value === "true" || token.value === "false")) {
      this.next();
      return token.value === "true";
    }
    if (token.type === "Number") {
      this.next();
      return token.value;
    }
    if (token.type === "Ident") {
      this.next();
      return token.value;
    }
    throw new GqlParseError(`expected value but found ${JSON.stringify(token.value)}`);
  }

  expectEof(): void {
    if (!this.at("EOF")) {
      throw new GqlParseError(`unexpected trailing token ${JSON.stringify(this.peek().value)}`);
    }
  }
}

export function parseQuery(input: string): Query {
  const trimmed = input.trim();
  if (trimmed === "") return new Query([new IdentityExpr()]);
  const parser = new Parser(tokenize(trimmed));
  const query = parser.parseQuery();
  parser.expectEof();
  return query;
}
