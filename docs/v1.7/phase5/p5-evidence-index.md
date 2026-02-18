# Phase 5 — Evidence Index

| ID | Gate | Evidence Command | Artifact |
|----|------|-----------------|----------|
| EV-v17-G1-001 | G1 Proto compatibility | `buf lint` | `artifacts/generated/v17-evidence/buf-lint.txt` |
| EV-v17-G1-002 | G1 Proto compatibility | `buf breaking --against '.git#branch=origin/dev'` | `artifacts/generated/v17-evidence/buf-breaking.txt` |
| EV-v17-G2-001 | G2 Moderation runtime | `go test ./...` | `artifacts/generated/v17-evidence/go-test-all.txt` |
| EV-v17-G4-001 | G4 Validation matrix | `go test ./tests/e2e/v17/...` | `artifacts/generated/v17-evidence/go-test-e2e-v17.txt` |
| EV-v17-G4-002 | G4 Validation matrix | `go test ./tests/perf/v17/...` | `artifacts/generated/v17-evidence/go-test-perf-v17.txt` |
| EV-v17-G5-001 | G5 Podman scenarios | `scripts/v17-moderation-scenarios.sh` | `artifacts/generated/v17-evidence/v17-moderation-scenarios.txt`, `artifacts/generated/v17-moderation-scenarios/result-manifest.json` |
| EV-v17-G7-001 | G7 Docs & evidence | `make check-full` | `artifacts/generated/v17-evidence/make-check-full.txt` |
| EV-v17-G7-002 | G7 Docs & evidence | `scripts/verify-roadmap-docs.sh` | `artifacts/generated/v17-evidence/verify-roadmap-docs.txt` |
| EV-v17-G8-001 | G8 Relay regression | `go test ./tests/e2e/v17/... -run '^TestAdversarialEvents$'` | `artifacts/generated/v17-evidence/go-test-relay-regression-v17.txt` |
