# Phase 5 — As-Built Conformance

This artifact maps implemented behavior to the v16 F17 specification. The deterministic engine in `pkg/v17/moderation` follows the rejection reasons and replay semantics described in the spec, while `pkg/v17/audit` exposes append-only visibility proofs.

Performance and adversarial scenarios are validated via the tests/e2e/perf suites and the `scripts/v17-moderation-scenarios.sh` manifest (see `artifacts/generated/v17-moderation-scenarios/result-manifest.json`).

Relay boundary evidence is recorded through `pkg/v11/relaypolicy` validation failures and documented in the Podman scenario runbook.
