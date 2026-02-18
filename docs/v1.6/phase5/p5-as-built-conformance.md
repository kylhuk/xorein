# P5 As-Built Conformance

v16 implementation is aligned to `docs/v1.5/phase4/f16-rbac-acl-spec.md` and `docs/v1.5/phase4/f16-acceptance-matrix.md`.

| Acceptance criterion (`docs/v1.5/phase4/f16-acceptance-matrix.md`) | v16 evidence | Result |
|---|---|---|
| Role model and permissions defined | `pkg/v16/rbac/rbac.go`, `pkg/v16/rbac/rbac_test.go` | pass |
| Merge precedence deterministic (`deny` over `allow`) | `pkg/v16/acl/acl.go`, `pkg/v16/acl/acl_test.go` | pass |
| Proto delta stays additive | `docs/v1.5/phase4/f16-proto-delta.md`, `artifacts/generated/v16-evidence/buf-breaking.txt` | pass |
| Enforcement bound to role + channel policy | `pkg/v16/enforcement/enforcement.go`, `tests/e2e/v16/enforcement_test.go`, `tests/e2e/v16/permission_matrix_test.go` | pass |
| Relay boundary preserved | `tests/e2e/v16/enforcement_test.go`, `artifacts/generated/v16-evidence/go-test-e2e-v16.txt` | pass |
| UI management explainability present | `pkg/v16/ui/admin_ui.go`, `pkg/v16/ui/admin_ui_test.go`, `tests/perf/v16/permission_steps_test.go` | pass |
