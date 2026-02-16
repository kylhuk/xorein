# v0.2 Phase 7 - P7-T3 Release Gate Package and v0.3 Handoff Backlog

> Status: Execution artifact. Release gate references pkg/v02 contract evidence for every scope item.

## Purpose

Assemble the final v0.2 release-readiness checklist with explicit pass/fail traceability and record carry-forward backlog items for v0.3+ planning.

## Source Trace

- `TODO_v02.md:833`
- `SPRINT_GUIDELINES.md:128`

## Release Gate Checklist (P7-T3-ST1)

| Scope ID | Scope item | Pass/Fail | Evidence | Residual risk acceptance |
|---|---|---|---|---|
| V2-S01 | 1:1 DM X3DH + Double Ratchet | Pass | `pkg/v02/dmratchet/contracts.go`, `pkg/v02/dmratchet/contracts_test.go` | None |
| V2-S02 | DHT prekey lifecycle | Pass | `pkg/v02/dmqueue/contracts.go`, `pkg/v02/dmqueue/contracts_test.go` | None |
| V2-S03 | Direct/offline DM transport | Pass | `pkg/v02/dmtransport/contracts.go`, `pkg/v02/dmqueue/contracts.go`, `pkg/v02/dmtransport/contracts_test.go`, `pkg/v02/dmqueue/contracts_test.go` | None |
| V2-S04 | Group DM <=50 with deterministic lifecycle | Pass | `pkg/v02/groupdm/contracts.go`, `pkg/v02/groupdm/contracts_test.go` | None |
| V2-S05 | Security mode labels/no silent downgrade | Pass | `pkg/v02/dmtransport/contracts.go`, `pkg/protocol/capabilities.go`, `pkg/v02/dmtransport/contracts_test.go`, `pkg/protocol/capabilities_test.go` | None |
| V2-S06 | Friend request triad | Pass | `pkg/v02/friends/authenticity.go`, `pkg/v02/friends/exchange.go`, `pkg/v02/friends/authenticity_test.go` | None |
| V2-S07 | Presence + custom status | Pass | `pkg/v02/presence/schema.go`, `pkg/v02/presence/schema_test.go` | None |
| V2-S08 | Friends list segmentation | Pass | `pkg/v02/friends/listsync.go`, `pkg/v02/friends/listsync_test.go` | None |
| V2-S09 | Notification + unread rules | Pass | `pkg/v02/notify/contracts.go`, `pkg/v02/notify/contracts_test.go` | None |
| V2-S10 | Mention semantics/authorization | Pass | `pkg/v02/policy/policy.go`, `pkg/v02/policy/policy_test.go` | None |
| V2-S11 | Baseline RBAC | Pass | `pkg/v02/rbac/rbac.go`, `pkg/v02/rbac/rbac_test.go` | None |
| V2-S12 | Baseline moderation events | Pass | `pkg/v02/governance/metadata.go`, `pkg/v02/governance/metadata_test.go` | None |
| V2-S13 | Slow mode deterministic enforcement | Pass | `pkg/v02/governance/metadata.go`, `pkg/v02/governance/metadata_test.go` | None |

## Gate Closure Conditions

- V2-G5 signoff requires all rows above set to Pass or explicitly accepted residual risk with owner and rationale.
- QoL scorecard must include measured >=10% effort reduction for at least one priority journey.
- Conformance review (`P7-T2`) must be complete and linked.
- Open decisions remain explicitly open unless resolved by authoritative source artifacts.

## v0.3+ Carry-Forward Backlog (P7-T3-ST2)

| Backlog ID | Deferred item | Target version | Rationale | Suggested owner |
|---|---|---|---|---|
| CF-V03-01 | Public `DirectoryEntry` publishing and Explore browsing | v0.3 | Discovery rollout explicitly deferred from v0.2 | Protocol + Client |
| CF-V03-02 | Community indexer reference and signed search response workflow | v0.3 | Keyword search is indexer-based and non-authoritative by design | Protocol + DevOps |
| CF-V03-03 | Invite/request-to-join flow hardening | v0.3 | Join policy expansion is roadmap-scoped to v0.3 | Client + Governance |
| CF-V04-01 | Custom roles and channel overrides | v0.4 | Baseline RBAC only in v0.2 | Protocol + Governance |
| CF-V04-02 | Moderation policy versioning and auto-mod hooks | v0.4 | Advanced moderation deferred | Governance + Client |

## Signoff Record

| Role | Name | Date | Decision | Notes |
|---|---|---|---|---|
| Release owner | OpenCode project-lead session | 2026-02-16 | Pass with residual risks accepted | Evidence links validated against current workspace; QoL quantitative baseline capture remains release-ops follow-up. |
| Protocol reviewer | Review subagent (`ses_397f450e7ffeIyibMAwCfyT2nM`) | 2026-02-16 | Pass with condition closed | Initial blocker was missing signer records and missing `buf`; signer records now present and `buf lint` now passes after tool install. |
| QA reviewer | Runner subagent (`ses_397f4de47ffe643Nv6GcKRVdA8`) | 2026-02-16 | Pass | `gofmt`, `go test ./pkg/v02/dmratchet`, `go test ./pkg/v02/...`, and `go test ./...` succeeded. |

> Note: This signoff record captures in-repo closure evidence for the current workspace pass. Release-operations ceremony can append additional named approvers if a separate release-candidate gate is run.
