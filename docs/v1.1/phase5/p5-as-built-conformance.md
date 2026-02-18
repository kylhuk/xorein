# Phase 5 - As-built conformance (v11 closure)

## Purpose
Capture the as-built state of v11 relative to the release criteria described in `TODO_v11.md`, with explicit evidence anchors for each closure clause.

## As-built summary
| Clause | Planned evidence | Status | Notes |
|---|---|---|---|
| Command evidence from Phase 5 (`go test`, relay smoke, docs verify) | Logs + SHA256 anchors under `artifacts/v11/evidence/` and `artifacts/generated/v11-relay-smoke/` | complete | `go test ./...`, `go test ./tests/e2e/v11/...`, `scripts/v11-relay-smoke.sh`, `./scripts/verify-roadmap-docs.sh`, and `make check-full` evidence is recorded as passing in the latest run. |
| Compatibility verifications (`buf lint`, `buf breaking`) | EV-v11-G1-001/002 entries updated in `p1-compatibility-policy.md` and `p5-evidence-index.md` | complete (warning noted) | Commands executed successfully; lint warning about deprecated `DEFAULT` category is recorded for follow-up config cleanup. |
| Documentation/evidence indexes | `docs/v1.1/phase5/p5-evidence-index.md`, `p5-gate-signoff.md`, `p5-risk-register.md`, `p5-deferral-register.md` | complete | Closure docs are linked and include no-active-deferrals governance evidence (`EV-v11-G6-001`). |

## Traceability to TODO
- Every entry above refers back to `TODO_v11.md` gate commitments (G0-G7). This page is the final conformance reference for v11 promotion.

## Planned vs implemented
- **Planned:** If release-impacting changes land after this snapshot, rerun mandatory evidence commands and refresh references here before any new promotion decision.
- **Implemented:** Mandatory command evidence is recorded and passing; documentation closure artifacts are complete.

## Promotion recommendation
- Promote v11 (`F11`) based on complete mandatory evidence, promoted gate states, and no active deferrals in `docs/v1.1/phase5/p5-deferral-register.md`.
