# 41 — Family: Chat (`/aether/chat/0.1.0`)

This document specifies the Chat family, which handles server-scoped message
delivery, channel management, and server joining.

## 1. Overview

The Chat family provides the primary channel-based communication substrate.
It is modeled on a server-centric topology where a server owner node acts as
the authority for channel structure, and members send messages scoped to a
channel within a server.

**Roles that use this family:**
- **Client**: All operations; initiates `chat.send` and `chat.join`.
- **Relay**: Routes `chat.send` deliveries via the Peer family; does not
  directly participate in Chat family operations.
- **Bootstrap**: Does not participate in Chat family operations.
- **Archivist**: MAY receive `chat.send` deliveries to archive; serves
  historical messages via the Sync family (`44-family-sync.md`).

**Security modes:** Chat family operations are bound to the server's negotiated
security mode. Seal, Tree, Crowd, Channel, and Clear modes are all applicable
(see §5).

**Protocol ID:** `/aether/chat/0.1.0`

**Required capabilities:**
- `cap.chat` — MUST be present in `advertised_caps` of both parties.
- `cap.peer.delivery` — MUST be present for `chat.send` to forward a Delivery.
- `cap.peer.join` + `cap.manifest` — MUST be present for `chat.join`.

## 2. Capability requirements

| Capability | Required for |
|------------|-------------|
| `cap.chat` | All Chat family operations |
| `cap.peer.delivery` | `chat.send` (Delivery forwarding) |
| `cap.peer.join` | `chat.join` (server join) |
| `cap.manifest` | `chat.join` (manifest retrieval as part of join response) |

When a node initiates a Chat family stream, it MUST set `required_caps` to
include at minimum `cap.chat`. The `cap.peer.delivery` requirement SHOULD also
be declared for `chat.send` operations. If the responder lacks a required
capability, the stream MUST fail with `MISSING_REQUIRED_CAPABILITY`.

## 3. Operations

| Operation | Required caps | Direction | Request payload type | Response payload type | Description |
|-----------|--------------|-----------|---------------------|----------------------|-------------|
| `chat.send` | `cap.chat`, `cap.peer.delivery` | initiator → server owner | `ChatSendRequest` | `ChatSendResponse` | Send a Delivery to a server scope |
| `chat.join` | `cap.chat`, `cap.peer.join`, `cap.manifest` | initiator → server owner | `JoinRequest` | `JoinResponse` | Join a server via a signed Invite deeplink |

The `operation` field in `PeerStreamRequest` MUST be one of the strings above.

## 4. Wire format details

All payloads are JSON-encoded and stored in `PeerStreamRequest.payload` /
`PeerStreamResponse.payload`.

### 4.1 `chat.send`

**Request:** `ChatSendRequest` — JSON-encoded:

```
{
  "delivery": { /* Delivery JSON object */ }
}
```

The `Delivery` struct (from `pkg/node/wire.go`) carried in `chat.send` MUST
have the following fields set:

```
{
  "id":                string,    // UUID v4; idempotency nonce
  "kind":              string,    // Delivery kind; see §4.1.1
  "scope_id":          string,    // channel ID this message targets
  "scope_type":        string,    // "channel"
  "server_id":         string,    // server this channel belongs to
  "sender_peer_id":    string,    // initiator's peer ID
  "sender_public_key": string,    // initiator's Ed25519 public key, base64url
  "recipient_peer_ids":[string],  // member peer IDs; empty means all server members
  "body":              string,    // see §4.1.1 for per-kind rules
  "data":              bytes,     // optional; kind-specific supplementary data (JSON)
  "created_at":        string,    // RFC3339Nano timestamp
  "signature":         string     // hybrid signature; see 02-canonical-envelope.md §3.3
}
```

**Response:** `ChatSendResponse` — JSON-encoded:

```
{
  "accepted":    bool,
  "delivery_id": string,   // mirrors delivery.id on success
  "error":       string    // human-readable; omitted on success
}
```

#### 4.1.1 Delivery kinds for `chat.send`

| `kind` | `body` content | `data` content | Notes |
|--------|---------------|----------------|-------|
| `channel_message` | Encrypted (or clear) message text | Absent | Primary message kind |
| `channel_create` | JSON of `ChannelRecord` (see §4.1.2) | Absent | Create a new channel; sender MUST be server owner |
| `channel_delete` | Channel ID string | `{"channel_id": "..."}` | Delete a channel; sender MUST be server owner |
| `message_edit` | Encrypted new text (same encoding as body) | `{"message_id": "..."}` | Edit an existing message |
| `message_delete` | Empty string `""` | `{"message_id": "..."}` | Soft-delete an existing message |
| `channel_topic_update` | New topic text (encrypted or clear) | `{"channel_id": "..."}` | Update channel topic; sender SHOULD be owner or moderator |

#### 4.1.2 `ChannelRecord` in `channel_create` body

When `kind` is `channel_create`, the `body` field is a JSON-encoded
`ChannelRecord` (from `pkg/node/types.go`):

```
{
  "id":         string,   // UUID v4; assigned by creator
  "server_id":  string,
  "name":       string,
  "voice":      bool,
  "created_at": string    // RFC3339Nano
}
```

In E2EE modes (Seal/Tree/Crowd/Channel), this struct MUST be AEAD-encrypted
before placement in `body`, and decoded as ciphertext by recipients who hold
the appropriate session key.

#### 4.1.3 Duplicate delivery handling

The receiver MUST maintain a deduplication set (`deliveries` bucket, keyed by
`delivery.id`). If a `delivery.id` is already in the set, the receiver MUST
return `accepted: true` without re-processing (idempotent acknowledgement).
This prevents double-application of edits and deletes.

### 4.2 `chat.join`

**Request:** `JoinRequest` — JSON-encoded using the `JoinRequest` struct from
`pkg/node/wire.go`:

```
{
  "invite": {
    "server_id":       string,
    "owner_peer_id":   string,
    "owner_public_key":string,
    "server_addrs":    [string],
    "bootstrap_addrs": [string],
    "relay_addrs":     [string],
    "manifest_hash":   string,   // 32-char base64url prefix; see 02-canonical-envelope.md §4
    "expires_at":      string,   // RFC3339Nano; zero means no expiry
    "signature":       string    // hybrid signature over canonical Invite
  },
  "requester": {
    "peer_id":    string,
    "role":       string,
    "addresses":  [string],
    "public_key": string
  },
  "capabilities": [string]   // capability flags the joining peer supports
}
```

The Invite struct MUST be verified by the server owner per
`02-canonical-envelope.md §3.2` before processing the join. If the invite has
expired (`expires_at` is non-zero and in the past), the server MUST reject with
`OPERATION_FAILED` and message `"invite expired"`.

**Response:** `JoinResponse` — JSON-encoded using the `JoinResponse` struct from
`pkg/node/wire.go`:

```
{
  "manifest": { /* Manifest JSON */ },
  "server":   { /* ServerRecord JSON */ },
  "channels": [ { /* ChannelRecord JSON */ } ],
  "history":  [ { /* MessageRecord JSON */ } ]
}
```

Notes:
- `manifest` carries the current signed server manifest (see `42-family-manifest.md`).
- `server` carries the full `ServerRecord` including the member list.
- `channels` is the current channel list for the server.
- `history` carries up to `history_retention_messages` recent messages (from
  the server's configured `HistoryLimit`; default 32). In E2EE modes, message
  bodies are ciphertext; recipients decrypt locally.
- The server owner MUST add the joining peer to `server.members` and MUST
  persist this change before sending the response.
- The deeplink format for invites is:
  `aether://join/<server_id>?invite=<base64url(Invite JSON)>`
  Parse and verify per `pkg/node/wire.go:ParseDeeplink`.

## 5. Security mode binding

The Chat family security mode is determined by the server's `security_mode`
field in its `Manifest` and `ServerRecord`. Mode is set at server creation and
SHOULD NOT change without a full manifest rotation.

| Security mode | `body` field encoding | Notes |
|--------------|----------------------|-------|
| `seal` | AEAD ciphertext (base64url, no padding) | 1:1 E2EE; not standard for channel messages; use for DM |
| `tree` | AEAD ciphertext (MLS-encrypted) | Small-group E2EE; default for new servers |
| `crowd` | AEAD ciphertext (sender-key epoch) | Large-scale E2EE |
| `channel` | AEAD ciphertext (broadcast epoch key) | Server-wide E2EE |
| `clear` | UTF-8 plaintext | MUST be explicitly opted-in; UI MUST display label |

For all non-Clear modes:
- The `body` field of every `Delivery` MUST be an AEAD ciphertext encoded as
  base64url (no padding). Receivers MUST reject plaintext bodies in non-Clear
  scopes with `OPERATION_FAILED`.
- Relay nodes MUST enforce the relay opacity invariant (`40-family-peer.md §4.4`)
  and MUST reject `relay.store` requests that carry non-encrypted bodies for
  non-Clear scopes.
- Senders MUST NOT log or expose the plaintext `body` after encryption.

For Clear mode:
- The `body` field is UTF-8 text, transmitted as-is.
- The UI MUST display a visible label (e.g., "This server uses Clear mode —
  messages are readable by infrastructure") on every message in a Clear scope.
- Clear mode conversations MUST NOT be silently downgraded to from an E2EE mode.

The security mode for a `chat.send` delivery MUST match the server's declared
`security_mode`. If the initiator includes a `security_mode` field in
`PeerStreamRequest` that conflicts with the server's mode, the receiver MUST
reject with `MODE_INCOMPATIBLE`.

## 6. State persistence

| State bucket | Key type | Value type | Description |
|-------------|----------|-----------|-------------|
| `servers` | `server_id` (string) | `ServerRecord` (JSON) | Server metadata, member list, channels, manifest |
| `messages` | `message_id` (string) | `MessageRecord` (JSON) | Chat message history |
| `deliveries` | `delivery_id` (string) | `struct{}` | Deduplication set for accepted deliveries |

Go types from `pkg/node/types.go`:

- `ServerRecord`: `ID`, `Name`, `Description`, `OwnerPeerID`, `SecurityMode`,
  `OfferedSecurityModes []string`, `CreatedAt`, `UpdatedAt`, `Members []string`,
  `Channels map[string]ChannelRecord`, `Manifest`, `Invite string`
- `ChannelRecord`: `ID`, `ServerID`, `Name`, `Voice bool`, `CreatedAt`
- `MessageRecord`: `ID`, `ScopeType`, `ScopeID`, `ServerID`, `SenderPeerID`,
  `Body`, `CreatedAt`, `UpdatedAt`, `Deleted bool`

`MessageRecord.ScopeType` MUST be `"channel"` for Chat family messages.
`MessageRecord.ScopeID` MUST be the channel ID. `MessageRecord.Body` stores the
received body verbatim (ciphertext in E2EE modes; plaintext in Clear mode).

The `deliveries` bucket acts as the idempotency set. Entries in this bucket have
no associated value; only key presence is checked.

## 7. Error codes

| Code | Trigger |
|------|---------|
| `INVITE_EXPIRED` | `chat.join` received an expired Invite (`expires_at` in the past) |
| `INVITE_INVALID` | `chat.join` Invite signature failed verification |
| `INVITE_SERVER_MISMATCH` | `chat.join` Invite `server_id` does not match this node's server |
| `MODE_INCOMPATIBLE` | `chat.send` Delivery mode does not match the server's declared mode |
| `NOT_A_MEMBER` | `chat.send` sender peer ID is not in the server's member list |
| `CHANNEL_NOT_FOUND` | `chat.send` `scope_id` does not correspond to a known channel |
| `DUPLICATE_DELIVERY` | `chat.send` Delivery ID was already accepted for this scope |
| `UNAUTHORIZED_OPERATION` | `channel_create` or `channel_delete` from non-owner |

## 8. Conformance

Implementations claiming Chat family conformance MUST pass the following KATs:

| KAT file | Covers |
|----------|--------|
| `pkg/spectest/chat/chat_send_clear_kat.json` | `chat.send` in Clear mode; plaintext body |
| `pkg/spectest/chat/chat_send_encrypted_kat.json` | `chat.send` in Tree mode; ciphertext body |
| `pkg/spectest/chat/chat_join_kat.json` | `chat.join` full round-trip; invite verification |
| `pkg/spectest/chat/chat_join_expired_kat.json` | `chat.join` with expired invite → `INVITE_EXPIRED` |
| `pkg/spectest/chat/chat_delivery_kinds_kat.json` | All five Delivery kinds (send, edit, delete, create, topic) |
| `pkg/spectest/chat/chat_dedup_kat.json` | Duplicate delivery ID → idempotent `accepted: true` |

All KAT files use the format defined in `90-conformance-harness.md`. Vectors
MUST include the full wire-level `PeerStreamRequest` and `PeerStreamResponse`
bytes, the Delivery JSON, and the expected response fields.

For mode compliance: a conformance suite claiming E2EE mode support (tree, crowd,
or channel) MUST additionally demonstrate rejection of a plaintext body in a
non-Clear scope and acceptance of a valid ciphertext body.
