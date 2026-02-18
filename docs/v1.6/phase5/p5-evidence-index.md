# P5 Evidence Index

| ID | Gate | Artifact | Evidence |
|----|------|----------|----------|
| EV-v16-G0-001 | G0 Scope lock | `docs/v1.6/phase0/p0-scope-lock.md` | `artifacts/generated/v16-evidence/verify-roadmap-docs.txt` |
| EV-v16-G1-001 | G1 Compatibility | Proto compatibility checks | `artifacts/generated/v16-evidence/buf-lint.txt`, `artifacts/generated/v16-evidence/buf-breaking.txt` |
| EV-v16-G2-001 | G2 RBAC runtime | `pkg/v16/rbac/*`, `pkg/v16/acl/*` | `artifacts/generated/v16-evidence/go-test-all.txt` |
| EV-v16-G3-001 | G3 Enforcement | `pkg/v16/enforcement/*`, `tests/e2e/v16/enforcement_test.go` | `artifacts/generated/v16-evidence/go-test-e2e-v16.txt` |
| EV-v16-G4-001 | G4 Validation | `tests/e2e/v16/permission_matrix_test.go`, `tests/perf/v16/permission_steps_test.go` | `artifacts/generated/v16-evidence/go-test-e2e-v16.txt`, `artifacts/generated/v16-evidence/go-test-perf-v16.txt` |
| EV-v16-G5-001 | G5 Podman policy | `scripts/v16-rbac-scenarios.sh`, `containers/v1.6/*` | `artifacts/generated/v16-evidence/v16-rbac-scenarios.txt`, `artifacts/generated/v16-rbac-scenarios/result-manifest.json` |
| EV-v16-G6-001 | G6 v17 spec | `docs/v1.6/phase4/f17-moderation-spec.md`, `docs/v1.6/phase4/f17-proto-delta.md`, `docs/v1.6/phase4/f17-acceptance-matrix.md` | `docs/v1.6/phase4/f17-proto-delta.md`, `docs/v1.6/phase4/f17-acceptance-matrix.md` |
| EV-v16-G7-001 | G7 Docs complete | `docs/v1.6/phase5/p5-evidence-bundle.md`, `docs/v1.6/phase5/p5-gate-signoff.md`, `docs/v1.6/phase5/p5-risk-register.md` | `artifacts/generated/v16-evidence/verify-roadmap-docs.txt`, `artifacts/generated/v16-evidence/make-check-full.txt` |
| EV-v16-G8-001 | G8 Relay regression | `tests/e2e/v16/enforcement_test.go`, `artifacts/generated/v16-rbac-scenarios/result-manifest.json` | `artifacts/generated/v16-evidence/go-test-e2e-v16.txt`, `artifacts/generated/v16-evidence/go-test-relay-regression-v16.txt` |
| EV-v16-G9-001 | G9 As-built conformance | `docs/v1.6/phase5/p5-as-built-conformance.md`, `docs/v1.5/phase4/f16-acceptance-matrix.md` | `artifacts/generated/v16-evidence/go-test-all.txt`, `artifacts/generated/v16-evidence/go-test-e2e-v16.txt`, `artifacts/generated/v16-evidence/buf-breaking.txt` |
