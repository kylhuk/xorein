# TODO v14 - Xorein Execution Plan (v1.4)

## Status
Planning artifact only. This file defines v14 implementation and validation requirements. It does not claim implementation completion.

## Version Isolation Contract (mandatory)
- v14 cannot advance unless all v14 in-scope spec, code, tests, docs, ops artifacts, and evidence are complete.
- v14 implements only `F14` from v13 specs.
- v14 must also publish full `F15` spec package for v15 implementation.

## Version role
- Implements: `F14` (voice channels baseline and recovery contracts).
- Specifies: `F15` (screen share baseline and adaptation contracts).

## Critical scope (v14)
- Implement voice channel signaling runtime and session lifecycle.
- Implement media baseline (Opus profile) and deterministic fallback behavior.
- Implement direct/mesh/SFU/TURN path selection policy for voice.
- Implement Gio voice controls and no-limbo call state UX.
- Validate call setup, reconnect, and degraded-path behavior.
- Preserve relay no-data-hosting boundary with explicit regression checks.

## Out of scope (defer)
- Screen share implementation is deferred to v15 implementation scope.
- Non-critical parity features remain deferred to `v20+`.

## Dependencies and relationships
- Inputs from v13: `docs/v1.3/phase4/f14-voice-spec.md`, `docs/v1.3/phase4/f14-proto-delta.md`, `docs/v1.3/phase4/f14-acceptance-matrix.md`.
- Outputs consumed by v15: `docs/v1.4/phase4/f15-screenshare-spec.md`, `docs/v1.4/phase4/f15-proto-delta.md`, `docs/v1.4/phase4/f15-acceptance-matrix.md`.

## Entry criteria (must be true before implementation starts)
- `v13` is in `promoted` state with evidence bundle and as-built conformance report.
- Required `F14` spec inputs from v13 exist and are approved.
- v14 deferred scope list is frozen and approved.

## Promotion gates (must all pass)
- `G0` Scope lock and artifact mapping complete.
- `G1` Additive wire compatibility checks pass.
- `G2` Voice signaling/runtime implementation complete.
- `G3` Gio voice user journeys complete.
- `G4` Voice reliability/performance test matrix complete.
- `G5` Podman call scenarios complete.
- `G6` v15 spec package complete.
- `G7` Docs and evidence complete.
- `G8` Relay no-data-hosting regression checks pass.
- `G9` `F14` as-built conformance report completed against v13 specification package.

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v14/...` (or declared equivalent if paths differ)
- `go test ./tests/perf/v14/...` (or declared equivalent if paths differ)
- `make check-full`
- `scripts/v14-voice-scenarios.sh`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v14-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## Phase plan

### Phase 0 - Scope lock (G0)
- [ ] `P0-T1` Freeze v14 requirements and SLOs.
  - `ST1` Produce requirement-to-artifact traceability matrix.
  - `ST2` Assign gate ownership and approvers using RACI template.
  - Artifacts: `docs/v1.4/phase0/p0-scope-lock.md`, `docs/v1.4/phase0/p0-traceability-matrix.md`, `docs/v1.4/phase0/p0-gate-ownership.md`.

### Phase 1 - Voice signaling and session engine (G2)
- [ ] `P1-T1` Implement voice signaling lifecycle.
  - `ST1` Session create/join/leave and retries.
  - `ST2` Deterministic signaling error taxonomy.
  - Artifacts: `pkg/v14/signaling/*`, `pkg/v14/signaling/*_test.go`.

- [ ] `P1-T2` Implement voice session engine.
  - `ST1` Codec and transport profile negotiation.
  - `ST2` Reconnect/backoff policy and state transitions.
  - `ST3` Fallback orchestration (direct -> mesh -> SFU/TURN).
  - Artifacts: `pkg/v14/voice/*`, `pkg/v14/voice/*_test.go`.

### Phase 2 - Client integration and UX (G3)
- [ ] `P2-T1` Implement Gio voice controls.
  - `ST1` Join/leave/mute/deafen/device selection.
  - `ST2` Call state and quality badge rendering.
  - `ST3` Recovery-first banners for transient failures.
  - Artifacts: `pkg/v14/ui/*`, `docs/v1.4/phase2/p2-voice-ux-contract.md`.

### Phase 3 - Validation matrix (G4, G5)
- [ ] `P3-T1` Add reliability and degraded-network tests.
  - `ST1` Call setup p50/p95 tests.
  - `ST2` Network-switch recovery tests.
  - `ST3` Restrictive-network force-SFU tests.
  - Artifacts: `tests/e2e/v14/*`, `tests/perf/v14/*`.

- [ ] `P3-T2` Add Podman multi-peer call scenarios.
  - `ST1` Voice call establishment and reconnect scenarios in container network.
  - `ST2` Relay no-data-hosting regression scenario for voice control paths.
  - `ST3` Deterministic pass/fail probes and result manifest output.
  - Artifacts: `containers/v1.4/*`, `scripts/v14-voice-scenarios.sh`, `docs/v1.4/phase3/p3-podman-scenarios.md`.

### Phase 4 - v15 spec package (G6)
- [ ] `P4-T1` Produce screen share specification package.
  - `ST1` Capture source, preset, and adaptation policy.
  - `ST2` Viewer-side degradation and recovery contract.
  - `ST3` Security and keying requirements for media frames.
  - Artifacts: `docs/v1.4/phase4/f15-screenshare-spec.md`, `docs/v1.4/phase4/f15-proto-delta.md`, `docs/v1.4/phase4/f15-acceptance-matrix.md`.

### Phase 5 - Closure and evidence (G7)
- [ ] `P5-T1` Publish v14 evidence bundle and promotion recommendation.
  - `ST1` Command outputs for compatibility/test/e2e/perf checks.
  - `ST2` Podman scenario outputs and deterministic result manifests.
  - `ST3` `F14` as-built conformance report against v13 `F14` specs.
  - `ST4` Risk closure and promotion recommendation.
  - `ST5` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts: `docs/v1.4/phase5/p5-evidence-bundle.md`, `docs/v1.4/phase5/p5-risk-register.md`, `docs/v1.4/phase5/p5-as-built-conformance.md`, `docs/v1.4/phase5/p5-gate-signoff.md`, `docs/v1.4/phase5/p5-evidence-index.md`.

## Risk register (v14)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R14-1 | Call instability across NAT types | Deterministic fallback ladder and tests | NAT matrix tests pass |
| R14-2 | UX limbo during reconnect | Recovery-first state contract | Journey tests pass |
| R14-3 | Voice quality regressions | SLO gating and perf regression checks | SLO gates pass |
| R14-4 | Relay boundary regresses during voice rollout | Dedicated regression checks in e2e and Podman scenarios | Relay-boundary regression tests pass |

## Decision log (v14)
- `D14-1`: Voice must ship with deterministic reconnect behavior, not best-effort only.
- `D14-2`: Voice implementation must not broaden relay durable storage permissions.
