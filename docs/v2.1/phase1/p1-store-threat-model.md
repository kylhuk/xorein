# Phase 1 Store Threat Model (v2.1)

## Adversaries
- Device thief with offline access to the filesystem but no live keying material.
- Malware that can read encrypted files but cannot inspect OS keystore secrets or runtime memory.
- Relay/archivist nodes that see replicas of ciphertext only and never plaintext.

## Non-Goals
- Defending against malware with live access to the decrypted process memory or the key derivation secrets.
- Providing remote keyword-intent inference or remote history guesses.

## Countermeasures
- Encrypt the local SQLite/SQLCipher-backed store with a key derived from the device secret plus the canonical identity secret.
- Wipe the store key and file whenever a "clear local history" action executes, ensuring offline copies lose their key.
- Emit deterministic failure reasons and audit logs (locked, corrupt, migration required, quota exceeded) so clients can recover safely.
