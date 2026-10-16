// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

// Package tests holds repository-level validation tests that span more than one
// implementation: the markdown code-block validator and the agent-skill
// contract tests. Shared conformance fixtures live alongside it under
// conformance/ and are executed by the individual implementations.
//
// The tests are written against repository-relative paths; TestMain (in
// main_test.go) chdirs to the repository root before running.
package tests

import (
	"os"
	"path/filepath"
)

// repoRoot returns the absolute path of the repository root by walking up from
// the current directory until a go.mod is found.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
