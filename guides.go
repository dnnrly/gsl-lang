package gsl

import "embed"

// Guides embeds the AI/LLM guide files
//
//go:embed GSL_GUIDE.md GQL_GUIDE.md GO_REFERENCE.md
var Guides embed.FS
