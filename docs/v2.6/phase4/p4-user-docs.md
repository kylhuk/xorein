# Phase 4 user documentation (as-built truth)

This artifact holds the user-focused documentation that `G7` requires. It pairs planning statements with the as-built evidence we already have so the read-to-ship narrative stays honest: the right-hand column below lists the already documented journeys or reports that will serve as the truth once this file is published.

## Planning vs implemented boundary
| User topic | Planning note | As-built note/reference | Artifact | Evidence placeholder |
| --- | --- | --- | --- | --- |
| Onboarding + backup guidance | Plan to cover identity creation/restore, multi-device safety nets, and guaranteed paths for primary users. | The regression narrative in `docs/v2.6/phase2/p2-regression-report.md` already proves identity creation and backup journeys; this doc will cite those tests without claiming new coverage. | `docs/v2.6/phase2/p2-regression-report.md` | `EV-v26-G7-401` |
| Privacy model + offline/history/backfill behavior | Plan to spell out how private Space anti-enumeration, offline history, and backfill reliability behave in v2.6. | Current boundary and regression probe reports (`docs/v2.6/phase1/p1-boundary-report.md`) attest to these invariants; the user doc will refer to them rather than introducing new guarantees. | `docs/v2.6/phase1/p1-boundary-report.md` | `EV-v26-G7-402` |
| Operator-facing user experience (relay, attachments, turn) | Plan to describe how users experience relay failover, attachment uploads, and TURN negotiation. | The planned experience narrative mirrors the area already captured in `docs/v2.6/phase2/p2-regression-report.md` and the `F26` acceptance matrix, so no new runtime claims are made. | `docs/v2.5/phase4/f26-acceptance-matrix.md` | `EV-v26-G7-403` |
| Evidence checklist + deferral register link | Plan to append `docs/templates/roadmap-deferral-register.md` and show the user-facing deferral rationale. | Template placeholders already exist, so the final doc simply maps the evidence index to `EV-v26-G7-###` entries with cross-links to the deferral register. | `docs/templates/roadmap-deferral-register.md` | `EV-v26-G7-404` |

## Evidence & tooling references
- The user doc must mention how the deferral register (`docs/templates/roadmap-deferral-register.md`) ties into the terminal deferral policy in phase 0, emphasizing that no core deferrals survive `G7`.
- The evidence index for this doc follows `docs/templates/roadmap-evidence-index.md`; include the placeholder entries above until the actual `EV-v26-G7-###` records are available.
- The doc build command(s) producing the user doc should be captured with `EV-v26-G7-405` for the artifact tracker.

## Next steps (planning)
- Draft the onboarding/backfill/attachment sections using the regression and boundary reports to anchor the story without claiming unverified runtime changes.
- Link the privacy narratives to the `F26` threat model and the boundary probes so readers can trace the invariants to concrete tests.
- Capture the doc packaging command along with the `docs/templates/roadmap-gate-checklist.md` entry to move this artifact out of planning and into implemented status.
