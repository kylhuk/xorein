# v2.4 Phase 3 Daemon Scenarios

## ST1-ST3 mapping

| Scenario | ST mapping | Test pattern |
| --- | --- | --- |
| dual-client-parallel-reads | ST1 | `TestScenarioDualClientsParallelReadsAndSerializedWrites` |
| crash-mid-call-recovery | ST2 | `TestScenarioDaemonCrashMidCallReattach` |
| stale-socket-repair | ST3 | `TestScenarioStaleSocketAutoRepair` |

## Command sequence

1. `./scripts/v24-daemon-scenarios.sh`
2. script reads `containers/v2.4/scenarios.conf`
3. script launches each test in Podman with `go test ./tests/e2e/v24 -run <pattern> -count=1 -v`
4. generated artifacts are written under `artifacts/generated/v24-daemon-scenarios`

## Harness output

- command: `./scripts/v24-daemon-scenarios.sh`
- manifest: `artifacts/generated/v24-daemon-scenarios/manifest.txt`
- run log: `artifacts/generated/v24-daemon-scenarios/run.log`
- per scenario logs: `artifacts/generated/v24-daemon-scenarios/<slug>.log`

## Manifest schema

The manifest is YAML with fields: `suite`, `generated_at`, `podman_image`, `podman_available`,
`podman_blocker`, `artifact_dir`, `log`, and `scenarios`.
Each scenario entry includes `slug`, `status`, `package`, `test_pattern`,
`expected_exit_code`, `actual_exit_code`, `required_output`, `failure_reason`, `command`, and `log`.

`status` values are `pass`, `fail`, or `BLOCKED:ENV`.

## Evidence references

- `EV-v24-G5-001`: command run output for `./scripts/v24-daemon-scenarios.sh`
- `EV-v24-G5-002`: `artifacts/generated/v24-daemon-scenarios/manifest.txt` with scenario coverage and status
- `EV-v24-G5-003`: `artifacts/generated/v24-daemon-scenarios/run.log` (Podman probe command history)
- `EV-v24-G5-004`: `containers/v2.4/scenarios.conf` (scenario matrix seed)
- `EV-v24-G6-001`: Podman/desktop harness parity evidence using `artifacts/generated/v24-daemon-scenarios/manifest.txt` + `run.log`
- `EV-v24-G6-002`: per-scenario logs under `artifacts/generated/v24-daemon-scenarios/<slug>.log` for desktop operator replay
