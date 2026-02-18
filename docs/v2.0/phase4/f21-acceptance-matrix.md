# F21 Acceptance Matrix

| Feature | Criterion | Evidence |
| --- | --- | --- |
| Relay hardening | No data persistence | `tests/e2e/v20/regression_matrix_test.go` |
| Runtime continuity | Continuity policy passes | `tests/e2e/v20/recovery_paths_test.go` |
| Operator readiness | Podman scenarios | `scripts/v20-podman-scenarios.sh` |
