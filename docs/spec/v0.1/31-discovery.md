# 31 — Discovery

This document specifies how Xorein nodes discover each other. Discovery is
layered: each layer narrows the search scope and adds resilience.

## 1. Discovery layers (in priority order)

| Layer | Scope | Mechanism | Staleness |
|-------|-------|-----------|-----------|
| 1. Live cache | Process | In-memory `livePeerRegistry` | Immediate |
| 2. mDNS | LAN subnet | `_aether._udp.local` | 15 min TTL |
| 3. Kademlia DHT | Global | Provider records in `/aether/kad/0.1.0` | 24 h TTL |
| 4. Bootstrap list | Configured | RPC `bootstrap.fetch` | Refreshed on connect |
| 5. Rendezvous | Configured | RPC `peer.exchange` with rendezvous node | Per-session |
| 6. Peer exchange (PEX) | Gossip | `peer.exchange` among connected peers | Best-effort |
| 7. Manual peers | Config | `--manual-peers` / `manual_peers` config | Always trusted |

Nodes iterate the layers in order and stop when enough peers are found
(threshold: max(5, target_connection_count)). The `discoveryInterval`
configuration key sets how often the discovery loop runs (default: 250 ms).

## 2. Layer 2 — mDNS (LAN)

### 2.1 Service name

```
_aether._udp.local
```

### 2.2 TXT record fields

mDNS announcements include a TXT record with:

```
peer_id=<libp2p PeerID>
addrs=<comma-separated multiaddrs>
role=<client|relay|bootstrap|archivist>
caps=<comma-separated capability flags>
```

### 2.3 Announcement interval

Nodes MUST announce via mDNS every **60 seconds** and MUST answer mDNS queries
within 500 ms. Announcements MUST include all current listen addresses.

### 2.4 mDNS peer validation

On receiving an mDNS record, the receiver MUST:

1. Parse `peer_id` and verify it matches the Noise identity from the
   subsequent Noise handshake.
2. Reject records where the PeerID is self.
3. Add the peer to the live cache with a 15-minute TTL.

mDNS-discovered addresses are not suitable for relay queuing because they
may be link-local. Relay targets MUST use publicly routable addresses.

## 3. Layer 3 — Kademlia DHT

### 3.1 DHT namespace

```
/aether/kad/0.1.0
```

This is the DHT protocol ID registered with the libp2p host. It ensures
Xorein nodes form their own DHT overlay distinct from the default IPFS DHT.

### 3.2 Provider records

A node announces itself as a provider for its own PeerID key. This allows
other nodes to find a peer's addresses by querying the DHT for `PeerID`.

```
DHT.PutProvider(key = SHA-256(peer_id), value = {addrs, role, signed_at})
```

Provider record TTL: **24 hours**. Nodes MUST re-announce every 22 hours
(10% before expiry).

### 3.3 Bootstrap to DHT seeding

When a node first joins, it connects to bootstrap nodes from `--bootstrap-addrs`
and uses them as DHT routing table seeds. The bootstrap nodes MUST be running
the Kademlia DHT and MUST NOT require authentication before serving routing
table entries.

### 3.4 DHT peer validation

Peers discovered via DHT MUST be contacted via a direct Noise connection
before being trusted. The `peer.info` operation (family `peer`) MUST succeed
before the peer is added to the live cache.

### 3.5 Server-rendezvous via DHT

To discover other members of a server, nodes store a server-specific provider
record:

```
key = SHA-256("xorein/server/" || server_id)
value = {peer_id, role, joined_at, signed_at}
```

Joining nodes query this key and use the results to find other members.

## 4. Layer 4 — Bootstrap list

### 4.1 Configuration

Bootstrap addrs are specified as a comma-separated list of multiaddrs in:

```
--bootstrap-addrs "/ip4/1.2.3.4/tcp/9999/p2p/QmBootstrap1,..."
```

or `bootstrap_addrs` in the JSON config file.

### 4.2 `bootstrap.register` operation

When a node connects to a bootstrap node, it calls:

```json
{
  "operation": "bootstrap.register",
  "payload": { "peer_id": "...", "addresses": [...], "role": "client", "caps": [...] }
}
```

The bootstrap node stores the registration with a 24-hour TTL.

### 4.3 `bootstrap.fetch` operation

To get peers from a bootstrap node:

```json
{
  "operation": "bootstrap.fetch",
  "payload": { "limit": 50, "exclude_peer_ids": ["..."] }
}
```

Response:
```json
{
  "peers": [
    { "peer_id": "...", "addresses": [...], "role": "...", "capabilities": [...] }
  ]
}
```

The bootstrap node MUST return at most `limit` entries (max: 200). It MUST
randomize the selection to prevent hot-spot clustering.

## 5. Layer 6 — Peer exchange (PEX)

### 5.1 `peer.exchange` operation

```json
{
  "operation": "peer.exchange",
  "payload": {
    "known_peer_ids": ["peer_id_1", "peer_id_2"],
    "request_new": true,
    "limit": 20
  }
}
```

Response:
```json
{
  "peers": [ { "peer_id": "...", "addresses": [...], "role": "...", "caps": [...] } ]
}
```

The responder returns peers it knows that are NOT in `known_peer_ids`, up to
`limit` entries (max: 50 per exchange). Returned peers MUST have been seen
within the last 24 hours.

### 5.2 PEX anti-flood

A node MUST NOT forward PEX-learned addresses without direct verification
(the `peer.info` round-trip). Unverified addresses may be stored tentatively
for up to 5 minutes before eviction.

## 6. Bootstrap canonical JSON format

Bootstrap address lists used in manifests and config are serialized as a
JSON array of multiaddr strings:

```json
["<multiaddr_with_peer_id>", ...]
```

Multiaddr strings MUST include the `/p2p/<peer_id>` suffix. Addresses without
a PeerID suffix MUST be rejected.

## 7. Peer record

A peer record shared via any discovery layer contains:

| Field | Type | Notes |
|-------|------|-------|
| `peer_id` | string | libp2p PeerID (multihash of Ed25519 pubkey) |
| `addresses` | []string | multiaddr strings with /p2p suffix |
| `role` | string | `client`, `relay`, `bootstrap`, `archivist` |
| `capabilities` | []string | declared capability flags |
| `signing_public_key` | string | base64url Ed25519 public key |
| `ml_dsa_65_public_key` | string | base64url ML-DSA-65 public key |
| `last_seen` | int64 | unix milliseconds |

Peer records MUST be signed by the peer's hybrid identity key before being
shared via PEX or stored in the DHT. Unsigned records MUST be rejected.

## 8. Discovery security notes

- mDNS records can be spoofed on a LAN. Always verify identity via Noise
  handshake before trusting a peer's peer_id claim.
- DHT provider records can be injected by malicious nodes. Validate freshness
  (`signed_at` within 25 hours) and verify the signature before using the
  addresses.
- Bootstrap nodes are semi-trusted (they learn what peers are online). Do not
  use bootstrap nodes as privacy infrastructure.
- Relay addresses announced via PEX SHOULD be verified by attempting a relay
  connection before adding to the relay list.
