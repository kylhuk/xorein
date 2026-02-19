# Phase 5 as-built conformance summary (P5-T1 ST1–ST4)

## Purpose
- Declares the "as-built" runtime/test/build posture audited by Gate `G7` based on the executed Phase 5 command suite and regression scenarios.

## ST1–ST4 evidence checkpoints
| ST | As-built target | Verification target |
| --- | --- | --- |
| ST1 | Regression harness scenarios (`scripts/v23-regression-scenarios.sh`, `containers/v2.3/scenarios.conf`) capture lifecycle coverage across offline catch-up, live continuity, degraded recovery, and restart resilience. | `artifacts/generated/v23-evidence/v23-regression-scenarios.txt` + manifest/logs under `artifacts/generated/v23-regression-scenarios/` (EV-v23-G7-009). |
| ST2 | Controlled `tests/e2e/v23` fixtures exercised against the latest scenario catalog with deterministic status reporting. | `artifacts/generated/v23-evidence/go-test-e2e-v23.txt` (EV-v23-G7-004). |
| ST3 | Perf suite (`tests/perf/v23`) validates latency and QoS budgets per v2.3 SLO targets. | `artifacts/generated/v23-evidence/go-test-perf-v23.txt` (EV-v23-G7-005). |
| ST4 | Build/lint hygiene surfaced by `buf` commands, `go build` for critical binaries, and `make check-full`. | `artifacts/generated/v23-evidence/buf-lint.txt`, `buf-breaking.txt`, `go-build-*.txt`, `make-check-full.txt` (EV-v23-G7-001, G7-002, G7-006, G7-007, G7-008). |

## Gate `G7` readiness narrative
- `G7` is satisfied when all ST1–ST4 checkpoints above reference the EV-v23-G7 artifacts and the go/no-go record reflects the GO decision.
- The gate sign-off checklist (`p5-gate-signoff.md`) must cite the same artifacts for traceability (advisory warnings only).

## Gate `G8` relay boundary assurance
- Relay boundary focus is documented by the dedicated `TestScenarioRelayNoHistoryHosting` run and the Podman harness manifest that captures offline/continuity/resilience variants while affirming relays store no long-term state. Evidence: `artifacts/generated/v23-evidence/go-test-relay-boundary.txt` and `artifacts/generated/v23-regression-scenarios/manifest.txt` (EV-v23-G8-001, EV-v23-G8-002).

## Gate `G9` as-built/spec review
- G9 reviewers rely on the published F23 spec input package manifest plus this as-built summary to ensure the executed posture matches the planned roadmap. Evidence: `artifacts/generated/v23-evidence/f23-spec-inputs.txt` lists the packages consulted and this document documents the narrative review (EV-v23-G9-001, EV-v23-G9-002). Command: N/A (document review and spec-package list compilation).
