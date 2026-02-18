# As-Built Conformance

This report maps v19 `F20` specification requirements to v20 artifacts/tests and confirms relay boundary behavior.

| v19 F20 input | v20 implementation evidence | Result |
|---|---|---|
| F20 hardening spec intent (TLS/handshake hardening and operator constraints) | `docs/v1.9/phase4/f20-release-hardening-spec.md`, `pkg/v20/security`, `pkg/v11/relaypolicy`, `tests/e2e/v20/regression_matrix_test.go` | pass |
| Deterministic orchestration (F20 acceptance criterion) | `tests/e2e/v20/regression_matrix_test.go`, `tests/e2e/v20/recovery_paths_test.go`, `docs/v2.0/phase2/p2-regression-report.md` | pass |
| Continuity/recovery behavior (F20 acceptance criterion) | `tests/e2e/v20/recovery_paths_test.go`, `docs/v2.0/phase2/p2-slo-scorecard.md` | pass |
| Chaos/ops readiness assertions (F20 acceptance criterion) | `scripts/v20-podman-scenarios.sh`, `artifacts/generated/v20-podman-scenarios/result-manifest.json`, `docs/v2.0/phase3/p3-podman-scenarios.md` | pass |
| Relay no-data boundary preservation (F20 hardening requirement) | `tests/e2e/v20/regression_matrix_test.go` (`TestRelayNoDataRegression`), `artifacts/generated/v20-evidence/go-test-relay-regression-v20.txt` | pass |

## Relay Boundary Proof

- `TestRelayNoDataRegression` rejects `relaypolicy.PersistenceModeDurableMessageBody` and requires session metadata mode.
- Test verifies `relaypolicy.ForbiddenClasses()` includes `StorageClassDurableMessageBody`.
- Exit evidence is captured in `artifacts/generated/v20-evidence/go-test-relay-regression-v20.txt` and is included as `EV-v20-G8-001` in the evidence index.
