<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Cookbook

**Audience:** people with a graph to maintain or a question to answer.
**Prerequisite:** [Concepts](../concepts/graph-model.md) + [Running queries](../tutorials/README.md).
**Next step:** [Language Guide](../../GSL_GUIDE.md) / [Query Guide](../../GQL_GUIDE.md) — or straight to the [Specification](../../SPEC.md).

Short, copy-pasteable jobs: the problem, a small model, the query, the answer, and a pointer to the flagship example doing it at real scale. Every recipe is **self-contained** — save the suggested file, run the commands, get the answer shown here.

The first four recipes use only the core language and attributes, so they are reachable right after [Getting Started](../getting-started/README.md) — the Cookbook is not a goal you work up to, it is a set of starting points. Recipe 6 (edge dependencies) is the one that assumes the [concepts on scopes](../concepts/edge-dependencies.md).

## How to read a recipe

1. **Save the model** exactly as shown (e.g. `model.gsl`).
2. **Run the commands.**
3. **Compare with the result.** Results are canonical GSL exactly as the current reference implementation emits them.
4. **Scale up:** every recipe names the flagship model that does the same thing at 20+ nodes.

### One honest note about ordering

GSL output is canonical — deterministic in any single implementation and version — but the *exact line ordering of same-degree elements can differ slightly between versions* while staying valid canonical GSL. What the recipe asserts is the **content**: the nodes, their membership and attributes, and the edges. The flagship results in this repository are regenerated and byte-checked by `go test ./examples -run Flagship`.

## The recipes

| Recipe | Job | Short answer |
|---|---|---|
| [1 — Retire a service](retire-a-service.md) | "Which `@critical` parts break if we retire it?" | The transitive blast radius, intersected with `@critical` |
| [2 — Zoom to one team](per-team-view.md) | "Show me just my team's services." | One subgraph predicate |
| [3 — What is `@critical`?](what-is-critical.md) | "Who is in the critical set, and how are they wired?" | Membership + edges between members |
| [4 — Migration backlog](deprecated-migration-backlog.md) | "What does a deprecated component still reach?" | Deprecated + immediate dependants |
| [5 — Render views in CI](render-a-view-for-ci.md) | "Keep diagrams and derived views fresh." | Regenerate, render, and diff |
| [6 — What does the gate unlock?](what-does-the-gate-unlock.md) | "What can run once the approval gate has passed?" | Edge dependencies in a release workflow |

Consumer notes: recipes 1–4 use only the core language and attributes; recipe 6 uses **scoped edge dependencies**, which (like GQL's edge-dependency predicates) were introduced in the reference implementation and are included in the v0.5.0 release. If you build from source, use the latest tag (`go install github.com/dnnrly/gsl-lang/cmd/gsl-query@latest`).