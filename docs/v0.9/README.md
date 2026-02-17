# v0.9 Forge Execution Slice

## Intentional plan vs implementation
Each artifact beneath `docs/v0.9/` stays explicit about what is implemented (deterministic Go helpers, CLI witness, in-repo deterministic test suites, evidence tables) and what remains a planned future validation. Statements like "implemented" refer to the contracts, CLI wiring, and deterministic suites inside `pkg/v09/`, `cmd/aether --scenario v09-forge`, `tests/e2e/v09/scenario_suite_test.go`, and `tests/perf/v09/deterministic_suite_test.go`; anything outside that scope (network deployment, stress runs on production hardware, UI polish) is marked as deferred.

## Document map
| Area | Documents |
|---|---|
| Phase 0 · Scope & governance lock | `phase0/p0-scope-governance.md` |
| Phase 1 · IPFS persistent hosting | `phase1/p1-ipfs-contracts.md` |
| Phase 2 · Large-server optimization | `phase2/p2-scale-optimization.md` |
| Phase 3 · Cascading SFU mesh | `phase3/p3-sfu-mesh.md` |
| Phase 4 · Performance profiling & optimization | `phase4/p4-profiling-optimization.md` |
| Phase 5 · Stress/test planning | `phase5/p5-stress-contract.md`, `phase9/perf-runbook.md`, `phase9/VA-E7-thresholds.md` |
| Phase 6 · Relay performance | `phase6/p6-relay-performance.md` |
| Phase 7 · Mobile battery optimization | `phase7/p7-mobile-battery.md` |
| Phase 8 · Governance readiness | `phase8/p8-governance-audit.md` |
| Phase 9 · Release conformance & handoff | `phase9/p9-release-conformance.md` |

## Evidence anchors (planned vs implemented linkages)
| VA ID | Artifact | Description |
|---|---|---|
| VA-G0x | `pkg/v09/conformance/gates.go` | Gate catalog, checklist helpers, and summary utilities that touch V9-G0 through V9-G9 criteria without touching runtime binaries. |
| VA-I1 | `pkg/v09/ipfs/contracts.go` | Deterministic content addressing, pin lifecycle, and retention classifier code. |
| VA-L1 | `pkg/v09/scale/plan.go` | Hierarchy + lazy member helpers, security-mode transition guards, and sharding guidance. |
| VA-S1 | `pkg/v09/sfu/topology.go` | Cascading tier descriptions, forwarding path selection, and failover decisions needed for 200+ participant planning. |
| VA-R1 | `pkg/v09/relay/policy.go` | Capacity envelope, overload priority, and recovery helpers for relay nodes. |
| VA-B1 | `pkg/v09/mobile/budget.go` | Background budget classes and wake/suppression logic required for mobile battery planning. |
| VA-X1 | `pkg/v09/governance/checklist.go` | Additive/major-path classifiers plus licensing audit helpers. |
| VA-C1 | `pkg/v09/scenario/scenario.go` & `cmd/aether/main.go` | `RunForgeScenario` exercises the helper stack and `cmd/aether --scenario v09-forge` reports pass/fail so reviewers can rerun the contract set deterministically. |

## Open reminders
- Deterministic stress/performance executions now exist in `tests/perf/v09/deterministic_suite_test.go`, are guided by `docs/v0.9/phase9/perf-runbook.md`, and replay the thresholds cataloged in `docs/v0.9/phase9/VA-E7-thresholds.md`; CLI witness coverage remains available via `tests/e2e/v09/scenario_suite_test.go`. That infrastructure satisfies the existing runbook requirements without relying on external harnesses.
- Compatibility/wire evolution remains additive-only; any future schema touches must reference the `VA-G*` checklist before reaching a gate.

## CLI witness reference
Run `cmd/aether --scenario v09-forge` to trigger the deterministic cross-package contract evaluation. The scenario prints `v0.9 forge: PASS` or `v0.9 forge: FAIL: <reason>` so reviewers can validate the implemented helpers without external traffic.
