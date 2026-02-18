# Phase 2 — Enforcement UX Contract

Official clients consume `pkg/v17/ui.StatusSignal` as the enforcement status feed. The contract guarantees:

- `Mode` is either `relaxed` or `strict`, and clients render the difference with explicit banners.
- `Sequence` is monotonic and increments whenever enforcement conditions change.
- `Reason` includes deterministic strings such as `lockdown active`, `slow mode`, or `bans`, allowing clients to align messaging.
- `TrustWarning` surfaces when the enforcement proof is not yet verified (`unverified enforcement`).

Any UI layout must pair the Summary string (`Summary()`) with accessible descriptors and expose `IsStrict()` to drive enforcement badges.
