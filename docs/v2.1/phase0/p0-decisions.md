# Phase 0 Decisions (v2.1)

1. **D21-1 (Timelines)**: Encrypt local timeline contents in an on-device database; no plaintext is held unencrypted on disk.
2. **D21-2 (Local search)**: Search remains strictly local with no remote keyword assistance in v21; coverage labels communicate what is locally available.
3. **D21-3 (Empty states)**: Every empty timeline must expose a deterministic state—"no history yet", "history cleared locally", or "history missing (no backfill)"—so UIs avoid ambiguous limbo.
4. **D21-4 (Relay boundary)**: Relays retain no long-lived history segments/manifests; durable history is an Archivist feature in v22.
5. **D21-5 (Binary naming)**: Xorein is the backend runtime binary and harmolyn is the frontend UI binary from v21 onward; dependencies between them must remain unidirectional.
