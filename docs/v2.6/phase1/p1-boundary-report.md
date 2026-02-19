# P1 boundary regression report

This planning snapshot catalogues the ST1–ST3 boundary probes that feed gate `G8` and clearly signals that the v2.6 release closure is still in planning mode; it does not imply downstream consumers may treat this as `G8` complete.

## ST coverage

- **ST1 – Relay no-long-history-hosting:** `TestRelayBoundaryRejectsHistoryHosting` exercises a relay-history probe that always rejects `fetchHistory`/`searchHistory` so operators can keep durable transcripts out of relay nodes. Evidence: `EV-v26-G8-001` (`go test ./tests/e2e/v26/boundaries/...` covering `TestRelayBoundaryRejectsHistoryHosting`).
- **ST2 – Relay no-durable-blob-hosting + deterministic refusal reasons:** `TestRelayBoundaryRefusesBlobUploads`, `TestRelayBoundaryMetadataOnlyPointers`, and `TestRelayBoundaryRefusalReasonTaxonomy` keep blob endpoints metadata-only, ensure payload bytes never persist, and assert the refusal reason string `relay_durable_blob_hosting_not_allowed` remains stable. Evidence: `EV-v26-G8-002` (`go test ./tests/e2e/v26/boundaries/...` covering the blob-boundary tests and refusal-reason taxonomy assertion).
- **ST3 – Private Space anti-enumeration:** `TestPrivateSpaceAntiEnumeration` verifies unauthorized lookups (present or missing blobs) all return the same `not-found` signal while membership holders can still read history. Evidence: `EV-v26-G8-003` (`go test ./tests/e2e/v26/boundaries/...` covering the private-space probe).

## Evidence ledger

| EV ID | Command | Notes |
| --- | --- | --- |
| EV-v26-G8-001 | `go test ./tests/e2e/v26/boundaries/...` | Relay-history probe (`TestRelayBoundaryRejectsHistoryHosting`) rejects both fetch and search operations.
| EV-v26-G8-002 | `go test ./tests/e2e/v26/boundaries/...` | Blob-boundary helpers remain metadata-only and keep the refusal reason `relay_durable_blob_hosting_not_allowed` deterministic via `TestRelayBoundaryRefusalReasonTaxonomy`.
| EV-v26-G8-003 | `go test ./tests/e2e/v26/boundaries/...` | Private-space anti-enumeration coverage via `TestPrivateSpaceAntiEnumeration`.

## Next steps

- Re-run the boundary suite after any relay/client or blobplane change that might affect metadata persistence, refusal reasons, or private-space membership semantics so the updated evidence files can replace the placeholders above.
