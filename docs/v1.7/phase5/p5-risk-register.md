# Phase 5 — Risk Register

| ID | Risk | Mitigation | Exit Criterion |
|---|---|---|---|
| R17-1 | Inconsistent moderation enforcement | Signed deterministic event model across `pkg/v17/moderation` | Convergence tests pass |
| R17-2 | Audit tampering concerns | Append-only verified log model plus role filters | Integrity tests pass |
| R17-3 | Client non-compliance ambiguity | Explicit enforcement status signaling contract with trust warnings | UX/runtime tests pass |
| R17-4 | Relay boundary regression | Dedicated e2e and Podman regression scenarios plus `pkg/v11/relaypolicy` checks | Regression tests (script + docs) pass |
