# 04 — Security Modes

Xorein defines six security modes. Every conversation scope (DM, channel, voice
session) carries exactly one mode. Modes are non-negotiable downward by
infrastructure — only the originating parties can choose a mode.

## 1. Mode identifiers

| Mode | String ID | Proto enum | Default for |
|------|-----------|-----------|-------------|
| Seal | `"seal"` | `SECURITY_MODE_SEAL = 1` | DMs |
| Tree | `"tree"` | `SECURITY_MODE_TREE = 2` | Servers ≤50 members |
| Crowd | `"crowd"` | `SECURITY_MODE_CROWD = 4` | Servers >50 members |
| Channel | `"channel"` | `SECURITY_MODE_CHANNEL = 5` | Broadcast channels |
| MediaShield | `"mediashield"` | `SECURITY_MODE_MEDIA_SHIELD = 6` | Voice/screen-share |
| Clear | `"clear"` | `SECURITY_MODE_CLEAR = 3` | Explicit opt-in only |

Proto values 1–7 are allocated. Value 3 (`SECURITY_MODE_CLEAR`) is deliberately
out of logical order; it is assigned historically and MUST NOT be renumbered.

## 2. Mode descriptions

### 2.1 Seal — 1:1 E2EE

Applied to: direct messages.

Cryptography: hybrid X3DH (X25519 + ML-KEM-768) for initial key establishment;
Double Ratchet for per-message forward secrecy. Details in `11-mode-seal.md`.

Properties: confidentiality, integrity, sender authentication, forward secrecy,
post-compromise recovery, store-now-decrypt-later resistance.

Relay behavior: relay nodes MUST NOT store Seal deliveries whose `body` field
is not AEAD ciphertext. Any `relay.store` request with a cleartext body for a
Seal-mode DM MUST be rejected with `RELAY_OPACITY_VIOLATION`.

### 2.2 Tree — Interactive group E2EE

Applied to: servers with expected membership ≤50.

Cryptography: hybrid MLS (RFC 9420 ciphersuite 0xFF01). Details in
`12-mode-tree.md`.

Properties: group confidentiality, membership authentication, O(log N) rekey,
forward secrecy, post-compromise security, store-now-decrypt-later resistance
for new messages after each commit.

### 2.3 Crowd — Sender-key large-group E2EE

Applied to: servers with expected membership >50.

Cryptography: epoch sender keys derived via HKDF-SHA-256. Details in
`13-mode-crowd.md`.

Epoch keys rotate on: membership change, 1000 messages, or 7 days (whichever
comes first). Legacy window: at most 2 prior epoch keys accepted.

Properties: confidentiality from infrastructure; epoch-level forward secrecy;
weaker post-compromise security than Tree (epoch-bound, not per-message).

### 2.4 Channel — Broadcast-epoch E2EE

Applied to: server-wide announcement channels with few writers and many readers.

Cryptography: same epoch key structure as Crowd; writers also produce
per-message authentication signatures. Details in `14-mode-channel.md`.

Properties: same as Crowd; reader list is not authenticated (any holder of the
epoch key can decrypt).

### 2.5 MediaShield — SFrame voice/screen-share E2EE

Applied to: voice frames and screen-share data within voice-capable channels.

Cryptography: RFC 9605 SFrame with per-participant keys derived from the parent
scope's group key material (MLS exporter for Tree; HKDF for Crowd/Channel).
Details in `15-mode-mediashield.md`.

Properties: frame-level confidentiality, integrity, and sender authentication.
Key rotation on participant join/leave. Codec-agnostic (Opus for audio,
AV1 optional for video).

### 2.6 Clear — Explicit plaintext

Applied to: any scope where participants have explicitly opted in.

Characteristics: message bodies readable by infrastructure (relays,
archivists). MUST be labeled in UI with a prominent indicator. MUST NOT be
the default for any private scope. MUST require explicit user action to enable.

Details in `10-mode-clear.md`.

## 3. Mode selection rules

### 3.1 DMs

New DMs MUST default to Seal. The creating client sets `dm.security_mode =
"seal"` in the `DMRecord`. A DM MUST NOT be created in Clear mode without both
participants explicitly opting in (both acknowledge the UI label in a distinct
confirmation step).

### 3.2 Server creation

`CreateServer` defaults to `"tree"` if `security_mode` is empty. Clients
SHOULD follow:

```
expected_members <= 50  →  tree
expected_members > 50   →  crowd
broadcast-only channel  →  channel
voice session           →  mediashield (layered on top of channel/crowd/tree)
```

The manifest carries both `security_mode` (the negotiated mode) and
`offered_security_modes` (all modes the server owner is willing to operate in).

### 3.3 Voice sessions

Voice sessions inherit the parent channel's mode. MediaShield encryption is
layered on top — the parent mode provides the key material for MediaShield key
derivation.

### 3.4 Mode downgrade invariant

A relay MUST reject any `relay.store` request for a delivery where:
- The delivery is for a Seal/Tree/Crowd/Channel/MediaShield scope, AND
- The delivery payload body is not AEAD ciphertext (i.e., the `body` field is
  plaintext rather than base64url-encoded ciphertext).

Error code: `RELAY_OPACITY_VIOLATION`. This invariant is enforced in
`peerRelayStore()` (`pkg/node/service.go`).

## 4. `offered_security_modes`

`ServerRecord.OfferedSecurityModes` and `Manifest.OfferedSecurityModes` list
every mode the server owner is willing to negotiate. Order is the owner's
preference order (most preferred first). Clients MUST display this list for
mode-selection UI.

The reference implementation populates `offered_security_modes` with
`["tree", "crowd", "channel", "seal", "clear"]` on new servers when the owner
has not specified a preference.

## 5. Mode audit events

Every mode change MUST emit a telemetry event:

```json
{"type": "mode.changed", "time": "...", "payload": {
    "scope_type": "...", "scope_id": "...",
    "old_mode": "...", "new_mode": "..."
}}
```

Mode changes require both-party consent (explicit re-negotiation for DMs;
owner-only decision for servers with manifest update). A node MUST NOT
silently change the mode of an existing scope.

## 6. Cross-references

| Mode | Full specification |
|------|--------------------|
| Clear | `10-mode-clear.md` |
| Seal | `11-mode-seal.md` |
| Tree | `12-mode-tree.md` |
| Crowd | `13-mode-crowd.md` |
| Channel | `14-mode-channel.md` |
| MediaShield | `15-mode-mediashield.md` |
