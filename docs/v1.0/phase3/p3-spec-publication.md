# Phase 3 - Protocol Specification Publication Readiness

Phase 3 closes V10-G3 by documenting the normative/informative architecture, compatibility/governance chapters, and publication workflow for the v1.0 spec.

## Evidence anchor table
| Anchor | Description | Evidence |
|---|---|---|
| VA-P7 | Spec architecture blueprint | `pkg/v10/publication/publication.go` section map that pairs informative/normative entries |
| VA-P8 | Terminology governance and glossary | `pkg/v10/governance/naming.go` canonical names + alias map |
| VA-P9 | Compatibility chapter requirements | `pkg/v10/governance/policies.go` compatibility classifier and additive rules |
| VA-P10 | Governance chapter requirements | `pkg/v10/governance/policies.go` governance chapter helper + open-decision table in this doc |
| VA-P11 | Publication bundle checklist and licensing assertions | `pkg/v10/release/manifest.go` bundler + `releases/VA-B1-release-manifest.md` checklist |
| VA-P12 | Publication approval workflow | `docs/v1.0/phase3/p3-spec-publication.md` workflow steps + CLI echo from `pkg/v10/scenario` finish path |

## Planned vs Implemented
- **Planned:** Blueprint the spec, define terminology/governance chapters, and lock the publication checklist + workflow before actual publication.
- **Implemented:** The blueprint lives in `pkg/v10/publication`, naming/governance helpers in `pkg/v10/governance`, and `pkg/v10/release` composes the bundle checklist used both by release docs and the CLI scenario.

## Notes
- RM-03 is closed and carried as a resolved baseline in `TODO_v10.md` and `open_decisions.md`.
