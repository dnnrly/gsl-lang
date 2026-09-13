# Recipe: Retire a Service — "Who Breaks?"

**Problem:** `legacy_orders` is being switched off. Which `@critical` services end up affected — directly or through the cascade?

**At scale:** the same question on the 20-service hero model — [flagship 01, query 3](../../examples/flagships/01-service-many-views/README.md).

## The model

Save this as `model.gsl`:

```gsl
set critical
set deprecated

node api_gw [team="gateway"] @critical
node checkout [team="checkout"] @critical
node orders [team="orders"] @critical
node orders_db [team="orders"]
node payments [team="payments"]
node payments_db [team="payments"]
node ml_model [team="payments"]
node legacy_orders [team="legacy"] @deprecated

api_gw -> checkout [protocol="http"]
checkout -> orders [protocol="http"]
orders -> payments [protocol="grpc"]
orders -> legacy_orders [protocol="jdbc"]
payments -> payments_db [protocol="sql"]
payments -> ml_model [protocol="grpc"]
legacy_orders -> orders_db [protocol="sql"]
```

## The query

Three moves in one pipeline:

1. **`(subgraph node.id == "legacy_orders" traverse in all) as BLAST`** — everything that transitively depends on `legacy_orders`, as a named graph.
2. **`(subgraph node in @critical) as CRIT`** — the critical set, as a named graph.
3. **`BLAST & CRIT`** — where they overlap.

```bash
gsl-query '(subgraph node.id == "legacy_orders" traverse in all) as BLAST | from * | (subgraph node in @critical) as CRIT | BLAST & CRIT' < model.gsl
```

## The answer

```gsl
set critical
set deprecated

node api_gw [team="gateway"] @critical
node checkout [team="checkout"] @critical
node orders [team="orders"] @critical

api_gw->checkout [protocol="http"]
checkout->orders [protocol="http"]
```

Retiring `legacy_orders` takes down `orders` — and because `checkout` depends on `orders` and `api_gw` depends on `checkout`, the blast **cascades** through all three `@critical` services. The transitive traversal found what a human staring at a diagram would be forgiven for missing.

## What to do with it

- Route this through `gsl-diagram` for a reviewable blast-radius picture for the change ticket.
- Re-run it after any model change: the answer is re-derived, never hand-recomputed.
- Make the set membership the review target — a service leaving `@critical` is a visible diff, not an assumption.

**Next:** [Zoom to one team](per-team-view.md) — the same idea, a different cut.