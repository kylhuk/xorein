# Phase 6 · P6-T1 Release Conformance Checklist

## Objective
Summarize v0.7 scope-bullet conformance with evidence links for V7-G6 readiness.

## Scope checklist

| Scope ID | Scope bullet | Status | Primary evidence |
|---|---|---|---|
| S7-01 | Robust DHT store-forward | Implemented | `pkg/v07/storeforward/contracts.go`, `pkg/v07/storeforward/contracts_test.go`, `tests/e2e/v07/archive_flow_test.go` |
| S7-02 | History sync + Merkle verification | Implemented | `pkg/v07/history/contracts.go`, `pkg/v07/history/contracts_test.go`, `tests/e2e/v07/archive_flow_test.go` |
| S7-03 | Configurable retention per server | Implemented | `pkg/v07/retention/contracts.go`, `pkg/v07/retention/contracts_test.go`, `tests/e2e/v07/archive_flow_test.go` |
| S7-04 | Archivist capability role | Implemented | `pkg/v07/archivist/contracts.go`, `pkg/v07/archivist/contracts_test.go`, `tests/e2e/v07/archive_flow_test.go` |
| S7-05 | SQLCipher FTS5 scoped search contract | Implemented (contract level) | `pkg/v07/search/contracts.go`, `pkg/v07/search/contracts_test.go`, `tests/e2e/v07/search_notification_flow_test.go` |
| S7-06 | Required search filters | Implemented | `pkg/v07/search/contracts.go`, `pkg/v07/search/contracts_test.go`, `tests/e2e/v07/search_notification_flow_test.go` |
| S7-07 | Encrypted push relay (relay-blind) | Implemented (contract + runtime skeleton) | `pkg/v07/pushrelay/contracts.go`, `cmd/aether-push-relay/main.go`, `tests/e2e/v07/search_notification_flow_test.go` |
| S7-08 | Desktop native notifications | Implemented (contract level) | `pkg/v07/notification/contracts.go`, `pkg/v07/notification/contracts_test.go`, `tests/e2e/v07/search_notification_flow_test.go` |
| S7-09 | History migration across security-mode epochs / History Capsule | Implemented (contract level) | `pkg/v07/history/contracts.go`, `pkg/v07/history/contracts_test.go`, `tests/e2e/v07/archive_flow_test.go` |

## Build and run summary
- `go test ./...` -> pass
- `go test ./pkg/v07/... ./cmd/aether-push-relay ./tests/e2e/v07/...` -> pass
- `go run ./cmd/aether --mode=client` -> pass
- `go run ./cmd/aether --mode=relay` -> pass
- `go run ./cmd/aether-push-relay --mode=relay` -> pass
- `go run ./cmd/aether-push-relay --mode=probe` -> pass
- `podman compose -f containers/v07/docker-compose.e2e.yml config` -> pass

## Residual risks
- CI execution of v0.7 compose/e2e suite remains pending external CI run.
- Upgrade test from a pre-existing encrypted v0.6 DB remains a tracked follow-up in release checklist.

## Evidence anchors
- `docs/v0.7/phase0/`
- `docs/v0.7/phase1/`
- `docs/v0.7/phase2/`
- `docs/v0.7/phase3/`
- `docs/v0.7/phase4/`
- `docs/v0.7/phase5/`
- `docs/v0.7/phase6/p6-t12-build-run-validation.md`
- `pkg/v07/`
- `cmd/aether/`
- `cmd/aether-push-relay/`
- `tests/e2e/v07/`
- `containers/v07/docker-compose.e2e.yml`
