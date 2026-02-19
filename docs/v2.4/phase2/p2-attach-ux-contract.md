# P2-T1 Attach UX Contract

This contract documents the deterministic harmolyn attach experience mandated by `P2-T1` so that both the docs-based checklist and any UI/UX implementation reference the same telemetry-friendly state machine. All failure reason codes below reuse the taxonomy from `docs/v2.4/phase0/p0-error-taxonomy.md` so phase0 tooling can correlate UI events with the same `EV-v24-*` reason buckets.

## ST1 – Attach bootstrap sequence
1. **Probe**: harmolyn inspects the configured socket/piped transport via the localized probe abstraction (phase0-compliant instrumentation reports missing sockets as `SOCKET_PERMISSION_DENIED`).
2. **Spawner**: if no daemon is running the coordinator invokes the daemon starter, waits for readiness, and ensures the probe reports a running state (any startup error produces `DAEMON_START_FAILED`).
3. **Handshake**: harmolyn opens the session handshake, records the returned session token, and surfaces `DAEMON_INCOMPATIBLE`/`AUTH_FAILED` for deterministic mismatch/auth errors.
4. **Caching**: once a token is recorded the coordinator returns success until the daemon restarts, at which point it gracefully clears the cached token and retriggers the bootstrap steps.

## ST2 – Failure state contract
| Failure reason | User-facing explanation | Next action | Telemetry anchor |
| --- | --- | --- | --- |
| `DAEMON_START_FAILED` | The daemon could not start or never reported `Running`. | Retry the launcher (retry button). | Maps directly to `phase0` `DAEMON_START_FAILED` bucket so evidence entries can link to `EV-v24-G3-###`. |
| `DAEMON_INCOMPATIBLE` | The running daemon speaks an unsupported API version. | Repair (prompt to reinstall/upgrade the daemon). | Telemetry reuses the same reason code from `docs/v2.4/phase0/p0-error-taxonomy.md`. |
| `AUTH_FAILED` | Handshake tokens were rejected or expired. | Reset the session (clear cache and reauthenticate). | Aligns with the `AUTH_FAILED` reason used by the local API threat model. |
| `SOCKET_PERMISSION_DENIED` | File-permission/Acl policy prevents harmolyn from touching the socket. | Open logs/inspect permissions before retrying. | Surface as `SOCKET_PERMISSION_DENIED` alongside the same phase0 reason. |

Each failure bubble should display the explicit next action text (Retry / Repair / Reset / Open Logs) so QA scripts can validate deterministic labels.

## ST3 – Graceful detach + reattach
1. When the daemon signals a restart (exit event, missing socket, new PID) harmolyn calls `Detach()` to purge the cached token.
2. `Reattach()` reruns the bootstrap sequence (probe, spawn, handshake) and emits the same deterministic failure reason codes above for any regression.
3. The UX must never show a “limbo” state between detach and reattach; there should always be either a success state or one of the four deterministic failures with their canonical next action.

This contract is intentionally scaffolded so the CLI/UI authoring the attach flow only needs to implement the exported `pkg/v24/harmolyn/attach` interfaces and will automatically observe the deterministic reason/action pairs described here.
