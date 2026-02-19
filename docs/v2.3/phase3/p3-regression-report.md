# v2.3 Phase 3 Regression Report

## Scope

P3-T2 requires the full history/search regression matrix (ST1 through ST5) before `G5` can be marked pass.

## Executions

- `./scripts/v23-regression-scenarios.sh`
- `go test ./tests/e2e/v23 -run TestScenario -count=1` (local equivalent check)

## Result matrix

| ST | Scenario | Evidence | Result | Notes |
| --- | --- | --- | --- | --- |
| ST1 | offline-catchup | `EV-v23-G5-002` | pass | one offline miss followed by recovered catch-up replay |
| ST2 | redaction-tombstone | `EV-v23-G5-002` | pass | tombstones hide local visibility, search matches, and backfill stream |
| ST3 | private-space-anti-enumeration | `EV-v23-G5-002` | pass | denied lookup for wrong token and missing space |
| ST4 | replica-healing | `EV-v23-G5-002` | pass | temporary degraded state then health recovery |
| ST5 | relay-no-history-hosting | `EV-v23-G5-002` | pass | relay path reports hardened boundary refusal |

## Evidence artifacts

- Podman manifest: `artifacts/generated/v23-regression-scenarios/manifest.txt` (`EV-v23-G5-002`)
- Podman run log: `artifacts/generated/v23-regression-scenarios/run.log` (`EV-v23-G5-003`)
- Scenario logs: `artifacts/generated/v23-regression-scenarios/*.log` (`EV-v23-G5-003`)

## Gate readiness

`P3-T2` is complete when all rows above are marked pass and evidence IDs for `G5` are populated in
`docs/v2.3/phase5/p5-evidence-index.md`.
