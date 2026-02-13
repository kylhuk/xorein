# TODO_v09.md

> Status: Planning artifact only. No implementation completion is claimed in this document.
>
> Authoritative v0.9 scope source: `aether-v3.md` roadmap bullets under **v0.9.0 — Forge**.
>
> Inputs used for sequencing, dependency posture, closure patterns, and guardrail carry-forward: `aether-v3.md`, `TODO_v01.md`, `TODO_v02.md`, `TODO_v03.md`, `TODO_v04.md`, `TODO_v05.md`, `TODO_v06.md`, `TODO_v07.md`, `TODO_v08.md`, and `AGENTS.md`.
>
> Guardrails that are mandatory throughout this plan:
> - Repository snapshot is documentation-only; maintain strict planned-vs-implemented separation.
> - Protocol-first priority is non-negotiable: protocol/spec contract is the product; client/UI behavior is downstream.
> - Network model invariant remains unchanged: single binary with mode flags `--mode=client|relay|bootstrap`; no privileged node classes.
> - Compatibility invariant remains mandatory: protobuf minor evolution is additive-only.
> - Breaking protocol behavior requires major-path governance evidence: new multistream IDs, downgrade negotiation, AEP flow, and validation by at least two independent implementations.
> - Open decisions remain unresolved unless source documentation explicitly resolves them.
> - Licensing alignment remains explicit: code licensing permissive MIT-like and protocol specification licensing CC-BY-SA.
> - v1.0+ and post-v1 capability is not promoted into v0.9 by implication, convenience, or adjacency.

## Sprint Guidelines Alignment

- This plan adopts `SPRINT_GUIDELINES.md` as the governing sprint policy baseline.
- Sprint model rule: one sprint maps to one minor version planning band and this document stays scoped to v0.9.
- Mandatory QoL target: this sprint must evidence at least one priority-journey improvement that achieves **10% less user effort**.
- Required closure gates: quality evidence, QA strategy traceability, review sign-off, and documentation plus release-note updates.
- Governance and status discipline remain mandatory: planned-vs-implemented separation stays explicit, and unresolved protocol decisions remain open unless authoritative sources resolve them.

---

## Stack Alignment Constraints (Parent Recommendation, Planning-Level)

- This section is recommendation-only planning guidance and does not claim implementation completion.
- Control plane default: libp2p secure channels use Noise_XX_25519_ChaChaPoly_SHA256 as the single supported suite; QUIC is preferred for reliable multiplexed streams, and this plan must not imply TCP-only operation.
- Media plane default: ICE (STUN/TURN), SRTP hop-by-hop, SFrame true media E2EE, and browser encoded-transform/insertable-streams integration where browser media clients apply.
- Key-management baseline carried forward: X3DH + Double Ratchet for DMs; MLS for group key agreement; any inherited Sender Keys mentions remain compatibility/migration context only.
- Crypto defaults carried forward: SFrame AES-GCM full-tag default (for example AES_128_GCM_SHA256_128 intent), avoid short tags unless explicitly justified; messaging AEAD baseline ChaCha20-Poly1305 with optional AES-GCM negotiation; Noise suite fixed as above; SRTP baseline unchanged.
- Latency/resilience baseline carried forward for dependent realtime behavior: race direct ICE and relay/TURN in parallel, continuous path probing with seamless migration, RTT-aware multi-region relay/SFU selection with warm standby, dynamic topology switching (P2P 1:1, mesh small groups, SFU larger groups) with no SFU transcoding, and background resilience controls.

## A. Status and Source-of-Truth Framing

### A.1 Planning-only status
This document is a v0.9 planning artifact for scope control, sequencing, validation design, and execution handoff readiness. It does not assert that implementation work has been completed.

### A.2 Source-of-truth hierarchy
1. `aether-v3.md` roadmap bullets for **v0.9.0 — Forge** are the sole authority for in-scope capability.
2. `AGENTS.md` provides mandatory repository/governance guardrails and framing constraints.
3. `TODO_v01.md` through `TODO_v08.md` provide continuity patterns for gate design, evidence discipline, and anti-scope-creep posture.

### A.3 Inherited framing constraints
- Protocol-first posture remains primary: protocol contract and interoperability are authoritative.
- Single-binary deployment model remains unchanged (`--mode=client|relay|bootstrap`).
- Compatibility/governance constraints apply to all protocol-touching deltas.
- Open decisions remain open unless explicitly resolved by source docs.

---

## B. v0.9 Objective and Measurable Success Outcomes

### B.1 Objective
Deliver **v0.9 Forge** as a protocol-first planning artifact that defines performance and scale contracts on top of v0.1-v0.8 baselines by specifying:
- IPFS integration for persistent file hosting with server-owner pinning posture.
- Large-server operation model including hierarchical GossipSub and lazy member loading.
- Cascading SFU mesh contract for 200+ participant voice scenarios.
- Cross-platform performance profiling and optimization planning envelope.
- Stress-testing contract for 1000-member server, 50-person voice, and latency benchmarks.
- Relay-node performance optimization and load-testing contract.
- Mobile battery optimization strategy via background-activity reduction constraints.

### B.2 Measurable success outcomes
1. Persistent-hosting contract defines deterministic IPFS pinning responsibilities, retention boundaries, and availability semantics for server-owner-managed pinning.
2. Hierarchical GossipSub and lazy-member-loading contracts define deterministic fanout and member-state loading behavior suitable for large-server planning targets.
3. Cascading SFU mesh contract defines deterministic topology transitions and relay/SFU layering behavior for 200+ participant voice planning scenarios.
4. Cross-platform profiling model defines measurement taxonomy, platform baselines, and optimization decision thresholds with repeatable evidence requirements.
5. Stress-test contract defines deterministic scenario classes and pass/fail thresholds for 1000-member server and 50-person voice benchmark campaigns.
6. Relay performance/load contract defines capacity envelopes, saturation behaviors, and degradation handling under load.
7. Mobile battery optimization contract defines background-activity budgets, wakeup boundaries, and energy-impact verification criteria.
8. Compatibility/governance controls and evidence model are complete and gate-auditable.
9. Integrated validation phase covers positive, adverse, saturation, and recovery pathways across all seven v0.9 bullets.
10. Release/handoff package provides complete traceability, explicit deferrals, and planning-only language integrity.

### B.3 QoL integration contract for v0.9 scale-performance journeys (planning-level)

- **Unified health and deterministic next action at scale:** high-load and saturation states in large-server, SFU, relay, and battery-sensitive flows must preserve explicit user state and recovery guidance.
  - **Acceptance criterion:** stress scenarios expose deterministic state/reason/action outcomes for throttling, saturation, and degradation conditions.
  - **Verification evidence:** `V9-G8` contains scale QoL score rows tied to `VA-S*`, `VA-R*`, and `VA-B*` artifacts.
- **Recovery-first call continuity at 200+ participant targets:** cascading-SFU degradation paths prioritize rejoin/switch-path/switch-device guidance and avoid ambiguous terminal outcomes.
  - **Verification evidence:** `V9-G3` and integrated validation include explicit recovery-path evidence with deterministic pass/fail criteria.
- **Cross-device continuity under performance constraints:** attention resume, read continuity, and interaction handoff remain deterministic despite profiling-driven optimizations.
  - **Verification evidence:** battery/perf optimization gates include continuity-regression checks with explicit fail conditions.

---

## C. Exact Scope Derivation from `aether-v3.md` (v0.9 Only)

The following roadmap bullets in `aether-v3.md` define v0.9 scope and are treated as exact inclusions:

1. IPFS integration for persistent file hosting (pinning by server owners)
2. Large server optimization: hierarchical GossipSub, lazy member loading
3. Cascading SFU mesh for 200+ participant voice
4. Performance profiling and optimization across all platforms
5. Stress testing: 1000-member server, 50-person voice, latency benchmarks
6. Relay node performance optimization and load testing
7. Battery optimization on mobile (background activity reduction)

No additional capability outside these seven bullets is promoted into v0.9 in this plan.

---

## D. Explicit Out-of-Scope / Anti-Scope-Creep Boundaries

### D.1 Deferred to v1.0+
- External security audit and publication of protocol specification v1.0.
- Full user/admin/developer/API documentation publication program.
- Landing/comparison sites, app-store release operations, bootstrap-infrastructure expansion.
- Community relay-node launch programs and reproducible-build release tracks.

### D.2 Deferred to post-v1 bands
- v1.1 bridge programs.
- v1.2 collaborative canvas capabilities.
- v1.3 forum/wiki/template tracks.
- v1.4 plugin/app-directory/reproducible-build ecosystem expansions.

### D.3 Anti-scope-creep enforcement rules for v0.9
1. Any proposal not traceable to one of the seven v0.9 bullets is rejected or formally deferred.
2. IPFS scope is bounded to persistent hosting + server-owner pinning semantics; no tokenized incentive system is introduced.
3. Large-server optimization scope is bounded to hierarchical GossipSub and lazy member loading; no unrelated feature expansion is imported.
4. Cascading voice topology scope is bounded to SFU mesh behavior for 200+ planning target; unrelated media-feature expansion is excluded.
5. Profiling/optimization scope is bounded to measurement and optimization planning across platforms; it does not import v1.0 publication tracks.
6. Stress-testing scope is bounded to the stated targets (1000-member server, 50-person voice, latency benchmarks).
7. Relay optimization scope is bounded to performance/load behavior; governance and role models remain consistent with prior constraints.
8. Mobile battery scope is bounded to background activity reduction and energy-budget planning constraints.
9. Any incompatible protocol behavior discovered during planning must enter major-path governance flow and cannot be silently absorbed into minor evolution.

---

## E. Entry Prerequisites from v0.1-v0.8 Outputs

v0.9 planning assumes prior-version contract outputs are available as dependencies.

### E.1 v0.1 prerequisite baseline
- Core networking, DHT, GossipSub, relay baseline, and mode-model assumptions exist.
- Single-binary operation model and foundational evidence discipline exist.

### E.2 v0.2 prerequisite baseline
- DM/social/presence/notification baselines exist for large-member-state and activity-surface assumptions.
- Baseline RBAC/moderation and slow-mode semantics exist and constrain large-scale abuse-handling assumptions.

### E.3 v0.3 prerequisite baseline
- Voice quality/SFU, screen-share, and file/media transfer baselines exist and constrain large-server voice/file performance assumptions under scale.
- Directory publishing/browse baseline exists and informs large-server discoverability assumptions under scale.
- Invite/request-to-join baseline exists and constrains admission and load-shaping assumptions.
- Optional community-run, non-authoritative indexer baseline and signed-response verification posture remain dependency context only.

### E.4 v0.4 prerequisite baseline
- Advanced moderation governance baseline (policy versioning + auto-mod hooks), RBAC/moderation/audit, and visibility policy context exist for large-server behavior boundaries.

### E.5 v0.5 prerequisite baseline
- API/event-surface and compatibility planning discipline exists for performance instrumentation and relay behavior touchpoints.

### E.6 v0.6 prerequisite baseline
- Discovery/safety/anti-spam controls exist and constrain large-server and relay optimization assumptions.

### E.7 v0.7 prerequisite baseline
- Durable history/store-forward/search/notification context exists for persistence and relay-load interaction assumptions.

### E.8 v0.8 prerequisite baseline
- Rich-content/theming/a11y/i18n and DTLN planning baseline exists for cross-platform profiling and battery-behavior interaction context.

### E.9 Carry-back dependency rule
- Missing prerequisites are blocking dependencies for affected v0.9 tasks.
- Missing prerequisites are carried back and are not silently re-scoped into v0.9.
- Gate owners must explicitly record carry-back status in evidence bundles.

---

## F. Gate Model and Flow (V9-G0..V9-G9)

### F.1 Gate definitions

| Gate | Name | Entry Criteria | Exit Criteria |
|---|---|---|---|
| V9-G0 | Scope/guardrails/evidence lock | v0.9 planning initiated | Scope lock, exclusions, prerequisites, compatibility controls, and evidence taxonomy approved |
| V9-G1 | IPFS persistent-hosting contract freeze | V9-G0 passed | IPFS pinning, retention, lifecycle, and failure/degraded behavior contracts complete |
| V9-G2 | Large-server optimization contract freeze | V9-G1 passed | Hierarchical GossipSub and lazy-member-loading contracts complete |
| V9-G3 | Cascading SFU mesh contract freeze | V9-G2 passed | 200+ participant cascading SFU topology and transition contracts complete |
| V9-G4 | Cross-platform profiling/optimization contract freeze | V9-G3 passed | Platform profiling taxonomy, baselines, and optimization decision rules complete |
| V9-G5 | Stress-testing contract freeze | V9-G4 passed | 1000-member/50-voice/latency benchmark scenario and threshold contracts complete |
| V9-G6 | Relay performance/load contract freeze | V9-G5 passed | Relay capacity, load-behavior, and degradation-handling contracts complete |
| V9-G7 | Mobile battery-optimization contract freeze | V9-G6 passed | Background-activity reduction and energy-budget validation contracts complete |
| V9-G8 | Integrated validation/governance readiness | V9-G7 passed | Cross-domain validation matrix, compatibility/governance audits, open-decision discipline complete |
| V9-G9 | Release conformance and execution handoff | V9-G8 passed | Full traceability closure, conformance checklist, handoff dossier, and explicit deferral register complete |

### F.2 Gate flow diagram

```mermaid
graph LR
  G0[V9 G0 Scope Guardrails Evidence Lock]
  G1[V9 G1 IPFS Persistent Hosting Contract Freeze]
  G2[V9 G2 Large Server Optimization Contract Freeze]
  G3[V9 G3 Cascading SFU Mesh Contract Freeze]
  G4[V9 G4 Cross Platform Profiling Contract Freeze]
  G5[V9 G5 Stress Testing Contract Freeze]
  G6[V9 G6 Relay Performance Load Contract Freeze]
  G7[V9 G7 Mobile Battery Optimization Contract Freeze]
  G8[V9 G8 Integrated Validation Governance Readiness]
  G9[V9 G9 Release Conformance Handoff]

  G0 --> G1 --> G2 --> G3 --> G4 --> G5 --> G6 --> G7 --> G8 --> G9
```

### F.3 Convergence rule
- **Single release-conformance exit:** V9-G9 is the only handoff exit for v0.9 planning.
- No phase is complete without explicit acceptance evidence linked to gate exit criteria.

---

## G. Detailed Execution Plan by Phase, Task, and Sub-Task

Priority legend:
- `P0` critical path
- `P1` high-value follow-through
- `P2` hardening and residual-risk controls

Validation artifact taxonomy IDs for v0.9:
- `VA-G*` scope/governance/evidence controls
- `VA-I*` IPFS persistent-hosting contracts
- `VA-L*` large-server optimization contracts
- `VA-S*` cascading SFU mesh contracts
- `VA-P*` cross-platform profiling/optimization contracts
- `VA-T*` stress-testing contracts
- `VA-R*` relay-performance/load contracts
- `VA-B*` battery-optimization contracts
- `VA-X*` integrated validation/governance artifacts
- `VA-H*` release-conformance/handoff artifacts

### Phase 0 - Scope, Governance, and Evidence Foundation (V9-G0)

- [ ] **[P0][Order 01] P0-T1 Freeze v0.9 scope contract and anti-scope boundaries**
  - **Objective:** Establish one-to-one mapping from the seven v0.9 bullets to task and artifact structure.
  - **Concrete actions:**
    - [ ] **P0-T1-ST1 Build v0.9 scope trace baseline (7 bullets to task families)**
      - **Objective:** Remove ambiguity in inclusion boundaries.
      - **Concrete actions:** Map each bullet to primary phase, acceptance anchors, and artifact IDs.
      - **Dependencies/prerequisites:** v0.9 scope extraction completed.
      - **Deliverables/artifacts:** Scope trace baseline (`VA-G1`).
      - **Acceptance criteria:** All 7 bullets mapped; no orphan and no extra capability.
      - **Suggested priority/order:** P0, Order 01.1.
      - **Risks/notes:** Unmapped scope introduces hidden execution gaps.
    - [ ] **P0-T1-ST2 Lock exclusion policy and escalation route**
      - **Objective:** Prevent v1.0+/post-v1 scope import.
      - **Concrete actions:** Define exclusions, escalation trigger criteria, and governance signoff path.
      - **Dependencies/prerequisites:** P0-T1-ST1.
      - **Deliverables/artifacts:** Exclusion/escalation policy (`VA-G2`).
      - **Acceptance criteria:** Gate submissions reference exclusion policy explicitly.
      - **Suggested priority/order:** P0, Order 01.2.
      - **Risks/notes:** Performance tracks have high adjacency to extra scope.
  - **Dependencies/prerequisites:** None.
  - **Deliverables/artifacts:** Scope control package (`VA-G1`, `VA-G2`).
  - **Acceptance criteria:** V9-G0 scope baseline approved.
  - **Suggested priority/order:** P0, Order 01.
  - **Risks/notes:** Scope drift here invalidates downstream planning.

- [ ] **[P0][Order 02] P0-T2 Lock compatibility/governance controls for v0.9 deltas**
  - **Objective:** Embed additive-evolution and major-path governance controls before domain freezes.
  - **Concrete actions:**
    - [ ] **P0-T2-ST1 Define additive-only protobuf checklist for v0.9 schema-touching surfaces**
      - **Objective:** Preserve minor-version compatibility invariants.
      - **Concrete actions:** Define field-addition constraints, reserved-field handling, and downgrade-safe defaults.
      - **Dependencies/prerequisites:** P0-T1.
      - **Deliverables/artifacts:** Additive schema checklist (`VA-G3`).
      - **Acceptance criteria:** All schema-touching tasks include checklist evidence.
      - **Suggested priority/order:** P0, Order 02.1.
      - **Risks/notes:** Hidden schema breaks harm interoperability.
    - [ ] **P0-T2-ST2 Define major-path trigger checklist for behavior-breaking proposals**
      - **Objective:** Enforce new multistream IDs + downgrade negotiation + AEP + multi-implementation validation.
      - **Concrete actions:** Specify mandatory evidence structure and escalation ownership.
      - **Dependencies/prerequisites:** P0-T2-ST1.
      - **Deliverables/artifacts:** Major-path trigger checklist (`VA-G4`).
      - **Acceptance criteria:** Any breaking candidate has complete governance package.
      - **Suggested priority/order:** P0, Order 02.2.
      - **Risks/notes:** Ambiguous triggers cause late governance conflict.
  - **Dependencies/prerequisites:** P0-T1.
  - **Deliverables/artifacts:** Compatibility/governance control pack (`VA-G3`, `VA-G4`).
  - **Acceptance criteria:** All protocol-touching tasks reference controls.
  - **Suggested priority/order:** P0, Order 02.
  - **Risks/notes:** Governance integrity is non-negotiable.

- [ ] **[P0][Order 03] P0-T3 Establish verification matrix and gate-evidence schema for v0.9**
  - **Objective:** Standardize evidence packaging and deterministic gate decisions.
  - **Concrete actions:**
    - [ ] **P0-T3-ST1 Define requirement-to-validation matrix template**
      - **Objective:** Ensure every scope bullet has positive, adverse, saturation, and recovery coverage.
      - **Concrete actions:** Define matrix fields for requirement ID, task IDs, artifact IDs, gate ownership, and evidence status.
      - **Dependencies/prerequisites:** P0-T1.
      - **Deliverables/artifacts:** Validation matrix template (`VA-G5`).
      - **Acceptance criteria:** Template supports all 7 bullets and all v0.9 gates.
      - **Suggested priority/order:** P0, Order 03.1.
      - **Risks/notes:** Weak template quality creates inconsistent gate closure.
    - [ ] **P0-T3-ST2 Define gate evidence-bundle conventions**
      - **Objective:** Normalize review quality across phase owners.
      - **Concrete actions:** Define bundle sections: scope references, decision logs, risk updates, checklists, open decisions.
      - **Dependencies/prerequisites:** P0-T3-ST1.
      - **Deliverables/artifacts:** Gate evidence schema (`VA-G6`).
      - **Acceptance criteria:** Every gate has auditable pass/fail package format.
      - **Suggested priority/order:** P0, Order 03.2.
      - **Risks/notes:** Inconsistent packaging impairs governance traceability.
  - **Dependencies/prerequisites:** P0-T1, P0-T2.
  - **Deliverables/artifacts:** Evidence baseline (`VA-G5`, `VA-G6`).
  - **Acceptance criteria:** V9-G0 exits only with approved evidence model.
  - **Suggested priority/order:** P0, Order 03.
  - **Risks/notes:** Missing evidence discipline creates rework later.

### Phase 1 - IPFS Persistent Hosting Contracts (V9-G1)

- [ ] **[P0][Order 04] P1-T1 Define IPFS content-addressing and pinning responsibility model**
  - **Objective:** Specify deterministic persistent-hosting contract with server-owner pinning posture.
  - **Concrete actions:**
    - [ ] **P1-T1-ST1 Define content envelope and addressing metadata contract**
      - **Objective:** Ensure deterministic content identity and lifecycle references.
      - **Concrete actions:** Define content identifiers, metadata constraints, and mapping boundaries with existing file contexts.
      - **Dependencies/prerequisites:** P0-T2, v0.3 file-transfer baseline, v0.7 retention baseline.
      - **Deliverables/artifacts:** Content-address contract (`VA-I1`).
      - **Acceptance criteria:** Equivalent content/metadata inputs resolve to deterministic address model.
      - **Suggested priority/order:** P0, Order 04.1.
      - **Risks/notes:** Address ambiguity undermines persistence guarantees.
    - [ ] **P1-T1-ST2 Define server-owner pinning roles and lifecycle responsibilities**
      - **Objective:** Bound operational responsibility without creating privileged protocol actors.
      - **Concrete actions:** Specify pin initiation/update/unpin semantics, authorization assumptions, and visibility of pin states.
      - **Dependencies/prerequisites:** P1-T1-ST1, v0.4 role/policy baseline.
      - **Deliverables/artifacts:** Pinning responsibility contract (`VA-I2`).
      - **Acceptance criteria:** Responsibility model is deterministic and role-bounded.
      - **Suggested priority/order:** P0, Order 04.2.
      - **Risks/notes:** Role ambiguity can create persistence inconsistency.
  - **Dependencies/prerequisites:** P0-T1 through P0-T3.
  - **Deliverables/artifacts:** IPFS addressing/pinning package (`VA-I1`, `VA-I2`).
  - **Acceptance criteria:** V9-G1 addressing/pinning criteria met with evidence links.
  - **Suggested priority/order:** P0, Order 04.
  - **Risks/notes:** Scope remains persistent hosting only.

- [ ] **[P0][Order 05] P1-T2 Define persistent-hosting retention, retrieval, and degraded behavior**
  - **Objective:** Specify deterministic retention/retrieval semantics and degradation handling.
  - **Concrete actions:**
    - [ ] **P1-T2-ST1 Define retention horizon and retrieval boundary semantics**
      - **Objective:** Ensure predictable availability expectations.
      - **Concrete actions:** Define retention states, retrieval resolution outcomes, and stale/missing-content handling taxonomy.
      - **Dependencies/prerequisites:** P1-T1, v0.7 history/retention context.
      - **Deliverables/artifacts:** Retention/retrieval contract (`VA-I3`).
      - **Acceptance criteria:** Retrieval outcomes are deterministic for present, stale, and missing content.
      - **Suggested priority/order:** P0, Order 05.1.
      - **Risks/notes:** Weak degraded semantics harm user trust.
    - [ ] **P1-T2-ST2 Define degraded-mode fallback and operator diagnostics expectations**
      - **Objective:** Provide deterministic behavior under partial pin availability.
      - **Concrete actions:** Define fallback ordering, operator evidence fields, and failure classification for unpinned/unreachable content.
      - **Dependencies/prerequisites:** P1-T2-ST1.
      - **Deliverables/artifacts:** Degraded behavior contract (`VA-I4`).
      - **Acceptance criteria:** Degraded outcomes are explicit and repeatable.
      - **Suggested priority/order:** P0, Order 05.2.
      - **Risks/notes:** Missing diagnostics delays triage.
  - **Dependencies/prerequisites:** P1-T1.
  - **Deliverables/artifacts:** Persistence behavior package (`VA-I3`, `VA-I4`).
  - **Acceptance criteria:** V9-G1 retention/degraded criteria met.
  - **Suggested priority/order:** P0, Order 05.
  - **Risks/notes:** Avoid importing unrelated storage-economics scope.

- [ ] **[P1][Order 06] P1-T3 Define IPFS governance boundary and interoperability notes**
  - **Objective:** Keep IPFS planning compatible with protocol governance and versioning constraints.
  - **Concrete actions:**
    - [ ] **P1-T3-ST1 Define additive evolution notes for persistence-related schema touches**
      - **Objective:** Preserve minor-version compatibility behavior.
      - **Concrete actions:** Record extension hooks, reserved fields, and backward-compatible defaults.
      - **Dependencies/prerequisites:** P1-T1, P0-T2.
      - **Deliverables/artifacts:** IPFS evolution note (`VA-I5`).
      - **Acceptance criteria:** All schema-delta notes are additive and auditable.
      - **Suggested priority/order:** P1, Order 06.1.
      - **Risks/notes:** Non-additive drift creates compatibility debt.
    - [ ] **P1-T3-ST2 Define major-path escalation examples for persistence behavior breakage**
      - **Objective:** Operationalize governance triggers.
      - **Concrete actions:** Document concrete trigger patterns requiring new multistream IDs + downgrade + AEP + multi-implementation validation.
      - **Dependencies/prerequisites:** P1-T3-ST1.
      - **Deliverables/artifacts:** IPFS major-path trigger annex (`VA-I6`).
      - **Acceptance criteria:** Breaking trigger examples are explicit and reusable at gate review.
      - **Suggested priority/order:** P1, Order 06.2.
      - **Risks/notes:** Vague examples reduce enforcement quality.
  - **Dependencies/prerequisites:** P1-T1, P1-T2.
  - **Deliverables/artifacts:** IPFS governance package (`VA-I5`, `VA-I6`).
  - **Acceptance criteria:** V9-G1 exits with governance-consistent IPFS contract set.
  - **Suggested priority/order:** P1, Order 06.
  - **Risks/notes:** IPFS scope remains v0.9-only.

### Phase 2 - Large-Server Optimization Contracts (V9-G2)

- [ ] **[P0][Order 07] P2-T1 Define hierarchical GossipSub topology contract for large servers**
  - **Objective:** Specify deterministic hierarchy behavior and fanout boundaries.
  - **Concrete actions:**
    - [ ] **P2-T1-ST1 Define hierarchy roles and topic-segmentation strategy**
      - **Objective:** Standardize large-server pubsub partitioning assumptions.
      - **Concrete actions:** Define hierarchy layers, topic segmentation boundaries, and relay/peer role assumptions without privileged protocol classes.
      - **Dependencies/prerequisites:** v0.1 GossipSub baseline, P0-T2.
      - **Deliverables/artifacts:** Hierarchical topology contract (`VA-L1`).
      - **Acceptance criteria:** Hierarchy behavior is deterministic and protocol-first.
      - **Suggested priority/order:** P0, Order 07.1.
      - **Risks/notes:** Role ambiguity can create uneven load.
    - [ ] **P2-T1-ST2 Define propagation and backpressure semantics under high membership**
      - **Objective:** Bound message fanout and congestion behavior.
      - **Concrete actions:** Define propagation thresholds, backpressure signals, and degraded delivery behavior.
      - **Dependencies/prerequisites:** P2-T1-ST1.
      - **Deliverables/artifacts:** Propagation/backpressure contract (`VA-L2`).
      - **Acceptance criteria:** High-load propagation outcomes are deterministic.
      - **Suggested priority/order:** P0, Order 07.2.
      - **Risks/notes:** Weak backpressure semantics amplify floods.
  - **Dependencies/prerequisites:** P0-T1 through P0-T3.
  - **Deliverables/artifacts:** Hierarchical pubsub package (`VA-L1`, `VA-L2`).
  - **Acceptance criteria:** V9-G2 hierarchical GossipSub criteria met.
  - **Suggested priority/order:** P0, Order 07.
  - **Risks/notes:** Must remain within v0.9 large-server bullet.

- [ ] **[P0][Order 08] P2-T2 Define lazy member-loading model and state-coherence boundaries**
  - **Objective:** Specify deterministic lazy loading for member state in large communities.
  - **Concrete actions:**
    - [ ] **P2-T2-ST1 Define member state classes and incremental-loading triggers**
      - **Objective:** Standardize when and how member subsets load.
      - **Concrete actions:** Define active/nearby/passive member state classes and deterministic load triggers.
      - **Dependencies/prerequisites:** P2-T1, v0.2 presence baseline.
      - **Deliverables/artifacts:** Lazy loading trigger contract (`VA-L3`).
      - **Acceptance criteria:** Load-trigger behavior is deterministic and test-mappable.
      - **Suggested priority/order:** P0, Order 08.1.
      - **Risks/notes:** Trigger ambiguity causes state churn.
    - [ ] **P2-T2-ST2 Define consistency guarantees and staleness handling**
      - **Objective:** Bound stale-view behavior and recovery expectations.
      - **Concrete actions:** Define staleness windows, refresh semantics, and fallback behavior when member snapshots diverge.
      - **Dependencies/prerequisites:** P2-T2-ST1.
      - **Deliverables/artifacts:** State coherence contract (`VA-L4`).
      - **Acceptance criteria:** Stale/refresh outcomes are deterministic.
      - **Suggested priority/order:** P0, Order 08.2.
      - **Risks/notes:** Inconsistent member state undermines moderation and presence trust.
  - **Dependencies/prerequisites:** P2-T1.
  - **Deliverables/artifacts:** Lazy-loading package (`VA-L3`, `VA-L4`).
  - **Acceptance criteria:** V9-G2 lazy-loading criteria met.
  - **Suggested priority/order:** P0, Order 08.
  - **Risks/notes:** Avoid importing non-scope UI redesign tracks.

- [ ] **[P1][Order 09] P2-T3 Define large-server optimization observability and rollback boundaries**
  - **Objective:** Provide deterministic instrumentation and fallback posture for hierarchy/lazy-load behavior.
  - **Concrete actions:**
    - [ ] **P2-T3-ST1 Define observability signals for hierarchy and lazy-load performance**
      - **Objective:** Ensure planning-level metrics support validation and tuning.
      - **Concrete actions:** Define signal catalog for fanout pressure, load latency, staleness rates, and recovery events.
      - **Dependencies/prerequisites:** P2-T1, P2-T2.
      - **Deliverables/artifacts:** Large-server observability contract (`VA-L5`).
      - **Acceptance criteria:** Signal set fully covers hierarchy and lazy-load behavior.
      - **Suggested priority/order:** P1, Order 09.1.
      - **Risks/notes:** Sparse telemetry hides failure modes.
    - [ ] **P2-T3-ST2 Define fallback/rollback decision thresholds**
      - **Objective:** Bound when to revert to safer behavior envelopes.
      - **Concrete actions:** Define threshold triggers, rollback sequencing assumptions, and compatibility-safe fallback behavior.
      - **Dependencies/prerequisites:** P2-T3-ST1.
      - **Deliverables/artifacts:** Rollback threshold contract (`VA-L6`).
      - **Acceptance criteria:** Rollback triggers are deterministic and governance-aligned.
      - **Suggested priority/order:** P1, Order 09.2.
      - **Risks/notes:** Poor rollback criteria prolong instability.
  - **Dependencies/prerequisites:** P2-T1, P2-T2.
  - **Deliverables/artifacts:** Large-server hardening package (`VA-L5`, `VA-L6`).
  - **Acceptance criteria:** V9-G2 exits with observability and rollback boundaries complete.
  - **Suggested priority/order:** P1, Order 09.
  - **Risks/notes:** Keep scope inside large-server bullet.

### Phase 3 - Cascading SFU Mesh Contracts for 200+ Voice (V9-G3)

- [ ] **[P0][Order 10] P3-T1 Define cascading SFU topology and election layering contract**
  - **Objective:** Specify deterministic cascade topology for 200+ participant voice targets.
  - **Concrete actions:**
    - [ ] **P3-T1-ST1 Define tiered SFU roles, segment boundaries, and election assumptions**
      - **Objective:** Standardize cascade structure while preserving single-binary mode model.
      - **Concrete actions:** Define tier roles, segment assignment logic, and election boundaries across peer/relay SFU candidates.
      - **Dependencies/prerequisites:** v0.3 voice/media baseline, P0-T2.
      - **Deliverables/artifacts:** Cascading topology contract (`VA-S1`).
      - **Acceptance criteria:** Segment/election behavior is deterministic and role-safe.
      - **Suggested priority/order:** P0, Order 10.1.
      - **Risks/notes:** Topology ambiguity creates instability at scale.
    - [ ] **P3-T1-ST2 Define inter-tier signaling and control-channel behavior**
      - **Objective:** Bound cascade-control chatter and transition timing.
      - **Concrete actions:** Specify signaling semantics, timeout handling, and degraded fallback when upstream tier is unavailable.
      - **Dependencies/prerequisites:** P3-T1-ST1.
      - **Deliverables/artifacts:** Inter-tier signaling contract (`VA-S2`).
      - **Acceptance criteria:** Control behavior is deterministic under nominal and degraded states.
      - **Suggested priority/order:** P0, Order 10.2.
      - **Risks/notes:** Unbounded control traffic can destabilize large sessions.
  - **Dependencies/prerequisites:** P0-T1 through P0-T3, v0.3 voice/media baseline.
  - **Deliverables/artifacts:** Cascading topology package (`VA-S1`, `VA-S2`).
  - **Acceptance criteria:** V9-G3 topology criteria met.
  - **Suggested priority/order:** P0, Order 10.
  - **Risks/notes:** Must not import unrelated video-feature scope.

- [ ] **[P0][Order 11] P3-T2 Define media forwarding, failover, and degradation semantics in cascades**
  - **Objective:** Specify deterministic forwarding and failover behavior across cascade tiers.
  - **Concrete actions:**
    - [ ] **P3-T2-ST1 Define forwarding path selection and quality-preservation boundaries**
      - **Objective:** Ensure predictable media path decisions at high participant counts.
      - **Concrete actions:** Define path-selection heuristics, jitter/latency boundaries, and forward-only assumptions.
      - **Dependencies/prerequisites:** P3-T1.
      - **Deliverables/artifacts:** Forwarding path contract (`VA-S3`).
      - **Acceptance criteria:** Path decisions are deterministic under equivalent conditions.
      - **Suggested priority/order:** P0, Order 11.1.
      - **Risks/notes:** Path ambiguity produces inconsistent quality outcomes.
    - [ ] **P3-T2-ST2 Define tier failover, rejoin, and split/merge behavior**
      - **Objective:** Bound recovery behavior when tiers disconnect or overload.
      - **Concrete actions:** Define failover triggers, re-election windows, split/merge semantics, and fallback-to-lower-scale topology.
      - **Dependencies/prerequisites:** P3-T2-ST1.
      - **Deliverables/artifacts:** Failover/degradation contract (`VA-S4`).
      - **Acceptance criteria:** Failure and recovery outcomes are deterministic.
      - **Suggested priority/order:** P0, Order 11.2.
      - **Risks/notes:** Undefined failover behavior can collapse sessions.
  - **Dependencies/prerequisites:** P3-T1.
  - **Deliverables/artifacts:** Cascading forwarding package (`VA-S3`, `VA-S4`).
  - **Acceptance criteria:** V9-G3 forwarding/failover criteria met.
  - **Suggested priority/order:** P0, Order 11.
  - **Risks/notes:** Preserve existing E2EE posture assumptions.

- [ ] **[P1][Order 12] P3-T3 Define governance and compatibility boundaries for cascade evolution**
  - **Objective:** Ensure cascade behavior changes remain compatibility-safe and governance-compliant.
  - **Concrete actions:**
    - [ ] **P3-T3-ST1 Define additive evolution guidance for cascade capability signaling**
      - **Objective:** Preserve minor-path compatibility.
      - **Concrete actions:** Define capability field extension boundaries and downgrade-safe defaults.
      - **Dependencies/prerequisites:** P3-T1, P0-T2.
      - **Deliverables/artifacts:** Cascade evolution note (`VA-S5`).
      - **Acceptance criteria:** Evolution guidance aligns with additive-only discipline.
      - **Suggested priority/order:** P1, Order 12.1.
      - **Risks/notes:** Capability drift can fragment interoperability.
    - [ ] **P3-T3-ST2 Define major-path trigger examples for topology-breaking changes**
      - **Objective:** Make major-path governance enforceable.
      - **Concrete actions:** Capture trigger patterns requiring new multistream IDs, downgrade negotiation, AEP, and multi-implementation validation.
      - **Dependencies/prerequisites:** P3-T3-ST1.
      - **Deliverables/artifacts:** Cascade major-path trigger annex (`VA-S6`).
      - **Acceptance criteria:** Trigger examples are explicit and gate-reusable.
      - **Suggested priority/order:** P1, Order 12.2.
      - **Risks/notes:** Under-specified triggers risk governance bypass.
  - **Dependencies/prerequisites:** P3-T1, P3-T2.
  - **Deliverables/artifacts:** Cascading governance package (`VA-S5`, `VA-S6`).
  - **Acceptance criteria:** V9-G3 exits with compatibility-safe cascade planning.
  - **Suggested priority/order:** P1, Order 12.
  - **Risks/notes:** Keep scope at v0.9 voice scaling bullet only.

### Phase 4 - Cross-Platform Performance Profiling and Optimization Contracts (V9-G4)

- [ ] **[P0][Order 13] P4-T1 Define cross-platform profiling taxonomy and baseline metrics**
  - **Objective:** Specify deterministic profiling model across desktop/mobile/web-capable targets.
  - **Concrete actions:**
    - [ ] **P4-T1-ST1 Define metric catalog and platform normalization rules**
      - **Objective:** Ensure consistent measurement semantics across environments.
      - **Concrete actions:** Define CPU, memory, network, latency, jitter, startup, and background activity metric definitions with normalization assumptions.
      - **Dependencies/prerequisites:** P0-T3, v0.8 cross-platform context.
      - **Deliverables/artifacts:** Profiling taxonomy contract (`VA-P1`).
      - **Acceptance criteria:** Metric definitions are unambiguous and comparable across platforms.
      - **Suggested priority/order:** P0, Order 13.1.
      - **Risks/notes:** Inconsistent metrics invalidate optimization decisions.
    - [ ] **P4-T1-ST2 Define baseline-capture protocol and environment controls**
      - **Objective:** Ensure baseline capture is repeatable and auditable.
      - **Concrete actions:** Define environment profiles, run conditions, sample sizes, and baseline evidence requirements.
      - **Dependencies/prerequisites:** P4-T1-ST1.
      - **Deliverables/artifacts:** Baseline capture protocol (`VA-P2`).
      - **Acceptance criteria:** Baseline results are reproducible under defined controls.
      - **Suggested priority/order:** P0, Order 13.2.
      - **Risks/notes:** Weak controls introduce noisy baselines.
  - **Dependencies/prerequisites:** P0-T1 through P0-T3.
  - **Deliverables/artifacts:** Profiling baseline package (`VA-P1`, `VA-P2`).
  - **Acceptance criteria:** V9-G4 baseline criteria met.
  - **Suggested priority/order:** P0, Order 13.
  - **Risks/notes:** Scope remains profiling/optimization planning only.

- [ ] **[P0][Order 14] P4-T2 Define optimization decision framework and threshold policy**
  - **Objective:** Specify deterministic criteria for prioritizing optimization paths.
  - **Concrete actions:**
    - [ ] **P4-T2-ST1 Define bottleneck classification and prioritization rules**
      - **Objective:** Standardize optimization triage decisions.
      - **Concrete actions:** Define classification for CPU-bound, network-bound, memory-bound, I/O-bound, and synchronization-bound cases.
      - **Dependencies/prerequisites:** P4-T1.
      - **Deliverables/artifacts:** Bottleneck classification contract (`VA-P3`).
      - **Acceptance criteria:** Equivalent profiles map to consistent bottleneck classes.
      - **Suggested priority/order:** P0, Order 14.1.
      - **Risks/notes:** Misclassification yields poor optimization sequencing.
    - [ ] **P4-T2-ST2 Define optimization acceptance thresholds and regression boundaries**
      - **Objective:** Bound when optimizations are considered viable.
      - **Concrete actions:** Define threshold bands, confidence requirements, and regression rejection criteria.
      - **Dependencies/prerequisites:** P4-T2-ST1.
      - **Deliverables/artifacts:** Threshold/regression contract (`VA-P4`).
      - **Acceptance criteria:** Go/no-go decisions are deterministic and evidence-linked.
      - **Suggested priority/order:** P0, Order 14.2.
      - **Risks/notes:** Vague thresholds create arbitrary acceptance.
  - **Dependencies/prerequisites:** P4-T1.
  - **Deliverables/artifacts:** Optimization decision package (`VA-P3`, `VA-P4`).
  - **Acceptance criteria:** V9-G4 decision-framework criteria met.
  - **Suggested priority/order:** P0, Order 14.
  - **Risks/notes:** Avoid importing release-marketing performance claims.

- [ ] **[P1][Order 15] P4-T3 Define cross-platform optimization evidence and reporting format**
  - **Objective:** Ensure optimization findings are traceable and governance-auditable.
  - **Concrete actions:**
    - [ ] **P4-T3-ST1 Define platform comparison reporting schema**
      - **Objective:** Normalize reporting of profile deltas and tradeoffs.
      - **Concrete actions:** Define schema for before/after baseline deltas, confidence, caveats, and unresolved constraints.
      - **Dependencies/prerequisites:** P4-T1, P4-T2.
      - **Deliverables/artifacts:** Performance reporting schema (`VA-P5`).
      - **Acceptance criteria:** Reports are comparable across platform families.
      - **Suggested priority/order:** P1, Order 15.1.
      - **Risks/notes:** Non-standard reports obscure regressions.
    - [ ] **P4-T3-ST2 Define optimization rollback and exception documentation rules**
      - **Objective:** Bound risk when optimization decisions regress behavior.
      - **Concrete actions:** Define rollback note format, exception handling workflow, and evidence-link requirements.
      - **Dependencies/prerequisites:** P4-T3-ST1.
      - **Deliverables/artifacts:** Rollback/exception reporting contract (`VA-P6`).
      - **Acceptance criteria:** Regression exceptions are deterministic and auditable.
      - **Suggested priority/order:** P1, Order 15.2.
      - **Risks/notes:** Poor exception hygiene hides systemic issues.
  - **Dependencies/prerequisites:** P4-T1, P4-T2.
  - **Deliverables/artifacts:** Optimization evidence package (`VA-P5`, `VA-P6`).
  - **Acceptance criteria:** V9-G4 exits with complete reporting standards.
  - **Suggested priority/order:** P1, Order 15.
  - **Risks/notes:** Preserve planning-only wording in all evidence templates.

### Phase 5 - Stress-Testing Contracts (V9-G5)

- [ ] **[P0][Order 16] P5-T1 Define 1000-member server stress-test scenario contract**
  - **Objective:** Specify deterministic large-server stress scenarios and boundary conditions.
  - **Concrete actions:**
    - [ ] **P5-T1-ST1 Define membership churn, activity distribution, and load-shape profiles**
      - **Objective:** Ensure scenarios represent realistic and adversarial load envelopes.
      - **Concrete actions:** Define join/leave churn patterns, message-rate distributions, and peak burst profiles.
      - **Dependencies/prerequisites:** P2-T1, P2-T2, P4-T1.
      - **Deliverables/artifacts:** 1000-member scenario contract (`VA-T1`).
      - **Acceptance criteria:** Scenario profiles are deterministic and reproducible.
      - **Suggested priority/order:** P0, Order 16.1.
      - **Risks/notes:** Narrow profiles under-represent real load.
    - [ ] **P5-T1-ST2 Define pass/fail thresholds and saturation classification**
      - **Objective:** Standardize stress outcome assessment.
      - **Concrete actions:** Define latency/error/throughput thresholds and saturation severity classes.
      - **Dependencies/prerequisites:** P5-T1-ST1.
      - **Deliverables/artifacts:** Saturation threshold contract (`VA-T2`).
      - **Acceptance criteria:** Stress outcomes resolve to deterministic pass/fail classes.
      - **Suggested priority/order:** P0, Order 16.2.
      - **Risks/notes:** Ambiguous thresholds block gate closure.
  - **Dependencies/prerequisites:** P4-T1, P4-T2, P4-T3.
  - **Deliverables/artifacts:** Large-server stress package (`VA-T1`, `VA-T2`).
  - **Acceptance criteria:** V9-G5 1000-member criteria met.
  - **Suggested priority/order:** P0, Order 16.
  - **Risks/notes:** Keep scope tied to stated stress bullet.

- [ ] **[P0][Order 17] P5-T2 Define 50-person voice stress and latency benchmark contract**
  - **Objective:** Specify deterministic voice-load and latency benchmark scenarios.
  - **Concrete actions:**
    - [ ] **P5-T2-ST1 Define 50-participant voice scenario classes and topology assumptions**
      - **Objective:** Align stress tests with cascade/topology planning boundaries.
      - **Concrete actions:** Define speaking-ratio profiles, packet-loss/jitter classes, and topology states.
      - **Dependencies/prerequisites:** P3-T1, P3-T2, P4-T1.
      - **Deliverables/artifacts:** Voice stress scenario contract (`VA-T3`).
      - **Acceptance criteria:** Scenario classes are deterministic and topology-aware.
      - **Suggested priority/order:** P0, Order 17.1.
      - **Risks/notes:** Missing topology context reduces benchmark relevance.
    - [ ] **P5-T2-ST2 Define latency benchmark method and acceptance bands**
      - **Objective:** Standardize latency benchmark interpretation.
      - **Concrete actions:** Define latency measurement points, percentile reporting, and pass/fail bands.
      - **Dependencies/prerequisites:** P5-T2-ST1.
      - **Deliverables/artifacts:** Latency benchmark contract (`VA-T4`).
      - **Acceptance criteria:** Latency benchmarking is deterministic and cross-run comparable.
      - **Suggested priority/order:** P0, Order 17.2.
      - **Risks/notes:** Inconsistent methods invalidate trend comparisons.
  - **Dependencies/prerequisites:** P3-T1, P3-T2, P3-T3, P4-T1, P4-T2, P4-T3.
  - **Deliverables/artifacts:** Voice stress/latency package (`VA-T3`, `VA-T4`).
  - **Acceptance criteria:** V9-G5 voice/latency criteria met.
  - **Suggested priority/order:** P0, Order 17.
  - **Risks/notes:** Avoid importing extra media feature scope.

- [ ] **[P1][Order 18] P5-T3 Define stress-campaign execution governance and evidence quality controls**
  - **Objective:** Ensure stress evidence is repeatable, auditable, and compatible with gate logic.
  - **Concrete actions:**
    - [ ] **P5-T3-ST1 Define runbook and evidence capture requirements for stress campaigns**
      - **Objective:** Normalize stress campaign execution and reporting.
      - **Concrete actions:** Define runbook fields, scenario IDs, environment fingerprints, and failure-capture minimums.
      - **Dependencies/prerequisites:** P5-T1, P5-T2.
      - **Deliverables/artifacts:** Stress campaign runbook contract (`VA-T5`).
      - **Acceptance criteria:** All stress runs produce complete, comparable evidence bundles.
      - **Suggested priority/order:** P1, Order 18.1.
      - **Risks/notes:** Incomplete runbook data breaks comparability.
    - [ ] **P5-T3-ST2 Define anomaly triage taxonomy and rerun criteria**
      - **Objective:** Standardize handling of anomalous stress outcomes.
      - **Concrete actions:** Define anomaly classes, mandatory rerun conditions, and closure criteria.
      - **Dependencies/prerequisites:** P5-T3-ST1.
      - **Deliverables/artifacts:** Anomaly triage contract (`VA-T6`).
      - **Acceptance criteria:** Anomaly handling is deterministic and traceable.
      - **Suggested priority/order:** P1, Order 18.2.
      - **Risks/notes:** Weak triage can mask systemic bottlenecks.
  - **Dependencies/prerequisites:** P5-T1, P5-T2.
  - **Deliverables/artifacts:** Stress governance package (`VA-T5`, `VA-T6`).
  - **Acceptance criteria:** V9-G5 exits with campaign governance controls complete.
  - **Suggested priority/order:** P1, Order 18.
  - **Risks/notes:** Keep all wording planning-only.

### Phase 6 - Relay Performance and Load-Testing Contracts (V9-G6)

- [ ] **[P0][Order 19] P6-T1 Define relay capacity model and saturation boundaries**
  - **Objective:** Specify deterministic relay capacity envelopes and saturation behavior.
  - **Concrete actions:**
    - [ ] **P6-T1-ST1 Define relay workload classes and resource budget model**
      - **Objective:** Standardize relay workload characterization.
      - **Concrete actions:** Define workload classes for DHT, relay circuits, store-forward, and SFU-forwarding interactions.
      - **Dependencies/prerequisites:** v0.1 relay baseline, v0.7 store-forward context, P4-T1.
      - **Deliverables/artifacts:** Relay workload model (`VA-R1`).
      - **Acceptance criteria:** Workload classes map deterministically to resource budget dimensions.
      - **Suggested priority/order:** P0, Order 19.1.
      - **Risks/notes:** Poor modeling leads to unstable capacity assumptions.
    - [ ] **P6-T1-ST2 Define saturation behavior and service-priority policies**
      - **Objective:** Bound degraded behavior under relay overload.
      - **Concrete actions:** Define priority semantics, shedding policies, and degradation classes.
      - **Dependencies/prerequisites:** P6-T1-ST1.
      - **Deliverables/artifacts:** Saturation policy contract (`VA-R2`).
      - **Acceptance criteria:** Overload behavior is deterministic and policy-consistent.
      - **Suggested priority/order:** P0, Order 19.2.
      - **Risks/notes:** Undefined overload policy can cause cascading failures.
  - **Dependencies/prerequisites:** P4-T1, P4-T2, P4-T3, v0.1 relay baseline.
  - **Deliverables/artifacts:** Relay capacity package (`VA-R1`, `VA-R2`).
  - **Acceptance criteria:** V9-G6 capacity criteria met.
  - **Suggested priority/order:** P0, Order 19.
  - **Risks/notes:** Maintain no-special-node invariant.

- [ ] **[P0][Order 20] P6-T2 Define relay load-testing scenarios and acceptance thresholds**
  - **Objective:** Specify deterministic relay load-test campaigns aligned with v0.9 scope.
  - **Concrete actions:**
    - [ ] **P6-T2-ST1 Define relay load profiles and fault-injection classes**
      - **Objective:** Cover nominal, peak, and faulted relay conditions.
      - **Concrete actions:** Define connection churn, traffic distributions, burst classes, and fault injection types.
      - **Dependencies/prerequisites:** P6-T1, P5-T1.
      - **Deliverables/artifacts:** Relay load scenario contract (`VA-R3`).
      - **Acceptance criteria:** Load profiles are complete and reproducible.
      - **Suggested priority/order:** P0, Order 20.1.
      - **Risks/notes:** Narrow profile coverage misses critical behavior.
    - [ ] **P6-T2-ST2 Define relay load pass/fail and degradation-recovery thresholds**
      - **Objective:** Standardize relay campaign outcome decisions.
      - **Concrete actions:** Define thresholds for latency, drop rate, queue pressure, and recovery windows.
      - **Dependencies/prerequisites:** P6-T2-ST1.
      - **Deliverables/artifacts:** Relay threshold contract (`VA-R4`).
      - **Acceptance criteria:** Campaign outcomes classify deterministically.
      - **Suggested priority/order:** P0, Order 20.2.
      - **Risks/notes:** Inconsistent thresholds undermine gate confidence.
  - **Dependencies/prerequisites:** P6-T1.
  - **Deliverables/artifacts:** Relay load-testing package (`VA-R3`, `VA-R4`).
  - **Acceptance criteria:** V9-G6 load-testing criteria met.
  - **Suggested priority/order:** P0, Order 20.
  - **Risks/notes:** Scope remains relay performance/load only.

- [ ] **[P1][Order 21] P6-T3 Define relay optimization-change governance and rollback controls**
  - **Objective:** Ensure relay optimization proposals remain governance-compliant and reversible.
  - **Concrete actions:**
    - [ ] **P6-T3-ST1 Define optimization-change classification and approval workflow**
      - **Objective:** Separate minor-safe adjustments from major-path behavior changes.
      - **Concrete actions:** Define classification criteria, approval checkpoints, and evidence requirements.
      - **Dependencies/prerequisites:** P6-T1, P6-T2, P0-T2.
      - **Deliverables/artifacts:** Relay optimization governance workflow (`VA-R5`).
      - **Acceptance criteria:** Every change proposal maps to a deterministic governance path.
      - **Suggested priority/order:** P1, Order 21.1.
      - **Risks/notes:** Ambiguous classification risks governance bypass.
    - [ ] **P6-T3-ST2 Define rollback protocol and post-incident evidence expectations**
      - **Objective:** Bound restoration behavior when optimization causes regressions.
      - **Concrete actions:** Define rollback sequencing, evidence minimums, and closure requirements.
      - **Dependencies/prerequisites:** P6-T3-ST1.
      - **Deliverables/artifacts:** Relay rollback contract (`VA-R6`).
      - **Acceptance criteria:** Rollback actions are deterministic and auditable.
      - **Suggested priority/order:** P1, Order 21.2.
      - **Risks/notes:** Missing rollback rigor increases outage risk.
  - **Dependencies/prerequisites:** P6-T1, P6-T2.
  - **Deliverables/artifacts:** Relay governance package (`VA-R5`, `VA-R6`).
  - **Acceptance criteria:** V9-G6 exits with relay governance controls complete.
  - **Suggested priority/order:** P1, Order 21.
  - **Risks/notes:** Preserve single-binary mode invariant throughout.

### Phase 7 - Mobile Battery Optimization Contracts (V9-G7)

- [ ] **[P0][Order 22] P7-T1 Define mobile background-activity budget and wakeup policy contract**
  - **Objective:** Specify deterministic background activity reduction boundaries for mobile.
  - **Concrete actions:**
    - [ ] **P7-T1-ST1 Define background task classes and energy budget envelopes**
      - **Objective:** Bound background workloads to explicit budgets.
      - **Concrete actions:** Define task categories, periodicity assumptions, and budget envelopes for networking/sync/activity updates.
      - **Dependencies/prerequisites:** P4-T1, v0.7 notification context, v0.8 platform context.
      - **Deliverables/artifacts:** Background budget contract (`VA-B1`).
      - **Acceptance criteria:** Background workloads map deterministically to budget classes.
      - **Suggested priority/order:** P0, Order 22.1.
      - **Risks/notes:** Unbounded tasks can degrade battery unpredictably.
    - [ ] **P7-T1-ST2 Define wakeup triggers, suppression rules, and fallback behavior**
      - **Objective:** Control wake frequency while preserving protocol continuity assumptions.
      - **Concrete actions:** Define wake trigger taxonomy, suppression precedence, and safe fallback when wake budget is exceeded.
      - **Dependencies/prerequisites:** P7-T1-ST1.
      - **Deliverables/artifacts:** Wakeup/suppression contract (`VA-B2`).
      - **Acceptance criteria:** Wake behavior is deterministic and budget-aligned.
      - **Suggested priority/order:** P0, Order 22.2.
      - **Risks/notes:** Over-suppression can harm timeliness.
  - **Dependencies/prerequisites:** P4-T1, P4-T2, P4-T3.
  - **Deliverables/artifacts:** Background activity package (`VA-B1`, `VA-B2`).
  - **Acceptance criteria:** V9-G7 background-activity criteria met.
  - **Suggested priority/order:** P0, Order 22.
  - **Risks/notes:** Keep scope to battery optimization only.

- [ ] **[P0][Order 23] P7-T2 Define battery-impact profiling and optimization validation contract**
  - **Objective:** Specify deterministic battery-impact measurement and acceptance logic.
  - **Concrete actions:**
    - [ ] **P7-T2-ST1 Define battery-impact measurement scenarios and controls**
      - **Objective:** Ensure energy measurements are comparable and reproducible.
      - **Concrete actions:** Define idle/background/active scenario classes, device-state controls, and capture intervals.
      - **Dependencies/prerequisites:** P7-T1, P4-T1.
      - **Deliverables/artifacts:** Battery scenario contract (`VA-B3`).
      - **Acceptance criteria:** Scenario design supports deterministic battery-impact comparison.
      - **Suggested priority/order:** P0, Order 23.1.
      - **Risks/notes:** Noisy measurement controls obscure improvements.
    - [ ] **P7-T2-ST2 Define battery optimization acceptance thresholds and regression guards**
      - **Objective:** Bound go/no-go decisions for battery-oriented optimization proposals.
      - **Concrete actions:** Define threshold bands, confidence minima, and regression blocker criteria.
      - **Dependencies/prerequisites:** P7-T2-ST1.
      - **Deliverables/artifacts:** Battery threshold contract (`VA-B4`).
      - **Acceptance criteria:** Acceptance decisions are deterministic and auditable.
      - **Suggested priority/order:** P0, Order 23.2.
      - **Risks/notes:** Weak regression guards can worsen battery behavior.
  - **Dependencies/prerequisites:** P7-T1.
  - **Deliverables/artifacts:** Battery validation package (`VA-B3`, `VA-B4`).
  - **Acceptance criteria:** V9-G7 battery-validation criteria met.
  - **Suggested priority/order:** P0, Order 23.
  - **Risks/notes:** Maintain planning-only framing, no completion claims.

- [ ] **[P1][Order 24] P7-T3 Define battery/performance tradeoff governance and user-impact boundaries**
  - **Objective:** Ensure battery reductions do not silently violate core performance assumptions.
  - **Concrete actions:**
    - [ ] **P7-T3-ST1 Define tradeoff matrix for battery vs latency/responsiveness**
      - **Objective:** Make compromise boundaries explicit and reviewable.
      - **Concrete actions:** Define tradeoff matrix dimensions and acceptable compromise ranges.
      - **Dependencies/prerequisites:** P7-T1, P7-T2, P4-T2.
      - **Deliverables/artifacts:** Tradeoff matrix contract (`VA-B5`).
      - **Acceptance criteria:** Tradeoff decisions are deterministic and policy-bounded.
      - **Suggested priority/order:** P1, Order 24.1.
      - **Risks/notes:** Hidden tradeoffs can degrade real-time experience.
    - [ ] **P7-T3-ST2 Define escalation triggers for behavior-breaking battery policies**
      - **Objective:** Route breaking behavior through major governance path.
      - **Concrete actions:** Document trigger examples requiring new multistream IDs, downgrade negotiation, AEP, and multi-implementation validation.
      - **Dependencies/prerequisites:** P7-T3-ST1, P0-T2.
      - **Deliverables/artifacts:** Battery major-path trigger annex (`VA-B6`).
      - **Acceptance criteria:** Trigger examples are explicit and gate-ready.
      - **Suggested priority/order:** P1, Order 24.2.
      - **Risks/notes:** Missing triggers risks incompatible policy changes.
  - **Dependencies/prerequisites:** P7-T1, P7-T2.
  - **Deliverables/artifacts:** Battery governance package (`VA-B5`, `VA-B6`).
  - **Acceptance criteria:** V9-G7 exits with battery tradeoff governance complete.
  - **Suggested priority/order:** P1, Order 24.
  - **Risks/notes:** Preserve protocol-first and compatibility posture.

### Phase 8 - Integrated Validation and Governance Readiness (V9-G8)

- [ ] **[P0][Order 25] P8-T1 Build integrated cross-domain validation matrix**
  - **Objective:** Validate interactions across all seven v0.9 scope bullets.
  - **Concrete actions:**
    - [ ] **P8-T1-ST1 Define end-to-end scenario matrix (positive/adverse/saturation/recovery)**
      - **Objective:** Ensure cross-domain behavior is validated, not only isolated contracts.
      - **Concrete actions:** Build scenario set spanning IPFS persistence, large-server pubsub, cascading SFU, profiling, stress, relay load, and battery controls.
      - **Dependencies/prerequisites:** P1-T1 through P7-T3.
      - **Deliverables/artifacts:** Integrated scenario matrix (`VA-X1`).
      - **Acceptance criteria:** Every scope bullet appears in integrated coverage.
      - **Suggested priority/order:** P0, Order 25.1.
      - **Risks/notes:** Missing interactions conceal systemic risks.
    - [ ] **P8-T1-ST2 Define integrated pass/fail thresholds and evidence-link rules**
      - **Objective:** Standardize gate-readiness judgments.
      - **Concrete actions:** Define threshold taxonomy, unresolved-item policy, and evidence-link requirements.
      - **Dependencies/prerequisites:** P8-T1-ST1.
      - **Deliverables/artifacts:** Integrated validation criteria (`VA-X2`).
      - **Acceptance criteria:** Gate outcomes are deterministic and auditable.
      - **Suggested priority/order:** P0, Order 25.2.
      - **Risks/notes:** Ambiguous thresholds create inconsistent closure.
  - **Dependencies/prerequisites:** All domain phases complete at contract level.
  - **Deliverables/artifacts:** Integrated validation package (`VA-X1`, `VA-X2`).
  - **Acceptance criteria:** V9-G8 integrated-validation criteria met.
  - **Suggested priority/order:** P0, Order 25.
  - **Risks/notes:** Integration phase is critical path to handoff.

- [ ] **[P0][Order 26] P8-T2 Perform compatibility/governance/invariant conformance audit**
  - **Objective:** Confirm additive evolution, major-path governance, and architecture invariants.
  - **Concrete actions:**
    - [ ] **P8-T2-ST1 Run additive-only conformance audit across schema/capability deltas**
      - **Objective:** Ensure minor-path deltas remain compatibility-safe.
      - **Concrete actions:** Audit all relevant artifacts against `VA-G3`; record pass/fail and exceptions.
      - **Dependencies/prerequisites:** P0-T2, P1-T1 through P7-T3.
      - **Deliverables/artifacts:** Additive conformance report (`VA-X3`).
      - **Acceptance criteria:** All minor-path deltas are compliant or escalated.
      - **Suggested priority/order:** P0, Order 26.1.
      - **Risks/notes:** Non-compliance silently breaks interoperability.
    - [ ] **P8-T2-ST2 Run major-path trigger and invariant conformance audit**
      - **Objective:** Ensure breaking candidates include full governance evidence.
      - **Concrete actions:** Validate new multistream ID evidence, downgrade negotiation, AEP path, multi-implementation validation, and single-binary invariant alignment.
      - **Dependencies/prerequisites:** P8-T2-ST1.
      - **Deliverables/artifacts:** Governance/invariant audit report (`VA-X4`).
      - **Acceptance criteria:** Every breaking candidate has explicit governance status.
      - **Suggested priority/order:** P0, Order 26.2.
      - **Risks/notes:** Incomplete governance blocks V9-G9.
  - **Dependencies/prerequisites:** P8-T1.
  - **Deliverables/artifacts:** Compatibility/governance package (`VA-X3`, `VA-X4`).
  - **Acceptance criteria:** V9-G8 governance criteria met.
  - **Suggested priority/order:** P0, Order 26.
  - **Risks/notes:** Governance integrity remains non-negotiable.

- [ ] **[P1][Order 27] P8-T3 Perform open-decision and licensing/repository-language conformance review**
  - **Objective:** Preserve unresolved decisions and document-language integrity.
  - **Concrete actions:**
    - [ ] **P8-T3-ST1 Validate open-decision handling discipline**
      - **Objective:** Ensure unresolved questions are not presented as settled facts.
      - **Concrete actions:** Audit open-decision references for status, owner role, revisit gate, and handling-rule consistency.
      - **Dependencies/prerequisites:** P1-T1 through P8-T2.
      - **Deliverables/artifacts:** Open-decision conformance report (`VA-X5`).
      - **Acceptance criteria:** All unresolved decisions remain explicitly open.
      - **Suggested priority/order:** P1, Order 27.1.
      - **Risks/notes:** Wording drift can create false certainty.
    - [ ] **P8-T3-ST2 Validate licensing and repository-state language alignment**
      - **Objective:** Preserve MIT-like/CC-BY-SA alignment and documentation-only framing.
      - **Concrete actions:** Audit artifacts for licensing and planning-only language consistency.
      - **Dependencies/prerequisites:** P8-T3-ST1.
      - **Deliverables/artifacts:** Licensing/repository-language conformance note (`VA-X6`).
      - **Acceptance criteria:** No section contradicts `AGENTS.md` constraints.
      - **Suggested priority/order:** P1, Order 27.2.
      - **Risks/notes:** Inconsistent legal framing undermines governance trust.
  - **Dependencies/prerequisites:** P8-T1, P8-T2.
  - **Deliverables/artifacts:** Open-decision/language package (`VA-X5`, `VA-X6`).
  - **Acceptance criteria:** V9-G8 exits with governance-readiness integrity confirmed.
  - **Suggested priority/order:** P1, Order 27.
  - **Risks/notes:** Keep unresolved policy questions visible.

### Phase 9 - Release Conformance and Execution Handoff (V9-G9)

- [ ] **[P0][Order 28] P9-T1 Close scope-to-task-to-artifact traceability**
  - **Objective:** Achieve complete auditable traceability for all v0.9 scope bullets.
  - **Concrete actions:**
    - [ ] **P9-T1-ST1 Build final traceability closure matrix with acceptance anchors**
      - **Objective:** Link every v0.9 bullet to tasks, artifacts, and gate acceptance anchors.
      - **Concrete actions:** Compile mapping table, verify evidence-link completeness, and mark unresolved gaps.
      - **Dependencies/prerequisites:** P8-T1 through P8-T3.
      - **Deliverables/artifacts:** Final traceability matrix (`VA-H1`).
      - **Acceptance criteria:** All 7 bullets have complete mapping and acceptance anchors.
      - **Suggested priority/order:** P0, Order 28.1.
      - **Risks/notes:** Missing traceability blocks handoff.
    - [ ] **P9-T1-ST2 Execute anti-scope-creep closure audit**
      - **Objective:** Confirm no v1.0+ or post-v1 capability entered v0.9 tasks.
      - **Concrete actions:** Run exclusion checklist and capture findings.
      - **Dependencies/prerequisites:** P9-T1-ST1.
      - **Deliverables/artifacts:** Scope-boundary closure audit (`VA-H2`).
      - **Acceptance criteria:** No unauthorized capability appears in v0.9 artifact set.
      - **Suggested priority/order:** P0, Order 28.2.
      - **Risks/notes:** Hidden expansion invalidates plan quality.
  - **Dependencies/prerequisites:** P8-T1, P8-T2, P8-T3.
  - **Deliverables/artifacts:** Traceability closure package (`VA-H1`, `VA-H2`).
  - **Acceptance criteria:** V9-G9 traceability criteria met.
  - **Suggested priority/order:** P0, Order 28.
  - **Risks/notes:** Primary release-conformance blocker set.

- [ ] **[P1][Order 29] P9-T2 Build release-conformance checklist with evidence linkage**
  - **Objective:** Provide deterministic go/no-go planning handoff checklist.
  - **Concrete actions:**
    - [ ] **P9-T2-ST1 Assemble gate-aligned conformance checklist sections**
      - **Objective:** Cover scope, sequencing, compatibility, governance, validation, and language integrity.
      - **Concrete actions:** Build checklist sections, include gate references and pass/fail fields.
      - **Dependencies/prerequisites:** P9-T1.
      - **Deliverables/artifacts:** Release-conformance checklist (`VA-H3`).
      - **Acceptance criteria:** Checklist supports deterministic handoff decision.
      - **Suggested priority/order:** P1, Order 29.1.
      - **Risks/notes:** Incomplete checklist creates ambiguous closure.
    - [ ] **P9-T2-ST2 Map every checklist item to owner role and evidence ID**
      - **Objective:** Ensure accountability and auditability.
      - **Concrete actions:** Add owner role, artifact ID, and gate linkage for each row.
      - **Dependencies/prerequisites:** P9-T2-ST1.
      - **Deliverables/artifacts:** Evidence-linked conformance registry (`VA-H4`).
      - **Acceptance criteria:** All checklist items have explicit owner/evidence mapping.
      - **Suggested priority/order:** P1, Order 29.2.
      - **Risks/notes:** Unowned items delay closure.
  - **Dependencies/prerequisites:** P9-T1.
  - **Deliverables/artifacts:** Conformance checklist package (`VA-H3`, `VA-H4`).
  - **Acceptance criteria:** V9-G9 conformance checklist criteria met.
  - **Suggested priority/order:** P1, Order 29.
  - **Risks/notes:** Keep planning-only language explicit.

- [ ] **[P1][Order 30] P9-T3 Prepare execution handoff dossier and forward deferral register**
  - **Objective:** Finalize execution-ready planning handoff with explicit future deferrals.
  - **Concrete actions:**
    - [ ] **P9-T3-ST1 Compile execution handoff dossier with gate outcomes and residual risks**
      - **Objective:** Provide complete orchestration input without implementation completion claims.
      - **Concrete actions:** Aggregate gate outcomes, evidence index, unresolved decisions, and residual-risk summary.
      - **Dependencies/prerequisites:** P9-T2.
      - **Deliverables/artifacts:** Execution handoff dossier (`VA-H5`).
      - **Acceptance criteria:** Dossier is complete, internally consistent, and planning-only.
      - **Suggested priority/order:** P1, Order 30.1.
      - **Risks/notes:** Missing context creates execution ambiguity.
    - [ ] **P9-T3-ST2 Build v1.0+/post-v1 deferral register from v0.9 residuals**
      - **Objective:** Preserve roadmap boundary clarity and prevent hidden carry-over.
      - **Concrete actions:** Record deferred items, rationale, target roadmap band, and owner role.
      - **Dependencies/prerequisites:** P9-T3-ST1.
      - **Deliverables/artifacts:** Forward deferral register (`VA-H6`).
      - **Acceptance criteria:** Every deferral maps explicitly to v1.0+ or post-v1 bands.
      - **Suggested priority/order:** P1, Order 30.2.
      - **Risks/notes:** Untracked residuals become hidden scope debt.
  - **Dependencies/prerequisites:** P9-T1, P9-T2.
  - **Deliverables/artifacts:** Handoff/deferral package (`VA-H5`, `VA-H6`).
  - **Acceptance criteria:** V9-G9 exit criteria satisfied with full handoff evidence.
  - **Suggested priority/order:** P1, Order 30.
  - **Risks/notes:** Preserve planned-vs-implemented distinction.

---

## H. Suggested Execution Waves and Sequencing

### Wave A - Scope/governance/evidence foundation (V9-G0)
1. P0-T1
2. P0-T2
3. P0-T3

### Wave B - IPFS persistent hosting contracts (V9-G1)
4. P1-T1
5. P1-T2
6. P1-T3

### Wave C - Large-server optimization contracts (V9-G2)
7. P2-T1
8. P2-T2
9. P2-T3

### Wave D - Cascading SFU mesh contracts (V9-G3)
10. P3-T1
11. P3-T2
12. P3-T3

### Wave E - Cross-platform profiling/optimization contracts (V9-G4)
13. P4-T1
14. P4-T2
15. P4-T3

### Wave F - Stress-testing contracts (V9-G5)
16. P5-T1
17. P5-T2
18. P5-T3

### Wave G - Relay performance/load contracts (V9-G6)
19. P6-T1
20. P6-T2
21. P6-T3

### Wave H - Mobile battery optimization contracts (V9-G7)
22. P7-T1
23. P7-T2
24. P7-T3

### Wave I - Integrated validation/governance readiness (V9-G8)
25. P8-T1
26. P8-T2
27. P8-T3

### Wave J - Release conformance/handoff (V9-G9)
28. P9-T1
29. P9-T2
30. P9-T3

---

## I. Verification Evidence Model and Traceability Expectations

### I.1 Evidence model rules
1. Every task produces at least one named artifact with artifact ID.
2. Every scope item appears in at least one positive-path and one adverse/saturation/recovery scenario.
3. Every gate submission includes explicit pass/fail decision and evidence links.
4. Every compatibility-sensitive delta includes additive and major-path checklist evidence.
5. Every unresolved decision remains explicitly open and linked to revisit gate.

### I.2 Traceability mapping: v0.9 scope to tasks/artifacts/acceptance anchors

| Scope Item ID | v0.9 Scope Bullet | Primary Tasks | Validation Artifacts | Acceptance Anchor |
|---|---|---|---|---|
| S9-01 | IPFS integration for persistent file hosting (pinning by server owners) | P1-T1, P1-T2, P1-T3 | VA-I1, VA-I2, VA-I3, VA-I4, VA-I5, VA-I6, VA-X1 | P1-T1/P1-T2 acceptance + P8-T1 integrated coverage |
| S9-02 | Large server optimization: hierarchical GossipSub, lazy member loading | P2-T1, P2-T2, P2-T3 | VA-L1, VA-L2, VA-L3, VA-L4, VA-L5, VA-L6, VA-X1 | P2-T1/P2-T2 acceptance + P8-T1 integrated coverage |
| S9-03 | Cascading SFU mesh for 200+ participant voice | P3-T1, P3-T2, P3-T3 | VA-S1, VA-S2, VA-S3, VA-S4, VA-S5, VA-S6, VA-X1 | P3-T1/P3-T2 acceptance + P8-T1 integrated coverage |
| S9-04 | Performance profiling and optimization across all platforms | P4-T1, P4-T2, P4-T3 | VA-P1, VA-P2, VA-P3, VA-P4, VA-P5, VA-P6, VA-X1 | P4-T1/P4-T2 acceptance + P8-T1 integrated coverage |
| S9-05 | Stress testing: 1000-member server, 50-person voice, latency benchmarks | P5-T1, P5-T2, P5-T3 | VA-T1, VA-T2, VA-T3, VA-T4, VA-T5, VA-T6, VA-X1 | P5-T1/P5-T2 acceptance + P8-T1 integrated coverage |
| S9-06 | Relay node performance optimization and load testing | P6-T1, P6-T2, P6-T3 | VA-R1, VA-R2, VA-R3, VA-R4, VA-R5, VA-R6, VA-X1 | P6-T1/P6-T2 acceptance + P8-T1 integrated coverage |
| S9-07 | Battery optimization on mobile (background activity reduction) | P7-T1, P7-T2, P7-T3 | VA-B1, VA-B2, VA-B3, VA-B4, VA-B5, VA-B6, VA-X1 | P7-T1/P7-T2 acceptance + P8-T1 integrated coverage |

### I.3 Traceability closure rules
- Any scope item without task mapping blocks V9-G9.
- Any scope item without artifact mapping blocks V9-G9.
- Any scope item without acceptance anchor blocks V9-G9.
- Any gate checklist item without evidence link is incomplete.

---

## J. Risk Register (Planning-Level)

| Risk ID | Description | Severity | Affected Scope | Mitigation in Plan | Owner Role |
|---|---|---|---|---|---|
| R9-01 | Scope creep from v0.9 into v1.0+ publication/release tracks | High | S9-01 to S9-07 | Exclusion policy + anti-scope audit at V9-G0 and V9-G9 | V9-G0 owner |
| R9-02 | IPFS pinning responsibility ambiguity | High | S9-01 | VA-I2 role/lifecycle contract + governance checks | V9-G1 owner |
| R9-03 | Hierarchical GossipSub fanout instability under scale | High | S9-02 | VA-L1/VA-L2 topology and backpressure contracts | V9-G2 owner |
| R9-04 | Lazy member loading produces stale or inconsistent member views | High | S9-02 | VA-L3/VA-L4 trigger and coherence contracts | V9-G2 owner |
| R9-05 | Cascading SFU failover complexity destabilizes large voice rooms | High | S9-03 | VA-S3/VA-S4 forwarding/failover semantics | V9-G3 owner |
| R9-06 | Cross-platform profiling metrics are not comparable | High | S9-04 | VA-P1 normalization + VA-P2 baseline controls | V9-G4 owner |
| R9-07 | Stress thresholds are ambiguous and non-actionable | High | S9-05 | VA-T2/VA-T4 threshold contracts + VA-T6 triage | V9-G5 owner |
| R9-08 | Relay overload behavior is undefined under saturation | High | S9-06 | VA-R2 saturation policy + VA-R4 thresholds | V9-G6 owner |
| R9-09 | Battery reductions degrade responsiveness or delivery posture | High | S9-07 | VA-B5 tradeoff matrix + VA-B6 escalation triggers | V9-G7 owner |
| R9-10 | Breaking behavior introduced without governance pathway | High | All | VA-G4 + VA-X4 mandatory audits | V9-G8 owner |
| R9-11 | Single-binary mode invariant erodes through wording drift | Medium | All | Invariant conformance checks in V9-G8 | V9-G8 owner |
| R9-12 | Open decisions are presented as settled architecture | Medium | All | VA-X5 open-decision conformance audit | V9-G8 owner |

---

## K. Open Decisions Tracking

| Open Decision ID | Open Question | Status | Owner Role | Revisit Gate | Trigger for Revisit | Handling Rule |
|---|---|---|---|---|---|---|
| OD9-01 | Preferred default retention expectation for server-owner pinning when storage availability is heterogeneous. | Open | Persistent Hosting Contract Lead | V9-G1 | Validation evidence reveals unresolved retention tradeoffs not resolved in source docs. | Keep defaults bounded and explicit; do not present one as final unless source docs resolve it. |
| OD9-02 | Preferred hierarchy depth profile for large-server GossipSub under mixed network quality conditions. | Open | Large-Server Topology Lead | V9-G2 | Scale validation shows unresolved hierarchy tradeoffs not fixed by source docs. | Keep depth profile as open decision with explicit owner and revisit gate. |
| OD9-03 | Preferred cascade split/merge aggressiveness for 200+ participant voice stability. | Open | Realtime Topology Lead | V9-G3 | Stress scenarios expose unresolved transition tradeoffs absent source-level resolution. | Keep aggressiveness open with deterministic safety boundaries. |
| OD9-04 | Cross-platform baseline confidence threshold for optimization go/no-go decisions. | Open | Performance Governance Lead | V9-G4 | Profiling variance indicates unresolved confidence policy tradeoffs. | Keep threshold policy open until evidence narrows acceptable range. |
| OD9-05 | Relay overload priority policy when circuit relay, store-forward, and SFU forwarding compete. | Open | Relay Operations Contract Lead | V9-G6 | Load testing identifies unresolved policy tradeoffs not decided by source docs. | Keep policy options explicit; avoid declaring final priority order prematurely. |
| OD9-06 | Mobile battery policy aggressiveness under constrained power modes vs responsiveness expectations. | Open | Mobile Runtime Contract Lead | V9-G7 | Battery/performance tradeoff evidence reveals unresolved boundaries. | Keep aggressiveness open and bounded; do not present as settled final policy. |
| OD9-07 | Preferred mobile wake-policy topology to minimize centralization risk while preserving delivery reliability. | Open | Mobile Runtime Governance Lead | V9-G7 | Battery and delivery validation reveals unresolved decentralization vs reliability tradeoffs. | Keep wake-policy topology options explicit and unresolved; do not present centralized defaults as settled architecture. |

Handling rule:
- Open decisions remain in `Open` status, include owner role and revisit gate, and are never represented as settled architecture in v0.9 artifacts.

---

## L. Release-Conformance Checklist for Execution Handoff (V9-G9)

v0.9 planning is execution-ready only when all items below are satisfied.

### L.1 Scope and boundary integrity
- [ ] All 7 v0.9 roadmap bullets are mapped to tasks, artifacts, and acceptance anchors.
- [ ] Out-of-scope boundaries are documented and referenced by gate checklists.
- [ ] No v1.0+ or post-v1 capabilities are imported into v0.9 tasks.

### L.2 Dependency and sequencing integrity
- [ ] v0.1 through v0.8 prerequisite assumptions are linked to dependent tasks.
- [ ] Task ordering is dependency-coherent across all phases.
- [ ] Gate exit criteria are deterministic and evidence-backed.

### L.3 Compatibility and governance integrity
- [ ] Additive-only protobuf discipline is applied to all schema-touching artifacts.
- [ ] Any breaking candidate includes major-path governance evidence.
- [ ] New multistream IDs and downgrade negotiation requirements are preserved where applicable.
- [ ] AEP flow and multi-implementation validation are referenced for breaking pathways.
- [ ] Single-binary and protocol-first invariants are preserved across all artifacts.

### L.4 Validation and traceability integrity
- [ ] Evidence model rules are enforced for all tasks and gates.
- [ ] Cross-domain positive/adverse/saturation/recovery scenarios cover all seven scope bullets.
- [ ] Traceability closure rules are satisfied with explicit evidence links.

### L.5 Documentation quality and handoff completeness
- [ ] Planned-vs-implemented separation is explicit in all sections.
- [ ] Open decisions remain unresolved and tracked with revisit gates.
- [ ] Licensing language remains aligned: code permissive MIT-like and protocol specification CC-BY-SA.
- [ ] Release-conformance checklist includes pass/fail status and evidence links per scope item.
- [ ] Execution handoff dossier and v1.0+/post-v1 deferral register are complete and roadmap-aligned.

---

## M. Definition of Done for v0.9 Planning Artifact

This planning artifact is complete when:
1. It captures all mandatory v0.9 scope bullets and excludes unauthorized scope expansion.
2. It provides gate/phase/task/sub-task detail with objective, concrete actions, dependencies, deliverables, acceptance criteria, order, priority, and risks.
3. It preserves protocol-first and single-binary architecture invariants.
4. It embeds compatibility/governance constraints for additive evolution and breaking-change pathways (new multistream IDs, downgrade negotiation, AEP, and multi-implementation validation).
5. It includes deterministic verification evidence model and traceability closure rules.
6. It includes integrated validation/governance-readiness and final release-conformance/handoff phases.
7. It remains planning-only and does not claim implementation completion.
