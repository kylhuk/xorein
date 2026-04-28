# 32 — NAT Traversal

This document specifies how Xorein nodes establish direct peer-to-peer
connections through Network Address Translation (NAT) using Circuit Relay v2
as a coordination medium and DCUtR (Direct Connection Upgrade through Relay)
for hole-punching.

## 1. Overview

Xorein uses a **relay-first, direct-upgrade** strategy:

1. Initial connection via Circuit Relay v2 (always available).
2. DCUtR hole-punch attempt to establish a direct path.
3. Fallback to Circuit Relay v2 if DCUtR fails.
4. Fallback to Xorein relay (store-and-forward) if Circuit Relay v2 fails.

Relaying adds latency and bandwidth cost. Implementations SHOULD attempt DCUtR
and report the connection type to the user layer.

## 2. Circuit Relay v2

### 2.1 Protocol

Xorein uses the [libp2p Circuit Relay v2 protocol](https://github.com/libp2p/specs/blob/master/relay/circuit-v2.md).

Relay nodes (`role=relay`) MUST support Circuit Relay v2. Bootstrap nodes MAY
support it. Client nodes SHOULD attempt to use relay nodes as Circuit Relay v2
coordinators before falling back to DHT-discovered relays.

### 2.2 Reservation flow

```
Client A (behind NAT) → Relay Node:
    /libp2p/circuit/relay/0.2.0/hop  RESERVE
    ← RESERVATION { relay_addrs, expire_in, voucher }

Client B → Relay Node → Client A (via /libp2p/circuit/relay/0.2.0/stop):
    B connects to A through the relay
```

Reservation TTL: minimum 30 minutes (implementations SHOULD renew at 80% of TTL).

A node that successfully obtains a relay reservation MUST announce the relay
address in its PEX records and manifest:

```
/ip4/<relay_ip>/tcp/<relay_port>/p2p/<relay_peer_id>/p2p-circuit/p2p/<my_peer_id>
```

### 2.3 Relay selection

Relay nodes are discovered via:
1. `--relay-addrs` configuration (most trusted; used first).
2. `relay.register` operation with known relay nodes.
3. DHT lookup for nodes with `cap.peer.relay` capability.

A node SHOULD maintain reservations on 2+ relay nodes for redundancy.

### 2.4 Relay quota

A relay MAY enforce quotas per connected client:
- Max 64 concurrent relay connections per client peer ID.
- Max 2 MiB/s throughput per client peer ID.
- Max reservation TTL: 24 hours.

Quotas MUST be documented in the relay's manifest (`history_retention_messages`
field overloaded → implementation SHOULD add a dedicated `relay_quota` field
in a future version).

## 3. DCUtR — Direct Connection Upgrade through Relay

### 3.1 Protocol

Xorein uses the [libp2p DCUtR protocol](https://github.com/libp2p/specs/blob/master/relay/DCUtR.md)
(`/libp2p/dcutr`) to attempt upgrading a relayed connection to a direct
connection via hole-punching.

### 3.2 DCUtR flow

```
A and B are connected via Circuit Relay.
A → B (via relay): /libp2p/dcutr  CONNECT { observed_addrs: [A's NAT-visible addrs] }
B → A (via relay): CONNECT { observed_addrs: [B's NAT-visible addrs] }
A and B simultaneously attempt direct connections to each other's
observed addresses (TCP simultaneous open / UDP hole-punch).
If either direction succeeds → direct connection established.
Relay connection closed (both sides).
If both directions fail after 3 attempts → remain on relay.
```

### 3.3 DCUtR attempt policy

- DCUtR SHOULD be attempted whenever a relayed connection is established.
- Max hole-punch attempts per pair: 3 (with 500 ms, 1 s, 2 s backoffs).
- DCUtR timeout per attempt: 5 seconds.
- If DCUtR fails, log the failure and remain on the Circuit Relay v2 path.
- DCUtR MUST NOT block the relayed stream; it runs concurrently.

### 3.4 IPv6 preference

When both IPv4 and IPv6 addresses are available, DCUtR SHOULD attempt IPv6
first (lower NAT traversal friction in dual-stack deployments).

## 4. Fallback to Xorein relay (store-and-forward)

If neither Circuit Relay v2 nor DCUtR succeeds (e.g., both peers behind
symmetric NAT with no relay reservation available), messages fall back to the
Xorein application-layer relay:

```
Sender → Relay (relay.store): encrypted delivery payload
Recipient → Relay (relay.drain): pull queued deliveries
```

This is the lowest-reliability path. Implementations MUST:
1. Display a latency indicator to the user when in relay-only mode.
2. Retry Circuit Relay v2 reservation every 10 minutes.
3. Attempt DCUtR again when the relay reservation is refreshed.

## 5. Connection type reporting

Implementations MUST track and report the connection type for each peer:

| Type | Description |
|------|-------------|
| `direct` | Direct TCP/QUIC connection (no relay) |
| `dcutr` | Direct connection established via DCUtR hole-punch |
| `circuit-relay-v2` | Relayed via Circuit Relay v2 |
| `xorein-relay` | Store-and-forward via Xorein relay node |

The connection type SHOULD be exposed via the local control API
(`/v1/peers/<peer_id>/connection`) and MAY be shown in UI.

## 6. Fallback sequence

```
1. Try direct TCP/QUIC connection (from known addresses).
2. Try DCUtR via an existing relay connection, if one exists.
3. Obtain Circuit Relay v2 reservation; connect via relay.
4. Attempt DCUtR via the new relay connection (concurrent).
5. If all fail: store message in Xorein relay queue.
```

Step 1 is skipped for the first connection to a new peer (no known address).
Steps 2-4 run in parallel with a 10-second timeout for DCUtR.

## 7. Security notes

- Circuit Relay v2 coordinators can observe connection metadata (who
  connects to whom) but cannot decrypt relayed traffic (Noise handles it).
- Xorein relay nodes for store-and-forward MUST NOT decrypt relay payloads
  (relay opacity invariant, `00-charter.md §3.4`).
- DCUtR-discovered IP addresses may reveal the node's true IP. This is
  acceptable in v0.1; the future Aether Tunnel layer will provide address
  hiding.
- Relay reservations are signed by the relay (voucher field in RESERVATION
  response). Clients MUST verify the voucher before accepting the reservation.
