# v0.2 Phase 0 - P0-T3 Verification Matrix and Evidence Schema

> Status: Execution artifact. Verification matrix now references implemented v0.2 contracts and tests.

## Purpose

Predefine validation expectations and evidence format for v0.2 so gate exits are auditable and deterministic.

## Source Trace

- `TODO_v02.md:214`
- `SPRINT_GUIDELINES.md:103`
- `aether-v3.md:729`

## Requirement-to-Test Matrix (P0-T3-ST1)

Each in-scope requirement includes at least one positive and one negative path.

| Scope ID | Requirement | Positive path | Negative path | Degraded/recovery path | Evidence anchor |
|---|---|---|---|---|---|
| V2-S01 | 1:1 DM X3DH + Double Ratchet | Valid bootstrap + encrypted roundtrip | Reject tampered handshake/ciphertext | Recover via retry with fresh prekey | `pkg/v02/dmratchet/contracts.go`, `pkg/v02/dmratchet/contracts_test.go` |
| V2-S02 | DHT prekey lifecycle | Publish/retrieve valid bundle | Reject expired/invalid signature bundle | Replenish after depletion threshold | `pkg/v02/dmqueue/contracts.go`, `pkg/v02/dmqueue/contracts_test.go` |
| V2-S03 | Direct/offline DM transport | Direct send to online peer | Reject malformed envelope/invalid receipt | Offline store-forward with retry + dedupe | `pkg/v02/dmtransport/contracts.go`, `pkg/v02/dmqueue/contracts.go`, `pkg/v02/dmtransport/contracts_test.go`, `pkg/v02/dmqueue/contracts_test.go` |
| V2-S04 | Group DM lifecycle (<=50) | Create/invite/join/leave with rekey | Reject invalid membership transition | Deterministic cap rejection and recovery guidance | `pkg/v02/groupdm/contracts.go`, `pkg/v02/groupdm/contracts_test.go` |
| V2-S05 | Security mode labels | Render explicit Seal/Tree mode | Reject unsupported mode silently downgrading | Deterministic unsupported-mode reason and action | `pkg/v02/dmtransport/contracts.go`, `pkg/protocol/capabilities.go`, `pkg/v02/dmtransport/contracts_test.go`, `pkg/protocol/capabilities_test.go` |
| V2-S06 | Friend request triad | Key/QR/link normalize to same request | Reject replayed/forged/expired request | Retry path with deterministic user feedback | `pkg/v02/friends/authenticity.go`, `pkg/v02/friends/exchange.go`, `pkg/v02/friends/authenticity_test.go` |
| V2-S07 | Presence + status | Online/idle/DND/invisible transitions | Reject invalid state transition | Stale cleanup converges to deterministic offline | `pkg/v02/presence/schema.go`, `pkg/v02/presence/schema_test.go` |
| V2-S08 | Friends list segmentation | Correct online/offline/pending projection | Reject contradictory state merge | Converge after delayed presence updates | `pkg/v02/friends/listsync.go`, `pkg/v02/friends/listsync_test.go` |
| V2-S09 | Notifications + unread | Increment/reset rules by context | Prevent duplicate inflation | Recovery on delayed read-marker sync | `pkg/v02/notify/contracts.go`, `pkg/v02/notify/contracts_test.go` |
| V2-S10 | Mention semantics | Parse/resolve valid mention tokens | Reject unauthorized or unknown mass-mention | Fallback rendering without false notifications | `pkg/v02/policy/policy.go`, `pkg/v02/policy/policy_test.go` |
| V2-S11 | Baseline RBAC | Deterministic allow/deny matrix | Reject forbidden actor-target operation | Stale-role reconciliation convergence | `pkg/v02/rbac/rbac.go`, `pkg/v02/rbac/rbac_test.go` |
| V2-S12 | Moderation events | Signed redaction/timeout/ban accepted | Reject bad signature/authority mismatch | Deterministic replay and audit visibility behavior | `pkg/v02/governance/metadata.go`, `pkg/v02/governance/metadata_test.go` |
| V2-S13 | Slow mode | Deterministic timing enforcement | Reject out-of-window messages | Rejoin/replay maintains same outcome | `pkg/v02/governance/metadata.go`, `pkg/v02/governance/metadata_test.go` |

## QoL 10 Percent Effort-Reduction Objective

- Priority journey: degraded DM send and recovery.
- Baseline metric: median number of user actions from failed send state to successful recovery.
- Target metric: at least 10 percent fewer user actions versus baseline.
- Scorecard anchor: `docs/v0.2/phase7/p7-t1-e2e-scenario-pack.md`.

## Evidence Artifact Schema (P0-T3-ST2)

Required fields for each completion package:

| Field | Required | Notes |
|---|---|---|
| Artifact ID | Yes | Stable identifier (`EVID-*`). |
| Gate | Yes | `V2-G0`..`V2-G5`. |
| Scope IDs covered | Yes | One or more `V2-Sxx`. |
| Validation type | Yes | Contract, integration, negative, degraded, recovery. |
| Scenario/Test ID | Yes | Deterministic scenario reference. |
| Result | Yes | Pass/Fail. |
| Reason class (if fail/degraded) | Yes | Deterministic taxonomy token. |
| User-visible next action | Yes | Required for no-limbo invariant. |
| Evidence link | Yes | File path or command-output reference. |
| Reviewer/signoff | Yes | Role + date. |

## Completion Review Checklist

- [ ] Every `V2-Sxx` has positive and negative evidence.
- [ ] Degraded/recovery coverage exists for interruption-prone flows.
- [ ] Reason taxonomy is stable across diagnostics and user messaging.
- [ ] QoL scorecard includes baseline, target, and measured delta.
- [ ] Planned-vs-implemented wording remains explicit.
