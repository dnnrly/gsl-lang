// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import { Lexer, Token, TokenType } from "./lexer.js";
import type { Program, Statement, NodeDecl, EdgeDecl, SetDecl, Attribute, AttrValue } from "./ast.js";

export class Parser {
  private tokens: Token[];
  private pos: number = 0;
  private errors: Error[] = [];

  constructor(input: string) {
    const lexer = new Lexer(input);
    this.tokens = lexer.tokenize();
  }

  getErrors(): Error[] {
    return this.errors;
  }

  private peek(): Token {
    return this.pos >= this.tokens.length ? new Token(TokenType.EOF, "", 0, 0) : this.tokens[this.pos]!;
  }

  private peekNext(): Token {
    return this.pos + 1 >= this.tokens.length ? new Token(TokenType.EOF, "", 0, 0) : this.tokens[this.pos + 1]!;
  }

  private advance(): Token {
    const tok = this.peek();
    if (this.pos < this.tokens.length) {
      this.pos++;
    }
    return tok;
  }

  private addError(format: string, ...args: any[]): void {
    const msg = format.replace(/{}/g, (match, i) => String(args[i] ?? match));
    // simpler: just format
    const result = format.replace(/\%s|\%d|\%q/g, (m) => {
      if (args.length === 0) return m;
      const arg = args.shift();
      return String(arg);
    });
    this.errors.push(new Error(result));
  }

  private isReserved(tok: Token): boolean {
    return tok.type === TokenType.NODE || tok.type === TokenType.SET || tok.type === TokenType.TRUE || tok.type === TokenType.FALSE;
  }

  parse(): Program {
    const prog: Program = { statements: [] };
    while (this.peek().type !== TokenType.EOF) {
      const stmt = this.parseStatement();
      if (stmt !== null) {
        prog.statements.push(stmt);
      }
    }
    return prog;
  }

  private parseStatement(): Statement | null {
    const tok = this.peek();
    switch (tok.type) {
      case TokenType.NODE:
        return this.parseNodeDecl();
      case TokenType.SET:
        return this.parseSetDecl();
      case TokenType.IDENT:
        return this.parseEdgeDecl();
      default:
        this.addError(`unexpected token ${tok.type} (${tok.literal}) at ${tok.line}:${tok.column}`);
        this.advance();
        return null;
    }
  }

  private parseNodeDecl(): NodeDecl {
    const tok = this.advance(); // consume NODE
    const nd: NodeDecl = {
      kind: "node",
      name: "",
      attrs: [],
      block: [],
      memberships: [],
      line: tok.line,
      col: tok.column,
    };

    const nameTok = this.peek();
    if (nameTok.type !== TokenType.IDENT) {
      if (this.isReserved(nameTok)) {
        this.addError(`reserved keyword "${nameTok.literal}" cannot be used as identifier at ${nameTok.line}:${nameTok.column}`);
        this.advance();
        nd.name = nameTok.literal;
        return nd;
      }
      this.addError(`expected identifier after 'node', got ${nameTok.type} (${nameTok.literal}) at ${nameTok.line}:${nameTok.column}`);
      return nd;
    }
    nd.name = this.advance().literal;

    if (this.peek().type === TokenType.LBRACKET) {
      nd.attrs = this.parseAttributeList(true);
    }

    if (nd.attrs.length === 0 && this.peek().type === TokenType.COLON) {
      this.advance();
      const strTok = this.expect(TokenType.STRING);
      if (strTok.type === TokenType.STRING) {
        nd.textValue = strTok.literal;
      }
    }

    if (this.peek().type === TokenType.LBRACE) {
      nd.block = this.parseBlock(nd.name);
    }

    nd.memberships = this.parseMemberships();
    return nd;
  }

  private parseBlock(parentName: string): NodeDecl[] {
    this.advance(); // consume {
    const children: NodeDecl[] = [];
    while (this.peek().type !== TokenType.RBRACE && this.peek().type !== TokenType.EOF) {
      if (this.peek().type === TokenType.NODE) {
        const nd = this.parseNodeDecl();
        if (nd !== null) {
          children.push(nd);
        }
      } else {
        const tok = this.peek();
        this.addError(`expected node declaration inside block, got ${tok.type} (${tok.literal}) at ${tok.line}:${tok.column}`);
        this.advance();
      }
    }
    this.expect(TokenType.RBRACE);
    return children;
  }

  private parseAttributeList(inNodeContext: boolean): Attribute[] {
    this.advance(); // consume [
    const attrs: Attribute[] = [];
    const seen = new Set<string>();

    if (this.peek().type !== TokenType.RBRACKET) {
      const attr = this.parseAttribute(inNodeContext);
      if (seen.has(attr.key)) {
        this.addError(`duplicate attribute key "${attr.key}" at ${attr.line}:${attr.col}`);
      }
      seen.add(attr.key);
      attrs.push(attr);

      while (this.peek().type === TokenType.COMMA) {
        this.advance();
        const attr = this.parseAttribute(inNodeContext);
        if (seen.has(attr.key)) {
          this.addError(`duplicate attribute key "${attr.key}" at ${attr.line}:${attr.col}`);
        }
        seen.add(attr.key);
        attrs.push(attr);
      }
    }
    this.expect(TokenType.RBRACKET);
    return attrs;
  }

  private parseAttribute(inNodeContext: boolean): Attribute {
    const tok = this.expectIdent();
    const attr: Attribute = { key: tok.literal, line: tok.line, col: tok.column };

    if (attr.key === "parent" && !inNodeContext) {
      if (this.peek().type === TokenType.EQUALS) {
        this.advance();
        const refTok = this.expect(TokenType.IDENT);
        if (refTok.type === TokenType.IDENT) {
          attr.value = { kind: "string", strVal: refTok.literal };
        }
      } else {
        this.addError(`parent requires a value (edge label reference) at ${tok.line}:${tok.column}`);
      }
      return attr;
    }

    if (this.peek().type === TokenType.EQUALS) {
      this.advance();
      attr.value = this.parseValue(inNodeContext);
    }
    return attr;
  }

  private parseValue(inNodeContext: boolean): AttrValue {
    const tok = this.peek();
    switch (tok.type) {
      case TokenType.STRING:
        this.advance();
        return { kind: "string", strVal: tok.literal };
      case TokenType.NUMBER:
        this.advance();
        const n = parseFloat(tok.literal);
        return { kind: "number", numVal: isNaN(n) ? 0 : n };
      case TokenType.TRUE:
        this.advance();
        return { kind: "bool", boolVal: true };
      case TokenType.FALSE:
        this.advance();
        return { kind: "bool", boolVal: false };
      case TokenType.IDENT:
        if (inNodeContext) {
          this.advance();
          return { kind: "noderef", refVal: tok.literal };
        }
        this.addError(`node references are not allowed in this context at ${tok.line}:${tok.column}`);
        this.advance();
        return { kind: "noderef", refVal: tok.literal };
      default:
        this.addError(`expected value, got ${tok.type} (${tok.literal}) at ${tok.line}:${tok.column}`);
        this.advance();
        return { kind: "string", strVal: "" };
    }
  }

  private parseMemberships(): string[] {
    const memberships: string[] = [];
    while (this.peek().type === TokenType.AT) {
      this.advance();
      const ident = this.expectIdent();
      if (ident.type === TokenType.IDENT) {
        memberships.push(ident.literal);
      }
    }
    return memberships;
  }

  private parseSetDecl(): SetDecl {
    const tok = this.advance();
    const sd: SetDecl = { kind: "set", name: "", attrs: [], line: tok.line, col: tok.column };
    const nameTok = this.expectIdent();
    if (nameTok.type !== TokenType.IDENT) {
      return sd;
    }
    sd.name = nameTok.literal;
    if (this.peek().type === TokenType.LBRACKET) {
      sd.attrs = this.parseAttributeList(false);
    }
    return sd;
  }

  private parseEdgeDecl(): EdgeDecl {
    const firstTok = this.peek();
    const ed: EdgeDecl = {
      kind: "edge",
      left: [],
      right: [],
      attrs: [],
      memberships: [],
      block: [],
      line: firstTok.line,
      col: firstTok.column,
    };

    if (firstTok.type === TokenType.IDENT && this.peekNext().type === TokenType.COLON) {
      const labelTok = this.advance();
      ed.label = labelTok.literal;
      this.advance(); // consume :
      const after = this.peek();
      ed.line = after.line;
      ed.col = after.column;
    }

    const left = this.parseNodeList();
    ed.left = left;
    this.expect(TokenType.ARROW);
    const right = this.parseNodeList();
    ed.right = right;

    if (left.length > 1 && right.length > 1) {
      this.addError(`both sides of edge cannot be grouped at ${firstTok.line}:${firstTok.column}`);
    }

    switch (this.peek().type) {
      case TokenType.LBRACKET:
        ed.attrs = this.parseAttributeList(false);
        if (this.peek().type === TokenType.COLON) {
          this.advance();
          const strTok = this.expect(TokenType.STRING);
          if (strTok.type === TokenType.STRING) {
            ed.textValue = strTok.literal;
          }
        }
        break;
      case TokenType.COLON:
        this.advance();
        const strTok = this.expect(TokenType.STRING);
        if (strTok.type === TokenType.STRING) {
          ed.textValue = strTok.literal;
        }
        break;
    }

    ed.memberships = this.parseMemberships();

    if (this.peek().type === TokenType.LBRACE) {
      ed.block = this.parseScopedBlock();
    }
    return ed;
  }

  private parseNodeList(): string[] {
    const names: string[] = [];
    const ident = this.expectIdent();
    if (ident.type === TokenType.IDENT) {
      names.push(ident.literal);
    }
    while (this.peek().type === TokenType.COMMA) {
      this.advance();
      const ident = this.expectIdent();
      if (ident.type === TokenType.IDENT) {
        names.push(ident.literal);
      }
    }
    return names;
  }

  private parseScopedBlock(): Statement[] {
    this.advance(); // consume {
    const stmts: Statement[] = [];
    while (this.peek().type !== TokenType.RBRACE && this.peek().type !== TokenType.EOF) {
      const tok = this.peek();
      switch (tok.type) {
        case TokenType.NODE:
          const nd = this.parseNodeDecl();
          if (nd !== null) stmts.push(nd);
          break;
        case TokenType.SET:
          const sd = this.parseSetDecl();
          if (sd !== null) stmts.push(sd);
          break;
        case TokenType.IDENT:
          const ed = this.parseEdgeDecl();
          if (ed !== null) stmts.push(ed);
          break;
        default:
          this.addError(`unexpected token ${tok.type} (${tok.literal}) inside edge scope at ${tok.line}:${tok.column}`);
          this.advance();
      }
    }
    this.expect(TokenType.RBRACE);
    return stmts;
  }

  private expect(t: TokenType): Token {
    const tok = this.peek();
    if (tok.type !== t) {
      this.addError(`expected ${t}, got ${tok.type} (${tok.literal}) at ${tok.line}:${tok.column}`);
      return tok;
    }
    return this.advance();
  }

  private expectIdent(): Token {
    const tok = this.peek();
    if (this.isReserved(tok)) {
      this.addError(`reserved keyword "${tok.literal}" cannot be used as identifier at ${tok.line}:${tok.column}`);
      this.advance();
      return new Token(TokenType.IDENT, tok.literal, tok.line, tok.column);
    }
    return this.expect(TokenType.IDENT);
  }
}

export function parseProgram(input: string): { program: Program; errors: Error[] } {
  const p = new Parser(input);
  const program = p.parse();
  return { program, errors: p.getErrors() };
}
