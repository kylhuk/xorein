# Phase 6 · P6-T3 Release Gate & Handoff

## Purpose
Deliver the final release-conformance checklist, evidence links, and deferral register that make V5-G7 auditable without implying runtime implementation completion.

## Contract
- The V5-G7 checklist enumerates all scope items, artifacts, and risk mitigations from Section 6 and Section 9 of `TODO_v05.md`, pairing each with pass/fail status and evidence references per P6-T3-ST1.
- The execution handoff backlog captures deferred work, rationale, and target roadmap bands (including any v0.6+ carry-overs) without claiming those items are complete, as required by P6-T3-ST2.
- Residual risks link back to the register in Section 9; any `High` severity item includes mitigation ownership and a short-term verification plan so downstream teams can continue the governance trail.

## V5-G7 checklist
| Checklist item | Pass/fail rule | Evidence anchor |
|---|---|---|
| All VA-B*..VA-W* artifacts cite deterministic positive/negative/recovery scenarios and reason codes, and each is referenced by the release gate dossier | Pass when every artifact row in the Phase 0 matrix covers the triad of paths and has a release doc trace link; fail if any entry lacks coverage or a trace. | `docs/v0.5/phase0/p0-t3-verification-evidence-matrix.md#artifact-matrix-by-scope` |
| Integrated validation (`VA-X1`), compatibility/governance (`VA-X2`), and release-conformance (`VA-X3`) evidence packages meet the V5-G7 pass rules | Pass when scenario, conformance, and checklist packages explicitly document the reason-code continuity described in the Phase 0 matrix; fail if any package is incomplete. | `docs/v0.5/phase6/p6-t1-cross-feature-scenario-pack.md#deterministic-contract-pack`, `docs/v0.5/phase6/p6-t2-conformance-review.md#purpose`, `docs/v0.5/phase6/p6-t3-release-gate-handoff.md#deferral-register-v06-carry-forward` |
| Residual-risk handoff (`R5-*`) table exists with owner, mitigation, and proof-of-review for each High severity item | Pass when the residual risk table below enumerates each critical risk with assigned owner and mitigation status; fail when a High risk remains unassigned or undocumented. | `docs/v0.5/phase6/p6-t3-release-gate-handoff.md#residual-risk-handoff` |

## Deferral register (v0.6+ carry-forward)
| Item | Reason for deferral | Target gate | Owner | Evidence anchor | Notes |
|---|---|---|---|---|---|
| P7-T1 Multi-region resilient connection orchestration | Multi-region telemetry/G6 instrumentation exceeds v0.5 scope; carry to v0.6 to avoid mixing runtime commitments | V0.6 planning gate | Connectivity lead | TODO_v05 entries for P7 | Keep deterministic handshake contract intact while deferring operational telemetry/template expansion. |
| P7-T3 Admin UI automation for governance controls | UI automation requires frontend stabilization not covered in v0.5 delivery; postpone to downstream wave | V0.6 planning gate | UX lead | TODO_v05 P7 section | Document dependencies on VA-X3 checklist and include timeline for future automation evidence. |
| OD5-01 Extended bot event disconnect guarantee | Open decision remains unresolved; preserve deterministic baseline while upgrading after v0.5 | V5-G7 revisit | Protocol lead | `TODO_v05.md` Section 10 | Keep as explicit open decision to avoid misrepresenting implementation. |

## Residual risk handoff
| Residual risk ID | Description | Severity | Mitigation / next step | Owner | Evidence anchor |
|---|---|---|---|---|---|
| R5-01 | Bot event ordering ambiguity creates divergent automation outcomes | High | Maintain canonical ordering keys and replay semantics in VA-B2; include scenario-level assertions in `VA-X1`. | V5-G1 owner | `TODO_v05.md` Section 9 risk register |
| R5-09 | Gateway session resume ambiguity causes dropped or duplicated events | High | Confirm heartbeat/reconnect semantics and reason codes in `VA-D2`; include fail/recovery scenarios in the integrated pack. | V5-G3 owner | `TODO_v05.md` Section 9 risk register |
| R5-15 | Webhook replay and duplicate submissions inject unintended messages | High | Stress idempotency/retry behavior in `VA-W3` and reference the replay window in scenario pack; include reason-coded audit logs for duplicates. | V5-G5 owner | `TODO_v05.md` Section 9 risk register |
| R5-18 | Protocol compatibility controls inconsistently applied | High | Require P6-T2 compatibility checklist plus traceable documentation; escalate any deviation before gate exit. | V5-G6 owner | `TODO_v05.md` Section 9 risk register |

These sections keep the V5-G7 signoff auditable, document what is deferred, and hand off residual risks clearly to downstream execution teams.
