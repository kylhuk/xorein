# TODO v18 - Xorein Execution Plan (v1.8)

## Status
Completed in repository scope; versioned verification reran `go test -count=1 ./pkg/v18/...`, `go test -count=1 ./tests/e2e/v18/...`, and `go test -count=1 ./tests/perf/v18/...`.

## Version Isolation Contract (mandatory)
- v18 cannot advance unless all v18 in-scope spec, code, tests, docs, ops artifacts, and evidence are complete.
- v18 implements only `F18` from v17 specs.
- v18 must also publish full `F19` spec package for v19 implementation.

## Version role
- Implements: `F18` (public discovery + indexer + join funnel contracts).
- Specifies: `F19` (connectivity orchestrator + QoL invariants + continuity contracts).

## Critical scope (v18)
- Implement signed public Space listing model (`DirectoryEntry`) and publication paths.
- Implement reference indexer service and signed search responses.
- Implement client-side verification, deduplication, and multi-indexer merge logic.
- Implement Explore and join funnel UX with invite/request/open modes.
- Validate discovery correctness, trust model, and abuse controls.
- Preserve relay no-data-hosting boundary with explicit regression checks.

## Out of scope (defer)
- Connectivity orchestrator and QoL continuity implementation is deferred to v19 scope.
- Non-critical parity features remain deferred to `v20+`.

## Dependencies and relationships
- Inputs from v17: `docs/v1.7/phase4/f18-discovery-spec.md`, `docs/v1.7/phase4/f18-proto-delta.md`, `docs/v1.7/phase4/f18-acceptance-matrix.md`.
- Outputs consumed by v19: `docs/v1.8/phase4/f19-connectivity-qol-spec.md`, `docs/v1.8/phase4/f19-proto-delta.md`, `docs/v1.8/phase4/f19-acceptance-matrix.md`.

## Entry criteria (must be true before implementation starts)
- `v17` is in `promoted` state with evidence bundle and as-built conformance report.
- Required `F18` spec inputs from v17 exist and are approved.
- v18 deferred scope list is frozen and approved.

## Promotion gates (must all pass)
- `G0` Scope lock and dependency map complete.
- `G1` Proto compatibility checks pass.
- `G2` DirectoryEntry and indexer runtime complete.
- `G3` Client discovery verification and join UX complete.
- `G4` Discovery/adversarial test matrix complete.
- `G5` Podman discovery scenarios complete.
- `G6` v19 spec package complete.
- `G7` Docs and evidence complete.
- `G8` Relay no-data-hosting regression checks pass.
- `G9` `F18` as-built conformance report completed against v17 specification package.

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v18/...` (or declared equivalent if paths differ)
- `go test ./tests/perf/v18/...` (or declared equivalent if paths differ)
- `make check-full`
- `scripts/v18-discovery-scenarios.sh`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v18-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## Phase plan

### Phase 0 - Scope lock (G0)
- [x] `P0-T1` Freeze discovery and indexer scope.
  - `ST1` Produce requirement-to-artifact traceability matrix.
  - `ST2` Assign gate ownership and approvers using RACI template.
  - Artifacts: `docs/v1.8/phase0/p0-scope-lock.md`, `docs/v1.8/phase0/p0-traceability-matrix.md`, `docs/v1.8/phase0/p0-gate-ownership.md`.

### Phase 1 - Directory and indexer runtime (G2)
- [x] `P1-T1` Implement `DirectoryEntry` publication lifecycle.
  - `ST1` Sign, publish, update, and revoke listing behavior.
  - `ST2` Deterministic keying and retrieval semantics.
  - Artifacts: `pkg/v18/directory/*`, `pkg/v18/directory/*_test.go`.

- [x] `P1-T2` Implement reference indexer service.
  - `ST1` Crawl and verify public listings.
  - `ST2` Build searchable index and signed response payloads.
  - `ST3` Add operator configuration and health endpoints.
  - Artifacts: `cmd/indexer/*`, `pkg/v18/indexer/*`, `pkg/v18/indexer/*_test.go`.

### Phase 2 - Client verification and join UX (G3)
- [x] `P2-T1` Implement multi-indexer query and verification.
  - `ST1` Merge/deduplicate by Space ID.
  - `ST2` Verify signatures and surface trust warnings.
  - Artifacts: `pkg/v18/discoveryclient/*`, `pkg/v18/discoveryclient/*_test.go`.

- [x] `P2-T2` Implement Explore and join funnel UX.
  - `ST1` Browse/search/preview flows.
  - `ST2` Invite/request/open join actions and error states.
  - Artifacts: `pkg/v18/ui/*`, `docs/v1.8/phase2/p2-discovery-ux-contract.md`.

### Phase 3 - Validation matrix (G4, G5)
- [x] `P3-T1` Add discovery integrity and abuse-path tests.
  - `ST1` Forged listing and forged indexer response tests.
  - `ST2` Duplicate/mismatch merge tests.
  - `ST3` Join abuse control tests (rate/reputation/PoW where specified).
  - Artifacts: `tests/e2e/v18/*`, `tests/perf/v18/*`.

- [x] `P3-T2` Add Podman discovery/indexer scenarios.
  - `ST1` Discovery/indexer/search/join scenarios in container network.
  - `ST2` Relay no-data-hosting regression scenario for discovery/join paths.
  - `ST3` Deterministic pass/fail probes and result manifest output.
  - Artifacts: `containers/v1.8/*`, `scripts/v18-discovery-scenarios.sh`, `docs/v1.8/phase3/p3-podman-scenarios.md`.

### Phase 4 - v19 spec package (G6)
- [x] `P4-T1` Produce connectivity orchestrator and QoL specification package.
  - `ST1` Path ladder and deterministic reason taxonomy.
  - `ST2` Global no-limbo UX invariants and journey contracts.
  - `ST3` Continuity contracts (draft/read/call handoff, wake behavior).
  - Artifacts: `docs/v1.8/phase4/f19-connectivity-qol-spec.md`, `docs/v1.8/phase4/f19-proto-delta.md`, `docs/v1.8/phase4/f19-acceptance-matrix.md`.

### Phase 5 - Closure and evidence (G7)
- [x] `P5-T1` Publish v18 evidence bundle and promotion recommendation.
  - `ST1` Command outputs for compatibility/test/e2e/perf checks.
  - `ST2` Podman scenario outputs and deterministic result manifests.
  - `ST3` `F18` as-built conformance report against v17 `F18` specs.
  - `ST4` Risk closure and promotion recommendation.
  - `ST5` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts: `docs/v1.8/phase5/p5-evidence-bundle.md`, `docs/v1.8/phase5/p5-risk-register.md`, `docs/v1.8/phase5/p5-as-built-conformance.md`, `docs/v1.8/phase5/p5-gate-signoff.md`, `docs/v1.8/phase5/p5-evidence-index.md`.

## Risk register (v18)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R18-1 | Indexer trust abuse | Signed result verification and multi-source merge | Adversarial tests pass |
| R18-2 | Discovery privacy leakage | Query minimization and source controls | Privacy checks pass |
| R18-3 | Join funnel abuse | Deterministic abuse controls and tests | Abuse-path tests pass |
| R18-4 | Relay boundary regresses during discovery rollout | Dedicated regression checks in e2e and Podman scenarios | Relay-boundary regression tests pass |

## Decision log (v18)
- `D18-1`: Indexers are optional and non-authoritative; signatures remain the trust anchor.
- `D18-2`: Discovery implementation must not broaden relay durable storage permissions.
