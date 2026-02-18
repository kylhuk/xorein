# Phase 4, F12: Identity + Backup Specification

Planning notes only; this document captures the immutable identity and local-backup contracts that v11 must deliver for v12 implementation (see `TODO_v11.md` and `TODO_v12.md`).

## Immutable global identity rules
- Each identity is globally unique, self-owned, and bound to a single cryptographic root key pair. No server or relay may reassign or rekey an existing identity without user consent and explicit backup replay. The identity identifier (e.g., `IdentityID`) is immutable once published by the client.
- Identities may include optional metadata (display name, avatar pointer), but these metadata fields cannot be repurposed to override the immutability invariant; updates produce deterministic metadata change events, not new identity references.
- Duplicate identity creation attempts are treated as conflicts (`reason=identity-duplicate`) and require the client to surface the deterministic reason taxonomy described in `f12-backup-recovery-flows.md`.

## No password reset + UX warning contract
- v12 deliberately omits any server-side password-reset endpoint; password loss without a valid local backup is unrecoverable and considered a user action outcome. Relay or operator tooling must never expose a recover-by-reset route.
- Clients must present the warning text defined in the UX contract during onboarding and backup export: “No host or relay can reset your password. Keep a local backup (BackupID + BackupPassword) to recover.” The warning must be surfaced before any irreversible identity persistence.
- The UX contract specifies an explicit callout in the onboarding flow that links to `docs/v1.1/phase2/p2-relay-data-boundary.md` so that the user understands relay nondurable storage guarantees while providing the backup.

## Local backup model
- Backups remain purely local artifacts encrypted under the `BackupPassword`. They are identified by the deterministic `BackupID` (UUID or similar) generated at export time.
- The envelope uses Argon2id (or equivalent) for `BackupPassword` derivation and an AEAD cipher for payload encryption. The payload contains the identity root key, metadata necessary for identity reconstruction, and the signed backup version marker. No relay-hosted or cloud-hosted backup service is permitted under this specification.
- Backup export must record the `BackupID` in a read-only manifest that clients can display or copy; the manifest is the only recovery locator shared with external parties.

## Conflict handling and deterministic reason taxonomy
- Restores must compare the incoming backup metadata with the local identity state. Every mismatch yields a deterministic reason label (`identity-mismatch`, `backup-outdated`, `backup-corrupt`, `identity-duplicate`), and clients must map these labels to UX copy derived from `f12-backup-recovery-flows.md`.
- Conflicts raised during restore (for example, trying to restore a backup for identity A when the device already holds identity B) must block automatic overwrite; the client must require user acknowledgement via an explicit dialog referencing the `identity-mismatch` reason.
- Invalid backups (wrong password, truncated payload, AEAD failure) emit `backup-corrupt` or `backup-password` reasons, enabling repeatable diagnostics and replayable gate evidence (EV-v11-G5-### placeholders).

## Reliability & security implications
- The identity spec demands deterministic onboarding restore flows to reduce user effort by at least the targeted 10% QoL improvement (see `f12-acceptance-matrix.md` for measurement plan).
- All identity persistence respects the relay no-data-hosting boundary by never writing backup or identity key material to the relay; workflows must rely on client-local storage only. Phase 2 (`docs/v1.1/phase2/p2-relay-data-boundary.md`) documents the required relay boundaries that this spec depends on.
- Every backup operation includes tamper detection so that a corrupted local backup surfaces the `backup-corrupt` deterministic reason before exposing any secret keys.

## Planned evidence traceability
- The badge for this planning artifact is EV-v11-G5-001 (spec approval) with follow-on EV-v12-G4-### reserved for execution and UX validation.
