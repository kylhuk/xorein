# 03 — Protocol Registry and Negotiation

## 1. Protocol ID namespace

All Xorein protocol IDs use multistream-select and follow the pattern:

```
/aether/<family>/<major>.<minor>.<patch>
```

The registry is authoritative in `pkg/protocol/registry.go`.

## 2. Registered families

### 2.1 v0.1.x core families

| Family | Canonical ID | Purpose |
|--------|-------------|---------|
| Chat | `/aether/chat/0.1.0` | Channel messages, server/channel management |
| Voice | `/aether/voice/0.1.0` | WebRTC signaling and MediaShield frame relay |
| Manifest | `/aether/manifest/0.1.0` | Server manifest fetch and publish |
| Identity | `/aether/identity/0.1.0` | Identity exchange and prekey distribution |
| Sync | `/aether/sync/0.1.0` | History sync and coverage exchange |
| Peer | `/aether/peer/0.1.0` | Peer info, exchange, relay, bootstrap ops |

### 2.2 v0.2.x extended families

| Family | Canonical ID | Purpose |
|--------|-------------|---------|
| DM | `/aether/dm/0.2.0` | Direct messages (Seal mode) |
| GroupDM | `/aether/groupdm/0.2.0` | Group DMs (Tree mode) |
| Friends | `/aether/friends/0.2.0` | Friend requests and roster |
| Presence | `/aether/presence/0.2.0` | Online/typing/presence indicators |
| Notify | `/aether/notify/0.2.0` | Push notification tokens and delivery |
| Moderation | `/aether/moderation/0.2.0` | Server moderation actions |
| Governance | `/aether/governance/0.2.0` | Role and permission changes |

A v0.1.x node MAY negotiate v0.2.x families if both sides support them. A
v0.2.x family is additive over v0.1.x — no v0.1.x behavior changes.

## 3. Capability flags

Capability flags are short strings exchanged in `PeerStreamRequest.advertised_caps`
and `PeerStreamRequest.required_caps`. They determine which operations a node
can handle. The authoritative list is in `pkg/protocol/capabilities.go`.

### 3.1 Peer transport flags

| Flag | Meaning |
|------|---------|
| `cap.peer.transport` | Required for any stream to proceed |
| `cap.peer.metadata` | Can serve peer metadata (addresses, role, keys) |
| `cap.peer.bootstrap` | Can register and serve bootstrap peer lists |
| `cap.peer.manifest` | Can publish and fetch server manifests |
| `cap.peer.join` | Can process server join requests |
| `cap.peer.delivery` | Can receive and forward message deliveries |
| `cap.peer.relay` | Can store relay queue entries and drain them |

### 3.2 Content capability flags

| Flag | Meaning |
|------|---------|
| `cap.chat` | Can send/receive channel messages |
| `cap.voice` | Can participate in voice sessions |
| `cap.dm` | Can send/receive direct messages |
| `cap.group-dm` | Can participate in group DMs |
| `cap.friends` | Friends roster and request handling |
| `cap.presence` | Presence and typing indicators |
| `cap.notify` | Notification delivery |
| `cap.identity` | Identity key distribution and prekey bundle |
| `cap.manifest` | Manifest fetch and publish |
| `cap.sync` | History sync |
| `cap.management` | Server creation and channel management |
| `cap.moderation` | Server moderation operations |
| `cap.rbac` | Role and permission changes |
| `cap.slow-mode` | Slow-mode enforcement |
| `cap.mentions` | @mention parsing and delivery |
| `cap.archivist` | Long-lived history storage and manifest coverage |

### 3.3 Security mode capability flags

| Flag | Meaning |
|------|---------|
| `mode.seal` | Supports Seal (X3DH + Double Ratchet + ML-KEM-768) |
| `mode.tree` | Supports Tree (hybrid MLS, ciphersuite 0xFF01) |
| `mode.crowd` | Supports Crowd (sender keys + epoch rotation) |
| `mode.channel` | Supports Channel (broadcast epoch) |
| `mode.mediashield` | Supports MediaShield (SFrame) |
| `mode.clear` | Willing to operate in Clear mode |

### 3.4 Role-capability matrix

| Role | Required capabilities |
|------|-----------------------|
| client | `cap.peer.transport`, `cap.peer.metadata`, `cap.peer.delivery`, all content caps, all mode caps |
| relay | `cap.peer.transport`, `cap.peer.relay` |
| bootstrap | `cap.peer.transport`, `cap.peer.bootstrap`, `cap.identity`, `cap.manifest` |
| archivist | `cap.peer.transport`, `cap.manifest`, `cap.sync`, `cap.chat`, `cap.archivist` |

## 4. Negotiation procedure

Negotiation runs at the start of every libp2p stream via the
`PeerStreamRequest` / `PeerStreamResponse` exchange defined in
`02-canonical-envelope.md §1`.

```
A opens stream to B via multistream-select on /aether/<family>/<version>

A → B: PeerStreamRequest{
    operation:       "peer.info",
    advertised_caps: ["cap.peer.transport", "cap.chat", ...],
    required_caps:   ["cap.peer.transport"],
    protocol_id:     "/aether/peer/0.1.0",
    request_id:      "<uuid>",
    payload:         <protobuf-encoded op-specific bytes>,
}

B computes:
    Accepted        = intersection(A.advertised_caps, B.localSupported)
    IgnoredRemote   = A.advertised_caps \ B.localSupported
    MissingRequired = A.required_caps  \ B.localSupported

if len(MissingRequired) > 0:
    B → A: PeerStreamResponse{
        error: PeerStreamError{
            code: "MISSING_REQUIRED_CAPABILITY",
            missing_capabilities: MissingRequired,
        },
        request_id: A.request_id,
    }
    B closes stream.
else:
    B → A: PeerStreamResponse{
        negotiated_protocol: "/aether/peer/0.1.0",
        accepted_caps:       Accepted,
        ignored_caps:        IgnoredRemote,
        payload:             <protobuf-encoded op response>,
        request_id:          A.request_id,
    }
```

If `MissingRequired` is non-empty, B MUST return a `PeerStreamError` with
code `MISSING_REQUIRED_CAPABILITY` and close the stream immediately. Streams
MUST NOT be left open after a negotiation error.

## 5. Version selection rules

These rules are implemented in `pkg/protocol/registry.go:NegotiateProtocol()`:

1. Family mismatch → reject.
2. Remote major > local candidate major → reject (`UNSUPPORTED_VERSION`).
3. Major downgrade (remote major < local) → reject unless explicitly allowed.
4. Minor downgrade → allowed; select the lower minor version.
5. Patch downgrade → allowed; select the lower patch version.
6. Deprecated candidates → skip.

The selected version is the highest mutually-acceptable minor.patch within the
same major. Both sides MUST record the negotiated protocol ID for the stream.

## 6. Operation constants

Operations are string constants of the form `<family>.<verb>`:

### 6.1 Peer family (`/aether/peer/0.1.0`)

| Operation | Required caps | Description |
|-----------|--------------|-------------|
| `peer.info` | `cap.peer.transport` | Fetch peer info and capabilities |
| `peer.exchange` | `cap.peer.transport` | Gossip known peers |
| `peer.deliver` | `cap.peer.delivery` | Deliver a message payload |
| `relay.store` | `cap.peer.relay` | Store payload in relay queue |
| `relay.drain` | `cap.peer.relay` | Drain relay queue for calling peer |
| `relay.register` | `cap.peer.relay` | Register this peer with a relay |
| `bootstrap.register` | `cap.peer.bootstrap` | Register with a bootstrap node |
| `bootstrap.fetch` | `cap.peer.bootstrap` | Fetch peers from bootstrap node |

### 6.2 Manifest family (`/aether/manifest/0.1.0`)

| Operation | Required caps | Description |
|-----------|--------------|-------------|
| `manifest.fetch` | `cap.manifest` | Fetch a server manifest by ID |
| `manifest.publish` | `cap.manifest` | Publish an updated manifest |

### 6.3 Chat family (`/aether/chat/0.1.0`)

| Operation | Required caps | Description |
|-----------|--------------|-------------|
| `chat.send` | `cap.chat`, `cap.peer.delivery` | Send a channel message |
| `chat.join` | `cap.peer.join`, `cap.manifest` | Join a server via deeplink |

### 6.4 Identity family (`/aether/identity/0.1.0`)

| Operation | Required caps | Description |
|-----------|--------------|-------------|
| `identity.fetch` | `cap.identity` | Fetch identity + prekey bundle |
| `identity.publish` | `cap.identity` | Publish updated prekeys |

### 6.5 DM family (`/aether/dm/0.2.0`)

| Operation | Required caps | Description |
|-----------|--------------|-------------|
| `dm.send` | `cap.dm`, `mode.seal` | Send a sealed DM delivery |

### 6.6 Friends family (`/aether/friends/0.2.0`)

| Operation | Required caps | Description |
|-----------|--------------|-------------|
| `friends.request` | `cap.friends` | Send a friend request |
| `friends.accept` | `cap.friends` | Accept a pending friend request |
| `friends.remove` | `cap.friends` | Remove a friend |

### 6.7 Voice family (`/aether/voice/0.1.0`)

| Operation | Required caps | Description |
|-----------|--------------|-------------|
| `voice.offer` | `cap.voice` | WebRTC offer (SDP) |
| `voice.answer` | `cap.voice` | WebRTC answer (SDP) |
| `voice.ice` | `cap.voice` | ICE candidate exchange |
| `voice.frame` | `cap.voice`, `mode.mediashield` | MediaShield-encrypted frame |

### 6.8 Presence, Notify, Moderation, Governance, Sync, GroupDM

Operations follow the same `<family>.<verb>` pattern. Full tables are in the
corresponding family documents (`48-family-presence.md` through
`52-family-voice.md`).

New operations MUST be registered with their required capabilities before use.

## 7. Wire framing

All peer-to-peer libp2p streams use the native protobuf framing from
`02-canonical-envelope.md §1`:

```
[4-byte big-endian uint32 payload length]
[proto.Marshal(PeerStreamRequest)]   // initiator → responder

[4-byte big-endian uint32 payload length]
[proto.Marshal(PeerStreamResponse)]  // responder → initiator
```

The multistream-select protocol handshake runs first (per libp2p convention),
then this framing carries the Xorein operation exchange. Each stream carries
exactly one request-response pair, then closes.

## 8. multistream-select handshake byte sequence

The complete byte-level handshake for opening an `/aether/peer/0.1.0` stream:

```
// multistream-select overhead (per libp2p spec)
A → B: "\x13/multistream/1.0.0\n"
B → A: "\x13/multistream/1.0.0\n"
A → B: "\x15/aether/peer/0.1.0\n"    // 0x15 = 21 = len("/aether/peer/0.1.0\n")
B → A: "\x15/aether/peer/0.1.0\n"    // echo = accepted; "na\n" = rejected

// After both sides echo: stream is now the Xorein application stream
// Xorein framing begins here:
A → B: [4B length][proto.Marshal(PeerStreamRequest{operation="peer.info", ...})]
B → A: [4B length][proto.Marshal(PeerStreamResponse{...})]
```

If B responds with `"na\n"` instead of echoing the protocol ID, the family is
not supported and A MUST try the next version or close the stream.

## 9. NegotiationError taxonomy

The following error codes are defined for `PeerStreamError.code`:

| Code | Cause |
|------|-------|
| `MISSING_REQUIRED_CAPABILITY` | Responder lacks a required cap from the initiator |
| `UNSUPPORTED_VERSION` | No common version for the requested family |
| `UNSUPPORTED_OPERATION` | Operation string not recognized |
| `SIGNATURE_MISMATCH` | Payload signature failed Ed25519 or ML-DSA-65 verification |
| `MODE_INCOMPATIBLE` | Requested security mode rejected by role policy |
| `RELAY_OPACITY_VIOLATION` | Relay received non-encrypted payload for E2EE scope |
| `OPERATION_FAILED` | Logical operation error (details in `message`) |
| `RATE_LIMITED` | Sender exceeded rate limit |
| `REPLAY_DETECTED` | Delivery ID seen before for this scope |
| `EXPIRED_SIGNATURE` | `signed_at` outside acceptance window |
