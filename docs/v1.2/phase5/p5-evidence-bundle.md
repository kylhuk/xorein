# v1.2 Phase 5 - Evidence Bundle

## Purpose
This artifact tracks mandatory command evidence and gate closure proof for v1.2 promotion.

## Command evidence tracker
| Command | Evidence ID | Output path | Status |
|---|---|---|---|
| `buf lint` | EV-v12-G1-001 | artifacts/generated/v12-evidence/buf-lint.txt | pass |
| `buf breaking --against '.git#branch=origin/dev'` | EV-v12-G1-002 | artifacts/generated/v12-evidence/buf-breaking.txt | pass |
| `go test ./...` | EV-v12-G4-001 | artifacts/generated/v12-evidence/go-test-all.txt | pass |
| `go test ./tests/e2e/v12/...` | EV-v12-G4-002 | artifacts/generated/v12-evidence/go-test-e2e-v12.txt | pass |
| `go test ./tests/perf/v12/...` | EV-v12-G4-003 | artifacts/generated/v12-evidence/go-test-perf-v12.txt | pass |
| `make check-full` | EV-v12-G7-001 | artifacts/generated/v12-evidence/make-check-full.txt | pass |
| `scripts/v12-recovery-scenarios.sh` | EV-v12-G5-001 | artifacts/generated/v12-recovery-scenarios/result-manifest.json | pass |

## Notes
- `buf lint` completed with a non-blocking deprecation warning for lint category `DEFAULT` in `buf.yaml`.

## Supporting artifacts
- Risk closure: `docs/v1.2/phase5/p5-risk-register.md`
- As-built conformance: `docs/v1.2/phase5/p5-as-built-conformance.md`
- Sign-off: `docs/v1.2/phase5/p5-gate-signoff.md`
- Evidence index: `docs/v1.2/phase5/p5-evidence-index.md`

## Planned vs implemented
- Mandatory evidence commands are captured and linked above.
