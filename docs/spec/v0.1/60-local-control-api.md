# 60 — Local Control API

The local control API is an HTTP interface that local clients (UIs, agents, scripts)
use to control and observe a running Xorein node. It is intentionally local-only;
remote access is rejected at the transport layer.

This document supersedes `docs/local-control-api-v1.md`.

## 1. Transport

### 1.1 Unix-like systems

The API server listens on a Unix-domain socket located at:

```
<data_dir>/xorein-control.sock
```

The socket path is overridden by `--control <path>`.

Clients MUST connect via the Unix socket. Non-loopback TCP connections are
rejected with HTTP 403 `forbidden` regardless of the bearer token.

### 1.2 Windows

On Windows, the API server listens on a loopback TCP port chosen at startup.
The bound address is written to:

```
<data_dir>/control.addr
```

Clients MUST read that file to discover the port. Only connections from
`127.0.0.1` and `::1` are accepted.

## 2. Authentication

Every request MUST include:

```
Authorization: Bearer <token>
```

The token is a 32-byte random value encoded as base64url (no padding),
generated at node startup and written to:

```
<data_dir>/control.token
```

Permissions: 0600 (owner read/write only). Missing or invalid token → HTTP 401
`unauthorized`.

## 3. Versioning

All stable endpoints are rooted at `/v1`. The API version is advertised in the
SSE `ready` event:

```
event: ready
data: {"version":"1"}
```

## 4. Error model

All errors return a JSON object:

```json
{"code": "<error_code>", "message": "<human-readable explanation>"}
```

### 4.1 Error codes

| Code | HTTP status | Meaning |
|------|------------|---------|
| `unauthorized` | 401 | Missing or invalid bearer token |
| `forbidden` | 403 | Non-local connection rejected |
| `method_not_allowed` | 405 | Wrong HTTP method for endpoint |
| `invalid_request` | 400 | Malformed or logically invalid request body |
| `not_found` | 404 | Resource or endpoint does not exist |
| `invalid_signature` | 400 | Invite or signed object has a bad signature |
| `expired_invite` | 400 | Deeplink invite is past its `expires_at` |
| `join_failed` | 400 | Could not join server (general failure) |
| `preview_failed` | 400 | Could not decode invite for preview |
| `backup_failed` | 500 | Identity backup could not be serialized |
| `unsupported` | 500 | Server-Sent Events not supported by responder |

## 5. Endpoints

### 5.1 State

#### `GET /v1/state`

Returns a full snapshot of local node state: identity, peers, servers, channels,
DMs, messages, voice sessions, friends, and relay queue.

**Response 200**: `Snapshot` JSON object (mirrors `pkg/node/service.go:Snapshot`).

---

### 5.2 Events

#### `GET /v1/events`

Opens a [Server-Sent Events](https://html.spec.whatwg.org/multipage/server-sent-events.html)
stream. The connection stays open until the client disconnects or the node shuts down.

**Response headers**:
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**First event** (sent immediately after connection):
```
event: ready
data: {"version":"1"}
```

**Subsequent events** — one per state change:

| `event` field | Trigger |
|--------------|---------|
| `peer` | Peer discovered, lost, or updated |
| `message` | Chat or DM message received |
| `channel` | Channel created or updated |
| `server` | Server joined or manifest changed |
| `voice` | Voice session join/leave/frame event |
| `dm` | DM record created |
| `friend` | Friend request or status change |
| `presence` | Presence status update |
| `notification` | New unread notification |
| `relay` | Relay queue entry added or drained |

Each event's `data` field is a JSON object specific to the event type.

---

### 5.3 Identity

#### `GET /v1/identities`

Returns the current identity (public key, peer ID, display name, bio).

**Response 200**: `Identity` JSON.

#### `POST /v1/identities`

Creates a new identity (destructive — replaces any existing identity).

**Request body**:
```json
{
  "display_name": "alice",
  "bio": "optional bio text"
}
```

**Response 201**: `Identity` JSON.

#### `GET /v1/identities/backup`

Exports the full identity (including private key material) as a JSON document
for offline backup.

**Response 200**: `application/json` backup document. Store this file securely;
it contains private key material.

#### `POST /v1/identities/restore`

Restores an identity from a backup document. Replaces the current identity.

**Request body**: the JSON backup document produced by `GET /v1/identities/backup`.

**Response 200**: `Identity` JSON of the restored identity.

---

### 5.4 Servers

#### `GET /v1/servers`

Lists all joined servers.

**Response 200**: `[]ServerRecord` JSON array.

#### `POST /v1/servers`

Creates a new server owned by the local identity.

**Request body**:
```json
{
  "name": "My Server",
  "description": "optional",
  "security_mode": "seal"
}
```

`security_mode` must be one of: `clear`, `seal`, `tree`, `crowd`, `channel`.

**Response 201**: `ServerRecord` JSON.

#### `POST /v1/servers/join`

Joins a server via a signed deeplink invite.

**Request body**:
```json
{"deeplink": "xorein://invite/<base64url>"}
```

**Response 200**: `ServerRecord` JSON.

Errors: `expired_invite`, `join_failed`, `invalid_signature`.

#### `POST /v1/servers/preview`

Decodes a deeplink invite and returns server metadata without joining.

**Request body**:
```json
{"deeplink": "xorein://invite/<base64url>"}
```

**Response 200**:
```json
{
  "server_id": "...",
  "name": "My Server",
  "description": "...",
  "owner_peer_id": "...",
  "expires_at": "RFC3339Nano",
  "security_mode": "seal"
}
```

Errors: `expired_invite`, `preview_failed`, `invalid_signature`.

#### `POST /v1/servers/{serverID}/channels`

Creates a new text or voice channel inside the server.

**Request body**:
```json
{"name": "general", "voice": false}
```

Set `voice: true` for a voice channel.

**Response 201**: `ChannelRecord` JSON.

---

### 5.5 Presence

#### `GET /v1/presence`

Returns the presence map for all known peers.

**Response 200**:
```json
{
  "peers": {
    "<peer_id>": {
      "status": "online|away|busy|offline",
      "status_text": "...",
      "typing_in_scope": "<scope_id or empty>",
      "updated_at": "RFC3339Nano"
    }
  }
}
```

---

### 5.6 Notifications

#### `POST /v1/notifications/search`

Searches notifications by filter.

**Request body**:
```json
{
  "server_id": "...",
  "scope_type": "channel|dm",
  "scope_id": "...",
  "unread_only": true,
  "limit": 50
}
```

All fields optional. `limit` defaults to 50, max 200.

**Response 200**: `{"notifications": [...]}`.

#### `GET /v1/notifications/summary`

Returns unread notification counts by scope.

**Response 200**:
```json
{
  "total_unread": 12,
  "by_server": {
    "<server_id>": {"unread": 5, "mentions": 1}
  },
  "dms_unread": 7
}
```

#### `POST /v1/notifications/read`

Marks notifications as read up to a given sequence point.

**Request body**:
```json
{
  "server_id": "...",
  "scope_type": "channel|dm",
  "scope_id": "...",
  "read_through_message_id": "..."
}
```

**Response 200**: updated read-through state for the scope.

---

### 5.7 Mentions

#### `POST /v1/mentions/search`

Searches messages that mention the local identity.

**Request body**:
```json
{
  "server_id": "...",
  "scope_id": "...",
  "limit": 50
}
```

All fields optional. `limit` defaults to 50, max 200.

**Response 200**: `{"mentions": [...]}`.

---

### 5.8 Peers

#### `POST /v1/peers/manual`

Adds a manual peer address (libp2p multiaddr).

**Request body**:
```json
{"address": "/ip4/1.2.3.4/tcp/9000/p2p/12D3Koo..."}
```

**Response 204**: no body.

#### `DELETE /v1/peers/manual`

Removes a manual peer address.

**Request body**: same as POST.

**Response 204**: no body.

---

### 5.9 Relays

#### `POST /v1/relays`

Registers a relay node (Circuit Relay v2 multiaddr).

**Request body**:
```json
{"multiaddr": "/ip4/1.2.3.4/tcp/1337/p2p/12D3Koo.../p2p-circuit"}
```

**Response 204**: no body.

---

### 5.10 Direct Messages

#### `GET /v1/dms`

Lists all DM records.

**Response 200**: `[]DMRecord` JSON array.

#### `POST /v1/dms`

Opens a DM with a peer (creates a `DMRecord` if one does not exist).

**Request body**:
```json
{"peer_id": "12D3Koo..."}
```

**Response 201**: `DMRecord` JSON.

#### `POST /v1/dms/{dmID}/messages`

Sends a message in the DM. Message body is encrypted using Seal mode before
delivery.

**Request body**:
```json
{"body": "Hello!"}
```

**Response 201**: `MessageRecord` JSON.

---

### 5.11 Channel messages

#### `POST /v1/channels/{channelID}/messages`

Sends a message to a server channel. Body is encrypted using the channel's
active security mode.

**Request body**:
```json
{"body": "Hello channel!"}
```

**Response 201**: `MessageRecord` JSON.

---

### 5.12 Messages

#### `POST /v1/messages/search`

Full-text and filter search across message history.

**Request body**:
```json
{
  "query": "optional text search",
  "scope_type": "channel|dm",
  "scope_id": "...",
  "server_id": "...",
  "sender_peer_id": "...",
  "before": "RFC3339Nano",
  "after": "RFC3339Nano",
  "limit": 50
}
```

All fields optional. `limit` defaults to 50, max 200.

**Response 200**:
```json
{
  "messages": ["<message_id>", ...],
  "results": [<MessageRecord>, ...]
}
```

#### `PATCH /v1/messages/{messageID}`

Edits the body of a locally-originated message.

**Request body**:
```json
{"body": "edited text"}
```

**Response 200**: updated `MessageRecord` JSON.

#### `DELETE /v1/messages/{messageID}`

Deletes a locally-originated message.

**Response 204**: no body.

---

### 5.13 Voice

All voice endpoints operate on `{channelID}` — a voice channel ID within a server.

#### `POST /v1/voice/{channelID}/join`

Joins a voice session in the given channel.

**Request body**:
```json
{"muted": false}
```

**Response 204**: no body.

#### `POST /v1/voice/{channelID}/leave`

Leaves the voice session.

**Request body**: empty (`{}`).

**Response 204**: no body.

#### `POST /v1/voice/{channelID}/mute`

Updates the local mute state without leaving.

**Request body**:
```json
{"muted": true}
```

**Response 204**: no body.

#### `POST /v1/voice/{channelID}/frames`

Submits a media frame (SFrame-encrypted audio/video payload) to the node for
forwarding to the SFU or direct peers. In v0.1 this carries raw binary data
that the WebRTC layer will encapsulate.

**Request body**:
```json
{"data": "<base64url SFrame payload>"}
```

**Response 204**: no body.

---

### 5.14 Friends

#### `GET /v1/friends`

Lists accepted friends.

**Response 200**: `{"friends": [<FriendRecord>, ...]}`.

#### `POST /v1/friends/requests`

Sends a friend request to a peer address.

**Request body**:
```json
{"peer_addr": "/ip4/1.2.3.4/tcp/9000/p2p/12D3Koo..."}
```

**Response 201**: `FriendRequestRecord` JSON.

#### `PUT /v1/friends/requests/{requestID}`

Performs an action on a pending friend request.

**Request body**:
```json
{"action": "accept|decline|cancel|block"}
```

**Response 200**: updated `FriendRequestRecord` JSON.

#### `DELETE /v1/friends/{friendID}`

Removes a friend.

**Response 204**: no body.

---

## 6. Request size limits

| Limit | Value |
|-------|-------|
| Maximum request body size | 1 MiB |
| Maximum `body` field in messages | 64 KiB |
| Maximum `data` field in voice frames | 8 MiB |

Requests exceeding these limits MUST be rejected with HTTP 413 and
`invalid_request`.

## 7. Security considerations

- The control API has no rate limiting; it is trusted as a local process.
- The bearer token MUST be treated as a secret equivalent to a session cookie.
  Clients MUST NOT log or expose the token.
- On node shutdown, the control socket is unlinked. Stale socket files from
  unclean exits MUST be unlinked before the next node start.
- Control API actions that send messages invoke the cryptographic path of the
  relevant security mode; the control API never handles plaintext message
  bodies directly on behalf of remote peers.
