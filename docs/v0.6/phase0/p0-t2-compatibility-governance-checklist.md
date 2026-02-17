# Phase 0 · P0-T2 Compatibility & Governance Checklist

## Purpose
Make V6-G0 gateable by cataloging every additive protobuf dependency implied by Section 6 tasks and embedding the major-change guardrails that protect planned-vs-implemented discipline.

## Contract items
- Enumerate every message or field addition the Phase 1–5 tasks demand, cite the enforcing gate (V6-G1..V6-G6), and link each row to the relevant `VA-*` artifact so reviewers see both the gate owner and the verification anchor.
- Capture the major-change trigger matrix from `TODO_v06.md`: list the downgrade-negotiation evidence, new multistream IDs, AEP obligations, and multi-implementation validation checkpoints any breaking candidate must supply before passing V6-G5/V6-G6.
- Reference the open-decision register (Section 8) and explain why OD6-01..OD6-03 remain `Open`, including the V6-G5 revisit plans, to prove no hidden resolution occurred.

### Compatibility matrix by task
| Compatibility item | Gate trigger | VA artifact | Evidence anchor |
|---|---|---|---|
| DHT freshness TTL headers (`VA-D1`) | V6-G1 adds TTL validation without new wire fields beyond optional freshness flag | `VA-D1` | `docs/v0.6/phase1/p1-discovery-hardening-contracts.md#deterministic-rule-table` |
| Partial search response diagnostics (`VA-S1`) | V6-G2 introduces deterministic reason codes but uses existing search schema | `VA-S1` | `docs/v0.6/phase2/p2-search-explore-preview-reliability.md#deterministic-rule-table` |
| PoW adaptation envelope metadata (`VA-A1`) | V6-G3 regulates difficulty bounds via config-only profile; no new wire IDs | `VA-A1` | `docs/v0.6/phase3/p3-anti-abuse-hardening-contracts.md#deterministic-rule-table` |
| Report routing idempotency keys (`VA-R2`) | V6-G4 reuses existing report schema plus deterministic routing metadata | `VA-R2` | `docs/v0.6/phase4/p4-reputation-report-filter-reliability.md#deterministic-rule-table` |
| Filter gating under SecurityMode (`VA-R3`) | V6-G4 constrains optional filters with server/E2EE flags; no new authoritative policy wire | `VA-R3` | same doc |

### Major-change trigger matrix
| Trigger area | Required evidence | Gate | Notes |
|---|---|---|---|
| Downgrade negotiation (`TODO_v06.md` Section 6) | Documented fallback negotiations before V6-G5, including AEP obligations | V6-G5 | Evidence anchored in `docs/v0.6/phase5/p5-t2-conformance-review.md` |
| New multistream IDs (search/discovery) | Multi-implementation validation plan plus upgrade signal | V6-G4/V6-G5 | Reason codes tracked in `VA-S*` and `VA-D*` tables |
| AEP obligations and multi-implementation checkpoints | Gate owner signoff verifying no new required clients | V6-G5 | References release dossier `docs/v0.6/phase5/p5-t3-release-gate-handoff.md` |

### Open decisions (Section 8 of `TODO_v06.md`)
| Decision | Current state | Revisit plan | Notes |
|---|---|---|---|
| OD6-01 · Default freshness TTL | `Open` | V6-G5 will review TTL tuning and publish evidence in `VA-D1` scenario row before handoff | TTL stays configurable; no default enforced yet so risk remains visible |
| OD6-02 · Default PoW adaptation envelope | `Open` | V6-G5 will leave envelope configurable and document envelope scenario outcomes in `VA-A1` before V6-G6 | Envelope is advisory; helpers in `pkg/v06/antiabuse` make bounds explicit for future signoff |
| OD6-03 · Default reputation weighting | `Open` | V6-G5 governance audit documents adjustable weighting/confidence in `VA-R1` before release | Reputation helper keeps anti-gaming and uncertainty annotations until decision closes |

This checklist keeps the open decisions auditable while proving no hidden resolutions occurred before V6-G5.
