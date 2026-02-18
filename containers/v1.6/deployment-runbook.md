# v1.6 Deployment Runbook

1. Validate container definitions via `docker compose -f docker-compose.yml config`.
2. Bring up `rbac-service` and `relay-policy` containers using `docker compose -f docker-compose.yml up -d`.
3. Run `scripts/v16-rbac-scenarios.sh` inside host to generate deterministic manifest in `artifacts/generated/v16-rbac-scenarios/result-manifest.json`.
4. Inspect manifest to confirm scenario `rbac-v16-admin-policy` passed and recorded timestamp.
5. Tear down containers with `docker compose -f docker-compose.yml down`.
