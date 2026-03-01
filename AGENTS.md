# AGENTS.md - GSL-Lang Development Guide

This document provides instructions for AI agents working on the GSL-Lang project.

**For comprehensive language specification, see [SPEC.md](SPEC.md).**  
**For formal grammar, see [GRAMMAR.md](GRAMMAR.md).**  
**For Go API patterns and algorithms, see [LLM_GUIDE.md](LLM_GUIDE.md).**

## Quick Commands

```bash
make test                # Run all tests with coverage
make lint                # Run linting
make fuzz                # Run fuzz tests
make clean               # Clean build artifacts
go test -v -run TestName # Run specific test
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

**Documentation:**
- `SPEC.md` - Normative spec (source of truth for language rules)
- `LLM_GUIDE.md` - Go API patterns and algorithms aimed specifically at LLMs and AI agents
- `README.md` - User-facing examples
- `GRAMMAR.md` - Formal grammar

## Before Submitting Changes

```bash
make test
go test -v -run TestMarkdownCodeBlocks
make lint
```

- All `gsl` code blocks in markdown must parse
- All `invalid-gsl` blocks must fail to parse
- Round-trip tests required for parser/serializer changes
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
| Documentation examples | README.md, SPEC.md, LLM_GUIDE.md |

## Key Design Patterns

- **Lenient parsing**: Warnings are non-fatal; invalid input generates errors
- **Last-write-wins**: Multiple declarations merge with attribute conflicts resolved by last occurrence
- **Canonical form**: `parse(serialize(parse(x))) == parse(x)` is a guarantee
- **No schema validation**: Graph properties (cycles, tree validity) are not enforced
- **Attributes are untyped**: Stored as `interface{}`; requires type assertion

## Review Checklist

Before submitting changes:
- [ ] All tests pass: `make test`
- [ ] Lint passes: `make lint`
- [ ] Markdown validates: `go test -v -run TestMarkdownCodeBlocks`
- [ ] Round-trip tests added (if touching parser/serializer)
- [ ] SPEC.md updated (if changing language semantics)
- [ ] GRAMMAR.md updated (if changing syntax)
- [ ] New code examples in markdown are valid `gsl` blocks
- [ ] Commit message references relevant files

## Notes

- Parser is hand-written, not generated
- Standard library only (no external dependencies)
- Edge order is preserved; duplicates allowed
- Set membership is O(1) via map lookup
