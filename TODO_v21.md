# TODO v21 - Xorein (backend) + harmolyn (frontend) Execution Plan (v2.1)

## Status
Planning artifact only. This file defines v21 implementation and validation requirements. It does not claim implementation completion.

## Naming + binary split note (architecture delta)
- **Xorein**: backend node/runtime binary (headless-capable). Provides the protocol runtime, storage, crypto, DHT/pubsub, relay/archivist capabilities as configured.
- **harmolyn**: frontend UI binary (Gio). Consumes Xorein runtime APIs (in-process library) and renders UX.  
  - v21 default posture: harmolyn embeds Xorein runtime as a library (single process) to avoid introducing a new local-IPC attack surface in the same sprint.
  - A “separate local daemon + UI attachment” mode is explicitly deferred to `F24` unless already implemented earlier.

## Version Isolation Contract (mandatory)
- v21 cannot advance unless all v21 in-scope spec, code, tests, docs, ops artifacts, and evidence are complete.
- v21 implements only `F21` from v20 `F21` seed package (additive proto only).
- v21 must also publish full `F22` spec package for v22 implementation.

## Version role
- Implements: `F21` (encrypted local timeline persistence + deterministic hydration + local search baseline).
- Specifies: `F22` (distributed history/backfill + archivist capability + E2EE-safe search coverage contracts).

## Critical scope (v21)
- Implement encrypted local timeline store for Spaces/channels (Xorein runtime):
  - restart-safe timeline hydration + pagination
  - idempotent apply + deduplication by canonical MessageID/EventID
  - retention/pruning baseline with deterministic reasons
- Implement local full-text search over locally decrypted history (Xorein runtime):
  - channel-scoped search + time range filters
  - explicit “search coverage” and “history availability” labeling (local-only; no implied network backfill)
- Integrate moderation redactions with persisted timelines (Xorein runtime):
  - redaction produces tombstones (plaintext removal) and search-index removal
  - **note:** redaction is client-enforced in E2EE; deletion cannot be guaranteed against previously-seen plaintext (document explicitly)
- Implement harmolyn UX for:
  - offline timeline availability states
  - search entry points + coverage labels
  - retention controls (minimal: defaults + “clear local history”)
- Preserve relay no-long-history-hosting boundary with explicit regression checks:
  - relays MUST NOT store long-lived history segments/manifests (Archivist is the durable-history capability; v22)
  - if any existing store-and-forward TTL exists elsewhere in the roadmap, it remains bounded and ciphertext-only (not expanded here)

## Out of scope (defer)
- Distributed history retrieval / network backfill (moved to `F22` / v22).
- Remote keyword search assistance (keyword leakage risks) remains out of scope by default.
- Introducing a mandatory “local daemon + UI” attachment model is out of scope for v21 unless already implemented earlier.
- Any change that requires breaking wire compatibility (major-path governance) is out of scope for v21.

## Dependencies and relationships
- Inputs from v20 (must exist before v21 starts):
  - `docs/v2.0/phase4/f21-backlog-and-spec-seeds.md`
  - `docs/v2.0/phase4/f21-proto-delta.md`
  - `docs/v2.0/phase4/f21-acceptance-matrix.md`
  - `docs/v2.0/phase4/f21-deferral-register.md`
- Outputs consumed by v22:
  - `docs/v2.1/phase4/f22-history-backfill-spec.md`
  - `docs/v2.1/phase4/f22-proto-delta.md`
  - `docs/v2.1/phase4/f22-acceptance-matrix.md`
- Binary deliverables (v21+ convention):
  - `cmd/xorein` (backend node/daemon)
  - `cmd/harmolyn` (frontend UI)

## Entry criteria (must be true before implementation starts)
- `v20` is in `promoted` state with evidence bundle and as-built conformance report.
- `F21` spec seeds from v20 exist and are approved (or explicitly superseded by v21 Phase 0 scope lock).
- v21 deferred scope list is frozen and approved.

## Promotion gates (must all pass)
- `G0` Scope lock and requirement-to-artifact traceability complete.
- `G1` Additive wire compatibility checks pass (no breaking changes).
- `G2` Encrypted local store implementation complete (schema + migrations + retention).
- `G3` harmolyn timeline/search UX complete with deterministic no-limbo states.
- `G4` Local search correctness + privacy test matrix complete.
- `G5` Podman persistence scenarios complete.
- `G6` v22 spec package complete.
- `G7` Docs and evidence bundle complete.
- `G8` Relay no-long-history-hosting regression checks pass.
- `G9` `F21` as-built conformance report completed against v20 `F21` seed package.

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v21/...` (or declared equivalent if paths differ)
- `go test ./tests/perf/v21/...` (or declared equivalent if paths differ)
- `go build ./cmd/xorein`
- `go build ./cmd/harmolyn`
- `make check-full`
- `scripts/v21-persistence-scenarios.sh`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v21-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## LLM agent execution contract (mandatory)
- `P0-T1`: Dependency rule—start only after the entry criteria (v20 promoted, `F21` seeds approved, deferred list frozen) and tooling for scope lock are ready; Acceptance rule—all ST1–ST5 substeps complete, associated artifacts refreshed, and the `G0` gate row marked Pass; Evidence rule—attach at least one `EV-v21-G0-###` entry tied to this task plus the required command outputs or a `not applicable` note; Blocker state taxonomy—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint rule—before executing any supporting command (e.g., `make check-full` to exercise gating docs), record the planned command and capture its observed output immediately afterward.
- `P1-T1`: Dependency rule—begin only after `G0` scope lock and `G1` additive wire checks pass; Acceptance rule—complete ST1–ST3 (store engine selection, threat model, keying model) with artifacts documented and `G2` marked Pass; Evidence rule—attach one `EV-v21-G2-###` entry plus command outputs or `not applicable`; Blocker codes—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint rule—note the next validation command (e.g., `go test ./pkg/v21/store/...` or `buf lint`) before running and log the output after.
- `P1-T2`: Dependency rule—stage after `P1-T1` artifacts land and `G2` stays green; Acceptance rule—complete ST1–ST3 (canon equality, ingestion pipeline, failure reasons), refresh artifact sets, and mark `G2` entry Pass; Evidence rule—≥1 `EV-v21-G2-###` entry plus command outputs or `not applicable`; Blocker taxonomy—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the next command (e.g., `go test ./pkg/v21/store/ingest/...`) before running and capture its output.
- `P1-T3`: Dependency rule—requires `P1-T2` stability and `G2` gating artifacts; Acceptance rule—finish ST1–ST3 (ordering, pagination, empty-history states) with docs/tests updated and gate row Pass; Evidence rule—≥1 `EV-v21-G2-###` entry plus required command outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—log planned command (e.g., `go test ./pkg/v21/store/hydrate/...`) and observed output.
- `P1-T4`: Dependency rule—after `P1-T3` and `G2` prerequisites; Acceptance rule—complete ST1–ST3 (retention defaults, deterministic pruning, clear-history action) with artifacts and G2 row Pass; Evidence rule—≥1 `EV-v21-G2-###` entry plus command outputs or `not applicable`; Blocker taxonomy—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the command you will run (e.g., `go test ./pkg/v21/store/retention/...`) and capture output.
- `P2-T1`: Dependency rule—wait for all Phase 1 artifacts and `G2` pass; Acceptance rule—complete ST1–ST3 (FTS index, coverage model, query limits) with docs/tests updated and mark `G3` Pass; Evidence rule—attach `EV-v21-G3-###` plus command outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—before running commands such as `go test ./pkg/v21/search/...`, note the planned command and log its output.
- `P2-T2`: Dependency rule—after `P2-T1` artifacts and `G3` coverage; Acceptance rule—complete ST1–ST3 (tombstones, idempotent apply, search removal) with artifacts and `G3` row Pass; Evidence rule—≥1 `EV-v21-G3-###` entry plus command outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the next command (e.g., `go test ./pkg/v21/store/tombstone/...`) and capture output.
- `P2-T3`: Dependency rule—after `P2-T2` redaction readiness; Acceptance rule—complete ST1–ST3 (timeline hydration UX, search labels, retention controls) with UX docs updated and `G3` Pass; Evidence rule—attach one `EV-v21-G3-###` entry plus command outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—log the command (e.g., `go test ./cmd/harmolyn/...`) before running and save its output.
- `P3-T1`: Dependency rule—start after Phase 2 artifacts and `G3` gate; Acceptance rule—complete ST1–ST4 (restart, migration, corruption, retention invariants) with tests/docs updated and mark `G4` Pass; Evidence rule—≥1 `EV-v21-G4-###` entry plus command outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the next command (e.g., `go test ./pkg/v21/**/*_test.go`) and note output.
- `P3-T2`: Dependency rule—requires P3-T1 validation done; Acceptance rule—complete ST1–ST3 (search correctness, time/channel scoping, stress tests) and mark `G4` Pass; Evidence rule—≥1 `EV-v21-G4-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—document the planned command (e.g., `go test ./tests/e2e/v21/search_*`) and capture its output.
- `P3-T3`: Dependency rule—after P3-T2 and `G4` tests; Acceptance rule—complete ST1–ST4 (Podman multi-peer scenarios, clear history, relay regression, deterministic probes), ensure manifests exist, and satisfy `G5`; Evidence rule—attach `EV-v21-G5-###` entry plus command outputs (e.g., `scripts/v21-persistence-scenarios.sh` results) or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—log the next command (`scripts/v21-persistence-scenarios.sh`) before execution and record observed output afterward.
- `P4-T1`: Dependency rule—after validation gates (G4/G5) and Phase 3 artifacts; Acceptance rule—complete ST1–ST6 for `F22` spec package and mark `G6` Pass; Evidence rule—attach `EV-v21-G6-###` entry plus relevant command outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—document the planned command (e.g., documentation rendering or `buf lint`) and capture output.
- `P5-T1`: Dependency rule—after all implementation/spec phases and `G6` completion; Acceptance rule—complete ST1–ST5 (command outputs, Podman manifests, conformance report, risk closure, gate sign-off), update artifacts, and mark `G7`, `G8`, `G9` Pass; Evidence rule—attach entries for `EV-v21-G7-###`, `EV-v21-G8-###`, and `EV-v21-G9-###` plus command outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the next command (`make check-full`, `go build ./cmd/xorein`, `go build ./cmd/harmolyn`, etc.) before running and capture output afterward.

### Phase dependency and evidence map
| Phase | Top-level tasks | Must start after | Gate target(s) | Minimum evidence | Command hints |
| --- | --- | --- | --- | --- | --- |
| Phase 0 | `P0-T1` | Entry criteria (v20 promoted, `F21` seeds documented, deferred scope frozen) | `G0` | `EV-v21-G0-###` | Plan scope-lock commands (e.g., `make check-full`) before running and capture output. |
| Phase 1 | `P1-T1`, `P1-T2`, `P1-T3`, `P1-T4` | `G0` + `G1` (additive wire checks) completion | `G1`, `G2` | `EV-v21-G1-###`, `EV-v21-G2-###` | Record store/integration commands (`go test ./pkg/v21/store/...`, `buf lint`) and log outputs. |
| Phase 2 | `P2-T1`, `P2-T2`, `P2-T3` | Phase 1 artifacts + `G2` | `G3` | `EV-v21-G3-###` | Note search and UX commands (`go test ./pkg/v21/search/...`, `go test ./cmd/harmolyn/...`) before running and capture outputs. |
| Phase 3 | `P3-T1`, `P3-T2`, `P3-T3` | `G3` + Phase 2 validation artifacts | `G4`, `G5` | `EV-v21-G4-###`, `EV-v21-G5-###` | Plan validation (`go test ./pkg/v21/**/*_test.go`, `go test ./tests/e2e/v21/*`) plus Podman script (`scripts/v21-persistence-scenarios.sh`); capture outputs. |
| Phase 4 | `P4-T1` | `G4` + `G5` completion | `G6` | `EV-v21-G6-###` | Record doc/spec commands (`buf lint`, doc render) before running and capture outputs. |
| Phase 5 | `P5-T1` | All prior gates (G0–G6) & artifacts | `G7`, `G8`, `G9` | `EV-v21-G7-###`, `EV-v21-G8-###`, `EV-v21-G9-###` | Capture command outputs from `go build ./cmd/xorein`, `go build ./cmd/harmolyn`, `make check-full`, plus supporting scripts and document logs. |

## Phase plan

### Phase 0 - Scope lock and seed import (G0)
- [ ] `P0-T1` Freeze v21 scope to `F21` persistence + local search baseline.
  - `ST1` Import and reconcile v20 `F21` seeds; record any deltas as `v21` scoped decisions.
  - `ST2` Produce requirement-to-artifact traceability matrix (spec → code → tests → docs).
  - `ST3` Freeze explicit non-goals (no network backfill; no remote keyword search; no mandatory local-daemon split).
  - `ST4` Assign gate ownership and approvers using RACI template.
  - `ST5` **Backend/frontend boundary lock (v21+ naming convention):**
    - Xorein runtime packages MUST NOT depend on Gio/UI packages.
    - harmolyn UI MUST consume Xorein via interfaces (or a narrow internal API layer) so that headless builds remain possible.
  - Artifacts:
    - `docs/v2.1/phase0/p0-scope-lock.md`
    - `docs/v2.1/phase0/p0-traceability-matrix.md`
    - `docs/v2.1/phase0/p0-gate-ownership.md`
    - `docs/v2.1/phase0/p0-decisions.md`
    - `docs/v2.1/phase0/p0-binary-boundary.md`

### Phase 1 - Encrypted local timeline store (G2)
- [ ] `P1-T1` Define local store architecture and encryption profile (Xorein runtime).
  - `ST1` Select store engine (SQLite + SQLCipher recommended) and document threat model:
    - adversary: filesystem compromise, offline copy, device theft
    - non-goals: malware with live access to decrypted process memory
  - `ST2` Define keying model and key rotation hooks:
    - local DB key derived from device secret + identity secret (or OS keystore where available)
    - explicit “wipe local history” destroys DB key + DB file
  - `ST3` Define schema v1 (versioned) for:
    - Spaces, channels, membership cache (minimal)
    - timeline rows (MessageID/EventID, sender, timestamps, delivery state, ciphertext envelope reference, plaintext body, attachments metadata)
    - tombstones/redactions (reason + audit pointer)
  - Artifacts:
    - `docs/v2.1/phase1/p1-store-threat-model.md`
    - `docs/v2.1/phase1/p1-store-schema.md`
    - `pkg/v21/store/*`

- [ ] `P1-T2` Implement store ingestion, dedupe, and idempotent apply rules (Xorein runtime).
  - `ST1` Canonical identity of a timeline event:
    - define “same event” equality rules (MessageID/EventID + channel + sender as needed)
    - define idempotent re-apply semantics under replay/duplicate delivery
  - `ST2` Implement write pipeline that persists:
    - received/decrypted messages
    - local-sent pending messages (draft delivery states) with convergence rules
  - `ST3` Implement deterministic failure reasons for store errors:
    - `STORE_LOCKED`, `STORE_CORRUPT`, `STORE_MIGRATION_REQUIRED`, `STORE_QUOTA_EXCEEDED`
  - Artifacts:
    - `pkg/v21/store/ingest/*`
    - `pkg/v21/store/ingest/*_test.go`
    - `docs/v2.1/phase1/p1-store-reason-taxonomy.md`

- [ ] `P1-T3` Implement deterministic timeline hydration + pagination (Xorein runtime).
  - `ST1` Define canonical ordering:
    - prefer protocol ordering primitives if available (Lamport/HLC/sequence)
    - otherwise define stable local ordering with explicit “order may differ across devices” disclaimer
  - `ST2` Implement pagination cursors and stable page boundaries across restarts.
  - `ST3` Implement explicit empty-history states:
    - “no local history yet”
    - “history cleared locally”
    - “history missing (backfill not supported in v21)”
  - Artifacts:
    - `pkg/v21/store/hydrate/*`
    - `pkg/v21/store/hydrate/*_test.go`
    - `docs/v2.1/phase1/p1-hydration-contract.md`

- [ ] `P1-T4` Implement retention/pruning baseline (Xorein runtime).
  - `ST1` Define default retention window and size cap (documented, configurable).
  - `ST2` Implement deterministic pruning policy:
    - prune oldest-first per channel, preserve pinned/system events if applicable
    - ensure pruning updates indexes consistently (no orphaned search rows)
  - `ST3` Implement “clear local history” action:
    - wipes DB + keys and resets local search coverage to empty
  - Artifacts:
    - `pkg/v21/store/retention/*`
    - `pkg/v21/store/retention/*_test.go`
    - `docs/v2.1/phase1/p1-retention-policy.md`

### Phase 2 - Local search + moderation integration (G3)
- [ ] `P2-T1` Implement local full-text search index over encrypted store (Xorein runtime).
  - `ST1` Implement FTS index (inside encrypted DB) with:
    - per-channel scoping
    - sender filter
    - time range filter
  - `ST2` Implement deterministic “search coverage” model:
    - coverage windows per channel (what time range is actually indexed)
    - visible UI label and machine-readable status
  - `ST3` Implement safe query limits and timeouts (avoid UI hangs on large DBs).
  - Artifacts:
    - `pkg/v21/search/*`
    - `pkg/v21/search/*_test.go`
    - `docs/v2.1/phase2/p2-search-contract.md`

- [ ] `P2-T2` Integrate moderation redaction into persisted timelines (Xorein runtime).
  - `ST1` Define tombstone behavior:
    - remove plaintext body and attachments preview text
    - preserve minimal envelope metadata + redaction reason code + audit pointer
    - explicitly document the E2EE limitation: redaction cannot retroactively delete plaintext from compromised/previously-synced endpoints
  - `ST2` Ensure idempotent apply under replay/duplication.
  - `ST3` Ensure redacted content is removed from FTS index (not searchable).
  - Artifacts:
    - `pkg/v21/store/tombstone/*`
    - `pkg/v21/store/tombstone/*_test.go`
    - `docs/v2.1/phase2/p2-redaction-persistence-contract.md`

- [ ] `P2-T3` Implement harmolyn UX for persistence + search.
  - `ST1` Timeline hydration from local store with explicit empty/history-missing banners.
  - `ST2` Search UI with coverage labels and deterministic empty states.
  - `ST3` Minimal retention controls:
    - show current retention window/size
    - clear local history action with confirmation
  - Artifacts:
    - `cmd/harmolyn/*`
    - `docs/v2.1/phase2/p2-ux-contract.md`

### Phase 3 - Validation matrix and Podman scenarios (G4, G5)
- [ ] `P3-T1` Add unit and integration tests for persistence invariants.
  - `ST1` Restart invariants: ordering stable, no duplication, no missing items after restart.
  - `ST2` Migration invariants: schema vN→vN+1 upgrades are deterministic.
  - `ST3` Corruption detection: corrupted DB produces deterministic refusal reason and safe recovery path.
  - `ST4` Retention/prune invariants: prune reasons deterministic; indexes consistent.
  - Artifacts: `pkg/v21/**/*_test.go`, `tests/e2e/v21/*`

- [ ] `P3-T2` Add search correctness and privacy tests.
  - `ST1` Redaction removes results from search.
  - `ST2` Time range and channel scoping correctness.
  - `ST3` Large-DB stress tests with enforced query limits/timeouts.
  - Artifacts: `tests/e2e/v21/search_*`, `tests/perf/v21/*`

- [ ] `P3-T3` Add Podman scenarios for persistence (Xorein nodes; harmolyn is not required in containers).
  - `ST1` Multi-peer chat + restart + resume scenarios in container network.
  - `ST2` Local history clear and recovery scenarios.
  - `ST3` Relay no-long-history-hosting regression scenario for persistence/search (ensure relay does not store durable history segments).
  - `ST4` Deterministic pass/fail probes and result manifest output.
  - Artifacts:
    - `containers/v2.1/*`
    - `scripts/v21-persistence-scenarios.sh`
    - `docs/v2.1/phase3/p3-podman-scenarios.md`

### Phase 4 - v22 spec package (G6)
- [ ] `P4-T1` Produce distributed history/backfill specification package (`F22`).
  - `ST1` Archivist capability model (non-privileged, opt-in):
    - advertisement + selection inputs
    - quota + retention windows + refusal reasons
    - replication-factor target (define default `r` and minimum acceptable `r_min`)
  - `ST2` History integrity model:
    - `HistoryHead` and segment manifests
    - verification commitments (signatures + hashes/Merkle proofs)
  - `ST3` Private Space anti-enumeration:
    - derive retrieval keys from join secret or MLS epoch secrets
    - no public listing of history endpoints for private Spaces
  - `ST4` Backfill protocol baseline:
    - time-range backfill requests only by default (avoid keyword leakage)
    - deterministic progress + failure reason taxonomy
  - `ST5` Search coverage and labeling contract extended for backfill (but no remote keyword search).
  - `ST6` Moderation interaction note:
    - redaction/tombstone events MUST override backfilled content deterministically
    - explicitly document limits of true deletion in E2EE
  - Artifacts:
    - `docs/v2.1/phase4/f22-history-backfill-spec.md`
    - `docs/v2.1/phase4/f22-proto-delta.md`
    - `docs/v2.1/phase4/f22-acceptance-matrix.md`

### Phase 5 - Closure and evidence (G7)
- [ ] `P5-T1` Publish v21 evidence bundle and promotion recommendation.
  - `ST1` Command outputs for compatibility/test/e2e/perf checks.
  - `ST2` Podman scenario outputs and deterministic result manifests.
  - `ST3` `F21` as-built conformance report against v20 `F21` seeds.
  - `ST4` Risk closure and promotion recommendation.
  - `ST5` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts:
    - `docs/v2.1/phase5/p5-evidence-bundle.md`
    - `docs/v2.1/phase5/p5-risk-register.md`
    - `docs/v2.1/phase5/p5-as-built-conformance.md`
    - `docs/v2.1/phase5/p5-gate-signoff.md`
    - `docs/v2.1/phase5/p5-evidence-index.md`

## Risk register (v21)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R21-1 | Local store corruption causes data loss or crashes | corruption detection + safe fallback + tests | corruption tests pass |
| R21-2 | Search index leaks redacted content | tombstone enforcement + index rebuild tests | redaction-search tests pass |
| R21-3 | Retention/prune breaks pagination | stable cursor contract + property tests | pagination tests pass |
| R21-4 | Store introduces UI limbo states | explicit history-availability banners + reason taxonomy | no-limbo journey tests pass |
| R21-5 | Relay boundary regresses (accidental durable storage) | dedicated regression scenarios | relay-boundary tests pass |
| R21-6 | Backend/frontend split causes dependency leakage (UI deps in runtime) | explicit boundary lock + build checks | `cmd/xorein` builds without UI deps |

## Decision log (v21)
- `D21-1`: v21 persistence stores plaintext only inside an encrypted local database; no plaintext is stored unencrypted on disk.
- `D21-2`: v21 search is strictly local; no remote keyword search or keyword-bearing queries exist in v21.
- `D21-3`: v21 introduces explicit “history availability / search coverage” labels; empty timelines must not be silently ambiguous.
- `D21-4`: Relays do not store long-lived history segments/manifests; durable history is an Archivist capability (v22).
- `D21-5`: Starting v21, Xorein refers to backend runtime/binary and harmolyn refers to frontend UI binary; earlier “Xorein includes UI” phrasing is legacy.
