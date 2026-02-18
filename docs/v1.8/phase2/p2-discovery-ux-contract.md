# Phase 2 - Discovery UX Contract

Official discovery and join-funnel UX behavior must remain deterministic across clients that render `pkg/v18` results.

## Discoverability contract
- The directory view orders entries deterministically by `NodeID`.
- Each row includes at minimum `NodeID`, `Relay`, and `Endpoint`.
- Entries with duplicate `NodeID` are deduplicated by timestamp freshness.

## Trust warning contract
- Each warning uses stable `NodeID` + message pair:
  - `NodeID` references the affected directory row.
  - Message is deterministic text such as `response[<index>] signature mismatch`.
- Clients render all warning rows in the warning block and keep stable ordering.
- Warning text must be surfaced before navigation actions if merge anomalies are present.

## Join funnel contract
- Funnel state is one of:
  - `discovery`
  - `relay-handshake`
  - `completed`
- `Failed` is set when any stage call reports failure.
- `Attempts` increments only on successful `completed` transitions.
- Summary string format is stable and parseable: `stage=<...> attempts=<n> status=<ok|failed>`.

## Stage mapping
- `discovery` renders search results and warning summary.
- `relay-handshake` is the only pre-completion trust-gate transition.
- `completed` requires prior handshake call and non-failing state.

## Interop expectation
- Clients should classify user-facing actions from the same state fields and warning set so UX stays deterministic for invite, request, and open flows.
