# Phase 5 evidence index (P5-T1 ST1–ST4)

## ST1–ST4 coverage summary
- **ST1**: Regression scenarios covering Podman harness and offline/continuity recovery permutations (`scripts/v23-regression-scenarios.sh`).
- **ST2**: Deterministic e2e suite (`tests/e2e/v23`) consuming the scenario catalog (`containers/v2.3/scenarios.conf`).
- **ST3**: Performance stability assertions from `tests/perf/v23` runs aligned with the v2.3 SLO profile.
- **ST4**: Build/lint hygiene commands (`buf`/Go builds/make check) that guard the binaries in `cmd/xorein` and `cmd/harmolyn`.

## Command evidence matrix (Gate `G7`)
| Evidence ID | Gate | Command | Artifact | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `EV-v23-G7-001` | G7 | `buf lint` | `artifacts/generated/v23-evidence/buf-lint.txt` | pass | ST4 lint/build hygiene (advisory `DEFAULT` warning only). |
| `EV-v23-G7-002` | G7 | `buf breaking --against '.git#branch=origin/dev'` | `artifacts/generated/v23-evidence/buf-breaking.txt` | pass | ST4 compatibility check against the v2.3 baseline (no output logged). |
| `EV-v23-G7-003` | G7 | `go test ./...` | `artifacts/generated/v23-evidence/go-test-all.txt` | pass | ST4 regression coverage across the entire workspace. |
| `EV-v23-G7-004` | G7 | `go test ./tests/e2e/v23/... -count=1` | `artifacts/generated/v23-evidence/go-test-e2e-v23.txt` | pass | ST2 deterministic e2e suite consumed the scenario catalog. |
| `EV-v23-G7-005` | G7 | `go test ./tests/perf/v23/... -count=1` | `artifacts/generated/v23-evidence/go-test-perf-v23.txt` | pass | ST3 performance regression suite confirmed SLOs. |
| `EV-v23-G7-006` | G7 | `go build ./cmd/xorein` | `artifacts/generated/v23-evidence/go-build-xorein.txt` | pass | ST4 backend binary build hygiene. |
| `EV-v23-G7-007` | G7 | `go build ./cmd/harmolyn` | `artifacts/generated/v23-evidence/go-build-harmolyn.txt` | pass | ST4 relay binary build hygiene. |
| `EV-v23-G7-008` | G7 | `make check-full` | `artifacts/generated/v23-evidence/make-check-full.txt` | pass | ST1/ST4 integrated checks (trivy deprecated-flag warning was advisory). |
| `EV-v23-G7-009` | G7 | `./scripts/v23-regression-scenarios.sh` | `artifacts/generated/v23-evidence/v23-regression-scenarios.txt` | pass | ST1 regression harness (manifest/logs under `artifacts/generated/v23-regression-scenarios/`). |

## Relay boundary evidence (Gate `G8`)
| Evidence ID | Gate | Command | Artifact | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `EV-v23-G8-001` | G8 | `go test ./tests/e2e/v23 -run TestScenarioRelayNoHistoryHosting -count=1 -v` | `artifacts/generated/v23-evidence/go-test-relay-boundary.txt` | pass | Dedicated relay boundary check proves relays host no long-term history. |
| `EV-v23-G8-002` | G8 | `./scripts/v23-regression-scenarios.sh` (relay boundary scenarios) | `artifacts/generated/v23-evidence/v23-regression-scenarios.txt` + `artifacts/generated/v23-regression-scenarios/manifest.txt` | pass | Podman harness logs and manifest document the relay boundary scenario plus offline/continuity coverage. |

## As-built/spec evidence (Gate `G9`)
| Evidence ID | Gate | Command | Artifact | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `EV-v23-G9-001` | G9 | N/A (spec input package list compiled for reviewers) | `artifacts/generated/v23-evidence/f23-spec-inputs.txt` | recorded | Provides the F23 spec input package set for reviewer intake and regression traceability. |
| `EV-v23-G9-002` | G9 | N/A (as-built conformance review) | `docs/v2.3/phase5/p5-as-built-conformance.md` | reviewed | Narrative review confirms the documented as-built posture matches the ST1–ST4 command bundle and relay boundary expectations. |
