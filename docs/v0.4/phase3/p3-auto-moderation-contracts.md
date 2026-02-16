# Phase 3 · P3 Auto-Moderation Contracts

## Purpose
Define trigger, action, bypass, and failure semantics for auto-moderation hooks operating under both Clear and E2EE modes as outlined in `TODO_v04.md` Phase 3.

## Contract
- Every trigger type (`rate`, `keyword`, `extensible hook`) specifies a precedence weight and required inputs; mode-aware gating asserts that E2EE channels only accept summary proofs, while Clear channels allow full evidence flows.
- Action contracts describe deterministic reason taxonomies (`warn`, `quarantine`, `block`, `escalate`), tie each to the triggering mode, and include any evidence expectations mandated by `VA-M5`.
- Bypass and failure semantics document when appeals or manual overrides can intervene; deterministic recovery states and failure reasons must be surfaced to `VA-M6` and traceable in the release dossier.
