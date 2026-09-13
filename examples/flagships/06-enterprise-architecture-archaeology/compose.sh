#!/usr/bin/env bash
# Compose the independently discovered service models into the enterprise
# architecture graph.
#
# There is no dedicated "merge" command in the GSL toolchain. Composition is
# the language's own mechanism: concatenate the fragments and the parser
# performs a declarative merge -- repeated node/set declarations merge
# (last-write-wins on attributes, membership accumulates) and every edge is
# preserved. Order matters only when two fragments claim different
# attributes for the same node; this example's fragments use unique node IDs
# and consistent attributes, so order is irrelevant.
set -euo pipefail
here="$(cd "$(dirname "$0")" && pwd)"
cd "$here"
cat discovered/catalogue-service.gsl \
    discovered/customer-api.gsl \
    discovered/fulfilment-service.gsl \
    discovered/inventory-service.gsl \
    discovered/notification-service.gsl \
    discovered/order-service.gsl \
    discovered/payment-service.gsl > model.gsl
echo "composed $(wc -l < model.gsl) lines into model.gsl"