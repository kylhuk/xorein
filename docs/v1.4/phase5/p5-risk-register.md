# Phase 5 Risk Register

| ID | Risk | Mitigation | Evidence |
| --- | --- | --- | --- |
| R14-1 | Call instability across NAT types | Deterministic fallback ladder and regression tests | `tests/e2e/v14/voice_flow_test.go` |
| R14-2 | UX limbo during reconnect | Recovery-first messaging contract in `pkg/v14/ui` | `tests/e2e/v14/reconnect_recovery_test.go` |
| R14-3 | Voice quality regressions | Quality badge tier and perf steps | `tests/perf/v14/call_setup_steps_test.go` |
| R14-4 | Relay boundary regression | pkg/v11/relaypolicy enforcement in e2e scenario | `tests/e2e/v14/voice_flow_test.go` |
