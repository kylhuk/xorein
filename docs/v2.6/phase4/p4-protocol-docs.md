# Phase 4 protocol documentation (as-built truth)

This artifact records the protocol documentation deliverable for `G7`. It keeps the planning-vs-implemented boundary explicit: every section notes what is still planned and what stage of the implementation narrative we can speak to today. The finished phase report will narrate the true, gate-ready protocol surface once the documents below are finalized, but today we anchor that work on the additive intent captured in the v2.5 `F26` corpus.

## Planning vs implemented boundary
| Protocol area | Planning note (phase 4 intent) | As-built note (current state) | Artifact reference | Evidence placeholder |
| --- | --- | --- | --- | --- |
| `F26` additive network schema and control plane | Plan to publish a consolidated protocol reference that explains relay handshakes, history-plane intent, and blob-plane invariants for v2.6. | As-built text draws from `docs/v2.5/phase4/f26-proto-delta.md` and the audited additive-only intent; no runtime wire changes are claimed ahead of gate closure. | `docs/v2.5/phase4/f26-proto-delta.md` | `EV-v26-G7-101` |
| History/relay boundary narrative | Plan to describe the relay no-long-history-hosting and no-durable-blob-hosting corridors, including private Space anti-enumeration. | Current narrative highlights the boundary probes that feed `G8` (`tests/e2e/v26/boundaries/*`) while deferring any new wire semantics. | `docs/v2.6/phase1/p1-boundary-report.md` | `EV-v26-G7-102` |
| Asset/bridge metadata expectations | Plan to document how attachments, bridges, and bot/webhook metadata behave on the wire for readers of the protocol spec. | Implementation side notes link to the existing `F26` acceptance matrix rows and the `p2-regression-report` to avoid overstating new behavior. | `docs/v2.6/phase2/p2-regression-report.md` | `EV-v26-G7-103` |
| Evidence indexing & narrative | Plan to include an evidence checklist section that follows `docs/templates/roadmap-evidence-index.md`. | Planners can already reference the template; the final deployed doc will be indexed the same way once lists are reviewed. | `docs/templates/roadmap-evidence-index.md` | `EV-v26-G7-104` |

## Evidence & checklist tooling
- Gate checklist for this document lives against `docs/templates/roadmap-gate-checklist.md` and will cite `G7` once all narrative sections and appendices are complete.
- Evidence index entries must follow the `EV-v26-G7-###` namespace; placeholders above may be supplanted by the actual ref files produced when the docs ship.
- The doc build command(s) that surface this artifact should be captured along with `EV-v26-G7-105` once the phase 4 documentation pipeline finishes.

## Next steps (planning)
- Fill in the detailed subsections (relay handshake, history plane, blob plane) from the final `F26` control plane audit without claiming new wire changes.
- Publish a formal appendix enumerating which proto fields remain stable-plus-additive, referencing `docs/v2.5/phase4/f26-acceptance-matrix.md` as trace input.
- Run the document build so the gate checklist entry from `docs/templates/roadmap-gate-checklist.md` can move from planning to implemented status.
