# v1.2 Phase 0 - Scope Lock

## Status
Scope is locked to `F12` implementation and `F13` specification handoff.

## In scope
- Immutable global identity creation and local persistence.
- No-password-reset behavior with explicit onboarding warning.
- Local backup export/import (`BackupID + BackupPassword`) with Argon2id + AEAD envelope.
- Gio-shell-level onboarding and restore flow state contract.
- Additive proto updates for identity and backup metadata.
- Unit, e2e, perf, and Podman scenario validation for v1.2.

## Deferred
- Hosted backup services or server reset workflows.
- Voice/video/screen-share runtime expansion.
- Discovery and advanced moderation/runtime parity features (`v20+`).

## Carry-back policy
- Any requirement not mapped below is deferred and must be approved through gate ownership.
- No deferred item is implemented in v1.2 code paths.

## Evidence anchors
| Evidence ID | Description | Path |
|---|---|---|
| EV-v12-G0-001 | Scope lock artifact | docs/v1.2/phase0/p0-scope-lock.md |
| EV-v12-G0-002 | Traceability matrix | docs/v1.2/phase0/p0-traceability-matrix.md |
| EV-v12-G0-003 | Gate ownership | docs/v1.2/phase0/p0-gate-ownership.md |

## Planned vs implemented
- This file documents governance closure for scope and does not assert runtime test outcomes.
