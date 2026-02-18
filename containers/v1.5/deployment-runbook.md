# v1.5 Containers Runbook

## Services
- `relay`: lightweight placeholder representing relay network for screenshare scenarios.
- `screenshare`: simulated screen sharing sender connected to relay.

## Deployment
1. Install Docker/Podman with Compose support.
2. Run `docker-compose -f containers/v1.5/docker-compose.yml up --build`.
3. Observe logs for both services to ensure the scenario exercises handshake + fallback path.
4. Tear down with `docker-compose -f containers/v1.5/docker-compose.yml down`.

## Evidence
- Scenario script writes `artifacts/generated/v15-screenshare-scenarios/result-manifest.json` with deterministic pass states.
