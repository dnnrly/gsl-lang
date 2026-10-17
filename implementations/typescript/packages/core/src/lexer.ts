// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

export enum TokenType {
  ILLEGAL = 0,
  EOF,
  IDENT,
  STRING,
  NUMBER,
  NODE,
  SET,
  TRUE,
  FALSE,
  LBRACKET,
  RBRACKET,
  LBRACE,
  RBRACE,
  COMMA,
  EQUALS,
  ARROW,
  AT,
  COLON,
}

export class Token {
  constructor(
    public readonly type: TokenType,
    public readonly literal: string,
    public readonly line: number = 1,
    public readonly column: number = 1,
  ) {}
}

const keywords: Record<string, TokenType> = {
  node: TokenType.NODE,
  set: TokenType.SET,
  true: TokenType.TRUE,
  false: TokenType.FALSE,
};

export class Lexer {
  private input: string[];
  private pos: number = 0;
  private line: number = 1;
  private col: number = 1;

  constructor(input: string) {
    this.input = input.split("");
  }

  private peek(): string {
    return this.pos >= this.input.length ? "\0" : this.input[this.pos]!;
  }

  private advance(): string {
    if (this.pos >= this.input.length) {
      return "\0";
    }
    const ch = this.input[this.pos++]!;
    if (ch === "\n") {
      this.line++;
      this.col = 1;
    } else {
      this.col++;
    }
    return ch;
  }

  private skipWhitespace(): void {
    while (this.pos < this.input.length) {
      const ch = this.input[this.pos]!;
      if (ch === " " || ch === "\t" || ch === "\r" || ch === "\n") {
        this.advance();
      } else {
        break;
      }
    }
  }

  private skipComment(): void {
    while (this.pos < this.input.length && this.input[this.pos] !== "\n") {
      this.advance();
    }
  }

  private readString(): Token {
    const line = this.line;
    const col = this.col;
    this.advance(); // consume opening "
    let lit = "";
    while (true) {
      if (this.pos >= this.input.length) {
        return new Token(TokenType.ILLEGAL, lit, line, col);
      }
      const ch = this.advance();
      if (ch === '"') {
        return new Token(TokenType.STRING, lit, line, col);
      }
      if (ch === "\\") {
        if (this.pos >= this.input.length) {
          return new Token(TokenType.ILLEGAL, lit, line, col);
        }
        const esc = this.advance();
        switch (esc) {
          case '"':
            lit += '"';
            break;
          case "\\":
            lit += "\\";
            break;
          case "n":
            lit += "\n";
            break;
          case "t":
            lit += "\t";
            break;
          default:
            lit += "\\" + esc;
            break;
        }
      } else {
        lit += ch;
      }
    }
  }

  private readNumber(): Token {
    const line = this.line;
    const col = this.col;
    const start = this.pos;
    while (this.pos < this.input.length) {
      const ch = this.input[this.pos]!;
      if (ch >= "0" && ch <= "9") {
        this.advance();
      } else {
        break;
      }
    }
    if (this.pos < this.input.length && this.input[this.pos] === ".") {
      this.advance();
      while (this.pos < this.input.length) {
        const ch = this.input[this.pos]!;
        if (ch >= "0" && ch <= "9") {
          this.advance();
        } else {
          break;
        }
      }
    }
    const literal = this.input.slice(start, this.pos).join("");
    return new Token(TokenType.NUMBER, literal, line, col);
  }

  private readIdentifier(): Token {
    const line = this.line;
    const col = this.col;
    const start = this.pos;
    while (this.pos < this.input.length) {
      const ch = this.input[this.pos]!;
      if ((ch >= "A" && ch <= "Z") || (ch >= "a" && ch <= "z") || (ch >= "0" && ch <= "9") || ch === "_") {
        this.advance();
      } else {
        break;
      }
    }
    const literal = this.input.slice(start, this.pos).join("");
    const tt = keywords[literal];
    if (tt !== undefined) {
      return new Token(tt, literal, line, col);
    }
    return new Token(TokenType.IDENT, literal, line, col);
  }

  nextToken(): Token {
    for (;;) {
      this.skipWhitespace();
      if (this.pos >= this.input.length) {
        return new Token(TokenType.EOF, "", this.line, this.col);
      }

      const ch = this.peek();
      if (ch === "#") {
        this.skipComment();
        continue;
      }

      const line = this.line;
      const col = this.col;

      switch (ch) {
        case "[":
          this.advance();
          return new Token(TokenType.LBRACKET, "[", line, col);
        case "]":
          this.advance();
          return new Token(TokenType.RBRACKET, "]", line, col);
        case "{":
          this.advance();
          return new Token(TokenType.LBRACE, "{", line, col);
        case "}":
          this.advance();
          return new Token(TokenType.RBRACE, "}", line, col);
        case ",":
          this.advance();
          return new Token(TokenType.COMMA, ",", line, col);
        case "=":
          this.advance();
          return new Token(TokenType.EQUALS, "=", line, col);
        case "@":
          this.advance();
          return new Token(TokenType.AT, "@", line, col);
        case ":":
          this.advance();
          return new Token(TokenType.COLON, ":", line, col);
        case "-":
          this.advance();
          if (this.pos < this.input.length && this.input[this.pos] === ">") {
            this.advance();
            return new Token(TokenType.ARROW, "->", line, col);
          }
          return new Token(TokenType.ILLEGAL, "-", line, col);
        case '"':
          return this.readString();
        default:
          if (ch >= "0" && ch <= "9") {
            return this.readNumber();
          }
          if ((ch >= "A" && ch <= "Z") || (ch >= "a" && ch <= "z") || ch === "_") {
            return this.readIdentifier();
          }
          this.advance();
          return new Token(TokenType.ILLEGAL, ch, line, col);
      }
    }
  }

  tokenize(): Token[] {
    const tokens: Token[] = [];
    for (;;) {
      const tok = this.nextToken();
      tokens.push(tok);
      if (tok.type === TokenType.EOF) {
        break;
      }
    }
    return tokens;
  }
}
