# P5 Evidence Bundle

This document now reports the assembled evidence bundle that closed Phase 5 gates for v2.2. Each EV entry references the recorded outputs so reviewers can verify the history hardening, proto compatibility, and governance alignment before signoff.

## Bundle structure

1. **History hardening ops** – Podman scenario logs, e2e/perf suites, and regression playbooks for replay-detection and coverage gaps.
2. **Proto delta validation** – `buf lint`/`buf breaking` outputs plus compatibility rationale for the optional metadata and enum additions.
3. **Risk register closure** – Cross-linked signoffs, warnings, and residual risks that feed into the release recommendation.

## Evidence catalog

| EV ID | Gate | Description | Source artifact | Notes |
| --- | --- | --- | --- | --- |
| `EV-v22-G6-001` | G6 | Podman history/replay scenario log covering offline catch-up, multi-archivist failover, quota refusal, replica healing, and relay regression proof | `artifacts/generated/v22-evidence/v22-history-scenarios.txt` | Deterministic manifest completed with status output for each stage. |
| `EV-v22-G4-001` | G4 | History backfill, retrieval, and coverage-gap suites (e2e + perf) covering integrity and UX invariants | `artifacts/generated/v22-evidence/go-test-e2e-v22.txt`, `artifacts/generated/v22-evidence/go-test-perf-v22.txt` | Reports assert deterministic failure reasons and coverage labels. |
| `EV-v22-G5-002` | G5 | Adversarial/privacy and regression checks plus security scans | `artifacts/generated/v22-evidence/make-check-full.txt` | Logged a non-fatal Trivy warning that `--security-checks` is deprecated; govulncheck/gosec results remain clean. |
| `EV-v22-G7-001` | G7 | Buf lint compatibility report for the proto delta | `artifacts/generated/v22-evidence/buf-lint.txt` | Warning notes the deprecation of the `DEFAULT` category, but Buf continues to accept the config. |
| `EV-v22-G7-002` | G7 | Buf breaking change check | `artifacts/generated/v22-evidence/buf-breaking.txt` | Command completed with the empty stub log. |
| `EV-v22-G8-001` | G8 | CLI/system build verification for `cmd/xorein` | `artifacts/generated/v22-evidence/go-build-xorein.txt` | Build output is empty (pass). |
| `EV-v22-G8-002` | G8 | CLI/system build verification for `cmd/harmolyn` | `artifacts/generated/v22-evidence/go-build-harmolyn.txt` | Build output is empty (pass). |
| `EV-v22-G8-003` | G8 | Full `go test ./...` suite | `artifacts/generated/v22-evidence/go-test-all.txt` | All relevant packages passed (cached). |
| `EV-v22-G8-004` | G8 | `make check-full` regression + scan pipeline | `artifacts/generated/v22-evidence/make-check-full.txt` | Completed with govulncheck/gosec plus the noted Trivy warning. |
| `EV-v22-G9-001` | G9 | Relay no-long-history-hosting regression scenario output | `artifacts/generated/v22-evidence/v22-history-scenarios.txt` | Traceably links the regression boundary output to the Podman stage markers. |
| `EV-v22-G9-002` | G9 | Relay regression manifest | `artifacts/generated/v22-history-scenarios/manifest.txt` | Confirms the status for every stage along the regression boundary and records artifact metadata. |
| `EV-v22-G9-003` | G9 | Relay regression Podman log capture | `artifacts/generated/v22-history-scenarios/run.log` | Podman stdout/stderr proves the no-long-history-hosting regression surface, matching the manifest. |

## Gate reference

Each EV entry links to the responsible gate owner so that signoff can cite the recorded artifact and any associated warnings, such as the Buf DEFAULT deprecation note and the Trivy flag warning in `make check-full`. The catalog now also highlights the EV-v22-G9 rows that tie the relay no-long-history-hosting regression boundary to the scenario output, manifest, and container log artifacts.
