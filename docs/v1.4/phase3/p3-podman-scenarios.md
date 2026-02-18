# Phase 3 Podman Scenarios

Podman scenario coverage includes:
1. `voice-call-setup`: containerized relay + client verifying call setup handshake with deterministic fallback. Script `scripts/v14-voice-scenarios.sh` writes `artifacts/generated/v14-voice-scenarios/result-manifest.json`.
2. `voice-reconnect`: scenario replays reconnect/backoff and ensures recovery messaging surfaces before voice stability.
3. Relay boundary regression checks confirm `pkg/v11/relaypolicy` still rejects durable storage for voice flows.

Pods are defined in `containers/v1.4/docker-compose.yml`, and operational steps live in `containers/v1.4/deployment-runbook.md`.
