# Phase 5 - Evidence bundle (v11 closure)

## Purpose
Collect the runnable evidence items required for v11 promotion. This bundle enumerates command evidence, documentation checkpoints, and release precursors captured during closure.

## Evidence categories
1. **Command evidence** – scripts and tests whose output anchors the gate runner checklist.
2. **Docs/evidence artifacts** – docs, checklists, and registries referenced from this phase.
3. **Scenario references** – deterministic scenarios and rollout notes captured elsewhere (Phase 4, TODO_v11).

### Command evidence tracker
| Evidence focus | Command | EV ID | Planned output | Status |
|---|---|---|---|---|
| Proto lint guardrail | `buf lint` | EV-v11-G1-001 | `artifacts/v11/evidence/EV-v11-G1-001-buf-lint.log` | pass (warning recorded) |
| Proto breaking guardrail | `buf breaking` | EV-v11-G1-002 | `artifacts/v11/evidence/EV-v11-G1-002-buf-breaking.log` | pass |
| Full repository tests | `go test ./...` | EV-v11-G5-004 | `artifacts/v11/evidence/EV-v11-G5-004-go-test-all.log` | pass |
| v11 relay e2e contracts | `go test ./tests/e2e/v11/...` | EV-v11-G4-001 | `artifacts/v11/evidence/EV-v11-G4-001-go-test-e2e-v11.log` | pass |
| Full check pipeline | `make check-full` | EV-v11-G5-005 | `artifacts/v11/evidence/EV-v11-G5-005-make-check-full.log` | pass |
| Relay smoke gating | `scripts/v11-relay-smoke.sh` | EV-v11-G4-002 | `artifacts/v11/evidence/EV-v11-G4-002-relay-smoke.log` | pass |
| Docs verification (supporting) | `./scripts/verify-roadmap-docs.sh` | EV-v11-G5-003 | `artifacts/v11/evidence/EV-v11-G5-003-roadmap-verify.log` | pass |

## Reference index
- All evidence items above link to `docs/v1.1/phase5/p5-evidence-index.md` once they run; the index will cite the immutable log paths and SHA256s required by the gate.
- Deferral governance and no-active-deferral status are tracked in `docs/v1.1/phase5/p5-deferral-register.md`.

## Planned vs implemented
- **Planned:** Keep this bundle current if release-impacting changes land after v11 promotion.
- **Implemented:** `buf lint`, `buf breaking`, `go test ./...`, v11 e2e, relay smoke, roadmap verification, and `make check-full` outputs are captured with concrete logs and checksums.
