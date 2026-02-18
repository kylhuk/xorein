# v1.2 Phase 5 - Evidence Index

## Evidence ID format
- `EV-v12-GX-###`

## Index
| Evidence ID | Gate | Command / Artifact | Output path | Owner | Result |
|---|---|---|---|---|---|
| EV-v12-G0-001 | G0 | Scope lock | docs/v1.2/phase0/p0-scope-lock.md | planning | pass |
| EV-v12-G0-002 | G0 | Traceability matrix | docs/v1.2/phase0/p0-traceability-matrix.md | planning | pass |
| EV-v12-G0-003 | G0 | Gate ownership | docs/v1.2/phase0/p0-gate-ownership.md | planning | pass |
| EV-v12-G1-001 | G1 | `buf lint` | artifacts/generated/v12-evidence/buf-lint.txt | protocol | pass |
| EV-v12-G1-002 | G1 | `buf breaking --against '.git#branch=origin/dev'` | artifacts/generated/v12-evidence/buf-breaking.txt | protocol | pass |
| EV-v12-G2-001 | G2 | Identity and backup runtime implementation | pkg/v12/identity/*, pkg/v12/backup/* | runtime | pass |
| EV-v12-G3-001 | G3 | Onboarding and restore UX contract implementation | pkg/v12/ui/*, pkg/ui/*, docs/v1.2/phase2/p2-ux-contract.md | client | pass |
| EV-v12-G4-001 | G4 | `go test ./...` | artifacts/generated/v12-evidence/go-test-all.txt | qa | pass |
| EV-v12-G4-002 | G4 | `go test ./tests/e2e/v12/...` | artifacts/generated/v12-evidence/go-test-e2e-v12.txt | qa | pass |
| EV-v12-G4-003 | G4 | `go test ./tests/perf/v12/...` | artifacts/generated/v12-evidence/go-test-perf-v12.txt | qa | pass |
| EV-v12-G5-001 | G5 | `scripts/v12-recovery-scenarios.sh` | artifacts/generated/v12-recovery-scenarios/result-manifest.json | ops | pass |
| EV-v12-G6-001 | G6 | F13 spaces lifecycle spec | docs/v1.2/phase4/f13-spaces-chat-spec.md | spec | pass |
| EV-v12-G6-002 | G6 | F13 chat flow spec | docs/v1.2/phase4/f13-chat-flows.md | spec | pass |
| EV-v12-G6-003 | G6 | F13 proto delta and acceptance matrix | docs/v1.2/phase4/f13-proto-delta.md, docs/v1.2/phase4/f13-acceptance-matrix.md | spec | pass |
| EV-v12-G7-001 | G7 | `make check-full` | artifacts/generated/v12-evidence/make-check-full.txt | release | pass |
| EV-v12-G8-001 | G8 | Relay no-data-hosting regression scenario | tests/e2e/v12/recovery_flow_test.go, artifacts/generated/v12-evidence/go-test-e2e-v12.txt | qa | pass |
| EV-v12-G9-001 | G9 | As-built conformance report | docs/v1.2/phase5/p5-as-built-conformance.md | release | pass |

## Planned vs implemented
- Index entries exist for all required command evidence.
- Result fields are updated after commands are executed.
