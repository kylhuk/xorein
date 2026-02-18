# v1.4 Voice Podman Deployment Runbook

1. Start the relay and client containers with `podman-compose -f containers/v1.4/docker-compose.yml up -d`.
2. Verify the `voice-relay` service reports `ready` via logs and that port 4001 is listening.
3. Run `scripts/v14-voice-scenarios.sh` from the repo root to materialize scenario artifacts.
4. Collect `artifacts/generated/v14-voice-scenarios/result-manifest.json` for evidence reporting.
5. Bring the stack down with `podman-compose -f containers/v1.4/docker-compose.yml down`.
