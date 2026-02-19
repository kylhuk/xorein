# Phase 1 Archivist Selection Contract

- Advertisements expose `SpaceID`, operator identifier, availability flag, quota budget, and refresh metadata (TTL + last refresh timestamp).
- Clients refresh advertisements periodically and treat stale records as unavailable to keep determinism.
- Selection order is deterministic: prefer advertisements for the requested space, then higher remaining quota, then more recent refresh timestamps.
- Selection fails with deterministic reasons:
  - `NO_ARCHIVIST_AVAILABLE` when no fresh, policy-allowed, quota-satisfying advertisement exists.
  - `ARCHIVIST_POLICY_DENIED` when every candidate was rejected by policy filters.
