# Phase 5 · As-built conformance (Planning in progress)

## Purpose
Capture the as-built (or as-planned) state of v11 relative to the release criteria described in `TODO_v11.md`, emphasizing which artifacts are complete, which remain planned, and how each gate will consume evidence.

## As-built summary
| Clause | Planned evidence | Status | Notes |
|---|---|---|---|
| Command evidence from Phase 5 (`go test`, relay smoke, docs verify) | Logs + SHA256 anchors under `artifacts/v11/evidence/` and `artifacts/generated/v11-relay-smoke/` | partially complete | `go test ./tests/e2e/v11/...`, `scripts/v11-relay-smoke.sh`, and `./scripts/verify-roadmap-docs.sh` evidence is recorded; `go test ./...` and `make check-full` remain pending. |
| Compatibility verifications (`buf lint`, `buf breaking`) | EV-v11-G1-001/002 entries updated in `p1-compatibility-policy.md` and `p5-evidence-index.md` | pending | Compatibility policy is documented, but buf command outputs are not yet attached. |
| Documentation/evidence indexes | `docs/v1.1/phase5/p5-evidence-index.md`, `p5-gate-signoff.md`, `p5-risk-register.md` | in progress | Index/checklist/risk docs are active and now include concrete G4/G5 supporting evidence plus remaining blockers. |

## Traceability to TODO
- Every entry above refers back to `TODO_v11.md` gate commitments (G0-G5). Use this page to explain outstanding dependencies before promotion.

## Planned vs implemented
- **Planned:** Once remaining mandatory command outputs and approvals are recorded, this document will capture final conformance and promotion recommendation.
- **Implemented:** Partial as-built evidence is now recorded; final conformance remains blocked by pending global verification commands and gate approvals.
