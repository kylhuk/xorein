# Phase 5 terminal deferral register (P5-T1)

## Policy basis
This register applies `docs/v2.6/phase0/p0-terminal-deferral-policy.md` and `TODO_v26.md` terminalization constraints.

## Core deferral snapshot
| DeferralID | Feature area | Deferred from version | Status | Evidence / reason | Revisit target |
| --- | --- | --- | --- | --- | --- |
| DEF-v26-F26-001 | Final closure spec publication | v26 | resolved | `docs/v2.5/phase4/f26-final-closure-spec.md`; evidence recorded in P5 bundle (`EV-v26-G9-001`, `EV-v26-G10-001`). | n/a |
| DEF-v26-F26-002 | Proto delta readiness and compatibility posture | v26 | resolved | `docs/v2.5/phase4/f26-proto-delta.md`; evidence recorded (`EV-v26-G10-001`). | n/a |
| DEF-v26-F26-003 | Regression matrix coverage (identity, messaging, media, moderation, discovery) | v26 | resolved | `docs/v2.6/phase2/p2-regression-report.md`; evidence recorded (`EV-v26-G9-003`, `EV-v26-G10-002`). | n/a |
| DEF-v26-F26-004 | Release packaging reproducibility (binaries, docs, evidence catalog) | v26 | resolved | `scripts/v26-repro-build-verify.sh`; evidence recorded (`EV-v26-G9-004`, `EV-v26-G9-005`, `EV-v26-G10-004`, `EV-v26-G10-005`). | n/a |

## Core deferral status
- Core deferrals at terminal sign-off: **0 unresolved**.
- Enhancement deferrals: not registered for P5 closure.

## Deferral policy outcome
- This register is complete for the documented F26 core categories and records no open blockers against `G9`/`G10` terminal sign-off.
