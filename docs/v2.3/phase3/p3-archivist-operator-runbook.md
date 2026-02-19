# Archivist Operator Runbook (G4)

This planning artifact documents the G4 operator readiness expectations for Archivist storage/search services. It captures the signal list, triage steps, and remediation actions needed to keep the storage plane within agreed SLOs. Evidence entries should reference the EV-v23-G4-### index when the runbook is exercised.

## Monitoring & alerts
The operator needs the following alarms wired into the monitoring stack (Prometheus, Grafana, Podman event alerts, etc.) and wired to on-call notification channels.

### Storage growth
- **Source:** Archivist volume (local disk and attached object store) used in the active region.
- **Threshold:** 12-hour growth > expected archival rate tier, or disk usage >75% of configured volume.
- **Triage:** Confirm the backing volume and node using `df -h` or `podman volume inspect`, check `archivist_ingest_bytes_total` and `archivist_prune_bytes_total` for anomalies, and review `storage-growth` dashboards for sustained spikes.
- **Remediation:** Pause ingest traffic via the API drain flag, run the prune workflow (`./scripts/archivist-prune.sh --confirm`), and, if needed, expand the volume or move a replica. Document all steps under `EV-v23-G4-001` (planned) or the next available G4 evidence entry.

### Quota exhaustion
- **Source:** Subscriber/account quota metrics (`archivist_quota_consumed_ratio`) and operator-visible refusal counts.
- **Threshold:** >90% quota burn rate per-tier or quota refusal rate rising for two consecutive 5-minute windows.
- **Triage:** Correlate active requests with `archivist_request_rate` and identify offending API keys/nodes via `journalctl -u archivist`. Confirm there is no downstream storage failure causing repeated retries.
- **Remediation:** Temporarily throttle offending keys, tighten admission controls, and notify the request owner of the backpressure reason. Escalate to security/privacy if the pattern matches abuse. Log actions under the next `EV-v23-G4-###` entry with “Quota alarm” context.

### Prune lag
- **Source:** `archivist_prune_lag_seconds` and the `prune.toml` job execution tracker; Podman scenario logs show backlog length.
- **Threshold:** Lag > 30 minutes for the configured prune cadence or the dequeued job count exceeds backlog depth.
- **Triage:** Check the `archivist-prune` pod logs for connectivity failures, confirm Cassandra/Postgres TTL scans are not blocked, and validate that disk I/O is not saturated.
- **Remediation:** Restart the prune job container with `podman restart archivist-prune`, verify prune tokens are not stalled (look for `prune: token acquired` messages), and, if necessary, rerun the prune workflow with elevated scheduling priority. Attach the execution log to `EV-v23-G4-###` when performed.

### Replica-target unmet (durability degraded)
- **Source:** Replica accounting metrics (`archivist_replica_target`, `archivist_replicas_available`) and health checks from each region.
- **Threshold:** Available replicas < target for >5 minutes or replica health z-tree showing `SYNCING`/`DOWN` states without recovery.
- **Triage:** Identify the affected archive shard, inspect network interface counters (`ip -s link`), and review `podman logs archivist-replica-$NODE` for error patterns (disk errors, handshake failures).
- **Remediation:** Shift traffic away from the degraded shard, trigger manual replication (`./scripts/archivist-replicate.sh --shard $ID --target $NODE`), or rebuild the replica from snapshot. Update the durability degradation log and reference `EV-v23-G4-###` for the gate evidence capture.

## Triage workflow
1. **Acknowledge the alarm:** assign the alert to the on-call engineer, note the incident number (e.g., G4-archivist-2026-02-19), and update the incident board.
2. **Gather context:** collect the relevant metrics, `podman ps`, and `journalctl` output before altering the system.
3. **Apply remediation:** follow the specific remediation steps above, document command outputs, and include any remainder gating updates or follow-up pulls.
4. **Verify recovery:** confirm the metrics return below threshold within two measurement intervals and send a summary to the gate owner.

## Gate evidence and follow-up
- **Evidence:** Log every action in `docs/v2.3/phase5/p5-evidence-index.md` using an `EV-v23-G4-###` identifier, even if the exercise is only tabletop planning (prefix with `planned:` to preserve honesty).
- **After-action:** If the incident triggers other gates (G5 regression, G7 go/no-go), reference this runbook in the subsequent incident follow-up documentation.

This runbook remains a planning artifact until the first live drill executes. Mark actions that remain theoretical as `pending` or `not yet run` within the gate evidence log.
