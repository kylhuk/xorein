# Phase 3 Podman Scenarios

- Use `containers/v1.5/docker-compose.yml` as baseline scenario with relay + screenshare placeholders.
- Run `scripts/v15-screenshare-scenarios.sh` inside container host to assert deterministic pass/fail and emit `artifacts/generated/v15-screenshare-scenarios/result-manifest.json`.
- Evidence for relay no-data regression is captured via end-to-end test referencing `pkg/v11/relaypolicy`.
