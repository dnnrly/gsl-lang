// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.
package tests

import (
	"os"
	"testing"
)

// TestMain makes repository-relative paths work regardless of where the test
// binary is invoked from by chdir-ing to the repository root.
func TestMain(m *testing.M) {
	root, err := repoRoot()
	if err != nil {
		os.Exit(1)
	}
	if err := os.Chdir(root); err != nil {
		os.Exit(1)
	}
	os.Exit(m.Run())
}
