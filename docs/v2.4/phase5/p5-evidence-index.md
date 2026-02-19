# Phase 5 evidence index (P5-T1 ST1–ST4)

## ST coverage summary
- **ST1**: `scripts/v24-daemon-scenarios.sh` exercises the multi-client crash/recovery matrix plus stale socket repair; manifest/log preserve the Podman/desktop Run details required by G5/G6.
- **ST2**: `tests/e2e/v24/...` verifies the API-only journeys (identity, spaces, history, attach flows).
- **ST3**: `tests/perf/v24/...` guards QoS targets while `go test ./...` covers the workspace.
- **ST4**: Lint/build hygiene (`buf lint`, `buf breaking`, `go build` for both binaries, `make check-full`) plus the boundary enforcement script capture CI expectations and G9 invariants.

## Command evidence matrix (Gate `G8`)
| Evidence ID | Gate | Command | Artifact | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `EV-v24-G8-001` | G8 | `buf lint` | `artifacts/generated/v24-evidence/buf-lint.txt` | pass | Advisory warning about deprecated `DEFAULT` name only. |
| `EV-v24-G8-002` | G8 | `buf breaking --against '.git#branch=origin/dev'` | `artifacts/generated/v24-evidence/buf-breaking.txt` | pass | Compatibility check vs. the living baseline. |
| `EV-v24-G8-003` | G8 | `go test ./...` | `artifacts/generated/v24-evidence/go-test-all.txt` | pass | Workspace-wide regression signal. |
| `EV-v24-G8-004` | G8 | `go test ./tests/e2e/v24/...` | `artifacts/generated/v24-evidence/go-test-e2e-v24.txt` | pass | ST2 deterministic e2e journey coverage. |
| `EV-v24-G8-005` | G8 | `go test ./tests/perf/v24/...` | `artifacts/generated/v24-evidence/go-test-perf-v24.txt` | pass | ST3 performance reference run. |
| `EV-v24-G8-006` | G8 | `go build ./cmd/xorein` | `artifacts/generated/v24-evidence/go-build-xorein.txt` | pass | Backend binary build hygiene. |
| `EV-v24-G8-007` | G8 | `go build ./cmd/harmolyn` | `artifacts/generated/v24-evidence/go-build-harmolyn.txt` | pass | UI binary build hygiene. |
| `EV-v24-G8-008` | G8 | `make check-full` | `artifacts/generated/v24-evidence/make-check-full.txt` | pass | Non-blocking Trivy flag warning; otherwise clean. |
| `EV-v24-G8-009` | G8 | `scripts/v24-daemon-scenarios.sh` | `artifacts/generated/v24-daemon-scenarios/manifest.txt` | pass | Primary scenario manifest for G5/G6; companion run log at `artifacts/generated/v24-daemon-scenarios/run.log` and command wrapper output at `artifacts/generated/v24-evidence/v24-daemon-scenarios.txt`. |

## Boundary and dependency evidence (Gate `G9`)
| Evidence ID | Gate | Command | Artifact | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `EV-v24-G9-001` | G9 | `scripts/ci/enforce-boundaries.sh` | `artifacts/generated/v24-evidence/enforce-boundaries.txt` | pass | Enforces Gio-free backend packages and keeps `cmd/harmolyn` tied to the local API. |

## G5/G6 scenario trace notes
- Podman/desktop scenario manifest (`artifacts/generated/v24-daemon-scenarios/manifest.txt`) and run log (`artifacts/generated/v24-daemon-scenarios/run.log`) are referenced by EV-v24-G8-009 and demonstrate the multi-client, crash, and stale socket scenarios required by Gates `G5` and `G6`. Include them when reviewing the manifest ID list referenced in `docs/v2.4/phase3/p3-daemon-scenarios.md`.

## Artifact hygiene
- All artifacts live under `artifacts/generated/v24-evidence/` (plus the additional daemon scenario directory) so automation can fetch them. Replace any future “pass” entries with updated outputs if the suite is rerun before release.
