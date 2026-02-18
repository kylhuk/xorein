# Phase 4, F12: Proto Delta (Planning)

Planning artifact only; actual `proto/aether.proto` edits happen in v12 (see `TODO_v12.md`). The list below enumerates the additive changes v12 must deliver without renumbering existing fields.

| Change | Type | Compatibility notes |
|---|---|---|
| Add `IdentityState` message under the identity section. Includes `identity_id` (string), `identity_public_key_fingerprint`, `created_at`, `metadata_version`, and `backup_status` fields. All field numbers must be new (e.g., `103`, `104` etc.) and never reuse retired numbers. | message | Additive only; other messages keep existing numbering. Clients that do not read this message should ignore unknown fields. |
| Add `BackupManifest` message that carries `backup_id`, `backup_version`, `created_at`, and the deterministic reason enum (see next row) for client-local diagnostics and duplicate-export detection. | message | Safe additive change. Manifest data is client-local and must not introduce relay-hosted backup payloads or key material. |
| Add `BackupReason` enum with values such as `BACKUP_REASON_USER_INITIATED`, `BACKUP_REASON_AUTOMATED_SNAPSHOT`. Values appended at the end and never renumbered; existing enum values remain untouched. | enum | Additive; reserves no value numbers and adds enumerants only. |
| Add `identity_reason` and `backup_reason` fields to relevant RPC response messages (e.g., `IdentityRestoreResponse`) so that v12 clients can emit deterministic reason labels. | field additions | Place these fields at the end of the message definition with new numbers (e.g., `identity_reason = 12`). Field presence is optional for older clients. |

## Forward-compatibility guardrails
- Do not renumber existing identity/backup-related fields in `proto/aether.proto`. If a field must be removed, mark its number as `reserved` and keep the name referenced in comments, but v12 targets addition-only changes per governance.
- Document each new field/message in the proto changelog (`docs/v1.1/phase4/f12-acceptance-matrix.md` and later the v12 proto changelog). Reserve future numbers before adding any new field to avoid collisions.
- Add descriptive comments to every field so downstream code generators can keep deterministic reason taxonomies aligned with `f12-backup-recovery-flows.md`.
- Do not add any private key or backup ciphertext field to on-wire protocol messages; private key material remains local-only inside encrypted backup artifacts.

Planned proto evidence sits under EV-v11-G5-003 (spec) and will extend to EV-v12-G1-### when actual proto diffs land.
