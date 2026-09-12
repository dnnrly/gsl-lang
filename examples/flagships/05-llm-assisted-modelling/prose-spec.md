# Source material: the "promotion engine" redesign

This is the messy source of truth an agent is handed. It is deliberately
incomplete and partly wrong. Nothing here is a graph, a query, or a
model — it is the raw material **for** a model.

---

## Email thread (subject: promo engine — can we retire the old calc?)

> From: Maya (engineering)
> To: promo-team
>
> Quick status. The new promo service is basically done — it reads offers
> from the offer store, computes discounts per order, and writes the
> resulting discount lines to the reward ledger for accounting.
>
> Key question: is the legacy promo calculator *really* only talking to the
> main product DB now? I seem to remember the old calc also used a cache.
>
> Also — the merchandising team "owns" the rules. The rules service is what
> they maintain. Discount rules are versioned; the promo service calls the
> rules service at checkout time.

> From: Priya (promo team)
> Reply:
> Legacy promo calc:
> - writes discounts to that same main product db (shared since last year's
>   migration)
> - I *think* it still reads prices from the catalog cache. Derek would know.
> - definitely still used by the returns system for re-printing receipts.
>
> The new promo service:
> - reads offers from offer-db (new)
> - calls rules service for the rule evaluation
> - publishes settlement events to the bus (kafka) for the ledger
> - ALSO writes directly to the reward ledger
>
> Merchandising owns offer-db and the rules service. Promo team owns the
> promo service and the legacy calc (until it dies).

---

## Stale README (promo-report.md, last touched 2 years ago)

```
# Promo report (legacy)

- promo-calc computes order discounts.
- reads prices from catalog cache.
- writes discount to orders table in the checkout DB (old).
- nightly batch re-prints receipts for returns.
- Rules live in setup/rules.yaml (outdated -> now a service).
```

Note the stale README references an "orders table in the checkout DB" —
this predates the shared-DB migration. Priya's email says the shared main
product db. Trust the email over the README; mark conflict.

---

## Tickets

- PROMO-114 — "promotion engine is our #1 area of technical debt. New
  promo service exists; retire legacy promo calc. Promo service and rules
  service are load-bearing."
- PROMO-121 — "catalog price cache is deprecated; the catalog service now
  serves prices from its own store."
- INFRA-77 — "event bus (kafka) is the settlement path for the ledger; do
  not bypass it with direct writes" — note: the email says promo service
  writes the ledger directly AND via kafka. Flag for the reviewer.

---

## Ownership cheat-sheet (from the infra onboarding doc)

| Component            | Team         | Notes |
|----------------------|--------------|-------|
| promo service        | promo        | new, retiring legacy calc |
| rules service        | merchandising | rules source of truth |
| offer store (db)     | merchandising | |
| catalog service      | storefront   | serves prices now |
| catalog price cache  | storefront   | deprecated |
| legacy promo calc    | promo        | to be retired |
| returns system       | fulfilment   | re-prints receipts |
| reward ledger        | accounting   | gets settlement events |
| event bus (kafka)    | platform     | settlement path |
| main product db      | platform     | shared product DB |

---