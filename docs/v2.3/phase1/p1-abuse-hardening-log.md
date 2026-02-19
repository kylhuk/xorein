# P1-T1: Abuse Hardening Log

Conservative abuse resistance primitives now live under `pkg/v23/security`. The implementation targets the Phase 1 gate requirements while keeping the API small, deterministic, and aligned with privacy-first telemetry.

## ST1 – Conservative quotas and retention
- `QuotaEnforcer` enforces entry and retention defaults, returning deterministic refusal reasons (`quota_entries_exceeded`, `quota_retention_exceeded`).
- The enforcer tolerates zero-valued limits (no enforcement) and keeps failure strings stable for automation and operator diagnostics.

## ST2 – Rate limiting and bounded responses
- `RateLimiter` applies a single-window counter and resets on window rollover; once the limit is reached it refuses with `rate_limit_exceeded` plus deterministic window metadata.
- Response payload bounds are enforced by `ValidateResponseSize`, refusing with `response_size_exceeded` when payloads would grow beyond configured bytes.

## ST3 – Privacy-preserving abuse telemetry
- `TelemetryAggregator` records only sorted, aggregated label tuples and keeps counters that can be snapshot without exposing raw request content.
- Recording and snapshotting are thread-safe and deterministic, satisfying privacy expectations for abuse metrics.

## Gate mapping (G2)
- G2 (Security hardening) is satisfied for abuse resistance by ST1–ST3: quota/refusal trade-offs, deterministic rate/size hard limits, and aggregated telemetry evidence are now codified alongside deterministic refusal reasons.
