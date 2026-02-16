# v0.3 Phase 2 - P2 Screen Share Baseline

> Status: Execution artifact. Screen-sharing capture, encode, simulcast, and viewer controls contracts now trace to `pkg/v03/screenshare` deliverables.

## Purpose

Detail deterministic contracts for native capture, hardware encoder selection, quality presets, simulcast layering, and viewer control behavior while preserving explicit security disclosures.

## Scope Summary

- Platform-native capture/encoder selection with deterministic per-platform capability tables.
- Quality presets (Low/Standard/High/Ultra/Auto) and simulcast up to three layers.
- Viewer controls for fullscreen, PiP, zoom, and pan transitions with render degradation behavior.

## Acceptance Anchors

1. Capture/encoder flows declare explicit security posture (E2EE vs. clear mode) and reason codes for permission denials.
2. Preset-to-parameter and simulcast layer mappings are deterministic under normal and degraded path conditions.
3. Viewer controls maintain deterministic state transitions and rendering fallbacks without implicit security mode changes.

## Evidence Mapping

| Element | Primary Evidence | Code/Test Reference |
|---|---|---|
| Native capture + encoder selection | `docs/v0.3/phase2/p2-screen-share-baseline.md` | `pkg/v03/screenshare/contracts.go`, `pkg/v03/screenshare/contracts_test.go` |
| Quality presets + simulcast | `docs/v0.3/phase2/p2-screen-share-baseline.md` | `pkg/v03/screenshare/contracts.go`, `pkg/v03/screenshare/contracts_test.go` |
| Viewer controls | `docs/v0.3/phase2/p2-screen-share-baseline.md` | `pkg/v03/screenshare/contracts.go`, `pkg/v03/screenshare/contracts_test.go` |

## Recovery Guidance Summary

- Capture restart guidance references permission/retry states documented in this file.
- Simulcast/backoff guidance ensures degraded clients fall back to fewer layers explicitly.
- Viewer control conflicts resolve via deterministic state machine entries linked to `pkg/v03/screenshare/contracts_test.go`.
