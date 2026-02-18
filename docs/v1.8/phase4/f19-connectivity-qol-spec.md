# F19 Connectivity and QoL Specification

This package defines the connectivity path ladder and continuity behavior to be implemented in v19.

## Scope
- Specify deterministic decision paths from first discovery result to active join attempt.
- Define no-limbo UX invariants when discovery, index trust checks, and relay transitions are delayed.
- Define continuity behavior for wake/pause and handoff paths.

## Path ladder
1. Search local cache and stale-safety checks.
2. Query primary indexer and validate signed responses.
3. If no trustworthy primary answer, query fallback indexers.
4. Validate selected path against `JoinPath` policy.
5. Execute relay handshake path with clear reason classification.

## Reason taxonomy
- `indexer-unreachable`
- `signature-missing`
- `signature-mismatch`
- `path-untrusted`
- `rate-limited`
- `join-request-denied`
- `invite-only-required`

## No-limbo invariants
- The UI must always show one of:
  - `discovering`
  - `awaiting-trust`
  - `ready-for-join`
  - `blocked`
- Blocked states include a deterministic reason from the taxonomy.
- Successful completion must clear temporary blockers and return to idle/ready.

## Continuity contracts
- When a join is paused or app returns from standby, the client preserves last verified path and failure class.
- Reasoned continuity state must be resumable with at most one additional retry and bounded counter metadata.
- Failed attempts must remain visible so a user can continue from the same join intent path.
