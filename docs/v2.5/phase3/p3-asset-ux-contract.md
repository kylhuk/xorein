# Phase 3 – P3 Asset UX Contract

## Overview
P3 asset UX work defines how harmolyn surfaces upload and download intent states, toggles, and offline markers without leaking protocol or plaintext data. The package under `pkg/v25/harmolyn/ui` provides deterministic plans for progress, cancellation, and badge presentation so every client renders the same failure class under ST1–ST3.

## ST1 – Upload/download progress and cancellation
- `PlanUploadProgress`, `PlanDownloadProgress`, and `PlanUploadCancelled` return stable stages (`uploading`, `downloading`, `cancelled`) with clamped progress percentages and explicit `Cancellable` flags that drive the harmolyn cancel affordance.
- Evidence: `go test ./pkg/v25/harmolyn/ui` proves the contract enforces the progress map and cancel transitions; gate: G6. Command hint: run the test before recording `EV-v25-G6-###` outputs.

## ST2 – Tap-to-download and Wi-Fi toggles
- `Controls` encodes the `TapToDownload` and `DownloadOnWiFiOnly` toggles in harmolyn; `PlanDownloadRequest` evaluates them against link state, emitting `tap-to-download` or `wifi-only` reasons until the user’s intent and Wi-Fi policy allow the transition to `downloading`.
- Evidence: `go test ./pkg/v25/harmolyn/ui` verifies the deterministic toggle logic and gating; gate: G6. Command hint: include the same test output for `EV-v25-G6-###`.

## ST3 – Offline badges and deterministic errors
- Every deterministic reason maps to a stable offline badge (e.g., `badge-offline`, `badge-network-error`, `badge-tap-to-download`), enabling harmolyn to paint consistent badges whenever downloads are blocked, including explicit network failure plans produced by `PlanDownloadNetworkError`.
- Evidence: `go test ./pkg/v25/harmolyn/ui` proves badge resolution and error surfaces remain deterministic; gate: G6. Command hint: run the test and attach its output to the `EV-v25-G6-###` entry.

## Evidence table
| ST | Test | Gate | Command |
| --- | --- | --- | --- |
| ST1 | `pkg/v25/harmolyn/ui` unit tests | G6 | `go test ./pkg/v25/harmolyn/ui` |
| ST2 | `pkg/v25/harmolyn/ui` unit tests | G6 | `go test ./pkg/v25/harmolyn/ui` |
| ST3 | `pkg/v25/harmolyn/ui` unit tests | G6 | `go test ./pkg/v25/harmolyn/ui` |

## Captured evidence
- `EV-v25-G6-001`: `go test ./pkg/v25/harmolyn/ui -count=1` -> pass (`artifacts/generated/v25-evidence/go-test-harmolyn-ui.txt`).
- `EV-v25-G6-002`: `go test ./tests/e2e/v25/... -run Harmolyn -count=1` -> not applicable (`[no tests to run]`) captured in the same artifact.
