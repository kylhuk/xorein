# Phase 5 · Stress-Testing Contracts

## Plan vs implementation
Stress contracts cover the 1000-member server, 50-person voice baseline, latency bench, and +50 increments, and the deterministic harness now exists in `tests/perf/v09/deterministic_suite_test.go` (guided by `phase9` runbook thresholds) plus `tests/e2e/v09/scenario_suite_test.go` for CLI witness coverage. The documentation here still records the contracts and runbook assumptions so reviewers can confirm the stress/perf automation matches the stated thresholds.

## Evidence anchors
| VA ID | Artifact | Files | Description |
|---|---|---|---|
| VA-T1..VA-T3 | Scenario definitions | `docs/v0.9/phase5/p5-stress-contract.md` | Membership churn, activity distribution, and latency profile descriptions for the baseline stress campaigns. |
| VA-T4..VA-T6 | Incremental +50 planning | `docs/v0.9/phase5/p5-stress-contract.md` | Hard limit expansion plan that accompanies the runbook thresholds in `phase9`. |
| VA-E7 | Perf reproducibility runbook | `docs/v0.9/phase9/perf-runbook.md`, `docs/v0.9/phase9/VA-E7-thresholds.md` | Documents step-by-step reproduction instructions plus metric thresholds. |
| VA-T7/VA-R7/VA-B7 | Cross-domain operator continuity | `docs/v0.9/phase5/p5-stress-contract.md` | Cover operator continuity, relay/mobile governance, and battery policy references needed for a multi-target validation matrix. |

## Planned-vs-implemented language
- Implemented: the deterministic harness under `tests/perf/v09/deterministic_suite_test.go` plus the scenario witness in `tests/e2e/v09/scenario_suite_test.go` execute the contracts described here; the runbook thresholds in `phase9` tie those runs back to the documented metrics and incremental limits.
