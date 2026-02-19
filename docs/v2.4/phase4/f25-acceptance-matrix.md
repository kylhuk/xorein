# F25 Acceptance Matrix

Planning-only matrix for `F25` that captures the performance, quota, and acquisition requirements promised in ST5 while keeping implementation completion implicit.

| Criterion | Target | Evidence placeholder |
| --- | --- | --- |
| Upload throughput | Sustain `X` MB/s aggregated per Archivist provider with headroom for `N` concurrent uploads and backpressure-aware chunking. | `EV-v24-G7-###` (placeholder for v25 execution) |
| Storage quota enforcement | Honor per-Space `/ per-user` quotas with `QuotaExceeded` codes; descriptive telemetry must appear when thresholds cross 80%. | `EV-v24-G7-###` (placeholder for quota evidence) |
| Retrieval latency | Private Space blob retrieval must respond within `Y` ms under nominal load and expose paginated cursors that do not expose additional `BlobRef`s. | `EV-v24-G7-###` (planning placeholder for latency measurements) |
| Anti-enumeration validation | API rejects non-member requests with generic `ResourceNotFound` while recording rate-limited counters; Archivist mirrors same behavior for cached copies. | `EV-v24-G7-###` (placeholder for anti-enumeration proof) |
| Evidence completeness | `F25` spec package is published alongside an evidence index entry covering scope + perf documentation, aligning with `G7` and `G8` gate requirements. | `EV-v24-G7-###` / `EV-v24-G8-###` (placeholders; `G7` for spec, `G8` for closing evidence bundle) |

Further implementation evidence (command outputs, test runs, etc.) will replace these placeholders during v25 execution; this matrix simply anchors the intended checks.
