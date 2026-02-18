# Phase 3 Podman Scenarios

- `docker-compose.yml` defines placeholder `rbac-service` and `relay-policy` containers.
- `scripts/v16-rbac-scenarios.sh` runs deterministic hooks and writes `artifacts/generated/v16-rbac-scenarios/result-manifest.json`.
- Scenario passes only if manifest reports `rbac-v16-admin-policy` success and ambient environment enforces relay no-data constraints.
