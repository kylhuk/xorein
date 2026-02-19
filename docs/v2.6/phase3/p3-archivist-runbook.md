# Phase 3 – Archivist Runbook (Planning)

## Overview
Architectural intent for the archivist/blob operator is captured here so that every rendezvous with storage, quota, and alert tooling is deterministic, planning-only, and ready for `G5`’s evidence requirements. Each outcome remains tied to the shared reason taxonomy, the no-limbo expectation, and the storage alarms that must fire before any human/automation action.

## ST1 – Storage alarms, quota handling, and deterministic telemetry
- Plan to surface storage alarms from every provider (object store, block store, retention tier) with deterministic reason codes (`REASON-ALARM-STORAGE-THRESHOLD`, `REASON-ALARM-QUOTA-EXCEEDED`, `REASON-ALARM-RECOVERY-ACK`) so dashboards/readout systems always settle on a single, documented reason before any remediation proceeds.
- Quota handling paths are captured as deterministic flows: when usage crosses configured thresholds the operator receives an explicit `REASON-QUOTA-METRIC` event, rate limits tie to documented margins, and no-limbo behavior ensures ingestion is paused before usage reports flip from `REASON-QUOTA-WARN` to `REASON-QUOTA-STOP`.
- Evidence capture will map to `EV-v26-G5-003` once the alarm/quota checklist and telemetry CLI references exist, along with annotated examples of the reason taxonomy in the plan.

## ST2 – Recovery flows, deterministic outcomes, and reason taxonomy consistency
- Recovery plans describe upgrade/rollback flows, replica reconciliation, and blob repair actions that all emit predictable reason codes (`REASON-RECOVERY-START`, `REASON-RECOVERY-VERIFY`, `REASON-RECOVERY-COMPLETE`) so automation never leaves the pipeline in a soft-failure or unknown state.
- No-limbo behavior is enforced by ensuring every recovery step concludes with a stable outcome; for example, replica resync either reports `REASON-SYNC-COMPLETE` or `REASON-SYNC-ABORT` with a declared next action, never a `REASON-UNKNOWN` placeholder.
- Cross-links to `ST1` ensure that storage alarms trigger recovery flows and throttle ingestion in the same deterministic taxonomy, so the reason field remains consistent across the lifecycle.
- Evidence capture will map to `EV-v26-G5-004` once the recovery flow documentation and reason taxonomy table are finalized.

## Command checklist
| Command | Purpose | Status | Notes |
| --- | --- | --- | --- |
| `scripts/v26-archivist-storage-alarm-check.sh` | Dry-run of storage alarm firing logic | TODO | Tie to `EV-v26-G5-003` when executed. |
| `scripts/v26-archivist-quota-tuning.sh` | Validate quota thresholds, notification reason codes, and pause/resume behavior | TODO | Capture the planned threshold values and reason taxonomy references for `EV-v26-G5-003`. |
| `scripts/v26-archivist-recovery-flow.sh` | Outline recovery flow steps (repair, resync, restart) with deterministic reasons | TODO | Attach output to `EV-v26-G5-004` once recovery plan is codified. |

## Evidence table
| ST | Artifact | Gate | EV identifier | Notes |
| --- | --- | --- | --- | --- |
| ST1 | Alarm/quota expectation doc + CLI output plan | G5 | EV-v26-G5-003 | Captures storage alarm triggers, quota thresholds, and reason taxonomy to keep states deterministic. |
| ST2 | Recovery flow playbook + reason taxonomy matrix | G5 | EV-v26-G5-004 | Ensures recovery steps report deterministic outcomes with no-limbo milestones. |
