# Phase 6 - Client Distribution Readiness

Phase 6 aligns the App Store/Play Store/Microsoft Store/Flathub submissions with release-quality evidence (V10-G6).

## Evidence anchor table
| Anchor | Description | Evidence |
|---|---|---|
| VA-A1 | App store submission dossier structure | `pkg/v10/release/manifest.go` distribution checklist + `docs/v1.0/phase6/p6-app-distribution.md` dossier table |
| VA-A2 | Platform policy mapping and compliance | `pkg/v10/release/manifest.go` compliance map |
| VA-A3 | Rejection/resubmission workflow | `releases/VA-B1-release-manifest.md` rejection/snippet section + `containers/v1.0/deployment-runbook.md` recovery steps |

## Planned vs Implemented
- **Planned:** Standardize app-store submission dossiers and compliance mapping as part of the release handoff, ensuring each platform has a deterministic evidence anchor.
- **Implemented:** The distribution checklist lives in `pkg/v10/release` and is echoed in the release manifest; this doc records the dossier while referencing the same evidence that the CLI scenario reports back as part of its summary.

## Notes
- Platform-specific compliance statements remain explicit to avoid future governance confusion.
