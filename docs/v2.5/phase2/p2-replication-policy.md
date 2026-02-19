# P2-T2 Replication Policy

## ST1 → G4
- Publish workflows target the configured `r_blob` catalog and record a deterministic replica set for each `BlobRef` so G4 can attest to persistent availability metadata.
- Evidence: `go test ./tests/e2e/v25 -run TestReplicationPolicyPublishPrefersLocalAndDiversifies`.

## ST2 → G4
- Health verification inspects metadata, reuses the healthy subset, and adds replacement replicas when churn is detected; repaired replica sets update verified timestamps so G4 can prove replica maintenance.
- Evidence: `go test ./tests/e2e/v25 -run TestReplicationPolicyRepairMaintainsSet`.

## ST3 → G5
- Provider selection prefers the configured preferred region and, when ASN data exists, avoids single-AS placement so that G5’s provider concentration guardrail is satisfied while still proving deterministic behavior.
- Evidence: `go test ./tests/e2e/v25 -run TestReplicationPolicyPublishPrefersLocalAndDiversifies`.
