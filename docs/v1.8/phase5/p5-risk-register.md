# Phase 5 - Risk Register

| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R18-1 | Indexer trust abuse | Enforce signature validation, deduplicate by NodeID, and surface warnings | All trust warning tests pass |
| R18-2 | Discovery privacy leakage | Query path and warning reason minimalism + explicit allowlist | No unscoped metadata is surfaced in warning payloads |
| R18-3 | Join-funnel abuse and replay | Deterministic stage transition checks and warning retention | Abuse-path tests pass |
| R18-4 | Relay boundary regression during discovery rollout | Dedicated relay-mode regression checks in e2e and scripted scenarios | Relay policy checks pass |
