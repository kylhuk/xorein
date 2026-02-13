# TODO_v06.md

> Status: Planning artifact only. No implementation completion is claimed in this document.
>
> Authoritative v0.6 scope source: `aether-v3.md` roadmap bullets under **v0.6.0 — Sentinel** after Addendum A pull-forward alignment.
>
> Inputs used for sequencing and dependency posture: `TODO_v01.md` through `TODO_v05.md`, and `AGENTS.md`.
>
> Guardrails that are mandatory throughout this plan:
> - Documentation-only repository snapshot; strict planned-vs-implemented separation.
> - Protocol-first priority; UI is a consumer of protocol contracts.
> - Single binary network model (`--mode=client|relay|bootstrap`); no special node classes.
> - Protobuf minor changes are additive-only.
> - Breaking behavior requires new multistream IDs, downgrade negotiation evidence, AEP path, and multi-implementation validation.
> - Open decisions remain unresolved unless source docs explicitly resolve them.

---

## Stack Alignment Constraints (Parent Recommendation, Planning-Level)

- This section is recommendation-only planning guidance and does not claim implementation completion.
- Control plane default: libp2p secure channels use Noise_XX_25519_ChaChaPoly_SHA256 as the single supported suite; QUIC is preferred for reliable multiplexed streams, and this plan must not imply TCP-only operation.
- Media plane default: ICE (STUN/TURN), SRTP hop-by-hop, SFrame true media E2EE, and browser encoded-transform/insertable-streams integration where browser media clients apply.
- Key-management baseline carried forward: X3DH + Double Ratchet for DMs; MLS for group key agreement; any inherited Sender Keys mentions are compatibility/migration context only.
- Crypto defaults carried forward: SFrame AES-GCM full-tag default (for example AES_128_GCM_SHA256_128 intent), avoid short tags unless explicitly justified; messaging AEAD baseline ChaCha20-Poly1305 with optional AES-GCM negotiation; Noise suite fixed as above; SRTP baseline unchanged.
- Latency/resilience baseline carried forward for dependent realtime behavior: race direct ICE and relay/TURN in parallel, continuous path probing with seamless migration, RTT-aware multi-region relay/SFU selection with warm standby, dynamic topology switching (P2P 1:1, mesh small groups, SFU larger groups) with no SFU transcoding, and background resilience controls.

## 1. v0.6 Objective and Measurable Success Outcomes

### 1.1 Objective
Deliver **v0.6 Sentinel** as a protocol-first **hardening/scaling/reliability** planning artifact for discovery, moderation-adjacent trust/safety, and anti-abuse systems already introduced in earlier versions by defining:
- Reliability and abuse-resistance hardening for directory publication, browse/search/explore, and pre-join preview flows.
- Scale-stable contracts for PoW admission, local rate-limiting, and GossipSub scoring.
- Reliability hardening for report routing, reputation signals, and optional per-server filtering.
- Conformance evidence and release handoff without importing v0.7+ history/archive scope.

### 1.2 Measurable Success Outcomes
1. Discovery publication/index **hardening** contracts define deterministic stale-data, poisoning, and degradation handling over v0.3-introduced discovery surfaces.
2. Search/explore/preview **hardening** contracts define deterministic partial-failure behavior and fallback paths over previously introduced browse/preview flows.
3. PoW, local limiter, and peer-scoring contracts define deterministic anti-abuse behavior under adversarial load.
4. Report routing and moderation-escalation reliability semantics are deterministic under retry/reorder/failure conditions.
5. Optional filter contracts remain optional, local-policy bounded, and non-authoritative.
6. Hardening contracts preserve decentralized enforcement assumptions and avoid privileged authority-node semantics.
7. Compatibility/governance and open-decision conformance are complete and gate-auditable.

### 1.3 QoL integration contract for v0.6 hardening journeys (planning-level)

- **Unified health/recovery clarity under stress:** discovery/search/preview and anti-abuse hardening paths must expose deterministic user state plus next-action guidance when throttled, delayed, or partially degraded.
  - **Acceptance criterion:** degraded hardening scenarios document state, reason class, and recommended recovery action without ambiguous limbo states.
  - **Verification evidence:** `V6-G5` scenario bundle includes stress/degradation state-transition proofs tied to `VA-S*` and `VA-A*` artifacts.
- **Reason taxonomy continuity under abuse controls:** PoW, limiter, scoring, reputation, and report-routing user-visible outcomes remain mapped to stable reason classes.
  - **Verification evidence:** taxonomy coverage table is included in integrated conformance audit and release checklist.

---

## 2. Exact Scope Derivation from `aether-v3.md` for v0.6 Only

The following roadmap bullets remain the exact v0.6 scope, interpreted as **hardening/scaling/reliability depth** over prior introductions:

1. Server directory via DHT (public discovery records)
2. Global search (keyword/category/language/member count)
3. Explore/Discover tab
4. Server preview before joining
5. Proof-of-Work on identity creation (~5s Hashcash-style)
6. Per-user local rate limiting (5 msgs/5s)
7. GossipSub peer scoring (flood protection/penalties)
8. Web-of-trust reputation system
9. Block user (local mute) + report to server moderators
10. Optional per-server filters (keyword/regex + on-device ML image classification)

No additional capability outside these ten bullets is promoted into v0.6 in this plan.

---

## 3. Explicit Out-of-Scope and Anti-Scope-Creep Boundaries

### 3.1 Already introduced in earlier versions (not first introduction in v0.6)
- Directory publishing/browse, invite/request-to-join, optional indexer reference, and signed response verification (v0.3).
- Baseline RBAC/moderation/slow-mode semantics and decentralized moderation-event enforcement (v0.2).
- Advanced moderation governance (policy versioning + auto-mod hooks) introduction (v0.4).

### 3.2 Deferred to v0.7+
- Archive/deep-history/push-relay expansion and long-retention retrieval ecosystems.

### 3.3 Deferred to v0.8+/v0.9+/v1.x
- Later roadmap ecosystem expansions not explicitly listed in v0.6 bullets.

### 3.4 Anti-scope rules
1. v0.6 is hardening/scaling/reliability for introduced systems; it does not claim first introduction of discovery/admission primitives.
2. Optional filter features must not be reframed as centralized authoritative moderation platforms.
3. Any incompatible behavior must enter major-path governance; no silent minor-version absorption.
4. Open decisions remain open and explicitly tracked.

---

## 4. Entry Prerequisites (Assumed Completed)

### 4.1 v0.2 prerequisites
- Baseline RBAC/moderation events/slow-mode semantics required for report and moderation-route compatibility.

### 4.2 v0.3 prerequisites
- Discovery publication/browse, invite/request, and optional indexer non-authoritative posture already exist as baseline introductions; v0.6 only hardens their reliability/abuse resistance.

### 4.3 v0.4 prerequisites
- Advanced moderation governance baseline (policy versioning + auto-mod hooks) used as hardening context.

### 4.4 Dependency handling rule
- Missing prerequisites are blocking dependencies and are carried back to prior-version backlog.

---

## 5. Gate Model and Flow for v0.6

| Gate | Name | Entry Criteria | Exit Criteria |
|---|---|---|---|
| V6-G0 | Scope/guardrails/evidence lock | v0.6 planning initiated | Scope lock, prerequisites, exclusions, compatibility controls, evidence schema approved |
| V6-G1 | Discovery hardening freeze | V6-G0 passed | Publication/index reliability, poisoning controls, and stale-data handling specified |
| V6-G2 | Search/explore/preview reliability freeze | V6-G1 passed | Deterministic degraded/partial-failure/fallback behavior specified |
| V6-G3 | Anti-abuse hardening freeze | V6-G2 passed | PoW + local limiter + peer-scoring hardening semantics fully specified |
| V6-G4 | Reputation/report/filter reliability freeze | V6-G3 passed | Report routing reliability, reputation safeguards, optional-filter boundaries specified |
| V6-G5 | Integrated validation and governance readiness | V6-G4 passed | Cross-feature hardening scenarios + compatibility/open-decision checks complete |
| V6-G6 | Release conformance and handoff | V6-G5 passed | Traceability closure and execution handoff package approved |

---

## 6. Detailed v0.6 Execution Plan by Phase

Validation artifacts:
- `VA-D*` discovery hardening
- `VA-S*` search/explore/preview reliability
- `VA-A*` anti-abuse hardening
- `VA-R*` reputation/report/filter reliability
- `VA-X*` integrated conformance

### Phase 0 (V6-G0)
- [ ] **P0-T1** Scope trace mapping for all ten bullets with hardening interpretation rules.
- [ ] **P0-T2** Compatibility/governance checklists (additive + major-path triggers).
- [ ] **P0-T3** Gate evidence schema and pass/fail templates.

### Phase 1 (V6-G1): Discovery hardening
- [ ] **P1-T1** DHT publication reliability contract (freshness, stale pruning, retry/backoff) (`VA-D1`).
- [ ] **P1-T2** Index integrity and poisoning defense contract (`VA-D2`).
- [ ] **P1-T3** Multi-source consistency and deterministic conflict handling (`VA-D3`).

### Phase 2 (V6-G2): Search/explore/preview reliability
- [ ] **P2-T1** Deterministic partial-failure semantics for search queries/responses (`VA-S1`).
- [ ] **P2-T2** Explore-feed degradation, freshness, and fallback ordering (`VA-S2`).
- [ ] **P2-T3** Preview-to-join reliability and mismatch handling (`VA-S3`).

### Phase 3 (V6-G3): Anti-abuse hardening
- [ ] **P3-T1** PoW anti-sybil hardening profile and adaptation boundaries (`VA-A1`).
- [ ] **P3-T2** Local limiter boundary/recovery hardening under burst and replay-like traffic (`VA-A2`).
- [ ] **P3-T3** GossipSub score-stability, reintegration, and false-positive controls (`VA-A3`).

### Phase 4 (V6-G4): Reputation/report/filter reliability
- [ ] **P4-T1** Reputation anti-gaming and uncertainty-handling hardening (`VA-R1`).
- [ ] **P4-T2** Report routing reliability (idempotency, retry, ordering, acknowledgement) (`VA-R2`).
- [ ] **P4-T3** Optional filter reliability and bounded-policy behavior (`VA-R3`).

### Phase 5 (V6-G5 → V6-G6): Integrated conformance and handoff
- [ ] **P5-T1** Cross-feature hardening scenarios (positive/adverse/abuse/recovery) (`VA-X1`).
- [ ] **P5-T2** Compatibility/governance/open-decision conformance audit (`VA-X2`).
- [ ] **P5-T3** Release checklist + execution handoff dossier + v0.7+ deferral register (`VA-X3`).

---

## 7. Traceability Mapping

| Scope Item ID | v0.6 Scope Bullet | Primary Tasks | Validation Artifacts |
|---|---|---|---|
| S6-01 | Server directory via DHT | P1-T1, P1-T2, P1-T3 | VA-D1, VA-D2, VA-D3, VA-X1 |
| S6-02 | Global search | P2-T1 | VA-S1, VA-X1 |
| S6-03 | Explore/Discover tab | P2-T2 | VA-S2, VA-X1 |
| S6-04 | Server preview before joining | P2-T3 | VA-S3, VA-X1 |
| S6-05 | Proof-of-Work identity creation | P3-T1 | VA-A1, VA-X1 |
| S6-06 | Per-user local rate limiting | P3-T2 | VA-A2, VA-X1 |
| S6-07 | GossipSub peer scoring | P3-T3 | VA-A3, VA-X1 |
| S6-08 | Web-of-trust reputation system | P4-T1 | VA-R1, VA-X1 |
| S6-09 | Block user local mute + report routing | P4-T2 | VA-R2, VA-X1 |
| S6-10 | Optional per-server filters | P4-T3 | VA-R3, VA-X1 |

---

## 8. Open Decisions (Must Remain Unresolved)

| Decision ID | Open Question | Status | Revisit Gate |
|---|---|---|---|
| OD6-01 | Default freshness/TTL hardening profile for intermittently available discovery publishers. | Open | V6-G5 |
| OD6-02 | Default PoW difficulty adaptation envelope across heterogeneous client classes. | Open | V6-G5 |
| OD6-03 | Baseline reputation weighting policy under sparse trust graphs. | Open | V6-G5 |

Handling rule: open decisions remain `Open` and must not be represented as settled architecture.

---

## 9. Release-Conformance Checklist (V6-G6)

- [ ] All ten v0.6 bullets are mapped to tasks and artifacts.
- [ ] v0.6 is framed as hardening/scaling/reliability, not first introduction of v0.3 discovery/admission features.
- [ ] Discovery/search/explore/preview wording consistently describes hardening of previously introduced capabilities rather than first introduction.
- [ ] Optional filters remain optional and non-authoritative.
- [ ] Anti-abuse and report-routing reliability semantics are deterministic and test-mapped.
- [ ] No v0.7+ archive/history scope is imported.
- [ ] Compatibility/governance/open-decision checks are complete.
- [ ] Planned-vs-implemented distinction remains explicit.

---

## 10. Definition of Done for v0.6 Planning Artifact

This artifact is complete when it captures v0.6 hardening/scaling scope exactly, preserves protocol-first constraints, and provides gate-testable reliability contracts and handoff evidence without implementation claims.
