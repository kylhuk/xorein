# Final Evidence Bundle

Mandatory v20 phase-5 command outputs are captured here:

- `buf lint` -> `artifacts/generated/v20-evidence/buf-lint.txt`
- `buf breaking --against '.git#branch=origin/dev'` -> `artifacts/generated/v20-evidence/buf-breaking.txt`
- `go test ./...` -> `artifacts/generated/v20-evidence/go-test-all.txt`
- `go test ./tests/e2e/v20/...` -> `artifacts/generated/v20-evidence/go-test-e2e-v20.txt`
- `go test ./tests/perf/v20/...` -> `artifacts/generated/v20-evidence/go-test-perf-v20.txt`
- `make check-full` -> `artifacts/generated/v20-evidence/make-check-full.txt`
- `scripts/v20-podman-scenarios.sh` -> `artifacts/generated/v20-evidence/v20-podman-scenarios.txt`
- `scripts/verify-roadmap-docs.sh` -> `artifacts/generated/v20-evidence/verify-roadmap-docs.txt`
- `go test ./tests/e2e/v20/... -run '^TestRelayNoDataRegression$'` -> `artifacts/generated/v20-evidence/go-test-relay-regression-v20.txt`
- Podman manifest -> `artifacts/generated/v20-podman-scenarios/result-manifest.json`

All artifact paths above are immutable and referenced by `docs/templates/roadmap-evidence-index.md` via `EV-v20-*` entries.
