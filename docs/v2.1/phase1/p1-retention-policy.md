# Phase 1 Retention Policy (v2.1)

- **Defaults**: 30-day window, capped at 1,000 non-pinned events per channel. These defaults are configurable but enforced by deterministic pruning logic.
- **Pruning order**: The policy evicts the oldest non-pinned, non-system events per channel first, ensuring that pinned/system rows and recently opened conversations survive longer.
- **Clear local history**: This action wipes the encrypted store, resets search coverage to empty, and derives a fresh key; clients receive a deterministic struct describing the timeline and search-state reset for auditability.

Retention behavior must run with deterministic inputs so that UI automations, diagnostics, and privacy reviews can recreate prune decisions and clear-history events exactly.
