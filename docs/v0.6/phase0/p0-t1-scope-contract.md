# Phase 0 · P0-T1 Scope Contract

## Goal
Anchor V6-G0 to the ten scope bullets listed in Section 7 of `TODO_v06.md`, tracing each `S6-*` item to its Phase 1–5 task and `VA-*` artifact so downstream validators can confirm no scope creep beyond the Sentinel hardening intent. The scope framing must explicitly cite the exclusions in Section 3 and keep the planned-vs-implemented guardrails visible in every downstream artifact.

## Contract
- Every `S6-*` bullet must be listed in the Phase 1–5 trace table before V6-G0 exits, with one-to-one mappings to the relevant `P*-T*` task and validation artifact (`VA-D*`, `VA-S*`, `VA-A*`, `VA-R*`, `VA-X*`).
- Reason-class continuity for each VA artifact must align with the taxonomy defined in Section 1.3 of `TODO_v06.md`, keeping discovery (`VA-D*`), search (`VA-S*`), anti-abuse (`VA-A*`), reputation/reporting/filters (`VA-R*`), and integrated (`VA-X*`) information auditable.
- Scope narrative must explicitly remind reviewers of Section 3 exclusions (anti-archive history, v0.7+ replays, etc.) so hardening/reliability remains the message rather than new feature introductions.

### Scope trace table
| S6 scope bullet | Phase task(s) | VA artifact(s) | Reason-class anchor | Evidence anchor |
|---|---|---|---|---|
| S6-01 · Server directory via DHT | `P1-T1` · freshness · `P1-T2` · poisoning · `P1-T3` · consistency | `VA-D1`, `VA-D2`, `VA-D3`, `VA-X1` | `discovery.freshness`, `discovery.poison`, `discovery.consistency` | `docs/v0.6/phase1/p1-discovery-hardening-contracts.md#deterministic-rule-table` |
| S6-02 · Global search | `P2-T1` | `VA-S1`, `VA-X1` | `search.partial` | `docs/v0.6/phase2/p2-search-explore-preview-reliability.md#deterministic-rule-table` |
| S6-03 · Explore/Discover tab | `P2-T2` | `VA-S2`, `VA-X1` | `search.explore` | same doc |
| S6-04 · Server preview before joining | `P2-T3` | `VA-S3`, `VA-X1` | `search.preview` | same doc |
| S6-05 · Proof-of-Work identity creation | `P3-T1` | `VA-A1`, `VA-X1` | `antiabuse.pow` | `docs/v0.6/phase3/p3-anti-abuse-hardening-contracts.md#deterministic-rule-table` |
| S6-06 · Per-user local rate limiting | `P3-T2` | `VA-A2`, `VA-X1` | `antiabuse.limiter` | same doc |
| S6-07 · GossipSub peer scoring | `P3-T3` | `VA-A3`, `VA-X1` | `antiabuse.score` | same doc |
| S6-08 · Web-of-trust reputation system | `P4-T1` | `VA-R1`, `VA-X1` | `reputation.weight` | `docs/v0.6/phase4/p4-reputation-report-filter-reliability.md#deterministic-rule-table` |
| S6-09 · Block user local mute + report routing | `P4-T2` | `VA-R2`, `VA-X1` | `reporting.route` | same doc |
| S6-10 · Optional per-server filters | `P4-T3` | `VA-R3`, `VA-X1` | `filters.process` | same doc |

This trace table keeps the hardening narrative explicit while connecting every bullet to deterministic phase artifacts and reason-class taxonomies.
