# Phase 5 evidence bundle (P5-T1)

## Scope of what is claimed
This file records only artifacts produced by the executed commands below and the v25 blob scenario harness output. It does not claim success for any command that produced non-zero exit status.

## Evidence mapping

| ST | Command | Evidence ID | Artifact | Command result |
| --- | --- | --- | --- | --- |
| ST1 | `buf lint` | EV-v25-G9-001 | `artifacts/generated/v25-evidence/buf-lint.txt` | pass (warning: deprecated `DEFAULT` category in `buf.yaml`) |
| ST1 | `buf breaking --against '.git#branch=v20'` | EV-v25-G10-001 | `artifacts/generated/v25-evidence/buf-breaking.txt` | pass (no breaking changes detected) |
| ST2 | `go test ./...` | EV-v25-G9-002 | `artifacts/generated/v25-evidence/go-test-all.txt` | pass |
| ST2 | `go test ./tests/e2e/v25/...` | EV-v25-G9-003 | `artifacts/generated/v25-evidence/go-test-e2e-v25.txt` | pass |
| ST2 | `go test ./tests/perf/v25/...` | EV-v25-G10-002 | `artifacts/generated/v25-evidence/go-test-perf-v25.txt` | pass |
| ST3 | `go build ./cmd/xorein` | EV-v25-G10-003 | `artifacts/generated/v25-evidence/go-build-xorein.txt` | pass |
| ST3 | `go build ./cmd/harmolyn` | EV-v25-G10-004 | `artifacts/generated/v25-evidence/go-build-harmolyn.txt` | pass |
| ST3 | `make check-full` | EV-v25-G9-004 | `artifacts/generated/v25-evidence/make-check-full.txt` | pass |
| ST3 | `scripts/v25-blob-scenarios.sh` | EV-v25-G9-005 | `artifacts/generated/v25-evidence/v25-blob-scenarios.txt`, `artifacts/generated/v25-blob-scenarios/manifest.txt`, `artifacts/generated/v25-blob-scenarios/run.log` | pass |

## Coverage notes
- Scenario harness output covers v25 e2e-level BlobRef, transfer, crypto, replication, and relay-boundary probes and emits deterministic statuses in the manifest.
- `go test ./tests/e2e/v25/...` and `go test ./tests/perf/v25/...` are command-mandated evidence items and both succeed in this capture.
