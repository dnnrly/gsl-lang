# Recipe: Migration Backlog — "What Does Deprecated Still Reach?"

**Problem:** `legacy_orders` is `@deprecated`. The migration story needs the concrete list of what it still touches, so ticket size is not a guess.

**At scale:** the flagship's [deprecated-migration backlog](../../examples/flagships/01-service-many-views/README.md) derives the same answer for its 20-service model.

## The model

Same model as [recipe 1](retire-a-service.md) — save it as `model.gsl`.

## The query

**Option A — just the deprecated service** (the "ticket for retiring it"):

```bash
gsl-query 'subgraph node in @deprecated' < model.gsl
```

**Option B — what it still reaches** (the "blast of the retirement" from the other side — walk *its* dependencies outward):

```bash
gsl-query 'subgraph node in @deprecated traverse out 1' < model.gsl
```

## The answer (Option B)

```gsl
set critical
set deprecated

node legacy_orders [team="legacy"] @deprecated
node orders_db [team="orders"]

legacy_orders->orders_db [protocol="sql"]
```

`legacy_orders`'s retirement must account for `orders_db` — an inventory fact that, on paper, is exactly the kind of thing that drifts until it bites.

> **Direction matters.** `traverse out` follows an edge *away from* the matched node (what `legacy_orders` calls). `traverse in` follows edges *toward* it (what calls `legacy_orders`) — which is the [retire-a-service](retire-a-service.md) question.

## Building the backlog

- Run Option B for each `@deprecated` node (or widen to `traverse out all` for full transitive reach).
- Pipe the union through `gsl-diagram` and attach the picture to the tickets.
- Re-run after every model change — the backlog stays derived, never hand-maintained.

**Next:** [Render views in CI](render-a-view-for-ci.md) — make the re-deriving automatic.