# TODO_v07_EXEC_ADDENDUM.md

> Status: **Execution addendum (implementation required).**
>
> Constraint: `TODO_v07.md` is already consumed and MUST NOT be edited. This addendum does not change v0.7 scope or numbering; it adds the missing “build + validate” requirements so that executing v0.7 yields runnable code and deterministic tests.
>
> Use pattern for agents:
> 1) Provide agents `TODO_v07.md` (scope + contracts + task IDs).
> 2) Also provide this addendum (implementation deliverables + evidence anchors).
> 3) Completion requires both: contract artifacts from v07 **and** runnable implementations + tests defined here.

---

## A) Evidence anchors (must exist by V7-G6)

These anchors are intentionally parallel to the v0.1–v0.6 “execution artifact” convention.

1) `pkg/v07/` (or equivalent)
- `pkg/v07/storeforward/` (TTL + replication + purge)
- `pkg/v07/retention/` (per-server policy + transitions)
- `pkg/v07/archivist/` (capability advertisement + full-history storage)
- `pkg/v07/historysync/` (protocol handlers + resume + checkpoints)
- `pkg/v07/merkle/` (Merkle tree + proof verification)
- `pkg/v07/search/` (SQLCipher FTS5 indexing + query/filter engine)
- `pkg/v07/push/` (push payload envelope + client registration)

2) `cmd/`
- `cmd/aether/` updates for v0.7 features.
- `cmd/push-relay/` (reference push relay service; minimal runnable).

3) `docs/v0.7/`
- Protocol / contract artifacts referenced by v07 (`VA-*`).
- A runnable “v0.7 demo” guide.

4) `tests/`
- `tests/e2e/v07/` deterministic scenario suite.
- Unit + integration tests colocated with packages.

5) `containers/`
- `containers/v0.7/` docker-compose (or equivalent) that can bring up: relay/bootstrap + push-relay + optional test harness.

---

## B) v0.7 Definition of Done (DoD)

v0.7 is “done” only when ALL are true:

1) **A client that was offline receives messages after reconnecting** using store-and-forward (30-day TTL target, k=20 replication target).
2) **A new device can sync history** for a joined server/channel/DM using the history-sync protocol, verifying integrity with Merkle proofs.
3) **Full-text search works locally** via SQLCipher FTS5 with required filters (from user, date range, has file, has link), scoped correctly.
4) **Push relay path exists and is runnable** (even if APNs/FCM are mocked in CI):
   - client registration
   - encrypted payload forwarding (relay-blind)
   - desktop notifications on receipt.
5) **E2E proof**: `tests/e2e/v07/*` runs via a single documented command and passes locally.

---

## C) Implementation requirements mapped to existing v07 tasks

This section adds “implementation deliverables” to each existing v07 task family. It does not rename tasks.

### C1) Phase 1 (V7-G1) — store-forward + retention + archivist

1) **P1-T1 (store-and-forward TTL + k=20 replication)**
- Required code deliverables:
  - Implement a store-forward record format (ciphertext payload + minimal metadata) and deterministic TTL evaluation.
  - Implement replication worker targeting k=20 across eligible relay/archivist peers.
  - Implement purge worker with deterministic purge reasons (expired, policy purge, storage pressure).
  - Implement degraded-mode signaling (cannot reach k=20) surfaced as reason codes (internal diagnostics).
- Required tests:
  - Unit tests for TTL start rules and purge evaluation.
  - Property-based tests for deterministic “retain vs purge” classification across randomized clock offsets.
  - Integration test with N relays to validate replication fanout and repair behavior.

2) **P1-T2 (per-server configurable retention)**
- Required code deliverables:
  - Policy storage in the server config/state (with signed/audited policy changes per v0.4 permission baseline).
  - Enforce policy in:
    - store-forward acceptance (whether to accept/retain)
    - purge scheduling
    - history sync boundaries (what is eligible to serve)
  - Implement audit events for policy changes (who, when, old→new).
- Required tests:
  - Unit tests for policy precedence + evaluation.
  - E2E scenario: policy tightened triggers deterministic purge and clients see expected history boundary.

3) **P1-T3 (Archivist role for full history storage)**
- Required code deliverables:
  - Capability advertisement: peers announce archivist willingness (bounded metadata only).
  - Selection rules: clients/relays choose archivists deterministically (e.g., peer scoring + diversity).
  - Storage obligations:
    - store full history ciphertext
    - serve history sync requests
    - optional periodic integrity check hooks.
  - Withdrawal behavior: grace period + deterministic “coverage dropped” signaling.
- Required tests:
  - Integration test: archivist join/leave does not break history sync; coverage drop yields deterministic fallback behavior.

### C2) Phase 2 (V7-G2) — Merkle history sync

1) **P2-T1 (history sync protocol lifecycle)**
- Required code deliverables:
  - Implement history sync stream handlers (request/response, resume tokens, checkpointing).
  - Implement mode-epoch segmentation handling:
    - represent epoch boundaries in history streams
    - render “locked history” when keys absent
    - optional “History Capsule” re-encryption workflow stub (if supported).
- Required tests:
  - Unit tests for resume tokens and checkpoint correctness.
  - Fuzz tests for malformed sync messages and proof verification.

2) **P2-T2 / P2-T3 (Merkle construction + proof verification)**
- Required code deliverables:
  - Deterministic Merkle tree construction over canonical history ordering.
  - Proof generation and client verification.
  - Explicit behavior for missing segments, divergent roots, and retry paths.
- Required tests:
  - Property tests: two independent builders produce identical roots for same history.
  - Negative tests: corrupted leaf fails verification deterministically.

### C3) Phase 3 (V7-G3) — SQLCipher FTS5 search

1) **P3-T1 (FTS5 index schema + lifecycle)**
- Required code deliverables:
  - DB migrations to add FTS5 virtual tables for message bodies and selected metadata.
  - Deterministic indexing lifecycle: insert/update/delete events.
  - Scope enforcement: search restricted by server/channel/DM boundaries.
- Required tests:
  - Migration test on a seeded DB.
  - Index rebuild test (drop/rebuild yields same results).

2) **P3-T2 (filters)**
- Required code deliverables:
  - Implement filters:
    - from user
    - date range
    - has file
    - has link
  - Ensure filter semantics are stable and documented.
- Required tests:
  - Query tests covering combinations and edge cases.

### C4) Phase 4 (V7-G4) — push relay + desktop notifications

1) **P4-T1 (encrypted push relay contract)**
- Required code deliverables:
  - Implement push relay reference service (`cmd/push-relay`) that:
    - accepts device registrations
    - accepts encrypted payloads
    - forwards to provider adapters
    - never requires plaintext
  - Provide provider adapters:
    - CI/mock adapter (always available)
    - optional APNs/FCM adapter behind build tags or config flags.
- Required tests:
  - Unit tests: payload size limits, auth, idempotency.
  - Integration test with mock adapter.

2) **P4-T2 (desktop notifications)**
- Required code deliverables:
  - Client notification pipeline that triggers desktop-native notifications on:
    - DM message
    - mention
    - call invite (if present)
  - Deterministic action handling: click opens correct server/channel/message.
- Required tests:
  - “Notification emitted” test using a test double (do not rely on OS UI in CI).

### C5) Phase 5–6 (V7-G5/V7-G6) — integrated validation + handoff

1) **P5-T1 / P5-T2 (integrated scenario suite)**
- Required deliverables:
  - `tests/e2e/v07/` suite that covers at minimum:
    - E2E-SF-01: offline store-forward delivery (client offline, message sent, reconnect, receive)
    - E2E-HS-01: new device history sync + Merkle verification
    - E2E-SR-01: search with each required filter
    - E2E-PR-01: push relay mocked delivery triggers local desktop notification pipeline
    - E2E-EP-01: mode-epoch boundary is represented and locked-history renders deterministically

2) **P6-T1 (release handoff)**
- Required deliverables:
  - `docs/v0.7/README.md` “run it” guide.
  - `docs/v0.7/runbooks/` operator notes for push relay + relays.
  - Deferral register: anything not implemented must be explicitly deferred (no silent gaps).

---

## D) Hard “missing-in-v01–v06” closures that v0.7 must not assume

This section is a safety net: if earlier versions implemented a feature but didn’t lock it down with runnable proof, v0.7 MUST add proof rather than assuming it.

1) **Offline delivery proof**: v0.7 E2E suite MUST include store-forward proof (it is a v0.7 scope bullet, and it is foundational).
2) **Search proof**: v0.7 MUST add deterministic local search proof; do not rely on manual testing.
3) **Upgrade/migration proof**: v0.7 MUST add DB migration tests (SQLCipher + FTS5), even if earlier migrations existed.

---

## E) Minimal operator-facing “working system” demo (v0.7)

A v0.7 PR is not mergeable unless the following can be executed from scratch on a dev machine:

1) Start infra:
- `docker compose -f containers/v0.7/docker-compose.yml up` (or documented equivalent)

2) Start 2 clients (same machine is ok):
- `aether --mode=client --profile=testA`
- `aether --mode=client --profile=testB`

3) Prove:
- A joins server, B goes offline
- A sends message
- B reconnects and receives message
- B adds new device profile and syncs history (Merkle verified)
- Search finds the message using at least one filter

If any step is not possible, it must be recorded as a ship-blocking gap with a specific follow-up task.
