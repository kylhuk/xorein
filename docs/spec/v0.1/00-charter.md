# 00 — Charter

## 1. Purpose

Xorein defines a wire protocol and reference runtime for a Discord-compatible
chat and voice platform that is:

- **Secure by default.** Every new conversation is end-to-end encrypted. A
  user must make an explicit, UI-labeled choice to operate in Clear mode.
- **Post-quantum resilient.** Identity signatures and key establishment use
  PQ-hybrid construction (Ed25519 + ML-DSA-65 signatures; X25519 + ML-KEM-768
  key agreement). An attacker who stores traffic today cannot decrypt it with
  a future quantum computer.
- **P2P-first.** Nodes communicate directly over libp2p. Relay nodes provide
  store-and-forward for offline delivery; they observe only ciphertext.
- **Resilient.** No single node is critical. Bootstrap and relay nodes can be
  replaced. Discovery cascades through mDNS, Kademlia DHT, Circuit Relay v2,
  DCUtR hole-punching, bootstrap lists, PEX, and manual peer lists.
- **Implementer-friendly.** A third party can build a compatible client using
  only this spec, the proto definitions in `proto/aether.proto`, and the
  conformance vector set in `pkg/spectest/`.

## 2. Scope

This specification covers:

1. Wire-level framing, signing, and verification (`02-canonical-envelope.md`).
2. Protocol family registration and capability negotiation
   (`03-protocol-registry-and-negotiation.md`).
3. All six security modes (`04-security-modes.md` and `10–15-mode-*.md`).
4. Cryptographic primitive profile, including PQ-hybrid construction
   (`01-cryptographic-primitives.md`).
5. Transport, discovery, and NAT traversal
   (`30-transport-and-noise.md`, `31-discovery.md`, `32-nat-traversal.md`).
6. All 13 protocol families (`40-family-peer.md` through `52-family-voice.md`).
7. Local HTTP control API (`60-local-control-api.md`).
8. Storage and key derivation (`70-storage-and-key-derivation.md`).
9. Conformance harness (`90-conformance-harness.md`).

Out of scope for v0.1:
- Anonymous routing / traffic analysis resistance (no anonymity guarantees
  are made; see threat model §3.3). A future anonymity layer (Aether Tunnel)
  is reserved for v0.2.
- Federation across distinct Xorein networks.
- QUIC transport (TCP is mandatory; QUIC is optional and additive).

## 3. Threat model

### 3.1 Trusted entities

- **Local endpoint.** The device running the node. Private key material lives
  only here. Endpoint compromise is out of scope for confidentiality guarantees.

### 3.2 Untrusted entities

- **Relay nodes.** May be operated by third parties. They observe only
  ciphertext for non-Clear traffic. A relay MUST NOT be able to read message
  content. This is an invariant enforced at every relay store point
  (see §3.4).
- **Bootstrap nodes.** Provide peer discovery only. They learn which peer IDs
  exist; they do not learn message content.
- **Network path.** All libp2p streams are encrypted at the transport layer
  (Noise XX protocol). An attacker on the network path sees encrypted,
  multiplexed streams.
- **Server operators.** Xorein servers (in the Discord sense) are logical
  groupings. The node hosting the server manifest learns member peer IDs and
  connection metadata, but not message plaintext.
- **Quantum adversary (store-now-decrypt-later).** An adversary that captures
  ciphertext today and decrypts it after a quantum computer becomes available.
  Mitigated by the ML-KEM-768 hybrid KEM and the Double Ratchet's forward
  secrecy chain.

### 3.3 Metadata

Xorein makes no anonymity guarantees for metadata. Peer IDs, server IDs,
channel IDs, membership lists, and timestamps are visible to parties with
whom a node communicates. A future anonymity layer is reserved for v0.2.

### 3.4 Relay opacity invariant

A relay node operating in `role=relay` MUST enforce:

1. Stored payloads MUST be encrypted (body is ciphertext, not plaintext) for
   all non-Clear scopes before enqueueing. Delivery requests with plaintext
   bodies for non-Clear scopes MUST be rejected with `RELAY_OPACITY_VIOLATION`.
2. The relay cannot decrypt stored payloads; it has no access to sender or
   recipient key material.
3. Queue entries expire according to TTL (default 24 h) and size limits
   (default 256 entries per recipient).

This invariant MUST be tested by the conformance harness (`90-conformance-harness.md §3`).

## 4. Design principles

**Additive-only evolution.** Every wire-level change MUST be additive. New
fields in JSON payloads use `omitempty`; new protobuf fields use reserved
numbers (outside 100-199); new capabilities are advertised as feature flags;
new protocol versions are distinct multistream IDs. A v0.1 node MUST
interoperate with a partial-v0.2 node using only the v0.1 protocol family.

**Explicit security modes.** Every conversation scope (DM, channel, voice
session) carries an explicit security mode string. The mode is non-negotiable
downward by infrastructure — only the originating parties can choose a mode.
UI clients MUST display the mode to the user.

**Capability-first negotiation.** Before processing any operation, both sides
exchange capability lists. Missing required capabilities cause a structured
`PeerStreamError`, not a silent fallback.

**Fail-closed.** Cryptographic verification failures, negotiation failures,
and missing capabilities all result in hard errors. There are no silent
silences or silent degradations.

**PQ-hybrid from day one.** The protocol does not have a classical-only phase.
Both Ed25519 and ML-DSA-65 signatures are required for every signed object.
Both X25519 and ML-KEM-768 KEM outputs are required for every key establishment
in Seal mode. This ensures no legacy classical-only traffic is ever emitted.

**No secrets in telemetry.** Telemetry and error messages MUST NOT contain key
material, private keys, plaintext message bodies, or other sensitive data.

## 5. Role taxonomy

All node roles run the same binary (`bin/aether`) with different capability
sets. The `--mode` flag or `mode` config key selects the role.

| Role | Canonical string | Purpose |
|------|-----------------|---------|
| Client | `client` | Full participant; sends/receives; hosts servers |
| Relay | `relay` | Store-and-forward for offline delivery; ciphertext only |
| Bootstrap | `bootstrap` | Kademlia DHT entry point; peer discovery; no message content |
| Archivist | `archivist` | Long-lived ciphertext history; manifests and coverage |

A node advertises its role in every peer exchange and in its manifest.

## 6. Normative language

Key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to
be interpreted as described in [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119)
and [RFC 8174](https://www.rfc-editor.org/rfc/rfc8174).

## 7. Versioning and evolution

This spec describes protocol version **v0.1**. Breaking wire changes require a
new spec version directory (`docs/spec/v0.2/`) and a new multistream major
version increment.

Additive changes within v0.1 use minor-version bumps in individual protocol
families per `03-protocol-registry-and-negotiation.md §5`. The `buf breaking`
tool gates all protobuf changes for wire compatibility.

## 8. Reference implementation

Canonical binary: `bin/aether` built by `make build` from the repository root.
Conformance vectors: `pkg/spectest/` (each subdirectory is a mode or family
chapter). The spec commit SHA is referenced in release notes for traceability.
