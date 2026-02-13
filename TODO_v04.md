# TODO_v04.md

> Status: Planning artifact only. No implementation completion is claimed in this document.
>
> Authoritative v0.4 scope source: `aether-v3.md` roadmap bullets under **v0.4.0 — Dominion** after Addendum A pull-forward alignment.
>
> Inputs used for sequencing and dependency posture: `TODO_v01.md`, `TODO_v02.md`, `TODO_v03.md`, and `AGENTS.md`.
>
> Guardrails that are mandatory throughout this plan:
> - Repository is documentation-only; maintain strict planned-vs-implemented separation.
> - Protocol-first is non-negotiable: protocol/spec contract is the product; UI is a consumer.
> - Network model invariant: single binary with `--mode=client|relay|bootstrap`; no privileged node classes.
> - Compatibility invariant: protobuf minor evolution is additive-only.
> - Breaking protocol behavior requires major-path governance: new multistream IDs, downgrade negotiation evidence, AEP path, and multi-implementation validation.
> - Open decisions remain unresolved unless source documents explicitly resolve them.

---

## Stack Alignment Constraints (Parent Recommendation, Planning-Level)

- This section is recommendation-only planning guidance and does not claim implementation completion.
- Control plane default: libp2p secure channels use `Noise_XX_25519_ChaChaPoly_SHA256` as the single supported suite; QUIC is preferred for reliable multiplexed streams, and this plan must not imply TCP-only operation.
- Media plane default: ICE (STUN/TURN), SRTP hop-by-hop, SFrame true media E2EE, and browser encoded-transform/insertable-streams integration where browser media clients apply.
- Key-management baseline carried forward: X3DH + Double Ratchet for DMs; MLS for group key agreement; historical Sender Keys references are compatibility/migration context only.
- Crypto defaults carried forward: SFrame AES-GCM full-tag default (for example `AES_128_GCM_SHA256_128` intent), avoid short tags unless explicitly justified; messaging AEAD baseline `ChaCha20-Poly1305` with optional AES-GCM negotiation; Noise suite fixed as above; SRTP baseline unchanged.
- Latency/resilience guidance carried forward for dependent realtime behavior: race direct ICE and relay/TURN in parallel, continuous path probing with seamless migration, RTT-aware multi-region relay/SFU selection with warm standby, dynamic topology switching (P2P 1:1, mesh small groups, SFU larger groups) with no SFU transcoding, and background resilience controls.

## 1. v0.4 Objective and Measurable Success Outcomes

### 1.1 Objective
Deliver **v0.4 Dominion** as a protocol-first execution plan focused on **advanced moderation and governance hardening** over v0.1–v0.3 baselines by defining:
- Full custom roles and deterministic channel-override hardening.
- Permission/role governance contracts with conflict-safe evaluation.
- Moderation policy versioning, migration, and rollback semantics.
- Auto-moderation hooks (rate/keyword/extensible policy pipeline).
- Signed audit-log expansion with policy traceability for authorized roles.

### 1.2 Measurable Success Outcomes
1. Custom role lifecycle and precedence behavior are deterministic under concurrent updates.
2. Channel permission override merge rules are deterministic and conflict-testable.
3. Moderation policy versions are immutable, migratable, and rollback-capable with explicit compatibility rules.
4. Auto-moderation hooks are deterministic for trigger, action, bypass, and failure behavior.
5. Signed audit entries include policy-version and enforcement-trace linkage.
6. Moderation semantics remain compatible with decentralized enforcement (signed events honored by compliant clients; no authoritative moderator node assumptions).
7. Compatibility/governance conformance controls and evidence schema are complete and gate-auditable.
8. Release handoff package preserves planned-only wording and explicit deferrals.

### 1.3 QoL integration contract for v0.4 governance surfaces (planning-level)

- **Deterministic reason taxonomy hardening:** moderation-policy, permission, and enforcement outcomes exposed to users map to stable reason classes shared with audit and diagnostics.
  - **Acceptance criterion:** equivalent deny/enforce/failure classes produce consistent reason-class outputs across role, policy, and auto-mod paths.
  - **Verification evidence:** `V4-G5` conformance bundle includes reason-taxonomy coverage table linked to `VA-M*` and `VA-A*` artifacts.
- **No-limbo enforcement for governance actions:** moderation and policy workflows that affect user-visible state must always present current state and next permitted action.
  - **Verification evidence:** cross-feature scenarios include denied-action and appeal/recovery pathways with deterministic state transitions.

---

## 2. Exact Scope Derivation from `aether-v3.md` for v0.4 Only

The following roadmap bullets in `aether-v3.md` define v0.4 scope and are treated as exact inclusions:

1. Permission bitmask system (server-level + channel overrides)
2. Role CRUD and role property contracts
3. Full custom roles + deterministic channel permission overrides
4. Advanced moderation policy (versioning, migration rules, rollback semantics)
5. Auto-moderation hooks (rate limits, keyword filters, extensible policy pipeline)
6. Audit-log expansion (signed entries + policy traceability for authorized roles)

No additional capability outside these six bullets is promoted into v0.4 in this plan.

---

## 3. Explicit Out-of-Scope and Anti-Scope-Creep Boundaries

### 3.1 Already introduced in earlier versions (not first introduction in v0.4)
- Baseline RBAC Owner/Admin/Moderator/Member, baseline moderation actions (redaction/delete, timeout, ban), and channel slow mode (v0.2).
- Directory publishing/browse, invite/request-to-join, optional indexer reference, and signed response verification (v0.3).

### 3.2 Deferred to v0.5+
- Bot API, Discord compatibility shim, slash commands, emoji/reaction platform.

### 3.3 Deferred to v0.6+
- Discovery/moderation/anti-abuse hardening and scale-reliability tracks.

### 3.4 Deferred to v0.7+
- Deep history/search/push relay architecture expansion.

### 3.5 Anti-scope rules
1. v0.4 expands governance depth; it does not re-introduce baseline capabilities already pulled into v0.2/v0.3.
2. Any incompatible protocol behavior must follow major-version governance path.
3. Open decisions remain open and are never represented as settled architecture.

---

## 4. Entry Prerequisites (Assumed Completed)

### 4.1 v0.2 baseline prerequisites
- Baseline RBAC and moderation-event enforcement semantics.
- Slow-mode baseline behavior and role-based mention controls.

### 4.2 v0.3 baseline prerequisites
- Directory/admission baselines and optional indexer non-authoritative posture.

### 4.3 Dependency handling rule
- Missing prerequisites are blocking dependencies and are carried back to prior-version backlog.

---

## 5. Gate Model and Flow for v0.4

### 5.1 Gate Definitions

| Gate | Name | Entry Criteria | Exit Criteria |
|---|---|---|---|
| V4-G0 | Scope and guardrails lock | v0.4 planning initiated | Scope lock, exclusions, prerequisites, and verification schema approved |
| V4-G1 | Role/permission hardening freeze | V4-G0 passed | Custom roles + override semantics + conflict resolution fully specified |
| V4-G2 | Moderation policy versioning freeze | V4-G1 passed | Policy versioning/migration/rollback contracts fully specified |
| V4-G3 | Auto-moderation contract freeze | V4-G2 passed | Trigger/action/bypass/failure semantics fully specified |
| V4-G4 | Audit traceability freeze | V4-G3 passed | Signed audit expansion and policy-trace linkage fully specified |
| V4-G5 | Integrated validation and governance readiness | V4-G4 passed | Cross-feature scenarios, compatibility checks, and open-decision discipline complete |
| V4-G6 | Release conformance and handoff | V4-G5 passed | Full traceability closure and execution handoff dossier approved |

### 5.2 Gate Flow Diagram

```mermaid
graph LR
  G0[V4 G0 Scope and Guardrails Lock]
  G1[V4 G1 Role and Permission Hardening]
  G2[V4 G2 Moderation Policy Versioning]
  G3[V4 G3 Auto-Moderation Contract Freeze]
  G4[V4 G4 Audit Traceability Freeze]
  G5[V4 G5 Integrated Validation Readiness]
  G6[V4 G6 Release Conformance and Handoff]

  G0 --> G1 --> G2 --> G3 --> G4 --> G5 --> G6
```

---

## 6. Detailed v0.4 Execution Plan by Phase

Validation artifacts:
- `VA-R*` role/permission hardening
- `VA-M*` moderation policy + auto-mod
- `VA-A*` audit traceability
- `VA-X*` cross-feature conformance

### Phase 0 (V4-G0): Scope, compatibility, and evidence controls
- [ ] **P0-T1** Scope trace mapping for all six v0.4 bullets.
- [ ] **P0-T2** Additive protobuf + major-change trigger checklist pack.
- [ ] **P0-T3** Evidence schema and gate pass/fail template.

### Phase 1 (V4-G1): Custom roles and override hardening
- [ ] **P1-T1** Custom-role CRUD/order conflict-resolution contract (`VA-R1`).
- [ ] **P1-T2** Permission-evaluation and channel-override merge contract (`VA-R2`).
- [ ] **P1-T3** Concurrency and stale-state reconciliation tests for authorization (`VA-R3`).

### Phase 2 (V4-G2): Moderation policy versioning governance
- [ ] **P2-T1** Policy version schema + immutability/integrity rules (`VA-M1`).
- [ ] **P2-T2** Policy migration/compatibility contract (`VA-M2`).
- [ ] **P2-T3** Policy rollback contract with deterministic fallback behavior (`VA-M3`).

### Phase 3 (V4-G3): Auto-moderation hook contracts
- [ ] **P3-T1** Trigger model (rate/keyword/extensible hooks) and precedence rules (`VA-M4`).
- [ ] **P3-T2** Action model (warn/quarantine/block/escalate) with reason taxonomy (`VA-M5`).
- [ ] **P3-T3** Bypass/appeal/failure semantics and deterministic recovery behavior (`VA-M6`).

### Phase 4 (V4-G4): Signed audit-log expansion and policy traceability
- [ ] **P4-T1** Signed audit entry expansion with policy-version references (`VA-A1`).
- [ ] **P4-T2** Authorized visibility and deterministic query behavior (`VA-A2`).
- [ ] **P4-T3** Coverage matrix for policy + auto-mod + manual moderation actions (`VA-A3`).

### Phase 5 (V4-G5 → V4-G6): Integrated validation and release handoff
- [ ] **P5-T1** Cross-feature scenario pack (positive/negative/recovery) (`VA-X1`).
- [ ] **P5-T2** Compatibility/governance/open-decision conformance audit (`VA-X2`).
- [ ] **P5-T3** Final release checklist + execution handoff dossier (`VA-X3`).

---

## 7. Traceability Mapping

| Scope Item ID | v0.4 Scope Bullet | Primary Tasks | Validation Artifacts |
|---|---|---|---|
| S4-01 | Permission bitmask + channel overrides | P1-T2, P1-T3 | VA-R2, VA-R3, VA-X1 |
| S4-02 | Role CRUD + properties | P1-T1 | VA-R1, VA-X1 |
| S4-03 | Full custom roles + deterministic overrides | P1-T1, P1-T2 | VA-R1, VA-R2, VA-X1 |
| S4-04 | Moderation policy versioning/migration/rollback | P2-T1, P2-T2, P2-T3 | VA-M1, VA-M2, VA-M3, VA-X1 |
| S4-05 | Auto-moderation hooks | P3-T1, P3-T2, P3-T3 | VA-M4, VA-M5, VA-M6, VA-X1 |
| S4-06 | Audit-log expansion + policy traceability | P4-T1, P4-T2, P4-T3 | VA-A1, VA-A2, VA-A3, VA-X1 |

---

## 8. Open Decisions (Must Remain Unresolved)

| Decision ID | Open Question | Status | Revisit Gate |
|---|---|---|---|
| OD4-01 | Final default precedence for simultaneous auto-mod and manual moderator actions under race conditions. | Open | V4-G5 |
| OD4-02 | Default rollback horizon for policy versions under prolonged partition/rejoin cycles. | Open | V4-G5 |
| OD4-03 | Baseline false-positive handling thresholds for keyword/rate auto-mod profiles. | Open | V4-G5 |

Handling rule: open decisions remain `Open` and are not presented as settled architecture.

---

## 9. Release-Conformance Checklist (V4-G6)

- [ ] All six v0.4 scope bullets are mapped to tasks and artifacts.
- [ ] No earlier-version baseline capability is re-introduced as first-introduction scope in v0.4.
- [ ] No v0.6 hardening/scaling scope is imported.
- [ ] Role/override, policy versioning, and auto-mod contracts are deterministic and test-mapped.
- [ ] Audit traceability includes signed policy linkage and authorized visibility rules.
- [ ] Compatibility/governance/open-decision checks are complete.
- [ ] Planned-vs-implemented distinction remains explicit.

---

## 10. Definition of Done for v0.4 Planning Artifact

This artifact is complete when it captures advanced moderation/governance scope exactly, preserves protocol-first constraints, and provides gate-testable contracts and handoff evidence without implementation claims.
