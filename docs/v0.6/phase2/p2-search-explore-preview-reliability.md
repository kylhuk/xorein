# Phase 2 · P2 Search/Explore/Preview Reliability

## Purpose
Turn search, explore, and preview failure paths into deterministic recovery stories so V6-G2 can gate partial failure handling without runtime ambiguity.

## Contract
- `P2-T1` (`VA-S1`) defines partial-failure semantics for keyword/category queries, including reason-coded coverage for success, service failures, and cached fallback deliveries (`search.partial.*`).
- `P2-T2` (`VA-S2`) clarifies explore-feed degradation modes, freshness-controlled ordering, and rerank recovery steps (`search.explore.*`) so feeds remain deterministic.
- `P2-T3` (`VA-S3`) spells out preview-to-join mismatch detection, mismatch advisories, and canonical preview recalibration (`search.preview.*`) to keep gating deterministic.

### Deterministic rule table
| Input | Outcome | Reason-class | Fallback/Recovery | Evidence anchor |
|---|---|---|---|---|
| Query response completeness (shards hit vs. expected) | If all shards respond, `search.partial.success`; if timeouts prevent completion, `search.partial.failure` with fallback `search.partial.fallback` | `search.partial` | Deterministic partial-failure classification triggers cached fallback with ordering hints | `pkg/v06/search/contracts.go#ClassifyPartialFailure` |
| Explore items with freshness timestamps plus freshness cutoff | Items ordered first by freshness then by stable `explorePriority`; degradation sets `search.explore.degraded` when cutoff exceeded | `search.explore` | Freshness-guided rerank reorders feed and emits `search.explore.recover` | `pkg/v06/search/contracts.go#OrderExploreFeed` |
| Preview hash vs. canonical join hash | Matching hashes emit `search.preview.success`; mismatch triggers `search.preview.mismatch` with advisory | `search.preview` | Recovery realigns preview via canonical metadata, tagging `search.preview.align` | `pkg/v06/search/contracts.go#DecidePreviewMismatch` |
