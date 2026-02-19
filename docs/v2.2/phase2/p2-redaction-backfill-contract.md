# P2-Redaction Backfill Contract

Backfilled events are always subject to moderation overrides. Tombstones win and their presence is propagated to indexing/search surfaces, with a disclosure helper that frames deletion coverage limits deterministically.

- Backfill time ranges must be strictly ordered (`Start < End`). Invalid ranges are rejected immediately with the deterministic reason `BACKFILL_INVALID_RANGE` and no segments are fetched.
- Fetched ciphertext segments flow through the apply semantics that merge them into the local store, and every application step updates an observable progress report that tracks applied vs total segments.
- Progress queries return the current reason (one of `BACKFILL_DENIED`, `BACKFILL_INCOMPLETE`, `BACKFILL_VERIFY_FAILED`, `BACKFILL_INVALID_RANGE`, `BACKFILL_APPLY_FAILED`) plus the applied and total segment counts so UX can highlight remaining gaps.
- Modern moderate backfill requests expose the same reason taxonomy so callers can surface deterministic next actions (backfill again, adjust range, or cancel).
