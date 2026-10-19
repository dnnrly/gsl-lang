// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

import { isNodeRef, type Attributes } from "@gsl-lang/core";

/**
 * Convert one element's GSL attribute map into a JSON-compatible object.
 *
 * The projection is deterministic and preserves the semantic distinctions
 * defined by GSL; it never coerces one value type into another. The mappings
 * are normative (`spec/dialects/DIALECTS_SPEC.md` §8.3):
 *
 * - string        -> JSON string
 * - number        -> JSON number (preserving the value)
 * - boolean       -> JSON boolean
 * - Empty (null)  -> `{ "$gslEmpty": true }`
 * - NodeRef       -> `{ "$gslRef": "<node-id>" }`
 *
 * An attribute that is absent in GSL is simply not present in the map. An
 * explicitly present Empty value is projected as `{ "$gslEmpty": true }`; it is
 * never treated as an absent attribute, JSON null or boolean true.
 */
export function projectAttributes(attrs: Attributes): Record<string, unknown> {
  const out: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(attrs)) {
    out[key] = projectValue(value);
  }
  return out;
}

function projectValue(value: unknown): unknown {
  if (value === null || value === undefined) return { $gslEmpty: true };
  if (isNodeRef(value)) return { $gslRef: value.id };
  if (typeof value === "string" || typeof value === "number" || typeof value === "boolean") {
    return value;
  }
  throw new Error(`unsupported GSL value type ${describe(value)}`);
}

function describe(value: unknown): string {
  if (typeof value === "object" && value !== null) return value.constructor.name;
  return typeof value;
}
