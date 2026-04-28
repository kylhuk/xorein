# 45 — Family: DM (`/aether/dm/0.2.0`)

Direct messages (DMs) provide 1:1 end-to-end encrypted communication between
two peers. Every message in this family MUST be protected by Seal mode (hybrid
X3DH + Double Ratchet), as specified in `11-mode-seal.md`. Clear mode is not
permitted for this family.

## 1. Overview

The DM family handles the setup and delivery of private, Seal-encrypted messages
between two clients. It operates over `/aether/dm/0.2.0` using the native
protobuf peer-stream envelope defined in `02-canonical-envelope.md §1`.

**Roles that participate:**

| Role | Participation |
|------|--------------|
| client | Initiator and recipient of DM sessions and messages |
| relay | Stores ciphertext-only `RelayQueueEntry` payloads; MUST NOT inspect body |
| bootstrap | Does not participate |
| archivist | Does not participate |

**Security mode:** Seal only (`mode.seal`). Negotiation of any other mode for
this family MUST fail with `MODE_INCOMPATIBLE`.

**Protocol dependency:** Before the first DM can be sent, the sender MUST have
fetched the recipient's prekey bundle via `identity.fetch` on
`/aether/identity/0.1.0`. The DM family itself does not carry key material
except in the first message of a new session.

## 2. Capability requirements

| Capability | Role |
|-----------|------|
| `cap.dm` | Required on both initiator and responder for any DM operation |
| `mode.seal` | Required on both sides; `dm.send` will not proceed without it |
| `cap.peer.delivery` | Required on the responder to accept a direct delivery |
| `cap.peer.relay` | Required on a relay node to store the fallback payload |

A peer that advertises `cap.dm` MUST also advertise `mode.seal`. A node that
advertises `mode.seal` without `cap.dm` MAY still participate in Seal-mode
sessions through other families.

## 3. Operations

| Operation | Required caps (initiator) | Direction | Description |
|-----------|--------------------------|-----------|-------------|
| `dm.send` | `cap.dm`, `mode.seal` | initiator → responder | Send a Seal-encrypted DM delivery to the recipient |

`dm.send` is a request-response operation: the responder acknowledges receipt
with an empty-payload success response, or returns a `PeerStreamError`. The
initiator MUST NOT consider delivery confirmed until a success response is
received.

## 4. Wire format details

### 4.1 Delivery flow

1. **Prekey fetch.** If no Seal session exists with the recipient, sender calls
   `identity.fetch` on `/aether/identity/0.1.0` to obtain the recipient's
   prekey bundle (see `11-mode-seal.md §1`).
2. **Hybrid X3DH.** Sender performs the hybrid X3DH computation from
   `11-mode-seal.md §2` to derive `hybrid_master`, `root_key`, and the
   initial `chain_key`.
3. **Encryption.** Sender encrypts the message body using the Double Ratchet
   from `11-mode-seal.md §3`. The ciphertext format is `"seal/v1"`.
4. **`dm_create` delivery.** If this is the very first message in a new DM
   relationship, sender MUST first send a `dm_create` Delivery (see §4.3)
   before sending the first `dm_message` Delivery.
5. **Direct delivery.** Sender opens a stream to the recipient and sends a
   `dm.send` request carrying the Delivery.
6. **Relay fallback.** If the recipient is unreachable (peer not found, stream
   error, or `DM_TRANSPORT_REASON_PEER_OFFLINE`), sender falls back to
   `relay.store` on `/aether/peer/0.1.0` and stores the ciphertext payload in
   the recipient's relay queue.

### 4.2 `dm.send` request payload

The `PeerStreamRequest.payload` bytes are `proto.Marshal(DmMessageEnvelope)`.
The JSON representation carried in the `Delivery.data` field is:

```json
{
  "delivery": {
    "id": "<uuid v4>",
    "kind": "dm_message",
    "scope_id": "<dm_record_id>",
    "scope_type": "dm",
    "sender_peer_id": "<base58 libp2p peer ID>",
    "sender_public_key": "<base64url(Ed25519_pub || ML-DSA-65_pub)>",
    "recipient_peer_ids": ["<base58 recipient peer ID>"],
    "body": "<base64url_no_padding(ChaCha20-Poly1305 ciphertext)>",
    "created_at": "<RFC3339Nano>",
    "signature": "<base64url_no_padding(Ed25519_sig || ML-DSA-65_sig)>"
  },
  "session_id": "<uuid v4>",
  "ek_pub": "<base64url_no_padding(EK_x25519 public key, 32 bytes)>",
  "ct_mlkem": "<base64url_no_padding(ML-KEM-768 ciphertext)>",
  "opk_index": 0,
  "ciphertext_format": "seal/v1",
  "data": "<base64url_no_padding(53-byte message header)>"
}
```

Fields `ek_pub`, `ct_mlkem`, and `opk_index` are present only in the **first
message of a new session** (the session-initiation message). They MUST be
absent in all subsequent messages within the same session.

If `opk_index` is absent and no one-time prekey (OPK) was available, the
initiator used the 3-DH variant (DH1, DH2, DH3 only, per `11-mode-seal.md
§2.1`). The responder MUST detect this by the absence of `opk_index` in the
first message and use the 3-DH decapsulation path.

**Message header** (53 bytes, base64url in `data`):

```
[1-byte version = 0x01]
[4-byte ratchet step counter, big-endian]
[4-byte previous chain length, big-endian]
[32-byte sender ratchet DH public key]
[12-byte ChaCha20 nonce]
```

### 4.3 `dm_create` Delivery

Before the first `dm_message` can be sent, the initiator MUST deliver a
`dm_create` Delivery to establish the DM relationship. This is sent as a
`dm.send` request with `kind="dm_create"`:

```json
{
  "delivery": {
    "id": "<uuid v4>",
    "kind": "dm_create",
    "scope_id": "<dm_record_id>",
    "scope_type": "dm",
    "sender_peer_id": "<base58>",
    "sender_public_key": "<base64url hybrid public keys>",
    "recipient_peer_ids": ["<base58>"],
    "body": "",
    "created_at": "<RFC3339Nano>",
    "signature": "<base64url hybrid sig>"
  },
  "session_id": "<uuid v4>",
  "data": "{\"display_name\":\"<initiator display name>\"}"
}
```

The `body` field is empty (no ciphertext). The `data` field carries a
JSON-encoded object with the initiator's `display_name`. This lets the
recipient create a `DMRecord` and render the conversation before the first
message arrives.

### 4.4 `dm.send` success response

```json
{
  "delivery_id": "<mirrors delivery.id>",
  "acknowledged": true,
  "acknowledged_at": "<RFC3339Nano>"
}
```

Carried as `proto.Marshal(DmDeliveryReceipt)` in `PeerStreamResponse.payload`.

### 4.5 Relay fallback payload

When the recipient is offline, the sender uses `relay.store` on
`/aether/peer/0.1.0`. The relay queue entry carries:

```json
{
  "recipient_peer_id": "<base58>",
  "payload": "<base64url(proto.Marshal(DmMessageEnvelope))>",
  "expires_at": "<RFC3339Nano; max 7 days from now>",
  "delivery_attempt": 1
}
```

The relay node MUST store only the ciphertext bytes. It MUST NOT decrypt,
inspect, or log the `DmMessageEnvelope.ciphertext` field. Relay opacity is
enforced by the `RELAY_OPACITY_VIOLATION` error code (see `02-canonical-
envelope.md §1.3`).

### 4.6 Transport decision

Before attempting delivery, the sender computes a `DmTransportDecision`:

| `DmDeliveryPath` | `DmTransportReason` | Meaning |
|-----------------|--------------------|----|
| `DIRECT` | `DIRECT_AVAILABLE` | Recipient is reachable; use `dm.send` directly |
| `OFFLINE` | `PEER_OFFLINE` | Recipient not reachable; use `relay.store` fallback |
| `REJECT` | `UNSUPPORTED_SECURITY_MODE` | Recipient does not support `mode.seal` |

The sender MUST record the transport decision in the `DmDeliveryReceipt` for
diagnostic purposes.

## 5. Security mode binding

This family is permanently bound to `SecurityMode = SECURITY_MODE_SEAL`.

All `PeerStreamRequest` messages in this family MUST carry
`security_mode = SECURITY_MODE_SEAL`. A responder that receives a request with
any other security mode MUST return `MODE_INCOMPATIBLE` and close the stream.

Relay nodes MUST enforce relay opacity: if a `relay.store` entry for the DM
family is received with a non-empty plaintext body (detected by absence of AEAD
tag or malformed ciphertext structure), the relay MUST return
`RELAY_OPACITY_VIOLATION` and discard the entry.

Key material (X3DH secrets, Double Ratchet root/chain keys, message keys) MUST
NOT appear in logs, diagnostic snapshots, or relay payloads at any time.

## 6. State persistence

State is serialized to the SQLCipher store (see `pkg/storage/store.go`).

### 6.1 `dms` bucket

Each DM relationship is a `DMRecord`:

```json
{
  "id": "<uuid v4; dm_record_id>",
  "peer_id": "<base58 remote peer ID>",
  "display_name": "<remote peer display name>",
  "security_mode": "seal",
  "created_at": "<RFC3339Nano>",
  "last_message_at": "<RFC3339Nano or null>",
  "session_id": "<uuid v4; current Seal session>",
  "unread_count": 0
}
```

### 6.2 `identities` bucket — session sub-key

Double Ratchet session state is stored under a sub-key of the form
`"dm_session/<session_id>"` within the `identities` bucket:

```json
{
  "session_id": "<uuid v4>",
  "dm_record_id": "<uuid v4>",
  "root_key": "<base64url(32 bytes)>",
  "send_chain_key": "<base64url(32 bytes)>",
  "recv_chain_key": "<base64url(32 bytes)>",
  "send_ratchet_pub": "<base64url(32 bytes)>",
  "send_ratchet_priv": "<base64url(32 bytes)>",
  "recv_ratchet_pub": "<base64url(32 bytes)>",
  "send_counter": 0,
  "recv_counter": 0,
  "prev_send_chain_length": 0,
  "skip_list": {}
}
```

Session state MUST be stored in the SQLCipher store and MUST NOT exist only
in memory. After delivery of every message, the updated session state MUST be
persisted before the delivery acknowledgement is sent to the caller.

Session state MUST be securely deleted when the DM relationship is removed by
either party. Deletion MUST overwrite the key material rather than just
removing the storage record.

### 6.3 `messages` bucket

Each delivered message is a `MessageRecord` with `scope_type = "dm"`:

```json
{
  "id": "<delivery.id>",
  "scope_type": "dm",
  "scope_id": "<dm_record_id>",
  "sender_peer_id": "<base58>",
  "body": "<decrypted plaintext, stored encrypted by SQLCipher>",
  "created_at": "<RFC3339Nano>",
  "received_at": "<RFC3339Nano>"
}
```

## 7. Error codes

All error codes are carried in `PeerStreamResponse.error` as `PeerStreamError`.

| Code | Trigger |
|------|---------|
| `MISSING_REQUIRED_CAPABILITY` | Responder does not advertise `cap.dm` or `mode.seal` |
| `MODE_INCOMPATIBLE` | `security_mode` in request is not `SECURITY_MODE_SEAL` |
| `RELAY_OPACITY_VIOLATION` | Relay received non-encrypted DM payload |
| `SIGNATURE_MISMATCH` | Delivery signature verification failed |
| `REPLAY_DETECTED` | `delivery.id` already seen for this `dm_record_id` |
| `EXPIRED_SIGNATURE` | `delivery.created_at` is more than 5 minutes in the past |
| `OPERATION_FAILED` | Session state corrupt, OPK exhausted, or other logical error |
| `RATE_LIMITED` | Sender has exceeded the per-peer DM rate limit |

Rate limit: max 60 `dm.send` requests per peer per minute. The responder MUST
include the retry-after duration in `PeerStreamError.message`.

## 8. Conformance

Conformance class: **W3** (family protocol conformance).

KATs in `pkg/spectest/dm/`:

- `dm_create.json` — `dm_create` Delivery round-trip: initiator creates, recipient
  produces DMRecord.
- `dm_send_first.json` — first message with X3DH header fields present; verifies
  responder derives same `hybrid_master`.
- `dm_send_subsequent.json` — second and third messages without X3DH fields;
  verifies Double Ratchet chain advancement.
- `dm_relay_fallback.json` — recipient offline; verify relay stores ciphertext
  without inspecting body.
- `dm_replay_reject.json` — duplicate `delivery.id`; verify `REPLAY_DETECTED`
  returned.
- `dm_mode_reject.json` — request with `SECURITY_MODE_CLEAR`; verify
  `MODE_INCOMPATIBLE` returned.
- `dm_session_persist.json` — session state persisted after each message; node
  restart recovers session correctly.

Implementations MUST pass all KATs in `pkg/spectest/dm/` to claim DM family
conformance.
