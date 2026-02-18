# v1.7 Moderation Deployment Runbook

1. Pull or build the relay image tagged `aether/relay:v1.7`.
2. Use `docker compose -f containers/v1.7/docker-compose.yml up` to launch the relay and audit viewer services.
3. Once containers are healthy, run `scripts/v17-moderation-scenarios.sh` from the repo root to execute deterministic moderation scenarios and capture `artifacts/generated/v17-moderation-scenarios/result-manifest.json`.
4. Inspect manifest entries for scenario status and rerun failing jobs with `go test` commands if needed.
5. Capture logs with `docker compose logs moderation-relay` when investigating enforcement or audit failures.
