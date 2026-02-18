# v1.1 Relay Operator Baseline

## Purpose

Provide a minimal operator reference for the v1.1 relay baseline used by the Podman smoke checks.

## Baseline commands

1. Build or refresh the local relay binary:
   - `go build -o bin/aether ./cmd/aether`
2. Start the minimal relay service:
   - `podman compose -f containers/v1.1/docker-compose.yml up`
3. Stop the relay service:
   - `podman compose -f containers/v1.1/docker-compose.yml down`

## Smoke outputs

- Baseline smoke logs and manifest are written under `artifacts/generated/v11-relay-smoke/`.
- If both checks pass, the expected success path is: one allowed persistence probe and one rejected forbidden probe.
