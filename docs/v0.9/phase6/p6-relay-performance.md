# Phase 6 · Relay Performance & Load Contracts

## Plan vs implementation
Relay capacity envelopes, overload prioritization, and recovery posture are implemented inside `pkg/v09/relay/policy.go`. This document keeps the deterministic descriptions and governance notes that gate reviewers rely on for `V9-G6` validation.

## Evidence anchors
| VA ID | Artifact | Files | Description |
|---|---|---|---|
| VA-R1 | Capacity envelope | `pkg/v09/relay/policy.go` | `CapacityEnvelope` tracks limits vs utilization without claiming network runtime behavior. |
| VA-R2 | Overload priority | `pkg/v09/relay/policy.go` | `OverloadPolicy` enumerates the priority ordering: control > active media > store-forward > bulk sync. |
| VA-R3 | Recovery posture | `pkg/v09/relay/policy.go` | `RecoveryPlan` outlines deterministic next steps when thresholds breach. |
| VA-R4/VA-R7 | Operator continuity | `docs/v0.9/phase6/p6-relay-performance.md` | Notes multi-provider wake failover and governance triggers for relay operators. |

## Planned-vs-implemented language
- Planned: additional telemetry instrumentation will tie into the capacity envelope when the runtime includes the actual relay store. |
- Implemented: the helpers already return definitive decisions that the `v09-forge` scenario exercises, so gate reviewers can confirm lane priority and recovery posture without running live traffic.
