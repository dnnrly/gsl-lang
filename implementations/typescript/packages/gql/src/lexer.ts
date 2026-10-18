// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

export type TokenType =
  | "Comment"
  | "String"
  | "Number"
  | "Operator"
  | "Assign"
  | "Pipe"
  | "LParen"
  | "RParen"
  | "Dot"
  | "At"
  | "AlgebraOp"
  | "Star"
  | "Ident"
  | "EOF";

export interface Token {
  type: TokenType;
  value: string;
  pos: number;
}

export class GqlLexError extends Error {}

const IDENT_START = /[a-zA-Z_]/;
const IDENT_REST = /[a-zA-Z0-9_]/;

export function tokenize(input: string): Token[] {
  const tokens: Token[] = [];
  let i = 0;
  const n = input.length;

  while (i < n) {
    const ch = input[i]!;

    if (ch === "#") {
      let j = i;
      while (j < n && input[j] !== "\n") j++;
      i = j;
      continue;
    }

    if (/\s/.test(ch)) {
      i++;
      continue;
    }

    if (ch === '"') {
      let j = i + 1;
      while (j < n && input[j] !== '"') j++;
      if (j >= n) throw new GqlLexError("unterminated string");
      tokens.push({ type: "String", value: input.slice(i, j + 1), pos: i });
      i = j + 1;
      continue;
    }

    // Number: [-+]?(\d+\.?\d*|\.\d+)
    if (/[0-9]/.test(ch) || ((ch === "-" || ch === "+") && /[0-9.]/.test(input[i + 1] ?? ""))) {
      const rest = input.slice(i);
      const m = /^[-+]?(?:\d+\.?\d*|\.\d+)/.exec(rest);
      if (m) {
        tokens.push({ type: "Number", value: m[0], pos: i });
        i += m[0].length;
        continue;
      }
    }

    if (ch === "=" && input[i + 1] === "=") {
      tokens.push({ type: "Operator", value: "==", pos: i });
      i += 2;
      continue;
    }

    if (ch === "!" && input[i + 1] === "=") {
      tokens.push({ type: "Operator", value: "!=", pos: i });
      i += 2;
      continue;
    }

    if (ch === "=") {
      tokens.push({ type: "Assign", value: "=", pos: i });
      i++;
      continue;
    }

    if (ch === "|") {
      tokens.push({ type: "Pipe", value: "|", pos: i });
      i++;
      continue;
    }

    if (ch === "(") {
      tokens.push({ type: "LParen", value: "(", pos: i });
      i++;
      continue;
    }

    if (ch === ")") {
      tokens.push({ type: "RParen", value: ")", pos: i });
      i++;
      continue;
    }

    if (ch === ".") {
      tokens.push({ type: "Dot", value: ".", pos: i });
      i++;
      continue;
    }

    if (ch === "@") {
      tokens.push({ type: "At", value: "@", pos: i });
      i++;
      continue;
    }

    if (ch === "+" || ch === "&" || ch === "-" || ch === "^") {
      tokens.push({ type: "AlgebraOp", value: ch, pos: i });
      i++;
      continue;
    }

    if (ch === "*") {
      tokens.push({ type: "Star", value: "*", pos: i });
      i++;
      continue;
    }

    if (IDENT_START.test(ch)) {
      let j = i + 1;
      while (j < n && IDENT_REST.test(input[j]!)) j++;
      tokens.push({ type: "Ident", value: input.slice(i, j), pos: i });
      i = j;
      continue;
    }

    throw new GqlLexError(`unexpected character ${JSON.stringify(ch)} at ${i}`);
  }

  tokens.push({ type: "EOF", value: "", pos: n });
  return tokens;
}
