# Phase 2: Voice UX Contract

## Controls
- Join/Leave buttons toggle deterministically via `CallState.ActionLabel()`.
- Mute/Deafen toggles follow `CallState.MuteLabel()` and `CallState.DeafLabel()` semantics.
- Selected device is surfaced via `DeviceLabel()` and defaults to `"Select Device"` when unset.

## State Feedback
- Quality badge is derived from `QualityBadge(score)` with HD/SD/Degraded tiers.
- `NoLimboMessage` communicates reconnection hints and ensures no ambiguous limbo states during recovery.
- Recovery-first guidance surfaces while `RecoveryHint` is true to avoid letting users get stuck.
