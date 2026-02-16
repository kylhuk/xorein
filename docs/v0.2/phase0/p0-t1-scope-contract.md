# v0.2 Phase 0 - P0-T1 Scope Contract and Traceability Map

> Status: Execution artifact. Scope-to-task trace is backed by the v0.2 contract packages listed below.

## Purpose

Lock v0.2 scope boundaries and provide one-to-one traceability from each in-scope roadmap bullet to execution tasks, acceptance anchors, and evidence placeholders.

## Source Trace

- `TODO_v02.md:33`
- `aether-v3.md:723`
- `SPRINT_GUIDELINES.md:42`

## v0.2 Scope-to-Task Trace Matrix

| Scope ID | v0.2 roadmap bullet (in-scope) | Primary task mapping | Acceptance anchor | Evidence placeholder |
|---|---|---|---|---|
| V2-S01 | 1:1 DMs with X3DH + Double Ratchet | P1-T2, P1-T3 | DM sessions negotiable, persistent, recoverable | `pkg/v02/dmratchet/contracts.go`, `pkg/v02/dmratchet/contracts_test.go` |
| V2-S02 | Prekey bundles via DHT | P2-T1 | Publish/retrieve/rotate/expire lifecycle deterministic | `pkg/v02/dmqueue/contracts.go`, `pkg/v02/dmqueue/contracts_test.go` |
| V2-S03 | Direct + offline DM transport | P2-T2, P2-T3 | Direct and store-forward behavior coherent | `pkg/v02/dmtransport/contracts.go`, `pkg/v02/dmqueue/contracts.go`, `pkg/v02/dmtransport/contracts_test.go`, `pkg/v02/dmqueue/contracts_test.go` |
| V2-S04 | Group DM up to 50 (MLS target + bounded compatibility bridge) | P3-T1, P3-T2, P3-T3 | Create/invite/join/leave + secure rekey + cap enforcement | `pkg/v02/groupdm/contracts.go`, `pkg/v02/groupdm/contracts_test.go` |
| V2-S05 | Security-mode labels (Seal/Tree) with no silent downgrade | P1-T1-ST3, P6-T1 | Explicit mode label and deterministic unsupported-mode reasons | `pkg/v02/dmtransport/contracts.go`, `pkg/protocol/capabilities.go`, `pkg/v02/dmtransport/contracts_test.go`, `pkg/protocol/capabilities_test.go` |
| V2-S06 | Friend requests via key, QR, `aether://` | P4-T1, P4-T2 | Canonical request model + trust/authenticity checks | `pkg/v02/friends/authenticity.go`, `pkg/v02/friends/exchange.go`, `pkg/v02/friends/authenticity_test.go` |
| V2-S07 | Presence states + custom status | P5-T1, P5-T2, P5-T3 | Deterministic presence/status propagation and staleness handling | `pkg/v02/presence/schema.go`, `pkg/v02/presence/schema_test.go` |
| V2-S08 | Friends list segmentation | P4-T3 | Online/offline/pending projection convergence | `pkg/v02/friends/listsync.go`, `pkg/v02/friends/listsync_test.go` |
| V2-S09 | In-app notifications + unread counters | P6-T1 | Counter and event behavior deterministic per context | `pkg/v02/notify/contracts.go`, `pkg/v02/notify/contracts_test.go` |
| V2-S10 | Mention semantics (`@user`, `@role`, `@everyone`, `@here`) | P6-T2, P6-T3 | Parser/resolver/authorization deterministic | `pkg/v02/policy/policy.go`, `pkg/v02/policy/policy_test.go` |
| V2-S11 | Baseline RBAC (Owner/Admin/Moderator/Member) | P6-T4 | Role hierarchy and authority outcomes deterministic | `pkg/v02/rbac/rbac.go`, `pkg/v02/rbac/rbac_test.go` |
| V2-S12 | Baseline moderation events (redaction/delete, timeout, ban) | P6-T5-ST1 | Signed event contracts and enforcement semantics auditable | `pkg/v02/governance/metadata.go`, `pkg/v02/governance/metadata_test.go` |
| V2-S13 | Channel slow mode | P6-T5-ST2 | Deterministic timing, replay, and rejoin enforcement | `pkg/v02/governance/metadata.go`, `pkg/v02/governance/metadata_test.go` |

## Explicit Non-Goals and Overlap Boundaries

These items remain deferred and are not part of v0.2 closure:

- v0.3 media and transfer expansion: RNNoise/voice pipeline expansion, screen share, file transfer.
- v0.3 discovery rollout: DirectoryEntry publishing, Explore/search indexers, request-to-join and invite expansion.
- v0.4 governance expansion: custom roles, channel overrides, policy versioning, auto-moderation hooks.
- v0.5 bot API/Discord shim/reactions.
- v0.6-v0.7 hardening and deep history/search/push relay work.

Boundary rule for `@role` in v0.2: authorization references only baseline roles above; no custom-role CRUD implied.

## Scope Control Checks (P0-T1-ST2)

- Every new proposal must map to one `V2-Sxx` row above before acceptance.
- Any proposal that cannot map to `V2-S01..V2-S13` is deferred to the target minor version backlog.
- Scope updates require trace matrix update and explicit risk-log entry.

## Planned Evidence Anchors

- Scope lock signoff record: `docs/v0.2/phase7/p7-t3-release-gate-handoff.md` with linked implementation evidence.
- Non-goal acknowledgment by downstream phases: `docs/v0.2/phase7/p7-t3-release-gate-handoff.md` and this trace matrix.
