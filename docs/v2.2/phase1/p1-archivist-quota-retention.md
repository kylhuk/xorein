# Phase 1 Archivist Quota & Retention Contract

- Storage keys are `{SpaceID, ChannelID, SegmentID}` and segments track ciphertext size plus insertion timestamp.
- Each space enforces an operator-configured quota; channels may provide tighter caps. Segments respect a global maximum size.
- `QUOTA_EXCEEDED` is returned deterministically when the space or channel budget would be overrun.
- `SEGMENT_TOO_LARGE` is returned when the segment exceeds the configured per-segment limit.
- A retention window governs pruning: any segment older than the configured duration is removed with reason `RETENTION_POLICY`, and usage counters are adjusted atomically.
