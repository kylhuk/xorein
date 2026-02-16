# v0.2 Phase 5 - P5 Presence and Status Contract

> Status: Execution artifact. Presence state and custom status contracts live in `pkg/v02/presence/schema.go` with dedicated tests.

## Purpose

Define the planning contract for presence state semantics, dissemination/staleness rules, and custom status behavior.

## Task Coverage

| Task | Coverage summary | Evidence anchor |
|---|---|---|
| P5-T1 | Canonical presence state machine and additive-safe schema/capability flags. | `pkg/v02/presence/schema.go`, `pkg/v02/presence/schema_test.go` |
| P5-T2 | Presence publication cadence, debounce/throttle, stale detection and convergence behavior. | `pkg/v02/presence/schema.go`, `pkg/v02/presence/schema_test.go` |
| P5-T3 | Custom status schema bounds, propagation, invalidation, and replacement semantics. | `pkg/v02/presence/schema.go`, `pkg/v02/presence/schema_test.go` |

## Planning Constraints

- Invisible-state semantics must avoid unintended activity leakage.
- Update cadence must prevent high-frequency flap behavior.
- Status payload constraints remain bounded and enforceable.
