# P0 Proto Compatibility Checklist

Planning artifact describing the additive wire commitments for the v25 blob store and asset distribution package. This document stays planning-only until the listed expectations have corresponding artifacts.

## ST1 — Expected proto surfaces
- Describe the blob transfer endpoints, chunk manifest references, and asset metadata that must land in the v25 delta:
  - `PutBlobChunk`/`GetBlobChunk` (chunk payload, offset, completion flag)
  - `PutManifest`/`GetManifest` (BlobRef, manifest hash, chunk list, size, mime)
  - `ConfirmManifest` (client-issued verification fingerprint)
  - `BlobRef` message (hash, size, mime, chunk_size, optional root_manifest_hash)
  - Asset-specific fields (avatar/icon/emoji references, access controls, Space/DM capability tags)
  - Provider capability metadata (quota buckets, retention tier, provider identifier)
  - Private-space anti-enumeration hints (membership-scoped salt, authorization tokens)
  - Replication state fields (replica listing, health summaries, repair hints)

## ST2 — Additive-only checklist
- Add fields/messages only; do not renumber existing field tags.
- Avoid changing field types, semantics, or wire encodings that older clients expect.
- Introduce only optional or `oneof`-guarded fields so no new server fields become required in older deployments.
- Reserve field numbers for any removed message/field to prevent reuse.
- Document each new tag in the compatibility register and update `docs/templates/roadmap-evidence-index.md` with the planned evidence placeholder.

## ST3 — Mixed-version downgrade/read-old expectations
- Relay and client downgrade paths must allow reading manifests/chunks produced by older versions without assuming new envelope fields.
- New clients must gracefully ignore unknown fields and continue to validate blobs with the oldest supported hashing/AEAD proofs.
- Mixed-version replication/repair flows must annotate asset metadata with the oldest-reader version so providers can choose the correct transfer path.
- Anti-enumeration controls must degrade to an explicit failure code when metadata lacks membership hints (older blobs) while still keeping the asset hidden.
- Note: evidence anchors for compatibility verification live under `EV-v25-G1-###` (e.g., `EV-v25-G1-001` for `buf lint`, `EV-v25-G1-002` for `buf breaking`) and act as placeholders until the commands execute.
