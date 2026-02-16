# v0.3 Phase 0 - P0-T2 Compatibility and Governance Checklist

> Status: Execution artifact. Compatibility and governance controls now reference `pkg/v03` schemas and trace to downstream gating docs.

## Purpose

Document additive-only protobuf, wire, and governance rules so that every protocol-affecting change in v0.3 follows the same checklist and the repo maintains protocol-first discipline.

## Compatibility Checklist

| Check | Requirement | Evidence placeholder |
|---|---|---|
| Proto additive-only policy | Any future schema/message evolution in v0.3 must remain additive; no field renumbering or type changes without new identifier families. | `docs/v0.3/phase6/p6-t2-conformance-review.md` |
| Wire stability | Contract helpers must keep deterministic, backwards-compatible behavior and explicit downgrade/fallback handling. | `pkg/v03/conformance/gates.go`, `pkg/v03/conformance/gates_test.go` |
| Security-mode constancy | Security mode labels and reason codes keep existing semantics; any major change triggers P0-T2-ST2 governance path. | `pkg/v03/voice/contracts.go`, `pkg/v03/transfer/contracts.go` |

## Governance Triggers

- **Major-change trigger:** Any proposal requiring incompatible behavior (new multistream IDs, tracing changes without negotiation, or breaking security posture) must include: new ID plan, downgrade negotiation steps, and multi-implementation verification schedule before approval.
- **Open decision guardrail:** Decisions OD3-01..OD3-04 stay `Open` until authoritative docs resolve them; no doc may reference them as resolved within this artifact.
- **Compliance review evidence:** Each claim of wire/gov compliance references a doc under `docs/v0.3/phase6/` and the matching `pkg/v03` tests.

## Implementation Checks

1. `pkg/v03` contract files include deterministic reason taxonomies tied to one of the phase-level baseline docs before merge.
2. Compatibility review requires evidence of `buf lint` and `buf breaking` outputs saved under `docs/v0.3/phase6/p6-t2-conformance-review.md`.
3. Governance checklist signoff (including major-change path) is part of the release gate handoff doc `docs/v0.3/phase6/p6-t3-release-gate-handoff.md`.
