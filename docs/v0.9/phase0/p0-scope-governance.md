# Phase 0 · Scope, Governance, and Evidence Foundation

## Plan vs implementation
Scope, compatibility, and gate-evidence discipline remain the primary plan deliverables. Implementation consists of the deterministic conformance catalog under `pkg/v09/conformance/gates.go`, the supporting README mapping VA IDs, and this phase document; nothing beyond that catalog (no runtime network behavior) is claimed as implemented.

## Evidence anchors
| VA ID | Artifact | Files | Description |
|---|---|---|---|
| VA-G1 | Scope trace map | `docs/v0.9/README.md`, `docs/v0.9/phase0/p0-scope-governance.md` | Maps the eight v0.9 bullets to phases and evidence, ensuring no scope drift. |
| VA-G2 | Governance checklist | `pkg/v09/conformance/gates.go` | Enumerates V9-G0..V9-G9, linkable checklists, and summary helpers for gate owners. |
| VA-LG1 | Licensing posture | `docs/v0.9/phase0/p0-scope-governance.md` | Notes AGPL/CC-BY-SA defaults plus transition evidence anchors as required of RM-04. |

## Planned-vs-implemented language
- Planned: future gates must document proof-per-gate via the `ValidateChecklist` helper before claiming closure.
- Implemented: `GateCatalog()` and `ValidateChecklist()` expose the additive-only AND evidence-bundle schema described in the checklist and make it re-runnable via the CLI witness.

## Legal posture snapshot
AGPL remains the code license and CC-BY-SA the normative spec license. There is no new legal language in this slice; instead, the `docs/v0.9/README.md` and `pkg/v09/governance/checklist.go` provide the audit helpers and traceable anchors for later consortium transition dialogues.
