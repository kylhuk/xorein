# Phase 5 - Risk register (v11 closure)

## Purpose
Track residual and post-promotion dependencies for v11. Gate-blocking risks remain `open` only while unresolved; mitigated items stay listed for audit traceability.

## Risk table
| RiskID | Description | Severity | Mitigation | Owner role | Status |
|---|---|---|---|---|---|
| R11-01 | Buf compatibility checks currently rely on a local CLI install; future environments may regress if `buf` is absent. | medium | Add deterministic buf provisioning in CI/automation (or containerized invocation), and keep EV-v11-G1-001/002 reproducible. | Protocol Lead | mitigated |
| R11-02 | `make check-full` previously failed at `gosec` scan; latest rerun passes, but scan regressions remain a release risk without routine monitoring. | medium | Keep security scan evidence fresh in CI and treat any future non-zero `gosec` result as an immediate gate blocker requiring triage. | Security Lead | mitigated |
| R11-03 | Deferral governance can regress if the no-active-deferrals record is not kept visible during later updates. | low | Keep `docs/v1.1/phase5/p5-deferral-register.md` linked from phase5 closure docs and update immediately if any unavoidable deferral is approved. | Plan Lead | mitigated |

## Planned vs implemented
- **Planned:** Continue updating this table for any post-v11 release-impacting risk.
- **Implemented:** All v11 gate-blocking risks in this register are mitigated in the current closure snapshot.
