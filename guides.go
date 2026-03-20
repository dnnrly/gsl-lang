package gsl

import "embed"

// Guides embeds the AI/LLM guide files
//go:embed LLM_GUIDE.md query/QUERY_AI_GUIDE.md
var Guides embed.FS
