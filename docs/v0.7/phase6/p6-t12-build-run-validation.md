# Phase 6 · P6-T12 Build and Run Validation

## Objective
Record concrete build/run smoke evidence for client, relay, push relay, and v0.7 test assets.

## Command evidence (executed)

1) `go test ./...`
- Exit: `0`
- Key output: package list with `ok` / `[no test files]`; no failing packages.

2) `go test ./pkg/v07/... ./cmd/aether-push-relay ./tests/e2e/v07/...`
- Exit: `0`
- Key output: all `pkg/v07/*`, `cmd/aether-push-relay`, and `tests/e2e/v07` reported `ok`.

3) `go run ./cmd/aether --mode=client`
- Exit: `0`
- Output: `Phase 2 foundation stub: cmd/aether mode=client`

4) `go run ./cmd/aether --mode=relay`
- Exit: `0`
- Output includes: `Relay runtime active ...` and `Relay health status: state=ready ...`

5) `go run ./cmd/aether-push-relay --mode=relay`
- Exit: `0`
- Output: `push relay runtime ready`

6) `go run ./cmd/aether-push-relay --mode=probe`
- Exit: `0`
- Output: `push relay probe ok`

7) Compose harness config validation
- `docker compose -f containers/v07/docker-compose.e2e.yml config` -> Exit `127` (`docker` unavailable in environment)
- `podman compose -f containers/v07/docker-compose.e2e.yml config` -> Exit `0` (compose configuration rendered successfully)

## Validation notes
- Client and relay mode flags are runnable locally.
- Push relay mode flags (`relay`, `probe`) are runnable locally.
- v0.7 package and e2e tests pass in current environment.
- Harness config is validated via Podman compose when Docker CLI is absent.

## Evidence anchors
- `pkg/v07/`
- `cmd/aether/main.go`
- `cmd/aether-push-relay/main.go`
- `tests/e2e/v07/`
- `containers/v07/docker-compose.e2e.yml`
