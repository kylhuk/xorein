# v0.2 Phase 2 - P2 DM Transport Baseline

> Status: Execution artifact. DM transport and offline delivery contracts are implemented in `pkg/v02/dmtransport/contracts.go` and `pkg/v02/dmqueue/contracts.go` with deterministic tests.

## Purpose

Define planning contracts for DHT prekey lifecycle, direct DM transport, and offline store-and-forward behavior.

## Task Coverage

| Task | Coverage summary | Evidence anchor |
|---|---|---|
| P2-T1 | Prekey bundle schema, signature validation, expiry metadata, rotation/depletion/republish policy. | `pkg/v02/dmqueue/contracts.go`, `pkg/v02/dmqueue/contracts_test.go` |
| P2-T2 | Direct DM stream lifecycle, session binding, metadata minimization, integrity checks. | `pkg/v02/dmtransport/contracts.go`, `pkg/v02/dmtransport/contracts_test.go` |
| P2-T3 | Offline mailbox addressing, TTL/replication policy, ack/retry/idempotency and dedupe semantics. | `pkg/v02/dmqueue/contracts.go`, `pkg/v02/dmqueue/contracts_test.go` |

## Validation Expectations

- Direct and offline paths must be coherent under retry and reconnection.
- Negative-path behavior is deterministic for invalid bundle, stale key, and duplicate-retrieval conditions.
- Delivery guarantees are represented as planned acceptance criteria with evidence placeholders.
