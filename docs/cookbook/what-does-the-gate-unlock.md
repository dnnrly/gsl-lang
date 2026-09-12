# Recipe: What Does the Approval Gate Unlock?

**Problem:** a release pipeline has an approval gate. Once the gate passes, what is allowed to run — and what is still held?

**At scale:** this is the whole of [flagship 03 — release prerequisites](../../examples/flagships/03-release-prerequisites/README.md), where two tracks rejoin before production.

> **Version note:** this recipe uses **scoped edge dependencies** — edges that depend on other edges. This is a current-reference-implementation feature (and part of the language spec), but the release binary may lag: if the commands fail to parse, build the tools from source with `make build` or `go install ./cmd/gsl-query` ([Getting Started](../getting-started/README.md#1-install-the-tools)).

## The model

Save this as `release.gsl`:

```gsl
node config
node controlplane
node approve
node registry
node canary
node prod
node smoke

CONFIG: config -> controlplane [stage="apply"]
GATE: controlplane -> approve [stage="gate", parent=CONFIG]
PROMOTE: registry -> canary [stage="promote", parent=GATE] {
    CANARY: canary -> prod [stage="ship"]
    SMOKE: prod -> smoke [stage="smoke"]
}
```

Read the middle three lines as *release choreography*: `GATE` runs after `CONFIG`; `PROMOTE` runs after `GATE`; and inside `PROMOTE`'s scope, `CANARY` and `SMOKE` follow its lead. No workflow DSL — just edges with parents.

## Query 1 — find the gate itself

```bash
gsl-query 'subgraph edge.stage == "gate"' < release.gsl
```

```gsl
node approve
node controlplane

GATE: controlplane->approve [stage="gate"]
```

## Query 2 — what does passing the gate unlock?

The dependency predicate `edge depends on ... scope` selects every edge whose parent chain reaches an edge with `stage="gate"`, and `scope` expands to all of its descendants:

```bash
gsl-query 'subgraph edge depends on edge.stage == "gate" scope' < release.gsl
```

## The answer

```gsl
node canary
node registry
node prod
node smoke

PROMOTE: registry->canary [stage="promote"] {
    CANARY: canary->prod [stage="ship"]
    SMOKE: prod->smoke [stage="smoke"]
}
```

Passing `GATE` unlocks promotion and the canary/smoke rollout; it does **not** unlock `CONFIG` or anything earlier. The answer is *exactly* the set of edges whose parent chain reaches the gate — derived, not read off a runbook.

## Turning it into a policy check

- **"Hold until approved"** — the gate-holding query is the complement: `subgraph edge parent exists` minus the unlocked set. GQL lets you combine these as named graphs ([query tutorial Step 9](../../QUERY_TUTORIAL.md#step-9-named-graphs-and-graph-algebra)).
- **Reuse the flagships' vocabulary** — flagship 03 queries by `edge.stage`, exactly as here, at six times the size.

**Next:** back to the [cookbook index](README.md), or up a level to the concept on [edge dependencies](../concepts/edge-dependencies.md).