# P2-T2 Journey migration to local API

This artifact records the Phase 2 migration work where harmolyn journeys were moved behind a local API coordinator, satisfying ST1–ST3 without touching the relay or database surfaces.

## ST1 – Local API-only journeys
- Every outbound journey action (`Send`, `Read`, `Search`, `FetchHistory`, `MediaControl`) now lives in `pkg/v24/harmolyn/journeys` and can *only* reach the daemon through the exported `LocalAPIClient` interface.
- The new `journeys.Coordinator` inserts attach gating before every call so no code path can invoke relay/database APIs directly, which keeps the implementation on a deterministic local API path.

## ST2 – Deterministic degraded-state reasons
- Degraded outcomes return a typed `JourneyError` that mirrors the attach failure taxonomy (`DAEMON_START_FAILED`, `DAEMON_INCOMPATIBLE`, `AUTH_FAILED`, `SOCKET_PERMISSION_DENIED`) and carries the matching `NextAction` guidance.
- Tests prove every attach failure reason is surfaced unchanged to the client and the local API is never called from a degraded path.

## ST3 – Attach-state gating and token enforcement
- The `journeys` coordinator depends on the attach provider/state interface from `pkg/v24/harmolyn/attach` so every operation requests a fresh session token before hitting the local API client.
- Any missing attach/state information now produces a deterministic `JourneyError`, which prevents bypassing the attach state machine or session token checks.

## Evidence mapping (G3 / G4)
| Gate | Command | Evidence target | Notes |
| --- | --- | --- | --- |
| G3 | `go test ./pkg/v24/harmolyn/journeys` | `EV-v24-G3-001` (journey package unit coverage) | Demonstrates ST1/ST3 paths and token gating.
| G4 | `go test ./tests/e2e/v24/journeys_*` | `EV-v24-G4-001` (journey degraded matrix) | Exercises ST2 deterministic reasons for every degraded scenario; `tests/e2e/v24/journeys_flow_test.go` covers the nominal path.

Capture the exact command outputs for these `go test` runs, link them to the EV rows above, and store the logs under `artifacts/generated/v24-evidence/` as per the evidence bundle conventions.
