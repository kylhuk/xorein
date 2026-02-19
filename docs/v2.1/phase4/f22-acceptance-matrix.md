# F22 Acceptance Matrix

> Phase 4 planning artifact; records how the `F22` specification package proves out each acceptance criterion before v22 implementation begins.

| Feature | Criterion | Evidence |
|---|---|---|
| Archivist capability definition | Advertisement + quota + retention semantics are documented with deterministic failure reasons. | `docs/v2.1/phase4/f22-history-backfill-spec.md` (ST1, ST4) |
| History segment integrity | Manifests include commitments, signatures, and explicit tombstone pointers that map to the `HistoryHead`. | `docs/v2.1/phase4/f22-history-backfill-spec.md` (ST2) |
| Backfill protocol baseline | Requests/responses cover time windows, coverage hints, and failure signals that avoid keyword leakage. | `docs/v2.1/phase4/f22-history-backfill-spec.md` (ST3, ST4) |
| Search coverage continuity | UI and protocol labels extend the v21 coverage model without exposing plaintext keywords. | `docs/v2.1/phase4/f22-history-backfill-spec.md`, `docs/v2.1/phase2/p2-search-contract.md` |
| Proto delta guardrails | Additive messages/services with reserved numbers and explicit compatibility notes are captured for reviewers. | `docs/v2.1/phase4/f22-proto-delta.md` |
| Phase 4 traceability | Spec package surface links directly to TODO_v21 Phase 4 acceptance criteria and gating evidence. | `docs/v2.1/phase4/f22-acceptance-matrix.md` (self), `docs/v2.1/phase5/p5-evidence-index.md` |

All evidence is tagged under the `EV-v21-G6` series until commands and manifests record the actual outputs. Current entries are placeholders; no promotion is claimed until the evidence index is populated.
