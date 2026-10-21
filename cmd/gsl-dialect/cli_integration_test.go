//go:build integration

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// getBinaryPath returns the path to the gsl-dialect binary, building it if
// necessary.
func getBinaryPath(t *testing.T) string {
	t.Helper()
	for _, path := range []string{"./gsl-dialect", "./tmp/gsl-dialect", "../../tmp/gsl-dialect"} {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	cmd := exec.Command("go", "build", "-o", "gsl-dialect", ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build gsl-dialect: %v\nOutput: %s", err, output)
	}
	return "./gsl-dialect"
}

const integrationDialect = `
id: integration
version: "1"
elements:
  Service:
    representation: node
    match:
      query: 'kind == "service"'
    attributes:
      type: object
      required: [name]
      properties:
        name:
          type: string
`

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
	return path
}

func TestCLIExitCodes(t *testing.T) {
	dir := t.TempDir()
	dialectPath := writeFile(t, dir, "dialect.yaml", integrationDialect)
	conforming := writeFile(t, dir, "conforming.gsl", "node a [kind=\"service\", name=\"A\"]\n")
	nonConforming := writeFile(t, dir, "non-conforming.gsl", "node a [kind=\"service\"]\n")
	invalidDialect := writeFile(t, dir, "broken.yaml", "id: broken\n")
	binary := getBinaryPath(t)

	tests := []struct {
		name string
		args []string
		want int
	}{
		{"conforming", []string{"validate", "-d", dialectPath, "-i", conforming}, 0},
		{"non-conforming", []string{"validate", "-d", dialectPath, "-i", nonConforming}, 1},
		{"invalid-dialect", []string{"validate", "-d", invalidDialect, "-i", conforming}, 2},
		{"missing-dialect-flag", []string{"validate", "-i", conforming}, 2},
		{"project", []string{"project", "-d", dialectPath, "-i", conforming}, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binary, tc.args...)
			output, err := cmd.CombinedOutput()
			code := 0
			if err != nil {
				exitErr, ok := err.(*exec.ExitError)
				if !ok {
					t.Fatalf("unexpected error running %v: %v\nOutput: %s", tc.args, err, output)
				}
				code = exitErr.ExitCode()
			}
			if code != tc.want {
				t.Fatalf("exit code = %d, want %d\nOutput: %s", code, tc.want, output)
			}
		})
	}
}
