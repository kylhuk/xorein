# Phase 3 – P3 Assets Contract

## Overview
P3 assets work scopes deterministic state transitions for preview/download, degraded placeholders, and telemetry-safe logging for BlobRef-backed assets that appear in the messaging surface. The model intentionally keeps plaintext bytes out of renderer state and telemetry while enumerating failure reasons and placeholders that downstream renderers must surface.

## ST1 – Attachment rendering contract
- Attachments tilt through BlobRef pointers and enter the render plan via `pkg/v25/assets.PlanPreview`/`PlanDownload`.
- Preview and download plans produce deterministic states (`preview`, `download`, `degraded`) that the harmolyn renderer surfaces without guessing at payloads.
- Evidence: `go test ./pkg/v25/assets` proves the state map and placeholder builders; gate: G6 (harmolyn asset UX) + G7 (Podman asset playback). Command hint: run the test before capturing `EV-v25-G6-###`/`EV-v25-G7-###` outputs.

## ST2 – Deterministic degraded placeholders
- `PlanDegraded` enumerates reason codes (missing blob, rate-limited, network timeout) and maps each to a stable placeholder ID.
- Without fetching the blob, the renderer falls back to the placeholder and reason codes, guaranteeing the same UI string every time for a given failure class.
- Evidence: `go test ./tests/e2e/v25/assets_flow_test.go`, covering the degraded flow and placeholder expectations; gate: G6/G7. Command hint: capture the command output as `EV-v25-G6-###` and `EV-v25-G7-###` evidence.

## ST3 – No plaintext blob leakage
- The telemetry builder (`TelemetryFields`) omits plaintext payloads, blob bytes, and content strings; it only exposes metadata-safe fields (`asset.kind`, `asset.state`, `asset.action`, `asset.reason`, `asset.placeholder`).
- Asset telemetry logging must consume the map produced by this helper to avoid accidental plaintext leaks.
- Evidence: re-run `go test ./tests/e2e/v25/assets_flow_test.go` to prove the telemetry guard; gate: G6/G7.

## Evidence table
| ST | Test | Gate | Command |
| --- | --- | --- | --- |
| ST1 | `pkg/v25/assets` unit tests | G6/G7 | `go test ./pkg/v25/assets` |
| ST2 | `tests/e2e/v25/assets_flow_test.go` degraded flow checks | G6/G7 | `go test ./tests/e2e/v25/assets_flow_test.go` |
| ST3 | `tests/e2e/v25/assets_flow_test.go` telemetry leak guard | G6/G7 | `go test ./tests/e2e/v25/assets_flow_test.go` |
