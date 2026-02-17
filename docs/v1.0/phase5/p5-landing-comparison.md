# Phase 5 - Landing and Comparison Launch Surface Readiness

Phase 5 documents the landing/comparison surface required for V10-G5. All claims are tied to deterministic artifacts.

## Evidence anchor table
| Anchor | Description | Evidence |
|---|---|---|
| VA-W1 | Landing page content taxonomy and evidence anchors | `website/landing.md` plus `pkg/v10/website/landing.go` claims map |
| VA-W2 | Comparison page governance controls | `website/comparison.md` and `pkg/v10/docs/docs.go` governance cross-links |
| VA-W3 | Landing claim hygiene rules | this doc’s hygiene checklist and `pkg/v10/conformance/gates.go` gate summary |

## Planned vs Implemented
- **Planned:** Lock landing claim taxonomy and comparison governance before publishing any marketing surface.
- **Implemented:** The landing/comparison markdown files plus `pkg/v10/website/landing.go` assert the same claims and feed the CLI scenario’s `summary` so that PASS only occurs once claims are traceable.

## Notes
- Evidence anchors stay inside the repo to honor the planned-vs-implemented discipline.
