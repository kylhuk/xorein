# F25 Blob Store Specification

This document captures the ciphertext blob/asset distribution plane that `F25` will publish, preserving v24 planning-only status and deferring implementation details to v25 while aligning with Phase 4 scope.

## ST1 – Supported Data Classes

- **Attachments**: canonical files attached to messages (documents, archives, binaries) carried as opaque ciphertext with metadata and streaming chunking hints.
- **Avatars**: user and Space avatars treated as deterministic, replaceable blobs with enforced size/mime constraints.
- **Custom emoji/stickers**: payloads incorporate creator attribution + optional delivery metadata; hosted assets must remain separable from message bodies.
- **Optional pinned media/thumbs**: when present, only ciphertext references are published (no plaintext thumbnails) to limit exposure in shared timelines.

Each data class documents the expected usage, lifecycle (create/rotate/archive), and any special encryption or expiry semantics so clients can align uploads/downloads before implementation.

## ST2 – BlobRef Model

`BlobRef` defines the canonical pointer that consumers, Archivist proxies, and repositories exchange:

- **content hash** (`BlobHash`): versioned digest (e.g., SHA-256) used for deduplication and integrity.
- **size** (`BlobSize`): exact byte count of the ciphertext payload (pre-chunking) for quota accounting.
- **mime** (`ContentType`): RFC 2046 value describing the original media type.
- **chunking parameters** (`ChunkSize`, `ChunkCount`, `ChunkHash[]`): optional per-chunk integrity for resumable uploads/downloads.
- **encryption envelope** (`EnvelopeVersion`, `EncryptionKeyID`, `KeyWrapMetadata`): describes how the ciphertext is derived from the workspace key material.
- **access control** (`SpaceID`, `Visibility`, `AccessPolicyRef`): determines whether the blob is public, Space-private, or limited to a capabilities list.

BlobRefs remain immutable once published; updates to the referenced ciphertext require new `BlobRef` instances and a revocation/expiry annotation in the metadata.

## ST3 – Storage Provider Capability Model

`F25` reuses the **Archivist** capability where feasible, exposing:

- `Archivist.BlobStore` as an extensible provider interface that declares supported classes, per-blob quotas, and eviction policies.
- Capability descriptors include `canRelay` (for cross-Space upload proxies) and `supportsChunkedUpload`.
- In environments without Archivist opt-in, a dedicated provider descriptor advertises the same fields plus `namespaceScoped` and `publisherSigned` requirements to keep compatibility additive.

The capability negotiation remains additive only: new provider hints or classes gain new optional fields without renumbering existing ones.

## ST4 – Retrieval and Anti-Enumeration

For private Spaces, retrieval must run through authenticated API endpoints that leak only the `BlobRef` and not enumeration state:

- Clients present a `SpaceID` + session token; servers verify membership before returning a `BlobRef` list.
- Anti-enumeration relies on capability-limited pagination (opaque cursors) and rate-limiting per token.
- Download endpoints require a nonce-based signed URL or short-lived session tied to the requesting client to prevent replay.
- Archivist proxies mirror the same authorization checks and must respect the `Visibility` flags stored in `BlobRef.access control`.

## ST5 – Acceptance Matrix

Performance, quota, and security concessions are captured in `docs/v2.4/phase4/f25-acceptance-matrix.md`. `F25` planning only defines the target metrics (upload throughput, storage quotas, anti-enumeration latency) so implementers can produce evidence in v25 without claiming runtime completion today.
