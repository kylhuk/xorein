# v1.2 Phase 0 - Traceability Matrix

## Requirement to artifact mapping
| Requirement | Code artifacts | Test artifacts | Docs artifacts |
|---|---|---|---|
| Immutable identity lifecycle | `pkg/v12/identity/identity.go` | `pkg/v12/identity/identity_test.go` | `docs/v1.2/phase0/p0-scope-lock.md` |
| Local backup envelope (`Argon2id + AEAD`) | `pkg/v12/backup/backup.go` | `pkg/v12/backup/backup_test.go` | `docs/v1.2/phase1/p1-backup-format.md` |
| Additive wire metadata | `proto/aether.proto`, `gen/go/proto/*` | `go test ./...` | `docs/v1.2/phase2/p2-proto-changelog.md` |
| Onboarding + restore UX contract | `pkg/v12/ui/flow.go`, `pkg/ui/shell.go` | `pkg/v12/ui/flow_test.go`, `pkg/ui/v12_identity_restore_test.go` | `docs/v1.2/phase2/p2-ux-contract.md` |
| Recovery + relay boundary scenarios | `scripts/v12-recovery-scenarios.sh` | `tests/e2e/v12/recovery_flow_test.go` | `docs/v1.2/phase3/p3-podman-scenarios.md` |
| QoL effort reduction objective | `tests/perf/v12/recovery_flow_steps_test.go` | `go test ./tests/perf/v12/...` | `docs/v1.2/phase5/p5-as-built-conformance.md` |

## Gate mapping
| Gate | Primary owner artifact |
|---|---|
| G0 | `docs/v1.2/phase0/p0-scope-lock.md` |
| G1 | `docs/v1.2/phase2/p2-proto-changelog.md` |
| G2 | `pkg/v12/identity/*`, `pkg/v12/backup/*` |
| G3 | `pkg/v12/ui/*`, `pkg/ui/shell.go`, `pkg/ui/v12_identity_restore_test.go` |
| G4 | `pkg/v12/**/*_test.go`, `tests/e2e/v12/*`, `tests/perf/v12/*` |
| G5 | `scripts/v12-recovery-scenarios.sh`, `containers/v1.2/*` |
| G6 | `docs/v1.2/phase4/*` |
| G7 | `docs/v1.2/phase5/*` |
| G8 | `tests/e2e/v12/recovery_flow_test.go` |
| G9 | `docs/v1.2/phase5/p5-as-built-conformance.md` |

## Planned vs implemented
- Traceability is implementation-backed for v1.2 code paths.
- v1.3 artifacts remain specification-only.
