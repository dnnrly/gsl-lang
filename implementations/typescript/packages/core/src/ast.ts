// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

export type AttrValueKind = "string" | "number" | "bool" | "noderef";

export interface AttrValue {
  kind: AttrValueKind;
  strVal?: string;
  numVal?: number;
  boolVal?: boolean;
  refVal?: string;
}

export interface Attribute {
  key: string;
  value?: AttrValue; // undefined means empty/flag
  line: number;
  col: number;
}

export interface NodeDecl {
  kind: "node";
  name: string;
  textValue?: string;
  attrs: Attribute[];
  block: NodeDecl[];
  memberships: string[];
  line: number;
  col: number;
}

export interface EdgeDecl {
  kind: "edge";
  label?: string;
  left: string[];
  right: string[];
  textValue?: string;
  attrs: Attribute[];
  memberships: string[];
  block: Statement[];
  line: number;
  col: number;
}

export interface SetDecl {
  kind: "set";
  name: string;
  attrs: Attribute[];
  line: number;
  col: number;
}

export type Statement = NodeDecl | EdgeDecl | SetDecl;

export interface Program {
  statements: Statement[];
}
