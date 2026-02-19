# Archivist Upgrade/Rollback Drill (G4/G5)

This drill proves the operator can upgrade Archivist to the v2.3 image and roll back safely while keeping the regression gates (G4 for operations readiness, G5 for historical regression scenarios) satisfied. Treat this document as the drill plan; add `EV-v23-G4-###` and `EV-v23-G5-###` entries per execution and mark any unrun steps as `pending` in the evidence index.

## Gate mapping
- **G4**: Operability drills (runbook follow-through and evidence logging).
- **G5**: Regression-latency verification after upgrade/rollback; ensure Podman scenario coverage remains green.

## Preconditions
1. Archivist nodes are backed up (`./scripts/archivist-backup.sh --snapshot` completed within the last 24 hours).
2. The `v2.3` container image is built (`make build-archivist`), tagged, and pushed to the private registry accessible by the staging cluster.
3. Monitoring/alerting and health checks for storage growth/quota/prune/replication are operational.
4. The upgrade plan and communication plan are approved by the gate owner (documented in `docs/v2.3/phase5/p5-evidence-index.md`).

## Upgrade / rollback script
The following script shows the procedural steps the operator will run (or simulate) during the drill. Adjust image names/containers to your environment.

```bash
#!/usr/bin/env bash
set -euo pipefail

# Step 1: Pause ingestion and mark nodes read-only
./scripts/archivist-control.sh --pause-ingest --tag drill-upgrade

# Step 2: Drain the node via Podman
podman stop archivist-node-1

# Step 3: Pull and run the new image
podman pull registry.local/xorein/archivist:v2.3
podman rm archivist-node-1
podman run --name archivist-node-1 -d registry.local/xorein/archivist:v2.3

# Step 4: Run smoke-checks (Podman scenarios)
./scripts/v23-podman-scenarios.sh --target archivist-node-1

# Step 5: If regression or storage alarms fire, rollback
podman tag archivist-node-1 registry.local/xorein/archivist:backup
podman stop archivist-node-1
podman run --name archivist-node-1 -d registry.local/xorein/archivist:v2.2
./scripts/archivist-control.sh --resume-ingest

# Step 6: Verify legacy metrics and note evidence
podman logs archivist-node-1 | tee artifacts/archivist-drill.log
```

When executing, capture the outputs of each step so that the evidence index can reference them. If the upgrade succeeds, the rollback section becomes a contingency plan that is still captured for `EV-v23-G4-###`.

## Verification checklist
| Step | Goal | Evidence requirement |
| --- | --- | --- |
| Storage/replication metrics | Confirm storage growth/quota/prune/replica alarms stay green after upgrade | Metric snapshots + screenshot/log referenced by `EV-v23-G4-###` |
| Podman scenarios | Re-run the key Podman scenarios from `docs/v2.3/phase3/p3-podman-scenarios.md` post-upgrade | Scenario log + pass/fail bits referenced by `EV-v23-G5-###` |
| Rollback readiness | Validate rollback script restores pre-upgrade artifact (configs, snapshots) | `podman ps`, `journalctl` clip, and backup hash with `EV-v23-G4-###` |

## Evidence capture template
| Gate | Evidence ID | Status | Notes | Responsible |
| --- | --- | --- | --- | --- |
| G4 | `EV-v23-G4-001` | pending | Scheduled drill run for upgrade script on `YYYY-MM-DD`; attach command outputs when available. | Archivist ops lead |
| G4 | `EV-v23-G4-002` | pending | Rollback rollback verification log (podman logs, metrics). | Archivist ops lead |
| G5 | `EV-v23-G5-001` | pending | Podman regression scenario log after upgrade; include pass/fail summary. | Regression engineer |
| G5 | `EV-v23-G5-002` | pending | Post-rollback check to ensure auto-regression gating still green. | Regression engineer |

Update the `Status` column from `pending` to `complete` after the drill runs, and keep the `Notes` column honest about whether the drill was a tabletop exercise or live run.
