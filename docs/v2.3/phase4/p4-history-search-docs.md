# Phase 4 (P4-T1): History + Search Plane Docs

This artifact captures the planned documentation set that closes gate `G6` (release docs) while outlining the interactions with `G10` (F24 seed/deferral readiness) and `G11` (architecture coverage audit). It is intentionally scoped to descriptions; operational behavior remains gated until these documents are published.

## Gate mapping
- **G6**: Serves as the release-document bundle covering operator, user, and security/privacy narratives for history/search.
- **G10**: Calls out the pending `F24` seeds that must absorb any missing persistence classes invoked from these docs (especially non-guaranteed metadata paths).
- **G11**: Links back to the architecture coverage audit so every referenced class is traced and any placeholder is marked as a seed or deferral.

## ST1 – Operator guidance

### Archivist role
Archivists are the only nodes permitted to serve long-lived history segments, manifest shards, and coverage indexes. Their duties include ingesting backfill requests, reconciling replicas, and maintaining the metadata catalog that maps manifests to durability labels. Archivists expose a deterministic Web API and CLI surface, so operators can automate scaling and observe the same status signals that clients rely on.

### Quotas and retention
Default quotas throttle Archivist interactions to protect storage and CPU resources. Each client is limited to configurable request rates, backfill byte limits, and manifest lookup counts. History retention is tiered: high-priority timeline segments retain for `60 days` by default, lower tiers degrade after `14 days`, and attachments follow a separate `30-day` rolling window unless an operator pin overrides it. Operators can adjust the per-tenant retention window via the `retention-profile` tag, but every change must be annotated in the runbook for auditability.

### Refusal reasons
Archivists return deterministic refusal reasons that surface in metrics and UI logs. Key codes include:

| Code | Condition |
| --- | --- |
| `quota_exceeded` | Client request exceeds the configured rate or backfill byte limit. |
| `retention_conflict` | Requested timeline segment has already expired under the retention profile. |
| `durability_degraded` | Replica targets cannot be satisfied; serves only best-effort reads. |
| `backfill_window_closed` | Backfill request falls outside of the allowed upload window (e.g., for relay-only leases). |

Operators must monitor these codes and correlate with alarms (quota exhaustion, retention drift, replica failure) before attempting retries.

### Durability labeling
Every manifest and section served by an Archivist carries a durability label exposed via the API and mirrored to the UI:

- **`durable`**: Replica targets met, checksum validated, and manifest signed. Archivists can serve data without further reconciliation.
- **`degraded`**: Replica targets unmet due to partial failures; metadata explains which nodes are lagging so operators can electively reroute traffic.
- **`cached`**: Data reconstructed from short-lived caches without full replica validation; suitable only for diagnostics.

Labels must propagate back to the alarm/monitoring stack so operators can correlate them with storage growth and retention drift events.

## ST2 – User guidance

### History availability
History coverage is bounded by Archivist persistence. Clients may ask for timelines up to the retention horizon defined per tenant, but the UI always surfaces the actual available window rather than assuming infinite history. Any gap (expired segment, relay policy) is annotated with the same refusal codes operators see.

### Search coverage labels
Search results carry coverage labels that describe how much data was searched:

- **`full`**: Query covered every known manifest and attachment in the requested timeline.
- **`partial`**: Query skipped data marked `degraded` or `cached`; users can tap the label to open the backfill controls.
- **`expedited`**: Query limited by rate quotas or CPU load, so it only covered high-priority segments.

Each label drives a consistent tooltip and is exposed through the API so automation can rely on the same semantics as the UI.

### Backfill behavior
Backfills are gated by both quotas and availability: `backfill_enabled` must be true for the target tenant, the requested window must overlap with metadata windows tracked by Archivists, and the request is rate-limited. When backfills cannot be satisfied, clients receive a `backfill_window_closed` refusal along with a human-readable timeframe and a recommended retry slot. The documentation calls out how relay-only deployments should advertise this refusal to avoid misinforming users about their timeline coverage.

## ST3 – Security & privacy posture

### Metadata model
Each manifest contains a metadata record with: producer node ID, timeline ID, sequence number, retention profile tag, durability label, and cryptographic checksum. Attachments and indices inherit the manifest’s metadata chain plus their own storage plane tags (e.g., Archivist segment ID, relay cache slot). The metadata model intentionally avoids keyword leakage; only hashed tokens necessary for routing are recorded unless a user explicitly opts into assisted search in a future release.

### Default guarantees
By default, the history/search plane guarantees:

- No keyword-bearing backfill requests are accepted; only time- and retention-bounded ranges are allowed.
- Durability labels accurately describe replica health and are propagated to clients and operators.
- Coverage labels reflect the actual manifest subset queried.
- Metadata includes only the minimum routing data plus retention/durability annotations; no plaintext keyword index exists unless opt-in is enabled later.

### Explicit non-guarantees
The docs also call out what is intentionally *not* guaranteed:

- Search coverage does not imply every relay holds persistent state; relays remain stateless and can only proxy short-term caches.
- Backfill requests are best-effort when durability is degraded; retries may be needed once replicas heal.
- Assisted search (if introduced later) will require separate opt-in flows, and this doc does not cover its leakage budgets yet.
- Any missing persistence class uncovered in the architecture coverage audit must be treated as a future `F24` seed (see `G11`).

## Evidence note
Docs-only verification—tests N/A (no runnable artifacts were executed for this write-up).
