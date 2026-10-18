// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

export interface Diagnostic {
  message: string;
  line?: number;
  column?: number;
}

export interface ParseErrorDetail extends Diagnostic {
}

export class ParseError extends Error {
  public readonly errors: ParseErrorDetail[];
  public readonly warnings: Diagnostic[];

  constructor(message: string, errors: ParseErrorDetail[] = [], warnings: Diagnostic[] = []) {
    super(message);
    this.name = "ParseError";
    this.errors = errors;
    this.warnings = warnings;
  }
}
