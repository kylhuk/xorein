# Phase 5 evidence index (P5-T1)

## Evidence table

| Evidence ID | Gate | Command | Artifact | TimestampUTC | Owner | Status | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| EV-v26-G9-001 | G9 | `buf lint` | `artifacts/generated/v26-evidence/buf-lint.txt` | 2026-02-19T11:37:56Z | QA lead | pass | Warning only: deprecated `DEFAULT` rule category in `buf.yaml`. |
| EV-v26-G9-002 | G9 | `go test ./...` | `artifacts/generated/v26-evidence/go-test-all.txt` | 2026-02-19T11:37:56Z | QA lead | pass | Workspace-wide test suite completed. |
| EV-v26-G9-003 | G9 | `go test ./tests/e2e/v26/...` | `artifacts/generated/v26-evidence/go-test-e2e-v26.txt` | 2026-02-19T11:37:56Z | QA lead | pass | E2E package pass (`ok`). |
| EV-v26-G9-004 | G9 | `go build ./cmd/xorein` | `artifacts/generated/v26-evidence/go-build-xorein.txt` | 2026-02-19T11:37:56Z | Release engineer | pass | Build command completed with exit 0. |
| EV-v26-G9-005 | G9 | `go build ./cmd/harmolyn` | `artifacts/generated/v26-evidence/go-build-harmolyn.txt` | 2026-02-19T11:37:56Z | Release engineer | pass | Build command completed with exit 0. |
| EV-v26-G10-001 | G10 | `buf breaking --against '.git#branch=v20'` | `artifacts/generated/v26-evidence/buf-breaking.txt` | 2026-02-19T11:37:56Z | Protocol lead | pass | Deterministic local-baseline compatibility check. |
| EV-v26-G10-002 | G10 | `go test ./tests/perf/v26/...` | `artifacts/generated/v26-evidence/go-test-perf-v26.txt` | 2026-02-19T11:37:56Z | QA lead | pass | Performance package pass (`ok`). |
| EV-v26-G10-003 | G10 | `make check-full` | `artifacts/generated/v26-evidence/make-check-full.txt` | 2026-02-19T11:37:56Z | QA lead | pass | `make` check command completed; includes scan and lint steps. |
| EV-v26-G10-004 | G10 | `scripts/v26-release-drills.sh` | `artifacts/generated/v26-evidence/release-drills.txt`, `artifacts/generated/v26-release-drills/manifest.txt` | 2026-02-19T11:37:56Z | Ops lead | pass | All drills passed with `overall_status: pass`. |
| EV-v26-G10-005 | G10 | `scripts/v26-repro-build-verify.sh` | `artifacts/generated/v26-evidence/repro-build-verify.txt`, `artifacts/generated/v26-evidence/repro-build/*` | 2026-02-19T11:37:56Z | Release engineer | pass | Deterministic rebuild + checksums captured. |

## Index notes
- No `BLOCKED:ENV` outcomes were produced by P5 command execution.
- Result column reflects command exit status mapping only (`pass` in this run).
