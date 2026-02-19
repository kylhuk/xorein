# Phase 3 – Auxiliary Services Runbook (Planning)

## Overview
- Document the planned operator guidance for the remaining auxiliary services (indexer, push relay, TURN) so the `G5` operator readiness gate can be signaled with deterministic reasons and no-limbo outcomes.
- Every subsection keeps planning-only language ("plan to", "ensure", "should") and traces evidence to `EV-v26-G5` IDs.
- Conditional paths differentiate between shipped and not-shipped scenarios; both must record a deterministic state and reason taxonomy before any drill executes.

## Indexer readiness

### Shipped path
- Plan for an indexer healthcheck that emits a deterministic reason from the taxonomy (`REASON-AUX-INDEXER-HEALTHY`, `REASON-AUX-INDEXER-FAIL`, `REASON-AUX-INDEXER-DEGRADED`) and ties the final outcome to the drill result so dashboards never pause in a no-limbo state.
- Failure drill: simulate `indexer down` while relay + push clients stay connected, confirm automation raises `REASON-AUX-INDEXER-FAIL`, and validate that clients switch to a fallback cache state with explicit guidance on when the indexer becomes available again; drill outcome should be recorded as `EV-v26-G5-011`.
- Deterministic outcome: either the indexer report transitions to healthy (with `REASON-AUX-INDEXER-HEALTHY`) or the drill halts with a documented mitigation (e.g., roll forward or retry cadence) so no intermediate status is exposed to diagnostics.
- Tie every update to the shared reason taxonomy so monitoring and alerting always read one reason per state change and can forward that reason to downstream operators without guessing.

### Not-shipped path
- When the indexer is not part of the v2.6 payload, plan an explicit `REASON-AUX-INDEXER-NA` declaration so dashboards and operator communications display the deterministic not-applicable policy instead of an ambiguous absence.
- Capture the statement that the service is intentionally unsupported (e.g., "indexer not shipped in v2.6") plus the fallback reason `REASON-AUX-INDEXER-NA` and map the evidence to `EV-v26-G5-NA-INDEXER`.
- Declare that no drills run in this path, but the policy still triggers the same reason taxonomy so downstream tooling knows why the service is skipped.

## Push relay readiness

### Shipped path
- Plan for the push relay health loop to surface `REASON-AUX-PUSH-HEALTHY`, `REASON-AUX-PUSH-FAIL`, and `REASON-AUX-PUSH-RATE-LIMIT` with deterministic client-facing descriptions and no shared state that could linger in a no-limbo state.
- Failure drill: bring the push relay offline while indexer/relay clients continue to operate, confirm the automation reports `REASON-AUX-PUSH-FAIL`, and validate that clients instead use polling or retransmit backoff per the documented mitigation plan; capture this drill under `EV-v26-G5-012`.
- Deterministic outcome: the drill either ends in a peaceful failover with `REASON-AUX-PUSH-FAIL` plus enforced retry policy, or the operation aborts with a clear failure notice so bridging clients do not attempt silent retries.
- Ensure telemetry and alerting consume the same taxonomy so no-limbo rule holds across both control-plane and UI monitoring surfaces.

### Not-shipped path
- If the push relay is not deployed for v2.6, publish a deterministic `REASON-AUX-PUSH-NA` note so the operator runbook shows the deliberate omission and clients do not expect the capability.
- Map this decision to `EV-v26-G5-NA-PUSH` and state the non-shipped policy in every operator briefing so failure classify steps never mix with production runs.

## TURN readiness

### Shipped path
- Deploy TURN validation that emits `REASON-AUX-TURN-ALLOC`, `REASON-AUX-TURN-FAIL`, and `REASON-AUX-TURN-SUCCESS` with deterministic timing and no-limbo semantics; drift between allocation and failure must not leave clients hanging without a reason.
- Failure drill: stop TURN service (or exhaust relay allocations) and confirm clients log `REASON-AUX-TURN-FAIL` while relays fall back to direct NAT traversal scripts; capture the drill outcome under `EV-v26-G5-013` and document the deterministic client behavior.
- Deterministic outcome: either the TURN rebuild sequence succeeds (with `REASON-AUX-TURN-SUCCESS`) or the failure path is logged with both automation mitigation steps and the assigned reason code so recovery steps remain consistent.
- Require the same reason taxonomy to feed dashboards so operators can read a single reason line per state change and avoid temporary or unmatched states.

### Not-shipped path
- When TURN is absent from the v2.6 deployment, plan to publish `REASON-AUX-TURN-NA` so the system records the deterministic not-applicable policy instead of leaving NAT traversal queries unanswered.
- Map this policy to `EV-v26-G5-NA-TURN` and note in runbooks that TURN-related failure drills are skipped for the terminal closure, including the deterministic reason that justifies the skip.

## Failure drill matrix & deterministic outcomes
- Indexer–down drill (`EV-v26-G5-011`), push relay–down drill (`EV-v26-G5-012`), and TURN–down drill (`EV-v26-G5-013`) must each produce a single final reason from the shared taxonomy, document the mitigation path, and publish deterministic telemetry so downstream automation never sits in a no-limbo digital purgatory.
- Each shipped-service drill must be paired with a deterministic outcome checklist (health reason, mitigation, client UX change) before being marked complete.
- Not-shipped policies should still publish their deterministic reasons (`REASON-AUX-*-NA`) plus the associated `EV-v26-G5-NA-*` entry so auditors can trace why the drill was not executed.

## Command checklist (placeholders)
| Command | Purpose | Status | Notes |
| --- | --- | --- | --- |
| `scripts/v26-aux-indexer-drill.sh` | Run indexer failure drill, reason verification, EV capture | TODO | Link output to `EV-v26-G5-011`. |
| `scripts/v26-aux-push-relay-drill.sh` | Run push relay failure drill with deterministic mitigation | TODO | Link output to `EV-v26-G5-012`. |
| `scripts/v26-aux-turn-drill.sh` | Run TURN failure drill + reason taxonomy confirmation | TODO | Link output to `EV-v26-G5-013`. |

## Evidence table
| Service | Scenario | Gate | EV identifier | Notes |
| --- | --- | --- | --- | --- |
| Indexer | Shipped failure drill (indexer down) | G5 | EV-v26-G5-011 | Failure reason taxonomy, mitigation, and no-limbo demand. |
| Push relay | Shipped failure drill (push relay down) | G5 | EV-v26-G5-012 | Deterministic outcome and client UX fallback. |
| TURN | Shipped failure drill (TURN down) | G5 | EV-v26-G5-013 | Allocation/fail reason taxonomy and recovery path. |
| Auxiliary | Indexer not shipped | G5 | EV-v26-G5-NA-INDEXER | Clearly stated not-applicable policy + reason. |
| Auxiliary | Push relay not shipped | G5 | EV-v26-G5-NA-PUSH | Clearly stated not-applicable policy + reason. |
| Auxiliary | TURN not shipped | G5 | EV-v26-G5-NA-TURN | Clearly stated not-applicable policy + reason. |
