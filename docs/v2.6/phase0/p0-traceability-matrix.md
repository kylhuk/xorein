# Phase 0 Traceability Matrix (Planning)

This planning artifact links each `F26` requirement from v2.5 planning to the v2.6 artifacts that must prove it once implemented. The matrix is intentionally additive: each row cites the prior doc and traces forward to the gate or artifact that records closure evidence.

| Requirement | Source (`v2.5`) | v2.6 Target artifact | Gate | Evidence placeholder |
| --- | --- | --- | --- | --- |
| Final closure spec publication must prove preparedness for v26 consumption. | `docs/v2.5/phase4/f26-final-closure-spec.md` (Section 1–2). | `docs/v2.6/phase5/p5-final-evidence-bundle.md` + gate checklist. | `G0` scope lock → later `G7`. | `EV-v26-G0-005` (planning). |
| Proto delta must be additive and audited with connectors that will read the new fields. | `docs/v2.5/phase4/f26-proto-delta.md`. | `docs/v2.6/phase4/p4-protocol-docs.md` + `buf breaking` outputs. | `G1`. | `EV-v26-G0-006` (proto audit placeholder). |
| Regression matrix coverage ensures identity, messaging, media/persistence, moderation, discovery pillars are rehearsed. | `docs/v2.5/phase4/f26-acceptance-matrix.md` row 3. | `docs/v2.6/phase2/p2-regression-report.md`. | `G3`. | `EV-v26-G0-007` (matrix placeholder). |
| Release packaging reproducibility catalogues fingerprints for binaries, docs, and evidence. | `docs/v2.5/phase4/f26-acceptance-matrix.md` row 4. | `docs/v2.6/phase4/p4-repro-build.md` + `docs/v2.6/phase4/p4-signing.md`. | `G6`. | `EV-v26-G0-008` (build planning placeholder). |

Each trace entry above feeds the `G0` freeze: before `G0` closes, owners must confirm these requirements are still planned with the referenced artifacts and that the `F26` inputs have not drifted.
