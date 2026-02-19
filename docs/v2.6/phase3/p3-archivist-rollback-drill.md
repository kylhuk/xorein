# Phase 3 – Archivist Rollback Drill (Planning)

## Overview
This drill plan stays deliberately planning-only: it defines how rollback scenarios interact with storage alarms, quota guardrails, and recovery flows without implying execution has occurred. The drill reinforces deterministic outcomes, no-limbo behavior, and a consistent reason taxonomy so `G5` evidence can cite precise signals.

## ST1 – Rollback scenario with storage alarms and quota guardrails
- Outline a rollback sequence that is gated on storage alarms (`REASON-ALARM-STORAGE-CRITICAL`, `REASON-ALARM-QUOTA-STOP`) so operators never rollback unless a documented threshold is breached; the scenario pauses new writes, snapshots the current state, and then triggers the rollback reason code `REASON-ROLLBACK-TRIGGERED`.
- Quota handling during rollback enforces deterministic throttles: the plan includes a `REASON-QUOTA-FREEZE` signal that prevents new blobs from being persisted until quota/usage metrics simultaneously report `REASON-QUOTA-NORMAL` and the rollback completes, preventing no-limbo ingestion spikes.
- The scenario ties to storage alarms so dashboards always know why the rollback started; the final state of the drill is either `REASON-ROLLBACK-COMPLETE` or `REASON-ROLLBACK-ABORT`, never an indeterminate `REASON-UNKNOWN`.

## ST2 – Recovery flows and deterministic outcomes after rollback
- Plan for recovery flows that rewind failed upgrades: once rollback completes, the playbook walks through post-recovery checks (metadata reconciliation, checksum verification, quota release) and ensures each step emits a consistent reason (`REASON-RECOVERY-VERIFY`, `REASON-RECOVERY-PASS`, `REASON-RECOVERY-ESCALATE`).
- No-limbo requirement: the drill requires a final verification step so operators see either `REASON-RECOVERY-PASS` or `REASON-RECOVERY-ESCALATE` before the drill is considered done; there are explicit next actions tied to each reason code.
- Evidence capture will map to `EV-v26-G5-005` once the rollback scenario narrative, alarm/quotas, and recovery reason taxonomy are documented.

## Command checklist
| Command | Purpose | Status | Notes |
| --- | --- | --- | --- |
| `scripts/v26-archivist-rollback-sim.sh` | Simulate rollback gated by storage alarms and quota signals | TODO | Reference `EV-v26-G5-005` when scenario is documented. |
| `scripts/v26-archivist-recovery-verify.sh` | Run recovery verification steps (checksum, quorum, quota release) | TODO | Capture deterministic reason outputs for `EV-v26-G5-005`. |

## Evidence table
| ST | Artifact | Gate | EV identifier | Notes |
| --- | --- | --- | --- | --- |
| ST1 | Rollback scenario + alarm/quotas plan | G5 | EV-v26-G5-005 | Shows storage alarms triggering rollback and ensures reason taxonomy keeps outcomes deterministic. |
| ST2 | Recovery verification playbook | G5 | EV-v26-G5-005 | Documents how recovery flows conclude with explicit no-limbo reasons. |
