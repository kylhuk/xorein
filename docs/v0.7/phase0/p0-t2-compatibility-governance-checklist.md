# Phase 0 · P0-T2 Compatibility & Governance Checklist

## Purpose
Document the additive protobuf/wire implications implied by the Phase 1–5 tasks, capture the governance triggers that guard the Archive plan, and keep every open decision visible to prevent premature scope closure.

## Contract items
- Enumerate every new schema surface that Phase 1–5 plans rely on, reference the owning gate (V7-G1..V7-G6), and cite the future `VA-*` artifact plus its doc anchor so reviewers can follow the compatibility story.
- Codify the major-change triggers (downgrade negotiation, multistream IDs, AEP obligations, N+2 legacy support) and point each to an evidence artifact path such as `pkg/v07/governance/metadata.go` and the governance-readiness doc.
- Reference the `TODO_v07.md` open-decision register (Section 8) and explain why OD7-01..OD7-04 remain `Open`, including the gate where each gets revisited.

### Compatibility matrix by task
| Compatibility item | Gate trigger | VA artifact | Evidence anchor |
|---|---|---|---|
| Store-forward TTL/replication metadata (`VA-D1`/`VA-D2`) | V7-G1 additive-only schema extension | `VA-D1`, `VA-D2` | `docs/v0.7/phase1/p1-t1-store-forward-retention-archivist.md` |
| Retention policy override schema (`VA-D3`/`VA-D4`) | V7-G1 additive policy fields | `VA-D3`, `VA-D4` | Same doc |
| History sync stream IDs & lifecycle (`VA-H1`..`VA-H6`) | V7-G2 gating history negotiation | `VA-H1`..`VA-H6` | `docs/v0.7/phase2/p2-t1-history-sync-merkle.md` |
| Scoped search index/filter schema (`VA-S1`..`VA-S6`) | V7-G3 adds FTS5 filter bindings without removing fields | `VA-S1`..`VA-S6` | `docs/v0.7/phase3/p3-t1-scoped-search-filters.md` |
| Push envelope + desktop notification relay metadata (`VA-P1`..`VA-P6`) | V7-G4 ensures relay-blind fields remain optional and additive | `VA-P1`..`VA-P6` | `docs/v0.7/phase4/p4-t1-push-relay-desktop-notifications.md` |
| Integrated validation + governance evidence structures (`VA-I1`..`VA-I3`, `VA-R1`) | V7-G5/V7-G6 release-gate demonstration of compliance | `VA-I1`..`VA-R1` | `docs/v0.7/phase5/p5-t1-integrated-validation.md`, `docs/v0.7/phase5/p5-t2-governance-readiness-audit.md`, `docs/v0.7/phase5/p5-t3-release-gate-handoff.md` |

### Major-change trigger matrix
| Trigger area | Required evidence | Gate | Notes |
|---|---|---|---|
| Downgrade negotiation and AEP entry (history sync, search, push) | Documented fallback negotiation paths plus multi-implementation validation matrix | V7-G5 | Evidence anchored to `docs/v0.7/phase5/p5-t2-governance-readiness-audit.md` and `pkg/v07/governance/metadata.go` |
| New multistream IDs for history sync/search flow | Capability registration entries plus compatibility checklist referenced by `VA-H1` and `VA-S1` | V7-G2/V7-G3 | Checklist resides in this doc and ties back to `pkg/v07/history/contracts.go` and `pkg/v07/search/contracts.go` |
| Relay enrollment and encrypted push metadata | Proof of additive provider mapping + ledger entry `VA-P1` | V7-G4 | Relay blindness enforced via `pkg/v07/pushrelay/contracts.go` |
| N+2 legacy protocol-ID support | Signoff mentions (legacy connectors) in release handoff and governance doc | V7-G6 | Release document `docs/v0.7/phase5/p5-t3-release-gate-handoff.md` preserves the audit trail |

### Open decisions
| Decision | Current state | Revisit plan | Notes |
|---|---|---|---|
| OD7-01 · Replica-placement strategy under heterogeneous relays | `Open` | V7-G1 stores the unresolved alternatives in the validation scenarios (`VA-I1`) for final assessment | Keep strategy options bounded and report tradeoffs rather than declaring a single winner until evidence resolves the tension |
| OD7-02 · Merkle chunk-size profile | `Open` | V7-G2 records competing profiles in `VA-H3`/`VA-H4` and compares proof-size vs. recovery outcomes | Maintain both profile candidates in the governance dossier and avoid presenting a final size contract without interop evidence |
| OD7-03 · Scoped-search ranking defaults | `Open` | V7-G3 stores ranking knobs in `VA-S3`/`VA-S4` so the release dossier can show deterministic tie-breaks without locking one profile | Trending ranking defaults stay advisory until telemetry validates them |
| OD7-04 · Relay topology default | `Open` | V7-G4 governance readiness audit catalogues deployment tradeoffs before V7-G6 handoff | Keep the federation vs. centralized relay question visible and avoid presenting a default as settled architecture |

This checklist keeps compatibility and governance evidence gateable while leaving open decisions transparently unresolved.
