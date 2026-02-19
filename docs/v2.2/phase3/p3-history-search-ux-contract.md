# P3-History Search UX Contract

This document describes the UX obligations for v22 history coverage:
- label searched ranges with coverage windows
- explain missing windows vs zero results
- provide deterministic next actions (start backfill, adjust range, cancel)

Coverage state is represented by the local window that was searched (`LocalStart`, `LocalEnd`) plus an explicit list of coverage gaps. Each `CoverageGap` carries its start/end range and a `MissingHistoryReason` (`no_local_history`, `missing_backfill`, or `adjust_range`). The coverage label must surface both the searched window and every gap summary so people understand which ranges still need work.

The UX should map each gap reason to a deterministic action:
1. `missing_backfill` or `no_local_history` → prompt `start_backfill` so users can request missing segments.
2. `adjust_range` → prompt `adjust_range` so users know the requested range is outside the local window.
3. No gaps reported → prompt `cancel_backfill`, signaling the coverage question is resolved.

Action labels must remain consistent (`Start backfill`, `Adjust search range`, `Cancel backfill`) so downstream copy can reuse the same phrasing without branching logic.
