# v1.0 Execution Artifacts Repository

This directory captures the deterministic repositories for the v1.0 gates described in `TODO_v10.md`. Every phase file below includes an evidence anchor table that points to the deterministic artifacts created in this pass, and the planned-vs-implemented notes stay explicit about the difference between intent and the actual files.

Phases and their core artifacts:

| Phase | File | Key anchors |
|---|---|---|
| Phase 0 | `phase0/p0-scope-governance.md` | Scope, governance, and evidence schema anchors (`VA-G*`) + `pkg/v10` conformance code |
| Phase 1 | `phase1/p1-protocol-freeze.md` | Protocol surface inventory and compatibility controls (`VA-P*`) |
| Phase 2 | `phase2/p2-security-audit.md` | Security audit scope, threat, and engagement (`VA-S*`) |
| Phase 3 | `phase3/p3-spec-publication.md` | Publication architecture, terminology, and approval controls (`VA-P7`..`VA-P12`) |
| Phase 4 | `phase4/p4-docs-publication.md` | Documentation quality + release workflow (`VA-D*`) |
| Phase 5 | `phase5/p5-landing-comparison.md` | Landing/comparison surface claims (`VA-W*`) |
| Phase 6 | `phase6/p6-app-distribution.md` | Client distribution readiness (`VA-A*`) |
| Phase 7 | `phase7/p7-bootstrap-infra.md` | Bootstrap topology and operator continuity (`VA-N*`) |
| Phase 8 | `phase8/p8-relay-program.md` | Relay program policies (`VA-N*` + `VA-R*`) |
| Phase 9 | `phase9/p9-repro-build.md` | Reproducible builds and container evidence (`VA-R*`, `VA-B*`) |
| Phase 10 | `phase10/p10-go-no-go.md` | Integrated go/no-go governance, CLI witness, and open-decision status (`VA-X*`, `VA-H*`, `pkg/v10/scenario`) |

Everything here is deterministic: the tables point to in-repo documents, `pkg/v10` helpers, release manifests, container references, and CLI artifact anchors. Evidence anchors are self-referential when needed to keep the loop closed.
