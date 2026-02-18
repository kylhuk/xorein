# Phase 5 Evidence Bundle

Evidence for v1.5 includes:
- Command captures:
  - `buf lint` -> `artifacts/generated/v15-evidence/buf-lint.txt`
  - `buf breaking --against '.git#branch=origin/dev'` -> `artifacts/generated/v15-evidence/buf-breaking.txt`
  - `go test ./...` -> `artifacts/generated/v15-evidence/go-test-all.txt`
  - `go test ./tests/e2e/v15/...` -> `artifacts/generated/v15-evidence/go-test-e2e-v15.txt`
  - `go test ./tests/perf/v15/...` -> `artifacts/generated/v15-evidence/go-test-perf-v15.txt`
  - `make check-full` -> `artifacts/generated/v15-evidence/make-check-full.txt`
  - `scripts/v15-screenshare-scenarios.sh` -> `artifacts/generated/v15-evidence/v15-screenshare-scenarios.txt`
  - `scripts/verify-roadmap-docs.sh` -> `artifacts/generated/v15-evidence/verify-roadmap-docs.txt`
- Scenario manifest and per-probe logs:
  - `artifacts/generated/v15-screenshare-scenarios/result-manifest.json`
