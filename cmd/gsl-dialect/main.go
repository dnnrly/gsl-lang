// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.

// Command gsl-dialect validates GSL graphs against dialect documents and
// projects graph attributes into the JSON instances a dialect validates.
//
// Exit codes (see README.md):
//
//	0  the graph conforms to the dialect
//	1  the graph does not conform (validation completed with violations)
//	2  validation could not be completed (invalid dialect, matching error,
//	   I/O or usage error)
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/dnnrly/gsl-lang/cmd/internal/cli"
	"github.com/dnnrly/gsl-lang/guides"
	"github.com/spf13/cobra"
)

// Build information injected by goreleaser.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		var ee *exitError
		if errors.As(err, &ee) {
			if ee.err != nil {
				fmt.Fprintln(os.Stderr, ee.err)
			}
			os.Exit(ee.code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitIncomplete)
	}
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "gsl-dialect",
		Short:         "Validate GSL graphs against dialect documents",
		Long:          "Validate GSL graphs against dialect documents and project graph attributes into the JSON instances a dialect validates.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var (
		validateDialect string
		validateInput   string
		validateJSON    bool
	)

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a graph against a dialect",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(&validateConfig{
				DialectFile: validateDialect,
				InputFile:   validateInput,
				JSON:        validateJSON,
				Stdin:       cmd.InOrStdin(),
				Stdout:      cmd.OutOrStdout(),
				Stderr:      cmd.ErrOrStderr(),
			})
		},
	}
	validateCmd.Flags().StringVarP(&validateDialect, "dialect", "d", "", "Dialect document (required)")
	validateCmd.Flags().StringVarP(&validateInput, "input", "i", "", "Input GSL file (read from stdin if not provided)")
	validateCmd.Flags().BoolVar(&validateJSON, "json", false, "Emit the validation report as JSON (provisional)")
	_ = validateCmd.MarkFlagRequired("dialect")

	var (
		projectDialect string
		projectInput   string
	)

	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "Project the attributes of matched elements to JSON",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProject(&projectConfig{
				DialectFile: projectDialect,
				InputFile:   projectInput,
				Stdin:       cmd.InOrStdin(),
				Stdout:      cmd.OutOrStdout(),
			})
		},
	}
	projectCmd.Flags().StringVarP(&projectDialect, "dialect", "d", "", "Dialect document (required)")
	projectCmd.Flags().StringVarP(&projectInput, "input", "i", "", "Input GSL file (read from stdin if not provided)")
	_ = projectCmd.MarkFlagRequired("dialect")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cli.PrintVersion("gsl-dialect", Version, Commit, BuildDate)
		},
	}

	rootCmd.AddCommand(validateCmd, projectCmd, versionCmd)

	dialectDesc, dialectContent, dialectErr := cli.LoadGuide(guides.Guides, "DIALECT_GUIDE.md")
	cli.BuildAICommand(rootCmd, []cli.GuideSpec{
		{Use: "dialect", Desc: dialectDesc, Content: dialectContent, Err: dialectErr},
	}, "Available AI/LLM guides:", "Run 'gsl-dialect ai dialect' to view the guide")

	return rootCmd
}
