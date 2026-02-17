# Open Decisions Register

Status: updated from maintainer decisions in `open_decisions_proposals.md`.

This register tracks the remaining unresolved decisions and keeps adopted baselines visible for auditability.

## 1) Adopted decision baselines

| Decision ID(s) | Adopted baseline | Propagated in planning docs |
|---|---|---|
| RM-01 | Naming baseline set: chat app is `Harmolyn`; protocol/backend/spec baseline is `xorein`; legacy `Aether` references retained only for traceability/migration context. | `TODO_v09.md`, `TODO_v10.md` |
| RM-02 | Relay participation is non-token and incentive-free; operators self-fund infrastructure. | `TODO_v09.md`, `TODO_v10.md` |
| RM-04 | Governance path: open-source now (AGPL code) with planned consortium model later; minimal legal/liability text remains a required artifact. | `AGENTS.md`, `TODO_v09.md`, `TODO_v10.md` |
| RM-05 | Decentralized continuity posture: initial single operator is allowed, but explicit operator succession/handover continuity is mandatory. | `TODO_v09.md`, `TODO_v10.md` |
| OD3-02 | Ranking tie-break baseline adopted (provisional): relevance -> trust -> recency -> deterministic lexical tie-break. | `TODO_v09.md`, `TODO_v10.md` |
| OD3-03 | RNNoise fallback remains mandatory through v0.9 baseline and is carried into v1.0 readiness checks. | `TODO_v09.md`, `TODO_v10.md` |
| OD3-04 | Discovery privacy default adopted: single-indexer query with rotation; multi-index parallel querying remains opt-in. | `TODO_v09.md`, `TODO_v10.md` |
| OD4-03 | Conservative auto-mod threshold posture adopted as moderation baseline. | `TODO_v09.md`, `TODO_v10.md` |
| OD5-01..OD5-05 | Bot delivery, Discord subset boundary, emoji retention, webhook signing, and SDK governance defaults adopted. | `TODO_v09.md`, `TODO_v10.md` |
| OD6-01..OD6-03 | Discovery hardening, PoW adaptation, and sparse-graph trust weighting defaults adopted. | `TODO_v09.md`, `TODO_v10.md` |
| OD7-01..OD7-04 | Replica placement, chunk size, scoped search ranking, and relay topology defaults adopted. | `TODO_v09.md`, `TODO_v10.md` |
| OD8-01..OD8-05 | v0.8 carry-forward defaults adopted (thread depth max 2, OG precedence, contrast policy, locale fallback, DTLN policy). | `TODO_v09.md`, `TODO_v10.md` |
| OD9-01..OD9-07 | v0.9 defaults adopted (retention, hierarchy depth, cascade aggressiveness, perf confidence, overload priority, battery policy, wake topology). | `TODO_v09.md`, `TODO_v10.md` |

## 2) Remaining open decisions

| Decision ID | Open question | Owner role | Revisit gate | Source |
|---|---|---|---|---|
| RM-03 | Exact hard limits per encryption/security mode and explicit policy when limits are reached (including no-silent-downgrade handling). | Performance Governance Lead + SecurityMode Contract Lead | `V9-G5`, `V9-G8`, carry-forward in `V10-G10` | `open_decisions_proposals.md`, `TODO_v09.md` |
| OD3-01 | Directory freshness/retention cliff and stale-state behavior when availability is fully peer-dependent. | Persistent Hosting Contract Lead | `V9-G1` | `open_decisions_proposals.md`, `TODO_v09.md` |
| OD4-01 | Exact race-window threshold for first-come-first-served moderator vs auto-mod actions. | Moderation Governance Lead | `V9-G8` | `open_decisions_proposals.md`, `TODO_v09.md` |
| OD4-02 | Exact rollback-window values by lifecycle stage (alpha/beta/live) for policy version rollback. | Moderation Governance Lead | `V9-G8`, carry-forward in `V10-G10` | `open_decisions_proposals.md`, `TODO_v09.md`, `TODO_v10.md` |

## 3) Totals

- Total tracked decisions (RM + OD3..OD9): 36
- Adopted baselines: 32
- Remaining open decisions: 4

## 4) Update rules

1. Resolve decisions in source planning files first, then update this register in the same change.
2. Do not present unresolved rows as settled architecture.
3. v1.0 handoff must explicitly report status for all remaining open decisions.
