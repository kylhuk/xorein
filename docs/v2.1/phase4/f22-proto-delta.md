# F22 Proto Delta

> Planning artifact that records the additive wire changes for the `F22` specification. No fields are removed or reshaped—every entry is a new, optional addition that keeps prior version compatibility intact.

## Additive messages and fields
| Message | Description | Field notes |
|---|---|---|
| `ArchivistAdvertisement` | Describes an Archivist node’s capacity, retention, and headroom so peers can make informed retrieval decisions. | Add optional `space_id`, `advertised_retention`, `advertised_quota`, `available_epochs`, `replication_target` fields. Reserve field numbers `10..19` for future telemetry metadata. |
| `HistorySegmentManifest` | Captures integrity proofs for a segment that an Archivist can serve. | Adds `segment_id`, `channel_id`, `range_start`, `range_end`, `manifest_hash`, `signature`, optional `tombstone_pointer`. Fields live in the 20..29 range to avoid collisions. |
| `BackfillRequest` | Models a time-bound backfill request that stays within v21 privacy constraints. | Includes `space_id`, `channel_id`, `from_timestamp`, `to_timestamp`, `coverage_hint`, `expected_epoch`. All fields are optional/nullable so that future extensions can add finer-grained selectors. |
| `BackfillResponse` | Streams segment payloads with manifest proofs or failure reasons. | Carries `segment_manifest`, repeated `segment_block`, `failure_reason`, `progress_nonce`. Reserve `failure_reason` enum values 30..39 for documented refusal reasons such as `BACKFILL_TOO_NEW` or `BACKFILL_INTEGRITY_MISMATCH`. |
| `SearchCoverageSummary` | Extends coverage metadata to include backfilled ranges. | Adds `channel_id`, `window_start`, `window_end`, `coverage_percent`, `coverage_state` (enum). Place these fields after existing coverage messages to maintain ordinal stability. |

## Services and RPC guidance
- Introduce an `ArchivistService` with read-only RPCs such as `Advertise`, `RequestBackfill`, `StreamBackfill`, and `ReportCoverage`. Each RPC is optional and uses streaming responses where necessary.
- Avoid adding required fields to existing messages; prefer optional wrappers that clients can ignore until they opt into backfill flows.

## Field allocation guardrails
- Reserve field numbers `40..59` inside `Metadata` messages for future Archivist telemetry and nonces so we can expand without renumbering later.
- Document every new field in `proto/aether.proto` with comments that note its introduction in `F22` and cite this spec doc for traceability.

## Compatibility requirements
- Run `buf lint` and `buf breaking --against 'origin/main'` to confirm no existing clients break. Tag new enums with explicit defaults so older binaries treat them as `UNSPECIFIED`.
- Keep serialization stable by avoiding `oneof` migrations: new data lives behind optional fields or entirely new messages.

## Traceability and licensing
- This delta feeds into `docs/v2.1/phase4/f22-acceptance-matrix.md` and satisfies the `P4-T1` `ST2` & `ST3` acceptance criteria.
- Code inherits the AGPL; documentation inherits CC-BY-SA, matching the repo’s baseline language in `AGENTS.md` and `docs/v2.1/phase0/p0-scope-lock.md`.
