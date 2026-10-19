// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Pascal Dennerly.
package guides

import "embed"

// Guides embeds the AI/LLM guide files distributed with the reference CLI
// tools. They live here (rather than in the language implementation) because
// they are language-neutral documentation assets that the command-line tools
// print via the `ai` subcommand; //go:embed cannot reference parent
// directories, so the files must sit alongside this package.
//
//go:embed GSL_GUIDE.md GQL_GUIDE.md GO_REFERENCE.md QUERY_TUTORIAL.md SEQUENCE_GUIDE.md DIALECT_GUIDE.md
var Guides embed.FS
