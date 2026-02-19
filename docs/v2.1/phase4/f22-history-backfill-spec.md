# F22 History + Backfill Spec

> Planning artifact for the v22 delivery. This spec package captures the interface, guardrails, and failure taxonomy that v22 will implement on top of the v21 timeline/search baseline.

## Purpose and scope
- Bound the Archivist role so that durable history segments can be optionally offered without forcing relays or clients to host unencrypted payloads beyond their own local stores.
- Document the contracts that unify the persistence, discovery, and retrieval of HistoryHead segments, with explicit labels for history availability, coverage, and privacy boundaries.
- Keep this artifact additive: no wire-breaking changes, no behavior that contradicts the AGPL (runtime code) or CC-BY-SA (normative spec) foundations already in place.

## Archivist capability model
- Advertise capacity in both space- and channel-scoped descriptors (`AdvertisedRetention`, `AdvertisedQuota`, `AdvertisedEpoch`) so peers can choose which Archivist to trust.
- Allow peers to enroll and refresh via deterministic input (join secret + recorded metadata) and document quota/retention enforcement with refusal reasons such as `ARCHIVIST_QUOTA_EXCEEDED`, `ARCHIVIST_EPOCH_MISMATCH`, and `ARCHIVIST_UNTRUSTED_JOINER`.
- Define default replication factors (`r`, `r_min`) and clearly state whether the Archivist accepts writes or only read requests for each segment.

## History integrity and segment manifests
- Define `HistoryHead`, `HistorySegmentManifest`, and per-segment commitment proofs that contain lightweight signatures or Merkle roots so clients can verify tamper resistance.
- Record segment bounds (start timestamp, end timestamp, hash), manifest metadata (channel identifiers, missing ranges, tombstone pointers), and canonical history proofs in a verifiable message.
- Outline defensive behaviors when an Archivist cannot prove a segment (e.g., return a `RETRIEVAL_PROOF_FAILURE` refusal and keep UI consumers in `history-unavailable` state).

## Backfill protocol baseline
- Support time-window-only requests (`BackfillRequest`) with optional coverage hint metadata instead of free-form keyword queries, keeping leakage risk minimal.
- Define deterministic streaming responses (`BackfillSegment`, `BackfillResponse`) that include manifests, segment data, and clear failure reasons while guarding replay/dup ordering via nonce/epoch descriptors.
- Document how clients can resume partial backfills, cancel open streams, and interpret progress/failure signals such as `BACKFILL_TOO_NEW`, `BACKFILL_RATE_LIMITED`, and `BACKFILL_INTEGRITY_MISMATCH`.

## Search coverage and privacy continuation
- Extend the v21 search coverage labels so that remote backfills can set shareable coverage metadata (time ranges, channel scopes) without exposing plaintext keywords.
- Require every backfill response to carry coverage guarantees and annotate UI surfaces with `coverage-available` or `coverage-missing` states that complement the local coverage labels defined in `docs/v2.1/phase2/p2-search-contract.md`.

## Moderation and E2EE interplay
- Ensure that redaction/tombstone events always override backfilled segments and are delivered with deterministic collapse semantics (`tombstone supersedes segment entry`), so clients never rehydrate removed plaintext.
- Document how Archivists propagate tombstone metadata to downstream peers without leaking edit history timelines beyond what legitimate participants already possess.

## Traceability
- This specification feeds `P4-T1` (Phase 4, `G6`) of TODO_v21 and depends on the Phase 3 validation matrices (`docs/v2.1/phase3/p3-podman-scenarios.md`) plus the v21 search/persistence contracts.
- Gate evidence for this spec will live in `docs/v2.1/phase5/p5-evidence-index.md` under `EV-v21-G6-###` entries until the commands and manifests are captured and promoted.

## Licensing
- Runtime artifacts derived from this spec follow the AGPL requirements documented in `AGENTS.md` and `go.mod`.
- Normative text and tables remain CC-BY-SA, consistent with the project baseline for docs.
