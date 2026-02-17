# Phase 0 - Scope, Governance, and Evidence Foundation

This document closes the V10-G0 scope, governance, and evidence-schema gate with deterministic in-repo anchors. Everything documented below references the artifacts that implement the policy rather than an external process.

## Evidence anchor table
| Anchor | Description | Evidence |
|---|---|---|
| VA-G1 | Scope trace baseline aligning the eight v1.0 bullets to delivery phases | `docs/v1.0/phase0/p0-scope-governance.md` + `pkg/v10/conformance/gates.go` gate catalog |
| VA-G2 | Exclusion policy + escalation path for post-v1 scope | `pkg/v10/governance/policies.go` naming and scope limits |
| VA-G3 | Additive protobuf checklist | `pkg/v10/conformance/gates.go` checklist helpers referenced from scenario |
| VA-G4 | Major-path trigger checklist + governance classifier | `pkg/v10/governance/policies.go` major-path classifier helpers |
| VA-G5 | Evidence matrix template for integrated validation | `docs/v1.0/phase0/p0-scope-governance.md` table + release docs `releases/VA-H1-handoff-dossier.md` |
| VA-G6 | Gate evidence bundle conventions | `containers/v1.0/deployment-runbook.md` (evidence section) |

## Planned vs Implemented
- **Planned:** Align scope with the eight authoritative v1.0 bullets, lock exclusions, and define the gate/evidence schema before any downstream phase documents.
- **Implemented:** This doc plus `pkg/v10/conformance` enshrine the gate catalog, checklist patterns, and references. The CLI scenario uses `pkg/v10/governance` to assert compliance, and release/container artifacts cite the same anchors when describing evidence packaging.

## Notes
- Deterministic evidence surfaces are limited to these in-repo files so reviewers can follow the entire trace from `TODO_v10.md` to `pkg/v10/scenario` to the release manifests.
