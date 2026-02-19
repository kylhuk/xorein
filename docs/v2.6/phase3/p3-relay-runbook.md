# Phase 3 – Relay Runbook (Planning)

## Overview
P3 relay operator work describes the planned deployment checks, alert templates, and automation steps that prove `G5` operator readiness. The runbook stays strictly planning-only: every heading captures the intent, deterministic outcomes, and reason taxonomy required before any drill is executed.

## ST1 – Deployment checks, health, and deterministic telemetry
- Plan to surface configuration and healthcheck status through the relay CLI/service so that every change emits a deterministic reason code from the shared taxonomy (`REASON-HEALTH-PASS`, `REASON-HEALTH-FAIL`, `REASON-ALERT-SKIPPED`) and never allows a no-limbo state—health transitions must always land on an explicit reason (up/down/degraded) before telemetry consumers react.
- No-limbo expectation: operators expect that clients and dashboards read a single stable reason per state change, avoiding intermediate or orphaned statuses, and that every `REASON-HEALTH-FAIL` is paired with an actionable `REASON` message describing the required next action (scale, config, restart).
- Evidence capture will map to `EV-v26-G5-001` once the healthcheck commands run.

## ST2 – Upgrade and rollback drill readiness
- Plan drills that stitch relay upgrade and rollback flows into a deterministic outcome: upgrade succeeds only if the new control plane reports `REASON-UPGRADE-READY`, otherwise automation halts with `REASON-UPGRADE-ABORT` and leaves traffic on the warmed predecessor to avoid no-limbo traffic loss.
- Rollbacks revert to the prior binary using the same reason taxonomy (`REASON-ROLLBACK-INIT`, `REASON-ROLLBACK-COMPLETE`, `REASON-ROLLBACK-VERIFY`) so monitoring always prints a predictable message and clients never see a floating state; drills are scoped to repeatable scripts with baked-in success/failure outcomes.
- Evidence capture will map to `EV-v26-G5-002` once the drill sequences are recorded.

## Command checklist
| Command | Purpose | Status | Notes |
| --- | --- | --- | --- |
| `scripts/v26-relay-healthcheck.sh` | Validate deterministic health reasons | TODO | Capture output for `EV-v26-G5-001`. |
| `scripts/v26-relay-upgrade-drill.sh` | Step through upgrade + rollback automation | TODO | Tie to `EV-v26-G5-002` once run. |
| `scripts/v26-relay-alert-check.sh` | Ensure alerts fire with reason taxonomy | TODO | Document expected alerts in runbook. |

## Evidence table
| ST | Artifact | Gate | EV identifier | Notes |
| --- | --- | --- | --- | --- |
| ST1 | Relay healthcheck doc + CLI output plan | G5 | EV-v26-G5-001 | Plans describe deterministic reason outputs and no-limbo states. |
| ST2 | Upgrade/rollback drill playbook | G5 | EV-v26-G5-002 | Specifies reason taxonomy and expected drill outcomes for rollback guardrails. |
