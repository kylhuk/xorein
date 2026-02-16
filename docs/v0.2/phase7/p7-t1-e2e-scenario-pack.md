# v0.2 Phase 7 - P7-T1 End-to-End Scenario Validation Pack

> Status: Execution artifact. Scenario pack links v0.2 journeys to pkg/v02 contract tests.

## Purpose

Define end-to-end scenario coverage for v0.2 journeys, including degraded and recovery behavior, before V2-G5 release readiness closure.

## Source Trace

- `TODO_v02.md:783`
- `SPRINT_GUIDELINES.md:105`
- `aether-v3.md:734`

## Scenario Matrix (P7-T1-ST1)

| Scenario ID | Journey | Scope coverage | Expected outcome | Evidence anchor |
|---|---|---|---|---|
| SCN-V2-01 | New 1:1 DM bootstrap and first send | V2-S01, V2-S02, V2-S03, V2-S05 | X3DH bootstrap succeeds, first encrypted DM delivered, explicit security mode label displayed | `pkg/v02/dmratchet/contracts_test.go`, `pkg/v02/dmtransport/contracts_test.go` |
| SCN-V2-02 | Offline recipient DM delivery and retrieval | V2-S02, V2-S03 | Store-forward retrieval is idempotent with deterministic receipt handling | `pkg/v02/dmqueue/contracts_test.go` |
| SCN-V2-03 | Group DM create/invite/join/leave | V2-S04, V2-S05 | Membership transitions and rekey semantics remain deterministic | `pkg/v02/groupdm/contracts_test.go` |
| SCN-V2-04 | Group DM cap boundary and conversion guidance | V2-S04 | 50-member cap enforced with deterministic reason and guided next action | `pkg/v02/groupdm/contracts_test.go` |
| SCN-V2-05 | Friend request via key/QR/deeplink | V2-S06, V2-S08 | All entry paths normalize and converge to one friend graph state | `pkg/v02/friends/authenticity_test.go`, `pkg/v02/friends/listsync_test.go` |
| SCN-V2-06 | Presence + custom status convergence | V2-S07, V2-S08 | Presence/status propagate and stale cleanup converges deterministically | `pkg/v02/presence/schema_test.go` |
| SCN-V2-07 | Notification/unread coherence across contexts | V2-S09, V2-S10 | Counters and mention badges align across DM/Group DM/server contexts | `pkg/v02/notify/contracts_test.go`, `pkg/v02/policy/policy_test.go` |
| SCN-V2-08 | Mention parsing/resolution/authorization | V2-S10, V2-S11 | Unauthorized mass-mentions rejected with deterministic reason classes | `pkg/v02/policy/policy_test.go` |
| SCN-V2-09 | Baseline moderation actions | V2-S11, V2-S12 | Signed redaction/timeout/ban events enforce deterministically | `pkg/v02/governance/metadata_test.go` |
| SCN-V2-10 | Slow mode under reconnect/replay | V2-S11, V2-S13 | Equivalent streams yield consistent accept/reject outcomes | `pkg/v02/governance/metadata_test.go`, `pkg/v02/policy/policy_test.go` |

## Failure-Recovery Scenarios (P7-T1-ST2)

| Scenario ID | Failure class | Recovery expectation | User-visible requirement | Evidence anchor |
|---|---|---|---|---|
| SCN-V2-R01 | Prekey depletion | Inventory republish and retry path completes | No-limbo state + next action displayed | `pkg/v02/dmqueue/contracts_test.go` |
| SCN-V2-R02 | Stale presence | Stale timeout converges to deterministic offline state | Reason class and refresh action shown | `pkg/v02/presence/schema_test.go` |
| SCN-V2-R03 | Delayed offline retrieval | Retry window delivers once with dedupe | Deterministic progress/failure state surfaced | `pkg/v02/dmqueue/contracts_test.go` |

## QoL Scorecard Placeholder

Priority journey: degraded DM send and recovery.

| Metric | Baseline | Target | Result | Evidence |
|---|---|---|---|---|
| User actions to recover from degraded DM send | TBD | >=10% reduction | Pending | `docs/v0.2/phase7/p7-t1-e2e-scenario-pack.md` |
| No-limbo compliance on degraded states | TBD | 100% of covered scenarios | Pending | `docs/v0.2/phase7/p7-t1-e2e-scenario-pack.md` |
| Deterministic reason class coverage | TBD | 100% of covered scenarios | Pending | `docs/v0.2/phase7/p7-t1-e2e-scenario-pack.md` |

> Note: These metrics await quantitative validation; instrumentation for the ≥10% user-effort reduction baseline is scheduled for the release-gate execution stage, so the table remains descriptive until that data exists.
