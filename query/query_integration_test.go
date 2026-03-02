package query

import (
	"testing"
)

// TestIntegrationParseSerializeRoundTrip demonstrates round-trip parsing and serialization
func TestIntegrationParseSerializeRoundTrip(t *testing.T) {
	queries := []string{
		"start A",
		"start A | flow out",
		"start A | flow out recursive",
		`start "Service" | flow out where edge.type = "calls"`,
		"start A | where status = \"active\" | where critical = true",
		"(start A | flow out) union (start B | flow in)",
		`(start "X" | flow out recursive) minus (start "Y")`,
	}

	for _, originalQuery := range queries {
		// Parse
		q, errs := ParseQuery(originalQuery)
		if len(errs) > 0 {
			t.Fatalf("Failed to parse %q: %v", originalQuery, errs)
		}

		// Serialize
		serialized := SerializeQuery(q)

		// Re-parse to verify it's valid
		q2, errs := ParseQuery(serialized)
		if len(errs) > 0 {
			t.Fatalf("Failed to re-parse serialized query %q: %v", serialized, errs)
		}

		// Serialize again and verify it's idempotent
		reserialized := SerializeQuery(q2)
		if serialized != reserialized {
			t.Errorf("Serialization not idempotent:\n  First:  %s\n  Second: %s", serialized, reserialized)
		}

		t.Logf("✓ %s", originalQuery)
	}
}
