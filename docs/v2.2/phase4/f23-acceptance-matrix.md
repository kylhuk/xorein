# F23 Acceptance Matrix

The acceptance criteria below are now tied directly to Phase 5 evidence rather than planning placeholders. Each row lists the completed gate, its acceptance targets, and the EV entries that document the as-built results.

| Task | Acceptance criteria | Gate | Status | Evidence |
| --- | --- | --- | --- | --- |
| Document history hardening remedial flows | Description of failure modes plus deterministic playbooks for `missing_backfill` and `stale-history` | G4 | Complete | `EV-v22-G4-001` (`artifacts/generated/v22-evidence/go-test-e2e-v22.txt` + `go-test-perf-v22.txt`) |
| Proto delta review | Compatibility memo showing optional new fields and enum usage | G7 | Complete | `EV-v22-G7-001`, `EV-v22-G7-002` (`buf lint` + `buf breaking`; lint run warns about the deprecated DEFAULT category but passed) |
| Podman scenario proof | Podman scenario run showing deterministic recovery for replay loops | G6 | Complete | `EV-v22-G6-001` (`artifacts/generated/v22-evidence/v22-history-scenarios.txt`) |

Completion of each gate now points reviewers to the concrete EV entries above for verification.
