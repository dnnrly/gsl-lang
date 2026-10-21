#!/usr/bin/env bash
# Optional structural gate for GSL dialects authored by the
# gsl-dialect-authoring skill.
#
# Loads the dialect and validates a model against it with gsl-dialect. If
# gsl-dialect is NOT available, report that validation could not run and exit
# 127, so a missing tool is never mistaken for a validated dialect.
#
# Usage: validate.sh dialect.yaml [model.gsl]
#
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2026 Pascal Dennerly.
set -euo pipefail

[ $# -ge 1 ] || { echo "usage: $0 dialect.yaml [model.gsl]" >&2; exit 2; }
dialect="$1"
model="${2:-}"
[ -f "$dialect" ] || { echo "no such file: $dialect" >&2; exit 2; }
if [ -n "$model" ]; then
  [ -f "$model" ] || { echo "no such file: $model" >&2; exit 2; }
fi

if ! command -v gsl-dialect >/dev/null 2>&1; then
  echo "gsl-dialect not found - dialect validation unavailable" >&2
  echo "install via 'go install github.com/dnnrly/gsl-lang/cmd/gsl-dialect@latest' or 'make build'" >&2
  exit 127
fi

echo "validating with dialect $dialect..." >&2
if [ -n "$model" ]; then
  gsl-dialect validate -d "$dialect" -i "$model"
else
  gsl-dialect project -d "$dialect"
fi
