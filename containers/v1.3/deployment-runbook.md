# v1.3 Deployment Runbook

1. Ensure the Podman scenario script `scripts/v13-e2e-podman.sh` is executable and the artifacts directory exists.
2. Launch `podman-compose -f containers/v1.3/docker-compose.yml up -d` to provision `relay` and `app` services on the isolated `v13-net` bridge.
3. Verify containers are healthy: `podman ps --filter "name=v13"`.
4. Execute the scenario script inside the app container or host shell to validate flows; the script writes `artifacts/generated/v13-e2e-podman/result-manifest.json` with deterministic pass/fail.
5. Collect logs from `podman logs relay` and `podman logs app` for evidence bundle.
