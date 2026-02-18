# v1.9 Deployment Runbook

## Objectives
- Validate CO path ladder and failover behaviors in Podman environments.
- Ensure relay no-data hosting regression tests accompany chaos scenarios.

## Steps
1. Pull or build the v1.9 orchestrator image and supporting containers.
2. Launch `docker-compose -f containers/v1.9/docker-compose.yml up`.
3. Confirm both `orchestrator` and `chaos` services report deterministic logs and exit cleanly.
4. Run `scripts/v19-chaos-scenarios.sh` to produce the manifest under `artifacts/generated/v19-chaos-scenarios/result-manifest.json`.
5. Archive manifest with e.g. `tar czf artifacts/v19-chaos-scenarios.tgz artifacts/generated/v19-chaos-scenarios/result-manifest.json` for evidence.
6. Tear down containers with `docker-compose down`.

## Recovery
- If a service fails, collect `docker-compose logs` for the failing service.
- Re-run the chaos script after the environment is stable.

## Signoff
- Document manifest hash and attach to Phase 5 evidence records.
