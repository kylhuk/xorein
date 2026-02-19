# P5 Gate Signoff

This artifact now tracks the completed gate signoffs for v2.2. Each gate references the specific EV entries that closed it so the review board can trace the required commands, Podman outputs, and security scans without ambiguity.

## Gate checklist

| Gate | Signoff owners | Evidence | Status | Notes |
| --- | --- | --- | --- | --- |
| G4 | Client & UX lead, search/backfill QA lead | `EV-v22-G4-001` (`go test ./tests/e2e/v22/...` + `go test ./tests/perf/v22/...`) | Signed | History hardening coverage proves deterministic failure reasons, in-flight backfill completeness, and coverage-gap telemetry. |
| G5 | Adversarial test lead, security & privacy board | `EV-v22-G5-001` (`go test ./tests/e2e/v22/...`), `EV-v22-G5-002` (`go test ./tests/perf/v22/...`) | Signed | QoL/resilience tests exercise quota refusals, replay controls, and performance baselines with the same binaries used for G4. |
| G6 | Podman scenario owner, reliability engineer | `EV-v22-G6-001` (`scripts/v22-history-scenarios.sh`), `EV-v22-G6-002` (manifest output), `EV-v22-G6-003` (run log), `EV-v22-G6-004` (`containers/v2.2/scenarios.conf`) | Signed | Offline catch-up, heal, quota-refusal, and relay regression proof recorded with manifest, log, and scenario-definition artifacts. |
| G7 | Protocol review board | `EV-v22-G7-001`, `EV-v22-G7-002` (`buf lint`, `buf breaking` with the DEFAULT category deprecation warning) | Signed | Compatibility memo confirms optional fields and `HardeningStatus` behavior remain additive. |
| G8 | Evidence bundle curator, governance board | `EV-v22-G8-001`, `EV-v22-G8-002`, `EV-v22-G8-003`, `EV-v22-G8-004` (see `p5-evidence-index` for command/checksum mapping) | Signed | Includes `make check-full` scan pipeline; Trivy warns `--security-checks` is deprecated but the scan finished. |
| G9 | Relay regression owner, reliability engineer | `EV-v22-G9-001`, `EV-v22-G9-002`, `EV-v22-G9-003` (`scripts/v22-history-scenarios.sh` outputs) | Signed | Relay no-long-history-hosting regression boundary is recorded across the manifest, log, and scenario output artifacts. |

## Gate-level instructions

- Collect command traces and test logs before requesting G6 approval.
- Attach buf lint/breaking outputs plus compatibility rationale to the G7 evidence packet (DEFAULT warning noted but non-fatal).
- Document the evidence index references once the bundle entries are complete for G8 (the index now maps every mandatory command to path and checksum, including `make check-full`'s Trivy warning).
