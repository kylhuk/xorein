# Phase 0 · P0-T2 Compatibility & Governance Checklist

## Governance posture
This checklist ties the v0.8 helper work to compatibility expectations. Each row notes whether the requirement is planned (doc-only) or implemented (Go helper or CLI scenario). No row claims a wider runtime completion beyond the additive, in-memory surfaces recorded here.

## Checklist
| Requirement | Status (planned vs implemented) | Evidence |
|---|---|---|
| Additive-only scope: no adjustments to protobuf schemas or wire encodings | Implemented | `pkg/v08/*` helpers keep changes limited to Go packages (VA-0801..VA-0807) |
| CLI scenario that exercises contracts without network I/O | Implemented | `cmd/aether --scenario v08-echo` invoking `pkg/v08/scenario/echo.go` (VA-0802) |
| Gate checklist referencing S8-01..S8-07 | Implemented | `pkg/v08/conformance/gates.go` exposes `Gates()` and `ValidateChecklist()` (VA-0801) |
| Bookmarks, link previews, themes, accessibility, and voice helpers documented for future review | Planned | `pkg/v08/*` contract files (VA-0803..VA-0807) and this doc set |

## Compatibility reminders
- Maintain deterministic behavior: all helper functions avoid global mutable state outside explicit structs (`Announcer`, `FocusGraph`, etc.).
- The CLI scenario prints only pass/fail messages and does not mutate network configuration or rely on secrets.

## Open decision register
- Ongoing governance questions (see `TODO_v01.md` P8..P11) remain out of scope but are noted here so reviewers can correlate them with this slice.
