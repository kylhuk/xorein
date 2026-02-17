# Phase 2 · P2 Policy Versioning Governance

## Purpose
Capture the immutability, migration, and rollback governance contracts for moderation policies described in `TODO_v04.md` Phase 2.

## Requirements
- Each policy version record is immutable once published; mutating a published version results in a deterministic rejection reason that is recorded in `VA-M1` evidence.
- Migration contracts must explicitly enumerate compatible versions and include a `PolicyMigrationPlan` that lists compatible major/minor pairs before progress to V4-G2.
- Rollback entries are allowed only when the fallback version is listed in the `PolicyTrace` history and includes a deterministic ``fallback-reason`` string; this history is audited under `VA-M3`.
