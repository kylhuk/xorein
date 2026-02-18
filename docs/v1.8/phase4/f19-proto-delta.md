# F19 Proto Delta

No protobuf fields are changed in v18. This file defines planned additive v19 additions to preserve contract-first continuity.

## Planned additions (additive)
1. Add a `JoinPath` enum with values `invite`, `request`, and `open`.
2. Add signed continuity telemetry in discovery responses for path handoff and wake-state resumptions.
3. Add deterministic reason metadata for connection blocks and trust transitions.
4. Add optional `continuity` fields for retry budget, stage age, and fallback indexer trace.

## Compatibility notes
- All additions are optional and additive.
- Do not reuse removed field numbers; explicitly reserve any field numbers removed during later cleanup.
- v19 execution will include concrete field assignments and wire compatibility verification.
