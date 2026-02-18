# v1.2 Phase 5 - Risk Register Closure

## Status
Risk review for v1.2 closure.

| ID | Risk | Mitigation implemented | Exit criterion | Status |
|---|---|---|---|---|
| R12-1 | Ambiguous identity uniqueness semantics | Deterministic identity ID + duplicate detection in `pkg/v12/identity` and tests | Property and unit checks pass | closed |
| R12-2 | Weak backup security defaults | Argon2id + AES-GCM envelope and tamper checks in `pkg/v12/backup` | Security and corruption tests pass | closed |
| R12-3 | Users expect password reset | Mandatory onboarding warning + no-reset flow contract in shell state methods | UX tests verify warning gate | closed |
| R12-4 | Relay boundary regression during rollout | Dedicated e2e relay policy regression scenario | Regression tests pass | closed |

## Planned vs implemented
- Mitigations are implemented in code.
- Verification closure is complete with attached phase5 command evidence.
