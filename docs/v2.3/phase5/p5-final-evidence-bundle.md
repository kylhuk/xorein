# Phase 5 final evidence bundle (P5-T1 ST1–ST4)

## Summary
- Captures the Phase 5 closure promises for ST1–ST4 now backed by the EV-v23-G7-001 through EV-v23-G7-009 artifacts plus the newly documented EV-v23-G8-001/EV-v23-G8-002 relay boundary checks and EV-v23-G9-001/EV-v23-G9-002 as-built/spec package traces.
- Aligns with Gates `G7`, `G8`, and `G9` by documenting the required command/regression/relay-boundary outputs along with the as-built review artifacts while keeping the existing advisory-warning note for non-blocking lint/trivy output.

## ST1–ST4 mapping
| ST | Focus | Evidence | Status |
| --- | --- | --- | --- |
| ST1 | Regression scenario coverage | Podman/log manifests covering offline catch-up, live stream, degraded continuity, and recovery recorded under `artifacts/generated/v23-regression-scenarios/`. | PASS (EV-v23-G7-009) |
| ST2 | E2E regression signal | Deterministic verification of the `tests/e2e/v23` suite against the latest scenario catalog (`artifacts/generated/v23-evidence/go-test-e2e-v23.txt`). | PASS (EV-v23-G7-004) |
| ST3 | Perf margin confirmation | `tests/perf/v23` deterministic check (`artifacts/generated/v23-evidence/go-test-perf-v23.txt`). | PASS (EV-v23-G7-005) |
| ST4 | Build and lint hygiene | `buf lint`, `buf breaking --against '.git#branch=origin/dev'`, `go build` for both binaries, plus `make check-full` (`artifacts/generated/v23-evidence/*.txt`). | PASS (EV-v23-G7-001, EV-v23-G7-002, EV-v23-G7-006, EV-v23-G7-007, EV-v23-G7-008) |

## Gate mapping
- **G7**: The final Phase 5 go/no-go gate. This document captures the ST-to-evidence mapping that must be satisfied before G7 can close.
- **G8**: Relay boundary assurances are derived from the dedicated boundary command output (`artifacts/generated/v23-evidence/go-test-relay-boundary.txt`) plus the Podman scenario manifest/log (`artifacts/generated/v23-regression-scenarios/manifest.txt`).
- **G9**: As-built/spec conformance relies on the documented spec input package list (`artifacts/generated/v23-evidence/f23-spec-inputs.txt`) and the narrative review captured in `p5-as-built-conformance.md`.
- **G7 Evidence knot**: Final status is driven by the command/manifest matrix recorded in `p5-evidence-index.md` plus the `G7` checklist in `p5-gate-signoff.md`, the as-built confirmations in `p5-as-built-conformance.md`, and the G8/G9 artifacts noted above.

-## Next steps
- All command outputs are captured; confirm reviewers know which EV rows to reference. Warnings were limited to the deprecated `DEFAULT` category name in `buf lint` and the non-blocking `trivy` flag warning in `make check-full`.
- Flag the relay boundary manifest/relay scenario logs and the spec input list for the G8/G9 reviewers so `p5-go-no-go-record.md` can reflect the GO decision with full gate traceability.
