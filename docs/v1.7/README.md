# v1.7 (F17) Execution Artifact Summary

This directory captures the implementation, validation, and governance artifacts for the v1.7 (F17) moderation delivery. Key deliverables include:

- `pkg/v17/moderation`, `pkg/v17/modsync`, `pkg/v17/audit`, and `pkg/v17/ui` for deterministic contract logic, replication, audit visibility, and enforcement status signaling.
- E2E/perf coverage under `tests/e2e/v17` and `tests/perf/v17` proving adversarial scenarios, partition convergence, and step pacing.
- Container + script scenarios (`containers/v1.7` and `scripts/v17-moderation-scenarios.sh`) that exercise the relay deployment and capture deterministic manifests.
- Governance docs for phases 0–5, including scope, traceability, UX contracts, Podman runbook, v18 spec plan, and Phase 5 evidence/risk packages.
