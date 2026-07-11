//go:build acceptance

package test_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
)

func TestMain(m *testing.M) {
	// Build binaries before running tests
	if err := buildBinaries(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build binaries: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format: "pretty",
			Paths:  []string{"features"},
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	tc, err := newTestContext()
	if err != nil {
		panic(fmt.Sprintf("failed to create test context: %v", err))
	}

	ctx.BeforeScenario(func(*godog.Scenario) {
		tc.reset()
	})

	ctx.AfterScenario(func(s *godog.Scenario, err error) {
		if err != nil {
			fmt.Printf(
				"Command line output for \"%s\"\nUsing parameters: %s\n%s",
				s.Name,
				tc.cmdParams,
				tc.cmdOutput,
			)
		}
		tc.cleanup()
	})

	// gsl-diagram steps
	ctx.Step(`^the gsl-diagram binary is available$`, tc.theGslDiagramBinaryIsAvailable)
	ctx.Step(`^I run gsl-diagram with parameters "(.*)"`, tc.iRunGslDiagramWithParameters)

	// gsl-query steps
	ctx.Step(`^the gsl-query binary is available$`, tc.theGslQueryBinaryIsAvailable)
	ctx.Step(`^I run gsl-query with parameters "(.*)"`, tc.iRunGslQueryWithParameters)
	ctx.Step(`^I run gsl-query with query "(.*)" using input file$`, tc.iRunGslQueryWithQueryUsingInputFile)

	// Shared steps
	ctx.Step(`^the app exits without error$`, tc.theAppExitsWithoutError)
	ctx.Step(`^the app exits with an error$`, tc.theAppExitsWithAnError)
	ctx.Step(`^the app output contains "(.*)"$`, tc.theAppOutputContains)
	ctx.Step(`^a GSL input file with content:$`, tc.aGslInputFileWithContent)
}
