# v0.7 Bootstrap Runbook

This runbook captures bootstrap-node operations for the v0.7 reference harness.

## Startup checks
- Validate compose topology: `podman compose -f containers/v07/docker-compose.e2e.yml config`
- Confirm service declarations include `bootstrap`, `relay-a`, `relay-b`, `pushrelay`, `client-a`, and `client-b`.

## Runtime checks
- Bootstrap/relay binary smoke: `go run ./cmd/aether --mode=relay`.
- Push relay smoke: `go run ./cmd/aether-push-relay --mode=relay` and `--mode=probe`.

## Recovery and rollback
- Stop and remove stack: `podman compose -f containers/v07/docker-compose.e2e.yml down -v`.
- Roll back binaries/containers if required while preserving retention and audit expectations.

## Canonical references
- `docs/v0.7/phase6/p6-t11-operator-runbooks-upgrade-notes.md`
- `docs/v0.7/phase6/p6-t12-build-run-validation.md`
