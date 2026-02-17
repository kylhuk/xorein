# TODO v15 - Xorein Execution Plan (v1.5)

## Status
Planning artifact only. This file defines v15 implementation and validation requirements. It does not claim implementation completion.

## Version Isolation Contract (mandatory)
- v15 cannot advance unless all v15 in-scope spec, code, tests, docs, ops artifacts, and evidence are complete.
- v15 implements only `F15` from v14 specs.
- v15 must also publish full `F16` spec package for v16 implementation.

## Version role
- Implements: `F15` (screen share baseline and adaptation contracts).
- Specifies: `F16` (full RBAC + ACL model and enforcement contracts).

## Critical scope (v15)
- Implement screen share capture, transport, and viewer pipeline runtime.
- Implement screen share quality presets and adaptive fallback behavior.
- Integrate secure media path requirements for screen share.
- Implement Gio screen share controls and status UX.
- Validate first-frame, adaptation, and recovery behavior.
- Preserve relay no-data-hosting boundary with explicit regression checks.

## Out of scope (defer)
- RBAC and ACL implementation is deferred to v16 implementation scope.
- Non-critical parity features remain deferred to `v20+`.

## Dependencies and relationships
- Inputs from v14: `docs/v1.4/phase4/f15-screenshare-spec.md`, `docs/v1.4/phase4/f15-proto-delta.md`, `docs/v1.4/phase4/f15-acceptance-matrix.md`.
- Outputs consumed by v16: `docs/v1.5/phase4/f16-rbac-acl-spec.md`, `docs/v1.5/phase4/f16-proto-delta.md`, `docs/v1.5/phase4/f16-acceptance-matrix.md`.

## Entry criteria (must be true before implementation starts)
- `v14` is in `promoted` state with evidence bundle and as-built conformance report.
- Required `F15` spec inputs from v14 exist and are approved.
- v15 deferred scope list is frozen and approved.

## Promotion gates (must all pass)
- `G0` Scope lock and artifact map complete.
- `G1` Proto compatibility checks pass.
- `G2` Screen share runtime complete.
- `G3` Gio control and viewer UX complete.
- `G4` Quality/adaptation/recovery test matrix complete.
- `G5` Podman screen-share scenarios complete.
- `G6` v16 spec package complete.
- `G7` Docs and evidence complete.
- `G8` Relay no-data-hosting regression checks pass.
- `G9` `F15` as-built conformance report completed against v14 specification package.

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v15/...` (or declared equivalent if paths differ)
- `go test ./tests/perf/v15/...` (or declared equivalent if paths differ)
- `make check-full`
- `scripts/v15-screenshare-scenarios.sh`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v15-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## Phase plan

### Phase 0 - Scope lock (G0)
- [ ] `P0-T1` Freeze v15 implementation and acceptance criteria.
  - `ST1` Produce requirement-to-artifact traceability matrix.
  - `ST2` Assign gate ownership and approvers using RACI template.
  - Artifacts: `docs/v1.5/phase0/p0-scope-lock.md`, `docs/v1.5/phase0/p0-traceability-matrix.md`, `docs/v1.5/phase0/p0-gate-ownership.md`.

### Phase 1 - Screen share runtime (G2)
- [ ] `P1-T1` Implement capture and encode pipeline.
  - `ST1` Capture source selection (display/window).
  - `ST2` Preset-to-encoder/bitrate mapping.
  - `ST3` Runtime error classes and recovery actions.
  - Artifacts: `pkg/v15/capture/*`, `pkg/v15/capture/*_test.go`.

- [ ] `P1-T2` Implement transport and adaptation logic.
  - `ST1` Stream start/stop and renegotiation paths.
  - `ST2` Adaptive layer fallback for uplink degradation.
  - `ST3` Deterministic viewer state transitions.
  - Artifacts: `pkg/v15/screenshare/*`, `pkg/v15/screenshare/*_test.go`.

### Phase 2 - Client integration and UX (G3)
- [ ] `P2-T1` Implement Gio screen share UX.
  - `ST1` Start/stop/source-select controls.
  - `ST2` Live quality indicator and degrade hints.
  - `ST3` Recovery-first reconnection cues.
  - Artifacts: `pkg/v15/ui/*`, `docs/v1.5/phase2/p2-screenshare-ux-contract.md`.

### Phase 3 - Validation matrix (G4, G5)
- [ ] `P3-T1` Add screen share test suites.
  - `ST1` First-frame target tests.
  - `ST2` Multi-viewer adaptation tests.
  - `ST3` Network degradation and recovery tests.
  - Artifacts: `tests/e2e/v15/*`, `tests/perf/v15/*`.

- [ ] `P3-T2` Add Podman end-to-end scenarios.
  - `ST1` Multi-peer share/start/stop scenarios in container network.
  - `ST2` Relay no-data-hosting regression scenario for screen-share paths.
  - `ST3` Deterministic pass/fail probes and result manifest output.
  - Artifacts: `containers/v1.5/*`, `scripts/v15-screenshare-scenarios.sh`, `docs/v1.5/phase3/p3-podman-scenarios.md`.

### Phase 4 - v16 spec package (G6)
- [ ] `P4-T1` Produce RBAC + ACL specification package.
  - `ST1` Role model (founder/admin/mod/member/guest + custom roles).
  - `ST2` ACL merge and precedence rules.
  - `ST3` Channel-level override semantics.
  - Artifacts: `docs/v1.5/phase4/f16-rbac-acl-spec.md`, `docs/v1.5/phase4/f16-proto-delta.md`, `docs/v1.5/phase4/f16-acceptance-matrix.md`.

### Phase 5 - Closure and evidence (G7)
- [ ] `P5-T1` Publish v15 evidence bundle and promotion recommendation.
  - `ST1` Command outputs for compatibility/test/e2e/perf checks.
  - `ST2` Podman scenario outputs and deterministic result manifests.
  - `ST3` `F15` as-built conformance report against v14 `F15` specs.
  - `ST4` Risk closure and promotion recommendation.
  - `ST5` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts: `docs/v1.5/phase5/p5-evidence-bundle.md`, `docs/v1.5/phase5/p5-risk-register.md`, `docs/v1.5/phase5/p5-as-built-conformance.md`, `docs/v1.5/phase5/p5-gate-signoff.md`, `docs/v1.5/phase5/p5-evidence-index.md`.

## Risk register (v15)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R15-1 | Encode pipeline instability across hardware | Deterministic fallback and compatibility matrix | Hardware matrix tests pass |
| R15-2 | Viewer desync during adaptation | Canonical state machine and recovery tests | E2E adaptation tests pass |
| R15-3 | UX ambiguity in degraded mode | Stable reason taxonomy in UI | No-limbo tests pass |
| R15-4 | Relay boundary regresses during screen-share rollout | Dedicated regression checks in e2e and Podman scenarios | Relay-boundary regression tests pass |

## Decision log (v15)
- `D15-1`: Screen share must degrade gracefully before failing terminally.
- `D15-2`: Screen-share implementation must not broaden relay durable storage permissions.
