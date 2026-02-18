# Phase 0 - Traceability Matrix

| Requirement | Artifact |
|---|---|
| Signed directory entries with deterministic signing model | `pkg/v18/directory`, `pkg/v18/directory/directory_test.go` |
| Signed indexer snapshots and freshness ordering | `pkg/v18/indexer`, `pkg/v18/indexer/indexer_test.go` |
| Multi-indexer merge, deduplication, and trust warnings | `pkg/v18/discoveryclient`, `pkg/v18/discoveryclient/discoveryclient_test.go` |
| Discovery UX + warning rendering + join stage summary | `pkg/v18/ui/discovery_ui.go`, `pkg/v18/ui/discovery_ui_test.go` |
| Discovery + abuse-path + perf evidence | `tests/e2e/v18/join_abuse_test.go`, `tests/e2e/v18/discovery_integrity_test.go`, `tests/perf/v18/discovery_steps_test.go` |
| Podman scenario coverage and relay boundary regression | `scripts/v18-discovery-scenarios.sh`, `containers/v1.8`, `docs/v1.8/phase3/p3-podman-scenarios.md` |
| v19 specification package | `docs/v1.8/phase4/f19-connectivity-qol-spec.md`, `docs/v1.8/phase4/f19-proto-delta.md`, `docs/v1.8/phase4/f19-acceptance-matrix.md` |
