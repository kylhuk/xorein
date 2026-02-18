# Phase 5 - Evidence Index

| EvidenceID | Gate | Command | OutputPath | Result | Owner | Notes |
|---|---|---|---|---|---|---|
| EV-v19-G0-001 | G0 | Scope lock review | `docs/v1.9/phase0/p0-scope-lock.md` | pass | Release Lead | Scope and dependencies frozen. |
| EV-v19-G0-002 | G0 | Traceability matrix review | `docs/v1.9/phase0/p0-traceability-matrix.md` | pass | Release Lead | Requirement coverage complete. |
| EV-v19-G0-003 | G0 | Gate ownership review | `docs/v1.9/phase0/p0-gate-ownership.md` | pass | Release Lead | RACI mapping captured. |
| EV-v19-G1-001 | G1 | `buf lint` | `artifacts/generated/v19-evidence/buf-lint.txt` | pass | Protocol Team | Lint output captured, warning-only deprecation notes. |
| EV-v19-G1-002 | G1 | `buf breaking --against '.git#branch=origin/dev'` | `artifacts/generated/v19-evidence/buf-breaking.txt` | pass | Protocol Team | No breaking proto changes detected. |
| EV-v19-G2-001 | G2 | CO contract tests | `pkg/v19/co/co_test.go` | pass | Networking Lead | Connectivity orchestrator contract tests present and exercised. |
| EV-v19-G3-001 | G3 | No-limbo contract test | `pkg/v19/ui/nolimbo_ui_test.go` | pass | UX Lead | Continuity and state-transition tests present. |
| EV-v19-G3-002 | G3 | QoL contract doc | `docs/v1.9/phase2/p2-nolimbo-ux-contract.md` | pass | UX Lead | Recovery-first state contract published. |
| EV-v19-G4-001 | G4 | `go test ./tests/e2e/v19/...` | `artifacts/generated/v19-evidence/go-test-e2e-v19.txt` | pass | QA Lead | E2E journey matrix coverage captured. |
| EV-v19-G4-002 | G4 | `go test ./tests/perf/v19/...` | `artifacts/generated/v19-evidence/go-test-perf-v19.txt` | pass | QA Lead | Performance/mobility matrix coverage captured. |
| EV-v19-G5-001 | G5 | `scripts/v19-chaos-scenarios.sh` | `artifacts/generated/v19-evidence/v19-chaos-scenarios.txt` | pass | Ops Lead | Podman scenario runner exit 0. |
| EV-v19-G5-002 | G5 | Scenario manifest | `artifacts/generated/v19-chaos-scenarios/result-manifest.json` | pass | Ops Lead | Deterministic manifest recorded. |
| EV-v19-G6-001 | G6 | Release hardening spec | `docs/v1.9/phase4/f20-release-hardening-spec.md` | pass | Plan Lead | v20 hardening package published. |
| EV-v19-G6-002 | G6 | Acceptance matrix | `docs/v1.9/phase4/f20-acceptance-matrix.md` | pass | Plan Lead | v20 acceptance criteria captured. |
| EV-v19-G7-001 | G7 | `make check-full` | `artifacts/generated/v19-evidence/make-check-full.txt` | pass | Evidence Captain | Full validation pipeline captured. |
| EV-v19-G7-002 | G7 | `scripts/verify-roadmap-docs.sh` | `artifacts/generated/v19-evidence/verify-roadmap-docs.txt` | pass | Evidence Captain | Roadmap doc checks captured. |
| EV-v19-G7-003 | G7 | Evidence bundle publish | `docs/v1.9/phase5/p5-evidence-bundle.md` | pass | Evidence Captain | Closure evidence bundle updated and linked. |
| EV-v19-G7-004 | G7 | Evidence index publish | `docs/v1.9/phase5/p5-evidence-index.md` | pass | Evidence Captain | Evidence index updated with G0..G9. |
| EV-v19-G8-001 | G8 | `go test ./tests/e2e/v19/... -run '^TestNATMatrixIncludesRelayBoundary$'` | `artifacts/generated/v19-evidence/go-test-relay-regression-v19.txt` | pass | Security Lead | Relay boundary regression guard exercised. |
| EV-v19-G9-001 | G9 | As-built conformance report | `docs/v1.9/phase5/p5-as-built-conformance.md` | pass | Program Board | Closure report published for implementation against v18 F19 package. |
| EV-v19-G9-002 | G9 | Residual risk register | `docs/v1.9/phase5/p5-risk-register.md` | pass | Program Board | Open risks and mitigations captured. |
