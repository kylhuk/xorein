# Phase 1 · P1 Discovery Hardening Contracts

## Purpose
Define deterministic reliability contracts for directory publication, index poisoning defense, and multi-source consistency so V6-G1 can assess compliance before advancing.

## Contract
- `P1-T1` (`VA-D1`) documents DHT publication retries, TTL validation, and stale-pruning behavior with reason-class labels (`discovery.freshness.*`) linked to the cross-feature scenario pack in `docs/v0.6/phase5/p5-t1-cross-feature-scenario-pack.md`.
- `P1-T2` (`VA-D2`) lays out the poisoning-detection guardrails, signature chasing, and fail-safe fallback with `discovery.poison.*` reason codes so auditors can see detection/invalidation/recovery triads.
- `P1-T3` (`VA-D3`) captures multi-source convergence rules plus deterministic conflict resolution paths (canonical source, conflict escalation, repair) so V6-G1 reviewers can confirm the `discovery.consistency.*` taxonomy ties into downstream artifacts.

### Deterministic rule table
| Input | Outcome | Reason-class | Fallback/Recovery | Evidence anchor |
|---|---|---|---|---|
| DHT entry age, TTL, retry budget | Entry marked `discovery.freshness.success` if age <= TTL; otherwise `discovery.freshness.stale` until retries exhaust (`discovery.freshness.retry`) | `discovery.freshness` | Retry/backoff helper re-publishes and triggers stale prune when TTL exceeded | `pkg/v06/discovery/contracts.go#AssessFreshness` |
| Signature validation across index peers | `discovery.poison.detected` when a signature mismatch occurs; escalates to `discovery.poison.invalid` if two successive signatures conflict | `discovery.poison` | Fallback pulls canonical index from deterministic conflict resolver | `pkg/v06/discovery/contracts.go#ClassifyPoisonAttempt` |
| Multi-source version vector comparison | Canonical source chosen deterministically (lexicographically minimal ID plus highest timestamp) with `discovery.consistency.success`; conflicting parents produce `discovery.consistency.conflict` | `discovery.consistency` | Repair path replays canonical source via resolver and ties back to scenario pack | `pkg/v06/discovery/contracts.go#ResolveMultiSourceConflict` |
