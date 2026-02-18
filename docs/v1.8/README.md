# v1.8 (F18) Execution Artifact Summary

This directory captures the implementation, validation, and governance artifacts for the v1.8 (F18) discovery and indexer delivery.

- `pkg/v18/directory`, `pkg/v18/indexer`, `pkg/v18/discoveryclient`, and `pkg/v18/ui` for signed listing, indexing, merge, and join-funnel runtime contracts.
- E2E/perf coverage under `tests/e2e/v18` and `tests/perf/v18` for signature checks, trust warnings, deduplication, and abuse-path probes.
- Container and script scenarios (`containers/v1.8` and `scripts/v18-discovery-scenarios.sh`) that exercise discovery/indexer behavior and emit deterministic manifests.
- Governance docs for phases 0-5, including scope lock, traceability, UX contract, Podman scenario runbook, and v19 connectivity/QoL specification package.
