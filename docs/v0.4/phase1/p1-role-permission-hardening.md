# Phase 1 · P1 Role & Permission Hardening

## Purpose
Document the deterministic contracts that guarantee custom-role CRUD, permission evaluation, and channel override merge semantics claimed in `TODO_v04.md`.

## Deterministic obligations
- Role lifecycle operations must produce stable ordering results when conflicting updates arrive; every transition emits a `RoleAssignment` diff with priority ordering.
- Permission evaluation combines base-role permissions with channel overrides per the ordered merge policy described in `VA-R2`; the helper must return the same effective permission set for any permutation of concurrently arriving deltas.
- Security mode interactions described in Phase 1 expectations rely on the `SecurityMode` schema, so role operations must annotate the `ModeChange` target and reason class.
