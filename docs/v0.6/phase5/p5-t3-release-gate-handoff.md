# Phase 5 · P5-T3 Release Gate & Handoff

## Purpose
Deliver the V6-G6 release-conformance checklist, deferral register, and residual risk handoff that make `VA-X3` auditable without implying runtime completion of downstream items.

## Contract
- The V6-G6 checklist enumerates every VA artifact plus scope bullet and ties each entry to a pass/fail rule, evidence anchor, and owner per the `verification-evidence-matrix` and the gate flow in Section 5 of `TODO_v06.md`.
- The handoff dossier captures residual risks, recovery guidance, and deferral targets (including `v0.7+` compliance/QA extensions) without implying those items are complete.
- Release messaging reiterates planned-vs-implemented discipline so the downstream team knows what remains in play and where to pick up the deterministic reasoning.

## Scenario-pack coverage
| Coverage area | Evidence requirement | Reference |
|---|---|---|
| Positive, adverse, abuse, recovery scenarios | Each scenario links to the VA reason-class triads and Section 9 release checklist entries | `docs/v0.6/phase5/p5-t1-cross-feature-scenario-pack.md#scenario-pack-table` |
| Release dossier cross-check | Scenario pack ties back to `p5-t3` Section 9 mapping and Open Decision register | same doc |

## Section 9 checklist mapping
| Section 9 item | Status | Gate reference | Evidence anchor |
|---|---|---|---|
| All ten v0.6 bullets are mapped to tasks and artifacts | Pass | V6-G5 | `docs/v0.6/phase0/p0-t1-scope-contract.md` |
| v0.6 framed as hardening/reliability | Pass | V6-G6 | `TODO_v06.md` Section 3 narrative |
| Discovery/search/explore/preview phrased as hardening | Pass | V6-G4 | `docs/v0.6/phase0/p0-t3-verification-evidence-matrix.md` |
| Optional filters remain optional | Pass | V6-G4 | `docs/v0.6/phase4/p4-reputation-report-filter-reliability.md` |
| Anti-abuse/report-routing semantics deterministic/tested | Pass | V6-G5 | `docs/v0.6/phase3/p3-anti-abuse-hardening-contracts.md`, `docs/v0.6/phase4/p4-reputation-report-filter-reliability.md` |
| No v0.7+ archive/history scope imported | Pass | V6-G5 | `TODO_v06.md` Section 3 exclusions |
| Compatibility/governance/open-decision checks complete | Pass | V6-G5 | `docs/v0.6/phase5/p5-t2-conformance-review.md` |
| Planned-vs-implemented distinction explicit | Pass | V6-G6 | `TODO_v06.md` front matter |

## V6-G6 checklist
| Checklist item | Pass/fail rule | Evidence anchor |
|---|---|---|
| `VA-D*`..`VA-R*` artifacts cite reason-coded positive/negative/recovery scenarios and release-trace links | Pass when every artifact above lists the triad of paths and cites the release dossier; fail when any path is missing or unlinked | `docs/v0.6/phase0/p0-t3-verification-evidence-matrix.md` |
| Integrated `VA-X1`, `VA-X2`, `VA-X3` packages align with the reason taxonomy and open-decision register | Pass when scenario, conformance, and release packages document the continuity described in Section 1.3; fail otherwise | `docs/v0.6/phase5/p5-t1-cross-feature-scenario-pack.md`, `docs/v0.6/phase5/p5-t2-conformance-review.md`, `docs/v0.6/phase5/p5-t3-release-gate-handoff.md` |

## Deferral register (v0.7+ carry-forward)
| Item | Reason for deferral | Target gate | Owner |
|---|---|---|---|
| V7 PoW telemetry instrumentation | Enhanced instrumentation screener exceeds v0.6 hardening scope; carry to v0.7 to avoid mixing evidence | V7 planning gate | Anti-abuse lead |
| V7 full archive/replay validation | Archive/history scope is explicitly out-of-scope; defer to keep v0.6 reliability focus intact | V7 planning gate | Release lead |

## Residual risk handoff
| Residual risk ID | Description | Severity | Mitigation / next step | Owner |
|---|---|---|---|---|
| R6-01 | Discovery freshness TTL disagreement across clients | Medium | Reference `VA-D1` TTL guardrails and cross-feature scenario pack for TTL tuning evidence | Discovery lead |
| R6-02 | Filter policy drift under SecurityMode gating | Medium | Keep `VA-R3` reason-state and include policy review in `VA-X2` governance audit | Filters lead |

These sections keep V6-G6 signoff auditable and frame what downstream teams continue to own without implying runtime implementation completion.
