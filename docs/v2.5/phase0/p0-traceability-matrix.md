# Traceability Matrix (G0)

| Requirement | Artifact | Notes |
| --- | --- | --- |
| BlobRef manifest definition + verification rules (ST1) | `docs/v2.5/phase1/p1-blobref-as-built.md`, `pkg/v25/blobref/*` | Links back to `docs/v2.4/phase4/f25-acceptance-matrix.md` columns for hash/size/mime.
| Provider refusal reasons + limited mime support (ST3) | `pkg/v25/blobproto/*`, `docs/v2.5/phase2/p2-relay-boundary-regressions.md` | Ensures provider APIs match locked refusal list.
| Anti-enumeration gating for private Spaces (ST1 + ST4) | `tests/e2e/v25/relay_boundary_*` | Validated by regression spec definitions in Phase 2.
| Asset coverage confirmation (ST4) | `docs/v2.5/phase3/p3-assets-contract.md`, `pkg/v25/assets/*` | Maps attachments, avatars, icons, emojis to BlobRef-backed paths.
| MIME & size tier freeze (ST3) | `docs/v2.5/phase0/p0-scope-lock.md`, `docs/templates/roadmap-gate-checklist.md` | Shared reference for gating mobile/desktop tiers.

> Evidence placeholder: `EV-v25-G0-004` documents the completed traceability matrix and reviewer sign-off.
