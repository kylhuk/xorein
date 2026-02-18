# TODO v19 - Xorein Execution Plan (v1.9)

## Status
Completed in repository scope; versioned verification reran `go test -count=1 ./pkg/v19/...`, `go test -count=1 ./tests/e2e/v19/...`, and `go test -count=1 ./tests/perf/v19/...`.

## Version Isolation Contract (mandatory)
- v19 cannot advance unless all v19 in-scope spec, code, tests, docs, ops artifacts, and evidence are complete.
- v19 implements only `F19` from v18 specs.
- v19 must also publish full `F20` spec package for v20 implementation.

## Version role
- Implements: `F19` (connectivity orchestrator + QoL invariants + continuity contracts).
- Specifies: `F20` (production/public-beta hardening and release conformance contracts).

## Critical scope (v19)
- Implement Connectivity Orchestrator (CO) path ladder and failover engine.
- Implement deterministic reason taxonomy across startup, messaging, sync, and call flows.
- Implement no-limbo UX contract and recovery-first transitions.
- Implement continuity contracts (draft/read/call handoff and wake flows).
- Validate NAT/transport/mobility matrix and QoL journey scorecards.
- Preserve relay no-data-hosting boundary with explicit regression checks.

## Out of scope (defer)
- v20 hardening and release closure implementation is deferred to v20 scope.
- Non-critical parity features remain deferred to `v20+`.

## Dependencies and relationships
- Inputs from v18: `docs/v1.8/phase4/f19-connectivity-qol-spec.md`, `docs/v1.8/phase4/f19-proto-delta.md`, `docs/v1.8/phase4/f19-acceptance-matrix.md`.
- Outputs consumed by v20: `docs/v1.9/phase4/f20-release-hardening-spec.md`, `docs/v1.9/phase4/f20-acceptance-matrix.md`.

## Entry criteria (must be true before implementation starts)
- `v18` is in `promoted` state with evidence bundle and as-built conformance report.
- Required `F19` spec inputs from v18 exist and are approved.
- v19 deferred scope list is frozen and approved.

## Promotion gates (must all pass)
- `G0` Scope lock and dependency map complete.
- `G1` Proto compatibility checks pass.
- `G2` Connectivity orchestrator implementation complete.
- `G3` QoL and continuity UX/runtime integration complete.
- `G4` NAT/transport/mobility and journey test matrix complete.
- `G5` Podman network-chaos scenarios complete.
- `G6` v20 hardening spec package complete.
- `G7` Docs and evidence complete.
- `G8` Relay no-data-hosting regression checks pass.
- `G9` `F19` as-built conformance report completed against v18 specification package.

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v19/...` (or declared equivalent if paths differ)
- `go test ./tests/perf/v19/...` (or declared equivalent if paths differ)
- `make check-full`
- `scripts/v19-chaos-scenarios.sh`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v19-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## Phase plan

### Phase 0 - Scope lock (G0)
- [x] `P0-T1` Freeze v19 CO and QoL acceptance criteria.
  - `ST1` Produce requirement-to-artifact traceability matrix.
  - `ST2` Assign gate ownership and approvers using RACI template.
  - Artifacts: `docs/v1.9/phase0/p0-scope-lock.md`, `docs/v1.9/phase0/p0-traceability-matrix.md`, `docs/v1.9/phase0/p0-gate-ownership.md`.

### Phase 1 - Connectivity orchestrator runtime (G2)
- [x] `P1-T1` Implement CO path-selection engine.
  - `ST1` Path ladder: direct QUIC/TCP -> tunnel -> relay -> TURN for media.
  - `ST2` Per-modality routing decisions and fallback triggers.
  - `ST3` Deterministic state and reason-code emissions.
  - Artifacts: `pkg/v19/co/*`, `pkg/v19/co/*_test.go`.

- [x] `P1-T2` Implement tunnel/recovery orchestration hooks.
  - `ST1` Opportunistic tunnel establishment policy.
  - `ST2` Auto-heal and teardown semantics.
  - Artifacts: `pkg/v19/tunnel/*`, `pkg/v19/tunnel/*_test.go`.

### Phase 2 - QoL and continuity integration (G3)
- [x] `P2-T1` Implement no-limbo UX contract in Gio.
  - `ST1` Canonical state/reason/next-action rendering.
  - `ST2` Recovery-first call and messaging transitions.
  - Artifacts: `pkg/v19/ui/*`, `docs/v1.9/phase2/p2-nolimbo-ux-contract.md`.

- [x] `P2-T2` Implement continuity contracts.
  - `ST1` Draft persistence and read-position continuity.
  - `ST2` Call handoff and wake-from-notification behavior.
  - Artifacts: `pkg/v19/continuity/*`, `pkg/v19/continuity/*_test.go`.

### Phase 3 - Validation matrix (G4, G5)
- [x] `P3-T1` Add CO and QoL test suites.
  - `ST1` NAT matrix tests (full cone/restricted/port-restricted/symmetric).
  - `ST2` Transport and mobility tests (UDP blocked/TCP-only/switching).
  - `ST3` Journey scorecards for startup, send, call setup, call recovery, wake, resume.
  - Artifacts: `tests/e2e/v19/*`, `tests/perf/v19/*`.

- [x] `P3-T2` Add Podman chaos and recovery scenarios.
  - `ST1` CO path/failover/recovery scenarios in container network.
  - `ST2` Relay no-data-hosting regression scenario for continuity and recovery flows.
  - `ST3` Deterministic pass/fail probes and result manifest output.
  - Artifacts: `containers/v1.9/*`, `scripts/v19-chaos-scenarios.sh`, `docs/v1.9/phase3/p3-podman-scenarios.md`.

### Phase 4 - v20 spec package (G6)
- [x] `P4-T1` Produce release hardening specification package.
  - `ST1` Security hardening and vulnerability response gates.
  - `ST2` Podman operator readiness and rollback contracts.
  - `ST3` Public beta vs production go/no-go matrix.
  - Artifacts: `docs/v1.9/phase4/f20-release-hardening-spec.md`, `docs/v1.9/phase4/f20-acceptance-matrix.md`.

### Phase 5 - Closure and evidence (G7)
- [x] `P5-T1` Publish v19 evidence bundle and promotion recommendation.
  - `ST1` Command outputs for compatibility/test/e2e/perf checks.
  - `ST2` Podman scenario outputs and deterministic result manifests.
  - `ST3` `F19` as-built conformance report against v18 `F19` specs.
  - `ST4` Risk closure and promotion recommendation.
  - `ST5` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts: `docs/v1.9/phase5/p5-evidence-bundle.md`, `docs/v1.9/phase5/p5-risk-register.md`, `docs/v1.9/phase5/p5-as-built-conformance.md`, `docs/v1.9/phase5/p5-gate-signoff.md`, `docs/v1.9/phase5/p5-evidence-index.md`.

## Risk register (v19)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R19-1 | CO complexity causes regressions | Incremental state machine tests + canary scenarios | CO tests pass with no regression |
| R19-2 | QoL inconsistency across flows | One reason taxonomy and journey scorecards | All journey gates pass |
| R19-3 | Mobility edge cases under-tested | Explicit network-switch matrix | Mobility tests pass |
| R19-4 | Relay boundary regresses during CO rollout | Dedicated regression checks in e2e and Podman scenarios | Relay-boundary regression tests pass |

## Decision log (v19)
- `D19-1`: CO and no-limbo UX are release blockers for v20 readiness.
- `D19-2`: CO and continuity implementation must not broaden relay durable storage permissions.
