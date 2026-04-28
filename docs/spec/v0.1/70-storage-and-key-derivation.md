# 70 — Storage and Key Derivation

This document specifies the at-rest storage format, encryption parameters,
and key derivation scheme for the Xorein node's persistent state database.

## 1. Storage engine

The node stores all persistent state in a single **SQLCipher** database file.
SQLCipher is SQLite extended with transparent AES-256 page-level encryption.

### 1.1 File layout

All storage files reside under the configured `data_dir`:

| File | Description |
|------|-------------|
| `state.db` | SQLCipher-encrypted SQLite database |
| `state.db.meta.json` | Key derivation metadata (salt, key check, timestamps) |
| `state.key` | Local secret (base64url, 32 bytes); absent if `XOREIN_STATE_KEY` env is set |
| `control.token` | Bearer token for the local control API (see `60-local-control-api.md`) |
| `control.addr` | TCP address of the control API server (Windows only) |

All files MUST be created with permissions 0600 (owner read/write only).

### 1.2 SQLCipher parameters

| Parameter | Value |
|-----------|-------|
| Encryption algorithm | AES-256 (SQLCipher default) |
| Key length | 32 bytes (256 bits) |
| Page size | 4096 bytes |
| Journal mode | WAL |
| Foreign keys | ON |
| Busy timeout | 5000 ms |

The SQLCipher connection string passes the derived key via `PRAGMA key = x'...'`
(raw hex, not passphrase mode).

## 2. Key material sources

The encryption key is derived from two inputs:

1. **Salt** — 32 bytes of cryptographic random, generated once at store creation
   and stored in `state.db.meta.json`. Never changes for the lifetime of the store.

2. **Secret** — 32 bytes of cryptographic random that proves authorization to
   open the store. Resolved in priority order:
   - The `XOREIN_STATE_KEY` environment variable (UTF-8 string, treated as raw bytes).
   - The `state.key` file in `data_dir` (base64url-encoded, no padding).
   - If neither exists and the store is being **created**, a fresh 32-byte secret
     is generated and written to `state.key` with permissions 0600.
   - If neither exists and the store **already exists**, the node MUST abort with
     `ErrWrongKey`.

## 3. Key derivation (current — v0.1)

**Current scheme (to be replaced by Argon2id in Phase 2):**

```
store_key = SHA-256(salt || secret)
```

Where `||` is byte concatenation, `salt` is the 32-byte raw salt bytes, and
`secret` is the raw secret bytes (not base64).

This scheme is recognized as weak (no memory/time hardness, no iterations).
It is retained in Phase 1 to match the existing runtime. Phase 2 MUST replace
it with the Argon2id scheme defined in §4 below.

### 3.1 Key verification

On store open, the node verifies that the derived key is correct before
attempting any database operation:

```
key_check = base64url_no_pad(SHA-256(store_key || b"xorein-state-store-key-check"))
```

The computed `key_check` MUST match `state.db.meta.json`'s `key_check` field.
Mismatch → `ErrWrongKey`; the node MUST NOT open the database.

## 4. Key derivation (target — Argon2id)

Phase 2 MUST migrate key derivation to Argon2id per
[RFC 9106](https://www.rfc-editor.org/rfc/rfc9106):

```
store_key = Argon2id(
    password = secret,          // 32-byte raw secret
    salt     = salt,            // 32-byte random salt from meta
    memory   = 65536,           // 64 MiB
    time     = 3,               // 3 iterations
    threads  = 4,               // parallelism degree
    keyLen   = 32,              // output key length
)
```

The `state.db.meta.json` `format_version` field MUST be incremented (3 → …)
when Argon2id is adopted. Nodes reading a v2 meta file MUST use the SHA-256
scheme (§3); nodes reading a v3 or later meta file MUST use Argon2id.

The migration procedure (§7) governs in-place upgrade from SHA-256 to Argon2id.

## 5. Metadata file format

`state.db.meta.json` is a JSON document with permissions 0600:

```json
{
  "format_version": 2,
  "salt": "<base64url 32 bytes, no padding>",
  "key_check": "<base64url SHA-256(store_key || check-label), no padding>",
  "created_at": "RFC3339Nano",
  "updated_at": "RFC3339Nano"
}
```

| Field | Type | Notes |
|-------|------|-------|
| `format_version` | int | 2 = SHA-256 KDF; 3+ = Argon2id |
| `salt` | string | base64url, no padding, 32 bytes decoded |
| `key_check` | string | base64url, no padding; see §3.1 |
| `created_at` | string | RFC3339Nano; store creation timestamp |
| `updated_at` | string | RFC3339Nano; last key rotation or migration |

The file MUST be written atomically (write to `.tmp`, then `rename`) to prevent
corruption on crash.

## 6. Database schema

The SQLCipher database contains two tables:

### 6.1 `store_metadata`

Singleton row (id = 1) tracking the in-database schema version:

```sql
CREATE TABLE store_metadata (
    id             INTEGER PRIMARY KEY CHECK (id = 1),
    format_version INTEGER NOT NULL,
    schema_version INTEGER NOT NULL,
    created_at     TEXT NOT NULL,   -- RFC3339Nano
    updated_at     TEXT NOT NULL    -- RFC3339Nano
);
```

`format_version` mirrors the meta file's value. `schema_version` is
incremented when the bucket payload schemas change (independent of the
encryption key scheme).

### 6.2 `store_buckets`

Named key-value store for all node state:

```sql
CREATE TABLE store_buckets (
    name    TEXT PRIMARY KEY,
    payload BLOB NOT NULL
);
```

Each `payload` is a JSON-encoded Go struct specific to that bucket name.
Bucket names are stable identifiers; implementations MUST NOT rename buckets
(treat as reserved field numbers in protobuf).

### 6.3 Bucket names

| Bucket name | Content |
|-------------|---------|
| `identity` | Local identity (peer ID, public/private keys, profile) |
| `peers` | Discovered peer records |
| `servers` | Joined server records |
| `channels` | Channel records (text and voice) |
| `messages` | Chat message history |
| `dms` | Direct message pair records |
| `friends` | Friend records and pending requests |
| `relay_queue` | Store-and-forward queue entries |
| `voice_sessions` | Active and recent voice session state |
| `notifications` | Notification records and read-through markers |
| `settings` | Node-level settings and preferences |

Additional buckets MAY be added by future schema versions. Unknown bucket
names encountered during load MUST be preserved (round-trip safe).

## 7. Key rotation procedure

Key rotation replaces `secret` (and optionally `salt`) while keeping all
stored data accessible. Rotation is performed offline (node stopped) or via
a dedicated control operation (future).

```
1. Derive current store_key from current (salt, secret).
2. Decrypt the SQLCipher database using the current store_key.
3. Generate new_secret (32 bytes random) and optionally new_salt (32 bytes random).
4. Derive new_store_key from (new_salt or old_salt, new_secret).
5. Re-encrypt the database using SQLCipher PRAGMA rekey = x'<new_store_key>'.
6. Write new state.db.meta.json with new_salt (if rotated), new key_check, and
   updated_at = now.
7. Write new state.key (or update XOREIN_STATE_KEY guidance to the caller).
```

On failure at any step between 5 and 7, the old files MUST be restored from
the backup taken in step 2.

### 7.1 Argon2id migration

When upgrading from SHA-256 (format_version = 2) to Argon2id (format_version = 3):

```
1. Perform key rotation as above, using Argon2id to derive new_store_key.
2. Write state.db.meta.json with format_version = 3.
3. Verify the new key_check before committing.
```

## 8. Legacy store migration

Nodes that shipped with the old flat-file store format (`state.store/` directory,
format_version = 1) are automatically migrated on first open:

1. Load the legacy store from `<data_dir>/state.store/`.
2. Save all buckets to the new SQLCipher database.
3. Rename `state.store/` to `state.store.migrated-<timestamp>` for archival.

If migration fails, the node MUST abort and leave the legacy store intact.

## 9. Cryptographic material storage

### 9.1 Identity keys

The local identity's private key material is stored in the `identity` bucket.
As of v0.1 the private keys are JSON-encoded alongside the public keys. Phase 2
MUST migrate to storing private key material in a sub-derived encryption layer
(double encryption: outer = store_key, inner = identity-specific KDF output).

### 9.2 Session state (Seal mode)

Double Ratchet session state (root key, chain keys, skip list) is stored in the
`identity` or `dms` bucket as a sub-field of the DMRecord. All key material
within session state is protected by the store encryption. Session state
MUST be deleted securely (zeroed then removed) when a DM is deleted or a
session is replaced.

### 9.3 Relay queue

Entries in the `relay_queue` bucket MUST contain only ciphertext payloads.
The relay node MUST NOT store any decryptable material. The `payload` field
of each relay queue entry is the raw AEAD ciphertext from the sending client;
the relay never holds a decryption key.

## 10. Security properties

| Property | Value |
|----------|-------|
| At-rest encryption | AES-256 (SQLCipher page-level) |
| Key derivation (v0.1) | SHA-256 — weak, no hardening |
| Key derivation (target) | Argon2id (memory=64 MiB, time=3, threads=4) |
| Salt length | 32 bytes (256 bits) |
| Secret entropy | 32 bytes (256 bits) |
| Key check | SHA-256(key \|\| label) — detects wrong key before DB open |
| Permissions | 0600 on all key material and token files |
| Journal mode | WAL — reduces exposure window of plaintext page flushes |
| Relay queue | Ciphertext only; no key material stored at relay |

## 11. Conformance

Implementations MUST:

- Use SQLCipher with the parameters in §1.2.
- Derive the store key using the scheme matching the `format_version` in
  `state.db.meta.json`.
- Write `state.db.meta.json` atomically.
- Reject opens where `key_check` does not match without attempting to open
  the SQLite database.
- Set file permissions to 0600 for `state.key`, `state.db`, `state.db.meta.json`,
  and `control.token`.
- Preserve unknown bucket names on load and re-save (round-trip safe).
- Migrate v1 (flat-file) stores to v2 (SQLCipher) automatically.
- NOT store relay queue entries with decryptable key material.
