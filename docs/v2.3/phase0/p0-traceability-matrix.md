# Traceability matrix (Phase 0 P0-T1 ST3)

This artifact keeps the “planning vs implementation” boundary explicit: every scope lock requirement maps to a downstream deliverable, and nothing is claimed as implemented yet.

| Scope requirement | Target artifact(s) | Notes / gating context |
| --- | --- | --- |
| Hardening invariants (privacy, quotas, relay boundary, durability state) | `docs/v2.3/phase1/*.md`, `pkg/v23/*/*`, `docs/v2.3/phase0/p0-hardening-matrix.md` | Evidence is produced through Phase 1 work; `EV-v23-G0-###` entries must cite both this matrix and the hardening matrix for traceability. |
| Requirement-to-artifact trace | All downstream docs referenced here | Use this matrix as the single source of truth for `ST3`; update before `G0` review to avoid false completions. |
| Architecture coverage audit requirements | `docs/v2.3/phase0/p0-architecture-coverage-audit.md`, followed by final audit copy `docs/v2.3/phase4/p4-architecture-coverage-audit-result.md` | `G11` expects the P0 artifact to feed the final report. |
| Evidence index schema use | `docs/templates/roadmap-evidence-index.md`, Phase 0 evidence list | Link each `EV-v23-G0-###` entry back to this matrix to show how the gate claim arose. |
| Gate ownership assignments | `docs/v2.3/phase0/p0-gate-ownership.md` | Ownership needed before `G0` approval. |
| Missing persistence classes flagged for `F24` seeds | `docs/v2.3/phase4/f24-*` + deferral register | This row remains open until the audit lists every class or explicitly defers it. |

Keep this matrix updated as Phase 0 artifacts evolve; any row that lacks a downstream artifact must be flagged as `BLOCKED:INPUT`. The matrix itself is planning evidence for `G0` and must be frozen (ST3) before `G0` can pass.
