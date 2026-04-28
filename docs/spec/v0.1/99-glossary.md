# 99 — Glossary

**AEAD** — Authenticated Encryption with Associated Data. A cipher mode that
provides both confidentiality and integrity in a single operation. The two AEAD
schemes in this spec are ChaCha20-Poly1305 (messages) and AES-128-GCM (media).

**Archivist** — A node role that stores long-lived ciphertext history with
coverage semantics. Archivists serve `cap.manifest`, `cap.sync`, and
`cap.archivist`.

**Bootstrap node** — A well-known, long-running node that acts as a Kademlia
DHT entry point for peer discovery. Bootstrap nodes do not store message content.

**Canonical form** — The deterministic serialization of a data structure used
as input to a signature or hash. See `02-canonical-envelope.md`.

**Capability flag** — A short string (e.g., `cap.chat`) exchanged during stream
setup to declare what operations a node supports. See §3 of
`03-protocol-registry-and-negotiation.md`. The authoritative list is in
`pkg/protocol/capabilities.go`.

**Channel** — In the security-mode sense: a broadcast-epoch E2EE mode for few
writers and many readers. In the Discord sense: a named grouping within a
server where messages are posted.

**CIRCL** — Cloudflare's cryptographic library (`github.com/cloudflare/circl`).
The reference Go implementation for ML-KEM-768 and ML-DSA-65 in Xorein.

**Circuit Relay v2** — A libp2p protocol that allows two peers to communicate
via an intermediary relay node when direct connection is impossible (e.g., both
peers are behind symmetric NAT). See `32-nat-traversal.md`.

**Clear** — The explicit plaintext security mode. Must be user-opted-in and
UI-labeled. Never the default for private scopes.

**Crowd** — Sender-key large-group E2EE mode for servers with >50 expected
members. Uses epoch rotation for forward secrecy.

**DCUtR** — Direct Connection Upgrade through Relay. A libp2p protocol that
uses a relay as a coordination point to attempt hole-punching and establish a
direct peer-to-peer connection. See `32-nat-traversal.md`.

**Delivery** — The JSON envelope used to transmit a message, voice frame, or
other payload peer-to-peer. Signed by the sender with the hybrid signature
scheme. See `pkg/node/wire.go`.

**DHT** — Distributed Hash Table. Used by libp2p (Kademlia variant) for
decentralized peer discovery. See `31-discovery.md`.

**Double Ratchet** — The forward-secure message ratchet used in Seal mode.
Provides per-message forward secrecy and post-compromise security. Specified
at https://signal.org/docs/specifications/doubleratchet/.

**DCUtR** — See "DCUtR" above.

**Ed25519** — The EdDSA signature scheme over Curve25519 (twisted Edwards
form). The classical signing identity key in Xorein. Always paired with
ML-DSA-65 in the hybrid signature scheme.

**Epoch** — In Crowd/Channel mode: a key generation period. An epoch ends when
a rotation trigger fires (member change, 1000-message limit, or 7-day timer).

**Feature flag** — Synonym for capability flag.

**Forward secrecy** — The property that compromise of long-term keys does not
compromise past session keys. Achieved in Seal by the Double Ratchet and in
Tree by MLS.

**HKDF** — HMAC-based Key Derivation Function. Per RFC 5869. Used throughout
the protocol for deriving symmetric keys from shared secrets. Every label used
in Xorein key derivation is defined in `pkg/crypto/labels.go`.

**Hybrid combiner** — The function that combines the classical (X25519) and
post-quantum (ML-KEM-768) shared secrets into a single master secret:
`HKDF-SHA-256(IKM = classical_secret || pq_secret, info = label)`.
See `01-cryptographic-primitives.md §5.1`.

**Hybrid KEM** — A key encapsulation mechanism that combines X25519 and
ML-KEM-768. Initiator performs both a classical X25519 exchange and an
ML-KEM-768 encapsulation to the responder's public key; shared secrets are
combined via the hybrid combiner.

**Hybrid signature** — A signature that uses both Ed25519 and ML-DSA-65 over
the same canonical payload. Both signatures MUST be present and verified
independently. See `01-cryptographic-primitives.md §6`.

**Identity** — A node's Ed25519 + ML-DSA-65 key pairs, peer ID, and profile.
Stored encrypted in the SQLCipher database. See `pkg/node/types.go:Identity`.

**Invite** — A signed, time-limited deeplink that a server owner issues for
others to join. Binds to a specific manifest hash. Signed with the hybrid
signature scheme.

**KAT** — Known-Answer Test. A test that provides fixed inputs and asserts a
fixed expected output, used to verify cryptographic primitive implementations
against reference vectors.

**Kademlia DHT** — The specific DHT variant used by libp2p for peer routing.
Nodes are organized by XOR distance in a 256-bit key space.

**Legacy window** — In Crowd/Channel mode: the number of prior epoch keys a
recipient may use to decrypt messages from senders that haven't yet advanced
their epoch. Bounded at ≤2 in v0.1.

**libp2p** — The P2P networking library underlying Xorein's transport layer.
Provides Noise encryption, Yamux multiplexing, Kademlia DHT, mDNS, Circuit
Relay v2, DCUtR, and stream multiplexing.

**Manifest** — A signed JSON document describing a server: its ID, owner,
addresses, capabilities, security mode, history policy, and channel list.
See `pkg/node/wire.go:Manifest`.

**mDNS** — Multicast DNS. Used for local network peer discovery on the subnet
`_aether._udp.local`. See `31-discovery.md`.

**MediaShield** — SFrame-based E2EE for voice and screen-share frames. Keys
derived from the parent scope's group material (MLS exporter for Tree,
HKDF-derived for Crowd/Channel).

**ML-DSA-65** — Module Lattice-based Digital Signature Algorithm (FIPS 204),
parameter set 65. The post-quantum signing algorithm in Xorein. Required
alongside Ed25519 for all signed objects (hybrid signature).

**ML-KEM-768** — Module Lattice-based Key Encapsulation Mechanism (FIPS 203),
parameter set 768. The post-quantum KEM in Xorein. Required alongside X25519
for all key establishment in Seal mode (hybrid KEM).

**MLS** — Messaging Layer Security (RFC 9420). The interactive group key
agreement protocol used in Tree mode. Xorein uses a hybrid ciphersuite
(ID 0xFF01) that extends the standard MLS_128_DHKEMX25519_AES128GCM_SHA256_Ed25519
ciphersuite with ML-KEM-768 and ML-DSA-65.

**multistream-select** — The libp2p protocol negotiation mechanism. Each
protocol is identified by a path string; the two sides negotiate which version
to use at the start of a stream via a byte-level exchange. See §8 of
`03-protocol-registry-and-negotiation.md`.

**NegotiationError** — A structured error (`PeerStreamError`) returned when
capability or protocol negotiation fails. Includes a machine-readable `code`
and missing capability names. See §9 of
`03-protocol-registry-and-negotiation.md`.

**Noise** — The cryptographic handshake protocol used by libp2p to establish
encrypted, authenticated transport connections between peers. Xorein uses the
Noise XX pattern.

**PeerID** — The base58/base64-encoded multihash of a node's public key.
Derived from the Ed25519 identity key. (ML-DSA-65 key material is not yet
incorporated into the peer-id derivation; see `30-transport-and-noise.md`.)

**PeerStreamError** — Defined in `proto/aether.proto`. The structured error
field in a `PeerStreamResponse`. See `02-canonical-envelope.md §1.3`.

**PeerStreamRequest** — Defined in `proto/aether.proto`. The native-protobuf
request envelope framing for all peer-to-peer stream operations. See
`02-canonical-envelope.md §1`.

**PeerStreamResponse** — Defined in `proto/aether.proto`. The native-protobuf
response envelope. See `02-canonical-envelope.md §1`.

**Peer exchange (PEX)** — A gossip mechanism for sharing known peers between
nodes, reducing reliance on bootstrap nodes. Implemented via the `peer.exchange`
operation in the peer family.

**Post-compromise security (PCS)** — The property that a session can
"recover" after key material is compromised, once new DH material is
contributed. Achieved in Seal by the Double Ratchet and in Tree by MLS.

**Post-quantum (PQ)** — Refers to cryptographic algorithms believed to be
secure against adversaries with quantum computers. In Xorein: ML-KEM-768 for
key exchange and ML-DSA-65 for signatures.

**Prekey** — A one-time or signed Diffie-Hellman public key published in
advance, used in X3DH to enable asynchronous key establishment. Prekey bundles
in Xorein also include the ML-KEM-768 public key for hybrid key establishment.

**Proto / protobuf** — Protocol Buffers. The wire encoding for all peer-stream
operation payloads and the `SignedEnvelope` format. Source of truth:
`proto/aether.proto`. Generated code: `gen/go/proto/aether.pb.go` (never
edit directly).

**Relay** — A node role that provides store-and-forward for offline delivery.
Relays observe only ciphertext; they cannot decrypt stored payloads.

**Relay opacity invariant** — The guarantee that a relay node never stores or
forwards plaintext for non-Clear scopes. Enforced in `peerRelayStore()`.
See `00-charter.md §3.4`.

**Role** — One of `client`, `relay`, `bootstrap`, `archivist`. All roles run
the same binary; role is set by the `--mode` flag or config.

**Seal** — The 1:1 E2EE security mode for direct messages. Uses hybrid X3DH
(X25519 + ML-KEM-768) and the Double Ratchet for forward secrecy.

**Security mode** — The encryption model applied to a conversation scope.
Exactly one of: Seal, Tree, Crowd, Channel, MediaShield, Clear.

**SFU** — Selective Forwarding Unit. A voice relay topology where one node
(the SFU) receives media from all participants and forwards selected streams
to other participants, without decrypting content. Used in MediaShield voice.

**SFrame** — Secure Frames. An E2EE media frame format defined in
[RFC 9605](https://www.rfc-editor.org/rfc/rfc9605). Used in MediaShield
mode for voice and screen-share frames.

**Sender key** — In Crowd/Channel mode: a symmetric key derived by HKDF that
a member uses to authenticate and encrypt their outgoing messages for an epoch.

**SignedEnvelope** — Defined in `proto/aether.proto`. Used inside
`PeerStreamRequest.payload` and `PeerStreamResponse.payload` for signed
binary artifacts (identity, manifest, prekey). See `02-canonical-envelope.md §2`.

**Snapshot** — The complete serialized state of a node at a point in time.
Returned by `/v1/state`. Contains identity, known peers, servers, DMs,
messages, and voice sessions.

**SQLCipher** — An encrypted SQLite extension used for local state persistence.
Key derived from a random salt and a secret via Argon2id (see
`70-storage-and-key-derivation.md`).

**StreamShield** — Segment-level E2EE for streaming media. Similar to
MediaShield but for recorded/live streams. Specified in `15-mode-mediashield.md`.

**Tree** — The hybrid MLS-based interactive group E2EE mode for servers with
≤50 members. Uses ciphersuite 0xFF01.

**Verification** — The process of checking a signature, a safety number, or
a SAS code. In the UI context: the user confirms a peer's identity via
out-of-band means. See `21-verification.md`.

**WebRTC** — Web Real-Time Communication. The peer-to-peer media protocol used
for voice and video in Xorein. Signaling (offer/answer/ICE) is carried over
the Xorein voice family stream; media frames are MediaShield-encrypted SFrame
payloads.

**Wire format** — The byte-level encoding for operations over libp2p streams:
4-byte big-endian length prefix followed by `proto.Marshal(PeerStreamRequest)`
or `proto.Marshal(PeerStreamResponse)`. See `02-canonical-envelope.md §1`.

**X25519** — The Diffie-Hellman function over Curve25519 (Montgomery form).
Used in X3DH and Double Ratchet for classical key exchange. Always combined
with ML-KEM-768 in the hybrid KEM.

**X3DH** — Extended Triple Diffie-Hellman. The asynchronous key establishment
protocol used in Seal mode. In Xorein, X3DH is extended with an ML-KEM-768
encapsulation step (hybrid X3DH). Specified at
https://signal.org/docs/specifications/x3dh/ with extensions in
`01-cryptographic-primitives.md §5`.

**Yamux** — Yet Another Multiplexer. The stream multiplexing protocol used on
top of Noise connections in libp2p, allowing multiple streams per connection.
