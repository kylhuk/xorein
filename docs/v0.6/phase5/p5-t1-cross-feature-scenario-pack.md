# Phase 5 · P5-T1 Cross-Feature Scenario Pack

## Purpose
Deliver deterministic scenario coverage for positive, adverse, abuse, and recovery flows across discovery, search, anti-abuse, reputation, reporting, and filters so V6-G5 reviewers see the integrated reliability story before V6-G6.

## Contract
- `VA-X1` documents scenario bundles that span discovery (`VA-D*`), search/explore (`VA-S*`), anti-abuse (`VA-A*`), and reputation/report/filter (`VA-R*`) flows, with annotated reason-class triplets for each stage.
- Each scenario references the gate owner, enumerates the `S6-*` scope items it covers, and includes trace links to the originating docs plus the release gate dossier (`docs/v0.6/phase5/p5-t3-release-gate-handoff.md`).
- Recovery scenarios show how V6-G3 and V6-G4 guardrails are re-entrant even during degraded or abusive conditions so downstream teams can continue the deterministic narrative.

### Scenario pack table
| Scenario type | Scenario ID | Steps / phases involved | Key reason classes | Evidence anchor |
|---|---|---|---|---|
| Positive · smooth join | SC-PP1 | `P1-T1` publish → `P2-T3` preview → `P4-T1` reputation weight | `discovery.freshness.success`, `search.preview.success`, `reputation.weight.success` | `docs/v0.6/phase5/p5-t3-release-gate-handoff.md#scenario-pack-coverage` |
| Adverse · stale directory | SC-AD1 | `P1-T1` stale detection + retry → `P2-T2` explore degrader → `P4-T2` reporting retries | `discovery.freshness.retry`, `search.explore.degraded`, `reporting.route.retry` | same doc |
| Abuse · Sybil rush | SC-AB1 | `P3-T1` PoW envelope limit → `P3-T2` limiter throttle → `P3-T3` peer score penalize | `antiabuse.pow.failure`, `antiabuse.limiter.throttled`, `antiabuse.score.penalize` | same doc |
| Recovery · filter fallback | SC-RC1 | `P4-T3` SecurityMode block → on-device limited enforcement + reporting `P4-T2` acknowledgement | `filters.process.recover`, `reporting.route.accept`, `reputation.weight.recover` | same doc |

Every scenario row maps reason-class triads back to both phase-specific docs and the release dossier so auditors can replay each flow deterministically.
