package gsl

import (
	"fmt"
	"strings"
)

// ParseError represents a structured error from parsing GSL code.
// It contains a fatal error (if any) and non-fatal warnings discovered during parsing.
type ParseError struct {
	Message  string  // Main error message (empty if no fatal error)
	Warnings []error // Non-fatal parse issues
	Err      error   // Fatal error, if any
}

// Error implements the error interface for ParseError.
func (pe *ParseError) Error() string {
	if pe == nil {
		return ""
	}

	var parts []string

	// Add fatal error if present
	if pe.Err != nil {
		parts = append(parts, pe.Err.Error())
	} else if pe.Message != "" {
		parts = append(parts, pe.Message)
	}

	// Add warnings summary
	if len(pe.Warnings) > 0 {
		warningMsgs := make([]string, len(pe.Warnings))
		for i, w := range pe.Warnings {
			warningMsgs[i] = w.Error()
		}
		parts = append(parts, fmt.Sprintf("warnings: %s", strings.Join(warningMsgs, "; ")))
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "; ")
}

// HasError returns true if there was a fatal error.
func (pe *ParseError) HasError() bool {
	return pe != nil && pe.Err != nil
}

// HasWarnings returns true if there were any non-fatal warnings.
func (pe *ParseError) HasWarnings() bool {
	return pe != nil && len(pe.Warnings) > 0
}
