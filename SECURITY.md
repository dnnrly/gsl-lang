# Security Policy

## Reporting a vulnerability

Please report security issues by opening a private advisory via the
[GitHub security advisory form](https://github.com/dnnrly/gsl-lang/security/advisories/new)
rather than a public issue.

You should receive an acknowledgement within a few days, and a coordinated
disclosure timeline will be agreed in the advisory thread.

## Scope

GSL is a graph modelling and query language with command-line tools
(`gsl-diagram`, `gsl-query`, `gsl-lsp`), a Go library, and a VS Code
extension. The following are in scope:

- Malicious GSL/GQL documents that trigger a crash, excessive resource use,
  or memory corruption in the parser, query engine, or language server.
- Path traversal or shell injection through CLI arguments, input files, or
  environment variables.
- Supply-chain issues in the release binaries or CI pipelines.

Diagram conversion invokes external renderers (mermaid-cli, PlantUML).
Treat diagram tooling output from untrusted graphs with the same care as any
external command execution.

## Supported versions

Security fixes are applied to the latest tagged release and backported on
request. Please include the version you are on when reporting.