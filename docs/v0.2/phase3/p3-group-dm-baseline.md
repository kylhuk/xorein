# v0.2 Phase 3 - P3 Group DM Baseline

> Status: Execution artifact. Group DM lifecycle, key distribution, and limits are captured in `pkg/v02/groupdm/contracts.go` and `pkg/v02/groupdm/contracts_test.go`.

## Purpose

Capture planning contracts for Group DM lifecycle, key distribution/rekey rules, and transport/limit behavior for v0.2.

## Task Coverage

| Task | Coverage summary | Evidence anchor |
|---|---|---|
| P3-T1 | Group DM create/invite/join/leave/remove state machine and ordering/conflict rules. | `pkg/v02/groupdm/contracts.go`, `pkg/v02/groupdm/contracts_test.go` |
| P3-T2 | Group key distribution profile and mandatory rekey triggers for membership/compromise events. | `pkg/v02/groupdm/contracts.go`, `pkg/v02/groupdm/contracts_test.go` |
| P3-T3 | Group DM routing/history bounds, 50-member cap enforcement, convert-to-server growth UX contract. | `pkg/v02/groupdm/contracts.go`, `pkg/v02/groupdm/contracts_test.go` |

## Planning Constraints

- Group DM remains distinct from server/channel architecture in v0.2.
- Cap enforcement is deterministic with explicit reason classes.
- History continuity behavior is explicit and must not imply deferred v0.7+ migration capabilities.
