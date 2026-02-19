# Phase 2 Relay Boundary Report

This report captures the Phase 2 P2-T3 relay boundary regression verification.

## ST coverage
- **ST1 (relay refuses durable blob payload hosting)**
  - The fake relay boundary rejects `PutManifest`/`PutBlobChunk` and returns `errRelayRefused`, so payload endpoints cannot store blobs.
- **ST2 (relay stores only metadata pointers/tokens)**
  - The boundary retains only `Manifest` pointers via `recordManifestPointer` while `hasPayload` stays false, proving no payload bytes are persisted.
- **ST3 (private-space anti-enumeration)**
  - Unauthorized queries for private blobs emit the `errNotFound` sentinel regardless of whether the blob exists, matching requests for missing blobs.

## Evidence
- `EV-v25-G10-001`: `go test ./tests/e2e/v25/relay_blob_boundary_test.go -count=1` -> pass (`artifacts/generated/v25-evidence/go-test-relay-blob-boundary.txt`)

## Notes
- The regression model is deterministic and fast; it exercises the fake boundary in isolation so that future Podman scenarios can reuse the same contract.
