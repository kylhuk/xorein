# Phase 1 P1-T1 BlobRef schema and manifest semantics

This artifact captures the as-built contract for `BlobRef` and manifest handling in support of gate `G2`. It maps the detailed scope items to acceptance evidence and highlights the deterministic validation/extension rules required by ST1–ST3.

## ST1: BlobRef schema (deterministic fields)
- **hash algorithm**: normalized text (BLAKE3, SHA-256) tracked by `BlobRef.HashAlgorithm` and refused via `RefusalCodeUnsupportedAlgorithm` when unknown.
- **content hash**: opaque digest stored in `BlobRef.ContentHash` and verified for presence via `RefusalCodeMissingField`.
- **size**: integer payload length (`BlobRef.Size`) with non-negative enforcement (`RefusalCodeInvalidSize`).
- **mime type**: canonical mime string (`BlobRef.MimeType`) required by `ValidateBlobRef`.
- **chunk size/profile**: size (`BlobRef.ChunkSize`) and profile (`BlobRef.ChunkProfile`) fields that drive chunking semantics and manifest validation.
- **optional encrypted metadata pointer**: `BlobRef.EncryptedMetadataPointer` stores ciphertext handles; zero-length pointers are rejected with `RefusalCodeInvalidMetadataPointer`.
- **extension metadata**: `BlobRef.Metadata` allows forward-compatible keys without invalidating existing validators.

## ST2: Manifest validation and refusal taxonomy
- Chunk correctness is enforced via `ValidateManifest`:
  - Each chunk must have a hash and positive size, or `RefusalCodeInvalidChunk` is raised.
  - No chunk may exceed the declared manifest `ChunkSize` (`RefusalCodeChunkSizeMismatch`).
  - The manifest sum must match `TotalSize` or `RefusalCodeChunkSumMismatch` occurs.
  - Manifest fields `ContentHash`, `TotalSize`, `ChunkSize`, and `ChunkProfile` are all required; missing values result in appropriate `RefusalCodeMissingField` or `RefusalCodeInvalidChunkSize` errors.
- Manifest validation always returns a `*RefusalError` so callers can log deterministic refusal reasons for `G2` evidence.

## ST3: Forward-extensible metadata
- `MetadataExtensions` is a JSON map that round-trips unknown keys without breaking validation.
- Helpers `MetadataExtensions.Set`/`Get` marshal and unmarshal payloads while ignoring unknown extension keys during manifest validation, satisfying additive wire rules.
- Both `BlobRef.Metadata` and `Manifest.Metadata` can safely carry unknown or future keys.

## G2 mapping
| Scope item | Acceptance | Evidence command |
| --- | --- | --- |
| ST1 schema fields | `ValidateBlobRef` enforces each deterministic field documented above | `go test ./pkg/v25/blobref` |
| ST2 manifest refusal taxonomy | `ValidateManifest` rejects malformed manifests with `RefusalError` codes | `go test ./pkg/v25/blobref` |
| ST3 metadata extensions | `MetadataExtensions` helpers round-trip unknown keys and are exercised by manifest tests | `go test ./tests/e2e/v25/blobref_manifest_test.go` |

## Evidence commands
- `go test ./pkg/v25/blobref` (unit coverage for schema, manifest, refusal codes, metadata helpers)
- `go test ./tests/e2e/v25/blobref_manifest_test.go` (early-phase e2e evidence harness demonstrating manifest validation code paths)
