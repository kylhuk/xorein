# Phase 1 Store Schema (v2.1)

- **Spaces/Channels**: Tables capture Space ID, channel ID, membership caches, and per-channel metadata such as last hydration cursor and search coverage ranges.
- **Timeline rows**: Each row records MessageID/EventID, channel ID, sender ID, delivery state, ciphertext reference, plaintext payload (decrypted on insert), attachments metadata, and tombstone markers.
- **Tombstones/Redactions**: Schemas track reason codes, audit pointers (e.g., redaction issuer, timestamp), and preserve minimal metadata for timeline ordering.
- **Versioning**: The schema is versioned so migrations can detect mismatches and emit `STORE_MIGRATION_REQUIRED` before any write occurs.

Every schema change is accompanied by deterministic failure reason metadata in the store contract so clients can react predictably when migrations or corruption are detected.

## Evidence cross-reference
- Gate-level verification for the store schema is tracked in the phase 5 evidence index (`docs/v2.1/phase5/p5-evidence-index.md`), where `EV-v21-G1-001`, `EV-v21-G1-002`, and `EV-v21-G2-001` tie to the `buf` commands and store regression suite referenced above.
