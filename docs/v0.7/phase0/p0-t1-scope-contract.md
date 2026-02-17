# Phase 0 · P0-T1 Scope Contract

## Goal
Lock v0.7.0 Archive scope as the nine bullets listed in `aether-v3.md` and `TODO_v07.md` without implying code completion beyond roadmap intent. Every downstream artifact must be able to trace a bullet to the Phase 1–5 task, its governing reason-class, and the required reference implementation anchor so reviewers can confirm no scope creep into v0.8+ territory.

## Contract items
- Each of the nine Archive bullets gets a deterministic Phase task, a Phase doc anchor, and a planned code path placeholder so reviewers can find the future implementation seam (`pkg/v07/...`).
- The narration must keep the anti-scope creep exclusions (sections 3.1–3.4 of `TODO_v07.md`) visible to gate reviewers so the story remains “hardening/Archive contract” rather than feature expansion.
- The table below connects each scope bullet to a planned doc, expected code surface, and the Phase 1–5 gate that will own the evidence.

### Scope trace table
| Scope bullet | Phase doc | Planned code path | Gate owner | Evidence anchor |
|---|---|---|---|---|
| S7-01 · Robust DHT store-and-forward (30-day TTL, k=20) | `docs/v0.7/phase1/p1-t1-store-forward-retention-archivist.md` | `pkg/v07/storeforward/contracts.go` | V7-G1 | Same doc |
| S7-02 · Configurable history retention per server | Same doc | `pkg/v07/retention/contracts.go` | V7-G1 | Same doc |
| S7-03 · Archivist volunteer role contract | Same doc | `pkg/v07/archivist/contracts.go` | V7-G1 | Same doc |
| S7-04 · History sync protocol with Merkle verification | `docs/v0.7/phase2/p2-t1-history-sync-merkle.md` | `pkg/v07/history/contracts.go` | V7-G2 | Same doc |
| S7-05 · Merkle tree integrity and proof exchanges | Same doc | `pkg/v07/history/contracts.go` | V7-G2 | Same doc |
| S7-06 · Scoped SQLCipher FTS5 search (channel/server/DM) | `docs/v0.7/phase3/p3-t1-scoped-search-filters.md` | `pkg/v07/search/contracts.go` | V7-G3 | Same doc |
| S7-07 · Mandatory search filters (from user, date range, has file, has link) | Same doc | `pkg/v07/search/contracts.go` | V7-G3 | Same doc |
| S7-08 · Encrypted push relay + desktop notification coherence | `docs/v0.7/phase4/p4-t1-push-relay-desktop-notifications.md` | `pkg/v07/pushrelay/contracts.go`, `pkg/v07/notification/contracts.go` | V7-G4 | Same doc |
| S7-09 · Integrated validation, governance readiness, release handoff | `docs/v0.7/phase5/p5-t1-integrated-validation.md`, `docs/v0.7/phase5/p5-t2-governance-readiness-audit.md`, `docs/v0.7/phase5/p5-t3-release-gate-handoff.md` | `tests/e2e/v07/archive_flow_test.go`, `tests/e2e/v07/search_notification_flow_test.go`, `pkg/v07/governance/metadata.go`, `cmd/aether/main.go` | V7-G5/V7-G6 | Same docs |

This contract keeps the Archive narrative explicit while pointing every bullet to the new v0.7 docs and the future `pkg/v07` seams so downstream reviewers can follow the governance story from V7-G0 through release handoff.
