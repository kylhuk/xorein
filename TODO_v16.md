# TODO v16 - Xorein Execution Plan (v1.6)

## Status
Planning artifact only. This file defines v16 implementation and validation requirements. It does not claim implementation completion.

## Version Isolation Contract (mandatory)
- v16 cannot advance unless all v16 in-scope spec, code, tests, docs, ops artifacts, and evidence are complete.
- v16 implements only `F16` from v15 specs.
- v16 must also publish full `F17` spec package for v17 implementation.

## Version role
- Implements: `F16` (full RBAC + ACL model and enforcement contracts).
- Specifies: `F17` (moderation events + audit and client enforcement contracts).

## Critical scope (v16)
- Implement default and custom role management.
- Implement ACL allow/deny logic with deterministic merge order.
- Enforce permissions consistently for text channels, voice channels, and screen share controls.
- Implement Gio admin UX for role and permission management.
- Validate complete permission matrix and conflict-path behavior.
- Preserve relay no-data-hosting boundary with explicit regression checks.

## Out of scope (defer)
- Moderation event implementation is deferred to v17 implementation scope.
- Non-critical parity features remain deferred to `v20+`.

## Dependencies and relationships
- Inputs from v15: `docs/v1.5/phase4/f16-rbac-acl-spec.md`, `docs/v1.5/phase4/f16-proto-delta.md`, `docs/v1.5/phase4/f16-acceptance-matrix.md`.
- Outputs consumed by v17: `docs/v1.6/phase4/f17-moderation-spec.md`, `docs/v1.6/phase4/f17-proto-delta.md`, `docs/v1.6/phase4/f17-acceptance-matrix.md`.

## Entry criteria (must be true before implementation starts)
- `v15` is in `promoted` state with evidence bundle and as-built conformance report.
- Required `F16` spec inputs from v15 exist and are approved.
- v16 deferred scope list is frozen and approved.

## Promotion gates (must all pass)
- `G0` Scope lock and dependency map complete.
- `G1` Proto compatibility checks pass.
- `G2` RBAC and ACL runtime complete.
- `G3` Permission enforcement in client/runtime complete.
- `G4` Permission matrix and adversarial tests complete.
- `G5` Podman policy scenarios complete.
- `G6` v17 spec package complete.
- `G7` Docs and evidence complete.
- `G8` Relay no-data-hosting regression checks pass.
- `G9` `F16` as-built conformance report completed against v15 specification package.

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v16/...` (or declared equivalent if paths differ)
- `go test ./tests/perf/v16/...` (or declared equivalent if paths differ)
- `make check-full`
- `scripts/v16-rbac-scenarios.sh`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v16-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## Phase plan

### Phase 0 - Scope lock (G0)
- [x] `P0-T1` Freeze v16 role/ACL scope and acceptance criteria.
  - `ST1` Produce requirement-to-artifact traceability matrix.
  - `ST2` Assign gate ownership and approvers using RACI template.
  - Artifacts: `docs/v1.6/phase0/p0-scope-lock.md`, `docs/v1.6/phase0/p0-traceability-matrix.md`, `docs/v1.6/phase0/p0-gate-ownership.md`.

### Phase 1 - RBAC and ACL runtime (G2)
- [x] `P1-T1` Implement role model and assignment lifecycle.
  - `ST1` Seed default roles and permissions.
  - `ST2` Custom role create/update/delete lifecycle.
  - `ST3` Founder and admin safety constraints.
  - Artifacts: `pkg/v16/rbac/*`, `pkg/v16/rbac/*_test.go`.

- [x] `P1-T2` Implement ACL and merge engine.
  - `ST1` Channel-level allow/deny overrides.
  - `ST2` Deterministic precedence and conflict resolution.
  - `ST3` Explainability output for denied actions.
  - Artifacts: `pkg/v16/acl/*`, `pkg/v16/acl/*_test.go`.

### Phase 2 - Client and runtime enforcement (G3)
- [x] `P2-T1` Wire permission checks into chat/voice/screenshare actions.
  - `ST1` Block unauthorized sends and channel joins.
  - `ST2` Block unauthorized moderation and management actions.
  - `ST3` Ensure no bypass paths in relay/client flows.
  - Artifacts: `pkg/v16/enforcement/*`, `tests/e2e/v16/enforcement_test.go`.

- [x] `P2-T2` Implement Gio admin/role management UX.
  - Artifacts: `pkg/v16/ui/*`, `docs/v1.6/phase2/p2-rbac-ux-contract.md`.

### Phase 3 - Validation matrix (G4, G5)
- [x] `P3-T1` Add permission matrix tests.
  - `ST1` Positive path tests per role.
  - `ST2` Negative path and escalation-attempt tests.
  - `ST3` Partition/rejoin consistency checks.
  - Artifacts: `tests/e2e/v16/*`, `tests/perf/v16/*`.

- [x] `P3-T2` Add Podman policy scenarios.
  - `ST1` Role/ACL policy scenarios in container network.
  - `ST2` Relay no-data-hosting regression scenario for policy-enforcement paths.
  - `ST3` Deterministic pass/fail probes and result manifest output.
  - Artifacts: `containers/v1.6/*`, `scripts/v16-rbac-scenarios.sh`, `docs/v1.6/phase3/p3-podman-scenarios.md`.

### Phase 4 - v17 spec package (G6)
- [x] `P4-T1` Produce moderation + audit specification package.
  - `ST1` Redaction, timeout, ban, slow mode, lockdown contracts.
  - `ST2` Signed-event requirements and audit visibility rules.
  - `ST3` Official-client enforcement status signaling.
  - Artifacts: `docs/v1.6/phase4/f17-moderation-spec.md`, `docs/v1.6/phase4/f17-proto-delta.md`, `docs/v1.6/phase4/f17-acceptance-matrix.md`.

### Phase 5 - Closure and evidence (G7)
- [x] `P5-T1` Publish v16 evidence bundle and promotion recommendation.
  - `ST1` Command outputs for compatibility/test/e2e/perf checks.
  - `ST2` Podman scenario outputs and deterministic result manifests.
  - `ST3` `F16` as-built conformance report against v15 `F16` specs.
  - `ST4` Risk closure and promotion recommendation.
  - `ST5` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts: `docs/v1.6/phase5/p5-evidence-bundle.md`, `docs/v1.6/phase5/p5-risk-register.md`, `docs/v1.6/phase5/p5-as-built-conformance.md`, `docs/v1.6/phase5/p5-gate-signoff.md`, `docs/v1.6/phase5/p5-evidence-index.md`.

## Risk register (v16)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R16-1 | Permission conflict ambiguity | Deterministic precedence algorithm | Conflict tests pass |
| R16-2 | Hidden bypass paths | End-to-end enforcement tests | No bypass found in test suite |
| R16-3 | Admin UX complexity | Explainable deny reasons and constrained workflows | UX acceptance checks pass |
| R16-4 | Relay boundary regresses during RBAC/ACL rollout | Dedicated regression checks in e2e and Podman scenarios | Relay-boundary regression tests pass |

## Decision log (v16)
- `D16-1`: Permission enforcement is protocol/runtime authoritative, not UI best-effort.
- `D16-2`: RBAC/ACL implementation must not broaden relay durable storage permissions.
