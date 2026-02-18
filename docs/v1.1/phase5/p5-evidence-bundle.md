# Phase 5 · Evidence bundle (Planning in progress)

## Purpose
Collect the runnable evidence items required for the v11 release gate while keeping the planning language explicit: the bundle enumerates command evidence, documentation checkpoints, and release precursors without claiming promotion is complete.

## Evidence categories
1. **Command evidence** – scripts and tests whose output anchors the gate runner checklist.
2. **Docs/evidence artifacts** – docs, checklists, and registries referenced from this phase.
3. **Scenario references** – deterministic scenarios and rollout notes captured elsewhere (Phase 4, TODO_v11).

### Command evidence tracker
| Evidence focus | Command | EV ID | Planned output | Status |
|---|---|---|---|---|
| Proto lint guardrail | `buf lint` | EV-v11-G1-001 | `pending/EV-v11-G1-001-buf-lint.log` | pending (not run) |
| Proto breaking guardrail | `buf breaking` | EV-v11-G1-002 | `pending/EV-v11-G1-002-buf-breaking.log` | pending (not run) |
| Full repository tests | `go test ./...` | EV-v11-G5-004 | `pending/EV-v11-G5-004-go-test-all.log` | pending (not run) |
| v11 relay e2e contracts | `go test ./tests/e2e/v11/...` | EV-v11-G4-001 | `artifacts/v11/evidence/EV-v11-G4-001-go-test-e2e-v11.log` | pass |
| Full check pipeline | `make check-full` | EV-v11-G5-005 | `pending/EV-v11-G5-005-make-check-full.log` | pending (not run) |
| Relay smoke gating | `scripts/v11-relay-smoke.sh` | EV-v11-G4-002 | `artifacts/v11/evidence/EV-v11-G4-002-relay-smoke.log` | pass |
| Docs verification (supporting) | `./scripts/verify-roadmap-docs.sh` | EV-v11-G5-003 | `artifacts/v11/evidence/EV-v11-G5-003-roadmap-verify.log` | pass |

## Reference index
- All evidence items above link to `docs/v1.1/phase5/p5-evidence-index.md` once they run; the index will cite the immutable log paths and SHA256s required by the gate.

## Planned vs implemented
- **Planned:** Capture the remaining TODO-mandated commands (`buf lint`, `buf breaking`, `go test ./...`, `make check-full`) and add their checksums before any promotion decision.
- **Implemented:** v11 e2e, relay smoke, and roadmap verification outputs are now captured with concrete log paths; mandatory global checks remain pending.
