# TODO v17 - Xorein Execution Plan (v1.7)

## Status
Completed in repository scope; versioned verification reran `go test -count=1 ./pkg/v17/...`, `go test -count=1 ./tests/e2e/v17/...`, and `go test -count=1 ./tests/perf/v17/...`.

## Version Isolation Contract (mandatory)
- v17 cannot advance unless all v17 in-scope spec, code, tests, docs, ops artifacts, and evidence are complete.
- v17 implements only `F17` from v16 specs.
- v17 must also publish full `F18` spec package for v18 implementation.

## Version role
- Implements: `F17` (moderation events + audit and client enforcement contracts).
- Specifies: `F18` (public discovery + indexer + join funnel contracts).

## Critical scope (v17)
- Implement signed moderation events: redaction, timeout, ban, slow mode, lockdown.
- Implement append-only moderation audit log visibility.
- Implement official-client moderation enforcement state signaling.
- Validate moderation behavior under partitions/rejoins and conflicting actions.
- Keep all moderation logic compatible with decentralized client enforcement model.
- Preserve relay no-data-hosting boundary with explicit regression checks.

## Out of scope (defer)
- Discovery/indexer implementation is deferred to v18 implementation scope.
- Non-critical parity features remain deferred to `v20+`.

## Dependencies and relationships
- Inputs from v16: `docs/v1.6/phase4/f17-moderation-spec.md`, `docs/v1.6/phase4/f17-proto-delta.md`, `docs/v1.6/phase4/f17-acceptance-matrix.md`.
- Outputs consumed by v18: `docs/v1.7/phase4/f18-discovery-spec.md`, `docs/v1.7/phase4/f18-proto-delta.md`, `docs/v1.7/phase4/f18-acceptance-matrix.md`.

## Entry criteria (must be true before implementation starts)
- `v16` is in `promoted` state with evidence bundle and as-built conformance report.
- Required `F17` spec inputs from v16 exist and are approved.
- v17 deferred scope list is frozen and approved.

## Promotion gates (must all pass)
- `G0` Scope lock and dependency map complete.
- `G1` Proto compatibility checks pass.
- `G2` Moderation event runtime complete.
- `G3` Audit log and enforcement signaling complete.
- `G4` Moderation/adversarial test matrix complete.
- `G5` Podman moderation scenarios complete.
- `G6` v18 spec package complete.
- `G7` Docs and evidence complete.
- `G8` Relay no-data-hosting regression checks pass.
- `G9` `F17` as-built conformance report completed against v16 specification package.

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v17/...` (or declared equivalent if paths differ)
- `go test ./tests/perf/v17/...` (or declared equivalent if paths differ)
- `make check-full`
- `scripts/v17-moderation-scenarios.sh`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v17-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## Phase plan

### Phase 0 - Scope lock (G0)
- [x] `P0-T1` Freeze v17 moderation and audit acceptance criteria.
  - `ST1` Produce requirement-to-artifact traceability matrix.
  - `ST2` Assign gate ownership and approvers using RACI template.
  - Artifacts: `docs/v1.7/phase0/p0-scope-lock.md`, `docs/v1.7/phase0/p0-traceability-matrix.md`, `docs/v1.7/phase0/p0-gate-ownership.md`.

### Phase 1 - Moderation event runtime (G2)
- [x] `P1-T1` Implement signed moderation event pipeline.
  - `ST1` Event creation, signing, verification, and ordering.
  - `ST2` Event application rules for redaction, timeout, ban, slow mode, lockdown.
  - `ST3` Deterministic rejection reasons for invalid events.
  - Artifacts: `pkg/v17/moderation/*`, `pkg/v17/moderation/*_test.go`.

- [x] `P1-T2` Implement moderation replication semantics.
  - `ST1` Consistency rules across replicas/peers.
  - `ST2` Replay and duplicate event handling.
  - Artifacts: `pkg/v17/modsync/*`, `pkg/v17/modsync/*_test.go`.

### Phase 2 - Audit and client enforcement (G3)
- [x] `P2-T1` Implement append-only audit log model.
  - `ST1` Query constraints by role.
  - `ST2` Signer and verification metadata visibility.
  - Artifacts: `pkg/v17/audit/*`, `pkg/v17/audit/*_test.go`.

- [x] `P2-T2` Implement official-client enforcement status signaling.
  - `ST1` Surface enforcement mode and trust warnings.
  - `ST2` Ensure UI and runtime semantics remain aligned.
  - Artifacts: `pkg/v17/ui/*`, `docs/v1.7/phase2/p2-enforcement-ux-contract.md`.

### Phase 3 - Validation matrix (G4, G5)
- [x] `P3-T1` Add moderation adversarial tests.
  - `ST1` Forged signature and stale-event tests.
  - `ST2` Concurrent conflicting moderation action tests.
  - `ST3` Partition/rejoin convergence tests.
  - Artifacts: `tests/e2e/v17/*`, `tests/perf/v17/*`.

- [x] `P3-T2` Add Podman moderation scenarios.
  - `ST1` Moderation event and audit scenarios in container network.
  - `ST2` Relay no-data-hosting regression scenario for moderation paths.
  - `ST3` Deterministic pass/fail probes and result manifest output.
  - Artifacts: `containers/v1.7/*`, `scripts/v17-moderation-scenarios.sh`, `docs/v1.7/phase3/p3-podman-scenarios.md`.

### Phase 4 - v18 spec package (G6)
- [x] `P4-T1` Produce discovery and indexer specification package.
  - `ST1` Signed `DirectoryEntry` contract and publication model.
  - `ST2` Indexer response signature and client verification model.
  - `ST3` Join funnel contract (invite/request/open + abuse controls).
  - Artifacts: `docs/v1.7/phase4/f18-discovery-spec.md`, `docs/v1.7/phase4/f18-proto-delta.md`, `docs/v1.7/phase4/f18-acceptance-matrix.md`.

### Phase 5 - Closure and evidence (G7)
- [x] `P5-T1` Publish v17 evidence bundle and promotion recommendation.
  - `ST1` Command outputs for compatibility/test/e2e/perf checks.
  - `ST2` Podman scenario outputs and deterministic result manifests.
  - `ST3` `F17` as-built conformance report against v16 `F17` specs.
  - `ST4` Risk closure and promotion recommendation.
  - `ST5` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts: `docs/v1.7/phase5/p5-evidence-bundle.md`, `docs/v1.7/phase5/p5-risk-register.md`, `docs/v1.7/phase5/p5-as-built-conformance.md`, `docs/v1.7/phase5/p5-gate-signoff.md`, `docs/v1.7/phase5/p5-evidence-index.md`.

## Risk register (v17)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R17-1 | Inconsistent moderation enforcement | Signed deterministic event model | Convergence tests pass |
| R17-2 | Audit tampering concerns | Append-only verified log model | Integrity tests pass |
| R17-3 | Client non-compliance ambiguity | Explicit enforcement-status signaling | UX and runtime tests pass |
| R17-4 | Relay boundary regresses during moderation rollout | Dedicated regression checks in e2e and Podman scenarios | Relay-boundary regression tests pass |

## Decision log (v17)
- `D17-1`: Moderation in decentralized mode is authoritative via signed events and official-client enforcement.
- `D17-2`: Moderation implementation must not broaden relay durable storage permissions.
