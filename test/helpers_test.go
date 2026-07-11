//go:build acceptance

package test_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

func buildBinaries() error {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("unable to determine test file location")
	}
	testDir := filepath.Dir(testFile)
	rootDir := filepath.Join(testDir, "..")
	binDir := filepath.Join(rootDir, "tmp")

	// Build gsl-diagram
	cmd := exec.Command("go", "build", "-o", filepath.Join(binDir, "gsl-diagram"), "./cmd/gsl-diagram")
	cmd.Dir = rootDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("building gsl-diagram: %w\nOutput: %s", err, string(output))
	}

	// Build gsl-query
	cmd = exec.Command("go", "build", "-o", filepath.Join(binDir, "gsl-query"), "./cmd/gsl-query")
	cmd.Dir = rootDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("building gsl-query: %w\nOutput: %s", err, string(output))
	}

	return nil
}
