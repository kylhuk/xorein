# Phase 5 evidence index (P5-T1)

## Evidence table

| Evidence ID | Gate | Command | Artifact | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| EV-v25-G9-001 | G9 | `buf lint` | `artifacts/generated/v25-evidence/buf-lint.txt` | pass | Advisory warning: deprecated `DEFAULT` lint category. |
| EV-v25-G9-002 | G9 | `go test ./...` | `artifacts/generated/v25-evidence/go-test-all.txt` | pass | Workspace-wide tests complete. |
| EV-v25-G9-003 | G9 | `go test ./tests/e2e/v25/...` | `artifacts/generated/v25-evidence/go-test-e2e-v25.txt` | pass | E2E package passes (`ok`). |
| EV-v25-G9-004 | G9 | `make check-full` | `artifacts/generated/v25-evidence/make-check-full.txt` | pass | Full local check command completed. |
| EV-v25-G9-005 | G9 | `go build ./cmd/xorein` | `artifacts/generated/v25-evidence/go-build-xorein.txt` | pass | Exit status 0. |
| EV-v25-G9-006 | G9 | `go build ./cmd/harmolyn` | `artifacts/generated/v25-evidence/go-build-harmolyn.txt` | pass | Exit status 0. |
| EV-v25-G10-001 | G10 | `buf breaking --against '.git#branch=v20'` | `artifacts/generated/v25-evidence/buf-breaking.txt` | pass | Deterministic local-baseline compatibility check; no breaking changes detected. |
| EV-v25-G10-002 | G10 | `go test ./tests/perf/v25/...` | `artifacts/generated/v25-evidence/go-test-perf-v25.txt` | pass | Deterministic v25 perf package now exists and passes. |
| EV-v25-G10-003 | G10 | `scripts/v25-blob-scenarios.sh` | `artifacts/generated/v25-evidence/v25-blob-scenarios.txt`, `artifacts/generated/v25-blob-scenarios/manifest.txt`, `artifacts/generated/v25-blob-scenarios/run.log` | pass | Scenario statuses are `pass` for all slugs. |

## Index notes
- No `BLOCKED:ENV` statuses were produced by the scenario command execution.
- `pass` and `fail` statuses in the matrix reflect the captured command exit codes.
