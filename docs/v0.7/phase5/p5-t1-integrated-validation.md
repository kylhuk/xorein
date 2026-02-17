# Phase 5 · P5-T1 Integrated Validation Scenario Pack

## Objective
Demonstrate positive, negative, degraded, and recovery paths spanning every v0.7 bullet so V7-G5 can declare the cross-feature validation pack ready without implying runtime completion.

## Contract
- Build scenario IDs that combine store-forward retention, history sync, scoped search, push relay, and desktop notification events, and capture expected outcomes plus reason codes in `tests/e2e/v07/archive_flow_test.go` and `tests/e2e/v07/search_notification_flow_test.go`.
- For each scenario, explicitly document the evidence type (doc reference, unit test, e2e harness entry) and note the dataset or topology it exercises (`tests/e2e/v07/` or `containers/v07/`).
- Residual-risk control entries must tie back to gate-level reason classes so reviewers can confirm the high-impact failure modes are acknowledged and scheduled for follow-up if not fully resolved.

## Evidence anchors
| Artifact | Description | Evidence anchor |
|---|---|---|
| `VA-I1` | Scenario pack coverage & reason classes | Section "Contract" + `tests/e2e/v07/archive_flow_test.go`, `tests/e2e/v07/search_notification_flow_test.go` |

This document cements the integrated validation narrative for V7-G5 while referencing the planned `tests/e2e/v07` harness and container artifacts for eventual evidence.
