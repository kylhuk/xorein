# v2.2 Phase 4 Podman Scenarios

## ST1–ST6 mapping
| Scenario | ST mapping | Description |
| --- | --- | --- |
| offline-catchup | ST1 | Simulate a disconnected client that relies on backfill progress after reconnecting. |
| multi-archivist-failover | ST2 | Verify replication handles a primary archivist failure and still produces a degraded health signal. |
| quota-refusal | ST3 | Confirm quota refusals occur with deterministic reasons when an archivist rejects writes. |
| replica-healing | ST4 | Trigger healing so missing replica tokens are reconciled when new candidates register. |
| relay-no-history-hosting | ST5 | Assert relays do not host history segments/manifests and that the harness reports failures without leakage. |
| _manifest sanity check_ | ST6 | The harness emits `artifacts/generated/v22-history-scenarios/manifest.txt` and `run.log` with pass/fail/BLOCKED:ENV semantics. |

## Command sequence
`./scripts/v22-history-scenarios.sh` is the deterministic harness: it reads `containers/v2.2/scenarios.conf`, boots a `docker.io/library/golang:1.24.12` Podman container for each row, then runs `go test <package> -run <pattern> -count=1 -v`. Each probe writes its log alongside the generated manifest/log under `artifacts/generated/v22-history-scenarios`.

## Manifest schema
The manifest file is YAML and includes:
- `suite`: `v22-history-scenarios`
- `generated_at`: UTC timestamp
- `podman_image`: container image used
- `podman_available`, `podman_blocker`: environment status
- `scenarios`: list of probe objects with `slug`, `status`, `package`, `test_pattern`, `actual_exit_code`, `required_output`, `failure_reason`, `command`, and `log`.

## Evidence references
- `EV-v22-G6-001`: `scripts/v22-history-scenarios.sh` drives the Podman harness and records `artifacts/generated/v22-evidence/v22-history-scenarios.txt` (checksum `dcb405683feb3ed54f9f112866df95453204b07c736c7534eff7a7ca8b70ca98`).
- `EV-v22-G6-002`: `artifacts/generated/v22-history-scenarios/manifest.txt` (checksum `d3b3445065537523057ce8b2fc18001f80d9b35cd4342c647a17bd5700220592`) enumerates ST1–ST6 statuses, commands, and coverage metadata.
- `EV-v22-G6-003`: `artifacts/generated/v22-history-scenarios/run.log` (checksum `516af4dba22d8f05460a88564fb0e24f1f5326d7c4862e75c1aa26c7126eb80f`) preserves every Podman container output for replay and debugging.
- `EV-v22-G6-004`: `containers/v2.2/scenarios.conf` (checksum `c37cfae1bb883e2d475211bb254f6907dfebbf30ca369872e1a4956a7acc68a2`) lists the deterministic scenario permutations that seed the harness.
