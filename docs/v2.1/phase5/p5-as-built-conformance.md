# As-Built Conformance

> As-built artifact that maps the v21 implementation/tests/gate evidence back to the v20 `F21` seeds now that every mandatory command output has been captured.

| F21 specification input (from v20) | v21 artifact/test | Status | Notes |
|---|---|---|---|
| `docs/v2.0/phase4/f21-acceptance-matrix.md` expectations for encrypted local timeline persistence | Packages under `pkg/v21/store` plus `docs/v2.1/phase1/p1-store-schema.md`, `docs/v2.1/phase1/p1-store-threat-model.md` | verified | Local store regression suite executed (`EV-v21-G2-001`), confirming deterministic corruption, hydration, and retention invariants. |
| `docs/v2.0/phase4/f21-proto-delta.md` hardening requirements | `docs/v2.1/phase4/f22-proto-delta.md` and `proto/aether.proto` additive surfaces | verified | Buf lint/breaking commands (`EV-v21-G1-001`, `EV-v21-G1-002`) affirm additive wire safety for the proto delta. |
| v20 search/persistence regression plan (`docs/v2.0/phase2/p2-regression-report.md`) | `docs/v2.1/phase2/p2-search-contract.md`, `docs/v2.1/phase2/p2-redaction-persistence-contract.md`, `cmd/harmolyn/*`, `tests/e2e/v21`, `tests/perf/v21` | verified | Full test matrix (`EV-v21-G4-001`, `EV-v21-G4-002`, `EV-v21-G4-003`) documents deterministic search/redaction/persistence coverage. |
| Podman persistence and relay boundary proof (v20 Phase 3) | `scripts/v21-persistence-scenarios.sh`, `docs/v2.1/phase3/p3-podman-scenarios.md` | verified | Scenarios and manifest (`EV-v21-G5-001`, `EV-v21-G5-002`, `EV-v21-G8-001`) confirm multi-peer persistence, clear/recover, and relay no-long-history-hosting invariants. |
| Risk & gate artifacts (Phase 5 templates) | `docs/v2.1/phase5/p5-evidence-index.md`, `docs/v2.1/phase5/p5-gate-signoff.md`, `docs/v2.1/phase5/p5-risk-register.md`, `docs/v2.1/phase5/p5-evidence-bundle.md` | verified | Gate-driven commands (`EV-v21-G7-001`, `EV-v21-G7-002`, `EV-v21-G7-003`) plus the updated index/gate docs now anchor the Phase 5 narrative. |

With every gate-linked EV row populated, these entries reflect the as-built state rather than planning placeholders.
