# TODO_v09.md

> Status: Execution plan (implementation required). Items are marked complete only when working code + validation evidence exist.
>
> Expected evidence anchors (update during execution): `docs/v0.9/`, `pkg/v09/`, `cmd/aether/`, `tests/perf/v09/`, `tests/e2e/v09/`.
>
> Authoritative v0.9 scope source: `aether-v3.md` roadmap bullets under **v0.9.0 — Forge**.
>
> Inputs used for sequencing, dependency posture, closure patterns, and guardrail carry-forward: `aether-v3.md`, `ENCRYPTION_PLUS.md`, `open_decisions.md`, `open_decisions_proposals.md`, `TODO_v01.md`, `TODO_v02.md`, `TODO_v03.md`, `TODO_v04.md`, `TODO_v05.md`, `TODO_v06.md`, `TODO_v07.md`, `TODO_v08.md`, and `AGENTS.md`.
>
> Guardrails that are mandatory throughout this plan:
> - Maintain strict planned-vs-implemented separation: specs remain explicit, and completion requires runnable implementations + tests + evidence.
> - Protocol-first priority is non-negotiable: protocol/spec contract is the product; client/UI behavior is downstream.
> - Network model invariant remains unchanged: single binary with mode flags `--mode=client|relay|bootstrap`; no privileged node classes.
> - Compatibility invariant remains mandatory: protobuf minor evolution is additive-only.
> - Breaking protocol behavior requires major-path governance evidence: new multistream IDs, downgrade negotiation, AEP flow, and validation by at least two independent implementations.
> - Open decisions remain unresolved unless source documentation explicitly resolves them.
> - Licensing alignment remains explicit: code licensing AGPL and protocol specification licensing CC-BY-SA.
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

### A.1 Execution-plan status
This document is a v0.9 execution plan for scope control, sequencing, validation design, and execution handoff readiness. It does not assert that implementation work has been completed.

### A.2 Source-of-truth hierarchy
1. `aether-v3.md` roadmap bullets for **v0.9.0 — Forge** are the sole authority for in-scope capability.
2. `AGENTS.md` provides mandatory repository/governance guardrails and framing constraints.
3. `TODO_v01.md` through `TODO_v08.md` provide continuity patterns for gate design, evidence discipline, and anti-scope-creep posture.

### A.3 Inherited framing constraints
- Protocol-first posture remains primary: protocol contract and interoperability are authoritative.
- Single-binary deployment model remains unchanged (`--mode=client|relay|bootstrap`).
- Compatibility/governance constraints apply to all protocol-touching deltas.
- Open decisions remain open unless explicitly resolved by source docs.
- RM-01 naming baseline: the v0.9 client/experience continues under the working name **Harmolyn** (traceable to the `Aether` bullet in `aether-v3.md`), and the protocol/backend stack proceeds under **xorein**, ensuring downstream documentation records both names without losing historical traceability.

---

## B. v0.9 Objective and Measurable Success Outcomes

### B.1 Objective
Deliver **v0.9 Forge** as a protocol-first implementation increment that ships performance and scale capabilities on top of v0.1-v0.8 baselines by delivering:
- IPFS integration for persistent file hosting with server-owner pinning posture.
- Large-server operation model including hierarchical GossipSub and lazy member loading.
- Cascading SFU mesh contract for 200+ participant voice scenarios.
- Cross-platform performance profiling and optimization planning envelope.
- Stress-testing contract for baseline 1000-member server, 50-person voice, and latency benchmarks, followed by +50 participant increments per encryption/security mode until hard limits are characterized.
- Relay-node performance optimization and load-testing contract.
- Mobile battery optimization strategy via background-activity reduction constraints.
- Scale-driven SecurityMode transition triggers and sharding guidance for huge interactive channels.

### B.2 Measurable success outcomes
1. Persistent-hosting contract defines deterministic IPFS pinning responsibilities, retention boundaries, and availability semantics for server-owner-managed pinning.
2. Hierarchical GossipSub and lazy-member-loading contracts define deterministic fanout and member-state loading behavior suitable for large-server planning targets.
3. Cascading SFU mesh contract defines deterministic topology transitions and relay/SFU layering behavior for 200+ participant voice planning scenarios.
4. Cross-platform profiling model defines measurement taxonomy, platform baselines, and optimization decision thresholds with repeatable evidence requirements.
5. Stress-test contract defines deterministic scenario classes and pass/fail thresholds for 1000-member server and 50-person voice baseline campaigns, plus incremental +50 expansions per encryption/security mode until hard limits are documented.
6. Relay performance/load contract defines capacity envelopes, saturation behaviors, and degradation handling under load.
7. Mobile battery optimization contract defines background-activity budgets, wakeup boundaries, and energy-impact verification criteria.
8. SecurityMode transition and sharding contract defines deterministic thresholding, hysteresis, mode-epoch behavior, and stress-validated fallback posture.
9. Compatibility/governance controls and evidence model are complete and gate-auditable.
10. Integrated validation phase covers positive, adverse, saturation, and recovery pathways across all eight v0.9 bullets.
11. Release/handoff package provides complete traceability, explicit deferrals, and planned-vs-implemented language integrity.

### B.3 QoL integration contract for v0.9 scale-performance journeys (planning-level)

- **Unified health and deterministic next action at scale:** high-load and saturation states in large-server, SFU, relay, and battery-sensitive flows must preserve explicit user state and recovery guidance.
  - **Acceptance criterion:** stress scenarios expose deterministic state/reason/action outcomes for throttling, saturation, and degradation conditions.
  - **Verification evidence:** `V9-G8` contains scale QoL score rows tied to `VA-S*`, `VA-R*`, and `VA-B*` artifacts.
- **Recovery-first call continuity at 200+ participant targets:** cascading-SFU degradation paths prioritize rejoin/switch-path/switch-device guidance and avoid ambiguous terminal outcomes.
  - **Verification evidence:** `V9-G3` and integrated validation include explicit recovery-path evidence with deterministic pass/fail criteria.
- **Cross-device continuity under performance constraints:** attention resume, read continuity, and interaction handoff remain deterministic despite profiling-driven optimizations.
  - **Verification evidence:** battery/perf optimization gates include continuity-regression checks with explicit fail conditions.

### B.4 Accepted defaults from `open_decisions_proposals.md`

- Thread depth stays capped at 2 levels of replies (OD8-01) and inherits the 2-level UX baseline for all downstream docs and tests.
- Open Graph metadata precedence (OG primary, Twitter secondary) is the default card source order (OD8-02) with deterministic conflict logging before presenting values.
- Locale fallback follows OD8-04: language-level fallback then `en-US`, ensuring consistent i18n behavior across profiling/battery work.
- DTLN is the preferred runtime with RNNoise fallback under constrained compute/power budgets (OD8-05); this appears across profiling/perf tasks without speculation.
- OD9-01 through OD9-07 defaults are adopted as v0.9 planning baselines in the relevant phase tasks (pin retention, hierarchy depth, cascade aggressiveness, optimization confidence, relay overload priority, mobile battery policy, multi-provider wake failover), with final closure still requiring V9-G8/V9-G9 evidence signoff.

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
8. Scale-driven security-mode transitions + sharding guidance for huge interactive channels

No additional capability outside these eight bullets is promoted into v0.9 in this plan.

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
1. Any proposal not traceable to one of the eight v0.9 bullets is rejected or formally deferred.
2. IPFS scope is bounded to persistent hosting + server-owner pinning semantics; no tokenized incentive system is introduced.
3. Large-server optimization scope is bounded to hierarchical GossipSub and lazy member loading; no unrelated feature expansion is imported.
4. Cascading voice topology scope is bounded to SFU mesh behavior for 200+ planning target; unrelated media-feature expansion is excluded.
5. Profiling/optimization scope is bounded to measurement and optimization planning across platforms; it does not import v1.0 publication tracks.
6. Stress-testing scope is bounded to the baseline targets (1000-member server, 50-person voice, latency benchmarks) plus explicit +50 increment campaigns used to characterize hard limits per encryption/security mode.
7. Relay optimization scope is bounded to performance/load behavior; governance and role models remain consistent with prior constraints.
8. Mobile battery scope is bounded to background activity reduction and energy-budget planning constraints.
9. Any incompatible protocol behavior discovered during planning must enter major-path governance flow and cannot be silently absorbed into minor evolution.
10. Relay operations remain incentive-free volunteers; no tokens/economic rewards are spun up (RM-02), so the plan layers explicit operator-transfer continuity preparation and multi-provider wake failover posture (RM-05) to avoid single-operator failure modes.

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
| V9-G5 | Stress-testing contract freeze | V9-G4 passed | 1000-member/50-voice/latency baseline scenario contracts complete and +50 increment hard-limit characterization plan/evidence defined |
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
- `VA-C*` reference implementation code and unit/integration tests
- `VA-E*` executable perf/stress harness evidence (commands, reports, and reproducible runs)
- `VA-LG*` legal/governance preparation artifacts (licensing framing + alpha/beta/live policy hardening planning)
- `VA-T7` encryption-method security-floor campaign artifacts
- `VA-R7` operator continuity plan artifacts for relay/mobile governance
- `VA-B7` adaptive battery policy + multi-provider wake failover posture evidence
- `VA-X7`/`VA-X8` policy hardening and governance/legal transition packages

### G.1 Artifact-to-path map (agent execution defaults)

| Artifact Prefix | Primary paths | Expected file patterns |
|---|---|---|
| `VA-G*` | `docs/v0.9/phase0/`, `pkg/v09/conformance/` | `docs/v0.9/phase0/*.md`, `pkg/v09/conformance/*.go` |
| `VA-I*` | `docs/v0.9/phase1/`, `pkg/v09/ipfs/` | `docs/v0.9/phase1/*.md`, `pkg/v09/ipfs/*.go` |
| `VA-L*` | `docs/v0.9/phase2/`, `pkg/v09/scale/` | `docs/v0.9/phase2/*.md`, `pkg/v09/scale/*.go` |
| `VA-S*` | `docs/v0.9/phase3/`, `pkg/v09/sfu/` | `docs/v0.9/phase3/*.md`, `pkg/v09/sfu/*.go` |
| `VA-P*` | `docs/v0.9/phase4/`, `tests/perf/v09/` | `docs/v0.9/phase4/*.md`, `tests/perf/v09/**/*.go` |
| `VA-T*` / `VA-T7` | `docs/v0.9/phase5/`, `tests/perf/v09/` | `docs/v0.9/phase5/*.md`, `tests/perf/v09/**/*.{go,md}` |
| `VA-R*` / `VA-R7` | `docs/v0.9/phase6/`, `pkg/v09/relay/` | `docs/v0.9/phase6/*.md`, `pkg/v09/relay/*.go` |
| `VA-B*` / `VA-B7` | `docs/v0.9/phase7/`, `pkg/v09/mobile/` | `docs/v0.9/phase7/*.md`, `pkg/v09/mobile/*.go` |
| `VA-X*` / `VA-X7` / `VA-X8` | `docs/v0.9/phase8/`, `pkg/v09/governance/` | `docs/v0.9/phase8/*.md`, `pkg/v09/governance/*.go` |
| `VA-H*` / `VA-E*` / `VA-C*` | `docs/v0.9/phase9/`, `tests/e2e/v09/`, `tests/perf/v09/`, `pkg/v09/` | `docs/v0.9/phase9/*.md`, `tests/e2e/v09/**/*.go`, `pkg/v09/**/*.go` |

### Phase 0 - Scope, Governance, and Evidence Foundation (V9-G0)

+ [x] **[P0][Order 01] P0-T1 Freeze v0.9 scope contract and anti-scope boundaries**
  - **Objective:** Establish one-to-one mapping from the eight v0.9 bullets to task and artifact structure.
  - **Concrete actions:**
    - [x] **P0-T1-ST1 Build v0.9 scope trace baseline (8 bullets to task families)**
      - **Objective:** Remove ambiguity in inclusion boundaries.
      - **Concrete actions:** Map each bullet to primary phase, acceptance anchors, and artifact IDs.
      - **Dependencies/prerequisites:** v0.9 scope extraction completed.
      - **Deliverables/artifacts:** Scope trace baseline (`VA-G1`).
      - **Acceptance criteria:** All 8 bullets mapped; no orphan and no extra capability.
      - **Suggested priority/order:** P0, Order 01.1.
      - **Risks/notes:** Unmapped scope introduces hidden execution gaps.
    - [x] **P0-T1-ST2 Lock exclusion policy and escalation route**
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

+ [x] **[P0][Order 02] P0-T2 Lock compatibility/governance controls for v0.9 deltas**
  - **Objective:** Embed additive-evolution and major-path governance controls before domain freezes.
  - **Concrete actions:**
    - [x] **P0-T2-ST1 Define additive-only protobuf checklist for v0.9 schema-touching surfaces**
      - **Objective:** Preserve minor-version compatibility invariants.
      - **Concrete actions:** Define field-addition constraints, reserved-field handling, and downgrade-safe defaults.
      - **Dependencies/prerequisites:** P0-T1.
      - **Deliverables/artifacts:** Additive schema checklist (`VA-G3`).
      - **Acceptance criteria:** All schema-touching tasks include checklist evidence.
      - **Suggested priority/order:** P0, Order 02.1.
      - **Risks/notes:** Hidden schema breaks harm interoperability.
    - [x] **P0-T2-ST2 Define major-path trigger checklist for behavior-breaking proposals**
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

+ [x] **[P0][Order 03] P0-T3 Establish verification matrix and gate-evidence schema for v0.9**
  - **Objective:** Standardize evidence packaging and deterministic gate decisions.
  - **Concrete actions:**
    - [x] **P0-T3-ST1 Define requirement-to-validation matrix template**
      - **Objective:** Ensure every scope bullet has positive, adverse, saturation, and recovery coverage.
      - **Concrete actions:** Define matrix fields for requirement ID, task IDs, artifact IDs, gate ownership, and evidence status.
      - **Dependencies/prerequisites:** P0-T1.
      - **Deliverables/artifacts:** Validation matrix template (`VA-G5`).
      - **Acceptance criteria:** Template supports all 8 bullets and all v0.9 gates.
      - **Suggested priority/order:** P0, Order 03.1.
      - **Risks/notes:** Weak template quality creates inconsistent gate closure.
    - [x] **P0-T3-ST2 Define gate evidence-bundle conventions**
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

+ [x] **[P0][Order 03.4] P0-T4 Capture legal/governance transition posture for RM-04**
  - **Objective:** Document planning-level legal guardrails for open-source v0.9 while preparing the minimal text required for the later consortium transition.
  - **Concrete actions:**
    - [x] **P0-T4-ST1 Document current open-source/licensing posture and liability disclaimer requirements**, referencing AGPL code + CC-BY-SA specification defaults and mapping required legal-text artifacts for the planned consortium transition.
      - **Dependencies/prerequisites:** AGENTS.md constraints, legal input.
      - **Deliverables/artifacts:** Legal posture note (`VA-LG1`).
      - **Acceptance criteria:** Note lists current AGPL/CC-BY-SA defaults, required disclaimers, and explicit evidence anchors for future review.
      - **Suggested priority/order:** P0, Order 03.4.1.
    - [x] **P0-T4-ST2 Outline alpha/beta/live policy hardening assumptions and required revisit triggers** to frame the future consortium governance path.
      - **Dependencies/prerequisites:** RM-04 policy expectations, P8-T3 evidence requirements.
      - **Deliverables/artifacts:** Policy hardening timeline (`VA-LG2`).
      - **Acceptance criteria:** Timeline links alpha/beta/live stages to concrete hardening tasks, evidence needs, and legal review triggers.
      - **Suggested priority/order:** P0, Order 03.4.2.
  - **Exposure:** Plan remains at documentation level; no implementation claim is made.

### Phase 1 - IPFS Persistent Hosting Contracts (V9-G1)

+ [x] **[P0][Order 04] P1-T1 Define IPFS content-addressing and pinning responsibility model**
  - **Objective:** Specify deterministic persistent-hosting contract with server-owner pinning posture.
  - **Concrete actions:**
    - [x] **P1-T1-ST1 Define content envelope and addressing metadata contract**
      - **Objective:** Ensure deterministic content identity and lifecycle references.
      - **Concrete actions:** Define content identifiers, metadata constraints, and mapping boundaries with existing file contexts.
      - **Dependencies/prerequisites:** P0-T2, v0.3 file-transfer baseline, v0.7 retention baseline.
      - **Deliverables/artifacts:** Content-address contract (`VA-I1`).
      - **Acceptance criteria:** Equivalent content/metadata inputs resolve to deterministic address model.
      - **Suggested priority/order:** P0, Order 04.1.
      - **Risks/notes:** Address ambiguity undermines persistence guarantees.
    - [x] **P1-T1-ST2 Define server-owner pinning roles and lifecycle responsibilities**
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

+ [x] **[P0][Order 05] P1-T2 Define persistent-hosting retention, retrieval, and degraded behavior**
  - **Objective:** Specify deterministic retention/retrieval semantics and degradation handling.
  - **Concrete actions:**
    - [x] **P1-T2-ST1 Define retention horizon and retrieval boundary semantics**
      - **Objective:** Ensure predictable availability expectations.
      - **Concrete actions:** Define retention states, retrieval resolution outcomes, and stale/missing-content handling taxonomy.
      - **Dependencies/prerequisites:** P1-T1, v0.7 history/retention context.
      - **Deliverables/artifacts:** Retention/retrieval contract (`VA-I3`).
      - **Acceptance criteria:** Retrieval outcomes are deterministic for present, stale, and missing content.
      - **Suggested priority/order:** P0, Order 05.1.
      - **Risks/notes:** Weak degraded semantics harm user trust.
    - [x] **P1-T2-ST2 Define degraded-mode fallback and operator diagnostics expectations**
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

+ [x] **[P1][Order 06] P1-T3 Define IPFS governance boundary and interoperability notes**
  - **Objective:** Keep IPFS planning compatible with protocol governance and versioning constraints.
  - **Concrete actions:**
    - [x] **P1-T3-ST1 Define additive evolution notes for persistence-related schema touches**
      - **Objective:** Preserve minor-version compatibility behavior.
      - **Concrete actions:** Record extension hooks, reserved fields, and backward-compatible defaults.
      - **Dependencies/prerequisites:** P1-T1, P0-T2.
      - **Deliverables/artifacts:** IPFS evolution note (`VA-I5`).
      - **Acceptance criteria:** All schema-delta notes are additive and auditable.
      - **Suggested priority/order:** P1, Order 06.1.
      - **Risks/notes:** Non-additive drift creates compatibility debt.
    - [x] **P1-T3-ST2 Define major-path escalation examples for persistence behavior breakage**
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

+ [x] **[P0][Order 07] P2-T1 Define hierarchical GossipSub topology contract for large servers**
  - **Objective:** Specify deterministic hierarchy behavior and fanout boundaries.
  - **Concrete actions:**
    - [x] **P2-T1-ST1 Define hierarchy roles and topic-segmentation strategy**
      - **Objective:** Standardize large-server pubsub partitioning assumptions.
      - **Concrete actions:** Define hierarchy layers, topic segmentation boundaries, and relay/peer role assumptions without privileged protocol classes.
      - **Dependencies/prerequisites:** v0.1 GossipSub baseline, P0-T2.
      - **Deliverables/artifacts:** Hierarchical topology contract (`VA-L1`).
      - **Acceptance criteria:** Hierarchy behavior is deterministic and protocol-first.
      - **Suggested priority/order:** P0, Order 07.1.
      - **Risks/notes:** Role ambiguity can create uneven load.
    - [x] **P2-T1-ST2 Define propagation and backpressure semantics under high membership**
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

+ [x] **[P0][Order 08] P2-T2 Define lazy member-loading model and state-coherence boundaries**
  - **Objective:** Specify deterministic lazy loading for member state in large communities.
  - **Concrete actions:**
    - [x] **P2-T2-ST1 Define member state classes and incremental-loading triggers**
      - **Objective:** Standardize when and how member subsets load.
      - **Concrete actions:** Define active/nearby/passive member state classes and deterministic load triggers.
      - **Dependencies/prerequisites:** P2-T1, v0.2 presence baseline.
      - **Deliverables/artifacts:** Lazy loading trigger contract (`VA-L3`).
      - **Acceptance criteria:** Load-trigger behavior is deterministic and test-mappable.
      - **Suggested priority/order:** P0, Order 08.1.
      - **Risks/notes:** Trigger ambiguity causes state churn.
    - [x] **P2-T2-ST2 Define consistency guarantees and staleness handling**
      - **Objective:** Bound stale-view behavior and recovery expectations.
      - **Concrete actions:** Define staleness windows, refresh semantics, and fallback behavior when member snapshots diverge.
      - **Dependencies/prerequisites:** P2-T2-ST1.
      - **Deliverables/artifacts:** State coherence contract (`VA-L4`).
      - **Acceptance criteria:** Stale/refresh outcomes are deterministic.
      - **Suggested priority/order:** P0, Order 08.2.
      - **Risks/notes:** Inconsistent member state undermines moderation and presence trust.
    - [x] **P2-T2-ST3 Define scale-driven `SecurityMode` transition triggers and sharding guidance for huge interactive channels**
      - **Objective:** Provide a deterministic, performance-aware contract for when and how very large channels change encryption posture and/or shard.
      - **Concrete actions:** Specify default thresholds + hysteresis, writer/reader semantics (interactive vs broadcast), shard model (naming, membership, fanout), and invariants (no silent downgrade; mode change = new `mode_epoch_id`).
      - **Dependencies/prerequisites:** P2-T2-ST2, v0.4 SecurityMode baseline, `ENCRYPTION_PLUS.md` scaling guidance (threshold profiles + hysteresis envelopes).
      - **Deliverables/artifacts:** Mode-transition + sharding guidance contract (`VA-L7`).
      - **Acceptance criteria:** Guidance is deterministic, testable via stress profiles, and includes explicit “cannot migrate history” fallback posture where applicable.
      - **Suggested priority/order:** P1, Order 08.3.
      - **Risks/notes:** Keep this as a contract/guidance surface; avoid importing full UX redesign into v0.9.
  - **Dependencies/prerequisites:** P2-T1.
  - **Deliverables/artifacts:** Lazy-loading package (`VA-L3`, `VA-L4`, `VA-L7`).
  - **Acceptance criteria:** V9-G2 lazy-loading criteria met.
  - **Suggested priority/order:** P0, Order 08.
  - **Risks/notes:** Avoid importing non-scope UI redesign tracks.

+ [x] **[P1][Order 09] P2-T3 Define large-server optimization observability and rollback boundaries**
  - **Objective:** Provide deterministic instrumentation and fallback posture for hierarchy/lazy-load behavior.
  - **Concrete actions:**
    - [x] **P2-T3-ST1 Define observability signals for hierarchy and lazy-load performance**
      - **Objective:** Ensure planning-level metrics support validation and tuning.
      - **Concrete actions:** Define signal catalog for fanout pressure, load latency, staleness rates, and recovery events.
      - **Dependencies/prerequisites:** P2-T1, P2-T2.
      - **Deliverables/artifacts:** Large-server observability contract (`VA-L5`).
      - **Acceptance criteria:** Signal set fully covers hierarchy and lazy-load behavior.
      - **Suggested priority/order:** P1, Order 09.1.
      - **Risks/notes:** Sparse telemetry hides failure modes.
    - [x] **P2-T3-ST2 Define fallback/rollback decision thresholds**
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

+ [x] **[P0][Order 10] P3-T1 Define cascading SFU topology and election layering contract**
  - **Objective:** Specify deterministic cascade topology for 200+ participant voice targets.
  - **Concrete actions:**
    - [x] **P3-T1-ST1 Define tiered SFU roles, segment boundaries, and election assumptions**
      - **Objective:** Standardize cascade structure while preserving single-binary mode model.
      - **Concrete actions:** Define tier roles, segment assignment logic, and election boundaries across peer/relay SFU candidates.
      - **Dependencies/prerequisites:** v0.3 voice/media baseline, P0-T2.
      - **Deliverables/artifacts:** Cascading topology contract (`VA-S1`).
      - **Acceptance criteria:** Segment/election behavior is deterministic and role-safe.
      - **Suggested priority/order:** P0, Order 10.1.
      - **Risks/notes:** Topology ambiguity creates instability at scale.
    - [x] **P3-T1-ST2 Define inter-tier signaling and control-channel behavior**
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

+ [x] **[P0][Order 11] P3-T2 Define media forwarding, failover, and degradation semantics in cascades**
  - **Objective:** Specify deterministic forwarding and failover behavior across cascade tiers.
  - **Concrete actions:**
    - [x] **P3-T2-ST1 Define forwarding path selection and quality-preservation boundaries**
      - **Objective:** Ensure predictable media path decisions at high participant counts.
      - **Concrete actions:** Define path-selection heuristics, jitter/latency boundaries, and forward-only assumptions.
      - **Dependencies/prerequisites:** P3-T1.
      - **Deliverables/artifacts:** Forwarding path contract (`VA-S3`).
      - **Acceptance criteria:** Path decisions are deterministic under equivalent conditions.
      - **Suggested priority/order:** P0, Order 11.1.
      - **Risks/notes:** Path ambiguity produces inconsistent quality outcomes.
    - [x] **P3-T2-ST2 Define tier failover, rejoin, and split/merge behavior**
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

+ [x] **[P1][Order 12] P3-T3 Define governance and compatibility boundaries for cascade evolution**
  - **Objective:** Ensure cascade behavior changes remain compatibility-safe and governance-compliant.
  - **Concrete actions:**
    - [x] **P3-T3-ST1 Define additive evolution guidance for cascade capability signaling**
      - **Objective:** Preserve minor-path compatibility.
      - **Concrete actions:** Define capability field extension boundaries and downgrade-safe defaults.
      - **Dependencies/prerequisites:** P3-T1, P0-T2.
      - **Deliverables/artifacts:** Cascade evolution note (`VA-S5`).
      - **Acceptance criteria:** Evolution guidance aligns with additive-only discipline.
      - **Suggested priority/order:** P1, Order 12.1.
      - **Risks/notes:** Capability drift can fragment interoperability.
    - [x] **P3-T3-ST2 Define major-path trigger examples for topology-breaking changes**
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

+ [x] **[P0][Order 13] P4-T1 Define cross-platform profiling taxonomy and baseline metrics**
  - **Objective:** Specify deterministic profiling model across desktop/mobile/web-capable targets.
  - **Concrete actions:**
    - [x] **P4-T1-ST1 Define metric catalog and platform normalization rules**
      - **Objective:** Ensure consistent measurement semantics across environments.
      - **Concrete actions:** Define CPU, memory, network, latency, jitter, startup, and background activity metric definitions with normalization assumptions.
      - **Dependencies/prerequisites:** P0-T3, v0.8 cross-platform context.
      - **Deliverables/artifacts:** Profiling taxonomy contract (`VA-P1`).
      - **Acceptance criteria:** Metric definitions are unambiguous and comparable across platforms.
      - **Suggested priority/order:** P0, Order 13.1.
      - **Risks/notes:** Inconsistent metrics invalidate optimization decisions.
    - [x] **P4-T1-ST2 Define baseline-capture protocol and environment controls**
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

+ [x] **[P0][Order 14] P4-T2 Define optimization decision framework and threshold policy**
  - **Objective:** Specify deterministic criteria for prioritizing optimization paths.
  - **Concrete actions:**
    - [x] **P4-T2-ST1 Define bottleneck classification and prioritization rules**
      - **Objective:** Standardize optimization triage decisions.
      - **Concrete actions:** Define classification for CPU-bound, network-bound, memory-bound, I/O-bound, and synchronization-bound cases.
      - **Dependencies/prerequisites:** P4-T1.
      - **Deliverables/artifacts:** Bottleneck classification contract (`VA-P3`).
      - **Acceptance criteria:** Equivalent profiles map to consistent bottleneck classes.
      - **Suggested priority/order:** P0, Order 14.1.
      - **Risks/notes:** Misclassification yields poor optimization sequencing.
    - [x] **P4-T2-ST2 Define optimization acceptance thresholds and regression boundaries**
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

+ [x] **[P1][Order 15] P4-T3 Define cross-platform optimization evidence and reporting format**
  - **Objective:** Ensure optimization findings are traceable and governance-auditable.
  - **Concrete actions:**
    - [x] **P4-T3-ST1 Define platform comparison reporting schema**
      - **Objective:** Normalize reporting of profile deltas and tradeoffs.
      - **Concrete actions:** Define schema for before/after baseline deltas, confidence, caveats, and unresolved constraints.
      - **Dependencies/prerequisites:** P4-T1, P4-T2.
      - **Deliverables/artifacts:** Performance reporting schema (`VA-P5`).
      - **Acceptance criteria:** Reports are comparable across platform families.
      - **Suggested priority/order:** P1, Order 15.1.
      - **Risks/notes:** Non-standard reports obscure regressions.
    - [x] **P4-T3-ST2 Define optimization rollback and exception documentation rules**
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
  - **Risks/notes:** Preserve planned-vs-implemented wording in all evidence templates.

### Phase 5 - Stress-Testing Contracts (V9-G5)

- [x] **[P0][Order 16] P5-T1 Define 1000-member server stress-test scenario contract**
  - **Objective:** Specify deterministic large-server stress scenarios and boundary conditions.
  - **Concrete actions:**
    - [x] **P5-T1-ST1 Define membership churn, activity distribution, and load-shape profiles**
      - **Objective:** Ensure scenarios represent realistic and adversarial load envelopes.
      - **Concrete actions:** Define join/leave churn patterns, message-rate distributions, peak burst profiles, and scale-up profiles that exercise mode transitions/sharding (e.g., channels growing across `SecurityMode` thresholds).
      - **Dependencies/prerequisites:** P2-T1, P2-T2, P4-T1.
      - **Deliverables/artifacts:** 1000-member scenario contract (`VA-T1`).
      - **Acceptance criteria:** Scenario profiles are deterministic and reproducible.
      - **Suggested priority/order:** P0, Order 16.1.
      - **Risks/notes:** Narrow profiles under-represent real load.
    - [x] **P5-T1-ST2 Define pass/fail thresholds and saturation classification**
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

- [x] **[P0][Order 17] P5-T2 Define 50-person voice stress and latency benchmark contract**
  - **Objective:** Specify deterministic voice-load and latency benchmark scenarios.
  - **Concrete actions:**
    - [x] **P5-T2-ST1 Define 50-participant voice scenario classes and topology assumptions**
      - **Objective:** Align stress tests with cascade/topology planning boundaries.
      - **Concrete actions:** Define speaking-ratio profiles, packet-loss/jitter classes, and topology states.
      - **Dependencies/prerequisites:** P3-T1, P3-T2, P4-T1.
      - **Deliverables/artifacts:** Voice stress scenario contract (`VA-T3`).
      - **Acceptance criteria:** Scenario classes are deterministic and topology-aware.
      - **Suggested priority/order:** P0, Order 17.1.
      - **Risks/notes:** Missing topology context reduces benchmark relevance.
    - [x] **P5-T2-ST2 Define latency benchmark method and acceptance bands**
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

- [x] **[P1][Order 18] P5-T3 Define stress-campaign execution governance and evidence quality controls**
  - **Objective:** Ensure stress evidence is repeatable, auditable, and compatible with gate logic.
  - **Concrete actions:**
    - [x] **P5-T3-ST1 Define runbook and evidence capture requirements for stress campaigns**
      - **Objective:** Normalize stress campaign execution and reporting.
      - **Concrete actions:** Define runbook fields, scenario IDs, environment fingerprints, and failure-capture minimums.
      - **Dependencies/prerequisites:** P5-T1, P5-T2.
      - **Deliverables/artifacts:** Stress campaign runbook contract (`VA-T5`).
      - **Acceptance criteria:** All stress runs produce complete, comparable evidence bundles.
      - **Suggested priority/order:** P1, Order 18.1.
      - **Risks/notes:** Incomplete runbook data breaks comparability.
    - [x] **P5-T3-ST2 Define anomaly triage taxonomy and rerun criteria**
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
  - **Risks/notes:** Keep all wording planned-vs-implemented.

- [x] **[P0][Order 18.5] P5-T4 Characterize security-floor limits per encryption method (RM-03)**
  - **Objective:** Execute incremental +50 participant campaigns for each encryption/security mode until the hard limit of that method is documented, ensure no silent downgrade occurs, and capture resilience evidence.
  - **Concrete actions:**
    - [x] **P5-T4-ST1 Run cascaded stress campaigns per applicable encryption method, starting from baseline and adding 50 participants per run, documenting the precise point each method reaches its hard limit and the triggered mitigation (security-mode transition or shard split).**
      - **Dependencies/prerequisites:** P5-T1, P2-T2, P4-T1.
      - **Deliverables/artifacts:** Security-floor campaign log (`VA-T7`).
      - **Acceptance criteria:** Each method has a reproducible +50 incremental run series plus documented decision on how thresholds translate to safeguards.
      - **Suggested priority/order:** P0, Order 18.5.
    - [x] **P5-T4-ST2 Summarize security-floor findings, highlight explicit non-silent downgrade guidance, and reference gating evidence for each mode.**
      - **Dependencies/prerequisites:** P5-T4-ST1.
      - **Deliverables/artifacts:** Security-floor summary (`VA-T7`).
      - **Acceptance criteria:** Summary links each threshold to deterministic behavior (no silent downgrade) and policy references.
      - **Suggested priority/order:** P1, Order 18.6.
  - **Risks/notes:** Campaigns must terminate before exposing insecure fallback; document residual uncertainties as open decision evidence.

### Phase 6 - Relay Performance and Load-Testing Contracts (V9-G6)

- [x] **[P0][Order 19] P6-T1 Define relay capacity model and saturation boundaries**
  - **Objective:** Specify deterministic relay capacity envelopes and saturation behavior.
  - **Concrete actions:**
    - [x] **P6-T1-ST1 Define relay workload classes and resource budget model**
      - **Objective:** Standardize relay workload characterization.
      - **Concrete actions:** Define workload classes for DHT, relay circuits, store-forward, and SFU-forwarding interactions.
      - **Dependencies/prerequisites:** v0.1 relay baseline, v0.7 store-forward context, P4-T1.
      - **Deliverables/artifacts:** Relay workload model (`VA-R1`).
      - **Acceptance criteria:** Workload classes map deterministically to resource budget dimensions.
      - **Suggested priority/order:** P0, Order 19.1.
      - **Risks/notes:** Poor modeling leads to unstable capacity assumptions.
    - [x] **P6-T1-ST2 Define saturation behavior and fixed service-priority policies**
      - **Objective:** Bound degraded behavior under relay overload.
      - **Concrete actions:** Define priority semantics, shedding policies, and degradation classes with the priority order: control plane -> active media -> store-forward -> bulk sync baked in as the default sequence.
      - **Dependencies/prerequisites:** P6-T1-ST1.
      - **Deliverables/artifacts:** Saturation policy contract (`VA-R2`).
      - **Acceptance criteria:** Overload behavior is deterministic, policy-consistent, and follows the mandated priority order.
      - **Suggested priority/order:** P0, Order 19.2.
      - **Risks/notes:** Undefined overload policy can cause cascading failures.
  - **Dependencies/prerequisites:** P4-T1, P4-T2, P4-T3, v0.1 relay baseline.
  - **Deliverables/artifacts:** Relay capacity package (`VA-R1`, `VA-R2`).
  - **Acceptance criteria:** V9-G6 capacity criteria met.
  - **Suggested priority/order:** P0, Order 19.
  - **Risks/notes:** Maintain no-special-node invariant.

- [x] **[P0][Order 20] P6-T2 Define relay load-testing scenarios and acceptance thresholds**
  - **Objective:** Specify deterministic relay load-test campaigns aligned with v0.9 scope.
  - **Concrete actions:**
    - [x] **P6-T2-ST1 Define relay load profiles and fault-injection classes**
      - **Objective:** Cover nominal, peak, and faulted relay conditions.
      - **Concrete actions:** Define connection churn, traffic distributions, burst classes, and fault injection types.
      - **Dependencies/prerequisites:** P6-T1, P5-T1.
      - **Deliverables/artifacts:** Relay load scenario contract (`VA-R3`).
      - **Acceptance criteria:** Load profiles are complete and reproducible.
      - **Suggested priority/order:** P0, Order 20.1.
      - **Risks/notes:** Narrow profile coverage misses critical behavior.
    - [x] **P6-T2-ST2 Define relay load pass/fail and degradation-recovery thresholds**
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

- [x] **[P1][Order 21] P6-T3 Define relay optimization-change governance and rollback controls**
  - **Objective:** Ensure relay optimization proposals remain governance-compliant and reversible.
  - **Concrete actions:**
    - [x] **P6-T3-ST1 Define optimization-change classification and approval workflow**
      - **Objective:** Separate minor-safe adjustments from major-path behavior changes.
      - **Concrete actions:** Define classification criteria, approval checkpoints, and evidence requirements.
      - **Dependencies/prerequisites:** P6-T1, P6-T2, P0-T2.
      - **Deliverables/artifacts:** Relay optimization governance workflow (`VA-R5`).
      - **Acceptance criteria:** Every change proposal maps to a deterministic governance path.
      - **Suggested priority/order:** P1, Order 21.1.
      - **Risks/notes:** Ambiguous classification risks governance bypass.
    - [x] **P6-T3-ST2 Define rollback protocol and post-incident evidence expectations**
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

- [x] **[P1][Order 21.5] P6-T4 Plan operator handover continuity for relay governance (RM-05)**
  - **Objective:** Ensure the single-operator startup phase can transition to new operator(s) without loss of continuity or telemetry.
  - **Concrete actions:**
    - [x] **P6-T4-ST1 Document the initial operator-to-successor handoff playbook (credentials, config, monitoring handover, emergency rollback)** with explicit roles and evidence anchors.
      - **Dependencies/prerequisites:** P6-T1 through P6-T3.
      - **Deliverables/artifacts:** Operator handover plan (`VA-R7`).
      - **Acceptance criteria:** Playbook describes every step required to transfer authority, ensure no single operator lock-in, and preserve security keys.
      - **Suggested priority/order:** P1, Order 21.5a.
    - [x] **P6-T4-ST2 Validate continuity plan with multi-provider-wake/backstop assumptions to guard against single-operator downtime.**
      - **Dependencies/prerequisites:** P7-T1, P7-T4.
      - **Deliverables/artifacts:** Continuity validation note (`VA-R7`).
      - **Acceptance criteria:** Simulation covers failure of the initial operator and resumption under successor or multi-provider architecture with no missing state.
      - **Suggested priority/order:** P1, Order 21.5b.
  - **Risks/notes:** Operator transfer must not imply new privileged node classes; align with economic incentive constraint.

### Phase 7 - Mobile Battery Optimization Contracts (V9-G7)

- [x] **[P0][Order 22] P7-T1 Define mobile background-activity budget and wakeup policy contract**
  - **Objective:** Specify deterministic background activity reduction boundaries for mobile.
  - **Concrete actions:**
    - [x] **P7-T1-ST1 Define adaptive background task classes and energy budget envelopes**
      - **Objective:** Bound background workloads with adaptive policies that scale with battery level and multi-provider wake/notification risk.
      - **Concrete actions:** Define task categories, periodicity assumptions, and envelope adjustments for networking/sync/activity updates when additional provider wake paths become available.
      - **Dependencies/prerequisites:** P4-T1, v0.7 notification context, v0.8 platform context.
      - **Deliverables/artifacts:** Background budget contract (`VA-B1`).
      - **Acceptance criteria:** Background workloads map deterministically to budget classes.
      - **Suggested priority/order:** P0, Order 22.1.
      - **Risks/notes:** Unbounded tasks can degrade battery unpredictably.
    - [x] **P7-T1-ST2 Define wakeup triggers, suppression rules, and multi-provider failover behavior**
      - **Objective:** Control wake frequency while preserving protocol continuity assumptions and ensuring multi-provider delivery failover.
      - **Concrete actions:** Define wake trigger taxonomy, suppression precedence, safe fallback when wake budget is exceeded, and provider failover/handoff mock-ups for multi-provider wake paths.
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

- [x] **[P0][Order 23] P7-T2 Define battery-impact profiling and optimization validation contract**
  - **Objective:** Specify deterministic battery-impact measurement and acceptance logic.
  - **Concrete actions:**
    - [x] **P7-T2-ST1 Define battery-impact measurement scenarios and controls**
      - **Objective:** Ensure energy measurements are comparable and reproducible.
      - **Concrete actions:** Define idle/background/active scenario classes, device-state controls, and capture intervals.
      - **Dependencies/prerequisites:** P7-T1, P4-T1.
      - **Deliverables/artifacts:** Battery scenario contract (`VA-B3`).
      - **Acceptance criteria:** Scenario design supports deterministic battery-impact comparison.
      - **Suggested priority/order:** P0, Order 23.1.
      - **Risks/notes:** Noisy measurement controls obscure improvements.
    - [x] **P7-T2-ST2 Define battery optimization acceptance thresholds and regression guards**
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
  - **Risks/notes:** Maintain planned-vs-implemented framing, no completion claims.

- [x] **[P1][Order 24] P7-T3 Define battery/performance tradeoff governance and user-impact boundaries**
  - **Objective:** Ensure battery reductions do not silently violate core performance assumptions.
  - **Concrete actions:**
    - [x] **P7-T3-ST1 Define tradeoff matrix for battery vs latency/responsiveness**
      - **Objective:** Make compromise boundaries explicit and reviewable.
      - **Concrete actions:** Define tradeoff matrix dimensions and acceptable compromise ranges.
      - **Dependencies/prerequisites:** P7-T1, P7-T2, P4-T2.
      - **Deliverables/artifacts:** Tradeoff matrix contract (`VA-B5`).
      - **Acceptance criteria:** Tradeoff decisions are deterministic and policy-bounded.
      - **Suggested priority/order:** P1, Order 24.1.
      - **Risks/notes:** Hidden tradeoffs can degrade real-time experience.
    - [x] **P7-T3-ST2 Define escalation triggers for behavior-breaking battery policies**
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

- [x] **[P1][Order 24.5] P7-T4 Plan mobile operator continuity and multi-provider wake failover (RM-05)**
  - **Objective:** Document how the initial mobile operator (single-server coordinator) hands over responsibilities and how clients fail over to alternative providers for wake paths.
  - **Concrete actions:**
    - [x] **P7-T4-ST1 Capture handover roles, secrets/keys, monitoring updates, and recovery procedures when the initial operator exits or fails.**
      - **Dependencies/prerequisites:** P7-T1, P7-T2, P6-T4.
      - **Deliverables/artifacts:** Mobile operator continuity plan (`VA-B7`).
      - **Acceptance criteria:** Plan describes rekeying steps, config sync, and evidence that clients can bootstrap from new operators with no lost messages.
      - **Suggested priority/order:** P1, Order 24.5a.
    - [x] **P7-T4-ST2 Define multi-provider wake failover testing posture, including provider discovery, jittered retries, and policy constraints to preserve decentralization.**
      - **Dependencies/prerequisites:** P7-T1-ST2, P5-T4-ST2.
      - **Deliverables/artifacts:** Wake failover validation note (`VA-B7`).
      - **Acceptance criteria:** Multi-provider test plan demonstrates seamless failover when one provider drops.
      - **Suggested priority/order:** P1, Order 24.5b.
  - **Risks/notes:** Continuity plan must not introduce new centralized control points.

### Phase 8 - Integrated Validation and Governance Readiness (V9-G8)

- [x] **[P0][Order 25] P8-T1 Build integrated cross-domain validation matrix**
  - **Objective:** Validate interactions across all eight v0.9 scope bullets.
  - **Concrete actions:**
    - [x] **P8-T1-ST1 Define end-to-end scenario matrix (positive/adverse/saturation/recovery)**
      - **Objective:** Ensure cross-domain behavior is validated, not only isolated contracts.
      - **Concrete actions:** Build scenario set spanning IPFS persistence, large-server pubsub, cascading SFU, profiling, stress, relay load, and battery controls.
      - **Dependencies/prerequisites:** P1-T1 through P7-T3.
      - **Deliverables/artifacts:** Integrated scenario matrix (`VA-X1`).
      - **Acceptance criteria:** Every scope bullet appears in integrated coverage.
      - **Suggested priority/order:** P0, Order 25.1.
      - **Risks/notes:** Missing interactions conceal systemic risks.
    - [x] **P8-T1-ST2 Define integrated pass/fail thresholds and evidence-link rules**
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

- [x] **[P0][Order 26] P8-T2 Perform compatibility/governance/invariant conformance audit**
  - **Objective:** Confirm additive evolution, major-path governance, and architecture invariants.
  - **Concrete actions:**
    - [x] **P8-T2-ST1 Run additive-only conformance audit across schema/capability deltas**
      - **Objective:** Ensure minor-path deltas remain compatibility-safe.
      - **Concrete actions:** Audit all relevant artifacts against `VA-G3`; record pass/fail and exceptions.
      - **Dependencies/prerequisites:** P0-T2, P1-T1 through P7-T3.
      - **Deliverables/artifacts:** Additive conformance report (`VA-X3`).
      - **Acceptance criteria:** All minor-path deltas are compliant or escalated.
      - **Suggested priority/order:** P0, Order 26.1.
      - **Risks/notes:** Non-compliance silently breaks interoperability.
    - [x] **P8-T2-ST2 Run major-path trigger and invariant conformance audit**
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

- [x] **[P1][Order 27] P8-T3 Perform open-decision and licensing/repository-language conformance review**
  - **Objective:** Preserve unresolved decisions and document-language integrity.
  - **Concrete actions:**
    - [x] **P8-T3-ST1 Validate open-decision handling discipline**
      - **Objective:** Ensure unresolved questions are not presented as settled facts.
      - **Concrete actions:** Audit open-decision references for status, owner role, revisit gate, and handling-rule consistency.
      - **Dependencies/prerequisites:** P1-T1 through P8-T2.
      - **Deliverables/artifacts:** Open-decision conformance report (`VA-X5`).
      - **Acceptance criteria:** All unresolved decisions remain explicitly open.
      - **Suggested priority/order:** P1, Order 27.1.
      - **Risks/notes:** Wording drift can create false certainty.
    - [x] **P8-T3-ST2 Validate licensing and repository-state language alignment**
      - **Objective:** Preserve AGPL/CC-BY-SA alignment and documentation-only framing.
      - **Concrete actions:** Audit artifacts for licensing and planned-vs-implemented language consistency.
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

- [x] **[P1][Order 27.5] P8-T4 Document alpha/beta/live policy hardening and legal transition assumptions (RM-04)**
  - **Objective:** Prepare governance/legal artifacts clarifying how the open-source v0.9 posture evolves toward a future consortium while capturing alpha/beta/live policy hardening assumptions.
  - **Concrete actions:**
    - [x] **P8-T4-ST1 Summarize legal transition assumptions, required liability text, and triggers for consortium-level review without claiming any license change is complete.**
      - **Dependencies/prerequisites:** P0-T4-ST1, legal/governance counsel.
      - **Deliverables/artifacts:** Legal transition summary (`VA-X7`).
      - **Acceptance criteria:** Summary references current hosting/licensing posture, outlines minimal liability text needs, and enumerates evidence triggers for future consortium formation.
      - **Suggested priority/order:** P1, Order 27.5a.
    - [x] **P8-T4-ST2 Align alpha/beta/live policy hardening assumptions with gate checkpoints and revisit triggers (OD4-02).**
      - **Dependencies/prerequisites:** P8-T3-ST1, OD4-02 inputs.
      - **Deliverables/artifacts:** Policy hardening plan (`VA-X8`).
      - **Acceptance criteria:** Plan ties each stage to measurable hardening tasks, evidence links, and revisit triggers, while explicitly treating the stages as in-progress.
      - **Suggested priority/order:** P1, Order 27.5b.
  - **Risks/notes:** Remain in planning language; do not assert the consortium or policy hardening is already in place.

### Phase 9 - Reference Implementation, Perf/Stress Validation, and v0.9 Shipping (V9-G9)

- [x] **[P0][Order 28] P9-T1 Close scope-to-task-to-artifact traceability**
  - **Objective:** Achieve complete auditable traceability for all v0.9 scope bullets.
  - **Concrete actions:**
    - [x] **P9-T1-ST1 Build final traceability closure matrix with acceptance anchors**
      - **Objective:** Link every v0.9 bullet to tasks, artifacts, and gate acceptance anchors.
      - **Concrete actions:** Compile mapping table, verify evidence-link completeness, and mark unresolved gaps.
      - **Dependencies/prerequisites:** P8-T1 through P8-T3.
      - **Deliverables/artifacts:** Final traceability matrix (`VA-H1`).
      - **Acceptance criteria:** All 8 bullets have complete mapping and acceptance anchors.
      - **Suggested priority/order:** P0, Order 28.1.
      - **Risks/notes:** Missing traceability blocks handoff.
    - [x] **P9-T1-ST2 Execute anti-scope-creep closure audit**
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

- [x] **[P1][Order 29] P9-T2 Build release-conformance checklist with evidence linkage (including perf/stress reports)**
  - **Objective:** Provide deterministic go/no-go planning handoff checklist.
  - **Concrete actions:**
    - [x] **P9-T2-ST1 Assemble gate-aligned conformance checklist sections**
      - **Objective:** Cover scope, sequencing, compatibility, governance, validation, and language integrity.
      - **Concrete actions:** Build checklist sections, include gate references and pass/fail fields.
      - **Dependencies/prerequisites:** P9-T1.
      - **Deliverables/artifacts:** Release-conformance checklist (`VA-H3`).
      - **Acceptance criteria:** Checklist supports deterministic handoff decision.
      - **Suggested priority/order:** P1, Order 29.1.
      - **Risks/notes:** Incomplete checklist creates ambiguous closure.
    - [x] **P9-T2-ST2 Map every checklist item to owner role and evidence ID**
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
  - **Risks/notes:** Keep planned-vs-implemented language explicit.

- [x] **[P1][Order 30] P9-T3 Prepare operator/reviewer dossier and forward deferral register**
  - **Objective:** Finalize execution-ready planning handoff with explicit future deferrals.
  - **Concrete actions:**
    - [x] **P9-T3-ST1 Compile execution handoff dossier with gate outcomes and residual risks**
      - **Objective:** Provide complete orchestration input without implementation completion claims.
      - **Concrete actions:** Aggregate gate outcomes, evidence index, unresolved decisions, and residual-risk summary.
      - **Dependencies/prerequisites:** P9-T2.
      - **Deliverables/artifacts:** Execution handoff dossier (`VA-H5`).
      - **Acceptance criteria:** Dossier is complete, internally consistent, and planned-vs-implemented.
      - **Suggested priority/order:** P1, Order 30.1.
      - **Risks/notes:** Missing context creates execution ambiguity.
    - [x] **P9-T3-ST2 Build v1.0+/post-v1 deferral register from v0.9 residuals**
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


- [x] **[P0][Order 31] P9-T4 Implement IPFS persistent hosting for server attachments and history objects**
  - **Objective:** Implement the IPFS hosting layer defined in `P1-T1`-`P1-T3` so servers can persist and retrieve large objects without central storage.
  - **Concrete actions:**
    - [x] **P9-T4-ST1 Implement IPFS integration module (pin, unpin, fetch, GC) with bounded resources**
      - **Objective:** Provide a deployable, testable IPFS adapter with safe defaults.
      - **Concrete actions:** Implement client wrapper; enforce timeouts and size bounds; implement pinset management; add mock mode for tests.
      - **Dependencies/prerequisites:** `P1-T1`, `P1-T2`.
      - **Deliverables/artifacts:** `pkg/v09/ipfs/*` + unit tests (`VA-C1`).
      - **Acceptance criteria:** Unit tests cover pin/fetch/GC; mock mode allows deterministic CI.
      - **Suggested priority/order:** P0, Order 31.1.
      - **Risks/notes:** Unbounded IPFS fetch can become a DoS vector.
    - [x] **P9-T4-ST2 Wire IPFS object addressing into attachment/history retrieval paths**
      - **Objective:** Ensure client/relay can resolve objects deterministically with degraded modes.
      - **Concrete actions:** Implement IPFS CID handling; implement fallback to direct transfer where allowed; implement “missing object” semantics.
      - **Dependencies/prerequisites:** `P1-T3`.
      - **Deliverables/artifacts:** Retrieval path integration + tests (`VA-C2`).
      - **Acceptance criteria:** Integration tests show attachments resolve via IPFS when pinned; degraded mode is deterministic when missing.
      - **Suggested priority/order:** P0, Order 31.2.
      - **Risks/notes:** CID canonicalization errors will break interoperability.
  - **Dependencies/prerequisites:** `P1-T1` through `P1-T3`.
  - **Deliverables/artifacts:** IPFS hosting reference implementation (`VA-C1`-`VA-C2`).
  - **Acceptance criteria:** A relay configured with IPFS can publish and retrieve attachments reliably with bounded resource use.
  - **Suggested priority/order:** P0, Order 31.
  - **Risks/notes:** Treat any unbounded resource behavior as release-blocking.

- [x] **[P0][Order 32] P9-T5 Implement hierarchical GossipSub tuning for large servers**
  - **Objective:** Implement the large-server messaging scalability profile (`P2-T1`-`P2-T3`) including topic hierarchy and fanout controls.
  - **Concrete actions:**
    - [x] **P9-T5-ST1 Implement topic hierarchy mapping and subscription policy**
      - **Objective:** Prevent “everyone subscribes to everything” failure modes.
      - **Concrete actions:** Implement topic naming; implement subscription rules by channel membership; implement join/leave churn handling.
      - **Dependencies/prerequisites:** `P2-T1`.
      - **Deliverables/artifacts:** Topic mapping implementation + tests (`VA-C3`).
      - **Acceptance criteria:** Simulated membership changes do not cause message loss or uncontrolled fanout.
      - **Suggested priority/order:** P0, Order 32.1.
      - **Risks/notes:** Topic naming becomes a long-term compatibility surface.
    - [x] **P9-T5-ST2 Implement fanout/batching controls and measure baseline throughput**
      - **Objective:** Achieve predictable throughput under load.
      - **Concrete actions:** Tune deterministic fanout/batching policy params; capture baseline contract metrics; add regression thresholds.
      - **Dependencies/prerequisites:** `P2-T2`, `P4-T1`.
      - **Deliverables/artifacts:** Tuning config + perf harness runs (`VA-C4`, `VA-E1`).
      - **Acceptance criteria:** Documented deterministic baseline throughput/latency contract numbers exist and are reproducible.
      - **Suggested priority/order:** P0, Order 32.2.
      - **Risks/notes:** Avoid tuning that only works in synthetic environments.
  - **Dependencies/prerequisites:** `P2-T1` through `P2-T3`.
  - **Deliverables/artifacts:** Large-server messaging implementation evidence (`VA-C3`-`VA-C4`, `VA-E1`).
  - **Acceptance criteria:** Large-server profile can be enabled and demonstrates deterministic throughput/latency contract improvements in reproducible tests.
  - **Suggested priority/order:** P0, Order 32.
  - **Risks/notes:** Treat uncontrolled resource amplification as a blocker.

- [x] **[P0][Order 33] P9-T6 Implement cascading SFU mesh for 200+ participant voice**
  - **Objective:** Implement deterministic cascading SFU topology contracts defined in `P3-T1`-`P3-T3` with observable QoS policy outputs and stable failure handling.
  - **Concrete actions:**
    - [x] **P9-T6-ST1 Implement SFU-to-SFU federation handshake and stream routing rules**
      - **Objective:** Make multi-SFU calls interoperable and stable.
      - **Concrete actions:** Implement deterministic federation control-plane contract; implement routing decisions; implement join/leave behavior policy; enforce encryption boundaries.
      - **Dependencies/prerequisites:** `P3-T1`, existing SFU baseline.
      - **Deliverables/artifacts:** Federation control-plane implementation + tests (`VA-C5`).
      - **Acceptance criteria:** Deterministic simulation tests show stable routing under join/leave churn.
      - **Suggested priority/order:** P0, Order 33.1.
      - **Risks/notes:** Federation adds complexity; keep control plane minimal.
    - [x] **P9-T6-ST2 Implement QoS telemetry and failure recovery (dropouts, SFU failover)**
      - **Objective:** Prevent “silent” call degradation.
      - **Concrete actions:** Add deterministic QoS telemetry policy outputs; implement failover rules; define deterministic recovery states; add integration tests.
      - **Dependencies/prerequisites:** `P3-T2`, `P4-T2`.
      - **Deliverables/artifacts:** QoS metrics + failover tests (`VA-C6`, `VA-E2`).
      - **Acceptance criteria:** Tests demonstrate deterministic recovery actions and bounded reconnection loops.
      - **Suggested priority/order:** P0, Order 33.2.
      - **Risks/notes:** Failover loops can melt networks; bound retries.
  - **Dependencies/prerequisites:** `P3-T1` through `P3-T3`.
  - **Deliverables/artifacts:** Cascading SFU implementation evidence (`VA-C5`-`VA-C6`, `VA-E2`).
  - **Acceptance criteria:** Deterministic cascade tests are reproducible for 50-participant baseline and 200+ planning profile policy outputs are recorded.
  - **Suggested priority/order:** P0, Order 33.
  - **Risks/notes:** If 200+ is not reachable initially, ship with explicit staged thresholds and deferral.

- [x] **[P0][Order 34] P9-T7 Implement cross-platform profiling harness and optimization loop**
  - **Objective:** Provide repeatable deterministic profiling policy runs for desktop and mobile and convert findings into prioritized fixes.
  - **Concrete actions:** Add profiling policy flags; add deterministic CPU/memory metric capture; define baseline metrics; produce hotspot policy log; land optimizations behind deterministic benchmarks.
  - **Dependencies/prerequisites:** `P4-T1`, `P4-T2`.
  - **Deliverables/artifacts:** Deterministic profiling harness scripts + baseline reports (`VA-E3`).
  - **Acceptance criteria:** A new contributor can reproduce deterministic baseline profiles and see stable metric outputs.
  - **Suggested priority/order:** P0, Order 34.
  - **Risks/notes:** Without repeatability, profiling data is not actionable.

- [x] **[P0][Order 35] P9-T8 Build and run 1000-member stress test campaign**
  - **Objective:** Convert the stress-test contracts (`P5-T1`-`P5-T3`) into executable deterministic campaigns with recorded results and regression thresholds.
  - **Concrete actions:** Implement deterministic load-model generator; implement scenario scripting; run 1000-member deterministic campaigns; record CPU/mem/bw policy metrics; produce pass/fail thresholds.
  - **Dependencies/prerequisites:** `P5-T1` through `P5-T3`, `P9-T5`.
  - **Deliverables/artifacts:** Deterministic stress harness + campaign reports (`VA-E4`).
  - **Acceptance criteria:** Campaigns are reproducible; results are stored as versioned artifacts; thresholds are enforced in CI where feasible.
  - **Suggested priority/order:** P0, Order 35.
  - **Risks/notes:** Synthetic tests must be documented as synthetic; avoid overstating production equivalence.

- [x] **[P0][Order 36] P9-T9 Achieve relay performance target (10k simultaneous clients) with evidence**
  - **Objective:** Implement optimizations and configuration required to support the 10k planning profile per relay and provide reproducible deterministic evidence.
  - **Concrete actions:** Profile relay policy hot paths; optimize deterministic decisions; tune connection-limit policy; add backpressure policy; run deterministic load-profile tests; document recommended sizing.
  - **Dependencies/prerequisites:** `P6-T1` through `P6-T3`, `P9-T7`.
  - **Deliverables/artifacts:** Relay perf policy changes + deterministic load-profile reports (`VA-C7`, `VA-E5`).
  - **Acceptance criteria:** Deterministic load profile meets target assumptions within defined hardware profile model; failure modes are bounded and documented.
  - **Suggested priority/order:** P0, Order 36.
  - **Risks/notes:** Define “10k clients” precisely (connected vs active) and keep evidence aligned.

- [x] **[P0][Order 37] P9-T10 Implement battery optimization policies and validate on mobile**
  - **Objective:** Implement the mobile battery policies defined in `P7-T1`-`P7-T3` and validate against measurable budgets.
  - **Concrete actions:** Add adaptive polling policy; implement wake budgeting; minimize background work policy; measure deterministic energy-impact indicators; document recommended settings.
  - **Dependencies/prerequisites:** `P7-T1` through `P7-T3`, platform build pipeline.
  - **Deliverables/artifacts:** Battery optimization policy code + measurement notes (`VA-C8`, `VA-E6`).
  - **Acceptance criteria:** Deterministic measurements show improvement vs baseline with no delivery-path regressions.
  - **Suggested priority/order:** P1, Order 37.
  - **Risks/notes:** Mobile OS policies are volatile; keep results platform-specific and date-stamped.

- [x] **[P0][Order 38] P9-T11 Publish v0.9 benchmark suite, thresholds, and reproducibility runbook**
  - **Objective:** Make v0.9 performance claims reproducible and reviewable.
  - **Concrete actions:**
    - [x] **P9-T11-ST1 Add deterministic CLI witness scenario `v09-forge` in `cmd/aether` and bind it to the gate evidence flow.**
      - **Objective:** Ensure the `V9-G9` command matrix requirement is backed by an explicit implementation task.
      - **Concrete actions:** Define scenario success/failure output contract; wire scenario dispatch; ensure it validates core v0.9 benchmark/evidence readiness checks without hidden side effects.
      - **Deliverables/artifacts:** Scenario hook specification + implementation plan (`VA-E7`).
      - **Acceptance criteria:** `go run ./cmd/aether --mode=client --scenario=v09-forge` is documented as deterministic and gate-auditable.
      - **Suggested priority/order:** P1, Order 38.1.
    - [x] **P9-T11-ST2 Publish benchmark reproducibility runbook and threshold registry.**
      - **Objective:** Keep performance claims reproducible by independent reviewers.
      - **Concrete actions:** Document perf/stress suite invocation, hardware profiles, threshold rationale, and known limitations.
      - **Deliverables/artifacts:** `docs/v0.9/phase9/perf-runbook.md` + threshold config (`VA-H3`, `VA-E7`).
      - **Acceptance criteria:** A reviewer can reproduce key benchmarks end-to-end with the same command set.
      - **Suggested priority/order:** P1, Order 38.2.
  - **Dependencies/prerequisites:** `P9-T4` through `P9-T10`.
  - **Deliverables/artifacts:** Scenario witness + perf runbook + threshold config (`VA-H3`, `VA-E7`).
  - **Acceptance criteria:** A reviewer can reproduce key benchmarks end-to-end.
  - **Suggested priority/order:** P1, Order 38.
  - **Risks/notes:** Benchmarks without reproducibility are not credible.


---

## H. Suggested Execution Waves and Sequencing

### Wave A - Scope/governance/evidence foundation (V9-G0)
1. P0-T1
2. P0-T2
3. P0-T3
4. P0-T4

### Wave B - IPFS persistent hosting contracts (V9-G1)
5. P1-T1
6. P1-T2
7. P1-T3

### Wave C - Large-server optimization contracts (V9-G2)
8. P2-T1
9. P2-T2
10. P2-T3

### Wave D - Cascading SFU mesh contracts (V9-G3)
11. P3-T1
12. P3-T2
13. P3-T3

### Wave E - Cross-platform profiling/optimization contracts (V9-G4)
14. P4-T1
15. P4-T2
16. P4-T3

### Wave F - Stress-testing contracts (V9-G5)
17. P5-T1
18. P5-T2
19. P5-T3
20. P5-T4

### Wave G - Relay performance/load contracts (V9-G6)
21. P6-T1
22. P6-T2
23. P6-T3
24. P6-T4

### Wave H - Mobile battery optimization contracts (V9-G7)
25. P7-T1
26. P7-T2
27. P7-T3
28. P7-T4

### Wave I - Integrated validation/governance readiness (V9-G8)
29. P8-T1
30. P8-T2
31. P8-T3
32. P8-T4

### Wave J - Shipping work: implementation + perf/stress evidence + release conformance (V9-G9)
33. P9-T1
34. P9-T2
35. P9-T3
36. P9-T4
37. P9-T5
38. P9-T6
39. P9-T7
40. P9-T8
41. P9-T9
42. P9-T10
43. P9-T11

---

## I. Verification Evidence Model and Traceability Expectations

### I.1 Evidence model rules
1. Every task produces at least one named artifact with artifact ID.
2. Every scope item appears in at least one positive-path and one adverse/saturation/recovery scenario.
3. Every gate submission includes explicit pass/fail decision and evidence links.
4. Every compatibility-sensitive delta includes additive and major-path checklist evidence.
5. Every unresolved decision remains explicitly open and linked to revisit gate.

### I.1A Gate command matrix (agent execution minimum)

| Gate | Minimum command set (captured verbatim in evidence) | Expected pass pattern |
|---|---|---|
| `V9-G0` | `go test ./pkg/v09/...` | Exit code `0`; conformance/governance package rows present in output. |
| `V9-G1` | `go test ./pkg/v09/...` and `go test ./tests/e2e/v09/...` | Exit code `0`; evidence links include `S9-01` scenarios. |
| `V9-G2` | `go test ./pkg/v09/...` and `go test ./tests/perf/v09/...` | Exit code `0`; large-server stress/baseline rows attached. |
| `V9-G3` | `go test ./pkg/v09/...` and `go test ./tests/perf/v09/...` | Exit code `0`; cascade failover and recovery cases linked. |
| `V9-G4` | `go test ./tests/perf/v09/...` | Exit code `0`; baseline and post-optimization reports attached. |
| `V9-G5` | `go test ./tests/perf/v09/...` | Exit code `0`; baseline + incremental `+50` campaign outputs (`VA-T7`) attached. |
| `V9-G6` | `go test ./pkg/v09/...` and `go test ./tests/perf/v09/...` | Exit code `0`; overload-priority evidence includes control/media/store-forward/bulk ordering. |
| `V9-G7` | `go test ./pkg/v09/...` and `go test ./tests/e2e/v09/...` | Exit code `0`; battery + wake-failover evidence linked. |
| `V9-G8` | `go test ./pkg/v09/...` + `go test ./tests/e2e/v09/...` + `go test ./tests/perf/v09/...` | Exit code `0`; integrated matrix rows mapped to all `S9-*` items. |
| `V9-G9` | `go test ./...` and `go run ./cmd/aether --mode=client --scenario=v09-forge` | Exit code `0`; release handoff bundle includes full command transcripts. |

If any command returns only `[no test files]`, gate owners must attach alternate evidence and an explicit justification; otherwise the gate remains incomplete.

If a listed command cannot run because a required path/tool/scenario is missing, gate owners must record the failed command output, execute the nearest scope-equivalent substitute command set, create a remediation task with owner + due gate, and keep the gate in conditional-fail state until the canonical command is restored.

### I.2 Traceability mapping: v0.9 scope to tasks/artifacts/acceptance anchors

| Scope Item ID | v0.9 Scope Bullet | Primary Tasks | Validation Artifacts | Acceptance Anchor |
|---|---|---|---|---|
| S9-01 | IPFS integration for persistent file hosting (pinning by server owners) | P1-T1 P1-T2 P1-T3 P9-T4 P9-T11 | VA-I1 VA-I2 VA-I3 VA-I4 VA-I5 VA-I6 VA-C1 VA-C2 VA-E1 VA-E7 | P9-T4 acceptance + P9-T11 reproducibility runbook |
| S9-02 | Large server optimization: hierarchical GossipSub, lazy member loading | P2-T1 P2-T2 P2-T3 P9-T5 P9-T8 P9-T11 | VA-L1 VA-L2 VA-L3 VA-L4 VA-L5 VA-L6 VA-L7 VA-C3 VA-C4 VA-E1 VA-E4 VA-E7 | P9-T5 acceptance + P9-T8 campaign evidence |
| S9-03 | Cascading SFU mesh for 200+ participant voice | P3-T1 P3-T2 P3-T3 P9-T6 P9-T8 P9-T11 | VA-S1 VA-S2 VA-S3 VA-S4 VA-S5 VA-S6 VA-C5 VA-C6 VA-E2 VA-E4 VA-E7 | P9-T6 acceptance + P9-T8 campaign evidence |
| S9-04 | Performance profiling and optimization across all platforms | P4-T1 P4-T2 P4-T3 P9-T7 P9-T11 | VA-P1 VA-P2 VA-P3 VA-P4 VA-P5 VA-P6 VA-E3 VA-E7 | P9-T7 baseline reports + P9-T11 runbook |
| S9-05 | Stress testing: 1000-member server, 50-person voice, latency benchmarks | P5-T1 P5-T2 P5-T3 P5-T4 P9-T8 P9-T11 | VA-T1 VA-T2 VA-T3 VA-T4 VA-T5 VA-T6 VA-T7 VA-E4 VA-E7 | P9-T8 campaign reports + thresholds |
| S9-06 | Relay node performance optimization and load testing | P6-T1 P6-T2 P6-T3 P6-T4 P9-T9 P9-T11 | VA-R1 VA-R2 VA-R3 VA-R4 VA-R5 VA-R6 VA-R7 VA-C7 VA-E5 VA-E7 | P9-T9 load test reports + sizing guidance |
| S9-07 | Battery optimization on mobile (background activity reduction) | P7-T1 P7-T2 P7-T3 P7-T4 P9-T10 P9-T11 | VA-B1 VA-B2 VA-B3 VA-B4 VA-B5 VA-B6 VA-B7 VA-C8 VA-E6 VA-E7 | P9-T10 measurement evidence + policy docs |
| S9-08 | Scale-driven security-mode transitions + sharding guidance for huge interactive channels | P2-T2 P5-T1 P5-T4 P9-T5 P9-T8 P9-T11 | VA-L7 VA-T1 VA-T2 VA-T7 VA-C3 VA-E4 VA-E7 | P9-T5 sharding behavior + P9-T8 stress evidence |

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
| R9-13 | Encryption-method scale limit is under-characterized, leaving unclear downgrade boundaries | High | S9-05, S9-08 | VA-T7 incremental +50 campaigns + security-floor summary | V9-G5 owner |
| R9-14 | Relay adoption slows because hosting has no economic incentives | Medium | S9-06 | Relay onboarding playbooks, explicit non-economic value messaging, and VA-R1/VA-R2 governance packages | V9-G6 owner |
| R9-15 | Operator continuity gap during handoff creates downtime, especially for mobile wake paths | High | S9-06, S9-07 | VA-R7 + VA-B7 operator handover and wake failover plans | V9-G6/V9-G7 owner |
| R9-16 | Governance/legal transition to future consortium lacks documentation, risking inconsistent policy hardening | Medium | All | VA-LG* + VA-X7/VA-X8 policy/legal plans | V9-G8 owner |

---

## K. Open Decisions Tracking

### K.1 Adopted carry-forward defaults from `open_decisions_proposals.md`

- **OD3-02 (ranking tie-break):** provisionally adopted as `relevance -> trust -> recency -> deterministic lexical tie-break`; integrated into large-server/discovery validation in `P2-T2`, `P2-T3`, and `P8-T1`.
- **OD3-03 (RNNoise fallback):** adopted as mandatory fallback posture through v0.9; reflected in cascade and battery/perf interaction checks (`P3-T2`, `P7-T1`, `P8-T1`).
- **OD3-04 (single-indexer default + rotation):** adopted with latency-watch trigger; integrated into discovery/privacy assumptions and revisited if profiling evidence indicates user-harmful slowness (`P4-T1`, `P4-T2`, `P8-T1`).
- **OD4-03 (conservative auto-mod thresholds):** adopted as moderation dependency baseline; enforced through governance and regression checks (`E.4`, `P8-T2`).
- **OD5-01..OD5-05 (bot/webhook/SDK/ecosystem defaults):** adopted as prerequisite baseline and enforced through compatibility/governance validation (`E.5`, `P8-T2`).
- **OD6-01..OD6-03 (discovery hardening, PoW adaptation, sparse-graph trust):** adopted as inherited safety/perf baseline (`E.6`, `P2-T3`, `P4-T2`, `P8-T1`).
- **OD7-01..OD7-04 (replica/chunk/search/relay defaults):** adopted as inherited persistence/relay baseline and verified in relay/load tasks (`E.7`, `P6-T1`-`P6-T4`, `P8-T1`).
- **OD8-01..OD8-05 (thread depth, preview precedence, contrast policy, locale fallback, DTLN policy):** adopted as v0.8 carry-forward constraints for profiling, stress, and battery regressions (`E.8`, `P4-T1`, `P5-T4`, `P7-T1`).
- **OD9-01..OD9-07 (v0.9 defaults):** adopted as planning defaults in v0.9 phase tasks, with final closure requiring explicit V9-G8/V9-G9 evidence signoff.
- **OD3-01 (directory freshness):** accepted with soft TTL 24h, stale grace to 72h, and explicit stale labeling plus next-action guidance in persistence surfaces.
- **OD4-01 (moderation race window):** accepted as first-come-first-served with a deterministic 5-second race window and full audit logging for both actions.
- **OD4-02 (policy rollback horizons):** accepted as stage-based windows (alpha 7d, beta 72h, live 24h) with privileged override up to 7d requiring explicit audit reason.
- **RM-03 (security-floor characterization):** closed with +50 participant campaign evidence and explicit no-silent-downgrade transition messaging tied to `mode_epoch_id` and sharding fallback guidance.

### K.2 Remaining open decisions

| Open Decision ID | Open Question | Status | Owner Role | Revisit Gate | Trigger for Revisit | Handling Rule |
|---|---|---|---|---|---|---|
| None | - | - | - | - | - | - |

Handling rule:
- No remaining open decisions exist for v0.9; any future ambiguity must be reopened with explicit owner, revisit gate, and trigger before it can affect scope.

---

## L. Release-Conformance Checklist for v0.9 Shipping (V9-G9)

v0.9 is considered **shipped** only when the scalability and performance scope is implemented and validated with reproducible evidence (benchmarks, stress campaigns, and documented thresholds).

### L.1 Build and test integrity
- [x] `go test ./...` passes for supported platforms/build tags.
- [x] Each v0.9 capability has reference implementation evidence (`VA-C*`) and a reproducible harness/run (`VA-E*`).
- [x] Any tuning parameters are documented with safe defaults and rollback posture.

### L.2 Reproducible perf/stress evidence (mandatory)
- [x] A perf/stress runbook exists (`P9-T11`) with exact commands, hardware profiles, and expected outputs.
- [x] 1000-member deterministic stress campaigns are executed and reproducible via versioned tests (`P9-T8`).
- [x] +50 increment encryption/security-mode campaigns are executed deterministically to hard-limit characterization with explicit no-silent-downgrade outcomes (`P5-T4` / `VA-T7` / `P9-T8`).
- [x] Relay overload policy for the 10k planning profile is exercised deterministically against saturation thresholds (`P9-T9`).
- [x] Voice scalability coverage (cascading SFU at 50 participants) is executed and recorded deterministically (`P9-T6`).

### L.3 Scope-bullet acceptance
- [x] IPFS hosting contracts are exercised end-to-end in deterministic scenarios with bounded degraded modes when content is missing.
- [x] Hierarchical GossipSub tuning contracts pass deterministic fanout/threshold coverage without correctness regressions.
- [x] Profiling and optimization thresholds are repeatable via deterministic harness checks and linked metrics.
- [x] Battery optimization policies pass deterministic improvement/guardrail checks without delivery-path regressions.

### L.4 Operational readiness
- [x] Sizing guidance exists for relays and SFU components (CPU/mem/bw and storage).
- [x] Monitoring and alert thresholds are documented for the perf-critical components.
- [x] Known limitations are explicitly listed (what scale is proven, what is aspirational).

---

## M. Definition of Done for v0.9

v0.9 is complete when:
1. All v0.9 roadmap bullets are implemented in deterministic reference components (IPFS hosting, large-server messaging optimization, cascading SFU, profiling+optimization, stress/load policy testing, relay perf policy, battery optimization).
2. Performance and scaling claims in this slice are backed by reproducible deterministic harnesses and versioned reports.
3. Configuration defaults are safe, documented, and rollbackable.
4. Release notes and docs accurately describe validated scale limits and known constraints.
