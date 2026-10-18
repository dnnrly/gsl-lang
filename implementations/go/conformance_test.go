// SPDX-License-Identifier: Apache-2.0
package gsl

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// conformanceGSLDir returns the path to the shared, language-neutral GSL
// canonical-form fixtures. This file lives in implementations/go, so the
// repository tests directory is two levels up.
func conformanceGSLDir() string {
	return filepath.Join("..", "..", "tests", "conformance", "gsl", "canonical")
}

// TestGSLConformance checks that serializing a parsed input matches the exact
// expected canonical form. Unlike the GQL suite, the comparison is against the
// literal bytes of expected.gsl (modulo a trailing newline), so canonical
// formatting rules such as numeric format and node-reference quoting are
// pinned.
func TestGSLConformance(t *testing.T) {
	root := conformanceGSLDir()
	entries, err := os.ReadDir(root)
	require.NoError(t, err, "failed to read GSL conformance corpus")

	found := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		found++
		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join(root, name)
			inputPath := filepath.Join(dir, "input.gsl")
			expectedPath := filepath.Join(dir, "expected.gsl")

			inputData, err := os.ReadFile(inputPath)
			require.NoError(t, err, "failed to read input.gsl")

			expectedData, err := os.ReadFile(expectedPath)
			require.NoError(t, err, "failed to read expected.gsl")

			graph, parseErr := Parse(bytes.NewReader(inputData))
			require.False(t, parseErr != nil && parseErr.HasError(), "failed to parse input.gsl: %v", parseErr)
			require.False(t, parseErr != nil && parseErr.HasWarnings(), "parsing warnings in input.gsl")

			actual := strings.TrimRight(Serialize(graph), "\n")
			expected := strings.TrimRight(string(expectedData), "\n")

			assert.Equal(t, expected, actual)
		})
	}

	assert.NotZero(t, found, "no fixtures found")
}
