package gsl

import "embed"

// Guides embeds the AI/LLM guide files
//
//go:embed LLM_GUIDE.md QUERY_AI_GUIDE.md GO_GUIDE.md
var Guides embed.FS
