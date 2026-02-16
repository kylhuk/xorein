# v0.2 Phase 4 - P4 Friend Graph Baseline

> Status: Execution artifact. Friend exchange, lifecycle, and list sync contracts live in `pkg/v02/friends` with accompanying tests.

## Purpose

Define planning contracts for friend identity exchange formats, request lifecycle semantics, and friends-list synchronization behavior.

## Task Coverage

| Task | Coverage summary | Evidence anchor |
|---|---|---|
| P4-T1 | Canonical friend input formats (public key, QR, `aether://`) with validation and normalization. | `pkg/v02/friends/authenticity.go`, `pkg/v02/friends/exchange.go`, `pkg/v02/friends/authenticity_test.go` |
| P4-T2 | Friend request state machine, authenticity checks, replay/expiry handling. | `pkg/v02/friends/lifecycle.go`, `pkg/v02/friends/lifecycle_test.go` |
| P4-T3 | Friends-list projection rules and deterministic UI states for load/empty/error/pending flows. | `pkg/v02/friends/listsync.go`, `pkg/v02/friends/listsync_test.go` |

## Planning Constraints

- No centralized identity lookup dependency is introduced.
- Concurrent request conflict outcomes are deterministic.
- Presence-driven list divergence is bounded and covered by validation planning.
