# Phase 4 local API documentation (as-built truth)

This planning-level artifact describes the local API surfaces (daemon control plane and attach endpoints) that must appear in the `G7` documentation bundle. It keeps the planning statement separate from the implementation narrative: the right-hand columns below reference the current, verifiable intent while noting that the live documentation is still being drafted.

## Planning vs implemented boundary
| API topic | Planning note | As-built note | Artifact reference | Evidence placeholder |
| --- | --- | --- | --- | --- |
| Daemon version negotiation + command set | Plan to explain how `cmd/xorein`/`cmd/harmolyn` negotiate schemas, fallback rules, and threat mitigations. | Reference uses the `F26` acceptance matrix rows that already capture command semantics without claiming new behavior. | `docs/v2.5/phase4/f26-acceptance-matrix.md` | `EV-v26-G7-201` |
| Attach protocol and attach authz weave | Plan to detail attach session lifecycle, service discovery, and anti-abuse markers once the docs settle. | Current implementation statements circle back to the `F26` final-closure spec and the boundary probes that keep local APIs compliant. | `docs/v2.5/phase4/f26-final-closure-spec.md` | `EV-v26-G7-202` |
| Local threat model + permission guarantees | Plan to publish sections on relay-only history, relay-only blob hosting, and local API permission defense policy. | Draft references the planned threat model from `docs/templates/roadmap-gate-checklist.md` so reviewers see how we intend to cite hazards. | `docs/templates/roadmap-gate-checklist.md` | `EV-v26-G7-203` |
| Evidence index + command audit | Plan to publish documented evidence of local API command runs (e.g., `go test ./cmd/xorein`). | Template alignment is already laid out in `docs/templates/roadmap-evidence-index.md`; once commands execute, the evidence list will be inserted. | `docs/templates/roadmap-evidence-index.md` | `EV-v26-G7-204` |

## Evidence & checklist tooling
- The local API doc will cite `docs/templates/roadmap-gate-checklist.md` for `G7` gating and `docs/templates/roadmap-signoff-raci.md` for role attribution.
- Evidence entries include command outputs per `docs/templates/roadmap-evidence-index.md` and are referenced via the `EV-v26-G7-###` namespace above.
- The documentation build commands that publish these pages should be recorded with `EV-v26-G7-205` for cross-checking when the doc boat is launched.

## Next steps (planning)
- Draft sections that describe version negotiation, attach lifecycle, and threat mitigations by pulling the vetted text from the v2.5 artifacts without claiming new wire implementations.
- Align diagram captions with the local API boundary probes documented in `docs/v2.6/phase1/p1-boundary-report.md` so readers can trace compliance to `G8`.
- Run the doc generation step and capture the output as required evidence so the checklist entry moves from planning to implemented.
