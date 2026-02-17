# Phase 1 · IPFS Persistent Hosting Contracts

## Plan vs implementation
The plan locks deterministic content addressing, pin lifecycle, retention, and degraded-mode behavior. Implementation is entirely in `pkg/v09/ipfs/contracts.go`; the document records how those helpers satisfy `VA-I1`..`VA-I6` and remains the authoritative reference for gate reviewers.

## Evidence anchors
| VA ID | Artifact | Files | Description |
|---|---|---|---|
| VA-I1 | Content envelope | `pkg/v09/ipfs/contracts.go` | `ContentMeta` + `ContentAddress` deterministically fold metadata into a stable identifier. |
| VA-I2 | Pinning lifecycle | `pkg/v09/ipfs/contracts.go` | `PinRole`, `PinLifecycleStage`, and lifecycle transition helpers encode server-owner responsibilities without creating privileged actors. |
| VA-I3/VA-I4 | Retention and degraded behavior | `pkg/v09/ipfs/contracts.go` | `ClassifyRetention` and `DegradedBehavior` enumerate states for in-scope retention horizons and failure responses. |
| VA-I5/VA-I6 | Governance notes | `docs/v0.9/phase1/p1-ipfs-contracts.md` | This document affirms additive evolution discipline for persistence schema and sketches major-path escalation examples for persistence breakage. |

## Planned-vs-implemented language
- Planned: future schema deltas must cite the checklist inside `phase0` before adding fields.
- Implemented: deterministic helpers already exercise `ClassifyRetention` and `DegradedBehavior` inside the CLI scenario so gate reviewers can rerun the explicit behavior without new infrastructure.
