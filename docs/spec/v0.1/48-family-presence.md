# 48 — Family: Presence (`/aether/presence/0.2.0`)

The Presence family handles real-time status indicators: online/offline states,
custom status strings, typing indicators, and activity context (active server
and channel). Presence is ephemeral — it is never stored beyond the relay
queue TTL and is always superseded by a fresher heartbeat.

## 1. Overview

Presence operates on a push model: each node that wishes to be visible MUST
proactively announce its state to directly-connected peers. There is no central
presence registry. Peers aggregate presence snapshots and apply a TTL-based
expiry to prevent stale indicators from persisting.

**Roles that participate:**

| Role | Participation |
|------|--------------|
| client | Announces its own presence; queries and displays peers' presence |
| relay | MUST NOT relay presence announcements (not store-and-forward) |
| bootstrap | MUST NOT relay presence announcements |
| archivist | Does not participate |

Relays and bootstrap nodes do not forward `presence.announce` messages. Presence
is inherently a live-connection feature. A client that reconnects after an
offline period MUST re-announce its full presence state immediately.

**Security mode:** Presence operations are unencrypted at the application layer
(presence metadata is not sensitive at the wire level). They are protected at
the transport level by Noise XX (see `30-transport-and-noise.md`). For Invisible
mode, a separate rule applies (see §3.3).

## 2. Capability requirements

| Capability | Role |
|-----------|------|
| `cap.presence` | Required on both initiator and responder for all presence operations |

No additional mode capability is required for presence. The `security_mode`
field in `PeerStreamRequest` MUST be `SECURITY_MODE_UNSPECIFIED` for this
family (presence is not mode-gated).

## 3. Operations

| Operation | Required caps | Direction | Description |
|-----------|--------------|-----------|-------------|
| `presence.announce` | `cap.presence` | initiator → responder | Push a `PresenceStateEntry` to a connected peer |
| `presence.query` | `cap.presence` | initiator → responder | Request the responder's current presence state; response is `PresenceStateEntry` or absent |

### 3.1 `presence.announce`

The initiator sends its current `PresenceStateEntry`. The responder stores the
entry in a local in-memory presence cache, keyed by `identity_id`. No
`PeerStreamResponse.payload` is returned for announce (acknowledgement is
implicit in the successful response with empty payload).

`PeerStreamRequest.payload = proto.Marshal(PresenceStateEntry)`.

Triggers that MUST cause a `presence.announce` to all directly-connected peers:

1. Status change: any transition among `online`, `idle`, `dnd`, `invisible`,
   `offline`.
2. Custom status string change (including clearing the custom status).
3. Typing indicator change (see §3.4).
4. Active server or channel change.
5. Heartbeat: exactly every 60 seconds, even if no state has changed.

Implementations MUST send a heartbeat on the 60-second tick regardless of any
recent change. Peers that do not receive a heartbeat within 90 seconds SHOULD
treat the sender as `offline`.

A node MUST send a final `presence.announce` with `state = PRESENCE_STATE_OFFLINE`
immediately before closing its last outbound stream (graceful shutdown). This
minimises the window before the 90-second TTL causes the peer to be marked
offline by remote nodes.

### 3.2 `presence.query`

The initiator requests the responder's current presence state. The responder
replies with its current `PresenceStateEntry` as `PeerStreamResponse.payload`.
If the responder does not support presence queries (e.g., it is a relay or
bootstrap node), it returns `MISSING_REQUIRED_CAPABILITY`. If the responder is
`invisible` and the querying peer is not in its friend list, the responder
returns a synthetic `PresenceStateEntry` with `state = PRESENCE_STATE_OFFLINE`
(see §3.3).

`PeerStreamRequest.payload` is empty (no request body needed).

`PeerStreamResponse.payload = proto.Marshal(PresenceStateEntry)` or empty if
the responder is not visible to the initiator.

### 3.3 Invisible mode

When `state = PRESENCE_STATE_INVISIBLE`:

- The node sends `presence.announce` only to peers in its `friends` bucket
  with `status = "accepted"`.
- To friend peers, it sends a `PresenceStateEntry` with
  `state = PRESENCE_STATE_OFFLINE` (not `invisible`). The invisible state is
  not transmitted; the node is indistinguishable from offline to all peers.
- To non-friend peers, the node does not send any announce. If queried via
  `presence.query`, it responds with `state = PRESENCE_STATE_OFFLINE`.
- The node continues to receive `presence.announce` messages from other peers
  normally.

Implementation note: The node's own local display (in the client UI) MAY show
`invisible` as the configured state while the transmitted state to all peers is
always `offline`.

### 3.4 Typing indicators

Typing indicators are conveyed within `presence.announce` via the `is_typing`
field of `PresenceStateEntry`.

Rules:

- A sender MUST set `is_typing = true` when the user has typed at least one
  character in a message composition box and has not sent the message yet.
- A sender MUST set `is_typing = false` when the user clears the composition
  box, sends the message, or has not typed for 5 seconds.
- **Debouncing**: A sender MUST NOT send more than one `presence.announce`
  with a changed `is_typing` value within a 2-second window.
- **Display timer**: The receiver MUST display the typing indicator for at
  most 5 seconds after receiving an `is_typing = true` announce. If no
  refresh arrives within 5 seconds, the indicator MUST be cleared.
- Typing indicators MUST NOT be persisted to disk.

The `PresenceDisseminationPolicy` proto captures the debounce and heartbeat
parameters:

| Field | Value |
|-------|-------|
| `debounce_seconds` | 2 |
| `min_publish_interval_seconds` | 2 |
| `heartbeat_interval_seconds` | 60 |
| `max_status_latency_seconds` | 90 |

## 4. Wire format details

### 4.1 `PresenceStateEntry` fields

`proto.Marshal(PresenceStateEntry)` with:

| Field | Type | Notes |
|-------|------|-------|
| `identity_id` | string | base58 libp2p peer ID |
| `state` | PresenceState | `ONLINE`, `IDLE`, `DND`, `INVISIBLE`, `OFFLINE` |
| `status` | string | optional custom status; max 128 UTF-8 chars |
| `last_updated_unix` | uint64 | unix milliseconds of the state change |
| `ttl_seconds` | uint32 | max age before entry is considered stale; MUST be 90 |
| `transition_reason` | PresenceTransitionReason | `APPLIED` or `IGNORED_BY_PRECEDENCE` |
| `status_version` | uint64 | monotonically increasing counter; receivers MUST discard entries with lower `status_version` than cached |
| `status_redacted` | bool | if `true`, `status` field is intentionally absent (privacy mode) |
| `published_at_unix` | uint64 | unix milliseconds when this announcement was generated |

`status_version` MUST be incremented by at least 1 on every `presence.announce`.
Receivers that receive an entry with `status_version <= cached_version` for the
same `identity_id` MUST discard the new entry (replay protection and
out-of-order delivery guard).

Additional context fields carried in the `data` JSON:

```json
{
  "identity_id": "<base58>",
  "is_typing": false,
  "active_server_id": "<uuid or null>",
  "active_channel_id": "<uuid or null>",
  "display_name": "<string; current display name>",
  "updated_at_unix_ms": "<int64; unix milliseconds>"
}
```

### 4.2 `presence.announce` full request JSON

```json
{
  "operation": "presence.announce",
  "protocol_id": "/aether/presence/0.2.0",
  "advertised_caps": ["cap.presence"],
  "required_caps": ["cap.presence"],
  "security_mode": "SECURITY_MODE_UNSPECIFIED",
  "request_id": "<uuid v4>",
  "payload": "<base64url(proto.Marshal(PresenceStateEntry))>",
  "data": {
    "identity_id": "<base58>",
    "is_typing": false,
    "active_server_id": null,
    "active_channel_id": null,
    "display_name": "<string>",
    "updated_at_unix_ms": 1234567890000
  }
}
```

### 4.3 `PresenceEventRecord`

The `PresenceEventRecord` proto is used internally to trace presence state
machine transitions. It is NOT transmitted over the wire; it is for local
diagnostics and conformance testing only.

| Field | Notes |
|-------|-------|
| `previous_state` | State before transition |
| `next_state` | State after transition |
| `event` | `IDLE_TIMEOUT`, `ACTIVITY`, or `DISCONNECT` |
| `reason` | `APPLIED` or `IGNORED_BY_PRECEDENCE` |
| `occurred_at` | unix seconds |
| `sequence` | monotonic counter |
| `debounce_until_unix` | earliest time next announce is permitted (unix seconds) |
| `throttle_until_unix` | earliest time next heartbeat tick is permitted (unix seconds) |
| `heartbeat` | `true` if this record was triggered by the 60-second heartbeat |

### 4.4 Idle detection

Clients SHOULD implement idle detection with a configurable threshold
(default: 5 minutes of no keyboard or mouse activity). On idle threshold
crossing:

1. Emit `PRESENCE_EVENT_IDLE_TIMEOUT`.
2. Transition from `ONLINE` to `IDLE`.
3. Send `presence.announce` with `state = PRESENCE_STATE_IDLE`.

On activity resumption:

1. Emit `PRESENCE_EVENT_ACTIVITY`.
2. Transition from `IDLE` to `ONLINE`.
3. Send `presence.announce` with `state = PRESENCE_STATE_ONLINE`.

`dnd` (do not disturb) is a user-explicit state, not an automatic transition.
It is set and cleared only by direct user action.

## 5. Security mode binding

`security_mode` in `PeerStreamRequest` MUST be `SECURITY_MODE_UNSPECIFIED` for
this family. Any other value MUST result in `MODE_INCOMPATIBLE`.

Presence data is visible at the transport level to the direct peer receiving the
announcement. It is NOT relayed or stored beyond the direct connection. For
Invisible mode, the protocol ensures the peer's real state is never transmitted.

## 6. State persistence

Presence state is **ephemeral** and MUST NOT be written to the SQLCipher
persistent store.

The local node stores its own presence state in memory:

```
localPresenceState {
    State:          PresenceState
    CustomStatus:   string
    IsTyping:       bool
    ActiveServerID: string
    ActiveChannelID: string
    StatusVersion:  uint64
    UpdatedAt:      time.Time
    NextHeartbeat:  time.Time
}
```

The remote peer cache is also in-memory:

```
presenceCache map[peerID]PresenceStateEntry  // keyed by identity_id
```

Entries in `presenceCache` expire after `ttl_seconds` (90 seconds) from
`published_at_unix`. Expired entries MUST be evicted before any UI render.
Presence state is rebuilt from heartbeats after a node restart; no warm-start
is required or permitted.

## 7. Error codes

| Code | Trigger |
|------|---------|
| `MISSING_REQUIRED_CAPABILITY` | Peer does not advertise `cap.presence` |
| `OPERATION_FAILED` | Malformed `PresenceStateEntry` (missing `identity_id`, invalid state enum, etc.) |
| `RATE_LIMITED` | Sender has exceeded the announce rate limit |

Rate limit: max 30 `presence.announce` operations per peer per minute. This
limit exists to prevent typing-indicator spam. The debounce rule (§3.4) ensures
well-behaved clients stay well under this limit.

Relay and bootstrap nodes MUST return `MISSING_REQUIRED_CAPABILITY` for any
presence operation since they do not advertise `cap.presence`.

## 8. Conformance

Conformance class: **W6** (Presence family conformance).

KATs in `pkg/spectest/presence/`:

- `presence_announce_online.json` — client announces `ONLINE`; receiver caches
  entry; verify `status_version` incremented.
- `presence_announce_idle.json` — client auto-transitions to `IDLE` after
  idle timeout; verify `PresenceEventRecord` emitted.
- `presence_announce_dnd.json` — user explicitly sets `DND`; verify announce sent.
- `presence_heartbeat.json` — 60-second heartbeat fires with no state change;
  verify announce sent with same state but incremented `status_version`.
- `presence_ttl_expiry.json` — no heartbeat received for 90 seconds; verify
  entry evicted from cache.
- `presence_invisible.json` — node configured `INVISIBLE`; verify `OFFLINE`
  sent to friend peers, nothing sent to non-friends; `presence.query` from
  non-friend returns `OFFLINE`.
- `presence_typing.json` — user types; `is_typing = true` sent; user pauses 5s;
  receiver clears indicator; sender sets `is_typing = false`.
- `presence_typing_debounce.json` — two typing changes within 2 seconds; verify
  only one announce sent.
- `presence_query.json` — `presence.query` returns current `PresenceStateEntry`.
- `presence_out_of_order.json` — lower `status_version` entry received after
  higher; verify discarded.
- `presence_relay_reject.json` — relay node receives `presence.announce`;
  verify `MISSING_REQUIRED_CAPABILITY` returned.

Implementations MUST pass all KATs in `pkg/spectest/presence/` to claim
Presence family conformance.
