# Phase 2 - External Security Audit Readiness and Governance

Security-audit readiness is anchored on a scoped asset inventory, threat model, and governance for engagement/resolution. The deterministic helpers below serve as proofs for V10-G2.

## Evidence anchor table
| Anchor | Description | Evidence |
|---|---|---|
| VA-S1 | Security audit asset inventory coverage | `pkg/v10/security/audit.go` asset map and `docs/v1.0/phase2/p2-security-audit.md` narrative |
| VA-S2 | Threat-model package and questionnaire template | `pkg/v10/security/threat.go` template data |
| VA-S3 | Auditor engagement framework | `pkg/v10/security/engagement.go` selection rubric |
| VA-S4 | Finding lifecycle taxonomy and disclosure gates | `pkg/v10/security/lifecycle.go` severity map + release docs `releases/VA-H1-handoff-dossier.md` disclosure table |
| VA-S5 | Self-assessment readiness checklist | `pkg/v10/conformance/gates.go` checklist usage + CLI scenario log |
| VA-S6 | Security readiness dossier | `releases/VA-H1-handoff-dossier.md` aggregated security section, `docs/v1.0/phase2/p2-security-audit.md` residual-risk table |

## Planned vs Implemented
- **Planned:** Define the audit scope, threat model, engagement criteria, and remediation lifecycle before handing gate V10-G2 over to ops.
- **Implemented:** The `pkg/v10/security` helpers encode the same information and feed the CLI scenario along with release/container manifests; the courtesy doc here plus `releases/VA-H1-handoff-dossier.md` constitute the deterministic dossier.

## Notes
- Every VA-S* anchor is tied to assets stored in-tree (packages plus release docs). The scenario leverages these helpers to confirm readiness before even printing `v1.0 genesis: PASS`.
