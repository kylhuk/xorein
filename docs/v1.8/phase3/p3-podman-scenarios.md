# Phase 3 - Podman Discovery Scenarios

## Scenario Steps
1. Launch discovery/indexer relay services using `docker compose -f containers/v1.8/docker-compose.yml up -d`.
2. Run `scripts/v18-discovery-scenarios.sh` from repository root.
3. Confirm the deterministic manifest is written to `artifacts/generated/v18-discovery-scenarios/result-manifest.json`.
4. Verify the manifest contains the scenario names:
   - `pkg-v18`
   - `e2e-v18`
   - `perf-v18`
   - `relay-boundary`
5. Review relay boundary evidence by checking that the `relay-boundary` scenario passed and `tests/e2e/v18/discovery_integrity_test.go` validates storage policy constraints.
6. Capture container and test logs for troubleshooting:
   - `docker compose -f containers/v1.8/docker-compose.yml logs`
   - `cat artifacts/generated/v18-discovery-scenarios/pkg-v18.log`
   - `cat artifacts/generated/v18-discovery-scenarios/e2e-v18.log`
   - `cat artifacts/generated/v18-discovery-scenarios/perf-v18.log`
   - `cat artifacts/generated/v18-discovery-scenarios/relay-boundary.log`
