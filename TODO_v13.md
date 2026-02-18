# TODO v13 - Xorein Execution Plan (v1.3)

## Status
Planning artifact only. This file defines v13 implementation and validation requirements. It does not claim implementation completion.

## Version Isolation Contract (mandatory)
- v13 cannot advance unless all v13 in-scope spec, code, tests, docs, ops artifacts, and evidence are complete.
- v13 implements only `F13` from v12 specs.
- v13 must also publish full `F14` spec package for v14 implementation.

## Version role
- Implements: `F13` (Spaces lifecycle + text channels/chat baseline).
- Specifies: `F14` (voice channels baseline and recovery contracts).

## Critical scope (v13)
- Implement Space creation, founder/admin bootstrap, and management baseline.
- Implement defaults: Spaces visible by default; join policy invite-only by default.
- Implement join modes: invite-only, request-to-join, open.
- Implement text channels and baseline chat lifecycle in Gio client and runtime.
- Validate channel/chat reliability and policy enforcement.
- Preserve relay no-data-hosting boundary with explicit regression checks.

## Out of scope (defer)
- Voice/video/screen-share implementation work is deferred to later versions.
- Non-critical parity features remain deferred to `v20+`.

## Dependencies and relationships
- Inputs from v12: `docs/v1.2/phase4/f13-spaces-chat-spec.md`, `docs/v1.2/phase4/f13-chat-flows.md`, `docs/v1.2/phase4/f13-proto-delta.md`, `docs/v1.2/phase4/f13-acceptance-matrix.md`.
- Outputs consumed by v14: `docs/v1.3/phase4/f14-voice-spec.md`, `docs/v1.3/phase4/f14-proto-delta.md`, `docs/v1.3/phase4/f14-acceptance-matrix.md`.

## Entry criteria (must be true before implementation starts)
- `v12` is in `promoted` state with evidence bundle and as-built conformance report.
- Required `F13` spec inputs from v12 exist and are approved.
- v13 deferred scope list is frozen and approved.

## Promotion gates (must all pass)
- `G0` Scope lock and mapping to artifacts.
- `G1` Proto compatibility checks pass.
- `G2` Spaces and join-policy runtime complete.
- `G3` Text channel and chat runtime + Gio flow complete.
- `G4` Permission/policy/adverse test matrix complete.
- `G5` Podman e2e scenarios complete.
- `G6` v14 spec package complete.
- `G7` Docs and evidence complete.
- `G8` Relay no-data-hosting regression checks pass.
- `G9` `F13` as-built conformance report completed against v12 specification package.

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v13/...` (or declared equivalent if paths differ)
- `go test ./tests/perf/v13/...` (or declared equivalent if paths differ)
- `make check-full`
- `scripts/v13-e2e-podman.sh`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v13-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## Phase plan

### Phase 0 - Scope lock and import (G0)
- [x] `P0-T1` Freeze v13 requirements and acceptance matrix.
  - `ST1` Map each v13 requirement to code/test/docs.
  - `ST2` Lock deferred set (`v20+` backlog only).
  - `ST3` Produce requirement-to-artifact traceability matrix.
  - `ST4` Assign gate ownership and approvers using RACI template.
  - Artifacts: `docs/v1.3/phase0/p0-scope-lock.md`, `docs/v1.3/phase0/p0-traceability-matrix.md`, `docs/v1.3/phase0/p0-gate-ownership.md`.

### Phase 1 - Space lifecycle and policy engine (G2)
- [x] `P1-T1` Implement Space lifecycle runtime.
  - `ST1` Create Space flow with founder/admin auto-assignment.
  - `ST2` Persist Space metadata and policy state.
  - `ST3` Implement space-level update and ownership transfer guardrails.
  - Artifacts: `pkg/v13/spaces/*`, `pkg/v13/spaces/*_test.go`.

- [x] `P1-T2` Implement join policies.
  - `ST1` Invite-only baseline and invite token verification.
  - `ST2` Request-to-join approval/deny lifecycle.
  - `ST3` Open mode with abuse controls baseline.
  - Artifacts: `pkg/v13/joinpolicy/*`, `tests/e2e/v13/join_policy_test.go`.

### Phase 2 - Text channels and chat baseline (G3)
- [x] `P2-T1` Implement channel model and channel operations.
  - `ST1` Channel create/list/update/archive lifecycle.
  - `ST2` Space-channel bindings and membership checks.
  - Artifacts: `pkg/v13/channels/*`, `pkg/v13/channels/*_test.go`.

- [x] `P2-T2` Implement chat lifecycle.
  - `ST1` Send/receive/ack states and deterministic error reasons.
  - `ST2` Read marker and unread counter convergence rules.
  - `ST3` Draft persistence hooks for later continuity work.
  - Artifacts: `pkg/v13/chat/*`, `pkg/v13/chat/*_test.go`.

- [x] `P2-T3` Implement minimal Gio Spaces/chat UX.
  - `ST1` Space list, preview, join controls.
  - `ST2` Channel list, timeline, composer, delivery state UI.
  - `ST3` Deterministic no-limbo messaging states.
  - Artifacts: `pkg/v13/ui/*`, `docs/v1.3/phase2/p2-ux-contract.md`.

### Phase 3 - Validation matrix (G4, G5)
- [x] `P3-T1` Add test coverage for policy and messaging flows.
  - `ST1` Positive join and chat journeys.
  - `ST2` Negative policy denials and replay/invalid invite tests.
  - `ST3` Degraded path tests (transient disconnect and retry).
  - Artifacts: `tests/e2e/v13/*`, `tests/perf/v13/*`.

- [x] `P3-T2` Add Podman multi-node scenarios.
  - `ST1` Space creation + invite join in container network.
  - `ST2` Request-to-join moderation decision scenario.
  - `ST3` Relay no-data-hosting regression scenario for chat/join paths.
  - `ST4` Deterministic pass/fail probes and result manifest output.
  - Artifacts: `containers/v1.3/*`, `scripts/v13-e2e-podman.sh`, `docs/v1.3/phase3/p3-podman-scenarios.md`.

### Phase 4 - v14 spec package (G6)
- [x] `P4-T1` Produce voice baseline spec.
  - `ST1` Voice join/leave/signaling state model.
  - `ST2` Topology and fallback contract (direct/mesh/SFU/TURN).
  - `ST3` Call setup and recovery SLO targets.
  - Artifacts: `docs/v1.3/phase4/f14-voice-spec.md`.

- [x] `P4-T2` Produce v14 proto and test delta package.
  - Artifacts: `docs/v1.3/phase4/f14-proto-delta.md`, `docs/v1.3/phase4/f14-acceptance-matrix.md`.

### Phase 5 - Closure and evidence (G7)
- [x] `P5-T1` Publish evidence bundle and promotion decision.
  - `ST1` Command outputs for compatibility/test/e2e/perf checks.
  - `ST2` Podman scenario outputs and deterministic result manifests.
  - `ST3` `F13` as-built conformance report against v12 `F13` specs.
  - `ST4` Risk closure and promotion recommendation.
  - `ST5` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts: `docs/v1.3/phase5/p5-evidence-bundle.md`, `docs/v1.3/phase5/p5-risk-register.md`, `docs/v1.3/phase5/p5-as-built-conformance.md`, `docs/v1.3/phase5/p5-gate-signoff.md`, `docs/v1.3/phase5/p5-evidence-index.md`.

## Risk register (v13)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R13-1 | Space policy ambiguity | Deterministic join-policy state machine | All policy transition tests pass |
| R13-2 | Message state drift in UI | Canonical delivery/read state model | No-limbo journey tests pass |
| R13-3 | Invite abuse | Baseline abuse controls and rate limits | Abuse-path tests pass |
| R13-4 | Relay boundary regresses during Space/chat implementation | Dedicated regression checks in e2e and Podman scenarios | Relay-boundary regression tests pass |

## Decision log (v13)
- `D13-1`: Spaces are visible by default, invite-only by default.
- `D13-2`: Founder is auto-assigned admin rights at Space creation.
- `D13-3`: Space/chat implementation must not broaden relay durable storage permissions.
