# F20 Release Hardening Spec

This spec describes the Phase 4 production hardening expectations for the v20 rollout:
- Attack surface reduction mandates TLS1.3+ and deterministic handshake order.
- Podman operators must have rollback scripts referencing `containers/v1.9/deployment-runbook.md`.
- Hardening checklists include relay no-data regression assertions already captured in v19.
