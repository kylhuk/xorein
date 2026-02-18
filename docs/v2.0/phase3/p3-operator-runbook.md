# Operator Runbook

1. Deploy `containers/v2.0/docker-compose.yml` via `podman-compose up -d` for operator staging tests.
2. Monitor `relay` logs and ensure `pkg/v20/hardening` continuity expectations are met.
3. Validate podman scenarios by running `./scripts/v20-podman-scenarios.sh` and reviewing the generated manifest at `artifacts/generated/v20-podman-scenarios/result-manifest.json`.
