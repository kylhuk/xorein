# Phase 3 – P3 Bridge Asset Policy

## Overview
The bridge asset policy keeps bots and bridge pathways strictly metadata-only so that the protocol never forwards raw blob bytes outside of the encrypted provider paths. It codifies capability gating, deterministic refusal reasons, and user-safe messages so that G6/G10 reviewers can expect reproducible, observable behavior for every denied request.

## ST1 – Metadata-only tokens for bridge/bot flows
- Bridge asset requests must ride the `TransferModeMetadata` path; `ForwardRawBytes` is always false and the policy refuses any `TransferModeRaw` attempts with a deterministic refusal reason.
- `pkg/v25/bridge/policy.go` contains the helper that enforces the metadata-only path and maps denial cases to stable reason/message pairs.
- Evidence: `go test ./tests/e2e/v25/bridge_asset_policy_test.go` (subtest `TestBridgeAssetMetadataOnlySuccess`). Gate: G6 (harmolyn asset UX) + G10 (bridge/regression). Command hint: run the test before capturing `EV-v25-G6-###`/`EV-v25-G10-###` outputs.

## ST2 – Provider capability allow/deny policy
- Providers advertise `SupportsAssetBridges`, and bridge requests check this capability before forwarding metadata tokens.
- The policy refuses unsupported providers with `ReasonProviderBridgeUnsupported` and a safe message so operators can track capability alignment deterministically.
- Evidence: same test file covering `TestBridgeAssetProviderUnsupportedRefused`. Gate: G6/G10. Command: `go test ./tests/e2e/v25/bridge_asset_policy_test.go`.

## ST3 – Deterministic refusal reasons and user-safe messages
- Every denial path (`raw blob`, `provider unsupported`, `bot denied`) uses a deterministic reason constant (`pkg/v25/bridge/refusal*`) and a friendly message exported for UI surfaces.
- Refusal messages are short, non-sensitive, and repeatable so UIs can present the same string for a given refusal code without guessing about the underlying condition.
- Evidence: `TestBridgeAssetRawBlobRefused` and `TestBridgeAssetBotCapabilityDeniedRefused`. Gate: G6/G10. Command: `go test ./tests/e2e/v25/bridge_asset_policy_test.go`.

## Evidence table
| ST | Test | Gate | Command |
| --- | --- | --- | --- |
| ST1 | `TestBridgeAssetMetadataOnlySuccess` (metadata-only success check) | G6/G10 | `go test ./tests/e2e/v25/bridge_asset_policy_test.go` |
| ST2 | `TestBridgeAssetProviderUnsupportedRefused` (provider capability refusal) | G6/G10 | `go test ./tests/e2e/v25/bridge_asset_policy_test.go` |
| ST3 | `TestBridgeAssetRawBlobRefused`, `TestBridgeAssetBotCapabilityDeniedRefused` (reason/message mapping) | G6/G10 | `go test ./tests/e2e/v25/bridge_asset_policy_test.go` |
