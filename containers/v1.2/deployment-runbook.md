# v1.2 Relay Recovery Baseline Runbook

## Purpose
This runbook defines the v1.2 relay-mode baseline used for identity and backup recovery validation.

## Baseline commands
1. Build relay binary:
   - `go build -o bin/aether ./cmd/aether`
2. Run compose baseline:
   - `podman compose -f containers/v1.2/docker-compose.yml up`
3. Stop baseline:
   - `podman compose -f containers/v1.2/docker-compose.yml down`

## Recovery scenario suite
- Execute deterministic suite:
  - `scripts/v12-recovery-scenarios.sh`
- Outputs:
  - `artifacts/generated/v12-recovery-scenarios/result-manifest.json`
  - per-scenario logs in `artifacts/generated/v12-recovery-scenarios/`

## Expected checks
- New-device restore passes.
- Lost-password path returns deterministic unrecoverable reason (`backup-password`) without server reset.
- Backup tamper detection returns deterministic corruption reason (`backup-corrupt`).
- Relay no-data-hosting regression check continues rejecting durable storage modes.
