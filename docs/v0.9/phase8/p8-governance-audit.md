# Phase 8 · Governance & Compatibility Readiness

## Plan vs implementation
Phase 8 maintains the additive checklist, major-path trigger classifier, and licensing/status language audits. Implementation lives entirely in `pkg/v09/governance/checklist.go`; this document records how the helpers satisfy `VA-X*` artifacts and what remains planned (additional legal reviews once the consortium paperwork surfaces).

## Evidence anchors
| VA ID | Artifact | Files | Description |
|---|---|---|---|
| VA-X1 | Additive checklist | `pkg/v09/governance/checklist.go` | `AdditiveChecklist` enumerates the additive evolution requirements. |
| VA-X2 | Major-path classifier | `pkg/v09/governance/checklist.go` | `MajorPathTriggerClassifier` determines when AEP/implementation proof is required. |
| VA-X3 | Licensing audit helper | `pkg/v09/governance/checklist.go` | `LicensingStatus` verifies AGPL/CC-BY-SA posture plus deferral reminders. |
| VA-X7/VA-X8 | Policy hardening | `docs/v0.9/phase8/p8-governance-audit.md` | Notes future policy hardening timelines and legal review anchors while preserving plan vs implementation language. |

## Planned-vs-implemented language
- Planned: formal legal review packages remain deferred; this doc now lists the trace anchors expected when policy hardening moves forward.
- Implemented: the governance helpers are deterministic and exercised through the scenario, so reviewers can rerun the additive/major-path classification logic without extra infrastructure.
