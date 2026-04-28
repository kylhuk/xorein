# 40 — Family: Peer (`/aether/peer/0.1.0`)

This document specifies the Peer family, which is the mandatory bootstrap
protocol for all Xorein node-to-node connections. Every stream MUST negotiate
the Peer family before any other family is used.

## 1. Overview

The Peer family handles:

- Peer identity exchange and gossip.
- Signed message delivery between online peers.
- Relay queue store-and-forward for offline message delivery.
- Bootstrap node peer list distribution.

**Roles that use this family:** All roles (client, relay, bootstrap, archivist).
Every node MUST implement the `peer.info` and `peer.exchange` operations.
Relay-specific operations (`relay.*`) MUST only be served by nodes with role
`relay`. Bootstrap-specific operations (`bootstrap.*`) MUST only be served by
nodes with role `bootstrap`.

**Security modes:** The Peer family itself is transport-layer metadata. It
operates below the conversation security mode layer. The relay opacity invariant
(§5) applies regardless of the negotiated conversation mode.

**Protocol ID:** `/aether/peer/0.1.0`

**Required capability:** `cap.peer.transport` — MUST be advertised and MUST be
present in `required_caps` for every Peer family stream. A node that does not
advertise `cap.peer.transport` MUST NOT be connected to.

## 2. Capability requirements

| Capability | Meaning |
|------------|---------|
| `cap.peer.transport` | Required for all Peer family operations; baseline peer connectivity |
| `cap.peer.metadata` | Enables full peer info including addresses and last-seen timestamps |
| `cap.peer.bootstrap` | Required for `bootstrap.register` and `bootstrap.fetch` |
| `cap.peer.manifest` | Enables manifest metadata in peer exchange responses |
| `cap.peer.join` | Enables join-request routing during server onboarding |
| `cap.peer.delivery` | Required for `peer.deliver`; enables signed Delivery forwarding |
| `cap.peer.relay` | Required for all `relay.*` operations |

All capability strings are defined in `pkg/protocol/capabilities.go` as
`FeaturePeer*` constants. Capability negotiation follows
`03-protocol-registry-and-negotiation.md §3`.

## 3. Operations

| Operation | Required caps | Direction | Request payload type | Response payload type | Description |
|-----------|--------------|-----------|---------------------|----------------------|-------------|
| `peer.info` | `cap.peer.transport` | initiator → responder | `PeerInfoRequest` | `PeerInfoResponse` | Fetch peer identity, role, capabilities, and addresses |
| `peer.exchange` | `cap.peer.transport` | initiator → responder | `PeerExchangeRequest` | `PeerExchangeResponse` | Gossip known peer records; filter by known IDs |
| `peer.deliver` | `cap.peer.transport`, `cap.peer.delivery` | initiator → responder | `PeerDeliverRequest` | `PeerDeliverResponse` | Forward a signed Delivery JSON to an online peer |
| `relay.store` | `cap.peer.transport`, `cap.peer.relay` | initiator → responder | `RelayStoreRequest` | `RelayStoreResponse` | Store an encrypted Delivery in relay queue for an offline recipient |
| `relay.drain` | `cap.peer.transport`, `cap.peer.relay` | initiator → responder | `RelayDrainRequest` | `RelayDrainResponse` | Drain relay queue entries addressed to the initiator |
| `relay.register` | `cap.peer.transport`, `cap.peer.relay` | initiator → responder | `RelayRegisterRequest` | `RelayRegisterResponse` | Register this peer's current addresses with a relay node |
| `bootstrap.register` | `cap.peer.transport`, `cap.peer.bootstrap` | initiator → responder | `BootstrapRegisterRequest` | `BootstrapRegisterResponse` | Announce this peer to a bootstrap node |
| `bootstrap.fetch` | `cap.peer.transport`, `cap.peer.bootstrap` | initiator → responder | `BootstrapFetchRequest` | `BootstrapFetchResponse` | Fetch known peers from a bootstrap node |

All operations use the peer-stream envelope defined in
`02-canonical-envelope.md §1`. The `operation` field in `PeerStreamRequest`
MUST be set to the exact operation string in the table above.

## 4. Wire format details

All request and response payload structs listed below are JSON-encoded and
stored in the `PeerStreamRequest.payload` / `PeerStreamResponse.payload` bytes
field. Implementations MUST NOT assume the payload is valid proto binary; it is
a JSON-encoded Go struct.

### 4.1 `peer.info`

**Request:** `PeerInfoRequest` — an empty JSON object `{}`. No fields are
required; implementations MUST accept an empty payload for this operation.

**Response:** `PeerInfoResponse` — a JSON-encoded `PeerInfo` struct:

```
{
  "peer_id":    string,   // libp2p peer ID (base58btc multihash)
  "role":       string,   // "client" | "relay" | "bootstrap" | "archivist"
  "addresses":  [string], // multiaddrs this peer is reachable at
  "public_key": string    // Ed25519 public key, base64url no-padding
}
```

Notes:
- `peer_id` MUST match the libp2p identity used in the Noise XX handshake.
- `addresses` MUST contain at least one multiaddr. When `cap.peer.metadata` is
  negotiated, addresses MUST include all known listen addresses including relay
  circuit addresses.
- The ML-DSA-65 public key is conveyed in the `PrekeyBundle` via the Identity
  family (§43); it is not included in `PeerInfo` to keep the Peer family minimal.
- `last_seen_at` and `source` fields from `pkg/node/types.go:PeerRecord` are
  not transmitted over the wire; they are local state only.

### 4.2 `peer.exchange`

**Request:** `PeerExchangeRequest` — JSON-encoded:

```
{
  "known_peer_ids": [string], // peer IDs the initiator already has; responder excludes these
  "server_ids":     [string], // optional; filter to peers associated with these servers
  "limit":          int       // max entries to return; 0 or absent means 50 (max 50)
}
```

**Response:** `PeerExchangeResponse` — JSON-encoded:

```
{
  "peers": [
    {
      "peer_id":    string,
      "role":       string,
      "addresses":  [string],
      "public_key": string
    }
  ]
}
```

Notes:
- The responder MUST return at most 50 peer records per response.
- Peers in `known_peer_ids` MUST be excluded from the response.
- Peers with no valid addresses MUST be excluded.
- The responder SHOULD prioritize peers with the most recent `last_seen_at`
  timestamps. Order is otherwise implementation-defined.
- Relay nodes MUST NOT include their internal state (relay queue contents,
  enqueued peer IDs) in peer exchange responses.

### 4.3 `peer.deliver`

**Request:** `PeerDeliverRequest` — JSON-encoded:

```
{
  "delivery": { /* Delivery JSON object; see 02-canonical-envelope.md §3.3 */ }
}
```

**Response:** `PeerDeliverResponse` — JSON-encoded:

```
{
  "accepted": bool,
  "error":    string   // human-readable; omitted on success
}
```

Notes:
- The receiver MUST verify the Delivery signature per
  `02-canonical-envelope.md §3.4` before accepting.
- Duplicate delivery IDs (same `id` for same `scope_id`) MUST be rejected with
  `OPERATION_FAILED` and a message of `"duplicate delivery"`. This is the
  idempotency gate for at-least-once delivery.
- The receiver MUST check that `recipient_peer_ids` contains its own peer ID or
  is empty (broadcast). Deliveries addressed to other peers MUST be rejected.

### 4.4 `relay.store`

**Request:** `RelayStoreRequest` — JSON-encoded:

```
{
  "recipient_peer_id": string,   // peer ID of the intended recipient
  "delivery": { /* Delivery JSON */ }
}
```

**Response:** `RelayStoreResponse` — JSON-encoded:

```
{
  "accepted":   bool,
  "queue_depth": int,  // current queue depth for this recipient after store
  "error":       string
}
```

**Relay opacity invariant:** A relay node MUST enforce the following rule before
accepting a `relay.store` request:

1. Inspect `delivery.scope_type`. If `scope_type` is not `"clear"`, the
   `delivery.body` MUST be a valid base64url-encoded ciphertext (non-empty,
   non-plaintext UTF-8 text). A relay MUST NOT attempt to decrypt; it MUST
   only check that the body is not a raw UTF-8 string that appears unencrypted.
   The practical check: base64url decode MUST succeed and the decoded length
   MUST be at least 1 byte.
2. If this check fails → reject with `RELAY_OPACITY_VIOLATION`.

Additional relay queue constraints:
- Max entries per recipient: **256** (from `relayQueueLimit` in `pkg/node/service.go`).
- When the per-recipient queue is full, the relay MUST reject new stores with
  `OPERATION_FAILED` and message `"relay queue full"`.
- TTL per entry: **24 hours** (from `relayQueueTTL` in `pkg/node/service.go`).
  Entries older than 24 hours MUST be silently purged before serving `relay.drain`.
- The relay MUST NOT read, log, or otherwise inspect `delivery.body` beyond the
  opacity check above.

### 4.5 `relay.drain`

**Request:** `RelayDrainRequest` — JSON-encoded using the `DrainRequest` struct:

```
{
  "requester": { /* PeerInfo */ },
  "signature": string   // base64url hybrid signature over canonical DrainRequest JSON
}
```

The canonical form for signing is the `DrainRequest` JSON with `signature`
field omitted.

**Response:** `RelayDrainResponse` — JSON-encoded:

```
{
  "deliveries": [ { /* Delivery JSON */ } ],
  "drained_count": int
}
```

Notes:
- The relay MUST verify the `signature` field using the requester's public key
  (from `requester.public_key`) before releasing any queue entries.
- Only entries addressed to `requester.peer_id` MUST be returned.
- After successful drain, the relay MUST remove the returned entries from
  the persistent relay queue.
- Expired entries (TTL elapsed) MUST be excluded from the drain response and
  MUST be purged.

### 4.6 `relay.register`

**Request:** `RelayRegisterRequest` — JSON-encoded using the `RegisterRelayRequest` struct:

```
{
  "multiaddr": string,   // the initiator's public multiaddr to register
  "role":      string    // informational; always stored as "relay" on the relay side
}
```

**Response:** `RelayRegisterResponse` — JSON-encoded:

```
{
  "accepted": bool,
  "error":    string
}
```

Notes:
- The relay MUST store the registered multiaddr in its known-peers list for
  peer exchange purposes.
- The relay MUST NOT use the registered role field for access control;
  only the initiator's negotiated capabilities govern what it may do.

### 4.7 `bootstrap.register`

**Request:** `BootstrapRegisterRequest` — JSON-encoded using `PeerInfo`:

```
{
  "peer_id":    string,
  "role":       string,
  "addresses":  [string],
  "public_key": string
}
```

**Response:** `BootstrapRegisterResponse` — JSON-encoded:

```
{
  "accepted": bool,
  "known_peers_count": int,
  "error": string
}
```

Notes:
- Bootstrap nodes MUST store registered peers in their known-peers list.
- Bootstrap nodes MUST NOT enforce capability restrictions beyond
  `cap.peer.bootstrap` for this operation.
- Registrations are ephemeral — bootstrap nodes SHOULD prune records not
  refreshed within 1 hour.

### 4.8 `bootstrap.fetch`

**Request:** `BootstrapFetchRequest` — JSON-encoded:

```
{
  "known_peer_ids": [string],
  "limit":          int
}
```

**Response:** `BootstrapFetchResponse` — JSON-encoded:

```
{
  "peers": [ { /* PeerInfo */ } ]
}
```

Notes:
- Semantics are identical to `peer.exchange` but served from the bootstrap
  node's persisted known-peers list rather than a live peer cache.
- Max 50 peers per response; `known_peer_ids` are excluded.

## 5. Security mode binding

The Peer family operates at the transport layer and is mode-agnostic. No
conversation security mode is applied to Peer family payloads themselves.

The relay opacity invariant (§4.4) bridges the Peer family with the
conversation security mode layer: the relay enforces that non-Clear conversation
payloads are ciphertext before forwarding.

The Noise XX hop-to-hop layer (see `30-transport-and-noise.md §1.2`) provides
confidentiality and authentication for all Peer family traffic in transit.
Peer family payloads are therefore never transmitted in cleartext on the wire,
even for Clear mode conversations.

## 6. State persistence

| State bucket | Key type | Value type | Description |
|-------------|----------|-----------|-------------|
| `known_peers` | `peer_id` (string) | `PeerRecord` (JSON) | Discovered peer metadata |
| `relay_queues` | `recipient_peer_id` (string) | `[]RelayQueueEntry` (JSON) | Per-recipient relay queue |

Go types from `pkg/node/types.go`:

- `PeerRecord`: `PeerID`, `Role`, `Addresses []string`, `PublicKey`, `Source`,
  `LastSeenAt time.Time`
- `RelayQueueEntry`: `Key`, `Payload []byte`, `EnqueuedAt time.Time`,
  `ExpiresAt time.Time`

The `known_peers` bucket is populated by all Peer family operations that
introduce new peer records (exchange, deliver, bootstrap.fetch,
relay.register). The in-memory `livePeerRegistry` (not persisted) caches
currently-connected peers by peer ID and address.

Relay queues are persisted in the `relay_queues` bucket under the recipient's
peer ID. Each entry holds the full Delivery JSON bytes as `Payload`. Entries
MUST be persisted atomically — a relay node that crashes after accepting a
`relay.store` but before persisting MUST treat the entry as lost and rely on
the sender's retry logic.

## 7. Error codes

The following `PeerStreamError.code` values are specific to the Peer family,
in addition to the generic codes in `02-canonical-envelope.md §1.3`:

| Code | Trigger |
|------|---------|
| `RELAY_OPACITY_VIOLATION` | `relay.store` received a non-encrypted body for a non-Clear scope |
| `RELAY_QUEUE_FULL` | Per-recipient relay queue has reached 256 entries |
| `RELAY_AUTH_FAILED` | `relay.drain` signature verification failed |
| `BOOTSTRAP_RATE_LIMITED` | Bootstrap node is rate-limiting registration/fetch requests |
| `DUPLICATE_DELIVERY` | `peer.deliver` received a delivery ID already seen for this scope |
| `UNKNOWN_RECIPIENT` | `relay.store` recipient peer ID is not registered with this relay |

## 8. Conformance

Implementations claiming Peer family conformance MUST pass the following
known-answer tests (KATs):

| KAT file | Covers |
|----------|--------|
| `pkg/spectest/peer/peer_info_kat.json` | `peer.info` request/response round-trip |
| `pkg/spectest/peer/peer_exchange_kat.json` | Exchange with `known_peer_ids` filter |
| `pkg/spectest/peer/relay_store_drain_kat.json` | Store + drain round-trip; opacity check |
| `pkg/spectest/peer/relay_opacity_kat.json` | `RELAY_OPACITY_VIOLATION` negative case |
| `pkg/spectest/peer/bootstrap_kat.json` | Register + fetch round-trip |

All KAT files use the format defined in `90-conformance-harness.md`. Each
vector MUST include the full `PeerStreamRequest` and `PeerStreamResponse`
serialized bytes and the expected `error` field value (empty for success cases).

A relay-role conformance suite MUST additionally demonstrate:
- Queue eviction at the 256-entry limit.
- TTL expiry purge (entries older than 24 hours excluded from drain).
- Correct `RELAY_AUTH_FAILED` on a tampered drain signature.
