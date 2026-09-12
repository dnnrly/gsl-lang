# Concept: Edge Dependencies

**Audience:** everyone modelling workflows, gates or sequenced steps. **Prerequisite:** [Structure and nesting](structure-and-nesting.md).
**Next step:** [Queries and derived views](queries-and-views.md).

Most graph relationships connect **nodes**. Edge dependencies are different: a *relationship itself* can participate in a dependency relationship with another relationship.

> Most graph relationships connect nodes. Edge dependencies allow relationships themselves to participate in dependency relationships.

That is unusual — and powerful. If you are modelling sequencing ("when can we promote?", "what does the approval gate unlock?"), the subject of the question is not a service or a decision; it is the *edge* between them.

## Label an edge, then make another edge depend on it

```gsl
CONFIG: config -> controlplane [stage="apply"]
GATE: controlplane -> approve [stage="gate", parent=CONFIG]
```

Read the second line as: *`GATE` runs only after `CONFIG` has completed.* Both are still ordinary edges — queryable, renderable, in the same graph — but the model now records an ordering **between dependencies**, not just between nodes.

A chain of these reads naturally as a workflow:

```gsl
GATE: controlplane -> approve
PROMOTE: registry -> canary [parent=GATE] {
    CANARY: canary -> prod
    SMOKE: prod -> smoke
}
```

`PROMOTE` depends on `GATE`, and the steps inside its scope (`CANARY`, `SMOKE`) follow the promotion chain. Note the scoped blocks themselves must not carry a `parent` attribute — the label-level `[parent=...]` plus nesting covers the relationship.

This is how flagship [03 — release prerequisites](../../examples/flagships/03-release-prerequisites/README.md) answers "when can we promote?" without a single workflow keyword — promotions, approvals, release gates and task pipelines are all just edges with edges depending on them. GQL exposes the structure through the edge-dependency predicates (`edge parent exists`, `edge.depth`, `edge depends on ... scope` — see the [query tutorial](../../QUERY_TUTORIAL.md)).

> **Version note:** scoped/edge-dependency syntax is part of the language and the current reference implementation, but the released binary may lag — build the tools from source if a `LABEL: a -> b { ... }` model fails to parse on your install ([Getting Started](../getting-started/README.md#1-install-the-tools)).

---

**Next:** [Queries and derived views](queries-and-views.md) — what GQL is, and how it turns a graph into the views you need.