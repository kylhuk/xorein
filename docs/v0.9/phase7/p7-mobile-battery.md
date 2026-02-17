# Phase 7 · Mobile Battery Optimization Contracts

## Plan vs implementation
Mobile battery optimization (background budgets, wake/suppression rules, and multi-provider failover) is implemented via `pkg/v09/mobile/budget.go`. This doc records the budgets, decision boundaries, and governance reminders that gate reviewers should reference for `V9-G7`.

## Evidence anchors
| VA ID | Artifact | Files | Description |
|---|---|---|---|
| VA-B1 | Background budgets | `pkg/v09/mobile/budget.go` | `BackgroundBudget` calculates deterministic budget classes tied to CPU/battery profiles. |
| VA-B2 | Wake/suppression policy | `pkg/v09/mobile/budget.go` | `EvaluateWakePolicy` captures when to allow wake events vs suppressing them based on battery + priority. |
| VA-B7 | Multi-provider wake failover | `docs/v0.9/phase7/p7-mobile-battery.md` | Notes governance language that still allows volunteers across providers. |
| VA-X7 | Governance alignment | `pkg/v09/governance/checklist.go` | Licensing/audit helpers referenced here ensure the mobile battery contracts remain additive. |

## Planned-vs-implemented language
- Planned: tuned budgets for specific handset classes will be added when platform data is available; this doc names the target entry points so future engineering has a clear reference.
- Implemented: `RunForgeScenario` consumes `BackgroundBudget` and `EvaluateWakePolicy`, making the described battery decisions re-playable in the CLI witness.
