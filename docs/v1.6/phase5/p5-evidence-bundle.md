# P5 Evidence Bundle

Evidence for v1.6 includes:

- Command captures:
  - `buf lint` -> `artifacts/generated/v16-evidence/buf-lint.txt`
  - `buf breaking --against '.git#branch=origin/dev'` -> `artifacts/generated/v16-evidence/buf-breaking.txt`
  - `go test ./...` -> `artifacts/generated/v16-evidence/go-test-all.txt`
  - `go test ./tests/e2e/v16/...` -> `artifacts/generated/v16-evidence/go-test-e2e-v16.txt`
  - `go test ./tests/perf/v16/...` -> `artifacts/generated/v16-evidence/go-test-perf-v16.txt`
  - `make check-full` -> `artifacts/generated/v16-evidence/make-check-full.txt`
  - `scripts/v16-rbac-scenarios.sh` -> `artifacts/generated/v16-evidence/v16-rbac-scenarios.txt`
  - `scripts/verify-roadmap-docs.sh` -> `artifacts/generated/v16-evidence/verify-roadmap-docs.txt`
  - `go test ./tests/e2e/v16/... -run '^TestEnforcementFlow$'` -> `artifacts/generated/v16-evidence/go-test-relay-regression-v16.txt`

- Podman scenario evidence and relay boundary proofs:
  - `artifacts/generated/v16-rbac-scenarios/result-manifest.json`
  - `artifacts/generated/v16-evidence/go-test-relay-regression-v16.txt`
