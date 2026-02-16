# Phase 4 · P4 Audit Traceability

## Purpose
Capture signed audit entry expansion, policy linkage, and deterministic visibility semantics described in `TODO_v04.md` Phase 4.

## Contract
- Signed audit entries must include policy version references, triggering action metadata, and authorized visibility scopes so that `VA-A1` and `VA-A2` artifacts can reconstruct enforcement lineage.
- Policy, auto-mod, and manual moderator actions are mapped to a coverage matrix (`VA-A3`) that lists which roles and channel modes can see each entry.
- Mode change audit events (`VA-A4`) reference both the originating channel `SecurityMode` and the resulting `ModeChange` reasoning to keep disclosures deterministic.
