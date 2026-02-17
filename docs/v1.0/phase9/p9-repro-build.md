# Phase 9 - Reproducible Build and Binary Verification Readiness

Phase 9 closes V10-G9 by cataloging deterministic build artifacts, signing checksums, and independent rebuild verification steps.

## Evidence anchor table
| Anchor | Description | Evidence |
|---|---|---|
| VA-R1 | Deterministic build and signing checklist | `pkg/v10/repro/verification.go` manifest functions + this doc’s checklist table |
| VA-R2 | Independent rebuild verification contract | `pkg/v10/repro/verification.go` verification steps + `releases/VA-B1-release-manifest.md` verification output section |
| VA-B1 | Release manifest for reproducible binaries | `releases/VA-B1-release-manifest.md` (checksum + provenance) |
| VA-B7 | Container baseline for evidence anchoring | `containers/v1.0/docker-compose.yml` + `containers/v1.0/deployment-runbook.md` |

## Planned vs Implemented
- **Planned:** Show deterministic build/signing steps, ensure independent rebuild coverage, and tie container evidence to the release manifest.
- **Implemented:** The `pkg/v10/repro` package codifies the deterministic steps, this doc enumerates the checklist, and the release/container artifacts document the actual signatures.

## Notes
- The CLI scenario inspects the same build manifest when assembling its `summary` before declaring PASS.
