# Aether Backend Protocol Specification (AEP-BACKEND) — v1.0 (Draft)

Status: Draft (normative language; implementation may be incomplete)
Last updated: 2026-02-17
License: CC-BY-SA 4.0 (specification text); code is expected to be permissive (MIT/BSD/Apache-2.0) per project guardrails.

This document specifies the backend-facing parts of the Aether Protocol: what must be implemented by headless nodes (relay/bootstrap) and auxiliary backend services (indexer, push relay, SFU/TURN) to support interoperable Aether clients.

“Backend” in Aether does not imply a central authority. It means infrastructure roles (still untrusted for content) that make the P2P network reliable and usable: discovery bootstrap, connection relays, encrypted store-and-forward, optional SFU/TURN forwarding, public directory indexing, and encrypted push wake-ups.

This specification intentionally describes behavior and semantics. Protobuf `.proto` schemas and multistream protocol IDs are treated as normative wire-format sources; this document defines the meaning and expected behavior.

---

## 1. Normative language

The key words “MUST”, “MUST NOT”, “REQUIRED”, “SHALL”, “SHALL NOT”, “SHOULD”, “SHOULD NOT”, and “MAY” are to be interpreted as described in RFC 2119.

If this document and the protobuf schema disagree, the schema is authoritative for the wire encoding, and this document is authoritative for semantics. Implementers MUST resolve disagreements by updating one or the other before claiming conformance.

---

## 2. Architecture and invariants

### 2.1 Single-binary, multi-mode invariant

All nodes run the same binary and select behavior via a runtime mode:

- `--mode=client`: full client node (UI + P2P core).
- `--mode=relay`: headless infrastructure node (bootstrap, circuit relay, store-forward, optional SFU/TURN).
- `--mode=bootstrap`: minimal DHT entrypoint node.

Implementations MUST NOT create privileged “authority node classes”. Differences are capability enablement only.

### 2.2 Protocol-first invariant

The protocol/spec contract is the product. Any official or reference backend MUST be implementable from this specification plus the protobuf schemas without requiring hidden behavior.

### 2.3 Compatibility invariant

- Minor protocol evolution MUST be additive at the protobuf schema level (new fields only; old clients ignore unknown fields).
- Breaking behavior MUST ship under new multistream protocol IDs and include downgrade negotiation, governance evidence, and multi-implementation validation.

### 2.4 Threat model baseline (backend)

- Relays, SFUs, TURN servers, indexers, and push relays are **untrusted for content confidentiality**.
- Backends MUST NOT require access to plaintext message content for core operation.
- Backends MAY observe unavoidable routing metadata; implementations MUST minimize, bound retention, and avoid logging sensitive identifiers where feasible.

---

## 3. Transport and addressing

### 3.1 libp2p transport profile

Backends (relay/bootstrap) MUST support:

- libp2p host with Noise-secured connections (recommended suite: `Noise_XX_25519_ChaChaPoly_SHA256`)
- QUIC transport as the preferred transport
- TCP as a fallback transport (for UDP-blocked environments)
- stream multiplexing as provided by libp2p (implementation-defined; QUIC provides multiplexing)

Backends SHOULD support:

- AutoNAT, DCUtR hole punching, and Circuit Relay v2 (relay nodes MUST support Circuit Relay v2).

### 3.2 Address format

All externally exchanged addresses MUST be libp2p multiaddrs.

For human exchange (QR codes, links), addresses SHOULD be encoded using base58btc or base32 without padding, and MUST be unambiguously decodable back into multiaddr form.

---



## 3.3 Discovery and bootstrap (Two-node guarantee)

Aether discovery is layered. Backends MUST support and MUST NOT block the layers relevant to their role.

### 3.3.1 Discovery cascade (normative order)

A node SHOULD attempt discovery in this order, proceeding to the next layer if the previous fails to yield a working peer connection:

0) Local peer cache (previously successful peers)
1) mDNS (LAN)
2) Hardcoded bootstrap list (shipped with binary)
3) DNS-based bootstrap discovery (TXT records)
4) DHT walking (Kademlia iterative lookups)
5) Rendezvous registration/lookup (for server/community affinity)
6) Peer exchange (PEX) with connected peers
7) Manual peer entry (multiaddr paste/QR)

Relays and bootstraps MUST interoperate with this cascade; in particular:

- Bootstrap nodes MUST be reachable via stable multiaddrs and SHOULD be present in the hardcoded bootstrap list.
- Relay nodes that offer DHT bootstrap MUST behave as well-formed DHT participants and SHOULD accept inbound connections from unknown peers (subject to abuse controls).

### 3.3.2 Local peer cache requirements (client-side, backend implications)

Relays/bootstraps MUST assume that many clients will reconnect using cached peer addresses rather than bootstrap infrastructure.

Therefore:
- Relays and bootstraps SHOULD maintain stable listen addresses where feasible to maximize cache usefulness.
- Relays SHOULD support address-change resilience (announce updates) to avoid “dead cache” failure for long-lived deployments.

### 3.3.3 DNS bootstrap discovery

Aether MAY publish bootstrap nodes via DNS TXT records. A recommended record shape:

- Name: `_aether._tcp.bootstrap.<domain>`
- Value: one or more multiaddrs (or multiaddrs wrapped in a minimal JSON list)

Clients MUST treat DNS bootstrap discovery as advisory only and MUST still validate remote peer identities as they would for any other connection.

### 3.3.4 Rendezvous (server affinity)

Rendezvous allows peers interested in the same server/community to find each other.

A rendezvous system MUST provide:
- `Register(server_id, peer_addrs, ttl)`
- `Discover(server_id, limit) -> peers`

This MAY be implemented:
- via the DHT under deterministic keys, or
- via libp2p rendezvous protocol if adopted, or
- via a server-event topic where members announce presence.

Relays MAY offer rendezvous acceleration (caching and serving recent registrations) but MUST NOT be a required authority.

### 3.3.5 Peer exchange (PEX)

Peers MAY exchange known peer multiaddrs periodically.

Relays SHOULD implement rate limits and peer scoring so PEX cannot be used to flood the network with bogus addresses.




## 3.4 Stream framing and message encoding (libp2p stream protocols)

For all libp2p stream subprotocols that exchange protobuf messages, peers MUST use a deterministic framing so multiple protobuf messages can be sent over a single stream.

Recommended framing (normative unless superseded by the `.proto`/implementation registry):

- Each protobuf message is encoded to bytes using proto3 encoding.
- Each message is preceded by a length prefix:
  - unsigned varint (protobuf varint encoding), representing the number of bytes in the message.
- Streams MAY carry multiple framed messages until closed.

Implementations MUST:
- enforce a maximum message size per subprotocol to prevent memory exhaustion
- close/reset the stream on invalid framing or oversized frames
- treat unknown/extra frames according to the subprotocol state machine (ignore if allowed, error otherwise)

If a subprotocol uses request/response semantics:
- the requester SHOULD close its write-half after sending the request
- the responder SHOULD send exactly one response (unless the protocol is explicitly streaming)


## 4. Versioning and negotiation

### 4.1 Multistream-select protocol IDs

Each Aether subprotocol is identified by a multistream-select protocol ID of the form:

`/aether/<subsystem>/<major>.<minor>.<patch>`

Examples:
- `/aether/chat/1.0.0`
- `/aether/manifest/1.0.0`

Implementations MUST negotiate the highest mutually supported version and MUST support downgrade to older minor versions within the same major version.

### 4.2 Protobuf schema evolution rules

For any `.proto` schema used by network protocols:

1. Field numbers MUST NOT be reused.
2. Existing field wire types MUST NOT change.
3. Removing fields MUST use `reserved` declarations.
4. Minor revisions MAY add fields only (compatibility).
5. Major revisions MUST use a new multistream protocol ID.

### 4.3 Capability negotiation

After establishing a secure libp2p connection, peers SHOULD exchange a `Capabilities` message describing:

- highest supported AEP version
- supported subprotocol versions
- feature flags (e.g., `voice.sframe`, `e2ee.mls`, `storeforward.dht.k20`)

Backends MUST make capability information available to internal routing decisions (e.g., whether a relay can offer SFU, whether store-forward supports the robust DHT replication profile).

---

## 5. Backend roles and required capabilities

### 5.1 Bootstrap node profile (`--mode=bootstrap`)

A bootstrap node is a minimal DHT entrypoint. It MUST:

- accept incoming libp2p connections
- participate in the Kademlia DHT for routing-table seeding
- advertise stable listen/announce multiaddrs

It SHOULD:

- minimize logging and retain no content
- expose a simple health endpoint and metrics (optional but recommended)
- be deployed in multiple geographic regions for resilience

A bootstrap node MUST NOT be required for ongoing operation after peers have established a working peer cache.

### 5.2 Relay node profile (`--mode=relay`)

A relay node provides “make the network work” services without learning plaintext. It MUST implement:

- DHT participation suitable for bootstrapping and steady-state routing
- Circuit Relay v2 service with admission control and abuse limits
- encrypted store-and-forward service (see Section 7)
- operational safety: quotas, TTL enforcement, and privacy-preserving logging

It MAY implement:

- SFU forwarding for voice/video/screen (E2EE media via SFrame at endpoints)
- TURN relay for WebRTC connectivity
- directory indexer service (separate process is recommended)
- push notification relay (separate process is recommended)
- history archival service and history sync support (v0.7+ profile)

### 5.3 Relay admission control, limits, and abuse protections

Relay nodes exist in hostile environments. They MUST implement resource limits and SHOULD ship with safe defaults.

#### 5.3.1 Circuit Relay v2 limits (recommended defaults)

A relay SHOULD enforce at least:

- `max_circuits`: default 1024 (operator configurable)
- `max_circuit_duration`: default 2 minutes, renewable (operator configurable)
- `max_circuit_bandwidth`: default 1 Mbps per circuit (operator configurable)
- per-peer reservation limits (to prevent one peer consuming all circuits)
- per-IP limits MAY be used but MUST NOT be relied upon as the only abuse control

Relays MUST provide deterministic rejection reasons for:
- circuit capacity exhausted
- peer exceeds reservation quota
- peer is banned/blocked by local policy
- invalid reservation protocol usage

#### 5.3.2 Store-and-forward limits (recommended defaults)

Relays MUST enforce:

- maximum stored bytes (quota): operator configurable; recommended default 10 GB
- maximum object size: operator configurable; recommended default 1–4 MB per mailbox item for v0.1 DM store-forward
- TTL bounds: enforce a maximum TTL regardless of sender request
- deletion policy: expired objects MUST be deleted (bounded delay acceptable)

Relays SHOULD implement:
- per-sender rate limits for store-forward `Put`
- spam controls tied to proof-of-work or reputation where available (v0.6+ profile)

#### 5.3.3 Privacy-preserving logging (normative)

Relays MUST NOT log:
- message ciphertext blobs
- full recipient mailbox identifiers
- long-term keys or session secrets

Relays SHOULD log only:
- aggregate counters
- coarse timing buckets
- anonymized/hardened identifiers (e.g., truncated hashes) if an identifier is needed for debugging

---

## 6. Shared data model (backend-relevant objects)

### 6.1 Identifiers

All identifiers listed below are binary on the wire; text encodings are for UX/logging only.

- `PeerID`: libp2p peer identifier.
- `ServerID`: stable identifier for a server/community.
- `ChannelID`: stable identifier for a channel within a server.
- `MessageID`: stable identifier for a message/event (see Section 8.2).

All IDs MUST be collision-resistant. Implementations SHOULD use 128-bit or 256-bit identifiers (or hashes thereof) depending on object lifetime and global scope.

### 6.2 Server manifest

A `ServerManifest` is the canonical signed description of a server: name, channels, join policy, security-mode policy, and moderation policy references.

Backends MUST treat the manifest as a signed object and MUST NOT rewrite it. Relays MAY cache manifests (public metadata only) if the manifest is intended to be publicly retrievable; otherwise manifests SHOULD be end-to-end protected and distributed among authorized peers.

### 6.3 Directory entry (public listing)

A `DirectoryEntry` is a public, signed summary for discovery. It MUST contain only public metadata, such as:

- ServerID
- manifest hash (for integrity binding)
- name, short description, tags, languages, region hints
- join policy (invite-only / request / open)
- safety labels (NSFW, topic, minimum age)
- optional suggested relays/SFUs
- signature by server owner key

Backends (indexers) MUST verify signatures and MUST serve the signed object unchanged.

### 6.4 DHT record types and key layout (backend-critical)

Aether uses a Kademlia DHT with a custom protocol prefix (recommended: `/aether/kad/1.0.0`).

The DHT stores *records*. A record consists of:
- a key (byte string)
- a value (byte string, typically a protobuf message)
- an implementation-defined TTL/expiration mechanism

Backends that participate in the DHT (relay/bootstrap) MUST:
- store and serve DHT records according to Kademlia behavior
- enforce maximum record size limits (to prevent abuse)
- enforce record TTLs where the DHT implementation supports them
- avoid logging record values

#### 6.4.1 Record classes (minimum set)

The system SHOULD define at least these record classes:

1) `IDENTITY_PROFILE`
- Key: derived from the user/account identifier.
- Value: signed user profile object (public or access-controlled blob).

2) `PREKEY_BUNDLE`
- Key: derived from the user/account identifier.
- Value: signed prekey bundle for DM establishment (Seal/X3DH).

3) `SERVER_MANIFEST`
- Key: derived from `ServerID`.
- Value: signed `ServerManifest` (or encrypted manifest capsule, depending on join policy).

4) `DIRECTORY_ENTRY`
- Key: derived from `ServerID` and a discovery namespace.
- Value: signed `DirectoryEntry` (public metadata only).

5) `STORE_FORWARD_MAILBOX_ITEM` (v0.7+ robust profile)
- Key: derived from recipient mailbox identifier and message identifier.
- Value: store-forward envelope (ciphertext).

6) `SERVER_AUDIT_EVENT` (v0.4+)
- Key: derived from `ServerID` and event log cursor.
- Value: signed moderation/audit event.

These classes are semantic; the exact on-wire representation MUST be defined in `.proto` schemas.

#### 6.4.2 Deterministic key derivation (recommended pattern)

To prevent collisions and enable safe multi-tenancy, DHT keys SHOULD be derived as:

`key = SHA-256("aether:" + record_type + ":" + namespace + ":" + raw_id_bytes)`

Where:
- `record_type` is a stable ASCII token (e.g., `server_manifest`)
- `namespace` is optional, used to avoid collisions across environments or time windows
- `raw_id_bytes` is the canonical binary identifier (e.g., `ServerID`)

The implementation MUST publish the exact derivation rules used (including encoding details) as part of the authoritative protocol registry.

#### 6.4.3 TTL guidance (recommended defaults)

- `IDENTITY_PROFILE`: 90 days, refresh on update
- `PREKEY_BUNDLE`: 7–30 days, refresh periodically
- `SERVER_MANIFEST`: 30 days, refresh on update (or pinned by server peers)
- `DIRECTORY_ENTRY`: 7–30 days, refresh periodically; clients MUST check freshness
- `STORE_FORWARD_MAILBOX_ITEM`: 7 days (v0.1 relay-local), 30 days (v0.7+ robust)
- `SERVER_AUDIT_EVENT`: long-lived (server policy), but storage is best-effort in a P2P network

Backends MAY enforce stricter TTLs to protect themselves, but MUST behave predictably and MUST disclose effective TTL behavior in operator documentation.

#### 6.4.4 Proof-of-work stamps (v0.6+ anti-abuse)

To reduce Sybil/spam pressure, the protocol MAY require proof-of-work (PoW) stamps on selected record publications and requests. A PoW stamp is intended to be:
- cheap to verify
- moderately expensive to create (seconds on consumer hardware)
- adjustable via a difficulty parameter

Recommended stamp shape (Hashcash-style, informative):

- Inputs:
  - `resource` (what is being protected; e.g., `identity:<account_id>` or `push:<to_account_id>`)
  - `timestamp_bucket` (coarse time, e.g., hour/day; prevents indefinite reuse)
  - `nonce` (search space)
  - `difficulty_bits` (required leading-zero bits)

- Condition:
  - `SHA-256(resource || timestamp_bucket || nonce)` has at least `difficulty_bits` leading zero bits.

Where PoW is applied, backends MAY:
- reject publications/requests without valid stamps
- rate-limit stampless requests more aggressively
- adjust required difficulty by observed load/abuse

PoW MUST NOT be relied upon as the only abuse control; it complements quotas, peer scoring, and reputation.

---

## 7. Encrypted store-and-forward (offline delivery)

Store-and-forward provides delivery when recipients are offline or unreachable directly. This is a backend-critical service.

### 7.1 Security and privacy requirements

A store-and-forward backend (relay, DHT replication node, or archivist) MUST:

- store only ciphertext payloads
- avoid logging message payloads
- minimize retained metadata (store only what is required for delivery)
- enforce TTL expiration and delete expired payloads
- apply storage quotas to bound abuse impact

The backend MUST NOT have access to plaintext message content.

### 7.2 v0.1 baseline profile (relay-local)

Baseline store-forward (sufficient for v0.1–v0.2 MVPs):

- relay stores encrypted DM payloads for offline peers
- default TTL: 7 days (configurable)
- retrieval occurs when recipient reconnects to the relay

### 7.3 v0.7 robust profile (DHT replicated)

Robust store-forward (v0.7+ target profile):

- payloads are stored under deterministic DHT keys
- replication target: `k = 20` replicas across the DHT keyspace
- default TTL: 30 days (configurable; must be enforced)
- retrieval supports parallel fetch from multiple replicas
- purge is best-effort: expiration is authoritative; explicit purge accelerates removal but is not relied upon for confidentiality

### 7.4 Store-forward envelope semantics

Store-forward messages MUST be wrapped in an envelope that includes:

- recipient routing identifier (PeerID or device mailbox ID)
- ciphertext blob (opaque to backend)
- creation timestamp
- expiry timestamp (or TTL)
- optional small routing hints (e.g., server scope) subject to metadata-minimization policy
- sender authentication binding (signature or AEAD tag) verifiable by recipient (backend does not validate)

The envelope MUST NOT include plaintext names, usernames, channel titles, or message bodies.

### 7.5 Store-forward API (conceptual)

Implementations MUST provide a mechanism equivalent to:

- `Put(envelope) -> ack`
- `Get(recipient, cursor) -> stream<envelope>`
- `Ack(receipt_ids) -> ack`
- `Purge(criteria) -> ack` (optional)

This MAY be implemented as:
- a libp2p stream protocol (`/aether/store/1.x.x`)
- DHT storage with client-side retrieval
- or both (recommended: relay-local for speed + DHT replication for resilience)

---

## 8. PubSub topics and event propagation

### 8.1 Topic naming

For server-scoped GossipSub topics, the naming convention SHOULD be:

- Text channel messages: `/aether/srv/<server_id>/ch/<channel_id>`
- Server event log (membership/moderation): `/aether/srv/<server_id>/events`
- Voice signaling: `/aether/srv/<server_id>/vc/<voice_channel_id>`

Backends that participate in PubSub (relays) MUST enforce topic subscription controls according to local policy and abuse mitigation configuration (peer scoring, rate limits).

### 8.2 Event identifiers and idempotency

All events delivered via PubSub or store-forward MUST be idempotent at the receiver.

Receivers MUST ignore duplicates (e.g., by `MessageID`), and MUST handle out-of-order arrival.

### 8.3 Moderation event propagation

Moderation actions MUST be represented as signed events (e.g., `Redaction`, `Timeout`, `Ban`, `SlowModeUpdate`). Official clients SHOULD enforce signed moderation events.

Relays MAY cache and serve moderation events to assist late joiners and recovery, but MUST treat them as opaque signed objects.

### 8.4 GossipSub peer scoring and flood protection (v0.6+ hardening)

Backends that participate in PubSub (especially relays) SHOULD apply peer scoring and flood protection to mitigate spam and network abuse.

Minimum recommended controls:

- Per-peer publish rate limits per topic (token-bucket or leaky-bucket).
- Validation of message size limits per topic.
- Penalties for:
  - invalid signatures (when verifiable without plaintext)
  - repeated oversized messages
  - repeated publish bursts beyond configured thresholds
  - gossip/ihave/iwant abuse patterns (implementation-defined)
- Decay of penalties over time to allow recovery from transient faults.

Relays SHOULD expose aggregate metrics for:
- dropped messages by reason
- peer score distributions (bucketed)
- topic-level publish rates

Implementations MUST ensure that peer scoring does not become a privacy leak (e.g., do not expose per-peer score values publicly).

---

## 9. Directory indexer service (optional, non-authoritative)

Full-text search over DHT is not expected to scale; indexers are optional helpers.

### 9.1 Trust model

- Indexers are community-run and non-authoritative.
- Indexers MUST return signed `DirectoryEntry` objects so clients can verify authenticity.
- Indexers SHOULD also sign their response envelope to provide integrity over the result set and pagination.

### 9.2 Indexer API (minimum)

An indexer SHOULD expose an HTTPS API with at least:

- `GET /v1/search?q=<query>&tags=<...>&lang=<...>&region=<...>&min_members=<n>&max_members=<n>&cursor=<...>`
- `GET /v1/server/<server_id>`

Responses MUST include:
- the signed `DirectoryEntry`
- a stable cursor/pagination token if paginated
- a deterministic ordering rule (e.g., relevance then freshness)

### 9.3 Privacy considerations

Indexers SHOULD support privacy-preserving query options:

- query padding (clients can send decoy queries)
- querying multiple indexers and merging locally
- optional proxy/Tor usage (client-side)

Indexers MUST NOT claim to provide anonymity.

---

## 10. Encrypted push notification relay (ENR)

Mobile background delivery often requires APNs/FCM-style wakeups. ENR is a backend helper that forwards encrypted wake payloads.

### 10.1 Security requirements

- ENR MUST not have access to plaintext message content.
- ENR MUST send minimal payloads; the payload SHOULD be a wake “ping” plus an opaque ciphertext blob that only the device can decrypt.
- Message content retrieval MUST happen via P2P after wake.

### 10.2 Device registration

ENR MUST support device-token registration:

- mapping from an Aether user/device identity to an APNs/FCM token
- revocation and rotation of tokens
- rate limiting and abuse controls to prevent token spamming

How the mapping is authenticated is implementation-defined, but MUST prevent unauthorized third parties from registering tokens for another account.

### 10.3 Delivery semantics

ENR MUST implement retry with bounded backoff and SHOULD provide delivery receipts where the push provider supports them.

ENR MUST not retain payloads longer than necessary for retries (bounded retention; default ≤ 24 hours).

### 10.4 ENR API surface (recommended minimum)

ENR is intentionally simple. A recommended minimal HTTPS API:

1) Register device token
- `POST /v1/devices/register`
- Auth: a device-bound Aether token (implementation-defined) OR a signed request by the device identity key.
- Body (JSON example):
  - `device_id` (stable per install)
  - `account_id` (public identifier)
  - `provider` (`apns` | `fcm`)
  - `token` (provider token)
  - `token_version` (monotonic counter for rotation)
  - `timestamp`

2) Unregister device token
- `POST /v1/devices/unregister`
- Body: `device_id`, `account_id`, `timestamp`

3) Send wake ping (called by peers or by the local node acting on behalf of peers)
- `POST /v1/push/send`
- Body:
  - `to_account_id`
  - `to_device_id` (optional; if omitted, send to all registered devices for account)
  - `payload` (opaque ciphertext bytes, base64)
  - `ttl_seconds` (MUST be capped by ENR; recommended max 3600)
  - `collapse_key` (optional; enables coalescing)
  - `timestamp`
- Response:
  - `accepted` boolean
  - `message_id` (for tracking)
  - `provider_status` (optional)

Authentication and authorization for `push/send` MUST prevent third parties from spamming arbitrary users. Acceptable approaches include:
- only allowing sends from accounts that are already “contacts” (verified relationship)
- requiring proof-of-work on `push/send` requests (v0.6+ profile)
- applying strong per-sender quotas and heuristics

---

## 11. SFU and TURN services (relay optional)

SFU/TURN exist for connectivity and scale. They are untrusted for content.

### 11.1 Media E2EE requirement (MediaShield)

If an SFU is used for voice/video/screen, endpoints MUST encrypt media frames end-to-end using SFrame (or an equivalent E2EE frame mechanism). The SFU MUST forward encrypted frames without decrypting them.

### 11.2 Topology switching

Backends that offer SFU MUST support deterministic topology switching policies (P2P mesh for small groups; SFU for larger groups). Exact thresholds are policy knobs, but MUST be disclosed to clients via capabilities.

### 11.3 TURN

TURN servers MUST be treated as untrusted and MUST not terminate media encryption. TURN is transport; it forwards SRTP/SFrame traffic.

---

## 12. Bot API (local backend surface)

Bots connect to an Aether node via a controlled local API surface (gRPC is recommended). This is a backend surface because it is served by the node daemon, not by the UI.

This section specifies minimum requirements for the **node-side Bot API**. The on-wire schemas are defined by the Bot API `.proto` files; this section defines behavior.

### 12.1 Binding and transport

1. The Bot API MUST be disabled by default OR MUST bind to loopback only (`127.0.0.1` / `::1`) unless explicitly enabled by the operator/user.
2. If the Bot API is bound to a non-loopback interface, it MUST require strong authentication (mTLS recommended) and SHOULD provide an allowlist of source IPs.
3. The Bot API SHOULD support TLS (mTLS preferred) even on loopback for defense-in-depth when multiple local users exist.

### 12.2 Authentication and authorization

The Bot API MUST implement:

- Authentication: verify the connecting bot process is authorized (token and/or certificate based).
- Authorization: enforce permissions using the server RBAC/permission model (server-level and channel-level overrides).

A bot MUST have an explicit identity (`BotID`) and MUST be issued credentials that can be revoked.

The node MUST support revocation:
- immediate revocation on explicit user action
- automatic revocation on key rotation where applicable

### 12.3 Event delivery, ordering, and replay

The Bot API MUST provide an event subscription mechanism with:

- Deterministic ordering: events delivered in the same order they are committed to the local node’s event log for that server.
- At-least-once delivery: reconnecting bots MAY receive duplicates; bots MUST dedupe by event ID.
- Replay support: a bot MUST be able to resume from a cursor/sequence number after reconnect.

Recommended (minimum) event stream semantics:

- Each event has:
  - `event_id` (globally unique within the node)
  - `server_id` (scope)
  - `sequence` (monotonic per server)
  - `timestamp`
  - `type` + type-specific payload
- The stream supports:
  - `Subscribe(server_id, from_sequence)` or `Subscribe(server_id, from_event_id)`

### 12.4 Commands and idempotency

The Bot API MUST support command invocation with:

- Idempotency key: caller-supplied `idempotency_key` so retries do not create duplicate side effects.
- Explicit acknowledgment and completion:
  - `ACCEPTED` (validated and queued)
  - `COMPLETED` (executed successfully)
  - `FAILED` (with stable error taxonomy)
- Timeouts:
  - server-defined max execution time
  - deterministic timeout error if exceeded

Command execution MUST be permission-checked at the time of execution (not only at submission), to avoid TOCTOU issues when roles/permissions change.

### 12.5 E2EE boundary and bot trust

In strict E2EE channels (Seal/Tree/Crowd/Channel), bots and webhooks MUST NOT receive plaintext by default.

If a bot is granted plaintext access:
- that channel MUST be explicitly labeled as granting plaintext to the bot
- the bot MUST be treated as an endpoint in the security model
- the node SHOULD surface bot trust and permissions to users/admins


---

## 13. Operational requirements

### 13.1 Configuration

Relay and bootstrap nodes MUST support configuration via file and/or CLI flags.

#### 13.1.1 Configuration precedence (recommended)

If both file and CLI flags are supported, the recommended precedence is:

1) CLI flags
2) Config file
3) Built-in defaults

Implementations MUST provide a way to print the **effective** resolved configuration (with secrets redacted) for operator debugging.

#### 13.1.2 Minimum required configuration fields (relay)

A relay configuration MUST cover at least:

- listen/announce multiaddrs
- enablement flags for relay/store-forward/SFU/TURN/metrics
- quotas (storage, bandwidth, max circuits)
- TTL defaults (store-forward and archive where applicable)
- metrics/health endpoints and bind addresses

#### 13.1.3 Example `relay.toml` (informative)

```toml
[node]
mode = "relay"
listen_addrs = [
  "/ip4/0.0.0.0/udp/4001/quic-v1",
  "/ip4/0.0.0.0/tcp/4001"
]
announce_addrs = [
  "/ip4/203.0.113.1/udp/4001/quic-v1",
  "/ip4/203.0.113.1/tcp/4001"
]

[relay]
enabled = true
max_circuits = 1024
max_circuit_duration = "2m"
max_circuit_bandwidth = "1Mbps"

[store_forward]
enabled = true
storage_path = "/data/store"
max_storage = "10GB"
message_ttl = "30d"

[sfu]
enabled = false
max_rooms = 50
max_participants_per_room = 100

[turn]
enabled = false

[metrics]
enabled = true
listen_addr = "127.0.0.1:9090"

[health]
enabled = true
listen_addr = "127.0.0.1:9091"
```

### 13.2 Observability (privacy-preserving)

Backends SHOULD expose:

- Prometheus metrics (`/metrics`)
- structured logs (JSON recommended)
- a basic health endpoint (`/healthz`)

#### 13.2.1 Logging restrictions (normative)

Logs MUST NOT include:
- message plaintext
- attachment plaintext
- long-term key material
- store-forward ciphertext blobs
- full recipient identifiers if avoidable (prefer hashed or bucketed)

#### 13.2.2 Reason-coded diagnostics

Backends SHOULD emit reason-coded diagnostics suitable for the “no-limbo” UX invariant, without leaking sensitive metadata.

Recommended practice:
- define a stable enumeration of reason codes for: connection, relay admission, store-forward, DHT, pubsub.
- expose aggregate counts by reason code.

### 13.3 Safety and abuse-control hooks (backend)

Backends SHOULD provide operator-configurable hooks for:

- allowlist/denylist of peer IDs (local policy)
- per-peer quotas (circuits, bandwidth, store-forward puts)
- proof-of-work verification policies for identities and/or send requests (v0.6+ profile)
- gossip peer scoring toggles and thresholds
- emergency “lockdown mode” that rejects new reservations/puts while keeping existing circuits alive (operator choice)

Backends MUST make these controls observable (counters/metrics) so operators can see when they are engaging.


---

## 14. Conformance requirements (backend)

A backend implementation claiming conformance to this spec MUST provide:

1. A documented list of supported multistream protocol IDs and versions.
2. A documented list of supported feature flags/capabilities.
3. Store-and-forward correctness:
   - TTL enforcement and purge behavior
   - quota enforcement
   - ciphertext opacity (no plaintext processing)
4. Circuit Relay v2 correctness:
   - reservation limits, session duration bounds, and abuse protections
5. DHT bootstrap correctness:
   - stable connectivity for new peers in typical environments
6. Optional services (if claimed): SFU/TURN, indexer, ENR push relay.

---

## Appendix A. Suggested protocol registry (informative)

This section is informative and may be adjusted by the authoritative protocol registry in the repository.

- `/aether/kad/1.0.0` — DHT routing
- `/aether/identity/1.0.0` — identity/profile/prekey exchange
- `/aether/manifest/1.0.0` — server manifest replication
- `/aether/chat/1.0.0` — text message envelopes + channel semantics
- `/aether/voice/1.0.0` — voice signaling
- `/aether/file/1.0.0` — file transfer
- `/aether/store/1.0.0` — store-and-forward protocol (relay-local API)
- `/aether/sync/1.0.0` — history sync
- `/aether/mod/1.0.0` — moderation event objects (signed)
- `/aether/rep/1.0.0` — reputation/reporting surfaces

Implementations SHOULD also maintain a machine-readable registry file (e.g., `docs/protocol/registry.json`) listing supported protocol IDs and their status (active/deprecated).

---

## Appendix B. Open items to close before calling this “v1.0 final”

1. Bind this spec to the exact `.proto` schema set used in the codebase (field numbers, names, and required/optional semantics).
2. Publish the authoritative protocol registry and capability flag registry.
3. Specify the canonical derivation/encoding for ServerID/ChannelID/MessageID used in deeplinks and DHT keys.
4. Specify the exact DHT key layout for:
   - ServerManifest publication
   - DirectoryEntry publication
   - store-and-forward replication objects (v0.7 profile)
5. Specify the exact ENR push relay API (device registration, send request format) and its authentication scheme.
6. Specify SFU signaling objects and election rules as normative message flows (currently described at a high level).

