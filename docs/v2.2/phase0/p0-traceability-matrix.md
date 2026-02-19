# Phase 0 traceability matrix (G0)

| Critical scope item | Planned artifact targets | Validation artifacts / gates |
| --- | --- | --- |
| Archivist capability (advertisement, quota/retention, deterministic refusals, replication policy). | `pkg/v22/archivist/advertise/*`, `pkg/v22/archivist/store/*`, `pkg/v22/archivist/replicate/*`, `docs/v2.2/phase1/p1-archivist-selection-contract.md`, `docs/v2.2/phase1/p1-archivist-quota-retention.md`, `docs/v2.2/phase1/p1-replication-contract.md`. | `go test ./pkg/v22/archivist/...`, `EV-v22-G2-###`, `G2` gate sign-off. |
| History retrieval head/manifest model and ciphertext-only endpoints (with manifest verification and anti-enumeration). | `pkg/v22/history/integrity/*`, `pkg/v22/history/retrieve/*`, `docs/v2.2/phase2/p2-integrity-reason-taxonomy.md`, `docs/v2.2/phase2/p2-private-space-anti-enumeration.md`. | `go test ./pkg/v22/history/...`, `EV-v22-G3-###`, `G3`. |
| Client backfill pipeline, redaction application, and harmolyn search coverage UX. | `pkg/v22/history/apply/*`, `pkg/v22/history/backfill/*`, `tests/e2e/v22/backfill_*`, `tests/e2e/v22/redaction_backfill_*`, `cmd/harmolyn/*`, `docs/v2.2/phase2/p2-redaction-backfill-contract.md`, `docs/v2.2/phase3/p3-history-search-ux-contract.md`. | `go test ./pkg/v22/history/...`, `go test ./cmd/harmolyn/...`, `EV-v22-G4-###`, `G4`. |
| Relay no-long-history-hosting boundary + Podman scenario coverage. | `containers/v2.2/*`, `scripts/v22-history-scenarios.sh`, `docs/v2.2/phase4/p4-podman-scenarios.md`, `tests/e2e/v22/*`, `tests/perf/v22/*`. | Scenario outputs (`scripts/v22-history-scenarios.sh`), `EV-v22-G6-###`, `EV-v22-G9-###`, `G6`/`G9`. |

Each artifact listed above provides the evidence referenced in `TODO_v22.md` (Phase 1-4 gate artifacts) and feeds into the `EV-v22-GX-###` entries required during Phase 5 evidence bundling.
