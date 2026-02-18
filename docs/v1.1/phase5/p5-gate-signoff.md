# Phase 5 - Gate signoff (v11 closure)

## Gate workflow reminders
- Gate statuses follow `docs/templates/roadmap-gate-checklist.md` (`open`, `blocked`, `ready_for_review`, `promoted`).
- Single-operator closure for v11: El Pollo Diablo fills owner and approver responsibilities while preserving role labels for traceability.

## Gate checklist
| GateID | Purpose | Entry checks | Exit checks | Owner role | Required approvers | Evidence IDs | Status | Notes |
|---|---|---|---|---|---|---|---|---|
| G0 | Scope lock and traceability | Phase0 scope + trace matrix exists | Phase0 docs cite gate runner contract | Plan Lead | QA Lead, Release Authority | EV-v11-G0-001 | promoted | Governance scope lock verified and signed. |
| G1 | Compatibility policy | `p1-compatibility-policy.md` drafted | `buf lint` + `buf breaking` logs captured | Protocol Lead | QA Lead, Release Authority | EV-v11-G1-001, EV-v11-G1-002 | promoted | Compatibility checks pass; lint warning documented as non-blocking follow-up. |
| G2 | Gate runner | `phase1/p1-gate-runner.md` describes runner | Runner software executed against gate artifacts | Runtime Lead | Client Lead, QA Lead | EV-v11-G2-001 | promoted | Gate runner reports all gates promoted and fresh. |
| G3 | Relay data boundary | Phase2 limit doc signed off | Scenario pack data + policy asserts | Runtime Lead | Protocol Lead, QA Lead | EV-v11-G3-001 | promoted | Relay policy and v11 boundary tests pass. |
| G4 | Podman smoke | Podman smoke doc + scripts referenced | Relay smoke script output + Go test evidence | QA Lead | Runtime Lead, Plan Lead | EV-v11-G4-001, EV-v11-G4-002, EV-v11-G4-003 | promoted | Podman smoke and e2e evidence are captured and passing. |
| G5 | Evidence bundle & release handoff | Phase5 docs assembled | Evidence index, risk register, gate signoff updated | Release Authority | All leads | EV-v11-G5-003, EV-v11-G5-004, EV-v11-G5-005 | promoted | Mandatory command evidence set is complete and passing. |
| G6 | Docs/evidence completeness | Gate checklist + evidence index drafted | Mandatory evidence table complete with checksums and ownership | Plan Lead | QA Lead, Release Authority | EV-v11-G6-001 | promoted | Closure docs and no-active-deferrals register are present and linked. |
| G7 | F11 as-built conformance | `p5-as-built-conformance.md` drafted | Release recommendation reviewed against all gate states | Release Authority | Plan Lead, QA Lead | EV-v11-G7-001 | promoted | Final gate-runner pass and as-built recommendation confirm promotion. |

## Sign-off record
| GateID | Owner role | Approver role(s) | Approver identifier | TimestampUTC | Evidence IDs |
|---|---|---|---|---|---|
| G0 | Plan Lead | QA Lead, Release Authority | El Pollo Diablo | 2026-02-18T08:36:35Z | EV-v11-G0-001 |
| G1 | Protocol Lead | QA Lead, Release Authority | El Pollo Diablo | 2026-02-18T08:36:35Z | EV-v11-G1-001, EV-v11-G1-002 |
| G2 | Runtime Lead | Client Lead, QA Lead | El Pollo Diablo | 2026-02-18T08:36:35Z | EV-v11-G2-001 |
| G3 | Runtime Lead | Protocol Lead, QA Lead | El Pollo Diablo | 2026-02-18T08:36:35Z | EV-v11-G3-001 |
| G4 | QA Lead | Runtime Lead, Plan Lead | El Pollo Diablo | 2026-02-18T08:36:35Z | EV-v11-G4-001, EV-v11-G4-002, EV-v11-G4-003 |
| G5 | Release Authority | All leads | El Pollo Diablo | 2026-02-18T08:36:35Z | EV-v11-G5-003, EV-v11-G5-004, EV-v11-G5-005 |
| G6 | Plan Lead | QA Lead, Release Authority | El Pollo Diablo | 2026-02-18T08:36:35Z | EV-v11-G6-001 |
| G7 | Release Authority | Plan Lead, QA Lead | El Pollo Diablo | 2026-02-18T08:36:35Z | EV-v11-G7-001 |

## RACI notes
- This file references `docs/templates/roadmap-signoff-raci.md` to keep ownership labels consistent with the standard workflow.
- The v11 single-operator signoff retains role labels while using one named approver identity.

## Planned vs implemented
- **Planned:** If evidence is refreshed later, update signoff timestamps and gate notes before any further promotion claim.
- **Implemented:** Gates G0..G7 are promoted with evidence-linked signoff records.
