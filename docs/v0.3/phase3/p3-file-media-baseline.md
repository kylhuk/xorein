# v0.3 Phase 3 - P3 File and Media Baseline

> Status: Execution artifact. File transfer chunking, integrity, inline preview, and fallback behaviors now align with `pkg/v03/transfer` contracts.

## Purpose

Define deterministic contracts for chunked peer-to-peer file transfers (<=25MB), inline image preview, file attachment cards, and recovery behavior across E2EE and clear modes.

## Scope Summary

- Chunked transfer with integrity checks, retry/resume, and explicit mode-aware encryption disclosure.
- Inline image rendering plus file/attachment card presentation, including metadata policies.
- Degraded/fallback handling that preserves integrity and transparency about security posture shifts.

## Acceptance Anchors

1. Chunking policy and retry/resume behavior are deterministic for success, partial failure, and resume paths with explicit security disclosure per conversation mode.
2. Inline previews and attachment cards have bounded metadata, rendering rules, and explicit failure states for unsupported content.
3. Fallback behavior never silently changes security posture; mode changes require explicit disclosure.

## Evidence Mapping

| Contract | Mentioned Document | Code/Test Evidence |
|---|---|---|
| Chunked P2P transfer | `docs/v0.3/phase3/p3-file-media-baseline.md` | `pkg/v03/transfer/contracts.go`, `pkg/v03/transfer/contracts_test.go` |
| Inline preview + attachment | `docs/v0.3/phase3/p3-file-media-baseline.md` | `pkg/v03/transfer/contracts.go`, `pkg/v03/transfer/contracts_test.go` |
| Degraded/fallback guidance | `docs/v0.3/phase3/p3-file-media-baseline.md` | `pkg/v03/transfer/contracts.go`, `pkg/v03/transfer/contracts_test.go` |

## Recovery Summary

- Resume statements refer to deterministic chunk acknowledgment and retry windows.
- Preview failure states map to explicit reason codes with user actions (reload, request upgrade, etc.) bound to this doc.
