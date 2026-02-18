# Phase 5 Evidence Bundle

- Required v13 evidence artifacts are under `artifacts/generated/v13-evidence/`:
  - `buf-lint.txt` (`buf lint`)
  - `buf-breaking.txt` (`buf breaking --against '.git#branch=origin/dev'`)
  - `go-test-all.txt` (`go test ./...`)
  - `go-test-e2e-v13.txt` (`go test ./tests/e2e/v13/...`)
  - `go-test-perf-v13.txt` (`go test ./tests/perf/v13/...`)
  - `make-check-full.txt` (`make check-full`)
  - `v13-e2e-podman.txt` (`scripts/v13-e2e-podman.sh`)
  - `verify-roadmap-docs.txt` (`scripts/verify-roadmap-docs.sh`)
- Pods manifest from `artifacts/generated/v13-e2e-podman/result-manifest.json` supports deterministic scenario results.
- Include log snippets for relay regression test referencing `pkg/v11/relaypolicy` to prove no-data boundary.
