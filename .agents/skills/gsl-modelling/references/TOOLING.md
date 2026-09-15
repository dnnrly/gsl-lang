---
name: gsl-modelling-tooling
type: skill-reference
audience: agent
description: Step-up tooling for the gsl-modelling skill — the gsl-query/gsl-diagram/gsl-lsp commands, when to step up, and the honesty rules that govern claiming validation, execution, or rendering.
---

<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Tooling & Step-Up

The skill's core work — reading, analysing, and writing GSL — needs **no
tooling**. Tools are optional *step-up*: they validate, execute, and
render. They are never implied. If a tool is not available, the skill
still produces the model and says honestly what it could and could not
check.

## Tools

| Tool | What it does | When to step up |
|---|---|---|
| `gsl-query` | Runs GQL over GSL. With an empty query it canonicalises and validates structure. | Before delivering a model; to answer review questions. |
| `gsl-diagram` | Renders GSL to Mermaid or PlantUML (component / graph / sequence). | When the user wants a visual view. Output is illustrative, not a contract. |
| `gsl-lsp` | Language-server diagnostics, completion, hover. | During interactive editing in a supporting editor. |
| `gsl-query ai gsl` / `gsl-query ai query` | Prints the authoritative, embedded language guides. | When tooling is present and you need the exact guide text. |

## The structural gate

The one step worth doing by default when a model is complete:

```bash
gsl-query "" < model.gsl
```

- **Exit 0 with no warnings** — structurally valid canonical output.
- **Exit 0 with warnings** (e.g. `implicit set creation: "foo"`) — valid
  but tells you something is under-declared. Read them; accept consciously
  or fix.
- **Non-zero exit** — a syntax error. Fix until it parses.

`scripts/validate.sh` wraps exactly this gate and distinguishes "tool
missing" (`127`) from "checked clean", so nobody can mistake "no tool" for
"validated".

## Honesty rules (non-negotiable)

1. **Never claim validation, execution, or rendering that did not happen.**
   If `gsl-query` did not run, say "not validated — no tooling
   available", and offer the command.
2. **The tool validates structure, not semantics.** A clean parse says
   nothing about whether the model is faithful, correct, or complete. The
   semantic gate is the human + the review checklists in MODELLING.md.
   Say so when presenting the model.
3. **Canonicalise with the same tool version before diffing.** Two models
   canonicalised by different versions may order bounded same-depth
   elements differently. Never attribute that to a content change.
4. **Rendering is illustrative.** `gsl-diagram` output depends on map-iteration
   ordering; the GSL source is the artifact of record.

## Detecting tools

Probe availability before assuming:

```bash
command -v gsl-query || echo "gsl-query not found"
```

Do not attempt to install tools on the user's behalf without being asked.

## Suggested step-up at delivery

After delivering a model, offer (do not silently perform):

1. Structural gate / canonicalise: `gsl-query "" < model.gsl`
2. A GQL question that shows the model in action:
   `gsl-query '<pipeline>' < model.gsl`
3. A diagram: `gsl-diagram -f graph -t graph < model.gsl`

Present these as offers with exact commands; let the user choose. When you
do run them, capture output and quote it accurately.