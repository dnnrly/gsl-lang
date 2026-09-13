<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Recipe: What Is `@critical`?

**Problem:** your architecture review needs the load-bearing services — who is in the `@critical` set, and how are they wired to each other?

**At scale:** the flagship model's [critical set](../../examples/flagships/01-service-many-views/README.md) is the same membership idea across 20 services.

## The model

Same model as [recipe 1](retire-a-service.md) — save it as `model.gsl`.

## The query

Membership is data, so this is a single predicate — no tag-hunting through a diagram:

```bash
gsl-query 'subgraph node in @critical' < model.gsl
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

The critical subgraph shows both the members *and* the edges between them: `api_gw`, `checkout` and `orders` form the checkout spine, and `orders_db`'s storage is not marked critical in this model — a deliberate fact the answer makes visible.

## What membership buys you

- **A reviewable diff.** Promoting a service to `@critical` is a one-line change whose blast you can then re-derive.
- **A shared vocabulary.** `@critical`, `@deprecated`, `@external` mean what your team says they mean — applied once, queried everywhere.
- **Faster questions.** "Is this in the blast radius of another critical set?" lines up naturally for review.

**Next:** [Migration backlog](deprecated-migration-backlog.md) — the other membership, worked the same way.