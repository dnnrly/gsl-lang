# Recipe: Render Views in CI

**Problem:** diagrams and derived views go stale. You want a pipeline that regenerates them from the model file, and a check that fails when they drift.

**At scale:** this is exactly how this repository itself is kept truthful — every committed flagship result is regenerated and byte-checked by `go test ./examples -run Flagship`, batch-processed with the converters.

## The idea in one line

The graph file is the source; *everything else is derived*. So CI derives it again — and diffs.

## 1. Regenerate the canonical results

The committed result files (`*.result.gsl`, e.g. in `examples/flagships/01-service-many-views/`) are the ground truth. Regenerate and diff them:

```bash
cd examples/flagships/01-service-many-views
gsl-query -f q3-critical-blast.gql -i model.gsl | diff -q - q3-critical-blast.result.gsl && echo "fresh"
gsl-query -f q7-deprecated.gql   -i model.gsl | diff -q - q7-deprecated.result.gsl   && echo "fresh"
```

If the model and the queries changed the derived answer, the diff is the review — a canonical, human-readable delta (see [Canonical form](../concepts/canonical-form.md)).

> **Use the tools as demonstrated.** Result files were created by a specific implementation. Build the tools from the repository source in CI (`make build`, then `tmp/gsl-query`) so byte checks match; see [Getting Started](../getting-started/README.md#1-install-the-tools).

## 2. Render the diagrams

Derive the views on every push and commit them back (or to an artifact store):

```bash
gsl-query 'from *' < model.gsl | gsl-diagram -f mermaid -t component > views/full.component.mmd
gsl-query -f q3-critical-blast.gql -i model.gsl | gsl-diagram -f mermaid -t graph > views/critical-blast.mmd
```

The `.mmd` files are Markdown-readable on GitHub — a diagram review without opening a drawing tool.

## 3. Make drift a failure

Two cheap guards:

- **Result freshness** — the `diff -q` above as a CI step: derived answers must match the committed files.
- **Parse integrity** — every changed file round-trips: `gsl-query "" < model.gsl > /dev/null` fails loudly on a broken graph. (This repository runs the same check over every GSL code block in its documentation.)

## What you get

A model edit that forgets to regenerate views is a **failed build**, not a forgotten chore. The reviewable object is the model diff itself — the diagrams and answers that advertise the change are re-derived and honest.

**Next:** [What does the gate unlock?](what-does-the-gate-unlock.md) — workflows as graphs, the flagship's third idea.