# Durability Conformance (Phase 1)

> This artifact maps the required durability hardening ST1–ST3 checks to the G2 security gate and enumerates the evidence commands that prove replica accounting, degraded-state surfacing, and service-functional-but-degraded coverage.

| ST | Description | Gate | Evidence command | Notes |
| --- | --- | --- | --- | --- |
| ST1 | Verify replica-set accounting under churn and partial failures | G2 | `go test ./pkg/v23/durability/...` | Unit tests exercise churn detection, capacity thresholds, and partial failure transitions in a deterministic replica accounting model. |
| ST2 | Surface degraded durability as explicit state (API + UI label semantics) | G2 | `go test ./pkg/v23/durability/...` | Status objects serialize `state`, `reason`, and `label` for downstream UI surfaces; labels include explicit human-readable messaging (e.g., "Durability degraded: ..."). |
| ST3 | Regression coverage for "replica target unmet but service still functional" | G2 | `go test ./tests/e2e/v23/replica_*` | E2E-level tests assert `DurabilityStateDegraded` with `ReadyReplicas >= MinReplicas` while `ReadyReplicas < TargetReplicas`, proving functional yet degraded churn responses. |

## Evidence commands

1. `go test ./pkg/v23/durability/...` – covers deterministic accounting, churn detection, status reason codes, and label semantics described above.
2. `go test ./tests/e2e/v23/replica_*` – simulates churn + partial failures plus the service-functional-but-degraded trajectory required by ST3.
