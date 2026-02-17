# Release Manifest VA-B1 – Release Bundle Checklist

This manifest captures the deterministic release bundle evidence for VA-B* and VA-R* anchors referenced throughout `TODO_v10.md`.

## Checklist
| Item | Anchor | Evidence |
|---|---|---|
| Gate catalog recorded | VA-B1 | `pkg/v10/conformance/gates.go` catalog + CLI scenario uses the same catalog |
| Distribution dossier | VA-A1–VA-A3 | `pkg/v10/release/manifest.go` plus `docs/v1.0/phase6/p6-app-distribution.md` |
| Build/repro provenance | VA-R1–VA-R2 | `pkg/v10/repro/verification.go` manifest + `docs/v1.0/phase9/p9-repro-build.md` |
| Container baseline | VA-B7 | `containers/v1.0/docker-compose.yml`, `containers/v1.0/deployment-runbook.md` |

## Compliance matrix <small>(app store + Flathub)</small>
| Platform | Compliance focus | Evidence |
|---|---|---|
| Google Play | Security + compatibility | `pkg/v10/security/lifecycle.go` severity table and `docs/v1.0/phase3/p3-spec-publication.md` compatibility chapter |
| App Store | Privacy & review workflow | `pkg/v10/security/engagement.go` selection rubric, `docs/v1.0/phase6/p6-app-distribution.md` dossier |
| Microsoft Store | Quality metadata | `pkg/v10/docs/docs.go` developer guide structure + `website/landing.md` claims |
| Flathub | Audit readiness | `docs/v1.0/phase2/p2-security-audit.md` threat model + `releases/VA-H1-handoff-dossier.md` disclosure list |

## Build verification summary
- Deterministic command: `pkg/v10/repro.DeterministicBuildCommand()` (printed by scenario) ensures signature alignment.
- Checksum manifest: `pkg/v10/repro.BuildPins()` aligns with `releases/v1.0-checksums.txt` generated from:
  - `artifacts/generated/release-pack/v1.0/aether`: `92b54a25d7ffbce4c4261130ff9769a7dd394284751a6ab4ccc2344078f30d26`
  - `artifacts/generated/release-pack/v1.0/aether-push-relay`: `864f7f323a4526e799a6bfe38f912243ad6d620cf0690521f641bf0bf1ad94ce`
