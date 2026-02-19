# P2 UX Contract (Harmolyn)

- **Timeline empty states**: when the store has history coverage but no new messages, the banner should call out the last indexed timestamp (see `EmptyTimelineBanner`). When the timeline is truly empty, it must invite the user to select a channel or import history explicitly.
- **Search coverage**: every search result view surfaces the coverage label (`CoverageLabel`) derived from the search index. The label pairs a machine-readable status with a human-friendly summary so users can distinguish between `COVERAGE_FULL`, `COVERAGE_PARTIAL`, and `COVERAGE_EMPTY` states.
- **Retention controls**: display the configured retention window and entry cap via `RetentionSummary`. Clearing history prompts `ClearHistoryConfirmation`, mentioning the affected channel (or `all channels` when no channel is selected) and warning that the action is irreversible.
- **Headless safety**: the Harmolyn package exposes these helpers without requiring a GUI framework (no Gio dependency) so that automated scenarios can evaluate UX contracts programmatically.
