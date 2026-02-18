# Phase 5 - As-Built Conformance

This artifact maps v18 implementation to v17 discovery/indexer requirements in `docs/v1.7/phase4/f18-discovery-spec.md` and `docs/v1.7/phase4/f18-acceptance-matrix.md`.

| Acceptance criterion | v18 evidence | Result |
|---|---|---|
| Deterministic `DirectoryEntry` and signed payload behavior | `pkg/v18/directory`, `pkg/v18/directory/directory_test.go`, `pkg/v18/indexer`, `pkg/v18/indexer/indexer_test.go` | pass |
| Multi-indexer dedupe and trust warning behavior | `pkg/v18/discoveryclient`, `pkg/v18/discoveryclient/discoveryclient_test.go`, `tests/e2e/v18/discovery_integrity_test.go` | pass |
| Join funnel state machine and summary rendering | `pkg/v18/discoveryclient`, `pkg/v18/ui`, `tests/e2e/v18/join_abuse_test.go` | pass |
| Relay boundary preserved in discovery path | `tests/e2e/v18/discovery_integrity_test.go`, `pkg/v11/relaypolicy` | pass |
| Scripted scenario pass/fail manifest | `artifacts/generated/v18-discovery-scenarios/result-manifest.json` | pass |
