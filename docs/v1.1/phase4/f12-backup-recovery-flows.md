# Phase 4, F12: Local Backup & Recovery Flows

Planning artifact only; execution details belong to v12 runtime work. This document enumerates the local-backup lifecycle, deterministic reason taxonomy, and recovery scenarios referenced by `f12-acceptance-matrix.md`.

## Backup export flow (client-local only)
1. User opts into backup and is shown the no-password-reset warning (see `f12-identity-backup-spec.md`).
2. Client derives an encrypted envelope from `BackupPassword` (Argon2id + AEAD) that includes the identity root key, latest metadata, and a version marker.
3. `BackupID` (UUID) is generated and written to a manifest displayed to the user; the manifest is never transmitted to the relay.
4. Export operation stores the encrypted payload and manifest in the local vault; the exported file is removable, copyable, or printable, but the relay stores nothing (per Phase 2 relay boundary constraints).
5. Export completion emits a deterministic success event with evidence ID placeholder `EV-v11-G5-002`.

## Restore-on-new-device flow
1. On a fresh device, the client prompts for `BackupID` (validation only) and `BackupPassword` before attempting decryption.
2. Decryption failure surfaces either `backup-password` (wrong password) or `backup-corrupt` (tampering/payload corruption).
3. Successful decryption compares the recovered identity metadata with any existing local identity:
   - If no identity exists, import succeeds and the client enters the onboarding completion path.
   - If a different identity exists, the client emits `identity-mismatch` and blocks the overwrite until user-approved fallback (see `f12-identity-backup-spec.md`).
4. Every restore path records deterministic telemetry (e.g., `reason=backup-password`) to accelerate debugging and gate evidence.

## Deterministic error & reason taxonomy
| Reason label | Trigger | Planned UX handling | Recovery implication |
|---|---|---|---|
| `backup-password` | Password derivation/AEAD rejects the password | Show “wrong backup password” prompt with fingerprint reminder | Retry allowed without data loss |
| `backup-corrupt` | AEAD or version marker mismatches | Surface tamper warning, mark backup as unusable | User must fetch new backup |
| `identity-mismatch` | Restored identity differs from existing local identity | Block automatic import, require manual acknowledgement | Failure path protects existing identity |
| `identity-duplicate` | Attempt to create identity already seen by other device | Display conflict reason; direct user to confirm migration | Allows explicit overwrite once acknowledged |
| `backup-outdated` | Backup version marker older than supported minimum | Inform user of incompatible backup version | Recovery requires new backup or defer |

This taxonomy drives deterministic dialog copy, consistent telemetry, and the replayable evidence IDs described in `f12-acceptance-matrix.md`.

## Negative/degraded/recovery scenarios (planned)
- **Negative:** Wrong `BackupPassword` path should stop before identity coercion and call out `backup-password`. Validation via unit/integration tests (EV-v11-G5-005). Planned UX tests cover the reassurance copy.
- **Degraded:** Corrupted local backup should emit `backup-corrupt` and fall back to “create new backup” guidance; the relay never receives corrupted material (evidence EV-v11-G5-006). Recovery plan includes new backup export UI and guidance.
- **Recovery:** Successful restore on new device ensures identity integrity and incremental config migration; expect reliability metrics showing <1% failure and QoL target of 10% less perceived steps (see `f12-acceptance-matrix.md`).

## Security & reliability notes
- All flows rely on local storage; no hosted backup or relay persistence is permitted (per `TODO_v11.md` and `TODO_v12.md`).
- Tamper detection and deterministic reason labels feed into security evidence (EV-v12-G4-###) so that replayable telemetry proves identity remains immutable even when facing conflict.

## QoL effort-reduction target
- The planned QoL objective for v12 is to reduce onboarding/recovery user effort by 10% by reusing the deterministic taxonomy and manifest display, compared with the legacy freeform instructions. Measurement plan: compare user steps counted in `tests/e2e/v12/recovery_flow_steps` with baseline from `v10` onboarding scenario and report in EV-v12-G4-###.
