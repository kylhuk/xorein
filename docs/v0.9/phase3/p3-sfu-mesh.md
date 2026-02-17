# Phase 3 · Cascading SFU Mesh Contracts

## Plan vs implementation
Large-voice cascades remain scoped to deterministic topology, forwarding, failover, and governance boundaries. Implementation lives in `pkg/v09/sfu/topology.go`; this doc keeps the control-channel rules and escalation examples explicit for `VA-S1`..`VA-S6` gate reviewers.

## Evidence anchors
| VA ID | Artifact | Files | Description |
|---|---|---|---|
| VA-S1 | Cascading tiers | `pkg/v09/sfu/topology.go` | `CascadingTiers` enumerates tier roles and segment assignments for 200+ participants. |
| VA-S2 | Inter-tier signaling | `pkg/v09/sfu/topology.go` | `SelectForwardingPath` defines control-plane signaling heuristics and appreciable degradations. |
| VA-S3/VA-S4 | Forwarding & failover | `pkg/v09/sfu/topology.go` | `FailoverDecision` expresses tier failover vs fallback to lower-scale topologies. |
| VA-S5/VA-S6 | Compatibility/governance | `docs/v0.9/phase3/p3-sfu-mesh.md` | Documents additive evolution guidance and major-path trigger examples for topology changes. |

## Planned-vs-implemented language
- Planned: any topology-breaking proposal must reference these governance examples before hitting a major-path review chair.
- Implemented: the CLI scenario executes `SelectForwardingPath` + `FailoverDecision`, ensuring the described heuristics remain deterministic and reproducible.
