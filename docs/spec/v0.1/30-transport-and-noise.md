# 30 — Transport and Noise

This document specifies the libp2p transport configuration, peer identity
construction, and stream lifecycle for Xorein nodes.

## 1. libp2p host construction

Every Xorein node creates a single libp2p host at startup. The host MUST be
configured as follows:

### 1.1 Transport

| Parameter | Value |
|-----------|-------|
| TCP listener | Required; address from `--listen` flag or config |
| QUIC-v1 listener | Optional and additive; same port preferred |
| Max connections | 1024 inbound; 512 outbound (implementation SHOULD enforce) |
| Connection timeout | 30 seconds |

Implementations MUST support TCP. QUIC support is OPTIONAL and MUST be
additive — a node that only speaks TCP MUST interoperate with a node that
speaks both TCP and QUIC.

### 1.2 Security — Noise XX

All connections MUST use the **Noise XX pattern** with:

| Parameter | Value |
|-----------|-------|
| Pattern | XX (mutual authentication) |
| DH curve | X25519 |
| Cipher | ChaChaPoly (ChaCha20-Poly1305) |
| Hash | SHA-256 |
| Prologue | `b"/aether/noise/1.0"` |

Reference: [libp2p Noise spec](https://github.com/libp2p/specs/tree/master/noise).

The Noise layer provides hop-to-hop confidentiality and authentication. It is
not E2EE — relay nodes and bootstrap nodes terminate Noise connections
independently.

### 1.3 Multiplexer — Yamux

All connections MUST use **Yamux** for stream multiplexing.

| Parameter | Value |
|-----------|-------|
| Multiplexer | Yamux |
| Max stream window | 16 MiB |
| Max concurrent streams per connection | 1024 |
| Keep-alive interval | 60 seconds |

Streams are independent and do not share ordering guarantees. Each Xorein
operation runs on a separate stream; streams are short-lived (one
request-response pair, then closed).

### 1.4 Protocol handlers

The host registers one stream handler per supported protocol family:

```
host.SetStreamHandler("/aether/peer/0.1.0",     peerHandler)
host.SetStreamHandler("/aether/chat/0.1.0",     chatHandler)
host.SetStreamHandler("/aether/voice/0.1.0",    voiceHandler)
host.SetStreamHandler("/aether/manifest/0.1.0", manifestHandler)
host.SetStreamHandler("/aether/identity/0.1.0", identityHandler)
host.SetStreamHandler("/aether/sync/0.1.0",     syncHandler)
host.SetStreamHandler("/aether/dm/0.2.0",       dmHandler)
host.SetStreamHandler("/aether/groupdm/0.2.0",  groupDmHandler)
host.SetStreamHandler("/aether/friends/0.2.0",  friendsHandler)
host.SetStreamHandler("/aether/presence/0.2.0", presenceHandler)
host.SetStreamHandler("/aether/notify/0.2.0",   notifyHandler)
host.SetStreamHandler("/aether/moderation/0.2.0", moderationHandler)
host.SetStreamHandler("/aether/governance/0.2.0", governanceHandler)
```

Nodes MUST only register handlers for families they support. A relay node
MUST NOT register the chat or DM handlers.

## 2. Peer identity and PeerID

### 2.1 Identity key pair

The libp2p peer identity is derived from the node's **Ed25519 signing key**.
The Go `crypto.PrivKey` is created from the 64-byte Ed25519 private key
(seed || public key) per the libp2p Ed25519 key standard.

```
peer_id = libp2p.PeerIDFromEd25519PublicKey(identity.signing_public_key)
```

PeerID is the multihash of the Ed25519 public key in the `identity/ed25519`
multihash namespace. The ML-DSA-65 key is NOT incorporated into the PeerID
derivation in v0.1; it is carried as additional identity metadata in
`IdentityProfile.ml_dsa_65_public_key`.

### 2.2 Listen addresses and multiaddr

Node listen addresses use libp2p multiaddr format:

```
/ip4/0.0.0.0/tcp/12345          // IPv4 TCP
/ip4/0.0.0.0/udp/12345/quic-v1 // IPv4 QUIC-v1 (optional)
/ip6/::1/tcp/12345              // IPv6 TCP
```

Announced addresses include the PeerID suffix:

```
/ip4/1.2.3.4/tcp/12345/p2p/QmPeerID
```

When running behind NAT, announced addresses come from the Circuit Relay
reservation or the DCUtR-established address (see `32-nat-traversal.md`).

## 3. Stream lifecycle

### 3.1 Initiator side

```
1. host.NewStream(ctx, peerID, "/aether/<family>/<version>")
   // multistream-select handshake runs automatically
2. Write: [4-byte length][proto.Marshal(PeerStreamRequest)]
3. Read:  [4-byte length][proto.Marshal(PeerStreamResponse)]
4. Close stream
```

Timeout per step: 30 seconds for the stream open; 60 seconds for read.
Implementations MUST close the stream after receiving the response.

### 3.2 Responder side

```
1. stream arrives via host.SetStreamHandler callback
2. Read:  [4-byte length][proto.Marshal(PeerStreamRequest)]
3. Validate PeerStreamRequest (max 8 MiB, required fields present)
4. Run capability negotiation (03 §4)
5. Dispatch to operation handler
6. Write: [4-byte length][proto.Marshal(PeerStreamResponse)]
7. Close stream
```

The responder MUST close the stream after writing the response. Half-open
streams SHOULD be garbage-collected after 120 seconds.

### 3.3 Max payload enforcement

```
if length_prefix > 8_388_608:
    close stream immediately (no response)
```

Reading beyond the max size MUST be refused without allocating the buffer.

## 4. Connection management

### 4.1 Peer store

The libp2p peer store caches connection metadata (addresses, protocols, key):

- Address TTL (bootstrap-learned): 24 hours
- Address TTL (mDNS-learned): 15 minutes
- Address TTL (directly connected): refreshed on heartbeat

### 4.2 Reconnection

Implementations SHOULD use an exponential backoff strategy for reconnection
to known peers:

| Attempt | Delay |
|---------|-------|
| 1 | immediate |
| 2 | 5 seconds |
| 3 | 30 seconds |
| 4+ | min(2× previous, 10 minutes) |

After 3 consecutive failures, the peer is marked "degraded" and de-prioritized
in peer selection. It is not removed from the peer store.

## 5. Relay-mode listeners

When `--mode relay` and `--relay-listen` are configured, the relay node opens
a **second listener** on the relay address. This second listener is used for
store-and-forward drain connections from clients. The relay announces both its
normal listen address and its relay listen address in peer exchange.

## 6. Security notes

- The Noise layer provides hop-to-hop confidentiality only. End-to-end
  confidentiality requires the application-layer security modes
  (Seal/Tree/Crowd/Channel/MediaShield).
- A relay node MUST NOT inspect stream payloads beyond the `PeerStreamRequest`
  envelope fields. The `payload` bytes MUST be treated as opaque ciphertext.
- Private keys MUST NOT be logged, traced, or included in error messages.
- Implementations MUST reject connections with self-signed certificates
  (Noise `s` public key) that do not match the announced PeerID.
