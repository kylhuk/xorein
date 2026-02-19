# v2.4 Phase 0 Gate Ownership

Phase 0 documents the gate owners, reviewers, and exit criteria for Promotion Gates G0–G9. Each gate is planning-only until the Phase that executes the work, but this matrix captures the intended accountability model and the deterministic exit signal required for sign-off.

| Gate | Owner Role | Reviewer Role | Exit Criteria (planning tone) |
| --- | --- | --- | --- |
| `G0` Scope & Traceability | Protocol Governance Lead | Architecture Reviewer | Scope lock + traceability matrix + RACI documented; evidence `EV-v24-G0-001` through `EV-v24-G0-004` available. |
| `G1` Local API Spec | Local API Product Lead | Security Reviewer | Local API evolution policy + spec drafts complete; compatibility taxonomy logged; evidence `EV-v24-G1-001`–`EV-v24-G1-003`. |
| `G2` Daemon Runtime | Runtime Engineering Lead | Operational Security Reviewer | Daemon API scaffolding + lifecycle docs approved; audit log policy enforced; placeholder evidence `EV-v24-G2-001`. |
| `G3` harmolyn Attach | UI Engineering Lead | Usability Reviewer | Attach workflows documented/tested; UX failure taxonomy validated; evidence `EV-v24-G3-001`. |
| `G4` Security Matrix | Security Lead | Threat Modeling Reviewer | Authz/replay/injection tests documented; security test matrix approved; evidence `EV-v24-G4-001`. |
| `G5` Multi-client & Recovery | Systems Reliability Lead | QA Reviewer | Multi-client scenarios + recovery scripts validated; evidence `EV-v24-G5-001`. |
| `G6` Harness Scenarios | Integration Lead | QA Reviewer | Podman + desktop harness scripts defined; evidence `EV-v24-G6-001`. |
| `G7` F25 Spec | Spec Lead | Product Governance Reviewer | F25 blob/asset spec + acceptance matrix complete; evidence `EV-v24-G7-001`. |
| `G8` Evidence Bundle | Evidence Steward | Program Review Board | Evidence index + command outputs compiled; `EV-v24-G8-001` entry ready. |
| `G9` No UI Dependencies | CI Owner | Security Reviewer | CI boundary gates pass; audit records showing `cmd/xorein` free of UI deps; evidence `EV-v24-G9-001`. |

## Notes
- Each owner/reviewer pair is responsible for producing the evidence placeholder listed in the exit criteria before the gate can be marked passed.
- `P0-T1` captures the first four artifacts (`p0-scope-lock`, `p0-traceability-matrix`, `p0-gate-ownership`, `p0-local-api-evolution`) required to show `G0` readiness.
