# Phase 0 - Traceability matrix for critical scope and gates

This planning artifact maps every critical v11 scope item and promotion gate to the future artifact or evidence placeholder that will carry the proof in later phases.

## Traceability table
| Scope item | Gate | Planned artifact | Evidence ID placeholder |
|---|---|---|---|
| Scope lock, dependency map, and deferred list | G0 | `docs/v1.1/phase0/p0-scope-lock.md` + planned review notes | EV-v11-G0-002 |
| Compatibility policy documentation (`additive-only`) | G1 | `docs/v1.1/phase1/p1-compatibility-policy.md` (planned) | EV-v11-G1-001 |
| Gate runner and status commands | G2 | `pkg/v11/gates/*` + `docs/v1.1/phase1/p1-gate-runner.md` | EV-v11-G2-001 |
| Relay no-data-hosting policy enforcement | G3 | `docs/v1.1/phase2/p2-relay-data-boundary.md` + `pkg/v11/relaypolicy/*` | EV-v11-G3-001 |
| Podman relay smoke checks (runtime) | G4 | `scripts/v11-relay-smoke.sh` + `docs/v1.1/phase3/p3-podman-smoke.md` | EV-v11-G4-001 |
| v12 identity and backup spec package | G5 | `docs/v1.1/phase4/f12-identity-backup-spec.md` + `docs/v1.1/phase4/f12-backup-recovery-flows.md` | EV-v11-G5-001 |
| Docs/evidence bundle completeness | G6 | `docs/v1.1/phase5/p5-evidence-bundle.md` + `docs/v1.1/phase5/p5-evidence-index.md` | EV-v11-G6-001 |
| `F11` as-built conformance report | G7 | `docs/v1.1/phase5/p5-as-built-conformance.md` + `docs/v1.1/phase5/p5-gate-signoff.md` | EV-v11-G7-001 |

## Notes
- Each evidence placeholder aligns with the EV-v11-GX-### schema defined in `docs/templates/roadmap-evidence-index.md`.
- Dependency edges between these gates and earlier versions are lifted from `TODO_v11.md` and will be updated in this table as executed artifacts and evidence IDs are produced.
