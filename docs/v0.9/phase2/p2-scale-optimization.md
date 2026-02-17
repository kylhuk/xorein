# Phase 2 · Large-Server Optimization Contracts

## Plan vs implementation
This phase fixes the hierarchical GossipSub fanout, lazy member-loading, scale-driven `SecurityMode` transitions, and rollback/observability contracts. Implementation is contained in `pkg/v09/scale/plan.go` with helpers that the CLI scenario can exercise; intentional planning language in this doc keeps the large-server journey traceable.

## Evidence anchors
| VA ID | Artifact | Files | Description |
|---|---|---|---|
| VA-L1 | Hierarchy topology | `pkg/v09/scale/plan.go` | `HierarchyPlan` partitions member counts into tiered fanout buckets without privileged nodes. |
| VA-L2 | Propagation/backpressure | `pkg/v09/scale/plan.go` | `MemberClass` and `LazyLoadPlan` encode deterministic thresholds for active, nearby, and passive member loading. |
| VA-L4/VA-L7 | State coherence & SecurityMode | `pkg/v09/scale/plan.go` | `NextSecurityMode` with hysteresis plus `ShardingPlan` document when channels shard, why mode epochs rotate, and what rollback posture looks like. |
| VA-L5/VA-L6 | Observability & rollback | `docs/v0.9/phase2/p2-scale-optimization.md` | Observability signals and rollback thresholds are listed as future review artefacts with explicit acceptance links. |

## Planned-vs-implemented language
- Planned: additional telemetry instrumentation can reference the `VA-L5`/`VA-L6` rows when the observability stack is wired; this doc names the signal set and fallback thresholds for gate reviewers.
- Implemented: the Go helpers already return deterministic plans for the CLI scenario so that ledgered values (fanout, class, security mode) stay reproducible for `V9-G2` validation.
