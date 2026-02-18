# v1.8 Discovery Deployment Runbook

1. Pull or build `aether/relay:v1.8`.
2. Start the discovery stack with `docker compose -f containers/v1.8/docker-compose.yml up -d`.
3. Confirm both `discovery-relay` and `indexer` are healthy.
4. From repo root run `scripts/v18-discovery-scenarios.sh` to capture `artifacts/generated/v18-discovery-scenarios/result-manifest.json`.
5. Save test output for any failures and include it in the evidence runbook entry.
6. Collect container logs with `docker compose -f containers/v1.8/docker-compose.yml logs` while diagnosing relay-policy and discovery indexer behavior.
