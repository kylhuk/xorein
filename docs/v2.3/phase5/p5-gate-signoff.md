# Phase 5 gate signoff checklist (P5-T1 ST1–ST4)

| Item | Owner | Gate `G7` criteria | Status |
| --- | --- | --- | --- |
| ST1 regression harness manifest captured and checksum recorded in `p5-evidence-index.md`. | Ops Lead | Confirms Podman scenarios executed for offline catch-up, live, degraded, and recovery coverage. | pass (EV-v23-G7-009) |
| ST2 E2E scenario suite passes with deterministic metadata consumed by `scripts/v23-regression-scenarios.sh`. | QA Lead | Ensures `G7` sees reproducible E2E regression evidence. | pass (EV-v23-G7-004) |
| ST3 perf regression logs appended to `p5-evidence-index.md`. | QA Lead | Verifies QoS/latency envelope described in Phase 5 planning. | pass (EV-v23-G7-005) |
| ST4 lint/build suite (`buf lint`, `buf breaking`, `go build`, `make check-full`) completes without lint-breaking output. | Plan Lead | Proves `G7` gate surfaces correct compile/test hygiene. | pass (EV-v23-G7-001, EV-v23-G7-002, EV-v23-G7-006, EV-v23-G7-007, EV-v23-G7-008; warnings were advisory) |
| Relay boundary harness (dedicated relay-history test + scenario manifest) documents `G8` invariants. | Security Lead | Confirms relays store no long-term state while offline/continuity/resilience scenarios also pass. | pass (EV-v23-G8-001, EV-v23-G8-002) |
| As-built/spec inputs and narrative review validate the `G9` posture. | Architecture Lead | Ensures the executed posture comments on the same spec packages described in the plan. | pass (EV-v23-G9-001, EV-v23-G9-002; review is document-only, no command run) |

## Additional approvals
- `p5-final-evidence-bundle.md`, `p5-as-built-conformance.md`, and `p5-evidence-index.md` must be reviewed together for consistency before `G7` status transitions from PENDING.
- `p5-final-evidence-bundle.md`, `p5-as-built-conformance.md`, `p5-evidence-index.md`, and `p5-go-no-go-record.md` now document that G7, G8, and G9 are satisfied (warnings remain advisory).
