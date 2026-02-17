# Roadmap Evidence Index Template

Use this template for `docs/vX.Y/phase5/p5-evidence-index.md`.

## Evidence ID format
- `EV-vNN-GX-###`
  - `NN`: roadmap version number (for example, `11`, `20`).
  - `GX`: gate identifier (for example, `G3`).
  - `###`: three-digit sequence (`001`, `002`, ...).

## Rules
- Every mandatory command in the version TODO must appear at least once.
- Every promoted gate must reference at least one evidence ID.
- `OutputPath` must be repository-relative and immutable once recorded.

## Evidence table

| EvidenceID | Gate | Command | OutputPath | TimestampUTC | Owner | Result | Checksum | Notes |
|---|---|---|---|---|---|---|---|---|
| EV-vNN-GX-001 | G0 | <command> | <path> | 2026-01-01T00:00:00Z | <role/name> | pass | <sha256> | |
