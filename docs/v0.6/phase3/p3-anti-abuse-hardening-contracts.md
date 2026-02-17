# Phase 3 · P3 Anti-Abuse Hardening Contracts

## Purpose
Ensure V6-G3 can enforce deterministic anti-abuse behavior across PoW, rate limiting, and GossipSub scoring without implying privileged enforcement nodes.

## Contract
- `P3-T1` (`VA-A1`) records PoW difficulty/adaptation boundaries, reason-coded success/failure/retry outcomes (`antiabuse.pow.*`), and explains how the profile ties into the cross-feature scenario pack for positive/adverse flows.
- `P3-T2` (`VA-A2`) catalogs local limiter thresholds, burst/replay hardening, and recovery guidance (`antiabuse.limiter.*`), highlighting how clients are informed of throttling and can resume once limits reset.
- `P3-T3` (`VA-A3`) documents GossipSub peer scoring stability, reintegration, and false-positive controls with reason classes (`antiabuse.score.*`) so scoring penalties never leave downstream signals ambiguous.

### Deterministic rule table
| Input | Outcome | Reason-class | Fallback/Recovery | Evidence anchor |
|---|---|---|---|---|
| Desired PoW difficulty plus adaptive envelope bounds | Difficulty clamped within envelope; success reason `antiabuse.pow.success`, envelope violation raises `antiabuse.pow.failure` | `antiabuse.pow` | Adaptive retry resets to highest allowed bound (`antiabuse.pow.retry`) | `pkg/v06/antiabuse/contracts.go#PowProfile.Adapt` |
| Local token bucket depth vs. limiter threshold | Under threshold emits `antiabuse.limiter.ok`; burst threshold crossing emits `antiabuse.limiter.throttled` | `antiabuse.limiter` | Recovery when tokens replenish signals `antiabuse.limiter.recover` | `pkg/v06/antiabuse/contracts.go#DecideLimiter` |
| Peer score delta and reintegration window | Scores inside stability window produce `antiabuse.score.allow`; penalized peers mark `antiabuse.score.penalize` | `antiabuse.score` | Reintegration window expiry emits `antiabuse.score.reintegrate` | `pkg/v06/antiabuse/contracts.go#EvaluatePeerScore` |
