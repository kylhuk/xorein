# v1.2 Phase 1 - Local Backup Format

## Status
Implemented in `pkg/v12/backup/backup.go`.

## Envelope schema
| Field | Type | Description |
|---|---|---|
| `version` | int | Envelope schema version (`1`). |
| `backup_id` | string | Deterministic backup identifier (`BKP-...`). |
| `created_at` | RFC3339 timestamp | Local export timestamp (UTC). |
| `salt` | base64 | Argon2id salt (16 bytes). |
| `nonce` | base64 | AEAD nonce (12 bytes). |
| `ciphertext` | base64 | Encrypted payload (identity + critical config). |
| `ciphertext_sha256` | hex | Tamper pre-check hash for deterministic corruption classification. |

## Cryptography profile
- KDF: `Argon2id` (`iterations=3`, `memory=64MiB`, `threads=1`, `key_len=32`).
- Cipher: `AES-256-GCM`.
- No relay/server storage for ciphertext, salt, nonce, or key material.

## Deterministic restore reasons
- `backup-password`
- `backup-corrupt`
- `identity-mismatch`
- `identity-duplicate`
- `backup-outdated`

## Planned vs implemented
- Envelope format and reason taxonomy are implemented and tested.
- Hosted backup remains out of scope.
