# Phase 5 Gate Sign-Off

- G0: Scope lock verified by release engineering.
- G1: Proto compatibility checks passed; see `artifacts/generated/v13-evidence/buf-breaking.txt`.
- G2: `pkg/v13` spaces/join policy logic validated by unit tests.
- G3: Chat/UX flow verified through e2e/perf tests and UI contract.
- G4: Validation matrix executed and closed; see `artifacts/generated/v13-evidence/go-test-e2e-v13.txt`, `artifacts/generated/v13-evidence/go-test-perf-v13.txt`, `artifacts/generated/v13-evidence/make-check-full.txt`.
- G5: Podman manifest recorded in `artifacts/generated/v13-e2e-podman/result-manifest.json`.
- G6: `F14` spec package drafted in phase4 docs.
- G7: Evidence bundle captured in `docs/v1.3/phase5/p5-evidence-bundle.md`.
- G8: Relay regression referenced via `pkg/v11/relaypolicy` tests.
- G9: As-built conformance recorded in `docs/v1.3/phase5/p5-as-built-conformance.md`.
