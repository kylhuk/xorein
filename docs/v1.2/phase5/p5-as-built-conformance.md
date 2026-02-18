# v1.2 Phase 5 - F12 As-Built Conformance

## Status
Conformance checklist against v1.1 `F12` specification package.

## Scope conformance
- Immutable identity lifecycle: implemented in `pkg/v12/identity/`.
- Local backup export/import with deterministic reasons: implemented in `pkg/v12/backup/`.
- No-password-reset user contract: implemented in `pkg/ui/shell.go` and tested.
- Additive proto metadata for identity/backup: implemented in `proto/aether.proto`.
- Recovery, relay-boundary, and perf checks: implemented in `tests/e2e/v12/` and `tests/perf/v12/`.

## QoL objective
- Target: >=10% effort reduction on restore journey.
- Verification hook: `tests/perf/v12/recovery_flow_steps_test.go`.

## Promotion recommendation
- Recommendation: promote v1.2 (`F12`) now that mandatory evidence commands are green and all gates are `promoted`.

## Planned vs implemented
- Runtime/test/docs artifacts are implemented.
- Final promotion state is captured in `docs/v1.2/phase5/p5-gate-signoff.md` and `artifacts/v12/gates/*.status.json`.
