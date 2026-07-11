# AGENTS.md - GSL-Lang Development Guide

This document provides instructions for AI agents working on the GSL-Lang project.

**For comprehensive language specification, see [SPEC.md](SPEC.md).**  
**For formal grammar, see [GRAMMAR.md](GRAMMAR.md).**  
**For a description of the query language, see [QUERY_SPEC.md](QUERY_SPEC.md) and [QUERY_GRAMMAR.md](QUERY_GRAMMAR.md).**
**LLM specific advice, see [GSL_GUIDE.md](GSL_GUIDE.md) and [GQL_GUIDE.md](GQL_GUIDE.md).**
**See [GO_REFERENCE.md](GO_REFERENCE.md) for Go reference implementation patterns and algorithms.**
**For a quick LLM-oriented overview, start with [llms.txt](llms.txt) or the ["For LLMs and AI Agents"](README.md#for-llms-and-ai-agents) section in README.md.**

## Quick Commands

```bash
make test                       # Run all tests with coverage (query fixtures run silently; use `go test -v -run TestFixtures ./query` to watch them)
make test-integration           # Run integration tests (skip if tools missing)
make test-integration-strict    # Run integration tests (fail if tools missing)
make test-acceptance            # Run acceptance tests (BDD/godog feature tests)
make lint                       # Run linting
make fuzz                       # Run fuzz tests
make build                      # Build CLI tools (gsl-diagram, gsl-query)
make clean                      # Clean build artifacts
go test -v -run TestName        # Run specific test
```

## Project Structure

**Core Implementation:**
- `lexer.go`, `parser.go`, `ast.go` - Parsing pipeline
- `model.go`, `build.go` - Graph construction
- `serialize.go` - Serialization to canonical form
- `gsl.go` - Public API

**Tests:**
- `*_test.go` files - Unit and integration tests
- `markdown_test.go` - Validates all code blocks in `.md` files
- `cmd/gsl-diagram/cli_integration_test.go` - CLI tool integration tests (requires `mmdc` and `plantuml`)
- `test/` - Acceptance tests using godog (BDD/Gherkin feature files)

**CLI Tools:**
- `cmd/gsl-diagram/` - Converts GSL graphs to diagram formats (Mermaid, PlantUML)
- `cmd/gsl-query/` - Runs GSL queries from the command line
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

**Documentation:**
- `SPEC.md` - Normative spec (source of truth for language rules)
- `GSL_GUIDE.md` - GSL language reference for LLMs and AI agents
- `README.md` - User-facing examples
- `GRAMMAR.md` - Formal grammar

## Before Submitting Changes

```bash
make test
make test-integration           # if mmdc and plantuml available
make test-integration-strict    # to enforce tool availability
go test -v -run TestMarkdownCodeBlocks
make lint
```

- All `gsl` code blocks in markdown must parse
- All `invalid-gsl` blocks must fail to parse
- Round-trip tests required for parser/serializer changes
- Integration tests validate all example GSL files with both Mermaid and PlantUML converters
- Update SPEC.md if changing language semantics
- Update GRAMMAR.md if changing syntax
- Commit message should reference relevant files

## Task Reference

| Task | Files |
|------|-------|
| Add new syntax | token.go, lexer.go, parser.go, ast.go, parser_test.go, GRAMMAR.md |
| Fix parsing bug | parser.go, parser_test.go, gsl_test.go |
| Serialization | serialize.go, serialize_test.go, gsl_test.go |
| Graph structure | model.go, build.go, build_test.go |
| Language rules | SPEC.md, GRAMMAR.md, markdown_test.go |
| CLI diagram tool | cmd/gsl-diagram/*, cli_integration_test.go |
| CLI query tool | cmd/gsl-query/* |
| LSP server | lsp/*, cmd/gsl-lsp/main.go |
| VS Code extension | editors/vscode/* |
| Documentation examples | README.md, SPEC.md, GSL_GUIDE.md |
| Query language tests | query/testdata/*, query/.test-plan.md |
| Acceptance tests | test/features/*.feature, test/*_test.go |

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

**Query Language Fixtures** — `query/testdata/README.md`

To avoid context bloat when working with many test fixtures:
- **All fixtures cataloged** in a single README with quick navigation table
- **Organized as a learning path** (numbered groups: basics → predicates → make → remove → traversal → edge-dependencies → collapse → named-graphs → pipelines → edge-cases)
- **Semantic notes** explain correct behavior for each category
- **Maintenance checklist** included so agents know when to update

**When adding/modifying query fixtures:**
1. Create fixture directory with `graph.gsl`, `query.gql`, `result.gsl`
2. Place it in the appropriate numbered group under `query/testdata/`
3. Verify it passes: `go test -v -run TestFixtures ./query`
4. Update `query/testdata/README.md`:
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
- [ ] Markdown validates: `go test -v -run TestMarkdownCodeBlocks`
- [ ] Round-trip tests added (if touching parser/serializer)
- [ ] Example fixture (`.gsl` file in `examples/`) added for serialization round-trip fixes
- [ ] Integration tests added (if modifying CLI tools or converters)
- [ ] LSP tests pass: `go test ./lsp/...`
- [ ] LSP file structure updated in this file if reorganizing lsp/
- [ ] SPEC.md updated (if changing language semantics)
- [ ] GRAMMAR.md updated (if changing syntax)
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
