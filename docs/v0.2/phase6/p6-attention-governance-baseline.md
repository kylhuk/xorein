# v0.2 Phase 6 - P6 Attention and Governance Baseline

> Status: Execution artifact. Notification, mention, RBAC, and governance contracts live in `pkg/v02/notify`, `pkg/v02/policy`, `pkg/v02/rbac`, and `pkg/v02/governance` with tests.

## Purpose

Capture planning contracts for notifications, mentions, baseline RBAC, moderation events, and slow-mode enforcement.

## Task Coverage

| Task | Coverage summary | Evidence anchor |
|---|---|---|
| P6-T1 | Notification event taxonomy and unread increment/reset rules across contexts. | `pkg/v02/notify/contracts.go`, `pkg/v02/notify/contracts_test.go` |
| P6-T2 | Mention parser/tokenization/resolution with deterministic fallback behavior. | `pkg/v02/policy/policy.go`, `pkg/v02/policy/policy_test.go` |
| P6-T3 | Mention authorization policy and in-app notification/badge/highlight surface contract. | `pkg/v02/notify/contracts.go`, `pkg/v02/policy/policy.go`, `pkg/v02/notify/contracts_test.go`, `pkg/v02/policy/policy_test.go` |
| P6-T4 | Baseline RBAC role hierarchy and conflict-resolution behavior. | `pkg/v02/rbac/rbac.go`, `pkg/v02/rbac/rbac_test.go` |
| P6-T5 | Signed moderation events (redaction/timeout/ban) and deterministic slow-mode semantics. | `pkg/v02/governance/metadata.go`, `pkg/v02/governance/metadata_test.go` |

## Planning Constraints

- Mass-mention controls must follow baseline v0.2 role model only.
- Moderation enforcement semantics are explicit and auditable.
- Slow-mode behavior remains deterministic under replay/reconnect conditions.
