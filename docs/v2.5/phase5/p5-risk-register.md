# Phase 5 risk register (P5-T1)

| ID | Risk | Mitigation | Evidence / Exit criterion |
| --- | --- | --- | --- |
| R25-1 | Blob anti-enumeration regression in private Spaces. | Keep anti-enumeration tests in the scenario and e2e harness as the regression safety net. | `EV-v25-G9-003` (`go test ./tests/e2e/v25/...`) includes `TestPrivateSpaceAntiEnumeration`; manifest entry `relay-boundary-private-anti-enum` is pass. |
| R25-2 | Storage/transfer DoS or unbounded behavior in provider protocol. | Exercise manifest and transfer refusal/acceptance paths in targeted harness scenarios. | `EV-v25-G9-005` and manifest scenarios `blob-transfer-success` and `blob-transfer-refusals` are still required for the next review cycle. |
| R25-3 | Blob metadata/proto compatibility breakage. | Require `buf lint` and `buf breaking` artifacts on every evidence refresh. | `EV-v25-G9-001` is advisory pass; `EV-v25-G10-001` now passes against a local deterministic baseline (`--against '.git#branch=v20'`) with no breaking changes. |
| R25-4 | Relay no-durable-blob-hosting regression. | Keep relay-boundary tests and scenario coverage in the command matrix. | `EV-v25-G9-003` and manifest slug `relay-boundary-private-anti-enum` remain passing for this run. |
| R25-5 | Missing package surface leaves performance obligations unverified. | Keep `tests/perf/v25` present with deterministic checks and rerun on every phase-5 evidence refresh. | `EV-v25-G10-002` (`go test ./tests/perf/v25/...`) now passes and is archived in `artifacts/generated/v25-evidence/go-test-perf-v25.txt`. |

## Residual risk disposition
- R25-1 and R25-4 are currently mitigated by passing v25 e2e coverage and scenario entries.
- R25-2 remains under ongoing regression observation; R25-5 is closed for this evidence cycle.
