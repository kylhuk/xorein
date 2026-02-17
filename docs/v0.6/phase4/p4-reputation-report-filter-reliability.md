# Phase 4 · P4 Reputation/Report/Filter Reliability

## Purpose
Document deterministic reliability semantics for reputation, report routing, and optional filters so V6-G4 can attest to anti-gaming and bounded enforcement behavior.

## Contract
- `P4-T1` (`VA-R1`) describes the anti-gaming rationale for the web-of-trust reputation system, including uncertainty signaling, reason-coded weight adjustments (`reputation.weight.success/failure/recover`), and how sparse trust graphs feed into the weighting policy mentioned in OD6-03.
- `P4-T2` (`VA-R2`) outlines report routing reliability (idempotency, retry/backoff, ordering, ack semantics, and encrypted evidence handling) with `reporting.route.*` reason classes and references to the cross-feature scenario pack for adverse flows.
- `P4-T3` (`VA-R3`) defines optional filter boundaries, SecurityMode gating, and on-device/E2EE prompts so filters remain non-authoritative, including reason classes (`filters.process.success/blocked/recover`) that describe when filters opt in or gracefully defer.

### Deterministic rule table
| Input | Outcome | Reason-class | Fallback/Recovery | Evidence anchor |
|---|---|---|---|---|
| Raw reputation weight plus trust path count | Weight capped to [0,1]; `reputation.weight.success` when inside bounds, `reputation.weight.failure` when anti-gaming cap hits | `reputation.weight` | Recovery still allows uncertainty signaling `reputation.weight.recover` by diluting weight | `pkg/v06/reputation/contracts.go#ComputeReputationWeight` |
| Report ID, previous IDs, ordering hint | Idempotency key derived from report ID; `reporting.route.accept` when ordering preserved | `reporting.route` | Duplicate/out-of-order detection sets `reporting.route.failure`; deterministic retry with preserved ordering `reporting.route.retry` | `pkg/v06/reporting/contracts.go#EvaluateReportRouting` |
| SecurityMode plus filter policy flag | `filters.process.success` when mode allows Clear server-side filters; `filters.process.blocked` when E2EE prohibits server-side enforcement | `filters.process` | Recovery chooses on-device limited enforcement path `filters.process.recover` with user prompt | `pkg/v06/filters/contracts.go#DecideFilterExecution` |
