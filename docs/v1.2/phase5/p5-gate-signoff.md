# v1.2 Phase 5 - Gate Sign-off Sheet

## Gate states
- `open`
- `blocked`
- `ready_for_review`
- `promoted`

## Sign-off table
| Gate | Owner role | Approver roles | Evidence IDs | Status | Timestamp (UTC) | Notes |
|---|---|---|---|---|---|---|
| G0 | Planning lead | Release authority | EV-v12-G0-001..003 | promoted | 2026-02-18T14:45:00Z | Scope lock and dependency freeze complete |
| G1 | Protocol engineer | Protocol steward | EV-v12-G1-001..002 | promoted | 2026-02-18T14:45:00Z | Additive proto checks passed |
| G2 | Runtime engineer | Security reviewer | EV-v12-G2-001 | promoted | 2026-02-18T14:45:00Z | Identity and backup runtime implementation complete |
| G3 | Client engineer | QA lead | EV-v12-G3-001 | promoted | 2026-02-18T14:45:00Z | Onboarding and restore UX contract complete |
| G4 | QA lead | Runtime engineer | EV-v12-G4-001..003 | promoted | 2026-02-18T14:45:00Z | Unit, e2e, and perf evidence passed |
| G5 | Ops engineer | Runtime engineer | EV-v12-G5-001 | promoted | 2026-02-18T14:45:00Z | Podman recovery scenarios passed with manifest |
| G6 | Spec editor | Protocol steward | EV-v12-G6-001..003 | promoted | 2026-02-18T14:45:00Z | F13 specification package complete |
| G7 | Release manager | Release authority | EV-v12-G7-001 | promoted | 2026-02-18T14:45:00Z | Docs and evidence bundle complete |
| G8 | QA lead | Protocol steward | EV-v12-G8-001 | promoted | 2026-02-18T14:45:00Z | Relay no-data-hosting regression checks passed |
| G9 | Release manager | Release authority | EV-v12-G9-001 | promoted | 2026-02-18T14:45:00Z | As-built conformance completed and recommended |

## Planned vs implemented
- Sign-off roles are assigned.
- All v1.2 promotion gates are marked `promoted` with linked evidence.
