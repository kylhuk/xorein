# Phase 0 Scope Lock (G0)

## ST1 – Imported v24 F25 acceptance matrix
- Source: `docs/v2.4/phase4/f25-acceptance-matrix.md`.
- Extracted go/no-go checks:
  | Requirement | Go? | Notes |
  | --- | --- | --- |
  | BlobRef manifest hash/size/mime integrity | Go | Aligns with v24 manifest definition; required.
  | Provider refusal reasons (quota/retention/unsupported mime/size) | Go | Scoped to deterministic refusal codes.
  | Client anti-enumeration validation | Go | Private Space membership gating is nondiscretionary.
  | Tokenized economics or public CDN hosting | No-go | Deferred per v24 decision; no new incentives or plaintext hosting.

## ST3 – MIME types and size tiers (frozen)
- Desktop tier: images (`image/png`, `image/jpeg`, `image/webp`), video (`video/mp4`, `video/webm`), audio (`audio/mp4`, `audio/opus`), archives (`application/zip`); max size **250 MiB**.
- Mobile tier: subset of desktop (image/video audio) with max size **64 MiB** and default Wi-Fi-only download prompts.
- Additions require updated spec and traceability approval.

## Locked non-goals
- Public CDN plaintext hosting (already marked No-go in v24).
- Relay nodes storing durable blob payloads (protocol invariant).
- Remote keyword search over blobs without membership context.

## Evidence placeholders
- `EV-v25-G0-001`: scope/go-no-go table confirmation.
- `EV-v25-G0-002`: size/mime freeze announcement.
- `EV-v25-G0-003`: gate sign-off sheet entry.
