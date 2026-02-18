# Phase 5 Evidence Bundle

Evidence for v1.4 includes:
- Command captures:
  - `buf lint` -> `artifacts/generated/v14-evidence/buf-lint.txt`
  - `buf breaking --against '.git#branch=origin/dev'` -> `artifacts/generated/v14-evidence/buf-breaking.txt`
  - `go test ./...` -> `artifacts/generated/v14-evidence/go-test-all.txt`
  - `go test ./tests/e2e/v14/...` -> `artifacts/generated/v14-evidence/go-test-e2e-v14.txt`
  - `go test ./tests/perf/v14/...` -> `artifacts/generated/v14-evidence/go-test-perf-v14.txt`
  - `make check-full` -> `artifacts/generated/v14-evidence/make-check-full.txt`
  - `scripts/v14-voice-scenarios.sh` -> `artifacts/generated/v14-evidence/v14-voice-scenarios.txt`
  - `scripts/verify-roadmap-docs.sh` -> `artifacts/generated/v14-evidence/verify-roadmap-docs.txt`
- Scenario manifest and per-probe logs:
  - `artifacts/generated/v14-voice-scenarios/result-manifest.json`
  - `artifacts/generated/v14-voice-scenarios/voice-call-setup.log`
  - `artifacts/generated/v14-voice-scenarios/voice-reconnect.log`
  - `artifacts/generated/v14-voice-scenarios/relay-boundary.log`
- Relay regression evidence:
  - `go test ./tests/e2e/v14/... -run '^TestVoiceFlowSequence$'` -> `artifacts/generated/v14-evidence/go-test-relay-regression-v14.txt`
  - Regression assertion uses `pkg/v11/relaypolicy` durable-storage rejection path.
- G6 v15 spec publication evidence:
  - `docs/v1.4/phase4/f15-screenshare-spec.md`
  - `docs/v1.4/phase4/f15-proto-delta.md`
  - `docs/v1.4/phase4/f15-acceptance-matrix.md`
- G9 as-built conformance evidence:
  - `docs/v1.4/phase5/p5-as-built-conformance.md`
  - `docs/v1.4/phase5/p5-risk-register.md`
