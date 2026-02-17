# v0.6 to v0.7 Upgrade Notes

This note tracks operator-facing upgrade considerations for v0.7 Archive assets.

## Preconditions
- Validate test baselines:
  - `go test ./...`
  - `go test ./pkg/v07/... ./cmd/aether-push-relay ./tests/e2e/v07/...`

## Runtime validation
- Client smoke: `go run ./cmd/aether --mode=client`
- Relay smoke: `go run ./cmd/aether --mode=relay`
- Push relay smoke: `go run ./cmd/aether-push-relay --mode=relay` and `--mode=probe`
- Harness config: `podman compose -f containers/v07/docker-compose.e2e.yml config`

## Data and retention notes
- Retention transitions and purge behavior are defined in `pkg/v07/retention/contracts.go`.
- History integrity and mode-epoch migration semantics are defined in `pkg/v07/history/contracts.go`.

## Rollback notes
- Binary/container rollback is supported.
- CI execution of compose/e2e suite and encrypted-DB migration validation remain tracked release checklist follow-ups in `TODO_v07.md`.

## Canonical references
- `docs/v0.7/phase6/p6-t11-operator-runbooks-upgrade-notes.md`
- `docs/v0.7/phase6/p6-t12-build-run-validation.md`
- `docs/v0.7/phase6/p6-t2-handoff-deferral-register.md`
