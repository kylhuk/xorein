# TODO v12 - Xorein Execution Plan (v1.2)

## Status
Planning artifact only. This file defines v12 implementation and validation requirements. It does not claim implementation completion.

## Version Isolation Contract (mandatory)
- v12 cannot be promoted unless all in-scope spec, code, tests, docs, ops artifacts, and evidence are complete.
- v12 implements only features specified in v11 (`F12`).
- v12 also must produce full spec package for v13 (`F13`).

## Version role
- Implements: `F12` (unique immutable identity + no-password-reset policy + local backup recovery).
- Specifies: `F13` (Spaces lifecycle + text channels/chat baseline).

## Critical scope (v12)
- Implement globally unique immutable identity generation and persistence.
- Implement explicit no-password-reset behavior (lost password without backup is unrecoverable).
- Implement local backup export/import (`BackupID + BackupPassword`) for identity and user configuration.
- Implement minimal Gio identity onboarding and restore flows.
- Validate identity and recovery flows with unit, integration, and e2e tests.
- Preserve relay no-data-hosting boundary with explicit regression checks.

## Out of scope (defer)
- Voice/video/screen-share runtime features and advanced moderation/discovery features.
- Non-critical parity features remain deferred to `v20+`.

## Dependencies and relationships
- Inputs from v11: `docs/v1.1/phase4/f12-identity-backup-spec.md`, `docs/v1.1/phase4/f12-proto-delta.md`, `docs/v1.1/phase4/f12-acceptance-matrix.md`.
- Outputs consumed by v13: `docs/v1.2/phase4/f13-spaces-chat-spec.md`, `docs/v1.2/phase4/f13-proto-delta.md`, `docs/v1.2/phase4/f13-acceptance-matrix.md`.

## Entry criteria (must be true before implementation starts)
- `v11` is in `promoted` state with evidence bundle and as-built conformance report.
- Required `F12` spec inputs from v11 exist and are approved.
- v12 deferred scope list is frozen and approved.

## Promotion gates (must all pass)
- `G0` v12 scope lock and dependency freeze.
- `G1` Additive wire compatibility checks pass.
- `G2` Identity and backup runtime implementation complete.
- `G3` Gio onboarding/restore flows complete.
- `G4` Test matrix (positive, negative, recovery) complete.
- `G5` Podman scenario validations complete.
- `G6` v13 spec package complete.
- `G7` Docs and evidence bundle complete.
- `G8` Relay no-data-hosting regression checks pass.
- `G9` `F12` as-built conformance report completed against v11 specification package.

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v12/...` (or declared equivalent if paths differ)
- `go test ./tests/perf/v12/...` (or declared equivalent if paths differ)
- `make check-full`
- `scripts/v12-recovery-scenarios.sh`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v12-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## Phase plan

### Phase 0 - Scope lock and spec import (G0)
- [x] `P0-T1` Freeze v12 implementation scope to `F12`.
  - `ST1` Map each requirement to exact code/test/docs artifacts.
  - `ST2` Freeze non-goals and escalation path.
  - `ST3` Produce requirement-to-artifact traceability matrix.
  - `ST4` Assign gate ownership and approvers using RACI template.
  - Artifacts: `docs/v1.2/phase0/p0-scope-lock.md`, `docs/v1.2/phase0/p0-traceability-matrix.md`, `docs/v1.2/phase0/p0-gate-ownership.md`.

### Phase 1 - Identity and backup implementation (G2)
- [x] `P1-T1` Implement immutable identity lifecycle.
  - `ST1` Add identity creation flow with uniqueness guarantees.
  - `ST2` Persist identity metadata and key references locally.
  - `ST3` Add deterministic errors for duplicate/corrupt identity state.
  - Artifacts: `pkg/v12/identity/*`, `pkg/v12/identity/*_test.go`.

- [x] `P1-T2` Implement local backup model.
  - `ST1` Define backup payload format for identity + critical configuration.
  - `ST2` Implement backup export with `BackupID + BackupPassword` using versioned envelope (`Argon2id` KDF + AEAD).
  - `ST3` Implement backup import and restore verification checks.
  - `ST4` Ensure no server-side password reset route is exposed.
  - Artifacts: `pkg/v12/backup/*`, `pkg/v12/backup/*_test.go`, `docs/v1.2/phase1/p1-backup-format.md`.

### Phase 2 - Proto and client integration (G1, G3)
- [x] `P2-T1` Apply additive proto updates for identity and recovery metadata.
  - `ST1` Add fields/messages without renumbering existing fields.
  - `ST2` Reserve deprecated numbers/names as needed.
  - `ST3` Regenerate code and validate schema compatibility.
  - Artifacts: `proto/aether.proto`, `gen/go/aether/*`, `docs/v1.2/phase2/p2-proto-changelog.md`.

- [x] `P2-T2` Implement Gio identity onboarding and recovery UX.
  - `ST1` New-user path: create identity, explain no-reset policy.
  - `ST2` Recovery path: import backup, validate password, show deterministic errors.
  - `ST3` Add configuration restore controls and confirmation UI.
  - Artifacts: `pkg/v12/ui/*`, `pkg/ui/*`, `docs/v1.2/phase2/p2-ux-contract.md`.

### Phase 3 - Validation matrix (G4, G5)
- [x] `P3-T1` Add unit and integration tests.
  - `ST1` Identity uniqueness and immutability tests.
  - `ST2` Wrong password, corrupt backup, truncated backup tests.
  - `ST3` Migration and restart consistency tests.
  - Artifacts: `pkg/v12/**/*_test.go`.

- [x] `P3-T2` Add e2e and Podman recovery scenarios.
  - `ST1` New device restore scenario.
  - `ST2` Lost password without backup scenario (expected unrecoverable).
  - `ST3` Backup tamper detection scenario.
  - `ST4` Relay no-data-hosting regression scenario under identity/backup flows.
  - `ST5` Deterministic pass/fail probes with result manifest output.
  - Artifacts: `tests/e2e/v12/*`, `containers/v1.2/*`, `scripts/v12-recovery-scenarios.sh`, `docs/v1.2/phase3/p3-podman-scenarios.md`.

### Phase 4 - v13 spec package (G6)
- [x] `P4-T1` Define Spaces lifecycle spec.
  - `ST1` Space create semantics, founder/admin auto-assignment.
  - `ST2` Visibility default and join policy defaults.
  - `ST3` Space metadata and channel model draft.
  - Artifacts: `docs/v1.2/phase4/f13-spaces-chat-spec.md`.

- [x] `P4-T2` Define text-channel/chat baseline spec.
  - `ST1` Channel contracts, send/receive states, read markers.
  - `ST2` Failure and recovery reason taxonomy.
  - `ST3` Acceptance matrix for v13 implementation closure.
  - Artifacts: `docs/v1.2/phase4/f13-chat-flows.md`, `docs/v1.2/phase4/f13-proto-delta.md`, `docs/v1.2/phase4/f13-acceptance-matrix.md`.

### Phase 5 - Closure and evidence (G7)
- [x] `P5-T1` Publish v12 evidence bundle.
  - `ST1` Command outputs for compatibility/test/e2e checks.
  - `ST2` Podman scenario outputs and deterministic result manifests.
  - `ST3` `F12` as-built conformance report against v11 `F12` specs.
  - `ST4` Risk closure and promotion recommendation.
  - `ST5` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts: `docs/v1.2/phase5/p5-evidence-bundle.md`, `docs/v1.2/phase5/p5-risk-register.md`, `docs/v1.2/phase5/p5-as-built-conformance.md`, `docs/v1.2/phase5/p5-gate-signoff.md`, `docs/v1.2/phase5/p5-evidence-index.md`.

## Risk register (v12)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R12-1 | Ambiguous identity uniqueness semantics | Explicit uniqueness invariant and tests | Property tests pass |
| R12-2 | Weak backup security defaults | Strict encryption profile and tamper tests | Security tests pass |
| R12-3 | Users expect password reset | Mandatory UX warning and docs | UX tests verify warning path |
| R12-4 | Relay boundary regresses during identity rollout | Dedicated regression checks in e2e and Podman scenarios | Relay-boundary regression tests pass |

## Decision log (v12)
- `D12-1`: Backup in v12 is local-only by default.
- `D12-2`: Recovery is possible only via backup; no server reset pathway.
- `D12-3`: Identity and backup features must not broaden relay durable storage permissions.
