# P1 Privacy Conformance

## ST1: Keyword-free backfill by default
- All backfill request validators reject keyword-bearing queries unless `PrivacyConfig.AllowKeywordBackfill` is opt-in.
- Backfill metadata remains limited to `{SpaceID, ChannelID, TimeRange}` and the request validator surfaces `keyword_backfill_not_allowed` reasons when keywords are blocked.

## ST2: Deterministic coverage labeling
- `CoverageState` records missing ranges and completion markers, and `LabelCoverage` never returns `coverage.full_history` when gaps exist.
- Partial or unknown histories are surfaced as `coverage.partial_history` or `coverage.incomplete_history` to prevent overclaiming.

## ST3: Assisted search opt-in scaffolding
- `AssistedSearchGate` starts disabled, and explicit consent tokens must be supplied before assisted search calls are permitted.
- The gate exposes helpers for reason hints so downstream UX/docs can explain why assisted search remains disabled.

## Tests
- `tests/e2e/v23/privacy_coverage_test.go`
- `tests/e2e/v23/privacy_assisted_test.go`
