<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Recipe: Zoom to One Team

**Problem:** the model is a system-wide graph. The payments team needs *just their slice* — services, databases, and the edges between them.

**At scale:** the flagship's [team-level view](../../examples/flagships/01-service-many-views/README.md), one of the seven derived views of the same 20-service model.

## The model

Same model as [recipe 1](retire-a-service.md) — save it as `model.gsl`.

## The query

One predicate. `node.team == "payments"` selects the payments-team services, and the subgraph automatically includes the edges *between* the selected nodes:

```bash
gsl-query 'subgraph node.team == "payments"' < model.gsl
```

## The answer

```gsl
set critical
set deprecated

node payments [team="payments"]
node payments_db [team="payments"]
node ml_model [team="payments"]

payments->payments_db [protocol="sql"]
payments->ml_model [protocol="grpc"]
```

The team slice is *exactly* their three nodes and the two edges joining them — no report-spectrum dump, no hand-filtered screenshot.

> **Ordering note:** a version difference may list `ml_model` before `payments` here. The content — three nodes, two edges — is identical either way.

## Variations

- **Plus the callers** — `subgraph node.team == "payments" traverse in 1` also pulls in the services that call payments-team endpoints (here, `orders`).
- **Overlay reach** — combine with a set to ask "what do we have that is *both* mine and critical?" — e.g. `(subgraph node.team == "orders") as MINE | (subgraph node in @critical) as CRIT | MINE & CRIT` (named-graph intersection, from [query tutorial Step 9](../../QUERY_TUTORIAL.md#step-9-named-graphs-and-graph-algebra)).
- **Render it** — `gsl-query 'subgraph node.team == "payments"' < model.gsl | gsl-diagram -f mermaid -t component` makes the team diagram.

**Next:** [What is `@critical`?](what-is-critical.md) — the flip side: the set, not the team.