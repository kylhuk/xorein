# Rollback Drill

- Triggered rollback after simulated pod restart: `podman pod stop v20-operator` followed by `podman pod rm -f v20-operator` and `podman pod create`.
- Confirmed relay continuity by rerunning `tests/e2e/v20/recovery_paths_test.go` to detect regressions.
- Evidence captured in `artifacts/generated/v20-podman-scenarios/result-manifest.json`.
