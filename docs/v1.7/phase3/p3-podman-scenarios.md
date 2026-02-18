# Phase 3 — Podman Moderation Scenarios

## Scenario Steps
1. Start the relay stack via Podman using `docker compose -f containers/v1.7/docker-compose.yml up -d`.
2. Run `scripts/v17-moderation-scenarios.sh` to execute deterministic scenario suites. This script drives `go test` for the v17 contracts and records `artifacts/generated/v17-moderation-scenarios/result-manifest.json` for auditing.
3. Validate the relay no-data-hosting regression by ensuring the manifest contains the `relay-policy` entry and that `pkg/v11/relaypolicy` rejects `durable-message-body` storage.
4. Capture logs for troubleshooting with `docker compose -f containers/v1.7/docker-compose.yml logs` and include them in the evidence bundle if failures arise.
