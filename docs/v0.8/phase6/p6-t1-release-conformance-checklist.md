# Phase 6 · P6-T1 Release Conformance Checklist

## Status message
Planned vs implemented: v0.8 release conformance remains at the helper/checklist level. The checklist below records which criteria have been satisfied by the in-memory contracts (`pkg/v08/*`) and the deterministic scenario. Nothing here pretends to deliver a full release artifact yet.

## Conformance checklist
| Criterion | Evidence | Status |
|---|---|---|
| Gate readiness (S8-01..S8-07) tracked in code and docs | `pkg/v08/conformance/gates.go`, this doc | Implemented (helper level) |
| Scenario witness for deterministic contract output | `pkg/v08/scenario/echo.go`, `cmd/aether/main.go` (`v08-echo`) | Implemented |
| Bookmarks, preview, accessibility, themes, voice contracts documented for inspection | `pkg/v08/bookmarks/contracts.go`, `pkg/v08/linkpreview/contracts.go`, `pkg/v08/accessibility/contracts.go`, `pkg/v08/themes/contracts.go`, `pkg/v08/voice/dtln.go` | Planned (helpers exist but need broader validation) |
| Release notes placeholder referencing this evidence | `docs/v0.8/README.md` evidence table | Planned |

## Evidence anchor table
| VA ID | Artifact | Notes |
|---|---|---|
| VA-0802 | `pkg/v08/scenario/echo.go` + `cmd/aether` | Deterministic contract check and concise pass/fail output used for release witness |
| VA-0808 | `docs/v0.8/phase6/p6-t3-release-gate-handoff.md` | Captures open questions and the planned governance handoff beneath this checklist |

## Out-of-scope reminder
- Runtime distribution, signing, telemetry enabling, and platform-specific upgrades remain outside this release slice. The CLI scenario does not imply a runnable release; it is a deterministic conformance check only.

## Open decision register
- Open decisions recorded in `TODO_v01.md` (P8..P11) apply to this release slice and remain unresolved; treat them as the authoritative register for release gating until updated.
