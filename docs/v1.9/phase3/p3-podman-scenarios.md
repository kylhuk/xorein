# Podman Chaos Scenarios

## Scenario Set
1. CO path failover in container network ensuring deterministic ladder traversal.
2. Relay boundary regression test ensuring no durable storage path is accepted.
3. Mobility and NAT matrix validations in controlled net namespaces.

## Evidence
- `scripts/v19-chaos-scenarios.sh` produces manifest at `artifacts/generated/v19-chaos-scenarios/result-manifest.json`.
- `containers/v1.9/docker-compose.yml` is the reproducible environment used in Phase 3 checks.
