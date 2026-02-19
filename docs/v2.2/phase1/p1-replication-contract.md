# Phase 1 Replication Contract

- Replication policy defines `r` (desired replicas) and `r_min` (minimum acceptable replicas).
- Successful writes attempt every configured archivist endpoint in the pipeline; success counts determine the deterministic result reason:
  - `REPLICA_TARGET_UNMET` when fewer than `r_min` copies succeed.
  - `REPLICA_WRITE_PARTIAL` when at least `r_min` but fewer than `r` succeed.
  - no reason when `r` copies succeed and durability is `HISTORY_DURABILITY_HEALTHY`.
- Durability health labels convey the same summary: `HISTORY_DURABILITY_HEALTHY` when `r` copies exist, otherwise `HISTORY_DURABILITY_DEGRADED`.
- Healing helpers gather missing replicas and re-run writes; they surface `REPLICA_HEALING_IN_PROGRESS` when recovery work is recorded and fall back to `REPLICA_TARGET_UNMET` if healing cannot raise counts above `r_min`.
