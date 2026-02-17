# Phase 0 · P0-T1 Scope Contract

## Goal
Keep the v0.8 scope contract tightly bound to the listed helper domains while clearly labeling each artifact as planned (documented gate progression) or implemented (minimal helper code and the `v08-echo` scenario). Reviewers can trace every planned scope item to a doc anchor, the additive Go surface, and the gate owner without assuming full runtime completion.

## Contract items
- Document the seven S8 scope gates (S8-01..S8-07) with explicit ownership and the in-memory helper required for each; no wire-level changes or storage upgrades are part of this contract.
- Make every contract item reference the `pkg/v08/*` helper where the behavior is documented and, where implemented, include an evidence anchor with the VA ID from the README table.
- Keep the anti-scope-creep reminder visible: networking, persistent storage, and UI surfaces remain outside this contract (see the out-of-scope section below).

## Scope trace table
| Scope bullet | Phase doc | Implementation anchor | Gate owner | Evidence anchor |
|---|---|---|---|---|
| S8-01 · Deterministic contract helpers | This doc set | `pkg/v08/conformance/gates.go`, `pkg/v08/*` | V8-G1 | VA-0801 |
| S8-02 · v0.8 scenario witness | Phase 6 docs | `pkg/v08/scenario/echo.go`, `cmd/aether/main.go` | V8-G1 | VA-0802 |
| S8-03 · Bookmark privacy lifecycle | Phase 0 docs | `pkg/v08/bookmarks/contracts.go` | V8-G2 | VA-0803 (planned) |
| S8-04 · Link preview metadata policy | Phase 0 docs | `pkg/v08/linkpreview/contracts.go` | V8-G2 | VA-0804 (planned) |
| S8-05 · Theme/token validation | Phase 0 docs | `pkg/v08/themes/contracts.go` | V8-G3 | VA-0805 (planned) |
| S8-06 · Accessibility role/announcement contracts | Phase 0 docs | `pkg/v08/accessibility/contracts.go` | V8-G3 | VA-0806 (planned) |
| S8-07 · Voice suppression selection | Phase 0 docs | `pkg/v08/voice/dtln.go` | V8-G4 | VA-0807 (planned) |

## Out-of-scope reminder
- Protocol wire changes (protobuf, HTTP endpoints) remain deferred to future minor versions.
- Persistent storage, networking runtimes, and UI integration are not part of this breach-safe helper contract.

## Open decision register
- Refer to `TODO_v01.md` for unresolved decisions that impact this scope, especially the entries under P8..P11, which are intentionally excluded from this contract and flagged for later review.
