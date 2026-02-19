# Phase 5 evidence bundle (P5-T1)

## Scope of what is claimed
This file records artifacts produced by the executed v2.6 phase-5 command set and runbook-drill scripts.

## Evidence mapping

| ST | Command | Evidence ID | Artifact | Command result |
| --- | --- | --- | --- | --- |
| ST1 | `buf lint` | EV-v26-G9-001 | `artifacts/generated/v26-evidence/buf-lint.txt` | pass (warning only: deprecated `DEFAULT` lint category in `buf.yaml`) |
| ST1 | `buf breaking --against '.git#branch=v20'` | EV-v26-G10-001 | `artifacts/generated/v26-evidence/buf-breaking.txt` | pass |
| ST2 | `go test ./...` | EV-v26-G9-002 | `artifacts/generated/v26-evidence/go-test-all.txt` | pass |
| ST2 | `go test ./tests/e2e/v26/...` | EV-v26-G9-003 | `artifacts/generated/v26-evidence/go-test-e2e-v26.txt` | pass |
| ST2 | `go test ./tests/perf/v26/...` | EV-v26-G10-002 | `artifacts/generated/v26-evidence/go-test-perf-v26.txt` | pass |
| ST3 | `go build ./cmd/xorein` | EV-v26-G9-004 | `artifacts/generated/v26-evidence/go-build-xorein.txt` | pass |
| ST3 | `go build ./cmd/harmolyn` | EV-v26-G9-005 | `artifacts/generated/v26-evidence/go-build-harmolyn.txt` | pass |
| ST3 | `make check-full` | EV-v26-G10-003 | `artifacts/generated/v26-evidence/make-check-full.txt` | pass |
| ST4 | `scripts/v26-release-drills.sh` | EV-v26-G10-004 | `artifacts/generated/v26-evidence/release-drills.txt`, `artifacts/generated/v26-release-drills/manifest.txt` | pass |
| ST4 | `scripts/v26-repro-build-verify.sh` | EV-v26-G10-005 | `artifacts/generated/v26-evidence/repro-build-verify.txt`, `artifacts/generated/v26-evidence/repro-build/*` | pass |

## Coverage notes
- `go test ./tests/e2e/v26/...` and `go test ./tests/perf/v26/...` passed in this capture and are mandatory ST2 evidence.
- Release drill output confirms v26 e2e/perf/repro/build-readiness checks and runbook presence probes all pass.
- `buf breaking` command executes with deterministic local baseline `--against '.git#branch=v20'`.
