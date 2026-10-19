// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validDialect = `
id: test-dialect
version: "1"
elements:
  Service:
    representation: node
    match:
      query: 'subgraph node.kind == "service"'
    attributes:
      type: object
      required: [name]
      properties:
        name:
          type: string
`

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
	return path
}

func exitCode(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	var ee *exitError
	if !errors.As(err, &ee) {
		t.Fatalf("expected an *exitError, got %T: %v", err, err)
	}
	return ee.code
}

func TestValidateConforming(t *testing.T) {
	var out bytes.Buffer
	err := runValidate(&validateConfig{
		DialectFile: writeTemp(t, "d.yaml", validDialect),
		Stdin:       strings.NewReader("node a [kind=\"service\", name=\"A\"]\n"),
		Stdout:      &out,
	})
	if code := exitCode(t, err); code != ExitConforming {
		t.Fatalf("exit code = %d, want %d (%v)", code, ExitConforming, err)
	}
	if !strings.Contains(out.String(), "Conforming:") {
		t.Errorf("output missing Conforming line: %q", out.String())
	}
}

func TestValidateNonConforming(t *testing.T) {
	var out bytes.Buffer
	err := runValidate(&validateConfig{
		DialectFile: writeTemp(t, "d.yaml", validDialect),
		Stdin:       strings.NewReader("node a [kind=\"service\"]\n"),
		Stdout:      &out,
	})
	if code := exitCode(t, err); code != ExitNonConforming {
		t.Fatalf("exit code = %d, want %d (%v)", code, ExitNonConforming, err)
	}
	if !strings.Contains(out.String(), "Non-conforming: 1 violation(s)") {
		t.Errorf("output missing violation summary: %q", out.String())
	}
}

func TestValidateInvalidDialect(t *testing.T) {
	err := runValidate(&validateConfig{
		DialectFile: writeTemp(t, "d.yaml", "id: broken\n"),
		Stdin:       strings.NewReader("node a\n"),
		Stdout:      &bytes.Buffer{},
	})
	if code := exitCode(t, err); code != ExitIncomplete {
		t.Fatalf("exit code = %d, want %d (%v)", code, ExitIncomplete, err)
	}
}

func TestValidateGraphParseError(t *testing.T) {
	err := runValidate(&validateConfig{
		DialectFile: writeTemp(t, "d.yaml", validDialect),
		Stdin:       strings.NewReader("this is not ; valid gsl"),
		Stdout:      &bytes.Buffer{},
	})
	if code := exitCode(t, err); code != ExitIncomplete {
		t.Fatalf("exit code = %d, want %d (%v)", code, ExitIncomplete, err)
	}
}

func TestValidateJSONOutput(t *testing.T) {
	var out bytes.Buffer
	_ = runValidate(&validateConfig{
		DialectFile: writeTemp(t, "d.yaml", validDialect),
		Stdin:       strings.NewReader("node a [kind=\"service\"]\n"),
		JSON:        true,
		Stdout:      &out,
	})
	got := out.String()
	for _, want := range []string{`"outcome": "non-conforming"`, `"definition": "Service"`, `"element": "a"`} {
		if !strings.Contains(got, want) {
			t.Errorf("JSON output missing %q:\n%s", want, got)
		}
	}
}

func TestProject(t *testing.T) {
	var out bytes.Buffer
	err := runProject(&projectConfig{
		DialectFile: writeTemp(t, "d.yaml", validDialect),
		Stdin:       strings.NewReader("node team\nnode a [kind=\"service\", name=\"A\", owner=team]\n"),
		Stdout:      &out,
	})
	if err != nil {
		t.Fatalf("runProject: %v", err)
	}
	got := out.String()
	for _, want := range []string{`"Service"`, `"a"`, `"$gslRef": "team"`} {
		if !strings.Contains(got, want) {
			t.Errorf("projection output missing %q:\n%s", want, got)
		}
	}
}
