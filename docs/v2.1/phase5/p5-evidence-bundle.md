# Evidence Bundle

> As-built summary of the mandatory command outputs that now anchor the v21 promotion checklist.

## Mandatory command outcomes
| Command | Gate | EV ID(s) | Artifact | Notes |
|---|---|---|---|---|
| `buf lint` | G1 | EV-v21-G1-001 | `artifacts/generated/v21-evidence/buf-lint-full.txt` | Additive schema guardrail pass; Buf warns only about the deprecated `DEFAULT` category that will be renamed later. |
| `buf breaking --against 'origin/main'` | G1 | EV-v21-G1-002 | `artifacts/generated/v21-evidence/buf-breaking-full.txt` | Breaking check succeeded with no output, demonstrating compatibility. |
| `go test ./pkg/v21/store/...` | G2 | EV-v21-G2-001 | `artifacts/generated/v21-evidence/go-test-pkg-v21-store.txt` | Store regression suite for the encrypted local timeline completed without failures. |
| `go test ./...` | G4 | EV-v21-G4-001 | `artifacts/generated/v21-evidence/go-test-all.txt` | Full unit/integration matrix including v21 packages ran cleanly. |
| `go test ./tests/e2e/v21/...` | G4 | EV-v21-G4-002 | `artifacts/generated/v21-evidence/go-test-e2e-v21.txt` | Scenario coverage for persistence/search flows passed. |
| `go test ./tests/perf/v21/...` | G4 | EV-v21-G4-003 | `artifacts/generated/v21-evidence/go-test-perf-v21.txt` | Perf regression suite ran to completion with no regressions. |
| `scripts/v21-persistence-scenarios.sh` | G5/G8 | EV-v21-G5-001, EV-v21-G8-001 | `artifacts/generated/v21-evidence/v21-persistence-scenarios.txt` + manifest | Podman multi-peer persistence, local-clear, and relay boundary probes all passed; scenario log plus `artifacts/generated/v21-persistence-scenarios/manifest.txt` capture the outcome. |
| `go build ./cmd/xorein` | G7 | EV-v21-G7-001 | `artifacts/generated/v21-evidence/go-build-xorein.txt` | Backend runtime builds cleanly without UI dependencies. |
| `go build ./cmd/harmolyn` | G7 | EV-v21-G7-002 | `artifacts/generated/v21-evidence/go-build-harmolyn.txt` | UI binary builds successfully and consumes the runtime library as expected. |
| `make check-full` | G7 | EV-v21-G7-003 | `artifacts/generated/v21-evidence/make-check-full.txt` | Meta-check passed; gosec reported 0 issues and trivy warned that `--security-checks` is deprecated (no failing findings). |

## Podman persistence manifest
- `artifacts/generated/v21-persistence-scenarios/manifest.txt` records probe results (relay boundary pass, multi-peer restart, local clear). It is already referenced by EV-v21-G5-002 and EV-v21-G8-001 and stored alongside the script log for post-mortem review.

## Evidence linkage
- Every EV ID listed above is documented in `docs/v2.1/phase5/p5-evidence-index.md`, which now records timestamps, checksums, and the exact artifacts for promotion day.
- See `docs/v2.1/phase5/p5-as-built-conformance.md` for how these outputs satisfy the v20 `F21` seeds.
