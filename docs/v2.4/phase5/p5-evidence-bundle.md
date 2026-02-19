# Phase 5 evidence bundle (P5-T1 ST1–ST4)

## Summary
- Evidence bundle centralizes the Phase 5 closure artifacts required by Gates `G8` and `G9` plus the traceability from G5/G6 scenario outputs.
- All command outputs listed below are copied to `artifacts/generated/v24-evidence/` (and the daemon scenario manifest/log) so reviewers can replay the regression surface or rerun the feel-good checks.
- Warnings observed during the runs (`buf lint` deprecated `DEFAULT` name and `make check-full`’s advisory Trivy flag) remain non-blocking; they are noted below and do not affect gate readiness.

## ST1–ST4 mapping
| ST | Focus | Representative evidence | Status |
| --- | --- | --- | --- |
| ST1 | Daemon scenario coverage (G5/G6) | `scripts/v24-daemon-scenarios.sh` plus the manifest/log under `artifacts/generated/v24-daemon-scenarios/` (EV-v24-G8-009). | PASS |
| ST2 | Deterministic E2E regression (`tests/e2e/v24/...`) | `artifacts/generated/v24-evidence/go-test-e2e-v24.txt` (EV-v24-G8-004). | PASS |
| ST3 | Performance guardrails (`tests/perf/v24/...`) | `artifacts/generated/v24-evidence/go-test-perf-v24.txt` (EV-v24-G8-005). | PASS |
| ST4 | Build/lint hygiene and CI rules | `buf lint`, `buf breaking`, `go build ./cmd/xorein`, `go build ./cmd/harmolyn`, `make check-full`, and `scripts/ci/enforce-boundaries.sh` outputs (EV-v24-G8-001…EV-v24-G8-008 plus EV-v24-G9-001). | PASS (warnings advisory) |

## Gate mapping
- **G8 (Docs + release evidence)** is satisfied by the command matrix recorded here (EV-v24-G8-001 through EV-v24-G8-009) plus the gate signoff (`p5-gate-signoff.md`) and as-built narrative (`p5-as-built-conformance.md`).
- **G9 (Backend dependency invariant)** is satisfied by the boundary enforcement script that confirms no Gio or protocol runtime imports exist on `cmd/xorein`/`pkg/xorein` and by calling out the same command in the gate signoff (EV-v24-G9-001). Residual risk for dependency drift is tracked in `p5-risk-register.md`.
- **G5/G6 (Multi-client + scenario harness)** references the daemon scenario manifest/log created in `artifacts/generated/v24-daemon-scenarios/`; EV-v24-G8-009 documents the multi-client and Podman scenario story and also serves as trace evidence for these prior gates that feed into this closure.

## Next actions
1. Reviewers should follow each `EV-v24-G8-###` link to replay the command outputs, paying attention to the advisory `DEFAULT` warning from `buf lint` and the Trivy flag warning from `make check-full`.
2. Cross-check the manifest/log in `artifacts/generated/v24-daemon-scenarios/` with the G5/G6 scenario catalogue to ensure the scenario IDs listed in `docs/v2.4/phase3/p3-daemon-scenarios.md` were covered.
3. Confirm the gate signoff, conformance narrative, and index all reference these same artifacts before the `P5-T1` closure decision is recorded in the go/no-go log.
