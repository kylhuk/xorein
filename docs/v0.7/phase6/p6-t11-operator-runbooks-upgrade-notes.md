# Phase 6 · P6-T11 Operator Runbooks and Upgrade Notes

## Objective
Provide concrete operator guidance for bootstrapping v0.7 archive/search/push assets and for validating the local reference deployment.

## Relay and bootstrap runbook
1. Validate harness configuration:
   - `podman compose -f containers/v07/docker-compose.e2e.yml config`
2. Start local topology:
   - `podman compose -f containers/v07/docker-compose.e2e.yml up -d --build`
3. Expected service topology in harness:
   - `bootstrap`
   - `relay-a`
   - `relay-b`
   - `pushrelay`
   - `client-a`
   - `client-b`
4. Stop/cleanup:
   - `podman compose -f containers/v07/docker-compose.e2e.yml down -v`

## Push relay runbook
- Binary smoke checks:
  - `go run ./cmd/aether-push-relay --mode=relay`
  - `go run ./cmd/aether-push-relay --mode=probe`
- CI/mock-provider posture:
  - Use local `probe` flow and harness-only wiring (no external provider credentials).
- Production/provider posture (bounded in v0.7 contract):
  - Keep relay payload-blindness constraints from `pkg/v07/pushrelay/contracts.go`.
  - Normalize provider failures to deterministic retry/fail terminal classes.

## Upgrade and rollback notes
- Upgrade entry:
  - Deploy v0.7 binaries/containers.
  - Run `go test ./pkg/v07/... ./cmd/aether-push-relay ./tests/e2e/v07/...` before promotion.
- Rollback posture:
  - Binary/container rollback is supported.
  - History/search retention behavior may reflect v0.7 policy-state transitions already applied; rollback should preserve purge/audit records.

## Storage sizing guidance (minimum)
- Store-forward sizing: assume 30-day TTL envelopes with k=20 replication intent.
- Archivist capacity: plan for full-history retention obligations with explicit withdrawal behavior.
- Search index capacity: include SQLCipher FTS5 index growth for channel/server/DM scopes.

## Privacy posture
- Do not log plaintext payload content, keys, or token secrets.
- Restrict relay-visible push metadata to routing-only fields.
- Preserve deterministic reason-class observability without sensitive material.

## Evidence anchors
- `containers/v07/docker-compose.e2e.yml`
- `containers/v07/Containerfile`
- `cmd/aether-push-relay/main.go`
- `pkg/v07/storeforward/contracts.go`
- `pkg/v07/history/contracts.go`
- `pkg/v07/search/contracts.go`
- `pkg/v07/pushrelay/contracts.go`
- `pkg/v07/notification/contracts.go`
