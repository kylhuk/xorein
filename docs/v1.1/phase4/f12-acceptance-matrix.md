# Phase 4, F12: Acceptance Matrix

Planning artifact only; this table maps v12 requirements back to v11 spec deliverables and traces to planned validation evidence (EV-v11-G5-### for spec approvals, EV-v12-### reserved for future runtime tests).

| Requirement | Validation approach | Planned evidence |
|---|---|---|
| Immutable global identity invariants (unique identity id, no rekey). | Unit/integration tests that attempt duplicate identity creation, metadata mutation, and conflict detection; deterministic reason labels must match `f12-backup-recovery-flows.md`. | `EV-v11-G5-010` (spec approval), later `EV-v12-G2-001` for runtime tests. |
| Explicit no-password-reset policy + UX warning. | UX checklist that proves warning shown before irreversible persistence and disallows server reset links; capture copy in onboarding spec. | `EV-v11-G5-011` (UX spec), `EV-v12-G2-002` (UX regression). |
| Local backup security envelope (`BackupID + BackupPassword`). | Security review of backup format (Argon2id + AEAD), manual proof that no server endpoint can accept backup payloads. Reference Phase 2 relay constraints to ensure relays do not store backup blobs. | `EV-v11-G5-012` (format spec + review). |
| Deterministic restore/error taxonomy (reasons defined in flows). | Table-driven documentation of each reason, target UX copy, and telemetry keys. Validate via doc review and planned instrumentation in v12 (telemetry events per reason). | `EV-v11-G5-013` (taxonomy spec), `EV-v12-G4-003` (telemetry validation). |

## Security coverage
- **Tamper detection (backup-corrupt):** Planned security walkthrough to confirm AEAD failures surface `backup-corrupt` reason and do not leak keys. Evidence: `EV-v11-G5-014` for spec, `EV-v12-G4-015` for test results.
- **Identity conflict protection:** Tests that ensure identity mismatch halts before overwriting existing local identity; includes deterministic dialog referencing `identity-mismatch`. Evidence: `EV-v11-G5-015`, `EV-v12-G4-016`.

## Reliability coverage (positive/negative/degraded/recovery)
| Scenario type | Description | Validation approach | Evidence |
|---|---|---|---|
| Positive | Successful restore on new device with valid backup. | e2e scenario `NewDeviceRestore` verifying restored identity and config migration with manifest `EV-v12-G4-017`. | `EV-v12-G4-017` |
| Negative | Wrong `BackupPassword` entry. | Unit/integration tests ensure client emits `backup-password` reason and UX halts. | `EV-v11-G5-005` (spec planning), `EV-v12-G4-018` (test). |
| Degraded | Corrupted backup payload (tamper). | Simulated tamper yields `backup-corrupt` reason and offers guidance to export anew. | `EV-v11-G5-006` (spec), `EV-v12-G4-019` (test). |
| Recovery | Conflict (existing identity) handling. | Scenario ensures `identity-mismatch` reason is surfaced, requiring manual override before overwrite. | `EV-v12-G4-020`. |

## QoL effort-reduction target
- **Objective:** Reduce onboarding/recovery effort by at least 10% compared to v10 baseline by using deterministic reason labels and manifest copy/paste instructions instead of freeform heuristics.
- **Measurement:** Count user steps via `tests/e2e/v12/recovery_flow_steps` vs. `tests/e2e/v10/onboarding_steps` (planned instrumentation). Record results and evidence in `EV-v12-G4-021` when the metric is captured.

## Notes
- All requirements defer to Phase 2 relay boundary constraints (`docs/v1.1/phase2/p2-relay-data-boundary.md`) to ensure backup artifacts are never relay-hosted.
- Evidence placeholders may evolve; update the matrix once EV IDs are assigned for actual v12 testing.
