# Ownership Matrix and ADR Template (Phase 1)

## Ownership Matrix
| Role | Primary Owner | Backup Owner | Notes |
|---|---|---|---|
| Protocol | Protocol Lead | Crypto Engineer | Oversees trait constraints and field numbering decisions. |
| Networking | Networking Engineer | DevOps Engineer | Owns libp2p host setup, DHT, relay flows. |
| Crypto | Crypto Engineer | Protocol Lead | Manages identity keys, MLS integration, and governance for cryptography defaults. |
| UI (Gio shell) | Client Engineer | QA Engineer | Maintains first-contact shell and journey instrumentation. |
| Ops (Build/Test) | Engineering Manager | DevOps Engineer | Coordinates pipelines, Podman builds, and Buf checks. |
| QA | QA Lead | Protocol Lead | Defines acceptance/test charters and first-contact validation. |

## Phase 1 P0 Task Owner Map
| Task | Primary Owner | Backup Owner | Acceptance Coverage |
|---|---|---|---|
| P1-T1 Freeze v0.1 scope contract | Tech Lead | Protocol Lead | Scope and exclusions are version-bounded and auditable. |
| P1-T2 Build protocol constraints checklist | Protocol Lead | Crypto Engineer | Compatibility/governance guardrails are explicit and reviewable. |
| P1-T3 Define ownership and decision cadence | Engineering Manager | Tech Lead | Every P0 Phase-1 task has accountable ownership and escalation route. |
| P1-T4 Define acceptance test charter | QA Lead | QA Engineer | First-contact journey has reusable preconditions/steps/evidence model. |

## Decision Cadence and Escalation
- Weekly decision sync anchoring Gate G0-G2 planning updates.
- Escalation path: Task owner → Phase lead (engineering manager) → Steering committee (AEP board) for breaking or scope-risk decisions.
- Merge/review policy: All critical path PRs require two owners (primary + backup) plus QA approval; release merges must reference the protocol constraints checklist.

## ADR Template
1. **Title:** Clear v0.1 decision name.
2. **Status:** Draft/Proposed/Accepted/Deprecated.
3. **Context:** Why this decision matters for Phase 1 (trace back to scope contract or acceptance criteria). [`docs/v0.1/phase1/scope-contract.md:1`](docs/v0.1/phase1/scope-contract.md:1)
4. **Decision:** Describe the selected option, mention additive evolution implications (no renumber). [`docs/v0.1/phase1/protocol-constraints.md:1`](docs/v0.1/phase1/protocol-constraints.md:1)
5. **Consequences:** Explain impact on tasks, owners, and constraints.
6. **Next Steps:** Link to follow-on actions (e.g., Buf config, QA verification, Podman runbooks). 

## Evidence Placeholders
- Ownership matrix reviewed by stakeholders and linked from next release notes. 
- ADR entries to be stored under `docs/adr/v0.1/`. 
