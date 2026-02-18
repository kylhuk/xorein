# Phase 5 · Risk register (Planning in progress)

## Purpose
Track residual and gating dependencies that must be mitigated before the v11 promotion gate can move from `blocked` to `ready_for_review`.

## Risk table
| RiskID | Description | Severity | Mitigation | Owner role | Status |
|---|---|---|---|---|---|
| R11-01 | Mandatory global command evidence remains incomplete (`buf lint`, `buf breaking`, `go test ./...`, `make check-full`), so promotion cannot proceed. | high | Run missing commands, store immutable logs/checksums, and update `p5-evidence-index.md` + `p5-gate-signoff.md`. | QA Lead | open |
| R11-02 | Proto guardrail gate (G1) remains blocked until buf outputs are attached, which risks additive-policy drift in later edits. | medium | Execute `buf lint` and `buf breaking`, bind outputs to EV-v11-G1-001/002, and require Protocol Lead review before status change. | Protocol Lead | open |
| R11-03 | Deferral register for out-of-scope runtime features is not yet published, reducing closure clarity for release handoff. | medium | Populate the v11 deferral register using `docs/templates/roadmap-deferral-register.md` and cross-link it from gate signoff + release notes draft. | Plan Lead | open |

## Planned vs implemented
- **Planned:** Keep this table current; escalate any blocker to a `blocked` gate status and tie the risk back to `p5-gate-signoff.md` before promotion.
- **Implemented:** G4 evidence collection has started (e2e + smoke + roadmap verify captured), while mandatory global checks and deferral publication remain open risks.
