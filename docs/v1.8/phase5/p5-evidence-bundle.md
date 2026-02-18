# Phase 5 - Evidence Bundle

Collected command evidence for v1.8:

- `buf lint` -> `artifacts/generated/v18-evidence/buf-lint.txt`
- `buf breaking --against '.git#branch=origin/dev'` -> `artifacts/generated/v18-evidence/buf-breaking.txt`
- `go test ./...` -> `artifacts/generated/v18-evidence/go-test-all.txt`
- `go test ./pkg/v18/...` -> `artifacts/generated/v18-evidence/go-test-pkg-v18.txt`
- `go test ./tests/e2e/v18/...` -> `artifacts/generated/v18-evidence/go-test-e2e-v18.txt`
- `go test ./tests/perf/v18/...` -> `artifacts/generated/v18-evidence/go-test-perf-v18.txt`
- `make check-full` -> `artifacts/generated/v18-evidence/make-check-full.txt`
- `scripts/v18-discovery-scenarios.sh` -> `artifacts/generated/v18-evidence/v18-discovery-scenarios.txt`
- `scripts/verify-roadmap-docs.sh` -> `artifacts/generated/v18-evidence/verify-roadmap-docs.txt`
- `go test ./tests/e2e/v18/... -run '^TestDiscoveryIntegrityEnforcesRelayBoundary$'` -> `artifacts/generated/v18-evidence/go-test-relay-regression-v18.txt`

Scenario artifacts:
- `artifacts/generated/v18-discovery-scenarios/result-manifest.json`
