# Phase 1 · P1-T1 Store-Forward Retention & Archivist Contract

## Objective
Frame the deterministic DHT store-and-forward behavior, per-server retention policy, and volunteer archivist role so V7-G1 can freeze the durability semantics before any implementation touches `pkg/v07`.

## Contract
- _Store-and-forward semantics:_ The envelope describing ciphertext, metadata, TTL, and purge markers must resolve deterministically to retain or drop outcomes for the 30-day target while keeping relay-visible metadata opaque (`pkg/v07/storeforward/contracts.go`).
- _Replication policy:_ k=20 replica targets, placement, and repair triggers are bounded by deterministic thresholds, degrade cleanly when placement cannot be satisfied, and do not expose new privileged node classes (`pkg/v07/storeforward/contracts.go`).
- _Retention policy:_ Server-level retention overrides obey defaults, precedence, and conflict resolution rules documented in `pkg/v07/retention/contracts.go`, with audit events providing deterministic transition evidence (`VA-D3`, `VA-D4`).
- _Archivist role:_ Enrollment, withdrawal, storage obligations, and integrity checks remain voluntary; the role cannot elevate a peer to a privileged node class and each transition must record nondisputable metadata (`pkg/v07/archivist/contracts.go`).

### Store-forward envelope & TTL lifecycle
- The envelope metadata and TTL/purge markers resolve deterministically to the 30-day retention decision without leaking plaintext to relays, and the deterministic purge path is captured in `pkg/v07/storeforward/contracts.go`.

### Replication & repair obligations
- k=20 replication placement, repair triggers, and degraded signaling reside in `pkg/v07/storeforward/contracts.go` so every target/reserve transition is auditable and repeatable.

### Retention policy overrides
- Server overrides follow the precedence and conflict rules in `pkg/v07/retention/contracts.go`, and each change logs an audit event that proves deterministic transition behavior before any data purge occurs.

### Archivist capability lifecycle
- Archivist enrollment, advertisement, withdrawal, and integrity obligations remain voluntary capability semantics with explicit metadata that makes it impossible to treat the role as a privileged node class (`pkg/v07/archivist/contracts.go`).

## Evidence anchors
| Artifact | Description | Evidence anchor |
|---|---|---|
| `VA-D1` | TTL lifecycle compliance for 30-day store-and-forward | Section "Store-forward envelope & TTL lifecycle" in this doc |
| `VA-D2` | k=20 replication reachability and degraded-mode repair signaling | Section "Replication & repair obligations" in this doc |
| `VA-D3` | Retention override & audit contract | Section "Retention policy overrides" in this doc |
| `VA-D4` | Retention transition and purge semantics | Same section |
| `VA-D5` | Archivist enrollment and capability advertisement | Section "Archivist capability lifecycle" in this doc |
| `VA-D6` | Archivist obligations and integrity fallback | Same section |

This doc provides the planning narrative that will be reviewed at V7-G1, referencing the future `pkg/v07` seams and tying every durability artifact to an evidence anchor so the gate owner can mark the contract as frozen.
