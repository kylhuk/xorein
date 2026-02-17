# Phase 2 · P2-T1 History Sync & Merkle Contract

## Objective
Freeze the deterministic history-sync negotiation, lifecycle, and Merkle integrity requirements before any implementation touches `pkg/v07/history`. V7-G2 relies on this narrative to prove the sync contracts behave across retention, archivist, and security-mode boundaries.

## Contract
### Sync negotiation & lifecycle
- Stream identifiers, capability negotiation, and lifecycle checkpoints are recorded in `pkg/v07/history/contracts.go`. Every request/response pair must resolve to deterministic success, rejection, or resume outcomes depending on the peer’s supported epochs and retention windows.
- Mode-epoch boundaries (including locked history and optional re-sharing workflows) remain explicit metadata that lets a client render a locked-history state without guessing, and each transition is documented in the same sync contract file.

### Merkle construction & proof semantics
- Canonical chunking, domain separation, and root derivation tie to `pkg/v07/history/contracts.go`; identical history inputs must produce identical commitments, and proof exchange mismatches follow the remediation taxonomy spelled out in this contract.

### Retention-aware sync & archivist fallback
- Sync windows respect retention policy, and archivist-assisted sources obey preference and fallback rules so evidence remains deterministic even when data moves between retention tiers; these expectations are captured in `pkg/v07/history/contracts.go`.

## Evidence anchors
| Artifact | Description | Evidence anchor |
|---|---|---|
| `VA-H1` | Sync negotiation success/failure taxonomy | Section "Sync negotiation & lifecycle" |
| `VA-H2` | Request/response lifecycle transitions | Same section |
| `VA-H3` | Canonical Merkle chunking | Section "Merkle construction & proof semantics" |
| `VA-H4` | Proof verification and remediation | Same section |
| `VA-H5` | Retention-aware sync window semantics | Section "Retention-aware sync & archivist fallback" |
| `VA-H6` | Archivist-assisted source fallback | Same section |
| `VA-H7` | Mode-epoch segmentation and locked history signaling | Section "Sync negotiation & lifecycle" |

This doc tells V7-G2 reviewers how the history sync and Merkle contracts behave, references the future `pkg/v07/history` seams, and keeps every artifact traceable to the planned `VA-H*` evidence.
