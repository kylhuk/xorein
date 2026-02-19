# P2 Redaction & Persistence Contract

- **Tombstone semantics**: applying a tombstone removes the plaintext payload and any attachment preview text while preserving envelope metadata, the redaction reason code, and the audit pointer. Tombstones are idempotent, so duplicate deliveries do not alter the preserved metadata.
- **Search integration**: tombstones must notify the local search index (via the integration hook) so that redacted entries no longer appear in search results. The store keeps a deterministic hook call count per entry to guarantee a single removal event.
- **Audit visibility**: each tombstone records the reason pointer and timestamp, allowing UI surfaces to explain why an entry is obfuscated without exposing original content.
- **E2EE limitation**: document that once plaintext leaves the device (e.g., via compromised synced peers), redaction cannot retroactively delete that copy. Local tombstones only affect the device apply-ing them.
