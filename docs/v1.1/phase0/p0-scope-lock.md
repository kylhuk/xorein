# Phase 0 - Scope lock and dependencies (Planning only)

This artifact freezes the v11 in-scope list, deferred list, and dependency trace before full gate promotion. Everything below remains a planning statement; implementation artifacts will cite these anchors once produced.

## Evidence anchors
| Anchor | Description | Evidence placeholder |
|---|---|---|
| EV-v11-G0-001 | Scope lock narrative that aligns TODO_v11 items to in-scope/delayed buckets | `TODO_v11.md` + this file |
| EV-v11-G0-002 | Sign-off that makes scope/deferred lists mutually exclusive and traceable | Planning review notes (TBD) |
| EV-v11-G0-003 | Dependency map and carry-back policy for version isolation | `docs/v1.1/phase0/p0-traceability-matrix.md` (dependency column) |

## In scope (mutually exclusive bucket)
- Gate runner and promotion checklist enforcement (_Gates G2-G7 prep_)
- Relay data-boundary enforcement so relays remain transport-only (G3)
- Xorein naming baseline (`Xorein Relay`, `Xorein Protocol`, `Spaces`) and related runtime naming checks
- v12 identity + local backup spec package (G5 readiness)
- Podman relay smoke baseline for relay-mode runtime checks (G4)
- Promotion gate automation commands, status outputs, and automation hygiene (all gates)

## Deferred (scope explicitly excluded from v11)
- Public discovery/indexers and advanced moderation/runtime features (audio/video/screen-share)
- Enterprise services (calendar, recording, SSO, compliance, bots marketplace)
- Any runtime persistence paths that contradict the relay no-data-hosting policy

## Dependencies and carry-back policy
- Inputs remain anchored in `TODO_v11.md`, `TODO_v10.md`, `aether-v3.md`, `aether-addendum-qol-discovery.md`, `SPRINT_GUIDELINES.md`, and `AGENTS.md`.
- These docs will trace every in-scope item to a downstream artifact; any addition outside this bucket must follow a documented carry-back review and be appended to the deferred list before gating.

## Planned vs implemented
- **Planned:** Lock the lists above, ensure they are disjoint, and document their traceability/readiness before any code or phase-level artifact is produced.
- **Implemented:** (This planning doc) records the frozen bucket names and references the future gate/evidence work that will assert compliance once the implementation phase executes.
