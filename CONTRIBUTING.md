# Contributing to GSL

Thanks for considering contributing to GSL. This guide covers how to work
with the repository, the supported build and test toolchain, and the
versioning and compatibility policy that keeps GSL reproducible.

## Repository layout

- `lexer.go`, `parser.go`, `ast.go` — parsing pipeline
- `model.go`, `build.go`, `serialize.go`, `gsl.go` — graph model and public API
- `query/` — the GSL Query Language (GQL) parser and execution engine
- `cmd/` — CLI tools (`gsl-diagram`, `gsl-query`, `gsl-lsp`)
- `lsp/` — the language server used by `gsl-lsp`
- `editors/vscode/` — the VS Code extension (grammars, client)
- `examples/` — flagship examples used by the integration and round-trip tests
- `docs/` — the learning journey (concepts, tutorials, cookbook)
- `test/` — acceptance tests (godog/Gherkin feature files)

## Prerequisites

- Go at the version declared in `go.mod` (currently 1.26).
- For integration tests (`make test-integration-strict`): `mmdc`
  (mermaid-cli) and `plantuml` on `PATH`.
- For acceptance tests (`make test-acceptance`): nothing extra; the make
  target builds the CLI binaries it needs.

## Quick commands

```bash
make test                       # unit + fixture + round-trip + markdown tests
make test-integration-strict    # converter integration tests (fail if tools missing)
make test-acceptance            # BDD/godog feature tests
make lint                       # golangci-lint
make build                      # build gsl-diagram, gsl-query, gsl-lsp into ./tmp
make fuzz                       # fuzz tests
```

`make test` includes `TestMarkdownCodeBlocks`, which validates every `gsl`
code block in the committed markdown (root and `docs/`). All `gsl` blocks
must parse and all `invalid-gsl` blocks must fail to parse. Keep this green
when editing documentation.

## Before submitting changes

- [ ] `make test`
- [ ] `make lint`
- [ ] `make test-integration-strict` (or `make test-integration` if tools are unavailable)
- [ ] `make test-acceptance`
- [ ] Existing markdown examples updated, or new ones added, with valid `gsl` blocks
- [ ] SPEC.md updated if language semantics changed
- [ ] GRAMMAR.md updated if syntax changed

## Versioning and compatibility policy

GSL does not use SemVer for the language itself.

- The language specification (SPEC.md) is a normative draft until it is
  promoted to "Version 1.0.0". Draft status means the language can change
  between GSL releases while the project is pre-1.0.
- The GQL specification is a "Revised Draft RFC" and can likewise change.
- CLI, library, and LSP releases follow `vX.Y.Z` tagging on the git
  repository. A release artifact (binary or module version) is
  reproducible: building the tagged commit with the Go version declared in
  `go.mod` produces byte-identical `serialize` output for a given graph.
- Within a released version, canonical form output is stable. If you change
  `serialize.go`, round-trip tests and a `TestFlagshipQueries` comparison
  must keep passing, and an example fixture update is required.

When you change observable behaviour (CLI flags, query semantics, LSP
messages, command output), note it in the release so users can tell what
changed between versions.

## Release process

1. Ensure `make test`, `make lint`, acceptance, and integration all pass.
2. Tag the commit: `git tag vX.Y.Z && git push origin vX.Y.Z`.
3. The Release workflow builds and uploads binaries through GoReleaser.
4. Verify the linux tarball reproducibly by running the flagship examples
   with both the release binary and a source build and diffing output.

## Coding conventions

- Standard library only for the core and query packages.
- The parser is hand-written, not generated.
- Edge order is preserved; duplicates are allowed.
- No schema validation: graph properties such as cycles or tree validity
  are intentionally not enforced.
- Attributes are untyped (`interface{}`); callers must type-assert.
