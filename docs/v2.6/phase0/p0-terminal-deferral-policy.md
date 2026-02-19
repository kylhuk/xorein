# Phase 0 Terminal Deferral Policy (Planning)

This planning-only policy captures how the v2.6 terminal closure will treat deferrals. It mirrors the expectation from `docs/v2.5/phase4/f26-final-closure-spec.md` that deferred work is explicit and that core functionality ships with zero unresolved deferrals.

## Core principles
- **Zero core deferrals at closure.** Terminal `G10` sign-off cannot happen while any core feature listed in `F26` reports an open deferral. Any remaining friction must be reclassified as an enhancement-targeted deferral (see below) or explicitly cancelled.
- **Enhancements are post-v2.6.** Items that keep `F26` scope intact but can wait for future roadmap tracks must live in a labelled enhancement bucket; these are not blockers for `G0` as long as they are documented and segregated from the core register.
- **Audit-ready trace.** The final deferral register (later `docs/v2.6/phase5/p5-terminal-deferrals.md`) will cite this policy and the `EV-v26-G0-010` placeholder to prove the segregation rules were enforced.

## Deferral categories
| Category | Description | Closure criterion | Evidence placeholder |
| --- | --- | --- | --- |
| Core | Any functionality required by `F26` (identity, messaging, media, moderation, discovery, operator runbooks, packaging, docs, evidence). | Must be resolved before `G0` closes; set to zero in the terminal register. | `EV-v26-G0-011` (core clearance placeholder). |
| Enhancement | Items outside `F26` that can be deferred to a post-v2.6 track. | Documented with owner, risk, and follow-on track (trace to release backlog). | `EV-v26-G0-012` (enhancement register placeholder). |
| Cancelled | Formerly scoped work that is no longer needed. | Capture rationale and gate review note; no active follow-up. | `EV-v26-G0-013` (cancellation placeholder). |

## Enforcement notes
- The terminal deferral register must include a row for every `F26` requirement (drawn from `docs/v2.5/phase4/f26-acceptance-matrix.md`) even if it is already satisfied; that row should show `core deferrals = 0` in the final state.
- `G0` reviewers verify that each `EV-v26-G0-0##` entry above exists and that the deferral status is frozen before allowing later gates to rely on the register.
