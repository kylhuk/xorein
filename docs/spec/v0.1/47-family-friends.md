# 47 — Family: Friends (`/aether/friends/0.2.0`)

The Friends family manages the friend roster: sending requests, accepting or
declining them, cancelling sent requests, removing friends, and blocking peers.
After mutual friendship is established, an identity share is exchanged so both
parties can locate each other directly without relying solely on the DHT.

## 1. Overview

Friend relationships are an opt-in layer on top of peer identity. A friend
relationship provides:

- **Persistent peer discovery**: friends share their current multiaddresses,
  reducing reliance on bootstrap/DHT lookups.
- **Trusted presence**: presence announcements from friends are given a higher
  trust level than from anonymous peers (see `48-family-presence.md §3`).
- **DM pre-authorization**: clients MAY use the friend relationship to
  pre-authorize DM sessions without an extra capability-negotiation round-trip.

This family does not carry message content. It carries only lifecycle signals
(requests, acceptances, removals, blocks) and the identity share that follows
mutual acceptance.

**Roles that participate:**

| Role | Participation |
|------|--------------|
| client | Initiator and recipient of all friend operations |
| relay | Stores ciphertext friend-request payloads for offline recipients |
| bootstrap | Does not participate |
| archivist | Does not participate |

**Security mode:** Friend operations are transport-level authenticated via
hybrid identity signatures. No Seal/Tree/Crowd/Channel encryption is applied to
friend request payloads themselves — they are signed, not encrypted. This is
intentional: friend requests are visible to the recipient's relay in metadata
(type and sender) but the signed payload binds integrity.

## 2. Capability requirements

| Capability | Role |
|-----------|------|
| `cap.friends` | Required on both initiator and responder for any operation |
| `cap.identity` | Required; sender's public key must be resolvable via `identity.fetch` |
| `cap.peer.delivery` | Required on responders accepting direct request delivery |
| `cap.peer.relay` | Required on relay nodes storing offline friend requests |

## 3. Operations

| Operation | Required caps (initiator) | Direction | Description |
|-----------|--------------------------|-----------|-------------|
| `friends.request` | `cap.friends` | initiator → target | Send a signed FriendRequest to the target peer |
| `friends.accept` | `cap.friends` | recipient → initiator | Accept (or decline/cancel/block) a pending request |
| `friends.remove` | `cap.friends` | either peer → the other | Remove an accepted friend from the roster |

The `friends.accept` operation carries all non-send friend lifecycle actions
via the `action` field. Only `friends.request`, `friends.accept`, and
`friends.remove` are distinct wire operations.

## 4. Wire format details

### 4.1 `friends.request`

The initiator constructs a `FriendRequest` and signs it with their hybrid
identity key.

**Canonical form for signing** (JSON, keys lexicographically sorted, no extra
whitespace; the `signature` field is excluded before signing):

```json
{
  "from_display_name": "<initiator display name; max 64 UTF-8 chars>",
  "from_peer_id": "<base58 libp2p peer ID>",
  "from_public_key": "<base64url(Ed25519_pub || ML-DSA-65_pub)>",
  "message": "<optional message; max 200 UTF-8 chars; empty string if absent>",
  "nonce": "<base64url(16-byte random nonce)>"
}
```

The canonical form MUST be UTF-8 JSON with keys in lexicographic order and no
trailing whitespace. The signature is:

```
canonical_json = UTF-8(JSON(canonical_form))
signature_bytes = Ed25519.Sign(ed_priv, canonical_json)
               || ML-DSA-65.Sign(mldsa_priv, canonical_json)
signature = base64url_no_padding(signature_bytes)  // 3373 bytes total
```

`PeerStreamRequest.payload = proto.Marshal(FriendRequest)` with:
- `from_id` = initiator peer ID
- `to_id` = target peer ID
- `state = FRIEND_REQUEST_STATE_PENDING`
- `action = FRIEND_REQUEST_ACTION_UNSPECIFIED`
- `nonce` = same nonce used in canonical form
- `signature_algorithm = SIGNATURE_ALGORITHM_HYBRID_ED25519_ML_DSA_65`
- `signature` = the 3373-byte hybrid signature

**Full JSON `data` field** (not part of the canonical signing form):

```json
{
  "from_peer_id": "<base58>",
  "from_public_key": "<base64url hybrid public key>",
  "from_display_name": "<string>",
  "message": "<optional string>",
  "nonce": "<base64url nonce>",
  "signature": "<base64url hybrid sig>",
  "sent_at": "<RFC3339Nano>",
  "expires_at": "<RFC3339Nano; max 7 days>"
}
```

On success, the responder returns an empty-payload `PeerStreamResponse`.

**Relay fallback:** If the target is offline, sender uses `relay.store` on
`/aether/peer/0.1.0` with the signed `FriendRequest` proto as the payload.
The relay stores the signed bytes without inspection.

### 4.2 `friends.accept`

The `friends.accept` operation covers four logical actions, differentiated by
the `action` field in the `FriendAction` JSON payload:

| `action` | `FriendRequestAction` | Meaning |
|----------|----------------------|---------|
| `accept` | `FRIEND_REQUEST_ACTION_ACCEPT` | Accept the request; transition to `accepted` |
| `decline` | `FRIEND_REQUEST_ACTION_DECLINE` | Decline the request; transition to `declined` |
| `cancel` | `FRIEND_REQUEST_ACTION_CANCEL` | Cancel a request that this peer sent; transition to `canceled` |
| `block` | `FRIEND_REQUEST_ACTION_BLOCK` | Block the peer; transition to `blocked`; suppress future requests |

`PeerStreamRequest.payload = proto.Marshal(FriendRequest)` with:
- `request_id` = the original `FriendRequest.request_id` this action responds to
- `state` = target state (see table)
- `action` = the action enum value
- `nonce` = new 16-byte random nonce for this action
- `signature` = hybrid signature over the canonical action form

**Canonical form for action signing:**

```json
{
  "action": "<accept|decline|cancel|block>",
  "nonce": "<base64url(16-byte nonce)>",
  "request_id": "<uuid v4; original request ID>",
  "signer_peer_id": "<base58>"
}
```

On `action = accept`: after the responder sends success, both peers MUST
exchange a `FriendIdentityShare` (see §4.4). The identity share is sent as an
additional `friends.accept` payload with the `data` field populated.

### 4.3 `friends.remove`

Either party may remove a friend. The operation is not required to be
acknowledged by the remote peer (removal is unilateral).

`PeerStreamRequest.payload = proto.Marshal(FriendRequest)` with:
- `from_id` = the peer initiating the removal
- `to_id` = the peer being removed
- `state = FRIEND_REQUEST_STATE_CANCELED` (reusing the proto field; removal uses
  this state as the closest semantic match)
- `action = FRIEND_REQUEST_ACTION_CANCEL`
- `nonce` = 16-byte random nonce
- `signature` = hybrid signature over the canonical removal form

**Canonical form for removal signing:**

```json
{
  "action": "remove",
  "nonce": "<base64url nonce>",
  "from_peer_id": "<base58>",
  "to_peer_id": "<base58>"
}
```

After removal, both peers MUST delete the `FriendRecord` and any cached address
information obtained via the identity share.

### 4.4 `FriendIdentityShare` — post-acceptance key exchange

After both peers have completed `friends.accept` (action = `accept`), each
MUST send a `FriendIdentityShare` to the other. This can be piggybacked on the
same `friends.accept` stream via the `data` field, or sent as a subsequent
`friends.accept` with empty action and the identity share payload.

**`FriendIdentityShare` JSON:**

```json
{
  "peer_id": "<base58 libp2p peer ID>",
  "display_name": "<string; max 64 UTF-8 chars>",
  "addresses": [
    "<multiaddr string>",
    "..."
  ],
  "public_key": "<base64url(Ed25519 public key, 32 bytes)>",
  "ml_dsa_65_public_key": "<base64url(ML-DSA-65 public key, 1952 bytes)>",
  "signed_at": "<int64; unix milliseconds>"
}
```

The `addresses` list MUST be sorted lexicographically before inclusion. The
`signed_at` field MUST be within 5 minutes of the current time. The entire JSON
object (with `signed_at` but without a separate signature field) is signed as
part of the enclosing `FriendRequest.signature`.

The proto type is `FriendIdentityShare` from `proto/aether.proto`
(`PAYLOAD_TYPE_FRIEND_IDENTITY_SHARE`).

### 4.5 Concurrent request tie-breaking

If two peers simultaneously send `friends.request` to each other:

1. Both requests arrive at the respective recipients.
2. Each recipient accepts their received request.
3. The tie is broken deterministically: the peer with the
   lexicographically-lower `peer_id` (base58) is treated as the canonical
   requester; the other is treated as the canonical recipient.
4. Both requests transition to `FRIEND_REQUEST_STATE_ACCEPTED`.
5. A `FriendConcurrentResolution` event is logged (not transmitted).

No special wire exchange is needed for tie-breaking. The deterministic rule
ensures both peers arrive at the same `canonical_requester` and
`canonical_recipient` without coordination.

### 4.6 Blocking

When `action = block`:

- The blocked peer is added to the local block list.
- Future `friends.request` operations from the blocked peer MUST be silently
  discarded (no error response; the blocker appears offline to the blocked peer).
- If the blocked peer was previously a friend, the `FriendRecord` is updated to
  `status = "blocked"` and the identity share addresses are removed from cache.
- The blocker MUST NOT send a `friends.remove` to the blocked peer (that would
  reveal the block). Removal is local-only.

## 5. Security mode binding

Friend operations are signed-but-not-encrypted at the application layer.
Transport-level encryption is provided by Noise XX (see `30-transport-and-
noise.md`), which protects all libp2p stream content including friend request
payloads from network observers.

The hybrid signature on every `FriendRequest` and `FriendAction` provides:
- **Integrity**: the payload has not been tampered with.
- **Authentication**: the sender possesses the private key corresponding to the
  advertised hybrid public key.
- **Non-repudiation**: the signature is a 3373-byte combined Ed25519 + ML-DSA-65
  value, quantum-safe against store-now-verify-later attacks.

Implementations MUST verify both the Ed25519 and ML-DSA-65 components
independently. Failure of either MUST result in `SIGNATURE_MISMATCH`.

## 6. State persistence

### 6.1 `friends` bucket

Each friend relationship is a `FriendRecord`:

```json
{
  "peer_id": "<base58 remote peer ID>",
  "public_key": "<base64url(Ed25519_pub, 32 bytes)>",
  "ml_dsa_65_public_key": "<base64url(ML-DSA-65_pub, 1952 bytes)>",
  "display_name": "<string>",
  "addresses": ["<multiaddr>", "..."],
  "status": "<pending_sent|pending_received|accepted|declined|blocked>",
  "request_id": "<uuid v4>",
  "nonce": "<base64url nonce of last action>",
  "created_at": "<RFC3339Nano>",
  "updated_at": "<RFC3339Nano>"
}
```

Status values and transitions:

| From | Action | To |
|------|--------|----|
| _(none)_ | `friends.request` sent | `pending_sent` |
| _(none)_ | `friends.request` received | `pending_received` |
| `pending_sent` | remote accepted | `accepted` |
| `pending_sent` | remote declined | `declined` |
| `pending_sent` | local cancel | _(deleted)_ |
| `pending_received` | local accept | `accepted` |
| `pending_received` | local decline | _(deleted)_ |
| `pending_received` | local block | `blocked` |
| `accepted` | `friends.remove` | _(deleted)_ |
| `accepted` | local block | `blocked` |
| `blocked` | manual unblock (future operation) | `pending_sent` |

Declined requests are deleted (not retained) by default. Blocked records are
retained so the block list persists across restarts.

### 6.2 Request expiry

Pending friend requests MUST expire after 7 days. On startup, the node MUST
scan the `friends` bucket and delete any `pending_sent` or `pending_received`
records whose `created_at` is more than 7 days old.

## 7. Error codes

| Code | Trigger |
|------|---------|
| `MISSING_REQUIRED_CAPABILITY` | Peer does not advertise `cap.friends` |
| `SIGNATURE_MISMATCH` | Hybrid signature on FriendRequest or FriendAction failed |
| `REPLAY_DETECTED` | Nonce already seen for this `(from_peer_id, to_peer_id)` pair |
| `EXPIRED_SIGNATURE` | `sent_at` more than 5 minutes in the past |
| `OPERATION_FAILED` | No pending request found for the given `request_id` |
| `RATE_LIMITED` | Sender has exceeded the friend-request rate limit |

Rate limit: max 10 `friends.request` operations per peer per hour. The
responder MUST include the retry-after duration in `PeerStreamError.message`.

## 8. Conformance

Conformance class: **W5** (Friends family conformance).

KATs in `pkg/spectest/friends/`:

- `friends_request.json` — send a FriendRequest; verify canonical JSON, hybrid
  signature, and nonce.
- `friends_accept.json` — accept action; verify state transition and identity
  share exchange.
- `friends_decline.json` — decline action; verify state transition and no
  identity share.
- `friends_cancel.json` — initiator cancels pending request; verify deletion.
- `friends_remove.json` — accepted friend removed; verify FriendRecord deleted.
- `friends_block.json` — block action; verify record transitions to `blocked`
  and future requests are silently discarded.
- `friends_concurrent.json` — simultaneous requests from both peers; verify
  deterministic tie-breaking and both records transition to `accepted`.
- `friends_relay_fallback.json` — target offline; verify signed FriendRequest
  stored in relay queue.
- `friends_expiry.json` — pending request older than 7 days deleted on startup.
- `friends_sig_mismatch.json` — tampered canonical JSON; verify
  `SIGNATURE_MISMATCH` returned.

Implementations MUST pass all KATs in `pkg/spectest/friends/` to claim Friends
family conformance.
