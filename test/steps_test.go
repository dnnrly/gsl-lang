//go:build acceptance

package test_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cucumber/godog"
)

type testContext struct {
	binDir    string
	cmdParams string
	cmdOutput string
	cmdErr    error
	tmpFiles  []string
}

func newTestContext() (*testContext, error) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("unable to determine test file location")
	}
	testDir := filepath.Dir(testFile)
	rootDir := filepath.Join(testDir, "..")
	binDir := filepath.Join(rootDir, "tmp")

	return &testContext{binDir: binDir}, nil
}

func (tc *testContext) reset() {
	tc.cmdParams = ""
	tc.cmdOutput = ""
	tc.cmdErr = nil
}

func (tc *testContext) cleanup() {
	for _, f := range tc.tmpFiles {
		os.Remove(f)
	}
	tc.tmpFiles = nil
}

func (tc *testContext) createTempFile(suffix, content string) (string, error) {
	f, err := os.CreateTemp("", "acceptance-*"+suffix)
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		return "", fmt.Errorf("writing temp file: %w", err)
	}
	f.Close()
	tc.tmpFiles = append(tc.tmpFiles, f.Name())
	return f.Name(), nil
}

func (tc *testContext) theGslDiagramBinaryIsAvailable() error {
	bin := filepath.Join(tc.binDir, "gsl-diagram")
	if _, err := os.Stat(bin); err != nil {
		return fmt.Errorf("gsl-diagram binary not found at %s: %w", bin, err)
	}
	return nil
}

func (tc *testContext) theGslQueryBinaryIsAvailable() error {
	bin := filepath.Join(tc.binDir, "gsl-query")
	if _, err := os.Stat(bin); err != nil {
		return fmt.Errorf("gsl-query binary not found at %s: %w", bin, err)
	}
	return nil
}

func (tc *testContext) iRunGslDiagramWithParameters(params string) error {
	return tc.runBinary("gsl-diagram", params)
}

func (tc *testContext) iRunGslQueryWithParameters(params string) error {
	return tc.runBinary("gsl-query", params)
}

func (tc *testContext) iRunGslQueryWithQueryUsingInputFile(query string) error {
	inputFile := tc.findTempFile(".gsl")
	bin := filepath.Join(tc.binDir, "gsl-query")
	cmd := exec.Command(bin, "--input", inputFile, query)
	output, err := cmd.CombinedOutput()
	tc.cmdOutput = string(output)
	tc.cmdErr = err
	return nil
}

func (tc *testContext) runBinary(name, params string) error {
	tc.cmdParams = params

	// Replace <input_file> placeholder with actual temp file path
	params = strings.ReplaceAll(params, "<input_file>", tc.findTempFile(".gsl"))

	args := strings.Fields(params)
	bin := filepath.Join(tc.binDir, name)

	cmd := exec.Command(bin, args...)
	output, err := cmd.CombinedOutput()
	tc.cmdOutput = string(output)
	tc.cmdErr = err

	return nil
}

func (tc *testContext) findTempFile(suffix string) string {
	for _, f := range tc.tmpFiles {
		if strings.HasSuffix(f, suffix) {
			return f
		}
	}
	return ""
}

func (tc *testContext) theAppExitsWithoutError() error {
	if tc.cmdErr != nil {
		return fmt.Errorf("expected no error, got: %v\nOutput: %s", tc.cmdErr, tc.cmdOutput)
	}
	return nil
}

func (tc *testContext) theAppExitsWithAnError() error {
	if tc.cmdErr == nil {
		return fmt.Errorf("expected an error, but command succeeded\nOutput: %s", tc.cmdOutput)
	}
	return nil
}

func (tc *testContext) theAppOutputContains(expected string) error {
	if !strings.Contains(tc.cmdOutput, expected) {
		return fmt.Errorf("expected output to contain %q, got:\n%s", expected, tc.cmdOutput)
	}
	return nil
}

func (tc *testContext) aGslInputFileWithContent(docString *godog.DocString) error {
	content := docString.Content
	_, err := tc.createTempFile(".gsl", content)
	return err
}
