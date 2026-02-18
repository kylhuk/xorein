# Regression Report

- Regression matrix run by `tests/e2e/v20/regression_matrix_test.go` ensures security controls, relay no-data boundary, and continuity hardened release paths remain deterministic.
- Recovery paths test `tests/e2e/v20/recovery_paths_test.go` covers restart/downtime expectations.
- No additional regressions observed compared to v19 baseline.
