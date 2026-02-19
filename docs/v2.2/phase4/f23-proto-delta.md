# F23 Proto Delta

This document now captures the as-built proto delta that closed Phase 5 gate G7. The suggested optional fields and enums have been exercised by the Buf lint/breaking runs listed below, so the review notes are grounded in recorded outputs rather than placeholders.

## Purpose

Capture the wire-format questions introduced by the history hardening work so proto stewards can review compatibility, optional extensions, and any telemetry hooks before gating.

## Final delta outline

1. Optional coverage-gap metadata (`CoverageGapStart`, `CoverageGapEnd`, `GapReason`) now propagate through the history manifest and are gated under the replay verification controls referenced in this spec.
2. The `ResyncAnchor` indicator stays optional and scoped to replay/resync requests so older clients ignore it safely.
3. `HardeningStatus` was added as an enum with default `HARDENING_STATUS_UNSPECIFIED` so clients that do not adopt the enum still decode the messages compatibly.

## Compatibility & governance notes

- All new fields stay optional and occupy new tag numbers, preserving additive behavior.
- `HardeningStatus`, scoped to `v22.history`, defaults to `HARDENING_STATUS_UNSPECIFIED` so legacy clients ignore the indicator.
- Gate G7 (proto compatibility) is now closed with the Buf commands listed in _Evidence_ below; the `DEFAULT` category warning is noted but non-fatal.

## Gate mapping & evidence

| Gate | Steward | Evidence |
| --- | --- | --- |
| G7 | Protocol review board | `EV-v22-G7-001`, `EV-v22-G7-002` (Buf lint/breaking outputs); the lint run logged a `DEFAULT` category deprecation warning but still exited cleanly.
| G8 | Evidence bundle curator | `EV-v22-G8-001`..`EV-v22-G8-004` (build/test suite + `make check-full`).

Each gate review now points at the artifacts listed above rather than placeholders.

## Evidence summary

- `EV-v22-G7-001`: `artifacts/generated/v22-evidence/buf-lint.txt` – Buf lint run with compatibility notes and the `DEFAULT` category deprecation warning explained in the log.
- `EV-v22-G7-002`: `artifacts/generated/v22-evidence/buf-breaking.txt` – Buf breaking check output.
