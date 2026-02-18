# F13 Text Channel and Chat Flow Spec (v1.3 planning)

## Status
Specification artifact only. No v1.3 runtime completion claim in v1.2.

## Channel contract
- Channel identifiers are immutable once published.
- Message ordering is monotonic per channel timeline segment.
- Read markers are per-identity and per-channel.

## Send/receive states
### Send
- `compose` -> `queued` -> `sent` -> `delivered` -> `read`

### Failure
- `send-failed-transient`
- `send-failed-policy`
- `send-failed-auth`

### Recovery
- Retry from `send-failed-transient` preserves draft and reply context.
- `send-failed-policy` is non-retriable until policy state changes.

## Deterministic reason taxonomy
- User-visible and diagnostics strings must share reason IDs.
- Degraded and recovery reasons are required for each terminal failure branch.

## Planned vs implemented
- This file is a planning contract for v1.3 implementation.
