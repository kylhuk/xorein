# Phase 9 · Release Conformance & Handoff

## Release conformance / handoff narrative
Release readiness rests on traceable evidence tables, risk logs, deterministic CLI scenarios, and deterministic contract-level QA/perf suites. Implementation includes this document, `pkg/v09/scenario/scenario.go`, `cmd/aether/main.go`, `tests/e2e/v09/scenario_suite_test.go`, and `tests/perf/v09/deterministic_suite_test.go` so reviewers rerunning the same checklist can reproduce outputs for contract-level behavior. Nothing outside this folder is claimed to be production-ready; known limitations are listed below.

## Evidence anchors
| VA ID | Artifact | Files | Description |
|---|---|---|---|
| VA-H1 | Release checklist | `docs/v0.9/phase9/p9-release-conformance.md` | Documents release handoff steps, risk log references, and QA evidence requirements. |
| VA-H2 | Scenario witness | `pkg/v09/scenario/scenario.go`, `cmd/aether/main.go`, `tests/e2e/v09/scenario_suite_test.go` | `RunForgeScenario` plus `v09-forge` CLI entry and the deterministic scenario suite provide deterministic evidence anchors for gates and QA. |
| VA-E7 | Perf runbook | `docs/v0.9/phase9/perf-runbook.md`, `docs/v0.9/phase9/VA-E7-thresholds.md`, `tests/perf/v09/deterministic_suite_test.go` | Contains reproducibility steps, threshold registry, and deterministic perf suite that exercises the stress contract at policy/helper level. |
| VA-C1 | Evidence bundle | `docs/v0.9/README.md` | Maps the entire document set to required evidence anchors so reviewers can validate planned vs implemented text. |

## Planned-vs-implemented language
- Implemented: QA/test automation now consists of `tests/e2e/v09/scenario_suite_test.go` and `tests/perf/v09/deterministic_suite_test.go`, which exercise deterministic contracts and threshold checks referenced here and produce deterministic outputs for gate reviewers.
- Implemented: this release document and the CLI witness document the evidence bundle for `RunForgeScenario`, so reviewers can declare contract-level gate completion by rerunning the CLI and deterministic suites while verifying the recorded outputs.

## Handoff checklist (evidence)
| Item | Status | Evidence path |
|---|---|---|
| Scope lock for v0.9 gates | implemented | `docs/v0.9/phase0/p0-scope-governance.md` |
| IPFS/persistence contract | implemented | `pkg/v09/ipfs/contracts.go` |
| Stress/perf reproducibility | implemented | `docs/v0.9/phase5/p5-stress-contract.md`, `docs/v0.9/phase9/perf-runbook.md`, `tests/perf/v09/deterministic_suite_test.go` |
| CLI witness scenario | implemented | `cmd/aether/main.go`, `pkg/v09/scenario/scenario.go` |
| QA/test automation | implemented | `tests/perf/v09/deterministic_suite_test.go`, `tests/e2e/v09/scenario_suite_test.go` |

## Known limitations
- Deterministic suites validate contract and threshold logic; they do not emulate live network traffic or production infrastructure.
- Claims in this slice are limited to reproducible in-repo deterministic evidence.
