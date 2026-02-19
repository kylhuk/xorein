# F26 Acceptance Matrix

Planning-only matrix for `F26` final closure coverage that anchors the targets and preparedness checks for gate `G8`. Implementation evidence will replace the placeholders during v26 execution.

| Criterion | Target | Evidence placeholder |
| --- | --- | --- |
| Final closure spec publication | Publish `docs/v2.5/phase4/f26-final-closure-spec.md` along with traceable requirement-to-gate references and sign-offs that prove the system is ready for v26 consumption. | `EV-v25-G8-012` |
| Proto delta readiness | Proto surfacess (`FinalClosureStatus`, `ReleaseArtifactCatalog`, `GateSignature`, `EvidenceIndexRef`) are approved for additive publication, with documented vetting of connectors that will read them. | `EV-v25-G8-013` |
| Regression matrix coverage | Identity, messaging, media/persistence, moderation, and discovery pillars all have regression scenarios defined and rehearsed so `G8` reflects cross-layer stability. | `EV-v25-G8-014` |
| Release packaging reproducibility | Every release artifact (binaries, docs, evidence indexes) is cataloged with cryptographic fingerprints, and a reproducible-build proof is captured for the final gate. | `EV-v25-G8-015` |

Additional acceptance rows (e.g., automation handoff, QA sign-off) can be appended as the release plan matures, but the rows above set the minimal proof anchors for `F26` as of v25 planning.
