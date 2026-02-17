# v0.7 Relay Runbook

This runbook captures relay-mode operations for v0.7 Archive assets.

## Startup
- Validate harness config: `podman compose -f containers/v07/docker-compose.e2e.yml config`
- Start topology: `podman compose -f containers/v07/docker-compose.e2e.yml up -d --build`

## Relay service expectations
- `bootstrap`, `relay-a`, and `relay-b` services are present in `containers/v07/docker-compose.e2e.yml`.
- Relay smoke command: `go run ./cmd/aether --mode=relay`.

## Rollback posture
- Binary/container rollback is supported.
- Retention and purge/audit behavior should be preserved during rollback.

## Canonical references
- `docs/v0.7/phase6/p6-t11-operator-runbooks-upgrade-notes.md`
- `docs/v0.7/phase6/p6-t12-build-run-validation.md`
