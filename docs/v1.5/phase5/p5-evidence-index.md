# Phase 5 Evidence Index

| ID | Gate | Artifact | Evidence |
|----|------|----------|----------|
| EV-v15-G0-001 | G0 Scope lock | `TODO_v15.md` | `artifacts/generated/v15-evidence/verify-roadmap-docs.txt`, `docs/v1.5/phase0/p0-traceability-matrix.md` |
| EV-v15-G1-001 | G1 Compatibility | Proto-wire checks | `artifacts/generated/v15-evidence/buf-lint.txt`, `artifacts/generated/v15-evidence/buf-breaking.txt` |
| EV-v15-G2-001 | G2 Runtime | `pkg/v15/capture`, `pkg/v15/screenshare` | `artifacts/generated/v15-evidence/go-test-all.txt` |
| EV-v15-G3-001 | G3 UX | `pkg/v15/ui/screenshare_ui.go` | `artifacts/generated/v15-evidence/go-test-e2e-v15.txt`, `artifacts/generated/v15-evidence/go-test-perf-v15.txt` |
| EV-v15-G4-001 | G4 Validation | `make check-full` | `artifacts/generated/v15-evidence/make-check-full.txt` |
| EV-v15-G5-001 | G5 Podman scenario | `scripts/v15-screenshare-scenarios.sh` + `artifacts/generated/v15-screenshare-scenarios/result-manifest.json` | `artifacts/generated/v15-evidence/v15-screenshare-scenarios.txt` |
| EV-v15-G6-001 | G6 v16 spec | `docs/v1.5/phase4/f16-acceptance-matrix.md`, `docs/v1.5/phase4/f16-proto-delta.md`, `docs/v1.5/phase4/f16-rbac-acl-spec.md` | `docs/v1.5/phase4/f16-acceptance-matrix.md` |
| EV-v15-G7-001 | G7 Docs | Phase 5 closure docs | `docs/v1.5/phase5/p5-evidence-bundle.md`, `docs/v1.5/phase5/p5-gate-signoff.md`, `docs/v1.5/phase5/p5-risk-register.md` |
| EV-v15-G8-001 | G8 Relay regression | `tests/e2e/v15/adaptation_recovery_test.go` | `artifacts/generated/v15-evidence/go-test-e2e-v15.txt` |
| EV-v15-G9-001 | G9 As-built | `docs/v1.5/phase5/p5-as-built-conformance.md` | `artifacts/generated/v15-evidence/go-test-all.txt`, `artifacts/generated/v15-evidence/go-test-e2e-v15.txt` |
