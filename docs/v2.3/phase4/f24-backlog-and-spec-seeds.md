# F24 backlog and spec seeds (Phase4 P4-T2 ST1)

This artifact publishes the `F24` seed catalog required by `G10` plus the additive proto delta for Phase 4. All seeds preserve the plan-only posture: they describe the scope that v24 must implement, they do not claim runtime or wire work is already done, and each seed explicitly resolves the coverage gaps surfaced in the Phase 0 architecture audit (`G11`).

## Architecture coverage audit closure (ST4)

- The Phase 0 audit (`docs/v2.3/phase0/p0-architecture-coverage-audit.md`) flagged two categories as unknown: relayed cache metadata (search hints, temporary coverage deltas) and assisted-search hints for future opt-in modes. `F24-A` and `F24-C` directly close those gaps by prescribing deterministic persistence models and privacy guardrails. Completing the audit with these seeds lets `G11` declare no remaining unknown persistence classes.
- `F24-B` and `F24-D` refine device and operator behaviors that otherwise risk smuggling additional cache state into relays or clients, aligning the backlog with the guarded guarantees already exercised by `G8` and `G9`.

## Seed catalog

1. **F24-A: Distributed encrypted blob store for attachments, avatars, and emojis**
   - Authoritative persistence: Archivist blob pipeline plus optional third-party cache helpers.
   - Guarantees: ciphertext-only storage, content hashes signed by the Archivist, per-space quotas, and retention-tier defaults that mirror the history segments audited in P0-T2.
   - Gate references: closes data-class unknowns for `attachments/blobs` and `custom emojis/stickers`, feeding `G11` and upstream `G8` relay boundary checks by keeping relays stateless.
   - Proto scope: reuse the existing blob metadata messages; there is no additive delta because the metadata already supports coverage-level labels.

2. **F24-B: Multi-device history policy and cache policy guidance**
   - Authoritative persistence: device-level metadata stored in Archivist index shards, reclaimed by automated pruning.
   - Guarantees: per-device backfill limits, explicit cache validity windows, and “bring-your-own-device cache” labels so relays/operators know when a client can reuse history.
   - Gate references: strengthens `G8` relay regression checks and prepares `G9` as-built reports by locking down per-device behavior before operators sign off.

3. **F24-C: Assisted search research pack (opt-in only)**
   - Authoritative persistence: optional hint ledger and leakage budget table that specify how assisted search metadata can persist without leaking keywords.
   - Guarantees: sealed envelopes for hints with explicit budget contracts, plus audit hooks so `G11` can verify no hidden persistence drifts into assisted search workflows.
   - Gate references: ties to `G10` for the seed catalog and ensures `G11` stays closed by never leaving assisted search hints as “unknown.”

4. **F24-D: Desktop background mode daemon for harmolyn⇄xorein (optional)**
   - Authoritative persistence: local service state limited to OS service metadata and refresh tokens; no new protobuf messages are required.
   - Guarantees: OS service metadata is encrypted on disk, automatically pruned, and recorded in the deferral register until `v24`+ proves the daemon can run without violating `G8` relay rules.
   - Gate references: identifies future `G8` regression work and points to `G10`/`G11` for the audit of any local state that may be introduced.

By publishing this catalog, we satisfy `P4-T2 ST1` and make the seeds discoverable for downstream implementers, while explicitly connecting each seed to the missing persistence classes from the architecture coverage audit.
