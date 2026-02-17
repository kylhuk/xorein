# Phase 0 · P0-T3 Verification Evidence Matrix

## Purpose
Describe the pass/fail evidence schema that V6-G0 uses to certify every deterministic contract in `TODO_v06.md`.

## Contract
- Map each `VA-*` artifact from Section 6 to a preferred evidence type (documentation, deterministic helper, scenario pack, audit record), annotate the gate owner, and assign the reason-class triad (positive, negative, recovery) referenced in Section 1.3.
- Include a sweep matrix that ties every V6-G0 verification row back to the release-handoff dossier in `docs/v0.6/phase5/p5-t3-release-gate-handoff.md` so reviewers can trace the narrative through V6-G6.

### Artifact matrix by scope

#### Discovery hardening (`VA-D1`..`VA-D3`)
| Artifact | Task | Positive | Negative | Recovery | Reason-class | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-D1 | P1-T1 | Fresh DHT entry observed before TTL (`discovery.freshness.success`) | Stale or missing entry flagged (`discovery.freshness.stale`) | Retry/backoff path after prune (`discovery.freshness.retry`) | Freshness | V6-G1 owner | `docs/v0.6/phase1/p1-discovery-hardening-contracts.md#deterministic-rule-table` |
| VA-D2 | P1-T2 | Poisoning attempt blocked (`discovery.poison.detected`) | Conflicting signature stream emits `discovery.poison.invalid` | Waiting for alternate source (`discovery.poison.fallback`) | Poisoning | V6-G1 owner | same doc |
| VA-D3 | P1-T3 | Multi-source convergence success (`discovery.consistency.success`) | Divergent versions flagged (`discovery.consistency.conflict`) | Roll-forward to canonical source (`discovery.consistency.repair`) | Consistency | V6-G1 owner | same doc |

#### Search/explore/preview reliability (`VA-S1`..`VA-S3`)
| Artifact | Task | Positive | Negative | Recovery | Reason-class | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-S1 | P2-T1 | Query responds with complete page (`search.partial.success`) | Timeout/failure reason (`search.partial.failure`) | Fallback to cached partial page (`search.partial.fallback`) | Partial-failure | V6-G2 owner | `docs/v0.6/phase2/p2-search-explore-preview-reliability.md#deterministic-rule-table` |
| VA-S2 | P2-T2 | Explore feed honors freshness (`search.explore.success`) | Degradation emits `search.explore.degraded` | Freshness-guided rerank (`search.explore.recover`) | Explore ordering | V6-G2 owner | same doc |
| VA-S3 | P2-T3 | Preview aligns with join outcome (`search.preview.success`) | Mismatch surfaces advisory (`search.preview.mismatch`) | Recovery path defers to canonical preview (`search.preview.align`) | Preview alignment | V6-G2 owner | same doc |

#### Anti-Abuse hardening (`VA-A1`..`VA-A3`)
| Artifact | Task | Positive | Negative | Recovery | Reason-class | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-A1 | P3-T1 | PoW profile stays within envelope (`antiabuse.pow.success`) | Envelope violation flagged (`antiabuse.pow.failure`) | Adaptive retry after bound reset (`antiabuse.pow.retry`) | PoW envelope | V6-G3 owner | `docs/v0.6/phase3/p3-anti-abuse-hardening-contracts.md#deterministic-rule-table` |
| VA-A2 | P3-T2 | Limiter decision samples inside bounds (`antiabuse.limiter.ok`) | Throttled status (`antiabuse.limiter.throttled`) | Recovery when burst clears (`antiabuse.limiter.recover`) | Local limiter | V6-G3 owner | same doc |
| VA-A3 | P3-T3 | GossipSub peer reintegration allowed (`antiabuse.score.allow`) | Penalize under false-positive detection (`antiabuse.score.penalize`) | Controlled reintegration (`antiabuse.score.reintegrate`) | Peer scoring | V6-G3 owner | same doc |

#### Reputation/report/filter reliability (`VA-R1`..`VA-R3`)
| Artifact | Task | Positive | Negative | Recovery | Reason-class | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-R1 | P4-T1 | Reputation weighting stays within anti-gaming bounds (`reputation.weight.success`) | Weight throttled by uncertainty (`reputation.weight.failure`) | Confidence restoration (`reputation.weight.recover`) | Reputation weight | V6-G4 owner | `docs/v0.6/phase4/p4-reputation-report-filter-reliability.md#deterministic-rule-table` |
| VA-R2 | P4-T2 | Report routing idempotent and ordered (`reporting.route.accept`) | Duplicate or out-of-order detection (`reporting.route.failure`) | Retry with preserved ordering (`reporting.route.retry`) | Reporting route | V6-G4 owner | same doc |
| VA-R3 | P4-T3 | Optional filters execute under allowed security mode (`filters.process.success`) | Disallowed path blocked (`filters.process.blocked`) | Fallback to no-filter path (`filters.process.recover`) | Filter gating | V6-G4 owner | same doc |

#### Integrated validation (`VA-X1`..`VA-X3`)
| Artifact | Task | Positive | Negative | Recovery | Reason-class | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-X1 | P5-T1 | Scenario pack covers positive hardening flows | Scenario gap flagged | Recovery scenarios illustrate mitigations | Coverage | V6-G5 owner | `docs/v0.6/phase5/p5-t1-cross-feature-scenario-pack.md` |
| VA-X2 | P5-T2 | Governance review passes | Non-compliance raised | Open-decision escalation path | Conformance | V6-G5 owner | `docs/v0.6/phase5/p5-t2-conformance-review.md` |
| VA-X3 | P5-T3 | Release checklist complete | Missing evidence listed | Deferrals cataloged | Release handoff | V6-G6 owner | `docs/v0.6/phase5/p5-t3-release-gate-handoff.md` |
