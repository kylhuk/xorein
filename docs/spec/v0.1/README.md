# Xorein Protocol Specification — v0.1

Xorein is a secure-by-default, P2P-first chat protocol. This directory is the
normative specification for the v0.1 wire protocol. A third party implementing
only the documents here, the conformance vectors in `91-test-vectors/`, and the
protobuf definitions in `proto/aether.proto` should be able to build a client
that interoperates with any v0.1 node.

## Document index

### Foundations

| File | Title | Status |
|------|-------|--------|
| [00-charter.md](00-charter.md) | Goals, threat model, normative language, design principles | Normative |
| [01-cryptographic-primitives.md](01-cryptographic-primitives.md) | Locked crypto profile: classical + PQ-hybrid | Normative |
| [02-canonical-envelope.md](02-canonical-envelope.md) | Wire envelope format (native protobuf) + signing rules | Normative |
| [03-protocol-registry-and-negotiation.md](03-protocol-registry-and-negotiation.md) | Protocol IDs, capability negotiation, multistream handshake | Normative |
| [04-security-modes.md](04-security-modes.md) | Mode overview, selection rules, relay opacity policy | Normative |
| [99-glossary.md](99-glossary.md) | Term definitions | Normative |

### Transport & Connectivity

| File | Title | Status |
|------|-------|--------|
| [30-transport-and-noise.md](30-transport-and-noise.md) | libp2p host, Noise XX, Yamux, stream lifecycle | Normative |
| [31-discovery.md](31-discovery.md) | mDNS, Kademlia DHT, PEX, bootstrap, peer records | Normative |
| [32-nat-traversal.md](32-nat-traversal.md) | Circuit Relay v2, DCUtR hole-punching, fallback | Normative |

### Security Modes

| File | Title | Conformance Wave |
|------|-------|-----------------|
| [10-mode-clear.md](10-mode-clear.md) | Clear — explicit plaintext opt-in, downgrade prevention | W5 |
| [11-mode-seal.md](11-mode-seal.md) | Seal — hybrid X3DH + Double Ratchet 1:1 E2EE | W1 |
| [12-mode-tree.md](12-mode-tree.md) | Tree — hybrid MLS RFC 9420 group E2EE | W2 |
| [13-mode-crowd.md](13-mode-crowd.md) | Crowd — epoch sender-key large-group E2EE | W3 |
| [14-mode-channel.md](14-mode-channel.md) | Channel — broadcast epoch E2EE (few writers) | W3 |
| [15-mode-mediashield.md](15-mode-mediashield.md) | MediaShield — RFC 9605 SFrame voice/video E2EE | W4 |

### Protocol Families (Discord-parity)

| File | Protocol ID | Description | Wave |
|------|-------------|-------------|------|
| [40-family-peer.md](40-family-peer.md) | `/aether/peer/0.1.0` | Core peer transport, relay, bootstrap | W0 |
| [41-family-chat.md](41-family-chat.md) | `/aether/chat/0.1.0` | Server channel messaging | W3 |
| [42-family-manifest.md](42-family-manifest.md) | `/aether/manifest/0.1.0` | Server manifests, signed invites, RBAC | W3 |
| [43-family-identity.md](43-family-identity.md) | `/aether/identity/0.1.0` | Prekey bundles, identity announce | W1 |
| [44-family-sync.md](44-family-sync.md) | `/aether/sync/0.1.0` | Own-device state sync | W1 |
| [45-family-dm.md](45-family-dm.md) | `/aether/dm/0.2.0` | Seal-mode direct messages | W1 |
| [46-family-groupdm.md](46-family-groupdm.md) | `/aether/groupdm/0.2.0` | Tree-mode group DMs | W2 |
| [47-family-friends.md](47-family-friends.md) | `/aether/friends/0.2.0` | Friend requests, blocks, list | W5 |
| [48-family-presence.md](48-family-presence.md) | `/aether/presence/0.2.0` | Online status, typing indicators | W5 |
| [49-family-notify.md](49-family-notify.md) | `/aether/notify/0.2.0` | Notifications, mentions, acks | W5 |
| [50-family-moderation.md](50-family-moderation.md) | `/aether/moderation/0.2.0` | Kick, ban, mute, slow mode | W6 |
| [51-family-governance.md](51-family-governance.md) | `/aether/governance/0.2.0` | RBAC roles, permission bitfield | W6 |
| [52-family-voice.md](52-family-voice.md) | `/aether/voice/0.1.0` | WebRTC signaling, SFU, SFrame media | W4 |

### Local Control & Storage

| File | Title | Status |
|------|-------|--------|
| [60-local-control-api.md](60-local-control-api.md) | HTTP control API: all endpoints, auth, error codes | Normative |
| [70-storage-and-key-derivation.md](70-storage-and-key-derivation.md) | SQLCipher params, KDF scheme, bucket layout | Normative |

### Conformance

| File | Title | Status |
|------|-------|--------|
| [90-conformance-harness.md](90-conformance-harness.md) | W0–W6 levels, pass/fail criteria, KAT format, release gate | Normative |
| [91-test-vectors/](91-test-vectors/) | KAT JSON files (RFC-derived + implementation-pinned) | Normative |

## Source of truth

| Concern | Authoritative source |
|---------|---------------------|
| Protocol IDs | `pkg/protocol/registry.go` |
| Capability flags | `pkg/protocol/capabilities.go` |
| Wire types | `proto/aether.proto` |
| Generated Go bindings | `gen/go/proto/aether.pb.go` (do not edit) |
| Crypto primitives | `pkg/crypto/` |
| Reference KATs | `pkg/v0_1/spectest/` and `docs/spec/v0.1/91-test-vectors/` |

Spec documents cite these sources; they do not re-define them.

## Normative language

RFC 2119 / RFC 8174 key words (MUST, MUST NOT, REQUIRED, SHALL, SHALL NOT,
SHOULD, SHOULD NOT, RECOMMENDED, NOT RECOMMENDED, MAY, OPTIONAL) carry their
standard meanings throughout all documents in this directory.

## Versioning

This spec describes protocol version `v0.1`. Breaking changes require a new
spec version directory (`docs/spec/v0.2/`) and a new multistream major version.
Additive changes within v0.1 use minor-version bumps in individual protocol
families per `03-protocol-registry-and-negotiation.md §7`.

## Reference implementation

Canonical binary: `bin/aether` built by `make build`.
Repository: `github.com/xorein/xorein` (branch `main`, tag `v0.1.0` pending).
Conformance KATs: `docs/spec/v0.1/91-test-vectors/` + `pkg/v0_1/spectest/`.

## Legacy documents

The following documents are **superseded** by this specification:

| Legacy document | Superseded by |
|-----------------|--------------|
| `aether-v3.md` | This spec directory (all documents) |
| `ENCRYPTION_PLUS.md` | `01-cryptographic-primitives.md`, `11-mode-seal.md` – `15-mode-mediashield.md` |
| `aether-addendum-qol-discovery.md` | `31-discovery.md`, `32-nat-traversal.md` |
| `docs/local-control-api-v1.md` | `60-local-control-api.md` |

Legacy documents are preserved for historical context. After the v0.1.0 tag,
they will be archived under `docs/legacy/`.
