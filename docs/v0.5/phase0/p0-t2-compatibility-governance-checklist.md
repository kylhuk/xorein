# Phase 0 · P0-T2 Compatibility & Governance Checklist

## Purpose
Document the additive-only protobuf requirements and major-change trigger guardrails that make V5-G0 gateable while preserving planned-vs-implemented discipline.

## Contract items
- Enumerate every proto message or field addition implied by the Phase 1–5 tasks listed in Section 6 of `TODO_v05.md`, note the gating rationale from Sections 5 and 6, and link each entry to its `VA-*` artifact.
- Capture the major-change trigger matrix described in P0-T2-ST2: list the downgrade-negotiation evidence, new multistream IDs, AEP obligations, and multi-implementation validation checkpoints that any breaking proposal must supply before approval.
- Reference the open-decision register in Section 10 to prove each unresolved item stays labeled `Open` and is revalidated during V5-G6; the compatibility summary must also explain why no v0.6+ hardening scope is imported per Section 3.
