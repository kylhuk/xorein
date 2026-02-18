# v1.2 Phase 3 - Podman Recovery Scenario Baseline

## Status
Implemented via `scripts/v12-recovery-scenarios.sh` and `containers/v1.2/*`.

## Deterministic probes
| Probe | Command | Expected |
|---|---|---|
| New-device restore | `go test ./tests/e2e/v12 -run TestNewDeviceRestoreScenario -count=1 -v` | pass |
| Lost password path | `go test ./tests/e2e/v12 -run TestLostPasswordWithoutBackupScenario -count=1 -v` | pass (deterministic `backup-password`) |
| Backup tamper detection | `go test ./pkg/v12/backup -run TestRestoreDetectsTamperedCiphertext -count=1 -v` | pass (deterministic `backup-corrupt`) |
| Relay boundary regression | `go test ./tests/e2e/v12 -run TestRelayBoundaryRegressionScenario -count=1 -v` | pass |

## Manifest contract
- JSON manifest path: `artifacts/generated/v12-recovery-scenarios/result-manifest.json`.
- Each probe records expected/actual exit code, status, failure reason, required output text, and log path.

## Planned vs implemented
- Script and container baseline are implemented.
- Promotion evidence is finalized in phase5.
