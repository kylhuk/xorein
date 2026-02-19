<!-- Replace kylhuk/YOUR_REPO and links. Keep only what is true for your implementation. -->

<p align="center">
  <img src="assets/xorein-banner.png" alt="Xorein" width="900" />
</p>

<h1 align="center">Xorein</h1>

<p align="center">
  A peer-to-peer protocol for end-to-end encrypted messaging, groups, and optional user-run nodes.
</p>

<p align="center">
  <a href="https://github.com/kylhuk/xorein/releases"><img alt="Release" src="https://img.shields.io/github/v/release/kylhuk/xorein?display_name=tag&sort=semver"></a>
  <a href="LICENSE"><img alt="License" src="https://img.shields.io/github/license/kylhuk/xorein"></a>
  <a href="https://github.com/kylhuk/xorein/actions"><img alt="Build" src="https://img.shields.io/github/actions/workflow/status/kylhuk/xorein/ci.yml"></a>
  <a href="https://github.com/kylhuk/xorein/security/policy"><img alt="Security Policy" src="https://img.shields.io/badge/security-policy-blue"></a>
  <a href="https://securityscorecards.dev/viewer/?uri=github.com/kylhuk/xorein"><img alt="OpenSSF Scorecard" src="https://img.shields.io/ossf-scorecard/github.com/kylhuk/xorein?label=OpenSSF%20Scorecard"></a>
</p>

<p align="center">
  <img alt="Spec" src="https://img.shields.io/badge/spec-versioned-informational">
  <img alt="Crypto" src="https://img.shields.io/badge/crypto-agile-informational">
  <img alt="Network" src="https://img.shields.io/badge/network-P2P-informational">
</p>

---

## What Xorein is

Xorein is the protocol layer behind Harmolyn. It defines:

- identity and device keys
- end-to-end encryption for messages and media
- peer discovery and routing over a peer-to-peer network
- optional nodes that improve availability (without gaining the ability to read E2EE content)
- version negotiation and crypto suite agility

Xorein is intended for **advanced end users** who want to understand (and optionally operate) the network they rely on.

## Security model (plain language)

Xorein aims to provide:

- **Confidentiality:** only intended participants can read content (E2EE)
- **Authentication:** participants can verify who they are talking to
- **Forward secrecy:** compromising a device later does not reveal all past traffic (depends on suite)
- **Crypto agility:** algorithms can be upgraded without rewriting the whole protocol

Limits (typical for P2P + E2EE systems):
- metadata (IP/timing/traffic analysis) may be visible to network participants
- discovery and availability may depend on nodes being online

## Crypto suites

Xorein should expose the active crypto suite in the client UI.
This repository should define the canonical list of supported suites and their IDs.

Example format (replace with your actual suites):

| Suite ID | Key agreement | Signatures | AEAD | Hash/KDF | Typical use |
|---:|---|---|---|---|---|
| 1 | X25519 | Ed25519 | XChaCha20-Poly1305 | HKDF-SHA-256 | default messages + media |
| 2 | P-256 | ECDSA P-256 | AES-256-GCM | HKDF-SHA-256 | FIPS-oriented environments |

See: `docs/crypto.md`

## Nodes (what they do)

Nodes are network participants that can provide some combination of:
- discovery (help peers find each other)
- routing/relaying (help messages traverse NATs/firewalls)
- encrypted object storage replication (store ciphertext only)
- public data distribution (e.g., signed network announcements)

Nodes **do not** automatically mean “can read your chats” if E2EE is correctly implemented.

## What happens when nodes change?

- With **one node**, the network can function but becomes availability-centralized (security still depends on E2EE).
- Adding nodes improves resilience and routing options.
- Removing nodes triggers rerouting and (if you replicate data) re-replication.

Clients should surface this as a connectivity/availability signal, not as a security downgrade (unless a downgrade actually occurs).

## Versioning & compatibility

Xorein should have:
- a protocol version handshake
- a capability list (features, suites)
- a policy for “minimum supported version” (to handle security fixes)

See: `docs/versioning.md`

## Operating a node (optional)

If you provide prebuilt binaries, document:
- hardware/network requirements
- ports and NAT expectations
- where encrypted data may be stored on disk
- how to update safely
- how the node learns “minimum supported versions”

See: `docs/node.md`

## Threat model

Document what you defend against and what you do not:

- passive network observers
- malicious peers
- compromised nodes
- compromised client device
- metadata adversaries (traffic analysis)

See: `docs/threat-model.md`

## Reporting vulnerabilities

See `SECURITY.md`.

## License

See `LICENSE`.
