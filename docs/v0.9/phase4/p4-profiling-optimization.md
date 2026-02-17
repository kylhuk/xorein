# Phase 4 · Cross-Platform Profiling & Optimization Contracts

## Plan vs implementation
Profiling contracts remain at the planning/reporting level while the deterministic helpers live in docs and the CLI scenario only peripherally (via runbook thresholds tied to `pkg/v09/scenario`). This doc records the metric taxonomy, baseline capture protocol, bottleneck classification, and optimization acceptance rules for `V9-G4` reviewers.

## Evidence anchors
| VA ID | Artifact | Files | Description |
|---|---|---|---|
| VA-P1 | Metric taxonomy | `docs/v0.9/phase4/p4-profiling-optimization.md` | Defines CPU, memory, network, latency, jitter, startup, and background metrics with normalization rules. |
| VA-P2 | Baseline capture | `docs/v0.9/phase4/p4-profiling-optimization.md` | Baseline run protocol (environment, sample size, artifact naming). |
| VA-P3/VA-P4 | Optimization decision policy | `pkg/v09/scenario/scenario.go`, `docs/v0.9/phase4/p4-profiling-optimization.md` | `RunForgeScenario` codifies classification/confidence logic derived from this doc. |
| VA-P5/VA-P6 | Reporting & rollback contracts | `docs/v0.9/phase4/p4-profiling-optimization.md` | Report schema and rollback/exemption process with planned-vs-implemented language. |

## Planned-vs-implemented language
- Planned: real profiling data and regression exceptions will fold into the schema referenced above when `tests/perf/v09/` exists in follow-up tasks.
- Implemented: this doc plus `RunForgeScenario` already encode the taxonomy, classification, and threshold policy that future profiling automation must meet, so reviewers have deterministic text to audit against.
