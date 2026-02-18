# Phase 5 - Evidence Bundle

Collected command evidence for v1.9:

- `buf lint` -> `artifacts/generated/v19-evidence/buf-lint.txt`
- `buf breaking --against '.git#branch=origin/dev'` -> `artifacts/generated/v19-evidence/buf-breaking.txt`
- `go test ./...` -> `artifacts/generated/v19-evidence/go-test-all.txt`
- `go test ./tests/e2e/v19/...` -> `artifacts/generated/v19-evidence/go-test-e2e-v19.txt`
- `go test ./tests/perf/v19/...` -> `artifacts/generated/v19-evidence/go-test-perf-v19.txt`
- `make check-full` -> `artifacts/generated/v19-evidence/make-check-full.txt`
- `scripts/v19-chaos-scenarios.sh` -> `artifacts/generated/v19-evidence/v19-chaos-scenarios.txt`
- `scripts/verify-roadmap-docs.sh` -> `artifacts/generated/v19-evidence/verify-roadmap-docs.txt`
- `go test ./tests/e2e/v19/... -run '^TestNATMatrixIncludesRelayBoundary$'` -> `artifacts/generated/v19-evidence/go-test-relay-regression-v19.txt`

Scenario manifest artifacts:
- `artifacts/generated/v19-chaos-scenarios/result-manifest.json`
