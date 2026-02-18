# TODO v22 - Xorein (backend) + harmolyn (frontend) Execution Plan (v2.2)

## Status
Planning artifact only. This file defines v22 implementation and validation requirements. It does not claim implementation completion.

## Naming + binary split note (carry-forward)
- **Xorein**: backend node/runtime binary (`cmd/xorein`).
- **harmolyn**: frontend UI binary (`cmd/harmolyn`) consuming Xorein runtime APIs (in-process library by default).
- v22 focuses on durable history-plane behavior implemented in Xorein; harmolyn provides UX surfaces only.

## Version Isolation Contract (mandatory)
- v22 cannot advance unless all v22 in-scope spec, code, tests, docs, ops artifacts, and evidence are complete.
- v22 implements only `F22` from v21 specs (additive proto only).
- v22 must also publish full `F23` spec package for v23 implementation.

## Version role
- Implements: `F22` (distributed history/backfill + archivist capability + E2EE-safe backfill and coverage semantics).
- Specifies: `F23` (history/search hardening, operator readiness, and privacy/abuse conformance gates).

## Critical scope (v22)
- Implement Archivist capability (non-privileged, opt-in) for **ciphertext-only** history segment storage (Xorein):
  - capability advertisement + selection inputs
  - quotas, retention windows, pruning, and deterministic refusal reasons
  - **replication policy**: publish segments to `r` archivists (default + configurable), track replica health, and heal under churn
- Implement distributed history retrieval (Xorein):
  - `HistoryHead` and `HistorySegmentManifest` generation and verification
  - ciphertext segment upload/download endpoints (no plaintext exposure)
  - private Space anti-enumeration keying
- Implement client history backfill (Xorein client runtime):
  - time-range backfill requests by default (no keyword leakage)
  - progress, retry, and deterministic failure reasons
  - apply redaction/tombstones deterministically over backfilled content
- Extend harmolyn search UX and coverage model:
  - show coverage gaps and offer backfill-by-time-range
  - preserve no-limbo invariant for “search with missing history”
- Preserve relay no-long-history-hosting boundary with explicit regression checks.

## Out of scope (defer)
- Remote keyword search or encrypted keyword search schemes are out of scope by default.
- Incentive/token economics for storage providers are out of scope.
- Any change requiring breaking wire compatibility is out of scope for v22.

## Dependencies and relationships
- Inputs from v21:
  - `docs/v2.1/phase4/f22-history-backfill-spec.md`
  - `docs/v2.1/phase4/f22-proto-delta.md`
  - `docs/v2.1/phase4/f22-acceptance-matrix.md`
- Outputs consumed by v23:
  - `docs/v2.2/phase4/f23-history-hardening-spec.md`
  - `docs/v2.2/phase4/f23-proto-delta.md`
  - `docs/v2.2/phase4/f23-acceptance-matrix.md`

## Entry criteria (must be true before implementation starts)
- `v21` is in `promoted` state with evidence bundle and as-built conformance report.
- Required `F22` spec inputs from v21 exist and are approved.
- v22 deferred scope list is frozen and approved.

## Promotion gates (must all pass)
- `G0` Scope lock and dependency/role map complete.
- `G1` Additive wire compatibility checks pass (no breaking changes).
- `G2` Archivist runtime complete (storage, quota, retention, pruning, refusal reasons, replication).
- `G3` History integrity + retrieval protocol implementation complete.
- `G4` Client backfill + UX integration complete (coverage labeling + no-limbo).
- `G5` Adversarial/privacy/abuse test matrix complete.
- `G6` Podman history/backfill scenarios complete.
- `G7` v23 hardening spec package complete.
- `G8` Docs and evidence complete.
- `G9` Relay no-long-history-hosting regression checks pass.
- `G10` `F22` as-built conformance report completed against v21 `F22` specs.

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v22/...` (or declared equivalent if paths differ)
- `go test ./tests/perf/v22/...` (or declared equivalent if paths differ)
- `go build ./cmd/xorein`
- `go build ./cmd/harmolyn`
- `make check-full`
- `scripts/v22-history-scenarios.sh`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v22-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## LLM agent execution contract (mandatory)
- `P0-T1`: Dependency rule—start only once the entry criteria (v21 promoted, `F22` seeds approved, deferred list frozen) are met; Acceptance rule—complete ST1–ST4 and gate ownership artifacts, refresh scope docs, and mark `G0` Pass; Evidence rule—attach at least one `EV-v22-G0-###` entry plus command outputs or a `not applicable` note; Blocker state taxonomy—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint rule—before running documentation or verification commands (e.g., `make check-full`), record the planned command and capture the observed output.
- `P1-T1`: Dependency rule—after `G0` and `G1` additive checks; Acceptance rule—finish ST1–ST3 of Archivist advertisement, selection, and refusal reasons with artifacts and mark `G2` Pass; Evidence rule—≥1 `EV-v22-G2-###` entry plus command outputs or `not applicable`; Blocker taxonomy—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the upcoming command (e.g., `go test ./pkg/v22/archivist/advertise/...`) and log its output.
- `P1-T2`: Dependency rule—after `P1-T1` completion; Acceptance rule—complete ST1–ST3 (store model, quotas, retention) with documentation and mark `G2` Pass; Evidence rule—attach `EV-v22-G2-###` plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record planned command (e.g., `go test ./pkg/v22/archivist/store/...`) before running and capture output.
- `P1-T3`: Dependency rule—after `P1-T2`; Acceptance rule—complete ST1–ST4 (replication policy, pipeline, healing, failure reasons), refresh artifacts, and mark `G2` Pass; Evidence rule—≥1 `EV-v22-G2-###` entry plus command outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—log the next command (e.g., `go test ./pkg/v22/archivist/replicate/...`) before running and capture output.
- `P2-T1`: Dependency rule—after `G2` and Phase 1 artifacts; Acceptance rule—complete ST1–ST3 (history head, manifest, verification errors) and mark `G3` Pass; Evidence rule—attach `EV-v22-G3-###` plus command outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—capture planned command (e.g., `go test ./pkg/v22/history/integrity/...`) and output.
- `P2-T2`: Dependency rule—after `P2-T1`; Acceptance rule—complete ST1–ST3 (retrieval endpoints, anti-enumeration, rate limits) and mark `G3` Pass; Evidence rule—≥1 `EV-v22-G3-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record command (e.g., `go test ./pkg/v22/history/retrieve/...`) before running and capture output.
- `P2-T3`: Dependency rule—after `P2-T2`; Acceptance rule—complete ST1–ST3 (redaction apply, disclosures, search removal) and mark `G3` Pass; Evidence rule—attach one `EV-v22-G3-###` entry plus outputs or `not applicable`; Blocker taxonomy—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note planned command (e.g., `tests/e2e/v22/redaction_backfill_*`) before running and capture output.
- `P3-T1`: Dependency rule—after `G3`; Acceptance rule—complete ST1–ST3 (backfill requests, local apply, progress/retries) and mark `G4` Pass; Evidence rule—≥1 `EV-v22-G4-###` entry plus command outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record command (e.g., `go test ./pkg/v22/history/backfill/...`) and log output.
- `P3-T2`: Dependency rule—after `P3-T1`; Acceptance rule—complete ST1–ST3 for harmolyn search UX and mark `G4` Pass; Evidence rule—attach `EV-v22-G4-###` plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the planned command (e.g., `go test ./cmd/harmolyn/...`) before running and record output.
- `P4-T1`: Dependency rule—after `G4` and Phase 3 artifacts; Acceptance rule—complete ST1–ST5 of adversarial/privacy/abuse tests and mark `G5` Pass; Evidence rule—≥1 `EV-v22-G5-###` entry plus command outputs or `not applicable`; Blocker taxonomy—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record commands (e.g., `go test ./tests/e2e/v22/*`, `tests/perf/v22/*`) and capture outputs.
- `P4-T2`: Dependency rule—after `P4-T1` and `G5` completion; Acceptance rule—complete ST1–ST6 Podman scenarios and mark `G6` Pass; Evidence rule—attach `EV-v22-G6-###` entry plus command outputs (`scripts/v22-history-scenarios.sh`) or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—log the next command (`scripts/v22-history-scenarios.sh`) before execution and note its output.
- `P5-T1`: Dependency rule—after Phases 0–4 and `G6`; Acceptance rule—complete ST1–ST3 for `F23` spec package and mark `G7` Pass; Evidence rule—≥1 `EV-v22-G7-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the planned command (e.g., spec publishing workflow) before running and capture output.
- `P5-T2`: Dependency rule—after `P5-T1` and `G7`; Acceptance rule—complete ST1–ST5 for command outputs, Podman manifests, conformance report, and sign-offs (G8/G9); Evidence rule—attach `EV-v22-G8-###` and `EV-v22-G9-###` entries plus command outputs (`go build`, `go test`, scripts) or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the next command (e.g., `go build ./cmd/xorein`, `go build ./cmd/harmolyn`, `make check-full`) before running and log output.

### Phase dependency and evidence map
| Phase | Top-level tasks | Must start after | Gate target(s) | Minimum evidence | Command hints |
| --- | --- | --- | --- | --- | --- |
| Phase 0 | `P0-T1` | Entry criteria (v21 promoted + `F22` seeds approved) | `G0` | `EV-v22-G0-###` | Plan scope-lock commands (e.g., `make check-full`) before running and log output. |
| Phase 1 | `P1-T1`, `P1-T2`, `P1-T3` | Completion of Phase 0 + `G1` | `G1`, `G2` | `EV-v22-G1-###`, `EV-v22-G2-###` | Record Archivist/runtime commands (`go test ./pkg/v22/archivist/...`, `buf lint`) and capture outputs. |
| Phase 2 | `P2-T1`, `P2-T2`, `P2-T3` | `G2` completion and Phase 1 artifacts | `G3` | `EV-v22-G3-###` | Note integrity/retrieval commands (`go test ./pkg/v22/history/...`) before running and capture outputs. |
| Phase 3 | `P3-T1`, `P3-T2` | `G3` completion | `G4` | `EV-v22-G4-###` | Capture command outputs for backfill (`go test ./pkg/v22/history/backfill/...`) and harmolyn search flows. |
| Phase 4 | `P4-T1`, `P4-T2` | `G4` completion | `G5`, `G6` | `EV-v22-G5-###`, `EV-v22-G6-###` | Run adversarial/regression commands (`go test ./tests/e2e/v22/*`, `tests/perf/v22/*`, `scripts/v22-history-scenarios.sh`) and log outputs. |
| Phase 5 | `P5-T1`, `P5-T2` | Phase 4 gates + Podman scenarios | `G7`, `G8`, `G9` | `EV-v22-G7-###`, `EV-v22-G8-###`, `EV-v22-G9-###` | Document spec/evidence commands (`go build`, `make check-full`, `scripts/v22-history-scenarios.sh`) before running and record outputs. |

## Phase plan

### Phase 0 - Scope lock and architecture map (G0)
- [ ] `P0-T1` Freeze v22 implementation scope and role boundaries.
  - `ST1` Confirm Archivist is a capability flag/config (no privileged protocol authority).
  - `ST2` Produce requirement-to-artifact traceability matrix.
  - `ST3` Produce threat model delta for history-plane:
    - storage abuse/DoS
    - metadata leakage risks
    - private space enumeration risks
    - replica-churn risks (data loss) and mitigation posture
  - `ST4` Assign gate ownership and approvers using RACI template.
  - Artifacts:
    - `docs/v2.2/phase0/p0-scope-lock.md`
    - `docs/v2.2/phase0/p0-traceability-matrix.md`
    - `docs/v2.2/phase0/p0-history-plane-threat-model.md`
    - `docs/v2.2/phase0/p0-gate-ownership.md`

### Phase 1 - Archivist capability runtime (G2)
- [ ] `P1-T1` Implement Archivist capability advertisement and selection signals.
  - `ST1` Implement `ArchivistAdvertisement` publication/refresh semantics (per F22 spec).
  - `ST2` Implement client selection inputs:
    - prefer same-Space operator list (if any)
    - otherwise rank by availability, quota, and policy
  - `ST3` Implement deterministic refusal reasons for selection failure (e.g., `NO_ARCHIVIST_AVAILABLE`, `ARCHIVIST_POLICY_DENIED`).
  - Artifacts:
    - `pkg/v22/archivist/advertise/*`
    - `pkg/v22/archivist/advertise/*_test.go`
    - `docs/v2.2/phase1/p1-archivist-selection-contract.md`

- [ ] `P1-T2` Implement ciphertext history segment storage engine.
  - `ST1` Implement storage model:
    - segments keyed by `{SpaceID, ChannelID, SegmentID}`
    - segment references are opaque; no plaintext metadata required beyond size/time range if specified
  - `ST2` Implement quota enforcement:
    - per-Space quota (operator-configured)
    - per-channel cap (optional)
    - deterministic refusal reasons (`QUOTA_EXCEEDED`, `RETENTION_POLICY`, `SEGMENT_TOO_LARGE`)
  - `ST3` Implement retention windows + pruning:
    - time-window retention baseline (e.g., last N days)
    - deterministic prune reasons + metrics
  - Artifacts:
    - `pkg/v22/archivist/store/*`
    - `pkg/v22/archivist/store/*_test.go`
    - `docs/v2.2/phase1/p1-archivist-quota-retention.md`

- [ ] `P1-T3` Implement replication and healing pipeline (ciphertext-only).
  - `ST1` Define replication policy:
    - default replica count `r` (e.g., 3) and minimum acceptable `r_min` (e.g., 2)
    - when `r` cannot be met: deterministic degraded-mode labeling (`HISTORY_DURABILITY_DEGRADED`)
  - `ST2` Implement multi-archivist write pipeline:
    - publish segment to `r` distinct archivists where possible
    - record replica-set in `HistorySegmentManifest` (or sidecar) without leaking private membership
  - `ST3` Implement background healing:
    - detect missing replicas (archivist offline, quota prune, manifest mismatch)
    - re-replicate to new archivists until `r` met, bounded by backoff and policy
  - `ST4` Implement deterministic refusal reasons and health states:
    - `REPLICA_TARGET_UNMET`, `REPLICA_WRITE_PARTIAL`, `REPLICA_HEALING_IN_PROGRESS`
  - Artifacts:
    - `pkg/v22/archivist/replicate/*`
    - `pkg/v22/archivist/replicate/*_test.go`
    - `docs/v2.2/phase1/p1-replication-contract.md`

### Phase 2 - History integrity + retrieval protocol (G3)
- [ ] `P2-T1` Implement history integrity primitives.
  - `ST1` Implement `HistoryHead` production/verification:
    - signed head commits to latest segment manifest(s)
  - `ST2` Implement `HistorySegmentManifest` structure:
    - segment list + hashes (or Merkle root) per spec
    - includes time range boundaries for backfill targeting
    - includes replica-set references where required by replication policy
  - `ST3` Implement deterministic verification errors:
    - `HISTORY_HEAD_INVALID_SIGNATURE`
    - `MANIFEST_HASH_MISMATCH`
    - `SEGMENT_NOT_FOUND`
  - Artifacts:
    - `pkg/v22/history/integrity/*`
    - `pkg/v22/history/integrity/*_test.go`
    - `docs/v2.2/phase2/p2-integrity-reason-taxonomy.md`

- [ ] `P2-T2` Implement retrieval endpoints and private Space anti-enumeration.
  - `ST1` Implement retrieval endpoints for:
    - `HistoryHead`
    - `HistorySegmentManifest`
    - ciphertext segment blocks
  - `ST2` Implement anti-enumeration:
    - retrieval keys derived from join secret / membership secret
    - requests without membership proof must fail without leaking existence
  - `ST3` Implement replay/abuse defenses:
    - request rate limits
    - bounded response sizes
  - Artifacts:
    - `pkg/v22/history/retrieve/*`
    - `pkg/v22/history/retrieve/*_test.go`
    - `docs/v2.2/phase2/p2-private-space-anti-enumeration.md`

- [ ] `P2-T3` Integrate moderation redaction/tombstones into backfill apply semantics.
  - `ST1` Ensure tombstone events override any backfilled plaintext deterministically.
  - `ST2` Define and test “cannot guarantee deletion” disclosure surfaces (docs + UI wording).
  - `ST3` Ensure tombstoned content is never re-indexed into local FTS during backfill.
  - Artifacts:
    - `pkg/v22/history/apply/*`
    - `tests/e2e/v22/redaction_backfill_*`
    - `docs/v2.2/phase2/p2-redaction-backfill-contract.md`

### Phase 3 - Client backfill + search coverage integration (G4)
- [ ] `P3-T1` Implement time-range backfill protocol in client runtime.
  - `ST1` Backfill request model:
    - request `{SpaceID, ChannelID, TimeRange, MaxBytes/MaxSegments}`
    - no keyword-bearing requests by default
  - `ST2` Integrate with local store:
    - verified segments decrypt → apply to timeline store
    - idempotent apply + dedupe
  - `ST3` Progress and retry:
    - bounded retries + exponential backoff
    - deterministic failure reasons (`BACKFILL_DENIED`, `BACKFILL_INCOMPLETE`, `BACKFILL_VERIFY_FAILED`)
  - Artifacts:
    - `pkg/v22/history/backfill/*`
    - `pkg/v22/history/backfill/*_test.go`
    - `tests/e2e/v22/backfill_*`

- [ ] `P3-T2` Extend harmolyn search UX to reflect coverage + offer backfill.
  - `ST1` Coverage labeling:
    - show “searched local history from A..B”
    - show missing windows and offer backfill-by-time-range
  - `ST2` No-limbo UX invariants:
    - explicit reasons for “no results” vs “history missing”
    - deterministic next actions (backfill, adjust range, cancel)
  - `ST3` Backfill UX surfaces:
    - progress, cancel, pause-on-metered-network hooks (policy only)
  - Artifacts:
    - `cmd/harmolyn/*`
    - `docs/v2.2/phase3/p3-history-search-ux-contract.md`

### Phase 4 - Validation matrix and Podman scenarios (G5, G6)
- [ ] `P4-T1` Add adversarial/privacy/abuse tests for history-plane.
  - `ST1` Anti-enumeration tests for private Spaces.
  - `ST2` Forged manifest/head signature tests.
  - `ST3` Quota/retention enforcement tests (abuse and edge cases).
  - `ST4` Keyword leakage checks (assert only time-range requests exist by default).
  - `ST5` Replica policy tests:
    - `r` and `r_min` degraded-mode behavior
    - healing under archivist churn
  - Artifacts: `tests/e2e/v22/*`, `tests/perf/v22/*`

- [ ] `P4-T2` Add Podman multi-node history scenarios (Xorein nodes).
  - `ST1` Offline catch-up: client offline → messages sent → client backfill on reconnect.
  - `ST2` Multi-archivist selection + failover scenario.
  - `ST3` Quota exceeded scenario with deterministic refusal reasons.
  - `ST4` Replica healing scenario: one archivist disappears → heal to new archivist.
  - `ST5` Relay no-long-history-hosting regression: relays must not store segments/manifests.
  - `ST6` Deterministic pass/fail probes and result manifest output.
  - Artifacts:
    - `containers/v2.2/*`
    - `scripts/v22-history-scenarios.sh`
    - `docs/v2.2/phase4/p4-podman-scenarios.md`

### Phase 5 - v23 spec package and closure (G7, G8)
- [ ] `P5-T1` Produce `F23` hardening specification package.
  - `ST1` Security hardening gates for history/search:
    - abuse resistance, quota defaults, retention defaults
    - privacy conformance (no keyword leakage by default)
    - replication defaults and durability labeling
  - `ST2` Operator readiness spec:
    - Archivist runbook, monitoring, storage growth alarms
    - upgrade/rollback drills
  - `ST3` SLO targets and perf regression gates for backfill/search.
  - Artifacts:
    - `docs/v2.2/phase4/f23-history-hardening-spec.md`
    - `docs/v2.2/phase4/f23-proto-delta.md`
    - `docs/v2.2/phase4/f23-acceptance-matrix.md`

- [ ] `P5-T2` Publish v22 evidence bundle and promotion recommendation.
  - `ST1` Command outputs for compatibility/test/e2e/perf checks.
  - `ST2` Podman scenario outputs and deterministic result manifests.
  - `ST3` `F22` as-built conformance report against v21 `F22` specs.
  - `ST4` Risk closure and promotion recommendation.
  - `ST5` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts:
    - `docs/v2.2/phase5/p5-evidence-bundle.md`
    - `docs/v2.2/phase5/p5-risk-register.md`
    - `docs/v2.2/phase5/p5-as-built-conformance.md`
    - `docs/v2.2/phase5/p5-gate-signoff.md`
    - `docs/v2.2/phase5/p5-evidence-index.md`

## Risk register (v22)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R22-1 | Archivist role drifts into privileged node class | explicit capability-only contract + tests | contract review passes |
| R22-2 | Private Space history becomes enumerable | join-secret-derived keys + negative tests | anti-enumeration tests pass |
| R22-3 | Storage abuse/DoS | quotas + retention + rate limits | abuse tests pass |
| R22-4 | Backfill introduces UX limbo | reason taxonomy + deterministic progress states | journey tests pass |
| R22-5 | Metadata leakage via search | no keyword-bearing requests by default | privacy tests pass |
| R22-6 | Relay boundary regresses | dedicated Podman regression scenarios | relay-boundary tests pass |
| R22-7 | Data loss due to insufficient replication or churn | replica policy + healing pipeline + tests | healing + degraded-mode tests pass |

## Decision log (v22)
- `D22-1`: Archivist is an opt-in capability; it stores ciphertext segments only and has no protocol authority.
- `D22-2`: Backfill requests are time-range only by default to avoid keyword leakage.
- `D22-3`: Private Space history endpoints are anti-enumeration protected (membership-derived keys; no existence leakage).
- `D22-4`: Relays do not store long-lived history segments/manifests; durable history is an Archivist capability.
- `D22-5`: History segments are replicated to `r` archivists by default with a bounded healing process; durability is explicitly labeled when `r` cannot be met.
