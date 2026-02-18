# Phase 3 Podman Scenarios

- Scenario 1: Space creation + invite-only join in Podman network; script runs e2e/perf tests to prove join policy enforcement.
- Scenario 2: Request-to-join moderation path exercised; manifest captures deterministic status.
- Scenario 3: Relay wariness assertion reuses `pkg/v11/relaypolicy`; framework ensures storage classes beyond session metadata stay forbidden.
- Result manifest `artifacts/generated/v13-e2e-podman/result-manifest.json` serves as evidence for G5.
