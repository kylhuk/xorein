# Phase 1 Store Reason Taxonomy (v2.1)

| Reason | Description | Remediation |
| --- | --- | --- |
| `STORE_LOCKED` | The local store is locked (e.g., clear-history in progress or replay guard). | Retry after lock clears or escalate if lock persists. |
| `STORE_CORRUPT` | Detected corruption in the database file or indexes. | Freshly derive a key, restore from backup, or clear local history. |
| `STORE_MIGRATION_REQUIRED` | Store schema version mismatched the expected version. | Run the migration path or clear local history after confirming compatibility. |
| `STORE_QUOTA_EXCEEDED` | The retention policy or configured limits block additional inserts. | Trigger pruning, adjust retention settings, or reduce payload size. |

Clients must treat these reasons as deterministic outcomes from ingestion and retention operations so UI and automation can present consistent recovery guidance.
