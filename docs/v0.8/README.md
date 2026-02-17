# v0.8 Execution Slice

## Intentional Plan vs Implementation
This document set keeps the v0.8 execution slice explicitly planned: narrative text states what is implemented (deterministic contract helpers and the `v08-echo` CLI witness) while calling out remaining adaptations needed before production hardening. No doc claims a release-ready runtime; instead, each artifact links to the planned gate and the additive Go seam that can be exercised in further passes.

## Document map
| Area | Documents |
|---|---|
| Phase 0 · Scope & Compatibility | `phase0/p0-t1-scope-contract.md`, `phase0/p0-t2-compatibility-governance-checklist.md` |
| Phase 6 · Release Readiness | `phase6/p6-t1-release-conformance-checklist.md`, `phase6/p6-t3-release-gate-handoff.md` |

## Evidence anchors
| VA ID | Artifact | Description |
|---|---|---|
| VA-0801 | `pkg/v08/conformance/gates.go` | Enumerates S8-01..S8-07 scope gates and exposes checklist helpers for deterministic review loops |
| VA-0802 | `pkg/v08/scenario/echo.go` & `cmd/aether/main.go` | Runs the v0.8 echo scenario that exercises contract helpers and prints pass/fail output without network side effects |

## Out-of-scope reminders
- This phase explicitly leaves networking, storage persistence, and protobuf changes to future version bumps; only in-memory helpers exist. Please keep that separation at every gate review.
- No user-visible UI or API surface changes beyond `cmd/aether --scenario v08-echo` are intended in this slice.

## Open decision register
- See `TODO_v01.md` (open tasks P8..P11) for ongoing governance decisions that intersect with v0.8 theory work. Those items remain unresolved and are deliberately parked outside the scope of this planned implementation.
