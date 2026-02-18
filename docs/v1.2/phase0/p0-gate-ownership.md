# v1.2 Phase 0 - Gate Ownership and Sign-off Roles

## Gate owners
| Gate | Owner role | Required approver roles |
|---|---|---|
| G0 | Planning lead | Release authority |
| G1 | Protocol engineer | Protocol steward |
| G2 | Runtime engineer | Security reviewer |
| G3 | Client engineer | QA lead |
| G4 | QA lead | Runtime engineer |
| G5 | Ops engineer | Runtime engineer |
| G6 | Spec editor | Protocol steward |
| G7 | Release manager | Release authority |
| G8 | QA lead | Protocol steward |
| G9 | Release manager | Release authority |

## Sign-off workflow
- Use `docs/templates/roadmap-signoff-raci.md` as the canonical role matrix.
- Each promoted gate requires evidence IDs (`EV-v12-GX-###`) and UTC timestamp.
- Fail-close rule applies: non-promoted gates block v1.2 promotion.

## Planned vs implemented
- Ownership is finalized for execution.
- Individual sign-off records are published in `docs/v1.2/phase5/p5-gate-signoff.md`.
