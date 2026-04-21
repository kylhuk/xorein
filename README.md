<div align="center">
  <h1>xorein</h1>
  <p><b>Protocol + runtime</b> for a secure-by-default, P2P-first chat system with explicit security modes.</p>

  <p>
    <img alt="protocol" src="https://img.shields.io/badge/protocol-xorein-black" />
    <img alt="p2p" src="https://img.shields.io/badge/network-libp2p%20%2B%20DHT%20%2B%20PubSub-blue" />
    <img alt="crypto" src="https://img.shields.io/badge/crypto-Noise%20%2B%20E2EE%20profiles-success" />
    <img alt="invariant" src="https://img.shields.io/badge/invariant-single%20binary%2C%20multi--mode-informational" />
  </p>

  <p>
    <a href="https://github.com/kylhuk/xorein/releases"><img alt="release" src="https://img.shields.io/github/v/release/kylhuk/xorein?display_name=tag&sort=semver" /></a>
    <a href="https://github.com/kylhuk/xorein/actions"><img alt="build" src="https://img.shields.io/github/actions/workflow/status/kylhuk/xorein/ci.yml" /></a>
    <a href="https://github.com/kylhuk/xorein/security"><img alt="security-policy" src="https://img.shields.io/badge/security-policy-blue" /></a>
    <a href="https://opensource.org/licenses/AGPL-3.0"><img alt="license" src="https://img.shields.io/github/license/kylhuk/xorein" /></a>
    <a href="https://creativecommons.org/licenses/by-sa/4.0/"><img alt="spec-license" src="https://img.shields.io/badge/spec-CC--BY--SA%204.0-lightgrey" /></a>
  </p>
</div>

---

## What xorein is

xorein is both:

1) the **protocol family** (wire formats + versioning + security modes), and
2) the **runtime binary** that can operate in multiple roles.

No privileged “authority node class” exists. Differences are capability enablement only.

---

## Roles (single binary, multiple modes)

- **client**: P2P node used by Harmolyn
- **relay**: connectivity + store-and-forward (ciphertext only)
- **bootstrap**: minimal DHT entrypoint
- **archivist** (capability role): long-lived ciphertext history segments + manifests + coverage semantics
- **blob provider** (capability role): ciphertext attachments/assets distribution plane

---

## Network design (what actually happens)

Discovery is layered and resilient:
- local peer cache → LAN discovery → bootstrap list/DNS → DHT walking → rendezvous → peer exchange → manual peers

Transport security uses **Noise-secured libp2p connections** (fixed suite policy by default). Content security is per-surface (see below).

---

## Security modes (explicit, user-visible)

xorein defines per-surface security modes so “E2EE” is never a vague claim:

- **Seal**: 1:1 E2EE (X3DH + Double Ratchet)
- **Tree**: interactive group E2EE (MLS)
- **Crowd / Channel**: large-scale E2EE using epoch rotation (removal semantics are rotation-based)
- **MediaShield**: E2EE media frames via SFrame where supported (fallback must be explicit)
- **Clear**: readable by infrastructure (must be labeled; not default for private conversations)

---

## Compatibility and versioning (built-in fragmentation control)

Three layers are used together:
1) **Multistream-select**: per-subprotocol version negotiation (`/aether/<subsystem>/<major>.<minor>.<patch>`)
2) **Protobuf evolution**: additive-only for minor changes
3) **Capabilities exchange**: feature-level negotiation without forcing protocol bumps

Breaking changes:
- ship under new multistream protocol IDs,
- require downgrade negotiation,
- require governance evidence and multi-implementation validation.

---

## Operational guarantees (what you can expect)

- Relays store **ciphertext only** and enforce quotas + TTLs.
- Store-and-forward supports an MVP “relay-local” profile and a replicated “DHT” profile.
- History/search surfaces are retention-aware and expose **coverage labels** and **durability labels**.

---

## License

- Runtime/client/backend code: **AGPL-3.0**
- Protocol/spec text: **CC-BY-SA 4.0**
