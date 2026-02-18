# Phase 0 - Gate ownership and approvals (Planning only)

This planning artifact assigns RACI roles and documents the per-gate owner/approver expectations for v11 promotion (G0..G7). Implementation deliverables will cite this sheet when opening sign-off records.

## RACI matrix (adapted from `docs/templates/roadmap-signoff-raci.md`)
| Gate | Responsible | Accountable | Consulted | Informed |
|---|---|---|---|---|
| G0 Scope + dependencies | Plan Lead | Plan Lead | Protocol Lead, QA Lead | Release Authority |
| G1 Compatibility + schema | Protocol Lead | Protocol Lead | Runtime Lead, QA Lead | Release Authority |
| G2 Gate runner implementation | Runtime Lead, Client Lead | Runtime Lead | Protocol Lead, QA Lead | Release Authority |
| G3 Relay boundary verification | Runtime Lead | Runtime Lead | Protocol Lead, QA Lead | Release Authority |
| G4 Podman smoke validation | Ops Lead, Runtime Lead | QA Lead | Client Lead, Protocol Lead | Release Authority |
| G5 Identity/backup spec | Protocol Lead | Plan Lead | Security Lead, QA Lead | Release Authority |
| G6 Docs/evidence bundle | QA Lead, Plan Lead | Plan Lead | Protocol Lead, Runtime Lead, Ops Lead, Security Lead | Release Authority |
| G7 Conformance report | Plan Lead | Release Authority | QA Lead, Protocol Lead | Runtime Lead, Ops Lead |

## Per-gate owner and approver records (planning placeholders)
- **G0 Scope lock:**
  - Owner role: Plan Lead
  - Approver roles: Plan Lead, Protocol Lead
  - Approver identifiers/timestamps: _TBD pending planning review_
  - Evidence anchors: EV-v11-G0-001 (scope lock narrative), EV-v11-G0-002 (mutually exclusive buckets)
- **G1 Compatibility policy:**
  - Owner role: Protocol Lead
  - Approver roles: Protocol Lead, Runtime Lead, QA Lead
  - Approver identifiers/timestamps: _TBD pending compatibility policy doc_
  - Evidence anchor: EV-v11-G1-001
- **G2 Gate runner:**
  - Owner role: Runtime Lead
  - Approver roles: Runtime Lead, Client Lead
  - Approver identifiers/timestamps: _TBD once pkg/v11/gates/ status output exists_
  - Evidence anchor: EV-v11-G2-001
- **G3 Relay boundary:**
  - Owner role: Runtime Lead
  - Approver roles: Runtime Lead, Protocol Lead
  - Approver identifiers/timestamps: _TBD after relay policy verification doc_
  - Evidence anchor: EV-v11-G3-001
- **G4 Podman smoke:**
  - Owner role: Ops Lead
  - Approver roles: Ops Lead, QA Lead
  - Approver identifiers/timestamps: _TBD after `scripts/v11-relay-smoke.sh` run plan_
  - Evidence anchor: EV-v11-G4-001
- **G5 Identity/backup spec:**
  - Owner role: Plan Lead
  - Approver roles: Protocol Lead, Security Lead
  - Approver identifiers/timestamps: _TBD after `docs/v1.1/phase4/f12-identity-backup-spec.md` review_
  - Evidence anchor: EV-v11-G5-001
- **G6 Docs/evidence bundle:**
  - Owner role: QA Lead
  - Approver roles: QA Lead, Plan Lead
  - Approver identifiers/timestamps: _TBD when `docs/v1.1/phase5/p5-evidence-bundle.md` is drafted_
  - Evidence anchors: EV-v11-G6-001, EV-v11-G6-002 (index)
- **G7 Conformance report:**
  - Owner role: Release Authority
  - Approver roles: Release Authority, Plan Lead
  - Approver identifiers/timestamps: _TBD with as-built report completion_
  - Evidence anchor: EV-v11-G7-001

## Notes
- Every gate row will ultimately include actual owner names, timestamps, approver signatures, and evidence links following the template in `docs/templates/roadmap-signoff-raci.md`.
- Plug in EV-v11-GX-### IDs consistently once command/test outputs are captured in Phase 5.
