# Roadmap Gate Checklist Template

Use this template for `docs/vX.Y/phase5/p5-gate-signoff.md`.

## Gate statuses
- `open`: not ready for review.
- `blocked`: failed or missing prerequisite/evidence.
- `ready_for_review`: owner complete, approvers pending.
- `promoted`: all approvers signed and evidence complete.

## Fail-close rule
- A gate may only move to `promoted` when every entry/exit check is satisfied and evidence links resolve.
- Any failed check returns gate status to `blocked`.

## Checklist table

| GateID | Purpose | Entry checks | Exit checks | Owner role | Required approvers | Evidence IDs | Status | Notes |
|---|---|---|---|---|---|---|---|---|
| G0 | Scope lock | | | | | | open | |

## Completion rule
- Version promotion is allowed only when all listed gates are `promoted`.
