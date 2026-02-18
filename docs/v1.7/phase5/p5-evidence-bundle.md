# Phase 5 — Evidence Bundle

Collected command evidence for v1.7 F17 delivery:

- `buf lint` -> `artifacts/generated/v17-evidence/buf-lint.txt`
- `buf breaking --against '.git#branch=origin/dev'` -> `artifacts/generated/v17-evidence/buf-breaking.txt`
- `go test ./...` -> `artifacts/generated/v17-evidence/go-test-all.txt`
- `go test ./tests/e2e/v17/...` -> `artifacts/generated/v17-evidence/go-test-e2e-v17.txt`
- `go test ./tests/perf/v17/...` -> `artifacts/generated/v17-evidence/go-test-perf-v17.txt`
- `make check-full` -> `artifacts/generated/v17-evidence/make-check-full.txt`
- `scripts/v17-moderation-scenarios.sh` -> `artifacts/generated/v17-evidence/v17-moderation-scenarios.txt`
- `scripts/verify-roadmap-docs.sh` -> `artifacts/generated/v17-evidence/verify-roadmap-docs.txt`
- `go test ./tests/e2e/v17/... -run '^TestAdversarialEvents$'` -> `artifacts/generated/v17-evidence/go-test-relay-regression-v17.txt`
- Scenario manifest: `artifacts/generated/v17-moderation-scenarios/result-manifest.json`
