# F13 Acceptance Matrix (v1.3 planning)

## Status
Planning artifact only.

## Matrix
| Requirement | Validation strategy | Evidence target |
|---|---|---|
| Space lifecycle transitions are deterministic | Unit tests for valid/invalid transitions | EV-v13-G4-001 |
| Founder/admin defaults enforced | Integration tests for create/update role flows | EV-v13-G4-002 |
| Channel send/receive and failure states deterministic | E2E chat scenario suite (positive/negative/recovery) | EV-v13-G4-003 |
| Read markers preserved across reconnect | E2E reconnect and replay validation | EV-v13-G4-004 |
| Additive wire compatibility maintained | `buf lint`, `buf breaking` | EV-v13-G1-001 |
| QoL effort reduction objective met | Perf test comparing baseline and target step count | EV-v13-G4-005 |

## Coverage obligations
- Positive paths, negative paths, degraded paths, and recovery paths are all mandatory.
- No ambiguous terminal user state is permitted in closure evidence.

## Planned vs implemented
- This matrix defines v1.3 closure expectations and does not assert implementation completion.
