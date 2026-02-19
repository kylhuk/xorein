# Phase 3 – Relay Rollback Drill (Planning)

## Overview
Relay rollback drills exercise the planned response when an upgrade or configuration change requires returning to a known-good binary without ever entering a no-limbo state. The scenario definitions remain tentative until the drill runs, ensuring we keep the language planning-only.

## ST1 – Rollback scenario definitions and deterministic expectations
- Scenario: `relay-rollback-config-drift` simulates a config validation failure that should trigger `REASON-ROLLBACK-INIT`, gracefully evict new routes, and shift traffic to the predecessor with `REASON-ROLLBACK-COMPLETE`. The planned outcome is deterministic: the script ends with a single stable reason per node and never leaves health or routing status unresolved.
- Scenario: `relay-rollback-binary-corrupt` models a failed upgrade whose checksum mismatch raises `REASON-ROLLBACK-VERIFY`. Clients and dashboards should only see the final reason (`REASON-ROLLBACK-COMPLETE` or `REASON-ROLLBACK-ABORT`), preventing floating or ambiguous states.
- Evidence capture for this scenario maps to `EV-v26-G5-003` once the drill script records the reason sequence.

## ST2 – Drill choreography and no-limbo behavior
- Plan to orchestrate the drill so that the rollback script updates every node in lockstep, sprays the same reason taxonomy values (`REASON-UPGRADE-ABORT`, `REASON-ROLLBACK-INIT`, `REASON-ROLLBACK-VERIFY`) across monitoring, and never leaves a node with a `REASON` placeholder alone; each status change resolves immediately to the next `REASON`, satisfying the no-limbo requirement.
- The expected deterministic outcome is a fully restored predecessor release with health checks passing under `REASON-ROLLBACK-COMPLETE`, plus documentation of the reasons that drove each step for audit.
- Evidence capture for the choreography maps to `EV-v26-G5-004` once the playbook artifacts and command outputs exist.

## Command checklist
| Command | Purpose | Status | Notes |
| --- | --- | --- | --- |
| `scripts/v26-relay-rollback-scenario.sh` | Drive scenario scripts for rollback and reason emission | TODO | Tie to `EV-v26-G5-003` once run. |
| `scripts/v26-relay-rollback-verify.sh` | Confirm nodes reach `REASON-ROLLBACK-COMPLETE` | TODO | Document verification steps in the drill log. |

## Evidence table
| ST | Artifact | Gate | EV identifier | Notes |
| --- | --- | --- | --- | --- |
| ST1 | Rollback scenario playbook (config drift + binary corrupt) | G5 | EV-v26-G5-003 | Plans list deterministic reason outcomes and scenario details. |
| ST2 | Drill choreography log and verification checklist | G5 | EV-v26-G5-004 | Shows no-limbo transitions plus final reason taxonomy coverage. |
