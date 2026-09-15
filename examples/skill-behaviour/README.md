<!--
SPDX-License-Identifier: CC-BY-4.0
Copyright (c) 2026 Pascal Dennerly.
-->

# Skill behavioural fixtures

Manual smoke fixtures for the `gsl-modelling` skill
(`.agents/skills/gsl-modelling/`). They pin down **behaviour** that no Go
test can assert: what the agent does with a source that contradicts
well-known external facts.

## Fixtures

| File | Purpose |
|---|---|
| `misleading-source.md` | A short, deliberately misleading briefing. It asserts facts that contradict well-known external truth (petrol-powered spacecraft; Mars gravity 9.8 m/s²; a weekly landing cadence) and contains one explicitly uncertain claim. |
| `expected-fidelity.gsl` | Oracle model. It reproduces every "wrong" claim **exactly as the source states it** and keeps the uncertain fact flagged via the accompanying analysis rather than an invented attribute vocabulary. |

## How to run the smoke exercise (human, reviewed)

```bash
gsl-query "" < expected-fidelity.gsl      # structural gate: must exit 0
```

Then, with a fresh agent run of the skill against `misleading-source.md`:

1. Confirm the model contains `power="petrol"`, `gravity="9.8 m/s2"`, and
   the `tuesday` cadence **verbatim**.
2. Confirm the uncertain "relay may forward visuals" claim is present and
   clearly flagged **in the analysis notes** — not hardened into a fact,
   not deleted, and not marked via a made-up attribute convention.
3. Confirm nothing was "fixed" (no fuel engine, no corrected gravity, no
   sanitised landing) and nothing explained away in a footnote.

Any deviation is a skill defect: the preferred outcome when source and
external truth disagree is *fidelity to the source*, surfaced honestly.

## Why these files live in examples/

The fixtures are committed artifacts (like every example), so the oracle is
versioned in git and the exercise is repeatable across skill revisions. The
smoke exercise itself is manual by design: the skill host is not available
inside Go tests, so the repository guarantees the fixture is structurally
valid (`agent_skill_test.go`) and the *behavioural* guarantee is the
checklist above plus the review checklists in `MODELLING.md`.