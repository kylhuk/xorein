# v2.3 Phase 3 Podman Scenarios

## ST1-ST5 mapping

| Scenario | ST mapping | Test pattern |
| --- | --- | --- |
| offline-catchup | ST1 | `TestScenarioOfflineCatchup` |
| redaction-tombstone | ST2 | `TestScenarioRedactionTombstoneRegression` |
| private-space-anti-enumeration | ST3 | `TestScenarioPrivateSpaceAntiEnumeration` |
| replica-healing | ST4 | `TestScenarioReplicaHealingUnderChurn` |
| relay-no-history-hosting | ST5 | `TestScenarioRelayNoHistoryHosting` |

## Command sequence

1. `./scripts/v23-regression-scenarios.sh`
2. script reads `containers/v2.3/scenarios.conf`
3. script launches each test in Podman with `go test ./tests/e2e/v23 -run <pattern> -count=1 -v`
4. generated artifacts are written under `artifacts/generated/v23-regression-scenarios`

## Harness output

- command: `./scripts/v23-regression-scenarios.sh`
- manifest: `artifacts/generated/v23-regression-scenarios/manifest.txt`
- run log: `artifacts/generated/v23-regression-scenarios/run.log`
- per scenario logs: `artifacts/generated/v23-regression-scenarios/<slug>.log`

## Manifest schema

The manifest is YAML with fields: `suite`, `generated_at`, `podman_image`, `podman_available`,
`podman_blocker`, `artifact_dir`, `log`, and `scenarios`.
Each scenario entry includes `slug`, `status`, `package`, `test_pattern`, `expected_exit_code`,
`actual_exit_code`, `required_output`, `failure_reason`, `command`, and `log`.

`status` values are `pass`, `fail`, or `BLOCKED:ENV`.

## Evidence references

- `EV-v23-G5-001`: command run output for `./scripts/v23-regression-scenarios.sh`
- `EV-v23-G5-002`: `artifacts/generated/v23-regression-scenarios/manifest.txt` with scenario coverage
  and status
- `EV-v23-G5-003`: `artifacts/generated/v23-regression-scenarios/run.log` (Podman probe command history)
- `EV-v23-G5-004`: `containers/v2.3/scenarios.conf` (scenario matrix seed)
