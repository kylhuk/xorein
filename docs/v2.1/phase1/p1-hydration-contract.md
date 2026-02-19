# Phase 1 Hydration Contract (v2.1)

- **Ordering**: Hydration orders events by timestamp, then by canonical identity (`ChannelID:MessageID:SenderID`) to ensure deterministic presentation even when clocks drift.
- **Pagination**: Each page exposes a cursor that encodes the next index; cursors persist across restarts, and page sizes are stable so downstream code can rely on deterministic boundaries.
- **Empty-history states**:
  - `no local history yet` indicates no event has ever been ingested.
  - `history cleared locally` indicates the user triggered the clear-local-history action and the timeline was wiped.
  - `history missing (backfill not supported in v21)` means a channel existed previously but no local events are available because backfill is intentionally out of scope.

The hydration contract requires both a deterministic ordering function and explicit state flags so UX teams can render precise banners instead of ambiguous placeholders.
