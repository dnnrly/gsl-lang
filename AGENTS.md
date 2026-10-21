<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# AGENTS.md - GSL-Lang Development Guide

This document provides instructions for AI agents working on the GSL-Lang project.

**For comprehensive language specification, see [spec/gsl/SPEC.md](spec/gsl/SPEC.md).**  
**For formal grammar, see [spec/gsl/GRAMMAR.md](spec/gsl/GRAMMAR.md).**  
**For a description of the query language, see [spec/gql/QUERY_SPEC.md](spec/gql/QUERY_SPEC.md) and [spec/gql/QUERY_GRAMMAR.md](spec/gql/QUERY_GRAMMAR.md).**
**LLM specific advice, see [GSL_GUIDE.md](guides/GSL_GUIDE.md) and [GQL_GUIDE.md](guides/GQL_GUIDE.md).**
**See [GO_REFERENCE.md](guides/GO_REFERENCE.md) for Go reference implementation patterns and algorithms.**
**For a quick LLM-oriented overview, start with [llms.txt](llms.txt).**
**The repository hosts multiple implementations under [implementations/](implementations/); Go is the canonical, reference implementation.**

## Quick Commands

```bash
make test                       # Run all Go tests with coverage (query fixtures run silently; use `go test -v -run TestFixtures ./implementations/go/query` to watch them)
make test-typescript            # Run the TypeScript workspace tests (installs deps; skips if npm missing)
make test-all                   # Run every test suite: Go, TypeScript, integration, acceptance
make test-integration           # Run integration tests (skip if tools missing)
make test-integration-strict    # Run integration tests (fail if tools missing)
make test-acceptance            # Run acceptance tests (BDD/godog feature tests)
make lint                       # Run linting
make fuzz                       # Run fuzz tests
make build                      # Build CLI tools (gsl-diagram, gsl-query, gsl-dialect, gsl-lsp)
make diagrams                   # Re-render README diagram SVGs from .mmd sources (needs mmdc + python3)
make clean                      # Clean build artifacts
go test -v -run TestName        # Run specific test
go test -v -run 'TestAgentSkill|TestSkillBehaviour' ./tests   # Validate the gsl-modelling agent skill
```

## Project Structure

The repository is organised so that language definitions, implementations and
shared test material are independent:

- `spec/` — language-neutral normative specifications (`spec/gsl/` for GSL,
  `spec/gql/` for GQL, `spec/dialects/` for dialect documents).
- `implementations/` — one directory per implementation (`implementations/go/`
  is the canonical reference; `implementations/typescript/` is an npm workspace
  with GSL and GQL packages — see below).
- `guides/` — explanatory, LLM/agent-oriented guides (embedded by the CLI `ai`
  command).
- `tests/` — repository meta-tests and the shared, language-neutral conformance
  suite.
- `cmd/`, `lsp/`, `editors/`, `examples/`, `docs/` — the CLI tools, language
  server, editor extension, examples and human documentation.

**Go Implementation (`implementations/go/`, package `gsl`):**
- `lexer.go`, `parser.go`, `ast.go` - Parsing pipeline
- `model.go`, `build.go` - Graph construction
- `serialize.go` - Serialization to canonical form
- `gsl.go` - Public API
- `query/` - the GQL library (package `query`)

**TypeScript Workspace (`implementations/typescript/`):**
- An npm workspace with three packages: `@gsl-lang/core` (GSL parser, graph
  build, canonical serializer), `@gsl-lang/gql` (GQL parser and query engine)
  and `@gsl-lang/dialect` (dialect loading, matching, projection and JSON
  Schema validation, using `ajv` and `yaml`). `private: true`, unpublished.
  Build/typecheck/test via `npm install`, `npm run typecheck`, `npm test`. No
  CLI or LSP.
- Tests mirror the shared corpora: the GQL conformance fixtures
  (`tests/conformance/gql`), the dialect conformance fixtures
  (`tests/conformance/dialects`), the `examples/` round-trip corpus, the
  flagship/advanced model+query fixtures, and the markdown `gsl`/`gql` blocks
  (`packages/{gql,dialect}/test/*`).

**Tests:**
- `implementations/go/*_test.go` - Unit and fuzz tests for the Go library
- `implementations/go/query/*_test.go` - GQL library tests and fixtures
- `tests/conformance/` - shared, language-neutral conformance fixtures (`gql/`)
- `tests/` - repository meta-tests: `markdown_test.go` (validates all code
  blocks in `.md` files), `agent_skill_test.go`, and `harness.go`/`main_test.go`
- `cmd/gsl-diagram/cli_integration_test.go` - CLI tool integration tests (requires `mmdc` and `plantuml`)
- `test/` - Acceptance tests using godog (BDD/Gherkin feature files)

**CLI Tools:**
- `cmd/gsl-diagram/` - Converts GSL graphs to diagram formats (Mermaid, PlantUML)
- `cmd/gsl-query/` - Runs GSL queries from the command line
- `cmd/gsl-dialect/` - Validates GSL graphs against dialect documents and projects attributes to JSON (exit 0 conforming, 1 non-conforming, 2 incomplete)
- `cmd/gsl-lsp/` - LSP server binary entry point

**LSP Server:**
- `lsp/server.go` — Server struct, lifecycle, document management
- `lsp/dispatch.go` — Parse routing, GSL/GQL diagnostic handling
- `lsp/handler.go` — LSP feature handlers (completion, hover, definition, etc.)
- `lsp/gsl.go` — GSL-specific features (completions, symbols, semantic tokens)
- `lsp/gql.go` — GQL-specific features (keyword completions, hover, symbols, tokens)
- `lsp/helpers.go` — Shared utilities (wordAt, tokenize, lineColToPos)

**VS Code Extension:**
- `editors/vscode/` — VS Code extension for GSL/GQL language support
- `editors/vscode/extension.js` — Extension entry point, binary discovery, client setup
- `editors/vscode/package.json` — Extension manifest
- `editors/vscode/syntaxes/` — TextMate grammars for GSL and GQL

**Examples:**
- `examples/flagships/` - the five flagship narratives (01 many-views, 02 release-prerequisites, 03 architecture-across-boundaries, 04 financial-network, 05 llm-assisted-modelling), each `model.gsl` + `README.md` + `qN-*.gql`/`.result.gsl` + `views/`
- `examples/advanced/architecture-archaeology/` - the advanced study (one-system + many-systems), same layout as a flagship
- `examples/flagship_test.go` - flagship model/query/result/diagram/README validation
- `examples/advanced_test.go` - the same guarantees for `advanced/`

**Documentation:**
- `spec/gsl/SPEC.md` - Normative spec (source of truth for language rules)
- `spec/gsl/GRAMMAR.md` - Formal grammar
- `spec/gql/QUERY_SPEC.md`, `spec/gql/QUERY_GRAMMAR.md` - GQL specification and grammar
- `spec/dialects/DIALECTS_SPEC.md`, `spec/dialects/SUPPORTED_GQL_SUBSET.md` - Dialect specification and the GQL subset available in dialect match queries
- `guides/` - explanatory guides embedded by the CLI `ai` command (`GSL_GUIDE.md`, `GQL_GUIDE.md`, `GO_REFERENCE.md`, `QUERY_TUTORIAL.md`, `SEQUENCE_GUIDE.md`, `DIALECT_GUIDE.md`)
- `README.md` - User-facing examples and storefront
- `docs/` - Human learning journey (the IA index is `docs/README.md`): `getting-started/`, `concepts/` (nine one-concept pages: graph-model, edges, attributes, sets, structure-and-nesting, edge-dependencies, queries-and-views, parsing-and-merging, canonical-form), `tutorials/` (modelling-with-gsl + path index), `cookbook/` (seven recipes). The specs and guides live in `spec/` and `guides/`; docs/ tiers reference them rather than duplicate them.
- `tests/markdown_test.go` - Validates all code blocks in root `*.md` and every `docs/**/*.md`, `guides/**/*.md`, `spec/**/*.md`, `implementations/**/*.md` recursively (via `findMarkdownFiles`; note that `filepath.Glob` does not support `**`)

**Agent Skills:**
- `.agents/skills/gsl-modelling/` - the `gsl-modelling` agent skill: source-faithful GSL modelling for agents (opencode/Cursor/Codex discover it natively; Claude Code needs a `.claude/skills` pointer). `SKILL.md` is the entry point; `references/` mixes authored method docs (`MODELLING.md`, `TOOLING.md`, `EXAMPLES.md`) with plain copies of the authoritative guides (`GSL_GUIDE.md`, `GQL_GUIDE.md`, `GRAMMAR.md`, `QUERY_GRAMMAR.md`, `modelling-with-gsl.md`) - copies are maintained upstream and refreshed from there (never symlinks: they do not survive packaging). `scripts/validate.sh` is an optional structural gate (`gsl-query ""`).
- `.agents/skills/gsl-dialect-authoring/` - the `gsl-dialect-authoring` agent skill: authoring/reviewing GSL dialect documents and checking models against them. `SKILL.md` is the entry point; `references/` mixes an authored method doc (`WORKFLOW.md`) with plain copies of `DIALECT_GUIDE.md`, `DIALECTS_SPEC.md` and `SUPPORTED_GQL_SUBSET.md` (refreshed the same way). `scripts/validate.sh` is an optional structural gate (`gsl-dialect validate`).
- `tests/agent_skill_test.go` - Validates every discovered skill (structure/frontmatter, reference mirror drift, reference link resolution, authored markdown code blocks, script syntax) plus the behaviour oracle. Skills and mirror maps are registered in the test; new skills are discovered automatically. Note: dot-directories are NOT scanned by `markdown_test.go`; the reference mirrors re-validate the same `gsl`/`gql` blocks as upstream. Refresh mirrors with `go test ./tests -run TestAgentSkillMirrorDrift -update`.
- `examples/skill-behaviour/` - Behavioural fixtures for the skill: a deliberately misleading source + an expected-fidelity oracle for the (human-reviewed) smoke exercise.

## Before Submitting Changes

```bash
make test
make test-integration           # if mmdc and plantuml available
make test-integration-strict    # to enforce tool availability
go test -v -run TestMarkdownCodeBlocks ./tests
make lint
```

- All `gsl` code blocks in markdown (root and `docs/`, `guides/`, `spec/`) must parse
- All `invalid-gsl` blocks must fail to parse
- If any mirrored reference guide (GSL_GUIDE, GQL_GUIDE, GRAMMAR, QUERY_GRAMMAR, modelling-with-gsl) is changed, refresh with `go test -run TestAgentSkillMirrorDrift -update ./tests` before committing
- Round-trip tests required for parser/serializer changes
- Integration tests validate all example GSL files with both Mermaid and PlantUML converters
- Update `spec/gsl/SPEC.md` if changing language semantics
- Update `spec/gsl/GRAMMAR.md` if changing syntax
- Add or update `docs/` pages if changing the learning narrative; keep docs/ errors-free so `TestMarkdownCodeBlocks` stays green
- Commit message should reference relevant files

## Task Reference

| Task | Files |
|------|-------|
| Add new syntax | implementations/go/{token,lexer,parser,ast}.go, implementations/go/parser_test.go, spec/gsl/GRAMMAR.md |
| Fix parsing bug | implementations/go/parser.go, implementations/go/parser_test.go, implementations/go/gsl_test.go |
| Serialization | implementations/go/serialize.go, implementations/go/serialize_test.go, implementations/go/gsl_test.go |
| Graph structure | implementations/go/model.go, implementations/go/build.go, implementations/go/build_test.go |
| Language rules | spec/gsl/SPEC.md, spec/gsl/GRAMMAR.md, tests/markdown_test.go |
| CLI diagram tool | cmd/gsl-diagram/*, cli_integration_test.go |
| CLI query tool | cmd/gsl-query/* |
| CLI dialect tool | cmd/gsl-dialect/*, implementations/go/dialect/* |
| LSP server | lsp/*, cmd/gsl-lsp/main.go |
| VS Code extension | editors/vscode/* |
| TypeScript workspace | implementations/typescript/{package,tsconfig}.json, implementations/typescript/packages/{core,gql,dialect}/* |
| Documentation examples | README.md, spec/gsl/SPEC.md, guides/GSL_GUIDE.md |
| Flagship/advanced examples | examples/flagship_test.go, examples/advanced_test.go, examples/flagships/*, examples/advanced/* |
| Documentation learning journey | docs/README.md (IA index), docs/getting-started/, docs/concepts/, docs/tutorials/ (+ modelling-with-gsl.md), docs/cookbook/ |
| Query language tests | tests/conformance/gql/*, tests/conformance/gql/README.md |
| Dialect validation | spec/dialects/*, implementations/go/dialect/*, cmd/gsl-dialect/*, tests/conformance/dialects/*, examples/dialects/*, guides/DIALECT_GUIDE.md |
| Dialect conformance tests | tests/conformance/dialects/* (runner: implementations/go/dialect/fixtures_test.go) |
| Acceptance tests | test/features/*.feature, test/*_test.go |
| Agent skill | .agents/skills/gsl-modelling/* (authored: SKILL.md, references/{MODELLING,TOOLING,EXAMPLES}.md and mirrored guides, scripts/validate.sh), .agents/skills/gsl-dialect-authoring/* (authored: SKILL.md, references/WORKFLOW.md and mirrored guides, scripts/validate.sh), tests/agent_skill_test.go (mirror drift: TestAgentSkillMirrorDrift), examples/skill-behaviour/* |

## Planning & Progress Tracking

Subsystems may have planning/progress files for organizing work across multiple agents or sessions.

**Pattern:**
Each subsystem may contain:
- `.test-plan.md` — Comprehensive plan with all proposed fixtures/work items
  - Organized by priority or phase
  - Includes spec references, complexity estimates, descriptions
  - Use to identify what to work on next
  
- `.progress.md` — Tracking checklist
  - Mark items as complete as you implement them
  - Update across sessions to maintain continuity
  - Organized to match the plan
  
- `.quick-ref.md` — Implementation reference guide
  - Templates and examples for the subsystem
  - Common commands and patterns
  - Quick syntax/structure examples

**Rules for planning files:**
- All plan/progress/reference files MUST be listed in `.gitignore` for their directory
- These files are **never committed** to git — they are local development aids only
- Actual artifacts (test fixtures, code, documentation) are committed normally
- Plan files help agents coordinate work across sessions without cluttering git history

**When working with planning files:**
1. Read the `.test-plan.md` to understand the scope and priorities
2. Consult `.quick-ref.md` for implementation patterns and examples
3. Update `.progress.md` as you complete work items
4. Create/commit only the actual artifacts (tests, code, docs) — never the plan files
5. Verify plan files are gitignored before finishing

## Test Fixture Documentation

**Query Language Fixtures** — `tests/conformance/gql/README.md`

To avoid context bloat when working with many test fixtures:
- **All fixtures cataloged** in a single README with quick navigation table
- **Organized as a learning path** (numbered groups: basics → predicates → make → remove → traversal → edge-dependencies → collapse → named-graphs → pipelines → edge-cases)
- **Semantic notes** explain correct behavior for each category
- **Maintenance checklist** included so agents know when to update

**When adding/modifying query fixtures:**
1. Create fixture directory with `graph.gsl`, `query.gql`, `result.gsl`
2. Place it in the appropriate numbered group under `tests/conformance/gql/`
3. Verify it passes: `go test -v -run TestFixtures ./implementations/go/query`
4. Update `tests/conformance/gql/README.md`:
   - Add entry to appropriate group table
   - Include brief description (one line)
   - Add semantic note if behavior is non-obvious
5. If adding a new group, create a new numbered directory and add it to the learning path
6. Follow maintenance checklist at end of README

## Key Design Patterns

- **Lenient parsing**: Warnings are non-fatal; invalid input generates errors
- **Last-write-wins**: Multiple declarations merge with attribute conflicts resolved by last occurrence
- **Canonical form**: `parse(serialize(parse(x))) == parse(x)` is a guarantee
- **No schema validation**: Graph properties (cycles, tree validity) are not enforced
- **Attributes are untyped**: Stored as `interface{}`; requires type assertion

## Review Checklist

Before submitting changes:
- [ ] All tests pass: `make test`
- [ ] Integration tests pass: `make test-integration` (or `make test-integration-strict`)
- [ ] Lint passes: `make lint`
- [ ] Markdown validates: `go test -v -run TestMarkdownCodeBlocks ./tests`
- [ ] Round-trip tests added (if touching parser/serializer)
- [ ] Example fixture (`.gsl` file in `examples/`) added for serialization round-trip fixes
- [ ] Integration tests added (if modifying CLI tools or converters)
- [ ] LSP tests pass: `go test ./lsp/...`
- [ ] LSP file structure updated in this file if reorganizing lsp/
- [ ] spec/gsl/SPEC.md updated (if changing language semantics)
- [ ] spec/gsl/GRAMMAR.md updated (if changing syntax)
- [ ] New code examples in markdown are valid `gsl` blocks
- [ ] Commit message references relevant files

## Refactoring and tool use

Where available, the `gopls` command can be used to:
- Format code: `gopls format`
- Analyze code: `gopls analyze`
- Generate code: `gopls generate`
- Inspect code: `gopls inspect`
- Rename identifiers: `gopls rename`

More details about `gopls` can be found by running `gopls help`.

## Notes

- Parser is hand-written, not generated
- Standard library only (no external dependencies)
- Edge order is preserved; duplicates allowed
- Set membership is O(1) via map lookup
