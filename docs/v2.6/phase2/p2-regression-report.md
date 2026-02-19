# Phase 2 regression matrix report (v2.6)

This artifact captures the `G3` regression gate evidence for journeys that are shipped in this repository snapshot together with the deterministic reasons for anything that remains deferred or absent.

## Matrix evidence (EV-v26-G3-###)

| Journey | Status | Evidence | Notes |
| --- | --- | --- | --- |
| Identity creation + backup recovery | Shipped | `EV-v26-G3-001` (`tests/e2e/v26/regression_matrix_test.go`: `TestRegressionMatrixIdentityJourney`) | Verifies `pkg/v12/identity` deterministic records, backup export, and restore settle on an identical identity fingerprint. |
| Space join + RBAC | Shipped | `EV-v26-G3-002` (`tests/e2e/v26/regression_matrix_test.go`: `TestRegressionMatrixSpaceJourney`) | Covers invite-only defaults, member promotion, founder transfer, and join-policy invite token enforcement from `pkg/v13/spaces` + `pkg/v13/joinpolicy`. |
| Chat send/receive + read markers | Shipped | `EV-v26-G3-003` (`tests/e2e/v26/regression_matrix_test.go`: `TestRegressionMatrixChatJourney`) | Exercises `pkg/v13/chat` message lifecycle, invalid transitions, and deterministic read-marker touch semantics. |
| Media (voice negotiation + screen-share adaptation) | Shipped | `EV-v26-G3-004` (`tests/e2e/v26/regression_matrix_test.go`: `TestRegressionMatrixMediaJourney`) | Proves `pkg/v14/voice` fallback/backoff ladders plus `pkg/v15/screenshare` transport transitions, adaptation labeling, and recovery hints. |
| Attachments/assets (preview/download/degraded) | Shipped | `EV-v26-G3-005` (`tests/e2e/v26/regression_matrix_test.go`: `TestRegressionMatrixAssetJourney`) | Covers deterministic asset planning for attachments, downloads, and degraded emoji flows from `pkg/v25/assets`. |
| Bridge/bot/webhook assets | Shipped | `EV-v26-G3-006` (`tests/e2e/v26/regression_matrix_test.go`: `TestRegressionMatrixBridgeJourney`) | Validates metadata-only bridges, raw-block refusals, provider capability gating, and bot capability gating via `pkg/v25/bridge`. |
| History/backfill search coverage | Not shipped (REASON=SEARCH-BACKFILL-MISSING) | `EV-v26-G3-007` | This repo snapshot lacks the deterministic history/backfill search contract that would power `P2-T1 ST3`; the matrix documents the gap so future versions can reference a deterministic reason. |

## Testing command

The regression matrix lives entirely inside `tests/e2e/v26/regression_matrix_test.go`. `go test ./tests/e2e/v26/...` proves the new coverage and is also recorded as the `EV-v26-G3-###` evidence above.
