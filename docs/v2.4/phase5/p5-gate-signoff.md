# Phase 5 gate signoff checklist (P5-T1 ST1–ST4)

| Item | Owner | Gate criteria | Status |
| --- | --- | --- | --- |
| ST1 regression scenarios capturing multi-client, crash recovery, stale socket, and continuity stories. | Ops Lead | `scripts/v24-daemon-scenarios.sh` runs with manifest/log captured, covering G5/G6 commitments. | pass (EV-v24-G8-009) |
| ST2 deterministic E2E suite configured for local API journeys. | QA Lead | `tests/e2e/v24/...` completes without flakes and the deterministic output is archived (`artifacts/generated/v24-evidence/go-test-e2e-v24.txt`). | pass (EV-v24-G8-004) |
| ST3 perf regression and workspace-wide tests protect QoS. | QA Lead | `go test ./tests/perf/v24/...` plus `go test ./...` ensure same day stability; perf output logged. | pass (EV-v24-G8-005, EV-v24-G8-003) |
| ST4 build/lint hygiene command bundle (`buf lint`, `buf breaking`, `go build` for both binaries, `make check-full`). | Plan Lead | All commands succeeded; advisory warnings limited to deprecated `DEFAULT` name in `buf lint` and the Trivy flag in `make check-full`. | pass (EV-v24-G8-001, EV-v24-G8-002, EV-v24-G8-006, EV-v24-G8-007, EV-v24-G8-008) |
| Boundaries enforcement (G9) to keep `cmd/xorein`/`pkg/xorein` Gio-free and align harmolyn to the local API. | Security Lead | `scripts/ci/enforce-boundaries.sh` asserts package import restrictions. | pass (EV-v24-G9-001) |

## Gate readiness and residual risk disposition
- `G8` is ready. The Phase 5 evidence index lists the full command matrix (EV-v24-G8-001…EV-v24-G8-009) and points reviewers to the manifest/log for scenario coverage plus the boundary check script that feeds `G9`.
- `G9` is ready per the boundary enforcement script; residual risk of dependency drift is noted in `p5-risk-register.md` but remains acceptable (no new Gio imports were introduced during this pass).
- Deployers should keep monitoring `scripts/ci/enforce-boundaries.sh` outputs as new packages are added, and re-run the scenario harness when G5/G6 requirements change.
