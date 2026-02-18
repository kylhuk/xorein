# Phase 5 · Gate signoff (Planning in progress)

## Gate workflow reminders
- Gate statuses follow the template in `docs/templates/roadmap-gate-checklist.md` (open/blocked/ready_for_review/promoted).  No gate moves to `promoted` without recorded evidence and approver timestamps.
- Role expectations mirror the `roadmap-signoff-raci` template: Plan Lead owns scope, Protocol Lead owns compatibility, Runtime/Client Leads own implementation, QA owns validation evidence, and Release Authority signs the final handoff.

## Gate checklist
| GateID | Purpose | Entry checks | Exit checks | Owner role | Required approvers | Evidence IDs | Status | Notes |
|---|---|---|---|---|---|---|---|---|
| G0 | Scope lock and traceability | Phase0 scope + trace matrix exists | Phase0 docs cite gate runner contract | Plan Lead | QA Lead, Release Authority | EV-v11-G0-001 (planned) | ready_for_review | Gate0 docs exist; awaiting final signoff from Release Authority. |
| G1 | Compatibility policy | `p1-compatibility-policy.md` drafted | `buf lint` + `buf breaking` logs captured | Protocol Lead | QA Lead, Release Authority | EV-v11-G1-001, EV-v11-G1-002 | blocked | Commands still pending; see `p1-compatibility-policy.md` and `p5-evidence-index.md`. |
| G2 | Gate runner | `phase1/p1-gate-runner.md` describes runner | Runner software executed against sample gate | Runtime Lead | Client Lead, QA Lead | EV-v11-G2-001 | ready_for_review | Runner doc exists; automated run still scheduled. |
| G3 | Relay data boundary | Phase2 limit doc signed off | Scenario pack data + policy asserts | Runtime Lead | Protocol Lead, QA Lead | EV-v11-G3-001 | ready_for_review | Phase2 doc produced; scenario evidence pending. |
| G4 | Podman smoke | Podman smoke doc + scripts referenced | Relay smoke script output + Go test evidence | QA Lead | Runtime Lead, Plan Lead | EV-v11-G4-001, EV-v11-G4-002, EV-v11-G4-003 | ready_for_review | v11 e2e + smoke logs and manifest are captured; awaiting approver review. |
| G5 | Evidence bundle & release handoff | Phase5 docs assembled | Evidence index, risk register, gate signoff updated | Release Authority | All leads | EV-v11-G5-003, EV-v11-G5-004, EV-v11-G5-005 | blocked | Bundle updated, but mandatory `go test ./...` and `make check-full` evidence is still pending. |
| G6 | Docs/evidence completeness | Gate checklist + evidence index drafted | Mandatory evidence table complete with checksums and ownership | Plan Lead | QA Lead, Release Authority | EV-v11-G6-001 | blocked | Waiting for remaining mandatory command outputs and final deferral register linkage. |
| G7 | F11 as-built conformance | `p5-as-built-conformance.md` drafted | Release recommendation reviewed against all gate states | Release Authority | Plan Lead, QA Lead | EV-v11-G7-001 | blocked | As-built report is partial until G1/G5/G6 blockers clear. |

## RACI notes
- This file references `docs/templates/roadmap-signoff-raci.md` to keep ownership consistent.
- When the remaining evidence runs complete, each owner will update the evidence ID column with actual log paths and approvals before changing the status to `ready_for_review`/`promoted`.

## Planned vs implemented
- **Planned:** Align `p5-evidence-index.md`, `p5-risk-register.md`, and this checklist so the gate runner sees the full decision record before any gate reaches `promoted`.
- **Implemented:** Checklist entries exist, but required command outputs and approvals are still pending; statuses remain `blocked` where appropriate until we capture the missing evidence.
