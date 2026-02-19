# Phase 0 scope lock (G0)

This artifact freezes the planning-versus-implementation contract for v22 history-plane work without claiming completion. It restates the boundaries that must hold while downstream teams convert scope into code, tests, and evidence.

## Locked commitments

- **Archivist capability is opt-in and non-privileged.** The capability flag/config lives in client/runtime policy; no peers are forced to become archivists. Archivists store only ciphertext segments, cannot veto protocol state transitions, and expose only deterministic refusal reasons (`NO_ARCHIVIST_AVAILABLE`, `QUOTA_EXCEEDED`, `REPLICA_TARGET_UNMET`) when a client is denied service.
- **No remote keyword leakage by default.** History backfill requests are scoped to time-range windows, and there is no reachable endpoint that accepts keyword-bearing arguments. Remote keyword search remains explicitly out of scope, and command/test artifacts (e.g., `tests/e2e/v22/*` checks) must verify that any new client query surface uses time-based, metadata-free parameters.
- **Relay no-long-history-hosting boundary stays firm.** Durable segments live on the Archivist mesh and private Space replicas; relays never persist segments or manifests beyond temporary buffering. Regression checks for `G9` (Podman relay scenarios and `scripts/v22-history-scenarios.sh`) will demonstrate that relays continue to act as pure transit nodes.

## Gate input expectations

- Scope lock output must be paired with the traceability matrix, threat model, and gate-ownership artifacts listed in `docs/v2.2/phase0`. These files supply the `EV-v22-G0-###` evidence entries before `G0` can be promoted.
- All teams must reference the `docs/templates/roadmap-gate-checklist.md` and `docs/templates/roadmap-signoff-raci.md` when submitting approvals for scope-lock deliverables.

## Next steps

- Once scope lock calibrates architecture, Phase 1 teams may proceed with `P1-T1` through `P1-T3` (Archivist runtime). All downstream work must cite these locked boundaries whenever new pull requests touch history-plane behavior.
