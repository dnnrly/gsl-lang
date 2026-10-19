// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

package dialect

import (
	"fmt"

	gsl "github.com/dnnrly/gsl-lang/implementations/go"
)

// JSON keys used to project GSL values that have no direct JSON equivalent.
const (
	keyGSLEmpty = "$gslEmpty"
	keyGSLRef   = "$gslRef"
)

// projectAttributes converts one element's GSL attribute map into a
// JSON-compatible object. The projection is deterministic and preserves the
// semantic distinctions defined by GSL; it never coerces one value type into
// another.
//
// The mappings are normative (spec §8.3):
//
//	string      -> JSON string
//	float64     -> JSON number (preserving the value)
//	bool        -> JSON boolean
//	Empty (nil) -> {"$gslEmpty": true}
//	NodeRef     -> {"$gslRef": "<node-id>"}
//
// An attribute that is absent in GSL is simply not present in the map. An
// explicitly present Empty value has a nil Go value and is projected as
// {"$gslEmpty": true}; it is never treated as an absent attribute, JSON null or
// boolean true.
func projectAttributes(attrs gsl.AttributeMap) (map[string]any, error) {
	out := make(map[string]any, len(attrs))
	for key, value := range attrs {
		pv, err := projectValue(value)
		if err != nil {
			return nil, fmt.Errorf("attribute %q: %w", key, err)
		}
		out[key] = pv
	}
	return out, nil
}

func projectValue(value any) (any, error) {
	switch v := value.(type) {
	case nil:
		return map[string]any{keyGSLEmpty: true}, nil
	case string:
		return v, nil
	case float64:
		return v, nil
	case bool:
		return v, nil
	case gsl.NodeRef:
		return map[string]any{keyGSLRef: string(v)}, nil
	default:
		return nil, fmt.Errorf("unsupported GSL value type %T", value)
	}
}
