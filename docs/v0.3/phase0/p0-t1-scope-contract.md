# v0.3 Phase 0 - P0-T1 Scope Contract and Traceability Map

> Status: Execution artifact. Scope lock and traceability now reference v0.3 phase docs and `pkg/v03` deliverables.

## Purpose

Freeze v0.3 boundaries, map every in-scope roadmap bullet to a discrete execution task, and keep the trace auditable through documented tasks and code packages.

## Source Trace

- `TODO_v03.md:42`
- `aether-v3.md:‡` (v0.3.0 Addendum A bullets)
- `AGENTS.md:25`

## v0.3 Scope-to-Task Trace Matrix

| Scope ID | v0.3 roadmap bullet | Primary tasks | Acceptance anchor | Evidence placeholder |
|---|---|---|---|---|
| S3-01 | RNNoise integration via C FFI | P1-T1 | Deterministic quality adaptation under noise | `pkg/v03/voice/contracts.go`, `pkg/v03/voice/contracts_test.go` |
| S3-02 | Opus adaptive bitrate (16-128 kbps) | P1-T1 | ABR transition rules + fallback state table | `pkg/v03/voice/contracts.go`, `pkg/v03/voice/contracts_test.go` |
| S3-03 | Adaptive jitter buffer (20-200ms) | P1-T1 | Buffer sizing + recovery taxonomy | `pkg/v03/voice/contracts.go`, `pkg/v03/voice/contracts_test.go` |
| S3-04 | FEC + DTX enabled | P1-T1 | Parity/repetition policy + DTX thresholds | `pkg/v03/voice/contracts.go`, `pkg/v03/voice/contracts_test.go` |
| S3-05 | Peer SFU election for 9+ voice sessions | P1-T2 | Election trigger/tie-break rules | `pkg/v03/voice/contracts.go`, `pkg/v03/voice/contracts_test.go` |
| S3-06 | Relay SFU mode (`--sfu-enabled=true`) | P1-T3 | Relay fallback semantics + compatibility matrix | `pkg/v03/voice/contracts.go`, `pkg/v03/voice/contracts_test.go` |
| S3-07 | Screen capture platform-native | P2-T1 | Capture source contract + security disclosure | `pkg/v03/screenshare/contracts.go`, `pkg/v03/screenshare/contracts_test.go` |
| S3-08 | Hardware encoder detection/selection | P2-T1 | Encoder capability table + selection determinism | `pkg/v03/screenshare/contracts.go`, `pkg/v03/screenshare/contracts_test.go` |
| S3-09 | Quality presets (Low/Standard/High/Ultra/Auto) | P2-T2 | Preset-to-parameter mapping table | `pkg/v03/screenshare/contracts.go`, `pkg/v03/screenshare/contracts_test.go` |
| S3-10 | Simulcast up to 3 layers | P2-T2 | Layer scheduling + degradation behavior | `pkg/v03/screenshare/contracts.go`, `pkg/v03/screenshare/contracts_test.go` |
| S3-11 | Viewer controls (fullscreen/PiP/zoom-pan) | P2-T3 | Control-state machine + render semantics | `pkg/v03/screenshare/contracts.go`, `pkg/v03/screenshare/contracts_test.go` |
| S3-12 | P2P file transfer up to 25MB chunked | P3-T1, P3-T3 | Chunking/integrity/retry contract across security modes | `pkg/v03/transfer/contracts.go`, `pkg/v03/transfer/contracts_test.go` |
| S3-13 | Image preview + attachment cards | P3-T2 | Inline rendering/metadata disclosure policy | `pkg/v03/transfer/contracts.go`, `pkg/v03/transfer/contracts_test.go` |
| S3-14 | DirectoryEntry publish + DHT retrieval | P4-T1 | Publish/withdraw/query deterministic flows | `pkg/v03/discovery/contracts.go`, `pkg/v03/discovery/contracts_test.go` |
| S3-15 | Explore/Discover + server preview | P4-T2 | Browse/preview failure taxonomy | `pkg/v03/discovery/contracts.go`, `pkg/v03/discovery/contracts_test.go` |
| S3-16 | Invite + request-to-join flow | P4-T3 | Invite/request lifecycle + policy matrix | `pkg/v03/discovery/contracts.go`, `pkg/v03/discovery/contracts_test.go` |
| S3-17 | Optional indexer + signed/verifiable responses | P5-T1, P5-T2, P5-T3 | Non-authoritative contract + signed response verification matrix | `pkg/v03/indexer/contracts.go`, `pkg/v03/indexer/contracts_test.go` |

## Explicit Non-Goals and Overlap Boundaries

- No v0.4 custom-role or channel override expansion.
- No v0.5 bot API, Discord shim, or slash commands.
- No v0.6 discovery/abuse hardening or v0.7 deep history/search expansions.
- Optional indexers remain non-authoritative; no discovery control-plane claim.

## Scope Control Checks (P0-T1-ST2)

- Every new proposal must first map to one `S3-01..S3-17` row before acceptance.
- Any work requiring incompatible behavior without a major-version plan is deferred to the governance backlog.
- Non-authoritative indexers are explicitly optional and reversible.

## Planned Evidence Anchors

- Scope lock signoff: `docs/v0.3/phase6/p6-t3-release-gate-handoff.md`.
- Traceability pick-up for downstream phases: this document and the phase-level baseline docs listed below.
